/* dq3_cmdmenu.c — 野外指令窗實作。標籤 glyph 取自 D3TXT00 rec400(官方字串)。 */
#include "dq3_cmdmenu.h"

/* 6 指令 2-glyph 標籤(rec400 驗證):對話/咒文/狀況/道具/裝備/調查 */
static const uint16_t LABEL[DQ3_CMD_COUNT][2] = {
    {443, 449},   /* 對話 */
    {429, 430},   /* 咒文 */
    {665, 666},   /* 狀況 */
    {402, 237},   /* 道具 */
    {197, 409},   /* 裝備 */
    {544, 545},   /* 調查 */
};
void dq3_cmdmenu_init(dq3_cmdmenu *m) { m->cursor = 0; }

int dq3_cmdmenu_input(dq3_cmdmenu *m, uint8_t sc)
{
    int r = m->cursor >> 1, c = m->cursor & 1;   /* 列優先:cell = r*2 + c */
    switch (sc) {
    case 0x48: r = (r + 3 - 1) % 3; break;        /* 上 */
    case 0x50: r = (r + 1) % 3; break;            /* 下 */
    case 0x4b: c = (c + 1) & 1; break;            /* 左 ↔ 右(2 欄繞回)*/
    case 0x4d: c = (c + 1) & 1; break;            /* 右 */
    case 0x1c: return m->cursor;                  /* Enter:選定 */
    case 0x01: return -2;                         /* ESC:取消 */
    default: return -1;
    }
    m->cursor = r * 2 + c;
    return -1;
}

void dq3_cmdmenu_render(const dq3_cmdmenu *m, const dq3_text *t,
                        uint8_t *fb, int fb_w, int fb_h, int x, int y, uint8_t fg, uint8_t curfg)
{
    int i;
    const int CW = 5 * DQ3_GLYPH_PX;   /* 每欄寬(► + 2 字 + 間距)*/
    const int RH = 18;
    /* 標題「命令」:命=274? 用 rec400 的 0x112/0x298。直接以 glyph 畫。 */
    dq3_text_draw_glyph(t, fb, fb_w, fb_h, x + DQ3_GLYPH_PX, y, 274, fg);     /* 命 (rec400 0x112) */
    dq3_text_draw_glyph(t, fb, fb_w, fb_h, x + 2 * DQ3_GLYPH_PX, y, 664, fg); /* 令 (rec400 0x298) */
    for (i = 0; i < DQ3_CMD_COUNT; i++) {
        int r = i >> 1, c = i & 1;
        int cx = x + c * CW, cy = y + RH + r * RH;
        if (i == m->cursor)   /* ► 游標 */
            dq3_text_draw_glyph(t, fb, fb_w, fb_h, cx, cy, 11, curfg);   /* glyph 11 = ▶ */
        dq3_text_draw_glyph(t, fb, fb_w, fb_h, cx + DQ3_GLYPH_PX,     cy, LABEL[i][0], fg);
        dq3_text_draw_glyph(t, fb, fb_w, fb_h, cx + 2 * DQ3_GLYPH_PX, cy, LABEL[i][1], fg);
    }
}
