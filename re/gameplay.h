/* gameplay.h — 核心遊戲玩法函式宣告(場景繪製 / 玩家移動碰撞 / HUD)
 *
 * 反編譯來源:tools/re_disasm.py(capstone,docker 內)反組譯 seg0 各函式;
 * 對照 docs/11-exe-gameplay.md。real-mode large-model C 風格。
 *
 * 位址慣例:sub_XXXX 的 XXXX = seg0 邏輯 offset(file off = 0x1370 + off)。
 * 全域變數沿用主資料段 DS=0x14dd 的 offset 標註(file base 0x16140)。
 *
 * 本檔只放本任務(render.c / player.c)自宣告的型別與全域;runtime far-call
 * wrapper 與既有素材載入請 include exe.h / assets.h(唯讀,不在此重複)。
 */
#ifndef DQ3_GAMEPLAY_H
#define DQ3_GAMEPLAY_H

#include <stdint.h>

#ifndef DQ3_EXE_H            /* 若未先 include exe.h,補基本型別 */
typedef uint8_t  u8;
typedef uint16_t u16;
typedef int16_t  i16;
#endif
typedef int8_t   i8;

/* real-mode far 指標構造(Turbo C / MSC 的 MK_FP 等價;移植時可改 huge ptr)。
 * 反編譯保留 seg:off 語意,實際重編譯時換成主機定址。 */
#ifndef MK_FP
#define MK_FP(seg, off) ((void far *)(((uint32_t)(uint16_t)(seg) << 16) | (uint16_t)(off)))
#endif
#ifndef far
#define far  /* 移植到 32/64-bit 主機時 far 視為無修飾;real-mode 編譯時為關鍵字 */
#endif

/* ===== 場景旗標 [0x4f2d](= exe.h 的 g_scene_flag / assets.h 的 is_town_flag)=====
 * 0 = 地表 overworld(讀 dq3con/und.map)
 * 1 = 城鎮 town(讀 cty 段,seg_cty)
 * 2 = 地表→城鎮過場(本幀切換,sub_255b 收尾)
 * 3 = 城鎮→地表過場(本幀切換,sub_19b8 收尾) */
#define SCENE_OVERWORLD     0
#define SCENE_TOWN          1
#define SCENE_TO_TOWN       2
#define SCENE_TO_OVERWORLD  3

/* ===== 移動方向 [0x4f1f](= exe.h 的 g_move_dir),0xff=無 ===== */
/* DIR_DOWN/LEFT/UP/RIGHT 0/1/2/3 已在 mainloop.c 定義;此處只補 NONE */
#ifndef DIR_NONE
#define DIR_NONE 0xff
#endif

/* ===== CTY 城鎮場景在記憶體中的版面(seg_cty = DS:0x2536)=====
 * CTY 檔整檔載入後就地解析(不另解壓到別的緩衝)。檔首為 section 偏移表
 * (docs/04 已解);本任務新解:section 內部版面與 per-tile 格式。
 *
 * 由 [0x256a](目前 section index)選一段:section_off = sect_table[idx]。
 * section 內偏移(皆相對 section_off,以 cty_seg 為基底段):
 *   +0x00  ... 0x0d   section 標頭(0xe bytes;含旗標、雜項,部分未逐欄定名)
 *   +0x0e  u16        row/layout 指標(指向版面資料起點,sub_30cf 用)
 *   +0x10  u16        旗標 word(bit15 → 設場景狀態旗標 [0x4f46] bit1)
 *   +0x15  4 bytes    出口/warp 參數([0xb56/0xb57/0xb58/0xd71])
 * 版面資料起點(layout_ptr 指到的位置):
 *   u16 width         城鎮寬(tile 數)→ [0xb28]
 *   u16 height        城鎮高(tile 數)→ [0xb2a]
 *   u8  spawn_x_lo, _hi(實際取 byte)→ 玩家進城起始 tile X [0x4f33]
 *   u8  spawn_y …                     → 起始 tile Y [0x4f35]
 *   之後 width*height 個 u16 = tile 陣列(row-major)。
 *   tile = layout_base[(y*width + x)] 為 u16:低 byte = BLK tile index,
 *         高 byte = 屬性 mask(0xc0 位元另存到 [0x2579],見碰撞)。
 *
 * 與 docs/04「CTY 容器:section 偏移表 + 各段 sub-header」呼應;先前未解的
 * 「RLE / 命令串」其實是 per-tile u16(index+屬性),非壓縮——see docs/11。 */
typedef struct {
    u16 width;          /* +0  城鎮寬(tile)*/
    u16 height;         /* +2  城鎮高(tile)*/
    u8  spawn_x;        /* +4  進城起始 X(取 byte)*/
    u8  spawn_x_hi;     /* +5 */
    u8  spawn_y;        /* +6  進城起始 Y */
    u8  spawn_y_hi;     /* +7 */
    /* +8 起:u16 tile[height][width],row-major;見上方說明 */
} cty_layout_hdr;

/* ===== 主要全域(DS=0x14dd offset;此處 forward declare 給 render/player 用)=====
 * 已在 exe.h / assets.h 宣告者不重複;以下為本任務新引用、尚未在他處宣告的。 */
extern i16 g_pos_x;        /* DS:0x4f33  玩家所在 tile X(地圖格座標,非螢幕)*/
extern i16 g_pos_y;        /* DS:0x4f35  玩家所在 tile Y */
extern u16 g_tile_cur;     /* DS:0x2576  目前(或目標)tile 的 BLK index */
extern u16 g_tile_attr;    /* DS:0x2577  目前目標 tile 的屬性 word(blkbm 查得)*/
extern u8  g_corner_attr;  /* DS:0x2579  目前 tile 高位屬性(& 0xc0)*/
extern u16 g_status_evt;   /* DS:0x4f46  狀態/事件旗標(HUD/事件 dispatcher 用)*/
extern u16 g_cty_section;  /* DS:0x256a  目前 CTY section 索引 */
extern i16 g_map_w;        /* DS:0x0b28  目前場景寬(tile)*/
extern i16 g_map_h;        /* DS:0x0b2a  目前場景高(tile)*/
extern u16 g_layout_base;  /* DS:0x0b26  tile 陣列在 seg_cty 內的 byte 偏移 */
extern u16 g_vram_off;     /* DS:0x4f09  目前繪製目的 VRAM 偏移(雙緩衝翻頁)*/

/* blkbm 屬性表:tile_attr_tab[tile_index] = 屬性 word(DS:0x308e,由 load_blk 載入)。
 * 每 tile 一個 u16:碰撞/觸發/地形旗標(見 docs/11 的位元語意)。 */
extern u16 tile_attr_tab[]; /* DS:0x308e */

/* ===== 函式(本任務反編譯)===== */

/* render.c —— 場景繪製 + 捲動 */
void draw_scene(void);        /* sub_255b  地表場景:gate [0x4f2d]==0,捲動+繪製 */
void update_player(void);     /* sub_19b8  城鎮場景:gate [0x4f2d]==1,移動+碰撞+捲動 */
void cty_select_section(void);/* sub_30cf  解析 CTY section → 版面/出口/起始座標 */

/* player.c —— 移動與碰撞 */
void player_move_collide(void); /* sub_2dda  town 路徑:算目標 tile、查屬性、決定可否移動 */
void scene_event_poll(void);    /* sub_2f01  [0x4f46]&1 觸發事件(call 0x488f)*/

/* HUD / 事件 dispatcher(render.c)*/
void update_hud(void);        /* sub_9530  狀態旗標逐位元 dispatch + NPC 更新 + HUD 重繪 */

/* 方向移動/捲動 handler(jump table 目標;forward declare,細節見 render.c)*/
void town_scroll_down(void);  /* sub_1ba2  jump[0x255a][0] */
void town_scroll_left(void);  /* sub_1c26  jump[0x255a][1] */
void town_scroll_up(void);    /* sub_1c99  jump[0x255a][2] */
void town_scroll_right(void); /* sub_1d09  jump[0x255a][3] */
void over_scroll_down(void);  /* sub_2880  jump[0x2562][0] */
void over_scroll_left(void);  /* sub_2904  jump[0x2562][1] */
void over_scroll_up(void);    /* sub_2977  jump[0x2562][2] */
void over_scroll_right(void); /* sub_29e7  jump[0x2562][3] */

/* 14-entry 主迴圈狀態機 handler 由「另一個 agent」反編譯;此處不定義,
 * 僅在需要時 forward-declare(本任務不引用其內部)。 */

#endif /* DQ3_GAMEPLAY_H */
