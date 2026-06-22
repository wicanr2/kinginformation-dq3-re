/* dq3_field.h — 地表(overworld)場景載入器。
 *
 * 把世界地圖素材(DQ3CON.MAP + DQ3.BLK + BLKBM.DAT + DQ3.PAL)解成統一的
 * dq3_scene(index_map + tiles + attr + palette),交給 scene 核心走動/繪製。
 * 載入器是邊界 adapter;移動/碰撞/渲染邏輯在 dq3_scene。
 */
#ifndef DQ3_FIELD_H
#define DQ3_FIELD_H

#include "dq3_scene.h"

/* 載入地表場景。失敗回 NULL(err 可帶錯誤訊息,容量 errcap)。
 * 玩家起點以開闊地啟發式挑選。 */
dq3_scene *dq3_field_load(const char *assets_dir, char *err, int errcap);

#endif /* DQ3_FIELD_H */
