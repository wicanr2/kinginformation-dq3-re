/* dq3_menu.c — 通用垂直選單 widget。 */
#include "dq3_menu.h"

void dq3_menu_init(dq3_menu *m, int x, int y)
{
    m->n_items = 0; m->cursor = 0;
    m->x = x; m->y = y; m->line_h = 18;
}

int dq3_menu_add(dq3_menu *m, const uint16_t *label, int len)
{
    int i, idx;
    if (m->n_items >= DQ3_MENU_MAX_ITEMS) return -1;
    if (len > DQ3_MENU_MAX_LABEL) len = DQ3_MENU_MAX_LABEL;
    idx = m->n_items++;
    for (i = 0; i < DQ3_MENU_MAX_LABEL; i++)
        m->label[idx][i] = (i < len && label) ? label[i] : 0xffff;
    return idx;
}

int dq3_menu_input(dq3_menu *m, uint8_t sc)
{
    if (m->n_items <= 0) return -1;
    switch (sc) {
        case 0x48: if (m->cursor > 0) m->cursor--; else m->cursor = m->n_items - 1; return -1; /* 上(繞回)*/
        case 0x50: if (m->cursor < m->n_items - 1) m->cursor++; else m->cursor = 0; return -1; /* 下(繞回)*/
        case 0x1c: return m->cursor;     /* Enter → 選定 */
        default:   return -1;
    }
}

void dq3_menu_render(const dq3_menu *m, const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                     uint8_t fg, uint8_t cursor_fg)
{
    int i, j;
    for (i = 0; i < m->n_items; i++) {
        int ly = m->y + i * m->line_h;
        uint8_t col = (i == m->cursor) ? cursor_fg : fg;
        /* 游標符號:選中項在標籤左側畫一個實心三角(►)*/
        if (i == m->cursor) {
            int r, c;
            for (r = 4; r < 12; r++)
                for (c = 0; c < (r < 8 ? r - 3 : 12 - r); c++) {
                    int xx = m->x + 2 + c, yy = ly + r;
                    if (xx >= 0 && xx < fb_w && yy >= 0 && yy < fb_h) fb[yy * fb_w + xx] = cursor_fg;
                }
        }
        for (j = 0; j < DQ3_MENU_MAX_LABEL; j++) {
            uint16_t g = m->label[i][j];
            if (g == 0xffff) break;
            dq3_text_draw_glyph(t, fb, fb_w, fb_h, m->x + 16 + j * DQ3_GLYPH_PX, ly, g, col);
        }
    }
}
