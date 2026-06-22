/* render.c — 場景繪製、捲動、CTY 版面解析、HUD/事件 dispatcher(RE → C)
 *
 * 反編譯來源(tools/re_disasm.py,capstone,docker 內):
 *   draw_scene        sub_255b  file 0x38cb   (地表場景 gate + 方向 dispatch + 捲動)
 *   cty_select_section sub_30cf file 0x443f   (CTY section → 版面/出口/起始座標)
 *   update_hud        sub_9530  file 0xa8a0   (狀態旗標逐位元 dispatch + NPC + HUD)
 *
 * 對照 docs/11-exe-gameplay.md。real-mode large-model C;far runtime 以
 * exe.h / assets.h 的 wrapper 表達。求結構正確可審閱,非位元精準。
 *
 * 命名澄清:exe.h 把 sub_255b 命名 draw_scene、sub_19b8 命名 update_player,
 * 但兩者其實是「同結構的兩個場景處理常式」:sub_255b 處理地表([0x4f2d]==0)、
 * sub_19b8 處理城鎮([0x4f2d]==1)。主迴圈每幀都呼叫兩者,只有與目前場景旗標
 * 相符的那個會實際工作,另一個第一行 gate 不符就 ret。沿用既有命名。
 */
#include "exe.h"
#include "assets.h"
#include "gameplay.h"

/* ---- 本檔用到、尚未在 exe.h/assets.h 出現的 runtime / 內部常式(forward) ---- */
extern void scene_redraw_full(void);    /* sub 0x3eae(file 0x521e)整屏重繪 */
extern void scene_finalize_a(void);     /* call 0xf003(wrap)場景收尾 a */
extern void scene_finalize_b(void);     /* call 0xf8b4(wrap)場景收尾 b */
extern void map_to_town_setup(void);    /* call 0x4289 城鎮初始化(scene>2)*/
extern void over_load_chunk(void);      /* call 0x396e 地表分塊載入(attr&0x1000)*/
extern void cty_attr_post(u16 ax);      /* sub 0x2bd5  section 後處理 */
extern void npc_update_one(u16 bx);     /* sub 0xa9be  逐 NPC 更新 */
extern void hud_event_tile(void);       /* sub 0xaa42  tile 觸發事件([0x4f46]&0x800)*/
extern void hud_timer_a(void);          /* sub 0xaa7e  計時器 a 到期 */
extern void hud_timer_b(void);          /* sub 0xaaca  計時器 b */
extern void hud_timer_c(void);          /* sub 0xab80  計時器 c */
extern void hud_anim_step(void);        /* sub 0xab18  tile 動畫([0x2577] hi nibble)*/
extern void npc_commit(void);           /* sub 0xd966 / 0xd062 NPC 提交 */
extern void hud_misc(void);             /* sub 0xd107 */

/* ================================================================= */
/* draw_scene —— sub_255b(地表場景處理:gate + 方向 dispatch + 捲動 + 過場)
 *
 * 結構與城鎮版 update_player(sub_19b8)逐行對應,差別只在 gate 值(地表 0)
 * 與方向 jump table 位址(地表 @0x2562、城鎮 @0x255a)。
 *
 * 流程:
 *   1. 若 [0x4f2d] != 0(非地表)→ 直接 ret(本幀交給城鎮版)。
 *   2. call 0x3ac8(地表互動/前置)。
 *   3. 若有移動方向([0x4f1f]!=0xff):
 *        - 設「移動中」狀態 [0x2784](0xff→2,事後→1):捲動狀態旗標。
 *        - 依方向經 jump table [0x4f1f*2 + 0x2562] 呼叫四個捲動 handler 之一。
 *        - scene_finalize_a/b 收尾。
 *   4. [0x4f3b] > 2 → call 0x4289(進階城鎮/室內初始化)。
 *   5. 無方向時做整屏重繪(scene_redraw_full)。
 *   6. 過場處理([0x4f2d]==2:地表→城鎮本幀切換):
 *        若 [0x2577]&0x1000(目標 tile 屬性帶城鎮旗標)→ over_load_chunk();
 *        否則切 [0x4f2d]=0→載入→=2,再 cty 流程;最終 [0x4f2d]==0 時關閉
 *        並重開資源(lcall 1053:240 / 1053:dc / call 0xa9 / call 0x2c70 /
 *        lcall 1104:123)。
 */
void draw_scene(void)
{
    if (g_scene_flag != SCENE_OVERWORLD)          /* cmp [0x4f2d],0; jne ret */
        return;

    over_interact_pre();                          /* call 0x3ac8 */

    if (g_move_dir != DIR_NONE) {                 /* [0x4f1f]!=0xff */
        if (g_scroll_state == 0xff)               /* [0x2784] */
            g_scroll_state = 2;                   /* 進入捲動 */

        /* 依方向呼叫捲動 handler:call word ptr [bx*2 + 0x2562] */
        over_scroll_dispatch[g_move_dir]();

        if (g_scroll_state == 0xff)
            g_scroll_state = 1;                   /* 捲動完成標記 */

        scene_finalize_a();                       /* call 0xf003 */
        scene_finalize_b();                       /* call 0xf8b4 */
    }

    if (g_scene_substate > 2)                     /* [0x4f3b]>2 */
        map_to_town_setup();                      /* call 0x4289 */

    if (g_move_dir == DIR_NONE) {                 /* 靜止:整屏重繪 */
        scene_redraw_full();                      /* call 0x3eae */
        scene_finalize_b();                       /* call 0xf8f8 */
    }

    /* 過場:地表→城鎮(本幀切換)*/
    if (g_scene_flag == SCENE_TO_TOWN) {          /* [0x4f2d]==2 */
        if (g_tile_attr & 0x1000) {               /* 目標屬性帶「城鎮入口」位 */
            over_load_chunk();                    /* call 0x396e */
        } else {
            g_scene_flag = SCENE_OVERWORLD;       /* 暫切 0 重載資源 */
            scene_reload();                       /* call 0xf147(wrap)*/
            g_scene_flag = SCENE_TO_TOWN;
            cty_enter();                          /* call 0x2c37 */
            if (g_scene_flag == SCENE_OVERWORLD) {/* 切換完成 */
                /* 關閉舊資源並重開新場景(檔案/音效/鍵盤再掛載)*/
                file_op_dx(0x240);                /* lcall 1053:240 */
                file_op_dx(0x0dc);                /* lcall 1053:dc */
                misc_reopen();                    /* call 0xa9 */
                map_reopen();                     /* call 0x2c70 */
                kbd_set_mode(1, 0x23);            /* lcall 1104:123 */
            }
        }
    }
}

/* ---- draw_scene 用到、仍以位址引用的近常式(forward,細節屬其他函式)---- */
extern void over_interact_pre(void);   /* sub 0x3ac8  地表互動前置 */
extern void scene_reload(void);        /* call 0xf147 */
extern void cty_enter(void);           /* sub 0x2c37  進城流程 */
extern void misc_reopen(void);         /* sub 0x00a9 */
extern void map_reopen(void);          /* sub 0x2c70 */
extern u16  g_scroll_state;            /* DS:0x2784  捲動狀態(0xff/1/2)*/
extern u8   g_scene_substate;          /* DS:0x4f3b  場景子狀態 */

/* 地表方向捲動 jump table(DS:0x2562,dir 0..3)。函式指標表,移植成函式陣列。*/
void (*const over_scroll_dispatch[4])(void) = {
    over_scroll_down,   /* 0x2880 */
    over_scroll_left,   /* 0x2904 */
    over_scroll_up,     /* 0x2977 */
    over_scroll_right,  /* 0x29e7 */
};

/* ================================================================= */
/* cty_select_section —— sub_30cf(file 0x443f)
 *
 * 解開 docs/04 留的 CTY 城鎮版面謎團。CTY 整檔已由 load_cty 載入 seg_cty
 * (DS:0x2536)。本函式:
 *   1. 由 [0x256a](section index)查檔首 section 偏移表:
 *        sect_off = ((u16*)cty_base)[g_cty_section]      →  [0xb24]
 *   2. section 內 +0x0e 取版面指標 layout_ptr,加 sect_off 得版面起點 si;
 *        讀 width=[si]→[0xb28]、height=[si+2]→[0xb2a]、
 *        起始座標 spawn_x=[di]、spawn_y=[di+2..](寫 [0x4f33]/[0x4f35]);
 *        si+=4 後 [0xb26]=si = tile 陣列基底(u16/tile,row-major)。
 *   3. 取玩家所在 tile 的屬性:idx=(y*width+x); tile=base[idx];
 *        [0x2576]=低 byte(BLK index)、[0x2579]=高 byte&0xc0(角落屬性)。
 *   4. section +0x15 取出口/warp 4 bytes → [0xb56/0xb57/0xb58/0xd71]。
 *   5. section +0x10 取旗標 word:bit15 → 設/清場景旗標 [0x4f46] bit1。
 *   6. call 0x2bd5 後處理(section attr → 繪製前置)。
 *
 * 結論:**CTY 的 tile 佈局不是壓縮/RLE,而是每格一個 u16(低=tile index、
 * 高=屬性),前置 4-byte 版面表頭(w,h,spawn)**。docs/04 之所以對不上,是
 * 因為先前把整段當「壓縮命令串」解;實際要先用 section 偏移表 + 0xe-byte
 * section 標頭 + 0x10 旗標 + 0x15 出口表,才到達 layout_ptr 指向的 w/h/tiles。
 */
void cty_select_section(void)
{
    u16 sect_off;       /* [0xb24]  本 section 在 CTY 檔內的 byte 偏移 */
    u16 layout_ptr;     /* section+0x0e:版面資料指標(相對 sect_off)*/
    u16 si;             /* 版面/tile 指標(seg_cty 內偏移)*/
    u16 di;
    u16 tile;
    i16 w, h, x, y;

    /* (1) section 偏移表:cty_base[g_cty_section] */
    sect_off = ((u16 far *)MK_FP(seg_cty, 0))[g_cty_section];  /* shl bx,1; ds=[0x2536]; [bx] */
    /* 原碼存 [0xb24] = sect_off */

    /* (2) +0x0e 版面指標 → 版面起點 */
    di = sect_off + 0x0e;
    layout_ptr = *(u16 far *)MK_FP(seg_cty, di);  /* mov si,[di] */
    si = layout_ptr + sect_off;                   /* add si,bx */

    w = *(u16 far *)MK_FP(seg_cty, si);           /* width  → [0xb28] */
    h = *(u16 far *)MK_FP(seg_cty, si + 2);       /* height → [0xb2a] */
    di += 2;
    /* spawn:di 處取 dl,(di+2)取 dh=spawn_x_hi? 原碼:dl=[di],dh=[di+2],
     * cl=[di+3],ch=[di+4];cl→[0x4f33]起始 X,ch→[0x4f35]起始 Y。*/
    {
        u8 far *p = (u8 far *)MK_FP(seg_cty, di);
        /* [0xb2c]=p[0]; [0xb2d]=p[2]; spawn_x([0x4f33])=p[3]; spawn_y([0x4f35])=p[4] */
        g_pos_x = p[3];
        g_pos_y = p[4];
    }
    si += 4;                                       /* [0xb26] = tile 陣列基底 */
    g_layout_base = si;
    g_map_w = w;
    g_map_h = h;

    /* (3) 玩家所在 tile 的屬性 */
    x = g_pos_x;  y = g_pos_y;
    {
        u16 idx = (u16)(y * w + x);
        u8 far *t = (u8 far *)MK_FP(seg_cty, si + idx * 2);
        g_corner_attr = t[1] & 0xc0;               /* [0x2579] */
        g_tile_cur    = t[0];                      /* [0x2576] = BLK index(低 byte)*/
    }

    /* (4) 出口/warp 表(section+0x15)*/
    di = sect_off + 0x15;
    {
        u8 far *p = (u8 far *)MK_FP(seg_cty, di);
        /* [0xb56]=p[0]; [0xb57]=p[1]; [0xb58]=p[2]; [0xd71]=p[3] */
        (void)p;
    }

    /* (5) section 旗標(section+0x10),bit15(test al,0x80 of low byte)*/
    di = sect_off + 0x10;
    {
        u8 far *p = (u8 far *)MK_FP(seg_cty, di);
        u8 flag = p[0];                            /* [0xb5c] */
        if (flag & 0x80) g_status_evt |= 2;        /* or [0x4f46],2 */
        else             g_status_evt &= 0xfffd;
        /* [0xd77] = p[1] */
    }

    /* (6) 後處理 */
    cty_attr_post(0 /* ax=[0xb58] */);             /* call 0x2bd5 */
}

/* ================================================================= */
/* update_hud —— sub_9530(file 0xa8a0)
 *
 * 與其說是「HUD」,實為**移動後事件 + NPC 更新 + 狀態列重繪的綜合 pass**:
 *   1. [0x4f46]&4(menu/redraw 請求)→ lcall 11dd:c(選單/特殊重繪),清 bit2。
 *   2. 無移動([0x4f1f]==0xff)→ ret。
 *   3. [0x4f3b]<2 → ret;[0x4f44]&0x10(對話/鎖定中)→ ret。
 *   4. 清狀態列暫存 [0x662..0x665]=0(4 bytes)。
 *   5. 逐位元 dispatch 事件旗標 bp=[0x4f46]:
 *        &0x800 → hud_event_tile()(踩到觸發 tile),ret;
 *        &0x20|0x40 → 倒數 [0x52f6];歸 0 → hud_timer_a();
 *        &0x8000|0x4000 → 倒數 [0x52f7];歸 0 → hud_timer_b();
 *        &0x200 → 倒數 [0x52f8];歸 0 → hud_timer_c();
 *        &4 → lcall 11dd:c;清 bit2;[0x13]|=1。
 *      另:[0x50b7]&4 → 倒數 [0x504e];<=0 → call 0x5474。
 *   6. tile 動畫:[0x2577] 高 nibble!=0 → hud_anim_step(),否則清 [0x4f46] bit10。
 *   7. NPC 迴圈:cx=[0x5077] 隻,逐隻 npc_update_one(bx+=2);若有變更
 *      ([0x24b4]!=0)→ npc_commit() 並(視 bp)重繪角色名(lcall 111b:264)、
 *      重設場景旗標 [0x4f2d]=1;否則 hud_misc()。
 */
void update_hud(void)
{
    u16 bp;     /* = [0x4f46] 快取 */
    u16 bx;
    int i, n;

    if (g_status_evt & 4) {                /* test [0x4f46],4 */
        menu_special();                    /* lcall 11dd:c */
        g_status_evt &= 0xfffb;            /* 清 bit2 */
    }
    if (g_move_dir == DIR_NONE) return;    /* [0x4f1f]==0xff */
    if (g_scene_substate < 2) return;      /* [0x4f3b]<2 */
    if (g_lock_flag & 0x10) return;        /* [0x4f44]&0x10 */

    /* 清狀態列暫存(4 bytes)*/
    g_statline[0] = g_statline[1] = g_statline[2] = g_statline[3] = 0;

    bp = g_status_evt;                     /* [0x4f46] */

    if (bp & 0x800) {                      /* 踩到觸發 tile */
        hud_event_tile();                  /* sub 0xaa42 */
        return;
    }
    if ((bp & 0x20) || (bp & 0x40)) {
        if (--g_timer_a == 0)              /* [0x52f6] */
            hud_timer_a();                 /* sub 0xaa7e */
    }
    if ((bp & 0x8000) || (bp & 0x4000)) {
        if (--g_timer_b == 0)              /* [0x52f7] */
            hud_timer_b();                 /* sub 0xaaca */
    }
    if (bp & 0x200) {
        if (--g_timer_c == 0)              /* [0x52f8] */
            hud_timer_c();                 /* sub 0xab80 */
    }
    if (bp & 4) {
        menu_special();                    /* lcall 11dd:c */
        g_status_evt &= 0xfffb;
        g_cmd_pending |= 1;                /* [0x13]|=1 */
    }
    if (g_anim_flag & 4) {                 /* [0x50b7]&4 */
        if (--g_anim_count <= 0)           /* [0x504e] */
            anim_expire();                 /* call 0x5474 */
    }

    /* tile 動畫:目前 tile 屬性高 nibble */
    if ((g_tile_attr >> 8) & 0x0f)         /* test ah,0xf of [0x2577] */
        hud_anim_step();                   /* sub 0xab18 */
    else
        g_status_evt &= 0xfbff;            /* 清 bit10 */

    /* NPC 迴圈:[0x5077] 隻 */
    n = g_npc_count;                       /* [0x5077] */
    g_npc_dirty = 0;                       /* [0x24b4] */
    for (i = 0, bx = 0; i < n; i++, bx += 2)
        npc_update_one(bx);                /* sub 0xa9be */

    if (g_npc_dirty != 0) {
        npc_commit();                      /* sub 0xd966 */
        if (bp != 0) {
            commit_more();                 /* sub 0xd062 */
        } else {
            text_set_src();                /* call 0x6372 */
            draw_string(0x169);            /* lcall 111b:264(角色名/HUD 文字)*/
            text_restore();                /* call 0x6380 */
            io_flush();                    /* call 0x100a9 */
            g_scene_flag = SCENE_TOWN;     /* [0x4f2d]=1 */
            town_post_a();                 /* call 0xd3af */
            town_post_b();                 /* call 0xdb49 */
        }
    } else {
        hud_misc();                        /* sub 0xd107 */
    }
}

/* ---- update_hud 用到、以位址引用的旗標/常式(forward)---- */
extern u8  g_scene_substate2; /* 同 0x4f3b,沿用 g_scene_substate */
extern u16 g_lock_flag;       /* DS:0x4f44 */
extern u8  g_statline[4];     /* DS:0x0662  狀態列暫存 4 bytes */
extern u8  g_timer_a;         /* DS:0x52f6 */
extern u8  g_timer_b;         /* DS:0x52f7 */
extern u8  g_timer_c;         /* DS:0x52f8 */
extern u16 g_cmd_pending;     /* DS:0x0013 */
extern u16 g_anim_flag;       /* DS:0x50b7 */
extern i8  g_anim_count;      /* DS:0x504e */
extern u8  g_npc_count;       /* DS:0x5077 */
extern u8  g_npc_dirty;       /* DS:0x24b4 */
extern void anim_expire(void);  /* call 0x5474 */
extern void commit_more(void);  /* sub 0xd062 */
extern void text_set_src(void); /* call 0x6372 */
extern void draw_string(u16 di);/* lcall 111b:264 包裝 */
extern void text_restore(void); /* call 0x6380 */
extern void io_flush(void);     /* call 0x100a9 */
extern void town_post_a(void);  /* call 0xd3af */
extern void town_post_b(void);  /* call 0xdb49 */
