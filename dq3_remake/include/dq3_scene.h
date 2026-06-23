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
#include "dq3_sprite.h"    /* dq3_charsprite */

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

    /* 事件層(docs/31):tile 高 byte = 事件 subid(屬性 attr&8 為事件格);
     * section 事件表(4-byte 項 type/param/p2,by subid)。城鎮/迷宮才有。 */
    uint8_t  *hi_map;             /* w*h,每格 tile 高 byte(owned);NULL=無事件層 */
    struct { uint8_t type; uint16_t param; uint8_t p2; } events[32];
    int       n_events;           /* 0 = 此 section 無事件表 */

    /* 主角 sprite(DQ3MAN.BLS);has_hero=0 時退回佔位方框 */
    dq3_charsprite hero;
    int has_hero;
    int frame_for_facing[4];      /* facing→BLS frame 對映(可調,對齊 oracle) */

    /* 釋放用:scene 接管的原始 buffer(loader 配置,scene_free 釋放) */
    void *owned[6];
    int   nowned;
} dq3_scene;

/* tile 可走:界外不可走;屬性 bit0=阻擋。 */
int  dq3_scene_walkable(const dq3_scene *s, int tx, int ty);

/* 查 (tx,ty) 的 tile 事件(docs/31:屬性 attr&8 + 高 byte subid → section 事件表)。
 * 有事件回 1 並填 *type/*param;無回 0。 */
int  dq3_scene_tile_event(const dq3_scene *s, int tx, int ty, int *type, int *param);

/* 處理一個方向 scancode(0x48/0x50/0x4b/0x4d),含碰撞。回傳 1=有移動。 */
int  dq3_scene_input(dq3_scene *s, uint8_t scancode);

/* 把以玩家為中心的視窗繪進 indexed framebuffer。 */
void dq3_scene_render(const dq3_scene *s, uint8_t *fb, int fb_w, int fb_h);

/* 把本場景 palette 套進 runtime DAC。
 * 契約(修 bug #8,docs/28):任何場景切換 / 戰鬥返回後都必須呼叫一次,
 * 確保 DAC 還原成目的場景的 palette,不殘留前一場景(如戰鬥 MNSBK)的色盤。 */
void dq3_scene_apply_palette(const dq3_scene *s);

/* 起點啟發式:挑 9×9 視窗內可走鄰居最多的可走 tile(用於地表開闊地)。 */
void dq3_scene_pick_open_start(dq3_scene *s);

/* 載入主角 sprite(DQ3MAN.BLS entry_base 起 4 方向 frame)。
 * facing_order:facing(0下1左2上3右)→ BLS frame 的 4 元素對映(NULL=預設 0,1,2,3)。
 * 失敗時 has_hero 維持 0(退回佔位方框),回傳 <0。 */
int dq3_scene_load_hero(dq3_scene *s, const char *assets_dir, int entry_base,
                        const int facing_order[4]);

/* 釋放 scene 與其接管的 buffer。 */
void dq3_scene_free(dq3_scene *s);

#endif /* DQ3_SCENE_H */
