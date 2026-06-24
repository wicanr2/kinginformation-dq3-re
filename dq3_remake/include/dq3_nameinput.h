/* dq3_nameinput.h — 姓名輸入(英數模式)widget。
 *
 * 對齊 DQ3 精訊版角色建立 modal 的英數鍵盤(docs/15,rec 453:0-9 / A-Z / OK)。
 * 英數 glyph 對映(D3TXT00.FON):數字 d → glyph d(0-9);字母 A..Z → glyph 15..40。
 * 注音組字模式(ni_zhuyin / compose,docs/15)為更深的後續,未在此 widget 實作。
 *
 * 格盤:0-9 + A-Z = 36 字,row-major(預設 9 欄),末尾兩特殊格 ←(退格)/ OK(完成)。
 * 方向鍵移游標、Enter 選格;選字 → 附加到名字緩衝(上限 DQ3_NAME_MAX)。
 */
#ifndef DQ3_NAMEINPUT_H
#define DQ3_NAMEINPUT_H

#include <stdint.h>
#include "dq3_text.h"

#define DQ3_NI_COLS      9
#define DQ3_NI_CHARS     36                 /* 0-9 + A-Z */
#define DQ3_NI_CELLS     (DQ3_NI_CHARS + 2) /* + ← + OK */
#define DQ3_NI_CELL_BS   (DQ3_NI_CHARS)     /* 退格格 index */
#define DQ3_NI_CELL_OK   (DQ3_NI_CHARS + 1) /* 完成格 index */
#define DQ3_NI_NAME_MAX  4

typedef struct {
    uint16_t buf[DQ3_NI_NAME_MAX];   /* 已選 glyph */
    int      len;
    int      cursor;                 /* 0..DQ3_NI_CELLS-1 */
    int      done;                   /* 1 = 按了 OK */
} dq3_nameinput;

void dq3_nameinput_init(dq3_nameinput *ni);

/* cell index → glyph(字元格)。特殊格(←/OK)回 0xffff。 */
uint16_t dq3_nameinput_cell_glyph(int cell);

/* 處理一個 scancode(上0x48/下0x50/左0x4b/右0x4d/Enter0x1c)。
 * Enter 選字元 → 附加;選 ← → 退格;選 OK → 設 done=1。回 1=有變化。 */
int  dq3_nameinput_input(dq3_nameinput *ni, uint8_t scancode);

/* 繪製:名字緩衝列 + 字盤(游標高亮)。 */
void dq3_nameinput_render(const dq3_nameinput *ni, const dq3_text *t,
                          uint8_t *fb, int fb_w, int fb_h, int x, int y,
                          uint8_t fg, uint8_t curfg);

#endif /* DQ3_NAMEINPUT_H */
