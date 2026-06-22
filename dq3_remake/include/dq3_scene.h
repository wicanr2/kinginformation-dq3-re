/* dq3_scene.h — 可走動場景核心(攝影機 + tile 貼圖 + 碰撞走動)。
 *
 * 地表(overworld)與城鎮(town)共用同一套「視窗貼圖 + bit0 碰撞」邏輯,
 * 差別只在資料來源(世界地圖 u8 陣列 vs CTY 容器 u16 陣列)與所用 tile 圖庫/
 * 屬性表。把共通部分收成 scene 核心(deep module),載入器(dq3_field /
 * dq3_town)只負責把素材解成統一的 index_map + tiles + attr,交給 scene。
 *
 * 統一中介表示:
 *   index_map[w*h]  每格一個 BLK tile index(u8);town 由 CTY u16 取低 byte。
 *   tiles[][24][32] 已解碼的 tile 像素(palette 索引)。
 *   attr[tile_index](u16)碰撞屬性,bit0=阻擋(對齊 re/player.c / docs/11)。
 */
#ifndef DQ3_SCENE_H
#define DQ3_SCENE_H

#include <stdint.h>
#include "dq3_runtime.h"   /* dq3_color */

#define DQ3_TILE_W 32
#define DQ3_TILE_H 24

typedef struct {
    int       map_w, map_h;
    uint8_t  *index_map;          /* w*h,每格 BLK tile index(owned) */
    uint8_t (*tiles)[DQ3_TILE_H][DQ3_TILE_W];  /* 解碼 tile(owned) */
    int       tile_count;
    const uint16_t *attr;         /* 屬性表 u16/tile;指向 owned buffer */
    int       attr_count;
    dq3_color pal[16];
    int       pal_count;

    int px, py;                   /* 玩家 tile 座標 */
    int facing;                   /* 0下 1左 2上 3右 */

    /* 釋放用:scene 接管的原始 buffer(loader 配置,scene_free 釋放) */
    void *owned[6];
    int   nowned;
} dq3_scene;

/* tile 可走:界外不可走;屬性 bit0=阻擋。 */
int  dq3_scene_walkable(const dq3_scene *s, int tx, int ty);

/* 處理一個方向 scancode(0x48/0x50/0x4b/0x4d),含碰撞。回傳 1=有移動。 */
int  dq3_scene_input(dq3_scene *s, uint8_t scancode);

/* 把以玩家為中心的視窗繪進 indexed framebuffer。 */
void dq3_scene_render(const dq3_scene *s, uint8_t *fb, int fb_w, int fb_h);

/* 起點啟發式:挑 9×9 視窗內可走鄰居最多的可走 tile(用於地表開闊地)。 */
void dq3_scene_pick_open_start(dq3_scene *s);

/* 釋放 scene 與其接管的 buffer。 */
void dq3_scene_free(dq3_scene *s);

#endif /* DQ3_SCENE_H */
