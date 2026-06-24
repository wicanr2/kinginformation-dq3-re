/* dq3_status.c — つよさ 能力值畫面實作。 */
#include "dq3_status.h"
#include "dq3_stats.h"
#include "dq3_spell.h"

/* 標籤 glyph(D3TXT00.FON 索引);順序對 dq3_stat_kind。 */
static const uint16_t L_LV[2]   = {419, 420};   /* 等級 */
static const uint16_t L_HP[2]   = {22, 30};     /* HP */
static const uint16_t L_MP[2]   = {27, 30};     /* MP */
static const uint16_t L_STR[2]  = {170, 280};   /* 力量 */
static const uint16_t L_VIT[2]  = {284, 170};   /* 耐力 */
static const uint16_t L_AGI[2]  = {282, 283};   /* 速度 */
static const uint16_t L_INT[2]  = {285, 286};   /* 聰明 */
static const uint16_t L_LUCK[2] = {276, 426};   /* 運氣 */
static const uint16_t L_ATK[2]  = {623, 624};   /* 攻擊 */
static const uint16_t L_DEF[2]  = {340, 409};   /* 守備 */
static const uint16_t L_EXP[2]  = {525, 526};   /* 經驗 */
static const uint16_t L_SEX[2]  = {674, 417};   /* 性別 */

static void draw_glyphs(const dq3_text *t, uint8_t *fb, int fw, int fh,
                        int x, int y, const uint16_t *g, int n, uint8_t fg)
{
    int i;
    for (i = 0; i < n; i++)
        dq3_text_draw_glyph(t, fb, fw, fh, x + i * DQ3_GLYPH_PX, y, g[i], fg);
}

/* 畫十進位數(右側不補零)。回傳畫了幾位。 */
static int draw_num(const dq3_text *t, uint8_t *fb, int fw, int fh,
                    int x, int y, long v, uint8_t fg)
{
    uint16_t tmp[10]; int n = 0, i;
    if (v < 0) v = 0;
    if (v == 0) tmp[n++] = 0;
    else { long x2 = v; while (x2 > 0 && n < 10) { tmp[n++] = (uint16_t)(x2 % 10); x2 /= 10; } }
    for (i = 0; i < n; i++)
        dq3_text_draw_glyph(t, fb, fw, fh, x + i * DQ3_GLYPH_PX, y, tmp[n - 1 - i], fg);
    return n;
}

/* 一列:label(n_lab glyph)+ 值 */
static void row(const dq3_text *t, uint8_t *fb, int fw, int fh, int x, int y,
                const uint16_t *lab, int n_lab, int val, uint8_t fg)
{
    draw_glyphs(t, fb, fw, fh, x, y, lab, n_lab, fg);
    draw_num(t, fb, fw, fh, x + (n_lab + 1) * DQ3_GLYPH_PX, y, val, fg);
}

void dq3_status_render(const dq3_recruit *rc, const dq3_text *t,
                       uint8_t *fb, int fb_w, int fb_h, int x, int y, uint8_t fg)
{
    /* 版面對齊 D3TXT00 rec407(精訊版狀態畫面):上=姓名+職業;
     * 左欄 性別/等級/HP/MP;右欄 力量/速度/耐力/聰明/運氣/攻擊/守備/經驗。
     * 攻擊力=力量(無武器)、守備力=耐力/2(無防具),與戰鬥衍生一致。 */
    const dq3_member *m = &rc->m;
    const int H = 18, RX = x + 9 * DQ3_GLYPH_PX;   /* 右欄 x */
    int yy = y, ry;

    /* 姓名 + 職業 */
    draw_glyphs(t, fb, fb_w, fb_h, x, yy, rc->name, rc->name_len, fg);
    {
        const struct dq3_class_name *cn = &dq3_class_names[m->cls];
        draw_glyphs(t, fb, fb_w, fb_h, x + (rc->name_len + 1) * DQ3_GLYPH_PX, yy, cn->g, cn->len, fg);
    }
    yy += H + 4;

    /* 左欄:性別 / 等級 / HP / MP */
    {
        uint16_t sg = (rc->gender == 0) ? 775 : 234;   /* 男 / 女 */
        draw_glyphs(t, fb, fb_w, fb_h, x, yy, L_SEX, 2, fg);
        dq3_text_draw_glyph(t, fb, fb_w, fb_h, x + 3 * DQ3_GLYPH_PX, yy, sg, fg);
    }
    row(t, fb, fb_w, fb_h, x, yy + H,   L_LV, 2, m->level, fg);
    row(t, fb, fb_w, fb_h, x, yy + 2*H, L_HP, 2, m->stat[DQ3_STAT_HP], fg);
    row(t, fb, fb_w, fb_h, x, yy + 3*H, L_MP, 2, m->stat[DQ3_STAT_MP], fg);

    /* 右欄:力量/速度/耐力/聰明/運氣/攻擊/守備/經驗 */
    ry = yy;
    row(t, fb, fb_w, fb_h, RX, ry,        L_STR,  2, m->stat[DQ3_STAT_STR],  fg);
    row(t, fb, fb_w, fb_h, RX, ry + H,    L_AGI,  2, m->stat[DQ3_STAT_AGI],  fg);
    row(t, fb, fb_w, fb_h, RX, ry + 2*H,  L_VIT,  2, m->stat[DQ3_STAT_VIT],  fg);
    row(t, fb, fb_w, fb_h, RX, ry + 3*H,  L_INT,  2, m->stat[DQ3_STAT_INT],  fg);
    row(t, fb, fb_w, fb_h, RX, ry + 4*H,  L_LUCK, 2, m->stat[DQ3_STAT_LUCK], fg);
    row(t, fb, fb_w, fb_h, RX, ry + 5*H,  L_ATK,  2, m->stat[DQ3_STAT_STR],      fg);  /* 攻擊=力量 */
    row(t, fb, fb_w, fb_h, RX, ry + 6*H,  L_DEF,  2, m->stat[DQ3_STAT_VIT] / 2,  fg);  /* 守備=耐力/2 */
    draw_glyphs(t, fb, fb_w, fb_h, RX, ry + 7*H, L_EXP, 2, fg);
    draw_num(t, fb, fb_w, fb_h, RX + 3 * DQ3_GLYPH_PX, ry + 7*H, (long)m->exp, fg);
}

void dq3_status_render_spells(const dq3_recruit *rc, const dq3_text *t,
                             uint8_t *fb, int fb_w, int fb_h, int x, int y, uint8_t fg)
{
    static const uint16_t L_SPELL[2] = {429, 430};   /* 咒文 */
    unsigned short recs[64];
    int n = dq3_spells_known(rc->m.cls, rc->m.level, recs, 64);
    int yy = y, i;

    /* 標題:姓名 + 咒文 */
    draw_glyphs(t, fb, fb_w, fb_h, x, yy, rc->name, rc->name_len, fg);
    draw_glyphs(t, fb, fb_w, fb_h, x + (rc->name_len + 1) * DQ3_GLYPH_PX, yy, L_SPELL, 2, fg);
    yy += 20;

    if (n == 0) {
        static const uint16_t L_NONE = 495;   /* 無 */
        dq3_text_draw_glyph(t, fb, fb_w, fb_h, x, yy, L_NONE, fg);
        return;
    }
    /* 兩欄列出咒文名(每名一筆 D3TXT00 record)*/
    for (i = 0; i < n; i++) {
        int col = i & 1, rowi = i >> 1;
        int sx = x + col * (5 * DQ3_GLYPH_PX);
        int sy = yy + rowi * 17;
        if (sy + 16 > fb_h) break;
        dq3_text_draw_record(t, fb, fb_w, fb_h, sx, sy, 5, 1, recs[i], fg);
    }
}
