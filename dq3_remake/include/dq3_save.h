/* dq3_save.h — 存檔 / 讀檔(名冊 + 隊伍 + 道具 + 位置)。
 *
 * remake 自有格式(非原版存檔):magic "DQ3SAVE2" + 版本 + 各結構原樣寫入 + 位置。
 * 同平台 round-trip 用;跨版本以 magic/大小欄位防呆。F10 自動存檔與「繼續遊戲」共用。
 */
#ifndef DQ3_SAVE_H
#define DQ3_SAVE_H

#include "dq3_roster.h"
#include "dq3_inventory.h"

/* 世界位置 + 船狀態(#2,docs/48)。ship_* 折入此結構以維持 write/read 窄介面。
 * in_town/layer/sec:讀檔後完整還原玩家所在場景(隨時存讀檔回原位置,使用者需求)。
 *   in_town=1 → 城鎮 CTY=cty 的 section=sec、(px,py);in_town=0 → overworld,layer(0=地表 1=下層)+ (px,py)。 */
typedef struct {
    int cty, px, py;
    int ship_owned, ship_aboard, ship_px, ship_py, ship_layer;
    int in_town, layer, sec;
    int daynight;   /* v6:晝夜相位 0白天 1黃昏 2黑夜 3黎明 */
    /* v8:魯拉/蓋美拉去過的城鎮(傳送目的)。存讀檔持久。 */
    int n_visited;
    struct { int cty, x, y, sec; } visited[16];
} dq3_save_pos;

/* 寫存檔(v6 加 storyflags 劇情進度)。回 0=成功、<0=失敗。 */
int dq3_save_write(const char *path, const dq3_roster *r, const dq3_party *p,
                   const dq3_inventory *inv, const dq3_storyflags *flags, dq3_save_pos pos);

/* 讀存檔到傳入結構。回 0=成功、<0=檔不存在 / 格式不符。 */
int dq3_save_read(const char *path, dq3_roster *r, dq3_party *p,
                  dq3_inventory *inv, dq3_storyflags *flags, dq3_save_pos *pos);

/* 存檔是否存在且 magic 正確。 */
int dq3_save_exists(const char *path);

#endif /* DQ3_SAVE_H */
