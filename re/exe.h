/* exe.h — 精訊 DQ3.EXE 反編譯共用標頭(real-mode large-model C)
 *
 * 來源:tools/re_recdis.py 遞迴反組譯 + docs/05/06/07/08。
 * 目標語言:16-bit real-mode large model C(當年 DOS C 編譯器,如 Turbo C / MSC large model)。
 * runtime = 一套手寫組語驅動(VGA planar / Sound Blaster / 鍵盤 / 滑鼠 / 檔案 / BIOS 視訊),
 * 主程式以 far call 進入。此處以函式原型 + 段位址註記表達,求結構正確可審閱,非位元精準。
 *
 * 位址慣例:函式名 sub_XXXX 的 XXXX = seg0 邏輯 offset(file offset = 0x1370 + off)。
 * runtime 以 seg:off 標註(seg = paragraph,file off = 0x1370 + seg*16 + off)。
 */
#ifndef DQ3_EXE_H
#define DQ3_EXE_H

#include <stdint.h>

typedef uint8_t  u8;
typedef uint16_t u16;
typedef int16_t  i16;

/* ---- 主資料段 DS=0x14dd(file base 0x16140)的全域變數(以 DS:off 表示)----
 * 這些在真正反編譯時會成為 .data 全域;此處保留原始 DS offset 方便對照反組譯。 */
extern i16 g_player_dx;     /* DS:0x2572  本幀玩家 X 位移/座標累加 */
extern i16 g_player_dy;     /* DS:0x2574  本幀玩家 Y 位移/座標累加 */
extern u8  g_player_face;   /* DS:0x26ad  朝向 0/2/4/6 = 下/左/上/右 */
extern i16 g_move_dir;      /* DS:0x4f1f  本幀移動方向 0/1/2/3,0xff=無 */
extern i16 g_scene_flag;    /* DS:0x4f2d  0=地表 overworld,否則城鎮(也決定 BLK 檔) */
extern u16 g_status_flag;   /* DS:0x4f46  狀態旗標(sub_9530 test bit2) */
extern u16 g_map_index;     /* DS:0x256c  地圖索引(BLK 載入查號) */

/* 14-entry 主迴圈狀態機跳表(DS 內):
 *   g_cmd_scancode[]  @ DS:0x19  (14 個 scancode)
 *   g_cmd_handler[]   @ DS:0x25  (14 個 far 函式指標,0=無) */
#define DQ3_CMD_TABLE_LEN 14
extern u8  g_cmd_scancode[DQ3_CMD_TABLE_LEN];      /* DS:0x19 */
extern void (far *g_cmd_handler[DQ3_CMD_TABLE_LEN])(void); /* DS:0x25 */

/* ---- runtime 驅動(手寫組語,far call)----
 * 以 (seg:off, ah/al 子功能) 表達為具名 wrapper。實際是 lcall seg:off,AH:AL 帶子功能碼。
 * 下面只列 startup / mainloop 用到的;完整段角色見 docs/06。 */

/* seg 0x1104 = 鍵盤 / 中斷驅動 */
void  kbd_init(void);                 /* lcall 1104:0d */
void  kbd_set_mode(u8 ah, u8 al);     /* lcall 1104:132,AH:AL=子功能 */
void  kbd_hook(void);                 /* lcall 1104:2e */
u8    kbd_read_scancode(void);        /* lcall 1104:8a → AH=scancode(0=無) */

/* seg 0x1053 = 檔案 I/O 庫 */
void  file_init(void);                /* lcall 1053:0c */
void  file_op_dx(u16 dx);             /* lcall 1053:154 / 172 …(DX 參數) */

/* seg 0x109c / 0x1209 = BIOS 視訊 / 圖形 */
void  vid_mode(u8 sub);               /* lcall 109c:xx  設模式 */
void  gfx_init(u16 sub);              /* lcall 1209:xx */

/* seg 0x10b6 = 滑鼠 + 雜項 */
void  mouse_op(u16 sub);              /* lcall 10b6:xx */

/* seg 0x11dd = 滑鼠 + 檔案(選單 0x3e 鍵入口) */
void  menu_special(void);            /* lcall 11dd:0c */

/* seg 0x129c = Sound Blaster 音效;seg 0x111b = VGA planar 繪圖(含文字繪製器 111b:0264) */

/* ---- 主程式函式(seg0,sub_XXXX)---- */

/* startup.c */
void cstartup(void);          /* sub_0000  C runtime startup */
void main_init(void);         /* sub_0030  主初始化(預設座標/視窗常數) */
void entry(void);             /* sub_9299  程式入口(對應 CS:IP) */
void game_body(void);         /* sub_92f0  遊戲主體(init + 主迴圈,return 後結束) */
void video_file_init(void);   /* sub_934f  影片/檔案/字型載入 */
void gfx_palette_init(void);  /* sub_9387  圖形 + 調色盤 + BLK 初始化 */
void seed_and_names(void);    /* sub_6f4b (file 0x82bb) 亂數種子 + 玩家名稱 */

/* mainloop.c */
void main_loop(void);         /* sub_93e3  主迴圈(輸入→移動→狀態機→繪製→節拍) */
void auto_step(void);         /* sub_94c3  自動移動一步(用 g_move_dir 重播) */
void draw_scene(void);        /* sub_255b  繪場景(地圖捲動) */
void update_player(void);     /* sub_19b8  玩家精靈/動畫 + 移動 jump table + 碰撞 */
void update_hud(void);        /* sub_9530  狀態/HUD */
void update_sub(void);        /* sub_991d  子系統更新 */
void frame_wait(void);        /* sub_ee23  計時/節拍(frame wait) */

/* 素材載入(檔名池 → 載入函式,見 docs/08) */
void load_item(void);         /* sub_17cb item.dat */
void load_menu_data(void);    /* sub_17ee d3mns.dat */
void load_text(void);         /* sub_1811 d3txt00.fon/.txt */
void load_setup(void);        /* sub_997d setup.dat */
void load_blk(void);          /* sub_eaf4 dq3.blk/dq31.blk(BLK tile 載入器) */
void load_palette(void);      /* sub_ecdc dq3.pal/mnsbk.pal */

#endif /* DQ3_EXE_H */
