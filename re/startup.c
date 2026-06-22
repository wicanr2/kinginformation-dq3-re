/* startup.c — 精訊 DQ3.EXE 啟動 / 初始化骨架(real-mode large-model C)
 *
 * 反編譯來源:tools/re_recdis.py 遞迴反組譯;對照 docs/08-exe-functions.md。
 * 涵蓋 entry(sub_9299)→ game_body(sub_92f0)→ 初始化各環節 → 進入 main_loop。
 *
 * 求結構正確、可審閱;非位元精準可重編譯版。far-call runtime 以 exe.h wrapper 表達,
 * 暫存器式子功能碼(AH:AL)以參數/註解標出。逐函式對 DOSBox 驗證為後續步驟。
 */
#include "exe.h"

/* 程式入口,對應 CS:IP = 0000:9299(file 0xa609)。
 * 一連串 runtime init + 設主資料段 DS,呼叫遊戲主體後以 int 21h/AH=4Ch 結束。 */
void entry(void)
{
    /* sub_02c4:早期環境設定(near call,16-bit 目標 wrap) */
    early_setup();                 /* sub_02c4 */

    /* runtime init:DOS / 計時 / 鍵盤 */
    /* lcall 0fe1:08 ; lcall 100a:16e ; lcall 1104:0d */
    dos_init();                    /* lcall 0fe1:08 */
    timer_init();                  /* lcall 100a:16e */
    kbd_init();                    /* lcall 1104:0d */

    /* mov ds, 0x14dd —— 設主資料段(原始碼此後所有 DS:off 全域生效) */
    /* (C 還原為直接存取 .data 全域,DS 載入由 startup runtime 負責) */

    kbd_set_mode(0x19, 0x00);      /* lcall 1104:132,AH=0x19 掛載中斷向量 */

    game_body();                   /* sub_92f0:遊戲主體;return 後即結束 */

    /* 收尾:還原視訊/滑鼠/視訊模式/檔案/計時/鍵盤,再 DOS 結束 */
    mouse_op(0x3a);                /* lcall 10b6:3a */
    vid_mode(0x35);                /* lcall 109c:35 */
    vid_mode(0x28);                /* lcall 109c:28 */
    file_op_dx(0x95);              /* lcall 1053:95 */
    dos_cleanup();                 /* lcall 0fe1:29 */
    timer_restore();               /* lcall 100a:18f */
    kbd_restore();                 /* lcall 1104:2e */
    kbd_set_mode(0x02, 0x00);      /* lcall 1104:132,AH=2 還原中斷 */

    dos_exit(0);                   /* mov ah,4Ch; int 21h —— 程式結束 */
}

/* sub_92f0(file 0xa660)= 遊戲主體。
 * 完成所有初始化後進入 main_loop;main_loop return 後本函式 ret → entry 收尾結束。 */
void game_body(void)
{
    video_file_init();             /* sub_934f:VGA A000、載入 item/d3mns/setup/字型 */

    gfx_init(0x0e);                /* lcall 1209:0e */
    file_init();                   /* lcall 1053:0c */
    file_op_dx(0x00);              /* lcall 1053:154,DX=0 */
    dos_op(0x13d);                 /* lcall 0fe1:13d */
    gfx_init(0x60);                /* lcall 1209:60 */

    gfx_palette_init();            /* sub_9387:GFX 暫存器 + 調色盤 + BLK */

    mouse_op(0x385);               /* lcall 10b6:385 */
    dos_op2(0x04);                 /* lcall 11d3:04 */

    /* DS:0x25d1 = 0x3352 + 9*0x30(版面/視窗基準計算) */
    /* lea dx,[0x3352]; ax=9; mul 0x30; add dx,ax; [0x25d1]=dx */
    g_win_base = 0x3352 + 9 * 0x30;   /* DS:0x25d1 */

    vid_mode(0x7a);                /* lcall 109c:7a */
    mouse_op(0x3a0);               /* lcall 10b6:3a0 */

    cstartup();                    /* sub_0000:C runtime startup */

    /* 清計時/中斷計數旗標(BIOS tick 區) */
    g_tick_a = 0;                  /* [0x13]=0 ; sti */
    g_tick_b = 0;                  /* [0x07]=0 ; cli */

    seed_and_names();              /* sub_6f4b:亂數種子 + 玩家名稱表 */

    main_loop();                   /* sub_93e3:主迴圈;return 後結束 */
}

/* sub_934f:影片 / 檔案系統 / 字型載入。設 ES=A000(VGA),載入主要資料檔。 */
void video_file_init(void)
{
    /* mov ax,0xa000; mov es,ax —— ES = VGA 顯示記憶體 */
    file_op_dx(0x172);             /* lcall 1053:172 */
    file_op_dx(0x3fb);             /* lcall 1053:3fb */
    file_op_dx(0x377);             /* lcall 1053:377 */

    load_setup();                  /* sub_997d:setup.dat */
    init_misc_e96d();              /* sub_e96d */
    init_misc_e932();              /* sub_e932 */
    vid_mode(0x0e);                /* lcall 109c:0e */
    load_text();                   /* sub_1811:d3txt00.fon + d3txt00.txt */

    g_flag_258e = 1;               /* [0x258e]=1 */
    load_blkbm();                  /* sub_ecc0:blkbm.dat */
    load_item();                   /* sub_17cb:item.dat */
    load_menu_data();              /* sub_17ee:d3mns.dat */
}

/* sub_9387:圖形 / 調色盤 / BLK 初始化。直接寫 VGA GFX 暫存器(0x3ce)。 */
void gfx_palette_init(void)
{
    vid_mode(0x47);                /* lcall 109c:47 */
    vid_mode(0x144);               /* lcall 109c:144 */
    load_palette();                /* sub_ecdc:dq3.pal/mnsbk.pal */

    /* VGA Graphics Controller 設定(out 0x3ce):
     *   reg3=0x00(Data Rotate),reg5=0x08(Mode),reg7=0x00,reg8=0xff(Bit Mask) */
    vga_gfx_out(0x0003);           /* out 0x3ce, ax=0x0003 */
    vga_gfx_out(0x0805);           /* out 0x3ce, ax=0x0805 */
    vga_gfx_out(0x0007);           /* out 0x3ce, ax=0x0007 */
    vga_gfx_out(0xff08);           /* out 0x3ce, ax=0xff08 */

    init_e6a3();                   /* sub_e6a3 */
    dos_op3(0x14);                 /* lcall 11c4:14 */
    init_ed39();                   /* sub_ed39 */
    init_2ddf();                   /* sub_2ddf */
    load_bls();                    /* sub_ec9e:dq3mst.bls/dq3man.bls */
    load_blk();                    /* sub_eaf4:dq3.blk/dq31.blk */

    mouse_op(0x08);                /* lcall 10b6:08 */
    mouse_op(0x385);               /* lcall 10b6:385 */
    vid_mode(0x144);               /* lcall 109c:144 */

    /* [0x4f09]=0x44f2; [0x4f0b]=[0x4f09](視窗座標/捲動基準對) */
    g_view_a = 0x44f2;             /* [0x4f09] */
    g_view_b = g_view_a;           /* [0x4f0b] */

    vid_mode(0x105);               /* lcall 109c:105 */
    mouse_op(0x3a0);               /* lcall 10b6:3a0 */
}

/* sub_0000(file 0x1370,seg0 起點)= large-model C runtime startup。
 * 設旗標、環境/argv 設定、呼叫主初始化 sub_0030。 */
void cstartup(void)
{
    g_crt_flag_72a = 1;            /* [0x72a]=1 */
    env_init();                    /* call 0xf4e3:環境/argv 設定 */

    if (g_crt_726 == 1 || g_crt_722 == 1) {   /* 已初始化 → 直接主初始化 */
        main_init();               /* sub_0030 */
        return;
    }
    crt_setup_14da();              /* call 0x14da */
    g_crt_flag_72a = 0;            /* [0x72a]=0 */
    if (g_crt_726 < 1)             /* jae 0x2c 的補集 */
        return;
    main_init();                   /* sub_0030 */
}

/* sub_0030:主初始化,設預設座標 / 視窗常數到主資料段。 */
void main_init(void)
{
    g_var_4f2f = 0x99;             /* [0x4f2f]=0x99 */
    g_var_4f31 = 0xae;             /* [0x4f31]=0xae */
    g_var_4f48 = 0x99;             /* [0x4f48]=0x99 */
    g_var_4f4a = 0xae;             /* [0x4f4a]=0xae */
    g_var_4f50 = 0x04;             /* [0x4f50]=4 */
    g_var_256a = 0x04;             /* [0x256a]=4 */
    /* … 其餘版面/狀態常數初始化(sub_0030 共 ~219 bytes,此處摘關鍵設定) … */
}

/* sub_6f4b(file 0x82bb)= 亂數種子 + 玩家名稱表初始化。
 * 讀 BIOS 資料區 0000:046C(系統 tick 計數器)當亂數種子;以名稱表填角色名。 */
void seed_and_names(void)
{
    u16 seed;
    u8  idx;

    /* push ds; ds=0; si=0x46c; ax=[si]; pop ds —— 讀 BIOS tick @ 0040:006C */
    seed = bios_tick();            /* 0000:046C */
    rng_state = seed;              /* cs:[0x701b]=ax */

    do {
        idx = (u8)(rng_next() & 0x7f);   /* call 0x701d 取亂數,& 0x7f */
    } while (idx > 0x63);          /* 範圍限制 0..0x63 */
    g_name_index = idx;            /* [0x9f1]=al */

    /* [0x8c0] + idx*3 取 3 byte 名稱碼 → [0x9ec..0x9ee],再逐一 call 0x6fde 展開 */
    {
        u8 *p = &name_table[idx * 3];   /* lea si,[0x8c0]+ax */
        g_name0 = p[0];            /* [0x9ec] */
        g_name1 = p[1];            /* [0x9ed] */
        g_name2 = p[2];            /* [0x9ee] */
        name_expand(g_name0);      /* call 0x6fde */
        name_expand(g_name1);
        name_expand(g_name2);
    }
}
