/* dq3_audio — 見 dq3_audio.h。SDL2 音訊子系統,播放 MBG.MCX 的 SB FM 音樂。 */
#include "dq3_audio.h"
#include "dq3_opl2.h"
#include "dq3_cmf.h"
#include "dq3_voc.h"
#include <SDL.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define AUD_SR    44100
#define AUD_BUF   1024
#define MAX_TRACK 64
#define SFX_VOICES 8

static struct {
    int            inited;
    int            enabled;
    int            have_device;
    SDL_AudioDeviceID dev;
    unsigned char *mcx;          /* MBG.MCX 全檔 */
    long           mcx_len;
    long           off[MAX_TRACK];
    long           tlen[MAX_TRACK];
    int            ntrack;
    /* 播放狀態(callback 與主執行緒共享,改動時 LockAudioDevice) */
    dq3_opl2      *opl;
    dq3_cmf       *cmf;
    int            cur_track;
    int            loop;
    int            volume;       /* 0..100 */
    double         tick_ms;
    /* 音效 */
    int            sfx_enabled;
    dq3_voc_bank   bankF, bankN;
    struct { const short *pcm; int len, pos; } voice[SFX_VOICES];
    /* 音源後端:0=SB FM 即時合成、1=MT-32 預錄 WAV(從使用者 MT-32 ROM render)*/
    int            backend;          /* DQ3_AUDIO_SB / DQ3_AUDIO_MT32 */
    char           mt32_dir[1024];   /* 放 track_NN.wav 的目錄 */
    int            mt32_avail;       /* 該目錄是否有可用 WAV(否則 MT-32 模式退回 SB)*/
    short         *wav;              /* 目前 MT-32 軌的 mono s16 */
    int            wav_len, wav_rate;
    double         wav_pos;          /* 小數播放位置(重取樣用)*/
} A;

/* 場景 → 軌號。MBG.MCX 18 軌(0-17)= 主音樂;EBG.MCX 6 軌(18-23)= 事件/場所短曲
 * (使用者聽感 2026-06-28:18=旅館/教堂復活、19=神宮…**非戰鬥曲**;先前「EBG=戰鬥」假設已更正)。
 * 使用者(玩過破關)聽感確認:TITLE=0(開頭)、FIELD=1(地圖走路)、TOWN=2、DUNGEON=3、BATTLE=14(地圖遇敵)、SHIP=6(搭船)、ENDING=17(結局返宮)。
 * 迷宮變體(用同一 DUNGEON slot=3):7=塔、9=鬼船地點、16=最後迷宮——日後可按 CTY 接專屬軌。
 * 待確認(? 標,結構推測):BOSS=12、CASTLE=8 —— 待使用者再聽更正。
 * EBG 事件 cue(無場景 slot,先載著):18=教堂復活、19=神宮、20=旅館睡覺;日後可接旅館/祠堂/復活事件。 */
static int g_scene_track[DQ3_MUS__COUNT] = {
    /* TITLE */ 0,  /* FIELD */ 1,  /* TOWN */ 2,   /* DUNGEON */ 3,
    /* BATTLE*/ 14, /* BOSS */ 12,  /* CASTLE */ 8, /* SHIP */ 6,
    /* ENDING*/ 17
};

static int parse_mcx_table(void)
{
    long i = 0; A.ntrack = 0;
    while (i + 4 <= A.mcx_len && A.ntrack < MAX_TRACK) {
        unsigned long v = (unsigned long)A.mcx[i] | ((unsigned long)A.mcx[i+1] << 8)
                        | ((unsigned long)A.mcx[i+2] << 16) | ((unsigned long)A.mcx[i+3] << 24);
        if (A.ntrack > 0 && (long)v <= A.off[A.ntrack-1]) break;
        if ((long)v >= A.mcx_len || (long)v < (A.ntrack + 1) * 4) break;
        A.off[A.ntrack++] = (long)v; i += 4;
    }
    for (i = 0; i < A.ntrack; i++)
        A.tlen[i] = (i + 1 < A.ntrack ? A.off[i+1] : A.mcx_len) - A.off[i];
    return A.ntrack;
}

/* 極簡 WAV(RIFF / PCM s16)讀取 → mono s16 buffer。回 1 成功;支援 mono/stereo、任意取樣率。 */
static int load_wav_mono(const char *path, short **out, int *out_len, int *out_rate)
{
    FILE *f = fopen(path, "rb"); unsigned char hdr[12]; long flen;
    int ch = 0, rate = 0, bits = 0; long data_off = 0, data_len = 0;
    if (!f) return 0;
    fseek(f, 0, SEEK_END); flen = ftell(f); fseek(f, 0, SEEK_SET);
    if (fread(hdr, 1, 12, f) != 12 || memcmp(hdr, "RIFF", 4) || memcmp(hdr+8, "WAVE", 4)) { fclose(f); return 0; }
    while (ftell(f) + 8 <= flen) {
        unsigned char ck[8]; long csz;
        if (fread(ck, 1, 8, f) != 8) break;
        csz = (long)ck[4] | ((long)ck[5]<<8) | ((long)ck[6]<<16) | ((long)ck[7]<<24);
        if (!memcmp(ck, "fmt ", 4)) {
            unsigned char fb[16];
            if (fread(fb, 1, (csz<16?csz:16), f) < 16) { fclose(f); return 0; }
            ch = fb[2] | (fb[3]<<8); rate = fb[4]|(fb[5]<<8)|(fb[6]<<16)|(fb[7]<<24); bits = fb[14]|(fb[15]<<8);
            if (csz > 16) fseek(f, csz - 16, SEEK_CUR);
        } else if (!memcmp(ck, "data", 4)) {
            data_off = ftell(f); data_len = csz; fseek(f, csz, SEEK_CUR);
        } else fseek(f, csz, SEEK_CUR);
        if (csz & 1) fseek(f, 1, SEEK_CUR);   /* word 對齊 */
    }
    if (bits != 16 || ch < 1 || data_off == 0 || data_len <= 0) { fclose(f); return 0; }
    {
        long nsamp = data_len / 2;                 /* total int16 */
        long frames = nsamp / ch;
        short *raw = (short *)malloc(data_len);
        short *mono = (short *)malloc((frames>0?frames:1) * sizeof(short));
        long i;
        if (!raw || !mono) { free(raw); free(mono); fclose(f); return 0; }
        fseek(f, data_off, SEEK_SET);
        if (fread(raw, 1, data_len, f) != (size_t)data_len) { free(raw); free(mono); fclose(f); return 0; }
        for (i = 0; i < frames; i++) {
            if (ch == 1) mono[i] = raw[i];
            else { int s = (int)raw[i*ch] + (int)raw[i*ch+1]; mono[i] = (short)(s/2); }  /* stereo→mono */
        }
        free(raw); fclose(f);
        *out = mono; *out_len = (int)frames; *out_rate = rate;
        return 1;
    }
}

#ifdef DQ3_HAVE_VORBIS
#include <vorbis/vorbisfile.h>
/* 用 libvorbisfile 解碼 OGG → mono s16(編譯有偵測到 libvorbis 時才有)。 */
static int load_ogg_mono(const char *path, short **out, int *out_len, int *out_rate)
{
    OggVorbis_File vf; vorbis_info *vi; short *buf = NULL; long cap = 0, n = 0;
    int ch, rate, bs = 0;
    if (ov_fopen(path, &vf) != 0) return 0;
    vi = ov_info(&vf, -1); if (!vi) { ov_clear(&vf); return 0; }
    ch = vi->channels; rate = (int)vi->rate;
    for (;;) {
        short tmp[2048];
        long got = ov_read(&vf, (char *)tmp, sizeof tmp, 0, 2, 1, &bs);  /* s16, little-endian */
        long frames, i;
        if (got <= 0) break;
        frames = got / 2 / ch;
        if (n + frames > cap) { cap = (n + frames) * 2 + 4096; buf = (short *)realloc(buf, cap * sizeof(short)); if (!buf) { ov_clear(&vf); return 0; } }
        for (i = 0; i < frames; i++) {
            if (ch == 1) buf[n+i] = tmp[i];
            else { int s = (int)tmp[i*ch] + (int)tmp[i*ch+1]; buf[n+i] = (short)(s/2); }
        }
        n += frames;
    }
    ov_clear(&vf);
    if (n <= 0) { free(buf); return 0; }
    *out = buf; *out_len = (int)n; *out_rate = rate;
    return 1;
}
#endif

/* 載入 MT-32 第 track_id 軌的預錄音檔。優先 .ogg(若編譯支援 libvorbis),否則 .wav。 */
static int mt32_load_track(int track_id)
{
    char path[1200]; short *w = NULL; int wl = 0, wr = 0, ok = 0;
#ifdef DQ3_HAVE_VORBIS
    snprintf(path, sizeof path, "%s/track_%02d.ogg", A.mt32_dir, track_id);
    ok = load_ogg_mono(path, &w, &wl, &wr);
#endif
    if (!ok) {
        snprintf(path, sizeof path, "%s/track_%02d.wav", A.mt32_dir, track_id);
        ok = load_wav_mono(path, &w, &wl, &wr);
    }
    if (!ok) return 0;
    if (A.have_device) SDL_LockAudioDevice(A.dev);
    free(A.wav); A.wav = w; A.wav_len = wl; A.wav_rate = wr; A.wav_pos = 0.0;
    if (A.have_device) SDL_UnlockAudioDevice(A.dev);
    return 1;
}

static void SDLCALL audio_cb(void *ud, Uint8 *stream, int len)
{
    short *out = (short *)stream;
    int n = len / (int)sizeof(short);
    (void)ud;
    /* 音樂:MT-32 後端串流預錄 WAV;否則 SB FM 即時合成。 */
    if (A.enabled && A.backend == DQ3_AUDIO_MT32 && A.wav && A.wav_len > 0) {
        double step = (double)A.wav_rate / AUD_SR;   /* 來源 rate → 輸出 rate 重取樣 */
        int i, v = A.volume;
        for (i = 0; i < n; i++) {
            int p = (int)A.wav_pos; double fr = A.wav_pos - p;
            int s0 = A.wav[p % A.wav_len];
            int s1 = A.wav[(p + 1) % A.wav_len];
            int s = (int)(s0 + (s1 - s0) * fr);
            out[i] = (short)(s * v / 100);
            A.wav_pos += step;
            if (A.wav_pos >= A.wav_len) { if (A.loop) A.wav_pos -= A.wav_len; else { A.wav_pos = A.wav_len; out[i]=0; } }
        }
    } else if (A.enabled && A.backend == DQ3_AUDIO_SB && A.cmf && A.opl) {
        dq3_cmf_render(A.cmf, out, n, A.loop);
        if (A.volume < 100) {
            int i, v = A.volume;
            for (i = 0; i < n; i++) out[i] = (short)((int)out[i] * v / 100);
        }
    } else {
        memset(stream, 0, len);
    }
    /* 音效:混在音樂之上(加總 + 飽和)。 */
    if (A.sfx_enabled) {
        int vi;
        for (vi = 0; vi < SFX_VOICES; vi++) {
            const short *pcm = A.voice[vi].pcm;
            int pos = A.voice[vi].pos, vl = A.voice[vi].len, i;
            if (!pcm) continue;
            for (i = 0; i < n && pos < vl; i++, pos++) {
                int m = (int)out[i] + pcm[pos];
                out[i] = (short)(m > 32767 ? 32767 : (m < -32768 ? -32768 : m));
            }
            A.voice[vi].pos = pos;
            if (pos >= vl) A.voice[vi].pcm = NULL;  /* 播完釋放 voice */
        }
    }
}

int dq3_audio_init(const char *assets_dir)
{
    char path[1024]; FILE *f; SDL_AudioSpec want, got;
    if (A.inited) return 0;
    memset(&A, 0, sizeof A);
    A.enabled = 1; A.cur_track = -1; A.volume = 70; A.tick_ms = 13.0;

    snprintf(path, sizeof path, "%s/MBG.MCX", assets_dir);
    f = fopen(path, "rb");
    if (f) {
        fseek(f, 0, SEEK_END); A.mcx_len = ftell(f); fseek(f, 0, SEEK_SET);
        A.mcx = (unsigned char *)malloc(A.mcx_len);
        if (A.mcx && fread(A.mcx, 1, A.mcx_len, f) == (size_t)A.mcx_len) parse_mcx_table();
        fclose(f);
    }
    if (A.ntrack <= 0) {
        fprintf(stderr, "dq3_audio: 找不到/解析不了 MBG.MCX(%s),音樂停用\n", path);
        /* 不致命:降級成無音樂 */
    }

    /* EBG.MCX = 戰鬥音樂(6 軌)獨立檔(RE docs/61):原版戰鬥/boss 曲在此檔,非 MBG。
     * 串接進同一 buffer,軌號接在 MBG 之後(18..23);g_scene_track[BATTLE/BOSS] 指向這些軌。
     * MT-32 側對應 work/mt32/track_18..23.ogg(munt render,export_music_mt32.sh)。 */
    if (A.ntrack > 0) {
        char ep[1024]; FILE *ef;
        snprintf(ep, sizeof ep, "%s/EBG.MCX", assets_dir);
        ef = fopen(ep, "rb");
        if (ef) {
            long elen; unsigned char *ebuf;
            fseek(ef, 0, SEEK_END); elen = ftell(ef); fseek(ef, 0, SEEK_SET);
            ebuf = (unsigned char *)malloc(elen > 0 ? elen : 1);
            if (ebuf && elen > 8 && fread(ebuf, 1, elen, ef) == (size_t)elen) {
                unsigned char *nb = (unsigned char *)realloc(A.mcx, A.mcx_len + elen);
                if (nb) {
                    long base = A.mcx_len, j = 0; int e0 = A.ntrack;
                    A.mcx = nb; memcpy(A.mcx + base, ebuf, elen); A.mcx_len += elen;
                    while (j + 4 <= elen && A.ntrack < MAX_TRACK) {   /* 解 EBG 偏移表(同 MBG 格式) */
                        unsigned long v = (unsigned long)ebuf[j] | ((unsigned long)ebuf[j+1] << 8)
                                        | ((unsigned long)ebuf[j+2] << 16) | ((unsigned long)ebuf[j+3] << 24);
                        if (A.ntrack > e0 && (long)v <= A.off[A.ntrack-1] - base) break;
                        if ((long)v >= elen || (long)v < (long)((A.ntrack - e0 + 1) * 4)) break;
                        A.off[A.ntrack++] = base + (long)v; j += 4;
                    }
                    for (j = e0; j < A.ntrack; j++)
                        A.tlen[j] = (j + 1 < A.ntrack ? A.off[j+1] : A.mcx_len) - A.off[j];
                }
            }
            free(ebuf); fclose(ef);
        }
    }

    A.opl = dq3_opl2_new(AUD_SR);

    /* 音效 bank(FVOC/NVOC.VCX)*/
    A.sfx_enabled = 1;
    snprintf(path, sizeof path, "%s/FVOC.VCX", assets_dir);
    dq3_voc_load(&A.bankF, path, AUD_SR);
    snprintf(path, sizeof path, "%s/NVOC.VCX", assets_dir);
    dq3_voc_load(&A.bankN, path, AUD_SR);

    /* MT-32 音源後端:預錄 WAV 目錄(env DQ3_MT32_DIR 可覆寫,預設 <assets>/mt32)。
     * 偵測 track_00.wav 是否存在 → 決定是否提供 MT-32 切換。 */
    A.backend = DQ3_AUDIO_SB;
    { const char *md = getenv("DQ3_MT32_DIR");
      if (md && *md) snprintf(A.mt32_dir, sizeof A.mt32_dir, "%s", md);
      else snprintf(A.mt32_dir, sizeof A.mt32_dir, "%s/mt32", assets_dir); }
    { char p[1200]; FILE *tf;
#ifdef DQ3_HAVE_VORBIS
      snprintf(p, sizeof p, "%s/track_00.ogg", A.mt32_dir);
      tf = fopen(p, "rb"); if (tf) { A.mt32_avail = 1; fclose(tf); }
#endif
      if (!A.mt32_avail) { snprintf(p, sizeof p, "%s/track_00.wav", A.mt32_dir);
          tf = fopen(p, "rb"); if (tf) { A.mt32_avail = 1; fclose(tf); } } }

    SDL_zero(want);
    want.freq = AUD_SR; want.format = AUDIO_S16SYS; want.channels = 1;
    want.samples = AUD_BUF; want.callback = audio_cb;
    A.dev = SDL_OpenAudioDevice(NULL, 0, &want, &got, 0);
    if (A.dev == 0) {
        fprintf(stderr, "dq3_audio: SDL_OpenAudioDevice 失敗(%s),音樂停用(headless?)\n", SDL_GetError());
        A.have_device = 0;
    } else {
        A.have_device = 1;
        SDL_PauseAudioDevice(A.dev, 0);     /* 開始播放 callback */
    }
    A.inited = 1;
    fprintf(stderr, "dq3_audio: init OK(軌數=%d, 音效 F=%d N=%d, MT-32=%s, device=%s)\n",
            A.ntrack, A.bankF.count, A.bankN.count,
            A.mt32_avail ? A.mt32_dir : "無(僅 SB)", A.have_device ? "on" : "off/headless");
    return 0;
}

void dq3_audio_shutdown(void)
{
    if (!A.inited) return;
    if (A.have_device) { SDL_LockAudioDevice(A.dev); }
    if (A.cmf) { dq3_cmf_free(A.cmf); A.cmf = NULL; }
    if (A.have_device) { SDL_UnlockAudioDevice(A.dev); SDL_CloseAudioDevice(A.dev); }
    if (A.opl) { dq3_opl2_free(A.opl); A.opl = NULL; }
    dq3_voc_free(&A.bankF); dq3_voc_free(&A.bankN);
    free(A.wav); A.wav = NULL;
    free(A.mcx); A.mcx = NULL;
    memset(&A, 0, sizeof A);
}

void dq3_audio_play(int track_id, int loop)
{
    if (!A.inited || A.ntrack <= 0) return;
    if (track_id < 0 || track_id >= A.ntrack) return;
    if (track_id == A.cur_track && (A.cmf || A.wav)) { A.loop = loop; return; }  /* 同軌不重啟 */

    if (A.backend == DQ3_AUDIO_MT32 && A.mt32_avail) {     /* MT-32:載預錄 WAV */
        if (mt32_load_track(track_id)) { A.cur_track = track_id; A.loop = loop; return; }
        /* 該軌 WAV 缺 → 退回 SB 合成 */
    }
    {   /* SB FM 即時合成 */
        dq3_cmf *nc = dq3_cmf_load(A.mcx + A.off[track_id], (int)A.tlen[track_id], A.opl, AUD_SR, A.tick_ms);
        if (!nc) return;
        if (A.have_device) SDL_LockAudioDevice(A.dev);
        if (A.cmf) dq3_cmf_free(A.cmf);
        A.cmf = nc; A.cur_track = track_id; A.loop = loop;
        if (A.have_device) SDL_UnlockAudioDevice(A.dev);
    }
}

void dq3_audio_set_backend(int backend)
{
    int b = (backend == DQ3_AUDIO_MT32) ? DQ3_AUDIO_MT32 : DQ3_AUDIO_SB;
    if (!A.inited || b == A.backend) return;
    if (b == DQ3_AUDIO_MT32 && !A.mt32_avail) return;     /* 無 MT-32 WAV 資產 → 不切 */
    {   /* 切後端:清掉另一邊的播放狀態,重播目前軌 */
        int t = A.cur_track;
        if (A.have_device) SDL_LockAudioDevice(A.dev);
        if (A.cmf) { dq3_cmf_free(A.cmf); A.cmf = NULL; }
        free(A.wav); A.wav = NULL; A.wav_len = 0;
        A.backend = b; A.cur_track = -1;
        if (A.have_device) SDL_UnlockAudioDevice(A.dev);
        if (t >= 0) dq3_audio_play(t, A.loop);
    }
}

int dq3_audio_backend(void) { return A.inited ? A.backend : DQ3_AUDIO_SB; }
int dq3_audio_mt32_available(void) { return A.inited ? A.mt32_avail : 0; }

void dq3_audio_stop(void)
{
    if (!A.inited) return;
    if (A.have_device) SDL_LockAudioDevice(A.dev);
    if (A.cmf) { dq3_cmf_free(A.cmf); A.cmf = NULL; }
    A.cur_track = -1;
    if (A.have_device) SDL_UnlockAudioDevice(A.dev);
}

void dq3_audio_set_enabled(int on)
{
    if (!A.inited) return;
    A.enabled = on ? 1 : 0;
    if (!A.enabled) dq3_audio_stop();
}
int dq3_audio_enabled(void) { return A.inited ? A.enabled : 0; }

void dq3_audio_set_volume(int v)
{
    if (v < 0) v = 0;
    if (v > 100) v = 100;
    A.volume = v;
}

int dq3_audio_track_count(void)   { return A.ntrack; }
int dq3_audio_current_track(void) { return A.inited ? A.cur_track : -1; }

int dq3_audio_scene_track(int scene_kind)
{
    if (scene_kind < 0 || scene_kind >= DQ3_MUS__COUNT) return -1;
    { int t = g_scene_track[scene_kind];
      return (t >= 0 && t < A.ntrack) ? t : -1; }
}

void dq3_audio_play_scene(int scene_kind, int loop)
{
    int t = dq3_audio_scene_track(scene_kind);
    if (t >= 0) dq3_audio_play(t, loop);
}

/* ── 數位音效 ── */
void dq3_audio_sfx(int bank, int sfx_id)
{
    dq3_voc_bank *b = (bank == DQ3_SFX_BANK_N) ? &A.bankN : &A.bankF;
    int vi;
    if (!A.inited || !A.sfx_enabled) return;
    if (sfx_id < 0 || sfx_id >= b->count || !b->sfx[sfx_id].pcm) return;
    if (A.have_device) SDL_LockAudioDevice(A.dev);
    for (vi = 0; vi < SFX_VOICES; vi++) if (!A.voice[vi].pcm) break;
    if (vi >= SFX_VOICES) vi = 0;             /* 全滿 → 搶占 voice 0 */
    A.voice[vi].pcm = b->sfx[sfx_id].pcm;
    A.voice[vi].len = b->sfx[sfx_id].len;
    A.voice[vi].pos = 0;
    if (A.have_device) SDL_UnlockAudioDevice(A.dev);
}

void dq3_audio_set_sfx_enabled(int on) { if (A.inited) A.sfx_enabled = on ? 1 : 0; }
int  dq3_audio_sfx_enabled(void) { return A.inited ? A.sfx_enabled : 0; }
int  dq3_audio_sfx_count(int bank)
{
    if (!A.inited) return 0;
    return (bank == DQ3_SFX_BANK_N) ? A.bankN.count : A.bankF.count;
}

/* 語意音效 → (bank,id)。對應為初步(FVOC 為主),日後依 RE 選音/聽感精校。 */
static const struct { int bank, id; } g_se[DQ3_SE__COUNT] = {
    /* CONFIRM */ {DQ3_SFX_BANK_F, 0}, /* CANCEL */ {DQ3_SFX_BANK_F, 1},
    /* ATTACK  */ {DQ3_SFX_BANK_F, 2}, /* HIT    */ {DQ3_SFX_BANK_F, 3},
    /* SPELL   */ {DQ3_SFX_BANK_F, 7}, /* DAMAGE */ {DQ3_SFX_BANK_F, 4},
    /* DOOR    */ {DQ3_SFX_BANK_F, 5}, /* ITEM   */ {DQ3_SFX_BANK_F, 6}
};
void dq3_audio_se(int se_kind)
{
    if (se_kind < 0 || se_kind >= DQ3_SE__COUNT) return;
    dq3_audio_sfx(g_se[se_kind].bank, g_se[se_kind].id);
}
