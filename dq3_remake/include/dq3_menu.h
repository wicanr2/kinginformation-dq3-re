/* dq3_menu.h — 通用垂直選單 widget(游標導航 + glyph 標籤繪製)。
 *
 * 供露依達酒場職業選單、野外指令選單、道具/咒文選單等共用。標籤以 D3TXT00.FON
 * glyph index 陣列表示(0xffff 終止),用 dq3_text_draw_glyph 繪製(retro-cjk 16×16)。
 */
#ifndef DQ3_MENU_H
#define DQ3_MENU_H

#include <stdint.h>
#include "dq3_text.h"

#define DQ3_MENU_MAX_ITEMS  16
#define DQ3_MENU_MAX_LABEL  8     /* 每標籤最多字數 */

typedef struct {
    uint16_t label[DQ3_MENU_MAX_ITEMS][DQ3_MENU_MAX_LABEL];  /* glyph index;0xffff 終止 */
    int      n_items;
    int      cursor;
    int      x, y;          /* 繪製左上角(像素;游標符號在標籤左側)*/
    int      line_h;        /* 行高(像素;預設 18)*/
} dq3_menu;

void dq3_menu_init(dq3_menu *m, int x, int y);

/* 加一項(label=glyph 陣列,len≤DQ3_MENU_MAX_LABEL)。回項 index;滿回 -1。 */
int  dq3_menu_add(dq3_menu *m, const uint16_t *label, int len);

/* 處理方向/確定 scancode:上(0x48)/下(0x50)移游標(夾在 [0,n));
 * Enter(0x1c)→ 回選中 index;其餘 → -1。 */
int  dq3_menu_input(dq3_menu *m, uint8_t scancode);

/* 繪製選單:每項 glyph 標籤,游標項用 cursor_fg + 左側「►」標記。 */
void dq3_menu_render(const dq3_menu *m, const dq3_text *t, uint8_t *fb, int fb_w, int fb_h,
                     uint8_t fg, uint8_t cursor_fg);

#endif /* DQ3_MENU_H */
