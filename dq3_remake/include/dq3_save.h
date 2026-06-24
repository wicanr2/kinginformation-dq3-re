/* dq3_save.h — 存檔 / 讀檔(名冊 + 隊伍 + 道具 + 位置)。
 *
 * remake 自有格式(非原版存檔):magic "DQ3SAVE2" + 版本 + 各結構原樣寫入 + 位置。
 * 同平台 round-trip 用;跨版本以 magic/大小欄位防呆。F10 自動存檔與「繼續遊戲」共用。
 */
#ifndef DQ3_SAVE_H
#define DQ3_SAVE_H

#include "dq3_roster.h"
#include "dq3_inventory.h"

typedef struct { int cty, px, py; } dq3_save_pos;

/* 寫存檔。回 0=成功、<0=失敗。 */
int dq3_save_write(const char *path, const dq3_roster *r, const dq3_party *p,
                   const dq3_inventory *inv, dq3_save_pos pos);

/* 讀存檔到傳入結構。回 0=成功、<0=檔不存在 / 格式不符。 */
int dq3_save_read(const char *path, dq3_roster *r, dq3_party *p,
                  dq3_inventory *inv, dq3_save_pos *pos);

/* 存檔是否存在且 magic 正確。 */
int dq3_save_exists(const char *path);

#endif /* DQ3_SAVE_H */
