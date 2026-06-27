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

static const int g_mod_off[9] = {0x00,0x01,0x02,0x08,0x09,0x0A,0x10,0x11,0x12};

/* 一組各異的 OPL 2-op FM 音色 patch(register 值:mod/car 各 0x20/0x40/0x60/0x80/0xE0 + 0xC0 fb/cnt)。
 * 精訊原版各軌 instrument 區的精確對應需深層 RE 驅動(0x1563d 鏈,見 docs/57);此處用音樂上合理、
 * 聲部分明的多音色,依 program-change 分配 channel,讓 bass/旋律/和聲音色不同(比單一 patch 大幅改善)。
 * 欄位:{m20,m40,m60,m80,mE0, c20,c40,c60,c80,cE0, C0} */
#define NPATCH 8
static const unsigned char g_patch[NPATCH][11] = {
 /* 0 brass lead  */ {0x21,0x1a,0xd5,0x53,0x00, 0x21,0x07,0xd5,0x53,0x00, 0x0a},
 /* 1 bass        */ {0x21,0x10,0xf2,0x74,0x00, 0x21,0x00,0xf2,0x76,0x00, 0x08},
 /* 2 organ(add) */ {0x31,0x00,0xf0,0x90,0x00, 0x31,0x00,0xf0,0x90,0x00, 0x0f},
 /* 3 string pad  */ {0x61,0x1c,0x52,0x24,0x01, 0x61,0x08,0x42,0x16,0x01, 0x0c},
 /* 4 flute/soft  */ {0x21,0x2a,0xf0,0x53,0x00, 0x21,0x12,0xf0,0x53,0x00, 0x0e},
 /* 5 pluck/harp  */ {0x01,0x16,0xf8,0x77,0x00, 0x11,0x06,0xf8,0x88,0x00, 0x0a},
 /* 6 reed/clar   */ {0x31,0x1f,0xd2,0x53,0x02, 0x21,0x09,0xd2,0x55,0x02, 0x0a},
 /* 7 bell/chime  */ {0x01,0x1a,0xf5,0x42,0x01, 0x12,0x05,0xf5,0x52,0x01, 0x0a},
};

/* 把第 idx 個 patch 寫進 OPL channel c。 */
static void apply_patch(dq3_opl2 *opl, int c, int idx)
{
    const unsigned char *p = g_patch[idx % NPATCH];
    int m = g_mod_off[c], car = g_mod_off[c] + 3;
    dq3_opl2_write(opl, 0x20 + m,   p[0]); dq3_opl2_write(opl, 0x40 + m,   p[1]);
    dq3_opl2_write(opl, 0x60 + m,   p[2]); dq3_opl2_write(opl, 0x80 + m,   p[3]);
    dq3_opl2_write(opl, 0xe0 + m,   p[4]);
    dq3_opl2_write(opl, 0x20 + car, p[5]); dq3_opl2_write(opl, 0x40 + car, p[6]);
    dq3_opl2_write(opl, 0x60 + car, p[7]); dq3_opl2_write(opl, 0x80 + car, p[8]);
    dq3_opl2_write(opl, 0xe0 + car, p[9]);
    dq3_opl2_write(opl, 0xc0 + c,   p[10]);
}

/* 初始:啟用 waveform select,各 channel 先給 patch 0(program-change 會覆寫成各聲部音色)。 */
static void load_default_instrument(dq3_opl2 *opl)
{
    int c;
    dq3_opl2_write(opl, 0x01, 0x20);
    for (c = 0; c < 9; c++) apply_patch(opl, c, 0);
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
    } else if (hi == 0xc0) {                      /* program change → 換該 channel 音色 */
        p->prog[ch] = e->d1;
        apply_patch(p->opl, ch, e->d1);
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
