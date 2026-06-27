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
} A;

/* 場景 → 軌號(初步合理對應;MBG.MCX 18 軌,日後依 RE 選曲精校)。 */
static int g_scene_track[DQ3_MUS__COUNT] = {
    /* TITLE */ 17, /* FIELD */ 0,  /* TOWN */ 2,   /* DUNGEON */ 5,
    /* BATTLE*/ 6,  /* BOSS */ 14,  /* CASTLE */ 3, /* SHIP */ 9,
    /* ENDING*/ 16
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

static void SDLCALL audio_cb(void *ud, Uint8 *stream, int len)
{
    short *out = (short *)stream;
    int n = len / (int)sizeof(short);
    (void)ud;
    /* 音樂 */
    if (A.enabled && A.cmf && A.opl) {
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

    A.opl = dq3_opl2_new(AUD_SR);

    /* 音效 bank(FVOC/NVOC.VCX)*/
    A.sfx_enabled = 1;
    snprintf(path, sizeof path, "%s/FVOC.VCX", assets_dir);
    dq3_voc_load(&A.bankF, path, AUD_SR);
    snprintf(path, sizeof path, "%s/NVOC.VCX", assets_dir);
    dq3_voc_load(&A.bankN, path, AUD_SR);

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
    fprintf(stderr, "dq3_audio: init OK(軌數=%d, 音效 F=%d N=%d, device=%s)\n",
            A.ntrack, A.bankF.count, A.bankN.count, A.have_device ? "on" : "off/headless");
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
    free(A.mcx); A.mcx = NULL;
    memset(&A, 0, sizeof A);
}

void dq3_audio_play(int track_id, int loop)
{
    dq3_cmf *nc;
    if (!A.inited || A.ntrack <= 0) return;
    if (track_id < 0 || track_id >= A.ntrack) return;
    if (track_id == A.cur_track && A.cmf) { A.loop = loop; return; }  /* 同軌不重啟 */
    nc = dq3_cmf_load(A.mcx + A.off[track_id], (int)A.tlen[track_id], A.opl, AUD_SR, A.tick_ms);
    if (!nc) return;
    if (A.have_device) SDL_LockAudioDevice(A.dev);
    if (A.cmf) dq3_cmf_free(A.cmf);
    A.cmf = nc; A.cur_track = track_id; A.loop = loop;
    if (A.have_device) SDL_UnlockAudioDevice(A.dev);
}

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
