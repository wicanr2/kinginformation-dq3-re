/* dq3_opl2 — 精簡 OPL2 (YM3812) FM 核心實作。見 dq3_opl2.h。
 * 近似:相位累積 + 正弦表(含 OPL2 四波形)+ 線性 ADSR(指數曲線)+ 2-op FM/additive。
 * 暫存器→參數對應依 YM3812 規格;envelope 速率/KSL 為簡化近似(非 cycle-accurate)。
 */
#include "dq3_opl2.h"
#include <stdlib.h>
#include <string.h>
#include <math.h>

#define NCH 9
#define SINLEN 2048
#define OPL_NATIVE 49716.0   /* OPL2 取樣率 = 3.579545MHz/72 */

enum { ENV_OFF = 0, ENV_ATK, ENV_DEC, ENV_SUS, ENV_REL };

typedef struct {
    /* 來自暫存器 */
    int   mult;       /* 倍頻 (已換成實際倍數 ×2,見 mult_tab) */
    double tl_amp;    /* total level → 線性振幅 0..1 */
    int   ar, dr, sl, rr;   /* 0..15 */
    int   wave;       /* 0..3 */
    int   ksr, egt, sus_flag;
    /* runtime */
    double phase;     /* 0..1 */
    double phinc;     /* 每樣本相位增量 */
    int    state;
    double env;       /* 0..1 線性振幅 */
    double last;      /* 上次輸出(回授用) */
} op_t;

typedef struct {
    int fnum, block, keyon, fb, con;
    op_t op[2];       /* 0=modulator, 1=carrier */
} ch_t;

struct dq3_opl2 {
    int sr;
    unsigned char reg[256];
    ch_t ch[NCH];
    float sintab[4][SINLEN];
    int   wse;        /* waveform select enable (reg1 bit5) */
};

/* MULT 暫存器 0..15 → 倍頻 ×2(OPL:0→0.5,1→1,2→2,...,10/11→10,12/13→12,14/15→15) */
static const int mult_x2[16] = {1,2,4,6,8,10,12,14,16,18,20,20,24,24,30,30};

/* 9 channel 的 modulator operator 暫存器 offset;carrier = +3 */
static const int mod_off[NCH] = {0x00,0x01,0x02,0x08,0x09,0x0A,0x10,0x11,0x12};

/* offset(0x00..0x15)→ (channel<<1 | opidx),非法為 -1 */
static int off2chop(int off, int *ch, int *opi)
{
    int c;
    for (c = 0; c < NCH; c++) {
        if (off == mod_off[c]) { *ch = c; *opi = 0; return 1; }
        if (off == mod_off[c] + 3) { *ch = c; *opi = 1; return 1; }
    }
    return 0;
}

dq3_opl2 *dq3_opl2_new(int sample_rate)
{
    dq3_opl2 *o = (dq3_opl2 *)calloc(1, sizeof *o);
    int i;
    if (!o) return NULL;
    o->sr = sample_rate > 0 ? sample_rate : 44100;
    for (i = 0; i < SINLEN; i++) {
        double t = (double)i / SINLEN;
        double s = sin(2.0 * M_PI * t);
        o->sintab[0][i] = (float)s;                              /* 全正弦 */
        o->sintab[1][i] = (float)(s > 0 ? s : 0.0);              /* 半波整流 */
        o->sintab[2][i] = (float)fabs(s);                        /* 絕對值(全波整流) */
        o->sintab[3][i] = (float)((t < 0.25 || (t >= 0.5 && t < 0.75)) ? fabs(s) : 0.0); /* 鋸狀脈衝 */
    }
    dq3_opl2_reset(o);
    return o;
}

void dq3_opl2_free(dq3_opl2 *o) { free(o); }

void dq3_opl2_reset(dq3_opl2 *o)
{
    int c, k;
    memset(o->reg, 0, sizeof o->reg);
    o->wse = 0;
    for (c = 0; c < NCH; c++) {
        memset(&o->ch[c], 0, sizeof(ch_t));
        for (k = 0; k < 2; k++) { o->ch[c].op[k].tl_amp = 1.0; o->ch[c].op[k].mult = 2; }
    }
}

/* tl 0..63(0.75dB/step)→ 線性振幅 */
static double tl_to_amp(int tl) { return pow(10.0, -(tl * 0.75) / 20.0); }
/* sustain level 0..15(3dB/step,15=極低)→ 線性振幅 */
static double sl_to_amp(int sl) { return sl >= 15 ? 0.0 : pow(10.0, -(sl * 3.0) / 20.0); }

static void update_phinc(dq3_opl2 *o, int c, int k)
{
    ch_t *ch = &o->ch[c];
    op_t *op = &ch->op[k];
    /* OPL 頻率 = fnum × 2^block × (OPL_NATIVE / 2^20);再乘 operator 倍頻 */
    double base = ch->fnum * (double)(1 << ch->block) * (OPL_NATIVE / 1048576.0);
    double f = base * (mult_x2[op->mult & 15] / 2.0);
    op->phinc = f / o->sr;
}

void dq3_opl2_write(dq3_opl2 *o, int reg, int val)
{
    int hi, lo, c, k;
    reg &= 0xff; val &= 0xff;
    o->reg[reg] = (unsigned char)val;
    if (reg == 0x01) { o->wse = (val >> 5) & 1; return; }
    hi = reg & 0xf0; lo = reg & 0x0f;

    if (hi == 0x20 || hi == 0x40 || hi == 0x60 || hi == 0x80 || hi == 0xE0) {
        if (!off2chop(lo, &c, &k)) return;
        { op_t *op = &o->ch[c].op[k];
          switch (hi) {
          case 0x20: op->mult = val & 0x0f; op->egt = (val >> 5) & 1;
                     op->ksr = (val >> 4) & 1; update_phinc(o, c, k); break;
          case 0x40: op->tl_amp = tl_to_amp(val & 0x3f); break;       /* 忽略 KSL 高 2 bit */
          case 0x60: op->ar = (val >> 4) & 0x0f; op->dr = val & 0x0f; break;
          case 0x80: op->sl = (val >> 4) & 0x0f; op->rr = val & 0x0f; break;
          case 0xE0: op->wave = o->wse ? (val & 0x03) : 0; break;
          } }
        return;
    }
    if (hi == 0xA0 && lo < NCH) {
        c = lo; o->ch[c].fnum = (o->ch[c].fnum & 0x300) | val;
        update_phinc(o, c, 0); update_phinc(o, c, 1); return;
    }
    if (hi == 0xB0 && lo < NCH) {
        int newkey;
        c = lo;
        o->ch[c].fnum = (o->ch[c].fnum & 0xff) | ((val & 0x03) << 8);
        o->ch[c].block = (val >> 2) & 0x07;
        newkey = (val >> 5) & 1;
        update_phinc(o, c, 0); update_phinc(o, c, 1);
        if (newkey && !o->ch[c].keyon) {            /* key-on:觸發兩 op attack(env 歸零→乾淨淡入,無 click) */
            for (k = 0; k < 2; k++) {
                o->ch[c].op[k].state = ENV_ATK;
                o->ch[c].op[k].phase = 0.0;
                o->ch[c].op[k].env = 0.0;
            }
        } else if (!newkey && o->ch[c].keyon) {     /* key-off:進入 release */
            for (k = 0; k < 2; k++) o->ch[c].op[k].state = ENV_REL;
        }
        o->ch[c].keyon = newkey; return;
    }
    if (hi == 0xC0 && lo < NCH) {
        c = lo; o->ch[c].fb = (val >> 1) & 0x07; o->ch[c].con = val & 1; return;
    }
}

/* envelope 速率 0..15 → 每樣本變化量(近似:速率越大越快;以時間常數換算)。
 * 關鍵:夾最短 ramp 時間,避免高 AR/RR 算出「一個 sample 跳到底」造成爆音 click。 */
static double atk_step(dq3_opl2 *o, int r)
{
    double ms;
    if (r == 0) return 0.0;
    ms = 8.0 * pow(0.5, (double)r);          /* r 大 → attack 短 */
    if (ms < 3.0) ms = 3.0;                  /* 最短 attack 3ms(無瞬跳 click) */
    return 1.0 / (ms * 0.001 * o->sr + 1.0);
}
static double dec_step(dq3_opl2 *o, int r)
{
    double ms;
    if (r == 0) return 0.0;
    ms = 80.0 * pow(0.5, (double)r);
    if (ms < 6.0) ms = 6.0;                  /* 最短 decay/release 6ms(無瞬切 click) */
    return 1.0 / (ms * 0.001 * o->sr + 1.0);
}

static double op_eval(dq3_opl2 *o, op_t *op, double pmod)
{
    double ph, frac; int idx; float *wt;
    /* 更新 envelope */
    switch (op->state) {
    case ENV_ATK:
        op->env += (1.0 - op->env) * atk_step(o, op->ar);
        if (op->env >= 0.999 || op->ar == 0) { op->env = 1.0; op->state = ENV_DEC; }
        break;
    case ENV_DEC: {
        double tgt = sl_to_amp(op->sl);
        op->env += (tgt - op->env) * dec_step(o, op->dr) * 2.0;
        if (op->env <= tgt + 0.001) { op->env = tgt; op->state = ENV_SUS; }
        break; }
    case ENV_SUS: break;
    case ENV_REL:
        op->env -= op->env * dec_step(o, op->rr);     /* 不再 ×4:平滑 release,不爆音 */
        if (op->env < 0.0003) { op->env = 0.0; op->state = ENV_OFF; }
        break;
    default: return 0.0;
    }
    /* 相位(含調變)→ 波形 */
    ph = op->phase + pmod;
    ph -= floor(ph);
    idx = (int)(ph * SINLEN); if (idx >= SINLEN) idx = SINLEN - 1;
    wt = o->sintab[op->wave & 3];
    frac = ph * SINLEN - idx;
    { double s = wt[idx] + (wt[(idx + 1) & (SINLEN - 1)] - wt[idx]) * frac;
      double out = s * op->env * op->tl_amp;
      op->phase += op->phinc; if (op->phase >= 1.0) op->phase -= floor(op->phase);
      op->last = out;
      return out; }
}

void dq3_opl2_render(dq3_opl2 *o, short *out, int nsamples)
{
    int n, c;
    for (n = 0; n < nsamples; n++) {
        double mix = 0.0;
        for (c = 0; c < NCH; c++) {
            ch_t *ch = &o->ch[c];
            double modout, carout, fbmod;
            if (ch->op[0].state == ENV_OFF && ch->op[1].state == ENV_OFF) continue;
            /* modulator:回授 = 上次輸出 × (fb 量) */
            fbmod = ch->fb ? ch->op[0].last * (ch->fb / 8.0) * 0.5 : 0.0;
            modout = op_eval(o, &ch->op[0], fbmod);
            if (ch->con) {                 /* additive:兩 op 各自發聲 */
                carout = op_eval(o, &ch->op[1], 0.0);
                mix += (modout + carout);
            } else {                       /* FM:carrier 被 modulator 調變 */
                carout = op_eval(o, &ch->op[1], modout * 0.5);
                mix += carout;
            }
        }
        mix *= 0.16;                       /* 降增益、留 headroom(避免多聲部相加爆掉) */
        mix = tanh(mix);                   /* 軟性飽和取代硬切峰 → 杜絕刺耳失真 */
        out[n] = (short)(mix * 28000.0);
    }
}
