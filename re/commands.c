/* commands.c — 野外指令子系統 + 對話流程反編譯(real-mode large-model C)
 *
 * 反編譯來源:tools/re_cmd_dis.py(docker capstone)反組譯
 *   - 指令選單二級派發表 DS:0x3baa 的 idx6..11 handler
 *   - 子動作 sub_16dd / sub_1a4c / sub_7c50
 *   - Enter 對話 / 事件 worker chain sub_9cd6 / sub_60c6 / sub_5b2d
 *   - 文字繪製器 111b:0264 控制碼分支
 * 對照 docs/12-exe-commands.md、docs/03(文字格式)、docs/10(狀態機)。
 *
 * 範圍界定:
 *   - 指令選單「開窗 + 選定 + 二級派發」框架在 states.c(h_cmd_menu)已反編譯,
 *     本檔接手「選定之後各 field 指令動作 + 對話流程」。
 *   - idx0..5(cmd_act0..cmd_item_list)已在 states.c 就地反編譯,本檔不重複,
 *     只反編譯 idx6..11 與對話流程鏈。
 *   - 文字繪製器本體在 runtime 段 0x111b(手寫組語),此處以 text_draw() wrapper
 *     表達其入口語意與控制碼;繪字底層(VGA planar)不在本 slice。
 *
 * 求結構正確、可審閱;非位元精準。far runtime 以 exe.h wrapper 表達;
 * 與本 slice 無關的核心函式 forward-declare 不反編譯。
 */
#include "commands.h"
#include "states.h"   /* g_menu_sel / g_action_result / game_flag_get/set / win_*  */

/* ---- 本 slice 呼叫到、但屬其他子系統的函式(forward-declare,不反編譯)---- */
extern void apply_redraw(void);     /* sub_7baa  選定後套用 + 重繪(別處詳解;此處只呼叫) */
extern void info_window(u16 mode);  /* sub_2719  開資訊 / 狀態視窗(= h_info_win caller) */
extern void msg_window_reset(void); /* sub_6372  清訊息視窗 */
extern void msg_return_loop(void);  /* sub_6380  訊息後回主選單迴圈 */
extern void seq_msg(void);          /* sub_7c0c  逐段動畫訊息序列 */
extern void seq_msg2(void);         /* sub_7bbe  訊息序列變體 */
extern void confirm_worker(void);   /* sub_9cd6  Enter 主動作 worker(states.c 已宣告) */
extern void confirm_stage2(void);   /* sub_60c6 */
extern void confirm_stage3(void);   /* sub_5b2d */

/* 場景 / 物件 / 隊伍存取(本 slice 只呼叫) */
extern void *obj_lookup(u16 idx);   /* sub_aba4  以索引取物件結構指標 → SI */
extern void map_obj_walk(void);     /* sub_fa39  走訪地圖物件(Enter worker 用) */
extern void event_run(u16 a, u16 b);/* sub_ba71 / sub_bae8  執行事件 / 對話 */
extern void event_none(void);       /* sub_ba41  「沒有任何事」訊息 */

/* ============================================================
 * idx6  cmd_talk — 話す(交談)
 * sub_5112 (file 0x6482):整段只是 call sub_16dd; ret。
 * sub_16dd 以 [0x722] 為頁碼,逐項把名冊 / 對話對象畫到視窗(lcall 111b:779 畫數字、
 * 111b:43e 畫框、111b:214 畫字串),掃 [si] 結構(stride 0x14)直到 cx 用盡,
 * 末了以 win_nav([0x3fba], [0x3fce]) 讓玩家選對象。是 DQ「話す → 選對象」的清單頁。
 * ============================================================ */
void cmd_talk(void)   /* sub_5112 */
{
    talk_page();      /* sub_16dd */
}

void talk_page(void)  /* sub_16dd (file 0x2a4d) */
{
    /* 反組譯骨架(逐頁 / 逐欄列出對話 / 名冊對象):
     *   for (每欄位 si, 共 cx 欄; si += 0x14, dx += 0x10) {
     *       text_num([0x722]);                 ; lcall 111b:779 畫序號
     *       if (si[0x12] != 0) {               ; 該欄有對象
     *           [0x716]=bp+8; [0x718]=dx;
     *           draw_box();                    ; lcall 111b:43e
     *           text_num(si[0x12]);            ; 畫對象屬性
     *           di = 0x216 + si[0x13] - 1;     ; 對象名字串 ID
     *           text_str(di);                  ; lcall 111b:214
     *       } else {
     *           di = 0x1d4; text_str(di);      ; 空欄提示
     *       }
     *       [0x722]++;
     *   }
     *   win_nav([0x3fba], [0x3fce]);           ; 選對象(call 0xae9)
     * 此處以語意骨架表達;逐欄迴圈與 111b 子功能屬繪字底層。 */
    g_menu_sel = g_menu_sel;   /* [0x722] 作為頁內序號累加(見反組譯 inc [0x722]) */
    win_nav((void *)0x3fba, *(u16 *)0x3fce);
}

/* ============================================================
 * idx7  cmd_flag_msg — 旗標查詢 + 訊息輸出
 * sub_5116 (file 0x6486)
 * ============================================================ */
void cmd_flag_msg(void)   /* sub_5116 */
{
    u16 di;
    msg_window_reset();                 /* sub_6372 清訊息視窗 */
    di = (game_flag_get(FLAG_CMD_AVAIL) == 1) ? 0xbd1 : 0xbd2;
    text_draw(di);                      /* 依旗標印不同訊息 */
    msg_return_loop();                  /* jmp 0x6380 回主選單迴圈 */
}

/* ============================================================
 * idx8  cmd_move_act — 子動作(地圖移動 / 座標解算)
 * sub_512f (file 0x649f):整段只是 call sub_1a4c; ret。
 * sub_1a4c:先 lcall 1053:240 / 1053:dc(資源 / 檔案),call 0xa9(?),
 * 再 call 0x3823 / 0x3866(畫面 / 捲動),kbd 1104:123 等鍵。其後段(0x2dda 起)
 * 是「前方座標 = [0x4f33]+dx / [0x4f35]+dy → 邊界檢查 → 查 tile 屬性
 * [bx+0x308e] → 設 [0x4f2d]=3(切城鎮 / 傳送)或更新移動旗標 [0x258a]」,
 * 為地圖移動 / 出入口 / 傳送的核心解算。
 * ============================================================ */
void cmd_move_act(void)   /* sub_512f */
{
    move_worker();        /* sub_1a4c */
}

void move_worker(void)    /* sub_1a4c (file 0x2dbc) */
{
    file_op_dx(0x240);    /* lcall 1053:240 資源 */
    file_op_dx(0xdc);     /* lcall 1053:dc */
    /* call 0xa9 / 0x3823 / 0x3866:畫面更新 + 捲動(屬繪製子系統) */
    apply_redraw();       /* 佔位:此處實為 0x3823/0x3866 畫面;語意=移動後重繪 */
    kbd_set_mode(0x01, 0x23);  /* lcall 1104:123 等鍵 */

    /* --- 0x2dda:前方目標座標 + tile 屬性解算(移動 / 傳送判定)---
     * i16 tx = g_word(0x4f33) + g_player_dx;
     * i16 ty = g_word(0x4f35) + g_player_dy;
     * if (out of [0xb28]x[0xb2a] bounds) { g_word(0x4f2d)=3; ... }  // 切場景
     * else { tile = mapseg[ty*W + tx]; attr = [tile*2 + 0x308e];     // 查屬性表
     *        if (attr & ...) set move dir [0x258a] / 牆面旗標 [0x4f46]; }
     * g_word(0x4f33)+=dx; g_word(0x4f35)+=dy;                        // 套用位移
     * 完整屬性位元語意屬地圖 / 碰撞子系統(另一 slice),此處標出入口與資料流。 */
}

/* ============================================================
 * idx9  cmd_examine — 調べる / 確認(指令選單版)
 * sub_5133 (file 0x64a3):整段只是 call sub_7c50; ret。
 * sub_7c50 = Enter 對話 / 事件 worker chain 的共用入口(見下 examine_chain)。
 * 「指令選單→調べる」與「按 Enter」走同一條對話推進鏈,只是入口不同。
 * ============================================================ */
void cmd_examine(void)    /* sub_5133 */
{
    examine_chain();      /* sub_7c50 */
}

/* ============================================================
 * idx10 cmd_cond_a — 條件動作(查旗標 → 訊息 + 資訊視窗)
 * sub_5137 (file 0x64a7)
 * ============================================================ */
void cmd_cond_a(void)     /* sub_5137 */
{
    msg_window_reset();                 /* sub_6372 */

    if (game_flag_get(FLAG_CMD_AVAIL) == 1) {
        text_draw(0xbe7);               /* 可用 → 提示訊息 */
        info_window(0);                 /* call 0x2719 開資訊視窗(mode 0) */
        return;
    }
    if (game_flag_get(FLAG_COND_B) != 1) {
        text_draw(0xbe8);               /* 條件不足訊息 */
        msg_return_loop();              /* jmp 0x6380 */
        return;
    }

    /* flag 0x4d 成立分支:逐格動畫(di=0xc1a/0xc1b),沿 [0x251d] 0x14→0xc8
     * 推進畫面,call sub_7baa 套用,設多個視窗矩形常數 [0x4f44/40/42/3c/3e],
     * 設旗標 0x4d / 0x1a,清 [0x506a] 9 bytes 為 0xff。語意:特定事件動畫 +
     * 狀態切換(如階梯 / 機關開啟)。 */
    text_draw(0xc1a);
    /* … 動畫迴圈 + apply_redraw() + 旗標設定 … */
    game_flag_set(FLAG_COND_B);
    game_flag_set(0x1a);
}

/* ============================================================
 * idx11 cmd_cond_b — 條件動作(查旗標 → 訊息 + 資訊視窗)
 * sub_51e9 (file 0x6559)
 * ============================================================ */
void cmd_cond_b(void)     /* sub_51e9 */
{
    u16 di;
    msg_window_reset();                 /* sub_6372 */

    if (game_flag_get(FLAG_CMD_AVAIL) == 1) {
        text_draw(0xbe9);               /* 已完成 / 不可用 */
        msg_return_loop();
        return;
    }
    di = (game_flag_get(FLAG_COND_B) == 1) ? 0xc19 : 0xbea;
    text_draw(di);
    info_window(0);                     /* call 0x2719 */
}

/* ============================================================
 * Enter 對話 / 事件流程:worker chain 共用入口
 * sub_7c50 (file 0x8fc0)
 *   = h_confirm(Enter)與 cmd_examine(指令調べる)共用。
 *   逐段呼叫 worker;每段以 [0x726] 結果碼判斷是否續下一段,
 *   構成「按 Enter / 選調べる → 逐段推進對話 / 事件」的流程。
 * ============================================================ */
void examine_chain(void)  /* sub_7c50 */
{
    confirm_worker();                   /* sub_9cd6:讀前方物件 → 跑事件 / 起對話 */
    if (g_action_result != 0) {         /* [0x726]!=0 才續段 */
        confirm_stage2();               /* sub_60c6 */
        if (g_action_result != 0) {
            confirm_stage3();           /* sub_5b2d */
            /* [0x726] 仍非 0 可再續;原碼此處 je 落到收尾 */
        }
    }
    kbd_set_mode(0x01, 0x23);           /* lcall 1104:123 等鍵 */
    g_action_active = 0;                /* [0x1f7]=0 動作結束 */
}

/* ------------------------------------------------------------
 * confirm_worker — sub_9cd6:Enter / 調べる 的主動作 worker
 * (states.c 已 forward-declare;此處給出對話流程的語意反編譯)
 *
 * 反組譯重點(file 0xb04b 起;函式入口在 0xb037 附近的迴圈頭):
 *   for (g_obj_loop=1; g_obj_loop<=cx; g_obj_loop++) {        // 掃前方 / 周邊格
 *       u16 ft = g_front_tile;                                 // [0x2511]
 *       u8  ev = *(u8 *)(ft*3 + OBJ_TABLE);                    // [ft*3 + 0x37c4]
 *       if (ev != 0xff) {                                      // 有事件 / 對象
 *           bp = ev; if (bp==0) bp=1;
 *           map_obj_walk();                                    // sub_fa39 取對象資料
 *           g_word(0x2593) = dx;                               // 對話 / 數值參數
 *           event_run(...);                                    // sub_ba71 跑事件 / 起對話
 *           g_event_hit++;                                     // [0x2702]++
 *       } else {
 *           event_run_else();                                  // sub_bae8
 *           g_event_hit++;
 *       }
 *   }
 *   if (g_event_hit == 0) event_none();                        // sub_ba41「沒有反應」
 *
 * 事件命中後,對話本體經 text_draw(msg_id) 逐字繪製:其讀 [0x252e] 遠段
 * 的 2-byte 字碼序列(即 D3TXT 記錄),遇 TXT_NEWLINE/PAGEFEED/VAR_* 控制碼
 * 按 docs/03 語意換行 / 換頁(等鍵)/ 插值。對話的「按 Enter 翻頁」即由
 * PAGEFEED(0xfffc)在繪製器內 lcall 1104:123 等鍵實現(見 text_draw 註解)。
 * ------------------------------------------------------------ */

/* ============================================================
 * text_draw — 文字 / 對話繪製器 111b:0264 入口語意
 * (本體為 runtime 段 0x111b 手寫組語;此處反編譯其控制流以記錄語意)
 *
 * 入口:DI=msg_id;由 [0x3e70]/[0x3e72] 取視窗左上座標,x+2、y+0x10,
 *      y 再加 g_txt_line*0x10(逐行下移)。取 SI=記錄起點後逐碼:
 *
 *   for (;;) {
 *       u16 bx = *(u16 far *)(g_txt_seg:si);   // 從遠段取 2-byte 字碼
 *       switch (bx) {
 *         case TXT_END:      // 0xffff:收尾(若 g_txt_nowait!=1 等鍵),retf
 *         case TXT_NEWLINE:  // 0xfffe
 *         case TXT_NEWLINE2: // 0xfffd:dx+=0x10;若 g_txt_line==3 換頁等待,否則 line++
 *         case TXT_PAGEFEED: // 0xfffc:lcall 111b:6 翻頁 + lcall 1104:123 等鍵
 *         case TXT_VAR_NAME: // 0xfffb:插主角 / 角色名
 *         case TXT_VAR_ITEM: // 0xfffa:插道具名
 *         case TXT_VAR_NUM:  // 0xfff9:插數值(g_txt_varnum)
 *         case TXT_VAR_B:    // 0xfff5
 *         case TXT_VAR_A:    // 0xfff6
 *         case TXT_VAR_ED:   // 0xffed:插值 / 控制
 *         default:           // glyph index:畫 16x16 字模,bp(x)+=字寬,si+=2
 *       }
 *   }
 *
 * 與 docs/03 控制碼表雙向印證:換行後接縮排、換頁等鍵、VAR_* 為 {變數} 占位。
 * g_txt_line==3 為一頁四行(0..3)上限,滿頁即停下等玩家按鍵翻頁,
 * 正是 DQ 對話框「一次顯示數行、按鍵續頁」的行為。
 * ============================================================ */
/* text_draw():實作在 runtime 段,宣告於 commands.h;此處不提供 C body
 * (lcall 111b:0264)。上方註解即其控制流反編譯記錄。 */
