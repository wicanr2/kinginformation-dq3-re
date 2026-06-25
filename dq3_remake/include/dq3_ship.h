/* dq3_ship.h — 船系統(#2,地表 gate 2)。
 *
 * 取船後可跨海。海 tile 辨識:overworld attr 指紋 (attr&1)&&(attr&0x20)
 * (海 0x25=BLK+0x20;山 0x07=BLK 無 0x20),RE 證據見 docs/48。
 * 純 remake 機制 + 旗標流(SHIP 里程碑 0x205);原版取船劇情走 runner,早期 build 未必接全。
 */
#ifndef DQ3_SHIP_H
#define DQ3_SHIP_H
#include "dq3_scene.h"

/* 船狀態。未登船時 (px,py,layer) = 停泊的 overworld tile 與層(地表0/下層1)。 */
typedef struct {
    int owned;          /* 已取得船(= SHIP 里程碑)*/
    int aboard;         /* 目前在船上 */
    int px, py;         /* 停泊格(overworld tile)*/
    int layer;          /* 停泊層 0=地表 1=下層 */
} dq3_ship;

/* dq3_ship_input 回傳碼。 */
enum {
    DQ3_SHIP_BLOCKED   = 0,   /* 擋住,不移動 */
    DQ3_SHIP_WALKED    = 1,   /* 步行移動 */
    DQ3_SHIP_SAILED    = 2,   /* 在船上航行至水格 */
    DQ3_SHIP_DISEMBARK = 3,   /* 上岸(船停在原水格)*/
    DQ3_SHIP_BOARD     = 4    /* 登船 */
};

void dq3_ship_init(dq3_ship *sh);

/* (tx,ty) 是否為可航行水域(海 attr 指紋;OOB→0)。 */
int dq3_ship_navigable(const dq3_scene *s, int tx, int ty);

/* overworld 專用移動(sc=方向掃描碼);honor 船規則並更新 scene/ship。layer=當前所在層。
 * 城鎮移動請仍用 dq3_scene_input(船只在 overworld)。回傳 DQ3_SHIP_*。 */
int dq3_ship_input(dq3_scene *s, dq3_ship *sh, uint8_t sc, int layer);

#endif
