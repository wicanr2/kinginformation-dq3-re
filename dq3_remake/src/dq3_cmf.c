/* dq3_cmf — 見 dq3_cmf.h。CMF 變體軌 parser + 排程器。 */
#include "dq3_cmf.h"
#include <stdlib.h>
#include <string.h>
#include <math.h>

typedef struct { unsigned char st, d1, d2, dl; } ev_t;

struct dq3_cmf {
    dq3_opl2 *opl;
    int sr;
    double samp_per_tick;
    ev_t *ev; int nev;
    /* 播放狀態 */
    int cursor;             /* 下一個要處理的事件 */
    double samp_to_next;    /* 到下個事件還要幾個樣本 */
    int finished;
    int prog[16];           /* 每 MIDI channel 目前 program(暫未用,留給 instrument 切換) */
    int cur_fnum[9], cur_block[9];   /* 每 OPL channel 目前的 fnum/block(note-off 保留音高用) */
};

/* 解析事件流(MIDI-like + 尾隨 delta + running status)。回傳事件數。 */
static int parse_events(const unsigned char *t, int len, ev_t **out)
{
    int base, p, cap, n; ev_t *ev; int status = -1;
    if (len < 8) { *out = NULL; return 0; }
    base = t[6] | (t[7] << 8);                 /* header word[3],恆 0x60 */
    if (base <= 0 || base >= len) base = 0x60;
    p = base;
    while (p < len && t[p] < 0x80) p++;        /* 跳到第一個 status */
    cap = 256; n = 0; ev = (ev_t *)malloc(cap * sizeof(ev_t));
    if (!ev) { *out = NULL; return 0; }
    while (p < len) {
        int b = t[p], hi;
        if (b >= 0x80) {                        /* 新 status */
            status = b; p++;
            if (status == 0xff || status == 0xf0) break;
        }
        if (status < 0) break;
        hi = status & 0xf0;
        if (hi == 0x80 || hi == 0x90 || hi == 0xb0 || hi == 0xa0 || hi == 0xe0) {
            if (p + 3 > len) break;
            if (n >= cap) { cap *= 2; ev = (ev_t *)realloc(ev, cap * sizeof(ev_t)); }
            ev[n].st = (unsigned char)status; ev[n].d1 = t[p]; ev[n].d2 = t[p+1]; ev[n].dl = t[p+2];
            n++; p += 3;
        } else if (hi == 0xc0 || hi == 0xd0) {
            if (p + 2 > len) break;
            if (n >= cap) { cap *= 2; ev = (ev_t *)realloc(ev, cap * sizeof(ev_t)); }
            ev[n].st = (unsigned char)status; ev[n].d1 = t[p]; ev[n].d2 = 0; ev[n].dl = t[p+1];
            n++; p += 2;
        } else break;
    }
    *out = ev; return n;
}

/* 預設 FM patch:一個 channel 的兩 operator + FB/CNT,寫進 OPL 各 channel。
 * 暖音色(organ/brass 風),供旋律先發聲;之後可換成各軌 instrument 區的實際 patch。 */
static void load_default_instrument(dq3_opl2 *opl)
{
    static const int mod_off[9] = {0x00,0x01,0x02,0x08,0x09,0x0A,0x10,0x11,0x12};
    int c;
    dq3_opl2_write(opl, 0x01, 0x20);             /* 啟用 waveform select */
    for (c = 0; c < 9; c++) {
        int m = mod_off[c], car = mod_off[c] + 3;
        /* modulator */
        dq3_opl2_write(opl, 0x20 + m, 0x21);     /* mult=1, sustain on */
        dq3_opl2_write(opl, 0x40 + m, 0x1a);     /* TL 衰減一些 */
        dq3_opl2_write(opl, 0x60 + m, 0xd5);     /* AR=13 DR=5 */
        dq3_opl2_write(opl, 0x80 + m, 0x53);     /* SL=5 RR=3 */
        dq3_opl2_write(opl, 0xe0 + m, 0x00);     /* sine */
        /* carrier */
        dq3_opl2_write(opl, 0x20 + car, 0x21);
        dq3_opl2_write(opl, 0x40 + car, 0x07);   /* 較大音量 */
        dq3_opl2_write(opl, 0x60 + car, 0xd5);
        dq3_opl2_write(opl, 0x80 + car, 0x53);
        dq3_opl2_write(opl, 0xe0 + car, 0x00);
        dq3_opl2_write(opl, 0xc0 + c, 0x0a);     /* FB=5, CNT=0(FM) */
    }
}

dq3_cmf *dq3_cmf_load(const unsigned char *track, int len, dq3_opl2 *opl,
                      int sample_rate, double tick_ms)
{
    dq3_cmf *p = (dq3_cmf *)calloc(1, sizeof *p);
    if (!p) return NULL;
    p->opl = opl; p->sr = sample_rate;
    p->samp_per_tick = (tick_ms > 0 ? tick_ms : 13.0) * 0.001 * sample_rate;
    p->nev = parse_events(track, len, &p->ev);
    if (p->nev <= 0) { free(p->ev); free(p); return NULL; }
    load_default_instrument(opl);
    p->cursor = 0; p->samp_to_next = 0; p->finished = 0;
    return p;
}

void dq3_cmf_free(dq3_cmf *p) { if (p) { free(p->ev); free(p); } }

/* MIDI note → (fnum, block):取能容下的最小 block 以保精度。 */
static void note_to_fb(int note, int *fnum, int *block)
{
    double freq = 440.0 * pow(2.0, (note - 69) / 12.0);
    double base = freq * 1048576.0 / 49716.0;   /* = fnum at block 0 */
    int b = 0;
    while (base >= 1024.0 && b < 7) { base /= 2.0; b++; }
    *fnum = (int)(base + 0.5); if (*fnum > 1023) *fnum = 1023;
    *block = b;
}

static void apply_event(dq3_cmf *p, const ev_t *e)
{
    int hi = e->st & 0xf0, ch = e->st & 0x0f;
    if (ch >= 9) return;                          /* 只有 9 個 OPL channel */
    if (hi == 0x90 && e->d2 > 0) {                /* note-on */
        int fnum, block;
        note_to_fb(e->d1, &fnum, &block);
        p->cur_fnum[ch] = fnum; p->cur_block[ch] = block;
        dq3_opl2_write(p->opl, 0xa0 + ch, fnum & 0xff);
        dq3_opl2_write(p->opl, 0xb0 + ch, ((block & 7) << 2) | ((fnum >> 8) & 3) | 0x20);
    } else if (hi == 0x80 || (hi == 0x90 && e->d2 == 0)) {  /* note-off:清 keyon bit、保留 fnum/block */
        int fnum = p->cur_fnum[ch], block = p->cur_block[ch];
        dq3_opl2_write(p->opl, 0xb0 + ch, ((block & 7) << 2) | ((fnum >> 8) & 3)); /* keyon=0 */
    } else if (hi == 0xc0) {
        p->prog[ch] = e->d1;                      /* program change:暫存(predefined patch) */
    }
    /* bx control / ex pitch:目前忽略 */
}

int dq3_cmf_render(dq3_cmf *p, short *out, int n, int loop)
{
    int done = 0;
    while (done < n) {
        if (p->samp_to_next <= 0.0) {
            if (p->cursor >= p->nev) {
                if (loop) { p->cursor = 0; }
                else { p->finished = 1; break; }
            }
            if (p->cursor < p->nev) {
                ev_t *e = &p->ev[p->cursor++];
                apply_event(p, e);
                p->samp_to_next += e->dl * p->samp_per_tick;
            }
            continue;                              /* 同一時間點可能有多個 delta=0 事件 */
        }
        {
            int chunk = (int)p->samp_to_next;
            if (chunk > n - done) chunk = n - done;
            if (chunk < 1) chunk = 1;
            dq3_opl2_render(p->opl, out + done, chunk);
            done += chunk; p->samp_to_next -= chunk;
        }
    }
    if (done < n) {                                /* 結束後補靜音 */
        memset(out + done, 0, (n - done) * sizeof(short));
        done = n;
    }
    return !p->finished;
}

int dq3_cmf_event_count(const dq3_cmf *p) { return p ? p->nev : 0; }

double dq3_cmf_length_sec(const dq3_cmf *p)
{
    long ticks = 0; int i;
    if (!p) return 0;
    for (i = 0; i < p->nev; i++) ticks += p->ev[i].dl;
    return ticks * p->samp_per_tick / p->sr;
}
