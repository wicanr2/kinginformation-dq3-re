/* dq3_cmdmenu.h — 野外指令窗(命令):精訊版 2×3 指令格,標籤取自 D3TXT00 rec400。
 * 對話/咒文/狀況/道具/裝備/調查(官方字串,非猜)。 */
#ifndef DQ3_CMDMENU_H
#define DQ3_CMDMENU_H

#include <stdint.h>
#include "dq3_text.h"

enum { DQ3_CMD_TALK = 0, DQ3_CMD_SPELL, DQ3_CMD_STATUS,
       DQ3_CMD_ITEM, DQ3_CMD_EQUIP, DQ3_CMD_EXAMINE, DQ3_CMD_COUNT };

typedef struct { int cursor; } dq3_cmdmenu;   /* 0..5(列優先:0對話1咒文/2狀況3道具/4裝備5調查)*/

void dq3_cmdmenu_init(dq3_cmdmenu *m);
/* 處理一鍵:方向移游標(2 欄×3 列繞回),Enter 回選定指令(0..5),其餘回 -1,ESC 回 -2。 */
int  dq3_cmdmenu_input(dq3_cmdmenu *m, uint8_t sc);
/* 在 (x,y) 繪指令窗(命令 + 6 格 + ► 游標)。fg 字色、curfg 游標色。 */
void dq3_cmdmenu_render(const dq3_cmdmenu *m, const dq3_text *t,
                        uint8_t *fb, int fb_w, int fb_h, int x, int y, uint8_t fg, uint8_t curfg);

#endif /* DQ3_CMDMENU_H */
