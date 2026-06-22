/* states.h — 主迴圈狀態機跳表與指令 handler 宣告(real-mode large-model C)
 *
 * 對應反組譯:主迴圈 sub_93e3 的 scancode 跳表(DS:0x19 / DS:0x25)與各 handler。
 * 詳見 docs/10-exe-states.md。
 *
 * 依賴 exe.h 的型別與 runtime wrapper(exe.h 唯讀,不在此重複定義它已有的東西)。
 * 本檔只放「狀態機 / 指令選單」這條 vertical slice 自己的宣告。
 */
#ifndef DQ3_STATES_H
#define DQ3_STATES_H

#include "exe.h"

/* ---- DOS Set-1 scancode(跳表用,AH 來自 kbd_read_scancode)---- */
#define SC_Q     0x10   /* 開指令選單 */
#define SC_S     0x1f   /* 開指令選單(同 Q) */
#define SC_SPACE 0x39   /* 快速互動 / 調べる */
#define SC_ENTER 0x1c   /* 確認 / 對話推進 */
#define SC_F1    0x3b   /* 開指令選單(同 Q) */
#define SC_F4    0x3e   /* 已註冊、無 handler */
#define SC_F5    0x3f   /* 資訊 / 狀態視窗 */
#define SC_F6    0x40   /* 存檔 / 載入槽 */
#define SC_F7    0x41   /* 已註冊、無 handler */
#define SC_F8    0x42   /* 已註冊、無 handler */
#define SC_F9    0x43   /* no-op stub */
#define SC_TERM  0xff   /* scancode 表終止碼 */

/* ---- 主迴圈狀態機跳表(DS 內;固定 11 項 + 0xff 終止)----
 * 主迴圈以 cx=14 固定上界線性掃 g_cmd_scancode[],命中後 shl bx,1 取
 * g_cmd_handler[] 的 near 指標;為 0 則跳過(已註冊但無動作)。
 * exe.h 已宣告 g_cmd_scancode[] / g_cmd_handler[](DS:0x19 / DS:0x25)。 */

/* ---- 指令選單二級派發跳表(DS:0x3baa,12 項 near 指標)----
 * h_cmd_menu 以 (g_menu_sel-1) 索引,呼叫對應 field 指令動作。 */
#define DQ3_CMD_MENU_LEN 12
extern void (*g_cmd_menu_disp[DQ3_CMD_MENU_LEN])(void); /* DS:0x3baa */

/* ---- 狀態機相關全域(主資料段 DS=0x14dd)---- */
extern i16 g_menu_sel;       /* DS:0x722  選單 / 槽位選中序號(1-based) */
extern i16 g_action_result;  /* DS:0x726  動作結果 / 取消碼(1/2=取消) */
extern i16 g_modal_busy;     /* DS:0x72a  選單 modal 忙碌旗標 */
extern u8  g_info_mode;      /* DS:0x1f0  資訊視窗模式 F5=0 / F6=1 */
extern u8  g_action_active;  /* DS:0x1f7  Enter 動作執行中旗標 */
extern u8  g_party_count;    /* DS:0x5077 隊伍人數 */
extern i16 g_front_obj;      /* DS:0x2511 玩家前方 tile / 物件索引 */
extern u8  g_mode_flag;      /* DS:0x4f3b 模式旗標(>=2 Space/Enter 改路) */
extern u16 g_input_flags;    /* DS:0x000f 輸入 / 修飾鍵旗標(test &0x18 / &1) */

/* ---- 14-entry 狀態機 handler(near,seg0)---- */
void h_cmd_menu(void);   /* sub_9842  Q/S/F1 指令選單(開窗 + 二級派發) */
void h_quick_act(void);  /* sub_7c83  Space 快速互動 / 調べる */
void h_confirm(void);    /* sub_7c43  Enter 確認 / 對話推進 */
void h_info_win(void);   /* sub_13a9  F5 資訊 / 狀態視窗(唯讀) */
void h_save_slot(void);  /* sub_14da  F6 存檔 / 載入槽選擇 */
void h_f9_noop(void);    /* sub_e915  F9 no-op stub */

/* 指令選單二級派發中已就地反編譯的前 5 個(0xabfd 群組)+ 道具清單 */
void cmd_act0(void);     /* sub_988d */
void cmd_act1(void);     /* sub_98b9 */
void cmd_act2(void);     /* sub_98cd */
void cmd_act3(void);     /* sub_98e6 */
void cmd_act4(void);     /* sub_9900  重入主流程 */
void cmd_item_list(void);/* sub_504e  道具 / 裝備清單 */

/* ---- forward-declare:核心遊戲函式(屬另一個 agent,本檔只呼叫不反編譯)---- */
/* (draw_scene/update_player/update_hud/update_sub/frame_wait 已在 exe.h) */
void save_write_slot(void);  /* sub_28ce  F6 存檔寫入(改檔名數字位 + DOS 開讀關) */
void save_pack_buffer(void); /* sub_2902  組裝存檔緩衝([0x4f15]→[0x613]) */
void info_win_build(void);   /* sub_29d0  F5/F6 共用視窗矩形建立 */

/* ---- 本 slice 需要、但 exe.h 未提供的輔助宣告 ----
 * exe.h 只有 mouse_op(u16);Space handler 還用到 int33h ax=3/4 存還原座標,
 * 以具名 wrapper 表達(實作對應 sub-call,exe.h 唯讀故宣告於此)。 */
void mouse_get(u16 *x, u16 *y);  /* int 33h ax=3 → 游標座標 */
void mouse_set(u16 x, u16 y);    /* int 33h ax=4 ← 設游標座標 */

/* 直接以 DS off 存取主資料段全域(反編譯保留原始位址,real-mode near 指標)。
 * 注意:real-mode large model 下這些是 DS:off;此處以 near 指標巨集表達邏輯位址,
 * 對照反組譯方便,非位元精準(逐函式對 DOSBox 驗證為後續步驟)。 */
#define g_word(off) (*(u16 *)(off))
#define g_byte(off) (*(u8  *)(off))

#endif /* DQ3_STATES_H */
