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
#include <stddef.h>        /* size_t */
#include "dq3_runtime.h"   /* dq3_color */
#include "dq3_sprite.h"    /* dq3_charsprite */
#include "dq3_npc.h"       /* dq3_npc(前向宣告 dq3_scene,不循環)*/

#define DQ3_TILE_W 32
#define DQ3_TILE_H 24
#define DQ3_SCENE_MAX_NPC 32

typedef struct dq3_scene {
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
    int dlg_bank;                 /* 對話檔 bank = section header +0x17(→ D3TXT0<bank>;0=未設,docs/42)*/
    int section;                  /* 目前 section 號(設施查表 by cty/sec/k 用,docs/40)*/

    /* 事件層(docs/31):tile 高 byte = 事件 subid(屬性 attr&8 為事件格);
     * section 事件表(4-byte 項 type/param/p2,by subid)。城鎮/迷宮才有。 */
    uint8_t  *hi_map;             /* w*h,每格 tile 高 byte(owned);NULL=無事件層 */
    struct { uint8_t type; uint16_t param; uint8_t p2; } events[32];
    int       n_events;           /* 0 = 此 section 無事件表 */

    /* 轉場層(docs/35):section header +0xc 轉場表(門/階梯/出城)。
     * 4-byte 項 {destCTY, destSec, X, Y},由轉場 tile 的 subid 索引;
     * 轉場 tile 判定 = 屬性 attr&0xe000(非 attr&8 的 examine 事件)。
     * destSec==0xff → 出城回地表;destCTY!=當前 → 跨 CTY。 */
    struct { uint8_t dest_cty, dest_sec, x, y; } transitions[32];
    int       n_transitions;      /* 0 = 此 section 無轉場表 */

    /* NPC(docs/34/35 §九):CTY section +0/+2 載入的 8-byte 槽;走動見 dq3_npc。
     * 載入時 stamp 到 hi_map(0x20|slot);每幀 dq3_scene_npc_tick 跑步進。 */
    dq3_npc   npcs[DQ3_SCENE_MAX_NPC];
    int       n_npcs;
    dq3_rng   npc_rng;
    /* NPC sprite 快取(by b2 去重;entry_base=(b2-4)*4,靜態 RE file 0xffc3:BLS off=(key-4)*0xf00+6;主角 key8↔entry16)。 */
    dq3_charsprite npc_spr[8];
    int       npc_spr_b2[8];
    int       n_npc_spr;

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

/* 找 (tx,ty) 上的 NPC 槽 index;無回 -1。供「話す」面向 NPC 偵測。 */
int  dq3_scene_npc_at(const dq3_scene *s, int tx, int ty);

/* 查 (tx,ty) 的 tile 事件(docs/31:屬性 attr&8 + 高 byte subid → section 事件表)。
 * 有事件回 1 並填 *type/*param;無回 0。 */
int  dq3_scene_tile_event(const dq3_scene *s, int tx, int ty, int *type, int *param);
/* 同上,另回傳事件 p2(一次性旗標 id);寶箱開啟用。 */
int  dq3_scene_tile_event_p2(const dq3_scene *s, int tx, int ty, int *type, int *param, int *p2);

/* 查 (tx,ty) 是否為轉場 tile(docs/35:屬性 attr&0xe000 + 高 byte subid → +0xc 轉場表)。
 * 是 → 回 1 並填目的 *dest_cty/*dest_sec/*dx/*dy(任一可為 NULL);否則回 0。 */
int  dq3_scene_tile_transition(const dq3_scene *s, int tx, int ty,
                               int *dest_cty, int *dest_sec, int *dx, int *dy);

/* 鎖門(docs/35 §八):門 tile 的 attr 低 byte bits6-7 = 所需鑰匙等級(1/2/3)。
 * 回傳 (tx,ty) 門所需鑰匙等級;0 = 非鎖門。鏡射 DQ3.EXE 0x4906 `test attr,0xc0`。 */
int  dq3_scene_door_tier(const dq3_scene *s, int tx, int ty);

/* 就地開門(鏡射 0x4977:新 index = 高 byte subid、高 byte &= 0xe0 保留轉場位)。
 * 不檢查鑰匙,呼叫端先用 dq3_scene_door_tier + key_tier 判定。回 1=已開、0=非門。 */
int  dq3_scene_open_door(dq3_scene *s, int tx, int ty);

/* 試開「面向 tile」的鎖門:若為門且 key_tier >= 所需等級 → 開門回 1;否則 0。
 * 面向格 = (px,py) + facing 位移(同 dq3_scene_input 慣例)。城鎮專用(EXE 0x4906 gate)。 */
int  dq3_scene_try_open_facing_door(dq3_scene *s, int key_tier);

/* 從 CTY section NPC 清單(+0/+2,7-byte 記錄)載 NPC 進 scene 槽並 stamp 到 hi_map。
 * cty=CTY 檔內容,so=section base offset。回載入隻數。 */
int  dq3_scene_load_npcs(dq3_scene *s, const uint8_t *cty, size_t cty_len, size_t so);

/* 每幀對所有 NPC 跑一次步進(隨機走動;靜止/凍結者不動)。回本幀移動隻數。 */
int  dq3_scene_npc_tick(dq3_scene *s);

/* 載入 NPC sprite(by b2 去重,DQ3MAN.BLS entry_base=b2*4)。回快取到的 sprite 種類數。 */
int  dq3_scene_load_npc_sprites(dq3_scene *s, const char *assets_dir);

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
