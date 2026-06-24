/* dq3_status.h — つよさ 能力值畫面:渲染一名隊員的完整數值表。
 * 標籤 glyph 取自 D3TXT00.FON;屬性語意三方交叉確認(見 dq3_stats.h dq3_stat_kind)。 */
#ifndef DQ3_STATUS_H
#define DQ3_STATUS_H

#include <stdint.h>
#include "dq3_roster.h"
#include "dq3_text.h"

/* 在 (x,y) 起繪一名隊員的能力值表:姓名+Lv、職業、HP/MP、力量/耐力/速度/聰明/運氣、經驗。
 * fg = 文字色 index。 */
void dq3_status_render(const dq3_recruit *rc, const dq3_text *t,
                       uint8_t *fb, int fb_w, int fb_h, int x, int y, uint8_t fg);

/* 在 (x,y) 起繪一名隊員「已學會的咒文」清單(由職業 + 等級算,DQ3.EXE 習得表)。
 * 無咒職業顯示「無」。 */
void dq3_status_render_spells(const dq3_recruit *rc, const dq3_text *t,
                             uint8_t *fb, int fb_w, int fb_h, int x, int y, uint8_t fg);

#endif /* DQ3_STATUS_H */
