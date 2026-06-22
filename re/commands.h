/* commands.h — 野外指令子系統 + 對話流程宣告(real-mode large-model C)
 *
 * 對應反組譯:
 *   - 指令選單二級派發表 DS:0x3baa(12 項 near 指標,由 h_cmd_menu 以 [0x722]-1 索引)
 *   - 各 field 指令 handler(sub_988d..sub_51e9)與其子動作(sub_16dd / sub_1a4c / sub_7c50)
 *   - Enter 對話 / 事件流程 worker chain(sub_9cd6 / sub_60c6 / sub_5b2d → sub_7c50 包裝)
 *   - 文字繪製器 111b:0264 的控制碼語意(對照 docs/03)
 * 詳見 docs/12-exe-commands.md。
 *
 * 依賴 exe.h(唯讀)與 states.h 已宣告的型別 / 變數;本檔只放「野外指令 + 對話」
 * 這條 vertical slice 自己新增的宣告,不重複 exe.h / states.h 已有者。
 *
 * 工具:tools/re_cmd_dis.py(docker capstone)反組譯本 slice 涉及的所有位址。
 */
#ifndef DQ3_COMMANDS_H
#define DQ3_COMMANDS_H

#include "exe.h"

/* ============================================================
 * 文字 / 對話繪製器 111b:0264 控制碼(對照 docs/03 與反組譯佐證)
 * text_draw(msg_id):DI=msg_id;由遠段 [0x252e] 取 2-byte 字碼序列逐字繪製。
 *   每碼為 glyph index(<0xffed)或下列控制碼之一。
 * ============================================================ */
#define TXT_END        0xffff  /* -1  記錄結束(record terminator) */
#define TXT_NEWLINE    0xfffe  /* -2  換行(dx+=0x10,line++;line==3 觸發換頁等待) */
#define TXT_NEWLINE2   0xfffd  /* -3  換行變體(同 -2 路徑) */
#define TXT_PAGEFEED   0xfffc  /* -4  換頁 / 段落分隔(等待按鍵 kbd 1104:123 後續頁) */
#define TXT_VAR_NAME   0xfffb  /* -5  動態插值:主角 / 角色名 */
#define TXT_VAR_ITEM   0xfffa  /* -6  動態插值:道具名 */
#define TXT_VAR_NUM    0xfff9  /* -7  動態插值:金額 / 數值 */
#define TXT_VAR_B      0xfff5  /* -0xb 動態插值(變數 B) */
#define TXT_VAR_A      0xfff6  /* -0xa 動態插值(變數 A) */
#define TXT_VAR_ED     0xffed  /* -0x13 動態插值 / 控制(最低控制碼) */

/* 文字繪製器狀態(主資料段 DS=0x14dd) */
extern u16 g_txt_seg;        /* DS:0x252e  字碼序列來源遠段(選單=D3MNS / 對話=D3TXT) */
extern u8  g_txt_line;       /* DS:0x259b  視窗內當前行號(0..3;==3 換頁) */
extern u8  g_txt_nowait;     /* DS:0x259e  1=自動推進不等鍵(系統訊息);0=逐頁等鍵 */
extern u16 g_txt_x;          /* DS:0x3e70  視窗左上 X(繪字起點 x+2) */
extern u16 g_txt_y;          /* DS:0x3e72  視窗左上 Y(繪字起點 y+0x10) */
extern u16 g_txt_varnum;     /* DS:0x2593  動態插值數值暫存(VAR_NUM 等用) */

void text_draw(u16 msg_id);  /* lcall 111b:0264 DI=msg_id */

/* ============================================================
 * 野外指令子系統:DS:0x3baa 12 項二級派發表
 *   h_cmd_menu(states.c)選定指令後 call g_cmd_menu_disp[sel-1]()。
 *   sel(=[0x722])1-based;以下為各 entry 反編譯。
 * ============================================================ */

/* idx 0..4:旗標 / 狀態 / 檔案類動作(states.c 已就地反編譯 cmd_act0..cmd_act4) */
/* idx 5  :道具 / 裝備清單(states.c 已就地反編譯 cmd_item_list / sub_504e) */

/* idx 6  sub_5112 = 話す(交談):呼叫 sub_16dd 列出對話 / 名冊頁並翻頁 */
void cmd_talk(void);         /* sub_5112 (file 0x6482) → sub_16dd */
/* idx 7  sub_5116 = 旗標查詢 + 訊息輸出(查 flag 0x29,印 0xbd1/0xbd2) */
void cmd_flag_msg(void);     /* sub_5116 (file 0x6486) */
/* idx 8  sub_512f = 子動作(sub_1a4c:移動 / 傳送 / 地圖切換相關) */
void cmd_move_act(void);     /* sub_512f (file 0x649f) → sub_1a4c */
/* idx 9  sub_5133 = 調べる / 確認(sub_7c50:Enter 對話 worker 的指令版包裝) */
void cmd_examine(void);      /* sub_5133 (file 0x64a3) → sub_7c50 */
/* idx 10 sub_5137 = 條件動作(查 flag 0x29/0x4d → 訊息 + 開資訊視窗 0x2719) */
void cmd_cond_a(void);       /* sub_5137 (file 0x64a7) */
/* idx 11 sub_51e9 = 條件動作(查 flag 0x29/0x4d → 訊息 + 開資訊視窗 0x2719) */
void cmd_cond_b(void);       /* sub_51e9 (file 0x6559) */

/* 指令選定後共用的「是 / 否 確認 + 套用」worker(多數 entry 收尾呼叫) */
void cmd_confirm_apply(void);/* sub_7baa (file 0x8f1a)  yes/no + apply + 重繪 */
void cmd_seq_msg(void);      /* sub_7c0c (file 0x8f9c)  逐段動畫訊息序列 */
void cmd_seq_msg2(void);     /* sub_7bbe (file 0x8f2e)  訊息序列變體 */
void cmd_msg_reset(void);    /* sub_6372 (file 0x76e2)  清 / 重置訊息視窗 */
void cmd_msg_loop(void);     /* sub_6380 (file 0x76f0)  訊息 + 回主選單迴圈 */

/* ============================================================
 * 子動作目標(指令派發呼叫的未具名子函式)
 * ============================================================ */
void talk_page(void);        /* sub_16dd (file 0x2a4d)  列名冊 / 對話頁(lcall 111b:779/43e/214) */
void move_worker(void);      /* sub_1a4c (file 0x2dbc)  地圖移動 / 座標 + tile 屬性解算 */

/* ============================================================
 * 對話 / 事件流程:Enter(h_confirm)的多段 worker chain
 *   h_confirm(states.c)→ confirm_worker → (stage2) → (stage3)。
 *   sub_7c50 是同一條鏈的「指令選單 idx9(調べる)」入口包裝。
 * ============================================================ */
void examine_chain(void);    /* sub_7c50 (file 0x8fc0)  包裝:呼叫 9cd6→60c6→5b2d 並依 [0x726] 續段 */
/* confirm_worker(sub_9cd6)/ confirm_stage2(sub_60c6)/ confirm_stage3(sub_5b2d)
 * 已在 states.h forward-declare(h_confirm 用);本檔補上對話流程語意註解,不重宣告。 */

/* ---- 對話 / 事件流程相關全域(主資料段 DS=0x14dd)---- */
extern i16 g_front_tile;     /* DS:0x2511 玩家前方 tile / 物件索引(=states.h g_front_obj) */
extern u16 g_event_hit;      /* DS:0x2702 本次互動命中事件計數(0=無事可做) */
extern u16 g_obj_loop;       /* DS:0x259c 物件 / 隊伍掃描迴圈索引 */
extern u16 g_sel_obj;        /* DS:0x24a5 選定物件 / 對象指標(視窗導航結果) */

/* 物件 / 事件表(near,主資料段內偏移) */
#define OBJ_TABLE      0x37c4   /* [front_tile*3 + 0x37c4] = 前方物件事件碼(0xff=無) */

/* ---- 旗標讀寫(遊戲進度旗標,bit 表)---- */
/* states.h 已宣告 game_flag_get(0x8279)/ game_flag_set(0x8264);
 * 本 slice 另用到的旗標索引(由各 handler 反組譯歸納): */
#define FLAG_CMD_AVAIL   0x29   /* 指令 / 場景可用性閘(多數 field 指令先查此旗標) */
#define FLAG_COND_B      0x4d   /* 條件動作第二旗標(cmd_cond_a/b 查) */
#define FLAG_EVENT_2A    0x2a   /* 事件旗標(cmd_cond_b) */
#define FLAG_EVENT_2C    0x2c   /* 事件旗標 */

void game_flag_set2(u16 idx); /* sub_824f  旗標設定(另一變體,worker chain 用) */

#endif /* DQ3_COMMANDS_H */
