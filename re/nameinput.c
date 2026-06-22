/* nameinput.c — 精訊 DQ3 注音(Zhuyin)姓名輸入子系統反編譯(real-mode large-model C)
 *
 * 反編譯來源:tools/re_name_dis.py(docker capstone)反組譯角色建立 modal sub_0854 進入的
 * 「輸入姓名」畫面。對照 docs/15-exe-nameinput.md。
 *
 * 子系統鏈:
 *   sub_0854(角色建立 modal)
 *     └─ sub_0d17  ni_dispatch     依 g_ni_mode 分派三畫面,迴圈到「完成 / 取消」
 *          ├─ sub_0dc8  ni_fn_list     功能列(英數/前進/後退/取消/完成)
 *          ├─ sub_0e8e  ni_alnum       英數鍵盤(0-9 A-Z 羅馬數字)
 *          └─ sub_0f5b  ni_zhuyin      注音鍵盤(ㄅㄆㄇㄈ… + 聲調)
 *               ├─ sub_23f7  ni_grid_nav      方向鍵 + Space/Enter 選格
 *               ├─ sub_2623  ni_compose_put   注音符號 → 組字緩衝
 *               └─ sub_2654  ni_compose_resolve 組字 → 查國字(選取國字候選窗)
 *
 * 注音輸入法核心:玩家在注音鍵盤上逐格選「聲母/介音/韻母/聲調」填入組字緩衝 [0x2730..0x2733],
 * 按下「組字完成格」後由注音引擎(seg 0x11c4)查出國字;若有多個同音字,開「選取國字」候選窗
 * 讓玩家挑一個。選定的國字寫進名字緩衝 [0x2710],名字上限 4 字。
 *
 * 求結構正確、可審閱;非位元精準。far runtime 以 wrapper 表達。
 */
#include "nameinput.h"

/* 文字 / 版面繪製器 111b:0214(以 DI=msg_id 畫整段版面;座標由 [0x716]/[0x718] 帶入)。
 * 與主迴圈用的 111b:0264 同段,0x214 是「畫一整筆 D3TXT 記錄到指定位置」入口。 */
extern void layout_draw(u16 msg_id, u16 x, u16 y); /* lcall 111b:0214, DI=id, [0x716]=x [0x718]=y */
extern void glyph_draw(u16 code, u16 x, u16 y);    /* lcall 111b:0006 畫單一字模(DX=x BP=y) */

/* 視窗 / 滑鼠 runtime(與 states.c 共用層) */
extern void win_open(void *desc);   /* sub_0900 建視窗 + 游標 */
extern void win_close(void *desc);  /* sub_0974 關視窗 */
extern void win_nav(void *desc, u16 rows); /* sub_0ae9 選單導航,結果 → g_menu_sel/g_action_result */
extern void beep_msg(u16 id);       /* sub_90e2 提示音 / 訊息(si=id) */

/* 注音輸入視窗描述結構位址(主資料段內) */
#define NI_WIN_TITLE  ((void *)0x42b8)  /* [0x42b8] 標題 / 名字列視窗 */
#define NI_WIN_GRID   ((void *)0x42f4)  /* [0x42f4] 注音 / 英數鍵盤格視窗 */
#define NI_WIN_PICK   ((void *)0x4310)  /* [0x4310] 「選取國字」候選窗 */
#define NI_FN_DESC    ((void *)0x42d6)  /* [0x42d6] 功能列(5 列)導航描述 */

/* ============================================================
 * ni_dispatch — 主控 dispatcher(sub_0d17, file 0x2087)
 * 依 g_ni_mode 位元分派到「注音 / 英數 / 功能列」三畫面,迴圈直到「完成」(bit3)或「取消」。
 * ============================================================ */
void ni_dispatch(void)
{
    int i;

    g_ni_len = 0;                    /* [0x270a]=0 名字字數歸零 */
    g_ni_maxlen = 0;                 /* [0x270c]=0 */
    for (i = 0; i < 16; i++)         /* 清名字緩衝 [0x2710] 16 word */
        g_ni_name[i] = 0;
    g_ni_grid_n = 0;                 /* [0x2712]=cx(迴圈結束 cx=0)→ 起始 maxlen 標記 */

    g_ni_mode = 0;                   /* [0x26fc]=1 起手 → 注音模式(下面 win_open 後進迴圈) */
    g_ni_mode = NI_MODE_FN >> 1;     /* (還原:實際 mov [0x26fc],1;此處示意起手非英數非功能列) */
    g_ni_mode = 1;                   /* 起手值 1(bit0,實際分派看 bit1/bit2) */

    win_open(NI_WIN_TITLE);          /* 建標題 / 名字列視窗 [0x42b8] */
    ni_redraw_cursor();              /* call 0x21c5 畫名字列游標 */

    for (;;) {                       /* loop @0x20b8 */
        if (g_ni_mode & NI_MODE_EXIT) {           /* bit3:離開 */
            if (g_ni_len < 1) {                   /* 名字空 → 不准離開,清旗標續輸入 */
                g_ni_mode = 0;
                continue;
            }
            ni_count_name();                      /* call 0x2114 重數字數 → g_ni_len */
            g_ni_cell = g_ni_len;                 /* [0x270e]=[0x270a](暫存最終字數) */
            break;
        }
        if (g_ni_mode & NI_MODE_FN) {             /* bit2:功能列 */
            ni_fn_list();                         /* call 0x2138 */
            if (g_action_result == 1)             /* 取消 */
                break;
            continue;
        }
        if (g_ni_mode & NI_MODE_ALNUM) {          /* bit1:英數鍵盤 */
            ni_alnum();                           /* call 0x21fe */
            if (g_action_result == 1)
                break;
            continue;
        }
        ni_zhuyin();                              /* 否則:注音鍵盤 call 0x22cb */
        if (g_action_result == 1)
            break;
    }

    win_close(NI_WIN_TITLE);         /* 關標題視窗 [0x42b8] */
}

/* ============================================================
 * ni_count_name — 數名字非空字數(sub_0da4, file 0x2114)
 * 掃名字緩衝每 4 byte(2 word)一字,兩 word 皆 0 視為空,數出實際字數。
 * ============================================================ */
void ni_count_name(void)
{
    u16 *p = g_ni_name;       /* si=[0x2710] */
    int n = 0;
    int i;
    for (i = 0; i < 4; i++) {            /* cx=4(最多 4 字) */
        if (p[0] != 0 || p[1] != 0)     /* 任一 word 非 0 = 有字 */
            n++;
        p += 2;                          /* 下一字(+4 byte) */
    }
    g_ni_len = n;             /* [0x270a]=bp */
}

/* ============================================================
 * ni_fn_list — 功能列(sub_0dc8, file 0x2138)
 * 在右欄 5 列功能鍵上導航:英數/注音切換、前進、後退、取消、完成。
 * win_nav 結果 g_menu_sel(1-based);g_action_result==1 表使用者按取消鍵離開導航。
 * 版面對照(rec 452 / 453 右欄,由上到下):
 *   row1 英數(注音模式時)/ 注音(英數模式時)= 切換鍵盤模式
 *   row2 前進  row3 後退  row4 取消  row5 完成
 * ============================================================ */
void ni_fn_list(void)
{
    win_nav(NI_FN_DESC, 5);              /* cx=5 列;結果 → g_menu_sel / g_action_result */

    if (g_action_result == 1)            /* 導航被取消 */
        goto done;

    if (g_menu_sel == 1) {               /* row1:英數/注音 切換 */
        if (g_ni_mode & NI_MODE_ALNUM) { /* 目前英數 → 切回注音(清 bit1,留低位 bit) */
            g_ni_mode = (g_ni_mode & 0xe) | NI_MODE_ALNUM;  /* 0x26fc &=0xe; |=2 */
        } else {                         /* 目前注音 → 切英數 */
            g_ni_mode = (g_ni_mode & 0xd) | 1;              /* 0x26fc &=0xd; |=1 */
        }
        goto done;
    }
    if (g_menu_sel == 3) {               /* row3:後退(游標往前一格,刪/退) */
        if (g_ni_len != 0) {             /* 有字才退 */
            ni_redraw_cursor();
            g_ni_len--;                  /* dec [0x270a] */
            ni_redraw_cursor();
        }
        goto done;
    }
    if (g_menu_sel == 2) {               /* row2:前進(游標往後一格) */
        if (g_ni_len < g_ni_maxlen) {    /* 未超過已輸入長度 */
            ni_redraw_cursor();
            g_ni_len++;                  /* inc [0x270a] */
            ni_redraw_cursor();
        }
        goto done;
    }
    if (g_menu_sel == 4) {               /* row4:取消(整個姓名輸入作廢) */
        g_action_result = 1;             /* [0x726]=1 → dispatcher break */
        goto done;
    }

    /* row5(else):★完成★ —— 設離開旗標,dispatcher 下一輪 bit3 命中即定案名字 */
    g_ni_mode |= NI_MODE_EXIT;           /* or [0x26fc], 8 */

done:
    g_ni_mode &= 0xb;                    /* 清 bit2(離開功能列模式),回主分派 */
}

/* ============================================================
 * ni_redraw_cursor — 重畫名字列游標(sub_0e55, file 0x21c5)
 * 名字列每字佔 2 字格,游標 X = 0x1f + len*2,Y=0x3e;並以 int33 ax=4 設定
 * 滑鼠命中熱區(讓玩家也能用滑鼠點名字列)。
 * ============================================================ */
void ni_redraw_cursor(void)
{
    u16 di;
    mouse_op(0x385);                     /* 滑鼠關 / 凍結 */
    di = g_ni_len;                       /* [0x270a] */
    if (g_ni_maxlen < di)                /* 更新已達最大字數(前進界限) */
        g_ni_maxlen = di;
    di = di * 2 + 0x1f;                  /* 游標 X(字格) */
    /* call 0x25df 畫游標方塊 @ (di, 0x3e) */
    /* int 33h ax=4:設滑鼠座標到 ((di*8)+1, 0x3e+8) 對齊名字列 */
    mouse_op(0x3a0);                     /* 滑鼠開 */
}

/* ============================================================
 * ni_alnum — 英數鍵盤畫面(sub_0e8e, file 0x21fe)
 * 畫 rect 0x1c5(rec 453:0-9 A-Z 羅馬數字 ★/☆/∞/⊙ + OK + 功能列),
 * 進 grid 導航;選到字元格 → 直接把對應 glyph index 寫進名字緩衝。
 * ============================================================ */
void ni_alnum(void)
{
    u16 save = g_ni_win_save;            /* push [0x2857] */
    u16 cell, glyph;

    layout_draw(NI_MSG_ALNUM, 0x13, 0x5e);  /* DI=0x1c5 畫英數版面 @ (0x13,0x5e) */
    g_ni_win_save = 1;

    for (;;) {
        if (ni_grid_nav() != 0) {        /* 導航;al!=0 表滑鼠移動取消本次,但這裡用 g_action_result */
            /* (實際以 g_action_result==1 判離開,見下) */
        }
        if (g_action_result == 1)        /* grid 導航回報取消 */
            break;

        /* 換算 cell:[0x26fe] raw → div 寬 → row*5+col */
        cell = ni_alnum_cell_index();    /* dl=col, ah=row; cell = row*5 + col */
        g_ni_cell = cell;

        if (cell <= NI_CELL_OK) {        /* <= 0x2a:字元格(0-9 A-Z 羅馬數字 + OK) */
            ni_redraw_cursor();
            g_ni_glyph_hi = 0xffff;      /* [0x2756]=0xffff(英數字只用低 word) */

            /* cell → glyph index 對照:
             *   cell 0..9   → glyph 0..9   (數字 0-9)
             *   cell 10..0x23(35) → glyph (cell-10)+0xf  = 15..40 (A-Z)
             *   cell >0x23  → glyph (cell-0x23)+0x71      (羅馬數字 / 符號 / OK) */
            {
                u16 g = cell;
                if (g < 0xa) {
                    /* 0-9:glyph = cell */
                } else if (g <= 0x23) {
                    g = (g - 0xa) + 0xf;          /* A-Z */
                } else {
                    g = (g - 0x23) + 0x71;        /* 羅馬數字 / ★ / OK… */
                }
                g_ni_glyph_lo = g;                /* [0x2754]=glyph */
                glyph = g;
            }
            glyph_draw(glyph, 0x3e, g_ni_len * 2 + 0x1f);  /* 在名字列畫該字 */

            /* 寫進名字緩衝第 g_ni_len 字(每字 4 byte) */
            g_ni_name[g_ni_len * 2 + 0] = g_ni_glyph_hi;   /* [bx+0x2710] */
            g_ni_name[g_ni_len * 2 + 1] = g_ni_glyph_lo;   /* [bx+0x2712] */
            g_ni_len++;
            if (g_ni_len == NI_NAME_MAX)          /* 滿 4 字 → 停在最後一格 */
                g_ni_len--;
            ni_redraw_cursor();
            continue;
        }
        if (cell == NI_CELL_TOFN) {      /* == 0x2b:切到功能列 */
            g_ni_mode |= NI_MODE_FN;     /* or [0x26fc], 4 */
            break;
        }
        g_ni_mode |= NI_MODE_EXIT;       /* 其他(右下角離開格)→ 離開 */
        break;
    }

    g_ni_win_save = save;                /* pop [0x2857] */
}

/* ============================================================
 * ni_zhuyin — 注音鍵盤畫面(sub_0f5b, file 0x22cb)
 * 畫 rect 0x1c4(rec 452:ㄅㄆㄇㄈ… 5×9 注音格 + 聲調 + OK + 功能列),
 * 清組字緩衝,進 grid 導航;依選到的 cell 分四路:
 *   注音/聲調符號格 → 組字(ni_compose_put)
 *   組字完成格      → 查國字 + 寫名字(ni_compose_resolve)
 *   OK 格           → 重置組字繼續
 *   切英數 / 離開
 * ============================================================ */
void ni_zhuyin(void)
{
    u16 save = g_ni_win_save;            /* push [0x2857] */
    u16 cell;

    layout_draw(NI_MSG_ZHUYIN, 0x13, 0x5e); /* DI=0x1c4 畫注音版面 */
    g_ni_win_save = 1;
    win_open(NI_WIN_GRID);               /* 建注音格視窗 [0x42f4] */

restart:
    g_ni_cursor = 0;                     /* [0x26fe]=0 游標回左上 */
    g_ni_grid_w = NI_GRID_W;             /* [0x2702]=9 */
    g_ni_grid_n = NI_GRID_N;             /* [0x2704]=0x2d(45) */
    g_ni_grid_x = NI_GRID_X;             /* [0x2706]=0x15 */
    g_ni_grid_y = NI_GRID_Y;             /* [0x2708]=0x5e */
    g_ni_zhu[0] = g_ni_zhu[1] = 0;       /* 清組字緩衝(聲母/介音) */
    g_ni_zhu[2] = g_ni_zhu[3] = 0;       /* (韻母/聲調) */

    for (;;) {
        ni_grid_nav();                   /* call 0x23f7;結果游標在 [0x26fe],選格時 g_action_result=0 */
        if (g_action_result == 1)        /* 取消 */
            goto leave;

        /* cell = (raw / 9 = row) , (raw % 9 = col) → cell = col*5 + row */
        cell = ni_zhuyin_cell_index();
        g_ni_cell = cell;

        if (cell < NI_CELL_SYMBOL_MAX) { /* < 0x25(37):注音 / 聲調符號 → 組字 */
            ni_compose_put();            /* call 0x2623 放進 [0x2730+group] */
            continue;                    /* 回 grid 繼續選下一個注音 */
        }
        if (cell < NI_CELL_COMMIT_MAX) { /* [0x25,0x2a):★組字完成 → 查國字★ */
            win_close(NI_WIN_GRID);
            if (ni_compose_resolve() != 0)   /* call 0x2654:查無字(beep)→ 重畫格重來 */
                goto reopen;
            ni_redraw_cursor();
            /* 在名字列畫選定國字 @ (0x3e, len*2+0x1f) */
            glyph_draw_pair(g_ni_glyph_lo, g_ni_glyph_hi, 0x3e, g_ni_len * 2 + 0x1f);
            /* 寫進名字緩衝第 g_ni_len 字 */
            g_ni_name[g_ni_len * 2 + 0] = g_ni_glyph_hi;  /* [bx+0x2710] */
            g_ni_name[g_ni_len * 2 + 1] = g_ni_glyph_lo;  /* [bx+0x2712] */
            g_ni_len++;
            if (g_ni_len == NI_NAME_MAX)
                g_ni_len--;
            ni_redraw_cursor();
            goto reopen;                 /* 重開注音格輸入下一字 */
        }
        if (cell == NI_CELL_OK) {        /* == 0x2a:OK 格 → 重畫「輸入注音」框,重置組字 */
            layout_draw(NI_MSG_COMPOSE, 0x13, 0xbe);  /* DI=0x1c8 */
            goto restart;
        }
        if (cell == NI_CELL_TOFN) {      /* == 0x2b:切到功能列 */
            g_ni_mode |= NI_MODE_FN;     /* or [0x26fc], 4 */
            goto leave;
        }
        g_ni_mode |= NI_MODE_EXIT;       /* 其他 → 離開 */
        goto leave;

reopen:
        win_open(NI_WIN_GRID);
        goto restart;
    }

leave:
    win_close(NI_WIN_GRID);
    g_ni_win_save = save;                /* pop [0x2857] */
}

/* ============================================================
 * ni_grid_nav — grid 鍵盤導航迴圈(sub_23f7, file 0x23f7)
 * 反覆讀鍵盤(lcall 1104:8a),方向鍵移游標(含環繞),Space/Enter 選格。
 * 注音 / 英數 / 選取國字 三畫面共用此迴圈(grid 幾何由呼叫端設好)。
 * 回傳:al(0=正常選格 / 1=滑鼠移動取消),並設 g_action_result。
 * ============================================================ */
i16 ni_grid_nav(void)
{
    u8 ah;

    /* (進入:call 0x24e2 初畫游標格;g_action_result=0) */
    g_action_result = 0;

    for (;;) {
        /* call 0x25ac 畫目前游標格高亮 / call 0x2536 滑鼠輪詢 */
        ah = kbd_read_scancode();        /* lcall 1104:8a → AH=scancode */
        if (ah == 0)                     /* 無輸入 → 續輪詢 */
            continue;

        if (ah == NI_SC_MOUSE_MV) {      /* 0x64:滑鼠移動 → 取消本次(al=1) */
            g_action_result = 1;         /* [0x726]=1 */
            return 1;
        }
        if (ah == NI_SC_MOUSE_CLK) {     /* 0x65:滑鼠點擊 → 以滑鼠座標取格 */
            g_ni_mouse_clk = 1;          /* [0x2850]=1 */
            /* call 0x2536 回算命中 index;若 0xff(框外)→ 視為取消 */
            if (/* hit == 0xff */ 0) {
                g_action_result = 1;
                return 1;
            }
            goto select;                 /* 命中 → 選該格 */
        }

        if (ah == NI_SC_UP) {            /* ↑ index -= 寬,環繞到底列 */
            g_ni_cursor -= g_ni_grid_w;
            if ((i16)g_ni_cursor < 0)
                g_ni_cursor += g_ni_grid_n;
            continue;
        }
        if (ah == NI_SC_DOWN) {          /* ↓ index += 寬,環繞回頂列 */
            g_ni_cursor += g_ni_grid_w;
            if (g_ni_cursor >= g_ni_grid_n)
                g_ni_cursor -= g_ni_grid_n;
            continue;
        }
        if (ah == NI_SC_LEFT) {          /* ← index--,環繞 */
            if ((i16)(--g_ni_cursor) < 0)
                g_ni_cursor += g_ni_grid_n;
            continue;
        }
        if (ah == NI_SC_RIGHT) {         /* → index++,環繞 */
            if (++g_ni_cursor >= g_ni_grid_n)
                g_ni_cursor -= g_ni_grid_n;
            continue;
        }

        if (ah == NI_SC_SPACE || ah == NI_SC_ENTER) {  /* ★ Space/Enter = 選取目前格 ★ */
        select:
            /* call 0x25ac 收游標 */
            g_action_result = 0;         /* [0x726]=0 = 正常選取 */
            return 0;                    /* al=0 */
        }
        /* 其他鍵忽略,續迴圈 */
    }
}

/* ============================================================
 * ni_compose_put — 把目前注音符號格放進組字緩衝(sub_2623, file 0x2623)
 * 查 cell→碼對照表 [0x2758]:byte = (group<<5)|code。
 *   group 0 聲母 → [0x2730]   group 1 介音 → [0x2731]
 *   group 2 韻母 → [0x2732]   group 3 聲調 → [0x2733]
 * 並在「輸入注音」框畫出該注音符號(glyph = 0x41 + cell;= D3TXT00 注音字模區 65..)。
 * ============================================================ */
void ni_compose_put(void)
{
    u8 enc = g_ni_codetab[g_ni_cell];    /* [bx(=cell)+0x2758] */
    u8 code  = enc & 0x1f;               /* 低 5 bit = 注音碼 */
    u8 group = enc >> 5;                 /* 高 3 bit = 組(聲母/介音/韻母/聲調) */

    g_ni_zhu[group] = code;              /* [group+0x2730] = code */

    /* 在組字框畫該注音符號:glyph = cell + 0x41(= 注音字模起始 65),
     * 位置 X = group*2 + 0x15,Y = 0xce */
    glyph_draw(g_ni_cell + 0x41, group * 2 + 0x15, 0xce);  /* lcall 111b:6 */
}

/* ============================================================
 * ni_compose_resolve — 組字 → 查國字(sub_2654, file 0x2654)
 * 把組字緩衝(聲母/介音/韻母/聲調)交給注音引擎 11c4:0x27 查碼。
 *   回傳國字碼 DX:AX → [0x2754]/[0x2756];BL = 同音候選字數。
 * - 候選數 0:查無字 → beep(si=0x1e1),回傳 1(失敗,呼叫端重來)。
 * - 候選數 >0:開「選取國字」候選窗 [0x4310],把同音字逐個畫出(11c4:0x81 取字),
 *   再用同一個 grid 導航迴圈讓玩家挑;選定索引交 11c4:0x97 算最終碼。回傳 0(成功)。
 * ============================================================ */
u8 ni_compose_resolve(void)
{
    u8 n_cand;
    int i;

    /* 11c4:0x27:in AH=聲母 AL=介音 BH=韻母 BL=聲調 → out DX:AX 碼、BL 候選數 */
    n_cand = (u8)ime_lookup(g_ni_zhu[0], g_ni_zhu[1], g_ni_zhu[2], g_ni_zhu[3],
                            &g_ni_glyph_lo, &g_ni_glyph_hi);

    if (n_cand == 0) {                   /* 查無同音字 */
        beep_msg(0x1e1);                 /* si=0x1e1 提示 */
        return 1;                        /* 失敗 */
    }

    /* 候選窗:寬 = 候選數(上限 15) */
    if (n_cand > 0xf)
        n_cand = 0xf;
    win_open(NI_WIN_PICK);               /* [0x4310] 「選取國字」窗(rec 454) */
    g_ni_grid_n = n_cand;                /* [0x2704]=候選數 */
    g_ni_grid_w = n_cand;                /* [0x2702]=候選數(單列) */
    g_ni_cursor = 0;                     /* [0x26fe]=0 */

    for (i = 0; i < n_cand; i++) {       /* 逐候選字取碼 + 畫到候選窗 */
        ime_get_candidate(g_ni_zhu);     /* 11c4:0x81 取第 i 個同音字到 [0x2734] */
        ni_draw_compose();               /* call 0x2703 畫該候選字 */
        g_ni_cursor++;
    }

    g_ni_cursor = 0;
    g_ni_grid_x = NI_GRID_X;             /* [0x2706]=0x15 */
    g_ni_grid_y = 0xce;                  /* [0x2708]=0xce(候選窗 Y) */
    /* [0x4324]=候選數 */

    ni_grid_nav();                       /* call 0x23f7:玩家挑一個候選字 */

    /* 選定索引 [0x26fe] → 國字碼 += idx<<5(進到引擎的碼空間),11c4:0x97 定案 */
    {
        u32 add = (u32)g_ni_cursor << 5;
        u32 code = ((u32)g_ni_glyph_lo) | ((u32)g_ni_glyph_hi << 16);
        code += add;
        g_ni_glyph_lo = (u16)(code & 0xffff);
        g_ni_glyph_hi = (u16)(code >> 16);
    }
    ime_finalize();                      /* 11c4:0x97 */

    win_close(NI_WIN_PICK);
    return 0;                            /* 成功 */
}

/* ============================================================
 * ni_draw_compose — 在「輸入注音」/ 候選框畫一個候選字(sub_2703, file 0x2703)
 * X = cursor*2 + 0x15,Y = 0xce。
 * ============================================================ */
void ni_draw_compose(void)
{
    glyph_draw_buf((void *)0x2734, g_ni_cursor * 2 + 0x15, 0xce);  /* lcall 111b:6c, si=[0x2734] */
}

/* ---- 下列為反組譯內聯的小工具,集中以函式表達(非獨立 sub_,僅為可讀性) ---- */

/* 注音 grid:raw=[0x26fe];div 9 → al=row, ah=col;cell = col*5 + row(sub_0f5b 0x2337..) */
static u16 ni_zhuyin_cell_index(void)
{
    u16 raw = g_ni_cursor;
    u8  row = (u8)(raw / NI_GRID_W);
    u8  col = (u8)(raw % NI_GRID_W);
    return (u16)(col * 5 + row);
}

/* 英數 grid:同算式(sub_0e8e 0x222d..),div 寬後 col*5+row */
static u16 ni_alnum_cell_index(void)
{
    u16 raw = g_ni_cursor;
    u8  row = (u8)(raw / g_ni_grid_w);
    u8  col = (u8)(raw % g_ni_grid_w);
    return (u16)(col * 5 + row);
}

/* glyph_draw 變體:畫一個雙 word 國字碼 / 一段碼緩衝(對應 lcall 111b:6 / 111b:6c)。
 * 此處僅宣告語意;實際是 runtime 文字繪製器入口。 */
extern void glyph_draw_pair(u16 lo, u16 hi, u16 x, u16 y);
extern void glyph_draw_buf(void *buf, u16 x, u16 y);
