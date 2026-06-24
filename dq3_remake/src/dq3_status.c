/* dq3_status.c — つよさ 能力值畫面實作。 */
#include "dq3_status.h"
#include "dq3_stats.h"

/* 標籤 glyph(D3TXT00.FON 索引);順序對 dq3_stat_kind。 */
static const uint16_t L_LV[2]   = {419, 420};   /* 等級 */
static const uint16_t L_HP[2]   = {22, 30};     /* HP */
static const uint16_t L_MP[2]   = {27, 30};     /* MP */
static const uint16_t L_STR[2]  = {170, 280};   /* 力量 */
static const uint16_t L_VIT[2]  = {284, 170};   /* 耐力 */
static const uint16_t L_AGI[2]  = {282, 283};   /* 速度 */
static const uint16_t L_INT[2]  = {285, 286};   /* 聰明 */
static const uint16_t L_LUCK[2] = {276, 426};   /* 運氣 */

static void draw_glyphs(const dq3_text *t, uint8_t *fb, int fw, int fh,
                        int x, int y, const uint16_t *g, int n, uint8_t fg)
{
    int i;
    for (i = 0; i < n; i++)
        dq3_text_draw_glyph(t, fb, fw, fh, x + i * DQ3_GLYPH_PX, y, g[i], fg);
}

/* 畫十進位數(右側不補零)。回傳畫了幾位。 */
static int draw_num(const dq3_text *t, uint8_t *fb, int fw, int fh,
                    int x, int y, int v, uint8_t fg)
{
    uint16_t tmp[6]; int n = 0, i;
    if (v < 0) v = 0;
    if (v == 0) tmp[n++] = 0;
    else { int x2 = v; while (x2 > 0 && n < 6) { tmp[n++] = (uint16_t)(x2 % 10); x2 /= 10; } }
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
    const dq3_member *m = &rc->m;
    int yy = y, cx;
    const int H = 18;

    /* 第一列:姓名 + 等級 */
    draw_glyphs(t, fb, fb_w, fb_h, x, yy, rc->name, rc->name_len, fg);
    cx = x + (rc->name_len + 1) * DQ3_GLYPH_PX;
    draw_glyphs(t, fb, fb_w, fb_h, cx, yy, L_LV, 2, fg);
    draw_num(t, fb, fb_w, fb_h, cx + 3 * DQ3_GLYPH_PX, yy, m->level, fg);
    yy += H;

    /* 職業 */
    {
        const struct dq3_class_name *cn = &dq3_class_names[m->cls];
        draw_glyphs(t, fb, fb_w, fb_h, x, yy, cn->g, cn->len, fg);
    }
    yy += H;

    /* HP / MP(member 存的是上限;野外滿血→現值=上限)*/
    row(t, fb, fb_w, fb_h, x, yy, L_HP, 2, m->stat[DQ3_STAT_HP], fg); yy += H;
    row(t, fb, fb_w, fb_h, x, yy, L_MP, 2, m->stat[DQ3_STAT_MP], fg); yy += H;
    /* 五項主屬性 */
    row(t, fb, fb_w, fb_h, x, yy, L_STR,  2, m->stat[DQ3_STAT_STR],  fg); yy += H;
    row(t, fb, fb_w, fb_h, x, yy, L_VIT,  2, m->stat[DQ3_STAT_VIT],  fg); yy += H;
    row(t, fb, fb_w, fb_h, x, yy, L_AGI,  2, m->stat[DQ3_STAT_AGI],  fg); yy += H;
    row(t, fb, fb_w, fb_h, x, yy, L_INT,  2, m->stat[DQ3_STAT_INT],  fg); yy += H;
    row(t, fb, fb_w, fb_h, x, yy, L_LUCK, 2, m->stat[DQ3_STAT_LUCK], fg);
}
