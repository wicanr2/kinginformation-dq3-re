/* states.c — 主迴圈狀態機 handler 反編譯(real-mode large-model C)
 *
 * 反編譯來源:tools/re_states_dis.py(docker capstone)反組譯 sub_93e3 跳表的 6 個 handler
 * (DS:0x25 near 指標表)與指令選單二級派發跳表(DS:0x3baa)。對照 docs/10-exe-states.md。
 *
 * scancode→handler 對照(DS:0x19 / DS:0x25):
 *   Q/S/F1 → h_cmd_menu | Space → h_quick_act | Enter → h_confirm
 *   F5 → h_info_win | F6 → h_save_slot | F9 → h_f9_noop
 *   F4/F7/F8 → 已註冊但指標為 0(吃掉按鍵不做事)
 *
 * 求結構正確、可審閱;非位元精準。far runtime 以 exe.h wrapper 表達;
 * 核心遊戲函式(玩家移動 19b8 / 場景繪製 255b / HUD 9530)只 forward-declare 不反編譯。
 */
#include "states.h"

/* 文字 / 對話繪製器 111b:0264(全程式最熱常式,見 docs/06)。
 * 入口慣例:DI = 訊息 / 字串 ID;由 [0x252e] 取字碼序列逐字繪製。 */
extern void text_draw(u16 msg_id);   /* lcall 111b:0264, DI=msg_id */

/* 額外 runtime / 工具(seg0 內共用,屬輔助層;此處宣告以表達呼叫關係)。 */
extern u8   game_flag_get(u16 idx);  /* sub_8279  bx=flag idx → al(0/1) */
extern void game_flag_set(u16 idx);  /* sub_8264  bx=flag idx,設定旗標 */
extern void win_open(void *desc);    /* sub_10900 建視窗 / 游標 */
extern void win_nav(void *desc, u16 rows); /* sub_10ae9 選單導航 */
extern void win_close(void *desc);   /* sub_10974 關視窗 */
extern void party_status_draw(void); /* sub_9592 畫隊伍 / 狀態列 */
extern void confirm_worker(void);    /* sub_9cd6 Enter 主動作 worker */
extern void confirm_stage2(void);    /* sub_60c6 Enter 第二段 */
extern void confirm_stage3(void);    /* sub_5b2d Enter 第三段 */
extern void info_load_block(void);   /* sub_ec59 F5 載入顯示資料(int21 42/3F) */
extern void info_page2(void);        /* sub_2793 F5 額外資訊頁(含座標備份) */
extern void mode_branch(void);       /* sub_434f Space/Enter 特定模式分支 */
extern void main_restart(void);      /* jmp 0xa62c 重入主流程(cmd_act4) */

/* 選單 / 視窗描述結構位址(主資料段內) */
#define WIN_CMD   ((void *)0x424c)    /* 指令選單視窗 [0x424c] */
#define WIN_INFO  ((void *)0x3e6e)    /* F5 資訊框 [0x3e6e] */
#define MSG_SAVE  0x3fba              /* F6 存檔框訊息 ID */

/* ============================================================
 * h_cmd_menu — Q / S / F1:開指令選單(field command window)
 * sub_9842 (file 0xabb2)
 * ============================================================ */
void h_cmd_menu(void)
{
    mouse_op(0x166);              /* 凍結畫面 / 滑鼠 off */
    g_modal_busy = 1;             /* [0x72a]=1 進 modal */
    mouse_op(0x19c);              /* 滑鼠 on */

    win_open(WIN_CMD);            /* 建指令視窗 + 游標 */
    win_nav(WIN_CMD, 5);          /* 5 列選單導航;結果 → [0x722]/[0x726] */

    if (g_action_result != 1) {   /* != 取消 → 派發選中指令 */
        int sel = g_menu_sel - 1; /* [0x722] 1-based → 0-based */
        if (g_cmd_menu_disp[sel] != 0)
            g_cmd_menu_disp[sel]();   /* call word ptr [bx*2+0x3baa] */
    }

    win_close(WIN_CMD);
    mouse_op(0x166);
    g_modal_busy = 0;             /* [0x72a]=0 離開 modal */
    mouse_op(0x19c);
}

/* ---- 指令選單二級派發:前 5 項(0xabfd 群組)+ 道具清單 ----
 * 多數以 [0x000f] 輸入旗標分支、檔案 I/O、設定 [0x286d]/[0x2784] 狀態。
 * 精確指令名(話す/呪文/装備…)待追入子函式對照 D3TXT;此處保留結構。 */

void cmd_act0(void)   /* sub_988d */
{
    if (g_input_flags & 0x18) {
        g_word(0x286d) = (g_input_flags >> 3) & 3;   /* 取 2-bit 子選擇 */
        /* 條件式檔案操作(lcall 1053:dc / 1053:154,改 [0x2896] 狀態) */
        file_op_dx(0xdc);
        file_op_dx(0x154);
    }
}

void cmd_act1(void)   /* sub_98b9 */
{
    if (g_input_flags & 0x18) {
        g_word(0x286d) = 0;
        file_op_dx(0x16c);
    }
}

void cmd_act2(void)   /* sub_98cd */
{
    if (g_input_flags & 1) {
        g_word(0x2784) = 0;
        mouse_op(0x71);
        mouse_op(0x3a0);
    }
}

void cmd_act3(void)   /* sub_98e6 */
{
    if (g_input_flags & 1) {
        g_word(0x2784) = 1;
        mouse_op(0x385);
        g_word(0x2784) = 0xff;
    }
}

void cmd_act4(void)   /* sub_9900 — 重入主流程 */
{
    kbd_set_mode(0x01, 0x23);     /* lcall 1104:123 等待 / flush */
    /* call 0x109ac;若 [0x722]!=2 則 call 0x7baa + 關窗 + jmp 0xa62c 重啟主流程 */
    if (g_menu_sel != 2) {
        confirm_stage2();         /* 佔位:此分支實為 sub_7baa,語意待精確化 */
        win_close(WIN_CMD);
        main_restart();           /* jmp 0xa62c:不回傳,重入主流程 */
    }
}

/* sub_504e:道具 / 裝備 / 隊伍清單(逐欄位列出,空欄印不同訊息 ID) */
void cmd_item_list(void)   /* sub_504e */
{
    /* 反組譯重點:讀 [0x507f] 結構,逐欄位判空(test [si+0x38]&0x80),
     * 以 text_draw 列出每欄(訊息 ID 0xbca/0xbcb/0xbce/0xbcf…),
     * call game_flag_get(0x29) 判可用,選定後 call 0x7baa 套用並重繪。
     * 完整迴圈龐大,此處以語意骨架表達。 */
    if (game_flag_get(0x29) != 1) {
        text_draw(0xbcb);         /* 不可用提示 */
        return;
    }
    text_draw(0xbca);             /* 清單抬頭 / 每欄項目(實際在迴圈內) */
    /* … 逐欄列出 + 選取 + 套用(sub_7baa)… */
}

/* ============================================================
 * h_quick_act — Space:快速互動 / 調べる
 * sub_7c83 (file 0x8ff3)
 * ============================================================ */
void h_quick_act(void)
{
    u16 mx, my;

    if (g_mode_flag >= 2) {       /* [0x4f3b]>=2:特定模式改走別處 */
        mode_branch();            /* sub_434f */
        return;
    }

    mouse_op(0x166);
    mouse_get(&mx, &my);          /* int 33h ax=3 → [0x2852]/[0x2854] */
    g_modal_busy = 1;
    mouse_op(0x19c);

    g_byte(0x2875) = 1;
    file_op_dx(0x240);            /* lcall 1053:240 資源 / 檔案操作 */

    /* 建隊伍狀態框 [0x3e9c]:座標常數 0x13/0xee + g_party_count 算高度 */
    if (game_flag_get(0x27) != 1) {
        win_open((void *)0x3e9c);
        party_status_draw();      /* sub_9592 畫隊伍 / 狀態列 */
    }

    kbd_set_mode(0x01, 0x23);     /* 等待鍵 */
    if (game_flag_get(0x27) != 1) {
        win_close((void *)0x3e84);
        win_close((void *)0x3e9c);
    }

    mouse_op(0x166);
    g_modal_busy = 0;
    mouse_set(mx, my);            /* int 33h ax=4 還原滑鼠 */
    mouse_op(0x19c);
}

/* ============================================================
 * h_confirm — Enter:確認 / 與前方互動 / 推進多段對話事件
 * sub_7c43 (file 0x8fb3)
 * ============================================================ */
void h_confirm(void)
{
    if (g_mode_flag >= 2) {       /* 同 Space:特定模式改走別處 */
        mode_branch();
        return;
    }

    mouse_op(0x166);
    g_action_active = 1;          /* [0x1f7]=1 動作執行中 */

    confirm_worker();             /* sub_9cd6:讀前方物件 [0x2511]→[bx+0x37c4],推進事件 */
    if (g_action_result != 0) {   /* 逐段:上一段非 0 才接續 */
        confirm_stage2();         /* sub_60c6 */
        if (g_action_result != 0)
            confirm_stage3();     /* sub_5b2d */
    }

    kbd_set_mode(0x01, 0x23);     /* 等待鍵 */
    g_action_active = 0;          /* [0x1f7]=0 */
}

/* ============================================================
 * h_info_win — F5:開狀態 / 隊伍資訊視窗(唯讀顯示)
 * sub_13a9 (file 0x2719)
 * ============================================================ */
void h_info_win(void)
{
    g_modal_busy = 1;

    /* 進入時 AX!=0 才開子視窗(由 caller 決定);[0x259b]=0 文字行計數歸零 */
    win_open(WIN_INFO);
    g_byte(0x259b) = 0;

    info_load_block();            /* sub_ec59:int21 AH=42/3F 從 [0x2534] 段載入區塊 */

    text_draw(0xfd);             /* 主訊息頁 */
    /* call 0x9ac(座標 / 計數前進) */
    if (g_menu_sel == 1) {
        text_draw(0xfa);         /* 額外資訊頁 */
        info_page2();            /* sub_2793:複製 [0x507f] 結構 + 備份 camera 座標 */
    }
    text_draw(0xfc);             /* 收尾訊息 */

    kbd_set_mode(0x01, 0xdb);     /* lcall 1104:db 等待 */
    win_close(WIN_INFO);
    g_modal_busy = 0;
}

/* ============================================================
 * h_save_slot — F6:存檔 / 載入槽選擇
 * sub_14da (file 0x284a)
 * ============================================================ */
void h_save_slot(void)
{
    mouse_op(0x166);
    g_modal_busy = 1;
    mouse_op(0x19c);

    g_info_mode = 1;              /* [0x1f0]=1 選槽 / 編輯模式(對比 F5 的 0) */
    info_win_build();            /* sub_29d0:與 F5 共用,建視窗矩形 [0x4f21..0x4f27] */

    if (g_action_result == 2)     /* 取消 */
        return;

    text_draw(MSG_SAVE);         /* 存檔框 [0x3fba] */

    if (g_action_result != 1) {  /* 確認選槽 */
        /* 掃 [0x128] 槽位陣列找空槽,設 g_menu_sel(+1) */
        u8 *slots = (u8 *)0x128;
        i16 idx = 0, n = 0;
        while (slots[idx * 0x14 + 0x12] != 0) {  /* 已用槽 */
            n++;
            if (n == g_menu_sel) break;
            idx++;
        }
        g_menu_sel = idx;
        g_menu_sel++;            /* inc [0x722] */

        if (slots[(g_menu_sel - 1) * 0x14 + 0x13a - 0x128] != 0) {
            mouse_op(0x166);
            save_write_slot();   /* sub_28ce:改檔名數字位 + DOS 開讀關 */
            save_pack_buffer();  /* sub_2902:組裝存檔緩衝 [0x4f15]→[0x613] */
            mouse_op(0x19c);
        }
    }

    mouse_op(0x166);
    g_modal_busy = 0;
    mouse_op(0x19c);
}

/* ============================================================
 * h_f9_noop — F9:no-op stub(入口即 ret)
 * sub_e915 (file 0xfc85)
 * ============================================================ */
void h_f9_noop(void)
{
    /* sub_e915: c3 (ret) — 已註冊但停用,不做任何事 */
}
