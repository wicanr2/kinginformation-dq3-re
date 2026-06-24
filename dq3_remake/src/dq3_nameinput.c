/* dq3_nameinput.c — 姓名輸入(英數)widget。 */
#include "dq3_nameinput.h"

#define ROWS ((DQ3_NI_CELLS + DQ3_NI_COLS - 1) / DQ3_NI_COLS)   /* 行數(含特殊格)*/

void dq3_nameinput_init(dq3_nameinput *ni)
{
    int i;
    ni->len = 0; ni->cursor = 0; ni->done = 0;
    for (i = 0; i < DQ3_NI_NAME_MAX; i++) ni->buf[i] = 0;
}

uint16_t dq3_nameinput_cell_glyph(int cell)
{
    if (cell < 0 || cell >= DQ3_NI_CHARS) return 0xffff;   /* 特殊格 / 越界 */
    if (cell < 10) return (uint16_t)cell;                  /* 數字 0-9 → glyph 0-9 */
    return (uint16_t)(15 + (cell - 10));                   /* 字母 A-Z → glyph 15-40 */
}

static void commit(dq3_nameinput *ni)
{
    int cell = ni->cursor;
    if (cell == DQ3_NI_CELL_BS) {                          /* 退格 */
        if (ni->len > 0) ni->len--;
    } else if (cell == DQ3_NI_CELL_OK) {                   /* 完成 */
        ni->done = 1;
    } else {                                               /* 字元 */
        uint16_t g = dq3_nameinput_cell_glyph(cell);
        if (g != 0xffff && ni->len < DQ3_NI_NAME_MAX) ni->buf[ni->len++] = g;
    }
}

int dq3_nameinput_input(dq3_nameinput *ni, uint8_t sc)
{
    int row = ni->cursor / DQ3_NI_COLS, col = ni->cursor % DQ3_NI_COLS, nc;
    switch (sc) {
        case 0x48: row = (row + ROWS - 1) % ROWS; break;   /* 上(繞回)*/
        case 0x50: row = (row + 1) % ROWS;        break;   /* 下 */
        case 0x4b: col = (col + DQ3_NI_COLS - 1) % DQ3_NI_COLS; break; /* 左 */
        case 0x4d: col = (col + 1) % DQ3_NI_COLS; break;   /* 右 */
        case 0x1c: commit(ni); return 1;                   /* Enter */
        default: return 0;
    }
    nc = row * DQ3_NI_COLS + col;
    if (nc >= DQ3_NI_CELLS) nc = DQ3_NI_CELLS - 1;         /* 末行不足欄 → 夾到最後格 */
    ni->cursor = nc;
    return 1;
}

void dq3_nameinput_render(const dq3_nameinput *ni, const dq3_text *t,
                          uint8_t *fb, int fb_w, int fb_h, int x, int y,
                          uint8_t fg, uint8_t curfg)
{
    int i;
    /* 名字緩衝列 */
    for (i = 0; i < DQ3_NI_NAME_MAX; i++) {
        uint16_t g = (i < ni->len) ? ni->buf[i] : 0xffff;
        if (g != 0xffff) dq3_text_draw_glyph(t, fb, fb_w, fb_h, x + i * DQ3_GLYPH_PX, y, g, fg);
        else { int r,c; for(r=14;r<16;r++)for(c=0;c<14;c++){ int xx=x+i*DQ3_GLYPH_PX+c,yy=y+r;  /* 底線 */
            if(xx>=0&&xx<fb_w&&yy>=0&&yy<fb_h) fb[yy*fb_w+xx]=fg; } }
    }
    /* 字盤(y 下移)*/
    for (i = 0; i < DQ3_NI_CELLS; i++) {
        int gr = i / DQ3_NI_COLS, gc = i % DQ3_NI_COLS;
        int cx = x + gc * DQ3_GLYPH_PX, cy = y + 28 + gr * 18;
        uint8_t col = (i == ni->cursor) ? curfg : fg;
        uint16_t g = dq3_nameinput_cell_glyph(i);
        if (g != 0xffff) {
            dq3_text_draw_glyph(t, fb, fb_w, fb_h, cx, cy, g, col);
        } else if (i == DQ3_NI_CELL_BS) {                  /* ← 退格:畫箭頭 */
            int r,c; for(r=6;r<10;r++)for(c=0;c<12;c++){ int xx=cx+c,yy=cy+r;
                if(xx>=0&&xx<fb_w&&yy>=0&&yy<fb_h) fb[yy*fb_w+xx]=col; }
        } else {                                           /* OK:畫實心方塊 */
            int r,c; for(r=2;r<14;r++)for(c=0;c<14;c++){ int xx=cx+c,yy=cy+r;
                if(xx>=0&&xx<fb_w&&yy>=0&&yy<fb_h) fb[yy*fb_w+xx]=col; }
        }
    }
}
