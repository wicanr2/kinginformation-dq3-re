/* nameinput.h — 精訊 DQ3 注音(Zhuyin)姓名輸入子系統宣告(real-mode large-model C)
 *
 * 對應反組譯:新遊戲建立主角時的「輸入姓名」畫面 —— 精訊為台灣玩家內建的注音輸入法。
 * 主控 sub_0d17(dispatcher)→ {sub_0e8e 英數 / sub_0f5b 注音} → sub_23f7(鍵盤導航迴圈)。
 * 由角色建立 modal sub_0854 透過 sub_0d17 進入。詳見 docs/15-exe-nameinput.md。
 *
 * 依賴 exe.h 的型別與 runtime wrapper(exe.h 唯讀,不在此重複)。
 * 本檔只放「注音姓名輸入」這條 vertical slice 自己的宣告。
 *
 * 求結構正確、可審閱;非位元精準。far runtime 以 wrapper 表達,
 * 暫存器式子功能碼(AH:AL / DI=msg_id)以參數/註解標出。
 */
#ifndef DQ3_NAMEINPUT_H
#define DQ3_NAMEINPUT_H

#include "exe.h"

/* ============================================================
 * DOS Set-1 scancode(注音鍵盤導航迴圈 sub_23f7 用,AH 來自 kbd 1104:8a)
 * ============================================================ */
#define NI_SC_UP     0x48   /* ↑  游標上移一列(grid index -= width) */
#define NI_SC_LEFT   0x4b   /* ←  游標左移一格(grid index--) */
#define NI_SC_RIGHT  0x4d   /* →  游標右移一格(grid index++) */
#define NI_SC_DOWN   0x50   /* ↓  游標下移一列(grid index += width) */
#define NI_SC_SPACE  0x39   /* Space 選取目前格(= Enter) */
#define NI_SC_ENTER  0x1c   /* Enter 選取目前格(= Space) */
#define NI_SC_MOUSE_MV 0x64 /* 滑鼠移動 pseudo-scancode(取消本次選取,al=1) */
#define NI_SC_MOUSE_CLK 0x65 /* 滑鼠點擊 pseudo-scancode(以滑鼠座標換算格位) */

/* ============================================================
 * dispatcher 模式旗標 g_ni_mode(DS:0x26fc)位元語意
 * sub_0d17 依此分派到三種子畫面;bit3 為離開訊號。
 * ============================================================ */
#define NI_MODE_FN    0x4    /* bit2:功能列模式(英數/前進/後退/取消/完成),走 sub_0dc8 */
#define NI_MODE_ALNUM 0x2    /* bit1:英數鍵盤模式,走 sub_0e8e;清除則為注音模式 sub_0f5b */
#define NI_MODE_EXIT  0x8    /* bit3:離開訊號(完成 / 強制結束) */

/* ============================================================
 * 注音 cell→碼 對照表的「組」編碼(DS:0x2758,每格 1 byte = (group<<5)|code)
 * sub_2623 以 group 當索引寫入組字緩衝 [0x2730+group]。
 * ============================================================ */
#define NI_GRP_INITIAL 0   /* 聲母(ㄅㄆㄇㄈ…,21 個)→ [0x2730] */
#define NI_GRP_MEDIAL  1   /* 介音(ㄧㄨㄩ,3 個)    → [0x2731] */
#define NI_GRP_FINAL   2   /* 韻母(ㄚㄛㄜ…,13 個)  → [0x2732] */
#define NI_GRP_TONE    3   /* 聲調(˙ˊˇˋ+一聲,5 個)→ [0x2733] */

/* 注音 grid 幾何(sub_0f5b 設定):5 列 × 9 欄 = 45 格 */
#define NI_GRID_W      9    /* [0x2702] 欄數 */
#define NI_GRID_N      0x2d /* [0x2704] 總格數 = 45 */
#define NI_GRID_X      0x15 /* [0x2706] grid 左上 X(字格) */
#define NI_GRID_Y      0x5e /* [0x2708] grid 左上 Y(像素) */

/* 注音 grid 內的特殊格(cell index = row*5 + col,sub_0f5b 計算後判斷) */
#define NI_CELL_SYMBOL_MAX 0x25 /* < 0x25(37):注音/聲調符號格 → 組字 */
#define NI_CELL_COMMIT_MAX 0x2a /* [0x25,0x2a):組字完成 → 查國字 + 寫名字 */
#define NI_CELL_OK    0x2a /* == 0x2a:OK(重置組字,繼續輸入) */
#define NI_CELL_TOFN  0x2b /* == 0x2b:切到功能列(set bit2) */

/* 名字長度上限(英數 / 注音兩模式皆 cap 4 字) */
#define NI_NAME_MAX   4

/* 文字繪製器訊息 ID(D3TXT00 記錄號;lcall 111b:0214 以 DI 帶入整段版面) */
#define NI_MSG_TITLE     0x1c3  /* 451:「輸入姓名」標題框 + 名字顯示列(《《 》》) */
#define NI_MSG_ALNUM     0x1c5  /* 453:英數鍵盤版面(0-9 A-Z 羅馬數字 ★ OK + 功能列) */
#define NI_MSG_ZHUYIN    0x1c4  /* 452:注音鍵盤版面(ㄅㄆㄇㄈ… + 聲調 OK + 功能列) */
#define NI_MSG_PICKCHAR  0x1c6  /* 454:「選取國字」候選同音字視窗 */
#define NI_MSG_COMPOSE   0x1c8  /* 456:「輸入注音」即時組字框 */

/* ============================================================
 * 注音姓名輸入子系統全域(主資料段 DS=0x14dd)
 * ============================================================ */
extern u16 g_ni_mode;        /* DS:0x26fc 模式旗標(NI_MODE_* 位元) */
extern u16 g_ni_cursor;      /* DS:0x26fe grid 內游標 raw index(0..[0x2704)) */
extern u16 g_ni_cell;        /* DS:0x2700 換算後 cell index(row*5+col) */
extern u16 g_ni_len;         /* DS:0x270a 已輸入名字字數(0..4) */
extern u16 g_ni_maxlen;      /* DS:0x270c 本次達到過的最大字數(游標前後界) */
extern u16 g_ni_grid_w;      /* DS:0x2702 grid 欄寬 */
extern u16 g_ni_grid_n;      /* DS:0x2704 grid 總格數 */
extern u16 g_ni_grid_x;      /* DS:0x2706 grid 左上 X */
extern u16 g_ni_grid_y;      /* DS:0x2708 grid 左上 Y */

extern u16 g_ni_name[16];    /* DS:0x2710 名字緩衝(每字 2 word:[0x2710]碼低 / [0x2712]碼高) */

extern u8  g_ni_zhu[4];      /* DS:0x2730 組字緩衝:[0]聲母 [1]介音 [2]韻母 [3]聲調 */
extern u8  g_ni_codetab[45]; /* DS:0x2758 cell→(group<<5)|code 對照表 */
extern u16 g_ni_glyph_lo;    /* DS:0x2754 查到 / 選到的國字碼(低 word) */
extern u16 g_ni_glyph_hi;    /* DS:0x2756 國字碼(高 word) */

extern u16 g_ni_mouse_clk;   /* DS:0x2850 滑鼠點擊旗標 */
extern i16 g_ni_mouse_idx;   /* DS:0x72c  滑鼠命中的 grid index(int33 ax=4 回算) */
extern u16 g_ni_win_save;    /* DS:0x2857 視窗狀態暫存(進出畫面 push/pop) */

/* ============================================================
 * 注音 IME 引擎 far 介面(seg 0x11c4 = 注音→國字 查表庫)
 * ============================================================ */
/* 11c4:0x27 — 組字查碼:in AH=聲母 AL=介音 BH=韻母 BL=聲調;
 *   out DX:AX=國字碼、BL=同音候選字數(0=查無)。 */
u16  ime_lookup(u8 initial, u8 medial, u8 final, u8 tone,
                u16 *glyph_lo, u16 *glyph_hi);  /* 回傳 BL=候選數 */
/* 11c4:0x81 — 依目前候選索引 [0x26fe] 取第 N 個同音字碼到 [DX] 緩衝。 */
void ime_get_candidate(void *buf);
/* 11c4:0x97 — 候選選定後,以選中索引算最終國字碼。 */
void ime_finalize(void);

/* ============================================================
 * 注音姓名輸入子系統函式(near,seg0)
 * ============================================================ */
void ni_dispatch(void);      /* sub_0d17 主控:依 g_ni_mode 分派三畫面,迴圈到離開 */
void ni_count_name(void);    /* sub_0da4 數名字非空字數 → g_ni_len */
void ni_fn_list(void);       /* sub_0dc8 功能列(5 列:英數/前進/後退/取消/完成) */
void ni_redraw_cursor(void); /* sub_0e55 重畫名字列游標(int33 ax=4 設滑鼠命中區) */
void ni_alnum(void);         /* sub_0e8e 英數鍵盤畫面 + 選字寫名字 */
void ni_zhuyin(void);        /* sub_0f5b 注音鍵盤畫面 + 組字 + 寫名字 */
i16  ni_grid_nav(void);      /* sub_23f7 grid 鍵盤導航迴圈(方向鍵 + Space/Enter 選格) */
void ni_compose_put(void);   /* sub_2623 把目前注音符號格放進組字緩衝(查 [0x2758]) */
u8   ni_compose_resolve(void); /* sub_2654 組字 → 查國字(可開「選取國字」候選窗) */
void ni_draw_compose(void);  /* sub_2703 在「輸入注音」框畫候選字 */

#endif /* DQ3_NAMEINPUT_H */
