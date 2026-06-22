/* dq3_field.h — 地表(overworld)場景引擎(features/field,deep module)。
 *
 * 把「世界地圖捲動 + tile 貼圖 + 玩家移動碰撞」收成一個窄介面的模組。
 * 對應原版:地圖佈局 DQ3CON.MAP(244×205 u8 tile index)、tile 圖庫 DQ3.BLK
 * (162 個 32×24 4-plane planar)、屬性表 BLKBM.DAT(u16/tile,bit0=阻擋)、
 * 調色盤 DQ3.PAL。移動/碰撞語意取自 re/player.c::player_move_collide
 * (docs/11):目標 tile = 玩家 tile + 方向位移;讀屬性 bit0 阻擋則不移動。
 *
 * 介面只暴露:建立/銷毀、輸入一個 scancode、把目前畫面繪進 framebuffer、
 * 取 palette 給 runtime 套色。tile 解碼、攝影機、碰撞全藏在實作內。
 */
#ifndef DQ3_FIELD_H
#define DQ3_FIELD_H

#include <stdint.h>
#include "dq3_runtime.h"   /* dq3_color */

typedef struct dq3_field dq3_field;

/* 載入地表所有素材(PAL/BLK/MAP/BLKBM),預解 tile。
 * map_name/blk_name 為 NULL 時用地表預設 DQ3CON.MAP / DQ3.BLK / BLKBM.DAT。
 * 失敗回 NULL(err 可帶錯誤字串緩衝,容量 errcap)。 */
dq3_field *dq3_field_create(const char *assets_dir, char *err, int errcap);
void       dq3_field_destroy(dq3_field *f);

/* 處理一個輸入 scancode(0x48/0x50/0x4b/0x4d 方向),含碰撞判定。
 * 回傳 1 表玩家有移動,0 表沒動(被擋或非方向鍵)。 */
int dq3_field_input(dq3_field *f, uint8_t scancode);

/* 把目前畫面(以玩家為中心的視窗)繪進 indexed framebuffer。 */
void dq3_field_render(dq3_field *f, uint8_t *fb, int fb_w, int fb_h);

/* 取目前 palette(已含海面藍色修正);count 回傳色數。 */
const dq3_color *dq3_field_palette(const dq3_field *f, int *count);

/* 查/設玩家 tile 座標(除錯/驗證用)。 */
void dq3_field_get_player(const dq3_field *f, int *tx, int *ty);
void dq3_field_set_player(dq3_field *f, int tx, int ty);

#endif /* DQ3_FIELD_H */
