/* dq3_town.h — 城鎮(CTY)場景載入器。
 *
 * CTY 為多段容器(docs/04 / docs/11):檔首 u16 n + n×u16 section 偏移表;
 * section_off+0x0e → layout_ptr;版面 = <u16 w><u16 h><spawn 4B> + w*h 個 u16
 * (低 byte = BLK tile index、高 byte = 屬性)。城鎮 tile 圖庫為 DQ3N.BLK、
 * 屬性表 BLKBMN.DAT(N=1..5)。解成統一 dq3_scene 交給 scene 核心。
 */
#ifndef DQ3_TOWN_H
#define DQ3_TOWN_H

#include "dq3_scene.h"

/* 載入 CTY 城鎮的某個 section 成 scene。
 *   cty_name : 如 "CTY00.DAT"
 *   section  : section 索引(0 起)
 *   blk_n    : 城鎮 tile 圖庫/屬性表編號 N(DQ3N.BLK + BLKBMN.DAT,1..5)
 * 失敗回 NULL(err 可帶訊息)。玩家起點取版面 spawn(不可走則退而求其次)。 */
dq3_scene *dq3_town_load(const char *assets_dir, const char *cty_name,
                         int section, int blk_n, char *err, int errcap);

#endif /* DQ3_TOWN_H */
