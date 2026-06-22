/* player.c — 城鎮場景:玩家移動、碰撞、觸發(RE → C)
 *
 * 反編譯來源(tools/re_disasm.py,capstone,docker 內):
 *   update_player        sub_19b8  file 0x2d28  (城鎮場景 gate + 方向 dispatch + 捲動)
 *   player_move_collide  sub_2dda  file 0x414a  (算目標 tile、查屬性、決定可否移動)
 *   scene_event_poll     sub_2f01  file 0x4271  ([0x4f46]&1 觸發事件)
 *
 * 對照 docs/11-exe-gameplay.md。real-mode large-model C;far runtime 以
 * exe.h / assets.h wrapper 表達。求結構正確可審閱,非位元精準。
 *
 * 命名澄清見 render.c 開頭:sub_19b8(exe.h 命名 update_player)其實是
 * 「城鎮場景處理常式」,與地表版 draw_scene(sub_255b)同結構;方向 jump
 * table 在 DS:0x255a(地表版在 0x2562)。
 */
#include "exe.h"
#include "assets.h"
#include "gameplay.h"

/* ---- forward(以位址引用的近常式/旗標)---- */
extern void scene_finalize_a(void);   /* call 0xf003 */
extern void scene_finalize_b(void);   /* call 0xf8b4 / 0xf8f8 */
extern void scene_reload(void);       /* call 0xf147 */
extern void scene_redraw_full(void);  /* sub 0x32be(file 0x462e)城鎮整屏重繪 */
extern void cty_enter(void);          /* (城鎮→地表收尾)*/
extern void misc_reopen(void);        /* sub 0x00a9 */
extern void over_reopen_a(void);      /* sub 0x3823 */
extern void over_reopen_b(void);      /* sub 0x3866 */
extern u16  g_scroll_state;           /* DS:0x2784 */
extern u8   g_scene_substate;         /* DS:0x4f3b */
extern u8   g_warp_pending;           /* DS:0x257b  出口/warp 待處理旗標 */
extern u16  g_status_evt;             /* DS:0x4f46 */
extern void warp_begin(void);         /* lcall 10b6:385 */
extern void warp_load(void);          /* sub 0x3685 */
extern void warp_apply(void);         /* sub 0x2cdb */
extern void warp_end(void);           /* lcall 10b6:3a0 */

/* 城鎮方向捲動 jump table(DS:0x255a,dir 0..3)*/
void (*const town_scroll_dispatch[4])(void) = {
    town_scroll_down,   /* 0x1ba2 */
    town_scroll_left,   /* 0x1c26 */
    town_scroll_up,     /* 0x1c99 */
    town_scroll_right,  /* 0x1d09 */
};

/* ================================================================= */
/* update_player —— sub_19b8(城鎮場景處理)
 *
 * 流程(與 draw_scene 鏡像,gate 值 / jump table 不同):
 *   1. [0x4f2d] != 1(非城鎮)→ ret。
 *   2. 有移動方向([0x4f1f]!=0xff):
 *        - call 0x2dda(player_move_collide):算目標 tile + 碰撞;若被擋,
 *          函式內會把 [0x4f1f]=0xff(取消本幀移動)。
 *        - 若移動仍成立:設捲動狀態 [0x2784];依方向經 jump table
 *          [0x4f1f*2 + 0x255a] 呼叫四個城鎮捲動 handler 之一;收尾 a/b。
 *        - warp 待處理([0x257b]==1)→ warp_begin/load/apply/end。
 *        - [0x4f46]&1 → scene_event_poll()(踩事件),跳到結尾。
 *   3. 無移動 / 被擋:若仍 [0x4f1f]==0xff 且非 warp → 城鎮整屏重繪。
 *   4. 過場([0x4f2d]==3:城鎮→地表本幀切換):切 0、重載資源、重開檔/音/鍵。
 */
void update_player(void)
{
    if (g_scene_flag != SCENE_TOWN)               /* cmp [0x4f2d],1; jne ret */
        return;

    if (g_move_dir != DIR_NONE) {                 /* [0x4f1f]!=0xff */
        player_move_collide();                    /* call 0x2dda(可能把 dir 設回 0xff)*/

        if (g_move_dir != DIR_NONE) {             /* 未被擋 */
            if (g_scroll_state == 0xff)
                g_scroll_state = 2;

            town_scroll_dispatch[g_move_dir]();   /* call [bx*2 + 0x255a] */

            if (g_scroll_state == 0xff)
                g_scroll_state = 1;

            scene_finalize_a();                   /* call 0xf003 */
            scene_finalize_b();                   /* call 0xf8b4 */

            if (g_warp_pending == 1) {            /* [0x257b]==1:踩到出口 */
                warp_begin();                     /* lcall 10b6:385 */
                warp_load();                      /* call 0x3685 */
                warp_apply();                     /* call 0x2cdb */
                warp_end();                       /* lcall 10b6:3a0 */
            }
            if (g_status_evt & 1) {               /* [0x4f46]&1:踩到事件 tile */
                scene_event_poll();               /* call 0x2f01 */
                goto done;
            }
        }
    }

    /* 無移動 / 被擋:條件式整屏重繪 */
    if (g_move_dir == DIR_NONE) {
        if (g_warp_pending != 1) {                /* 非 warp 中 */
            scene_redraw_full();                  /* call 0x32be */
            scene_finalize_b();                   /* call 0xf8f8 */
        }
    }

done:
    /* 過場:城鎮→地表(本幀切換)*/
    if (g_scene_flag == SCENE_TO_OVERWORLD) {     /* [0x4f2d]==3 */
        g_scene_flag = SCENE_TOWN;                /* 暫切 1 重載 */
        scene_reload();                           /* call 0xf147 */
        g_scene_flag = SCENE_OVERWORLD;           /* → 地表 */
        file_op_dx(0x240);                        /* lcall 1053:240 */
        file_op_dx(0x0dc);                        /* lcall 1053:dc */
        misc_reopen();                            /* call 0xa9 */
        over_reopen_a();                          /* call 0x3823 */
        over_reopen_b();                          /* call 0x3866 */
        kbd_set_mode(1, 0x23);                    /* lcall 1104:123 */
    }
}

/* ================================================================= */
/* player_move_collide —— sub_2dda(file 0x4154)
 *
 * 城鎮路徑的移動 + 碰撞核心。流程:
 *   1. 目標 tile 座標 = (玩家 tile + 本幀位移):
 *        tx = [0x4f33] + [0x2572]  (Y? 見下;原碼 bx=tile_x + dx)
 *        ty = [0x4f35] + [0x2574]
 *      註:[0x4f33]/[0x4f35] 是玩家當前 tile 座標,[0x2572]/[0x2574] 是本幀
 *      ±1 位移(來自主迴圈方向鍵)。
 *   2. 邊界檢查:tx∈[0,[0xb28]) 且 ty∈[0,[0xb2a])。出界 → 設過場
 *        [0x4f2d]=3(城鎮→地表)、[0x2576]=[0xb2d](邊界 tile),收尾離開。
 *   3. 界內:讀目標 tile 的 BLK index:
 *        si = layout_base[0xb26] + (ty*width + tx)*2
 *        tile_word = seg_cty:si(u16);tile_index = 低 byte。
 *   4. 查屬性表:attr = tile_attr_tab[tile_index*2]  (DS:0x308e);
 *        存 [0x2577] = attr。
 *   5. 依屬性決定行為:
 *        attr_hi & 0x20(ah&0x20)→ 特殊「可走且觸發」:套用位移後跳 0x2ec5
 *          (走過去 + 呼叫互動 0x...19a;失敗回退)。
 *        attr & 0x0001 → **阻擋**(牆):bp=0、設碰撞音 [0x2875]=1、
 *          lcall 1053:240、[0x4f1f]=0xff(取消移動),return。
 *        attr & 0xe000 → 特殊地形(事件):or [0x4f46],1(觸發旗標);
 *          再依 ah&0x1f 細分(門/階梯/觸發點)。
 *        否則(可走):記 [0x2576]=tile_index、依 ah&0xc0 比較 [0x2579]
 *          (角落屬性變化 → 設 warp 旗標 [0x257b]=1);
 *          最後**套用位移**:[0x4f33]+=[0x2572];[0x4f35]+=[0x2574](真正移動)。
 *
 * → 解開 BLKATT/blkbm 屬性語意:每 tile u16 屬性,bit0=阻擋牆、bits13-15
 *   (0xe000)=事件地形、bit5(0x20)=可走且觸發、bits6-7(0xc0)=角落/出口
 *   類別。(docs/04 BLKATT「每 tile 2 byte mask」由此對齊到具體位元。)
 */
void player_move_collide(void)
{
    i16 tx, ty;
    u16 si, tile_word;
    u8  tile_index, attr_hi;
    u16 attr;

    tx = g_pos_x + g_player_dx;     /* [0x4f33] + [0x2572] */
    ty = g_pos_y + g_player_dy;     /* [0x4f35] + [0x2574] */

    /* (2) 邊界:出界 → 過場到地表 */
    if (tx < 0 || tx >= g_map_w || ty < 0 || ty >= g_map_h) {
        g_scene_flag = SCENE_TO_OVERWORLD;   /* [0x4f2d]=3 */
        g_tile_cur = g_border_tile;          /* [0x2576]=[0xb2d] */
        goto apply_move;                     /* jmp 0x2e66(套用位移後返回)*/
    }

    /* (3) 讀目標 tile */
    si = g_layout_base + (u16)(ty * g_map_w + tx) * 2;   /* [0xb26] + idx*2 */
    tile_word = *(u16 far *)MK_FP(seg_cty, si);
    tile_index = (u8)tile_word;          /* 低 byte = BLK index */

    /* (4) 查屬性 */
    attr = tile_attr_tab[tile_index];    /* [bx*2 + 0x308e] */
    g_tile_attr = attr;                  /* [0x2577] */
    attr_hi = (u8)(attr >> 8);           /* ah */

    /* (5) 行為分支 */
    if (attr_hi & 0x20) {
        /* 可走且觸發:先走過去,再呼叫互動;互動結果決定是否回退 */
        i16 ck;
        g_pos_x += g_player_dx;          /* [0x4f33]+=[0x2572] */
        g_pos_y += g_player_dy;          /* [0x4f35]+=[0x2574] */
        ck = tile_interact((u8)tile_word);  /* call 0x...19a(AL 帶 tile)*/
        if (ck == 0) {
            /* 互動失敗 → 回退 */
            g_pos_x -= g_player_dx;
            g_pos_y -= g_player_dy;
            g_collide_sfx = 1;           /* [0x2875]=1 */
            file_op_dx(0x240);           /* lcall 1053:240 撞牆音 */
        }
        g_move_dir = DIR_NONE;           /* [0x4f1f]=0xff(本幀已處理)*/
        return;
    }

    if (attr & 0x0001) {
        /* 阻擋(牆):不移動 */
        g_collide_sfx = 1;               /* [0x2875]=1 */
        file_op_dx(0x240);               /* lcall 1053:240 */
        g_move_dir = DIR_NONE;           /* [0x4f1f]=0xff */
        return;
    }

    if (attr & 0xe000) {
        /* 特殊地形/事件 */
        g_status_evt |= 1;               /* or [0x4f46],1 */
        if ((attr_hi & 0x1f) && !(attr & 0xc8)) {
            /* 子事件:記事件碼,進一步分類(門/階梯…)*/
            g_status_evt |= 0x800;       /* or [0x4f46],0x800 */
            g_event_code = attr_hi & 0x1f;  /* [0x258c] */
        }
        /* fallthrough 至可走分支(特殊地形多半也可走)*/
    }

    /* 可走(或事件後可走):記錄並比較角落屬性 */
    g_tile_cur = tile_index;             /* [0x2576] */
    {
        u8 corner = attr_hi & 0xc0;
        if (corner != g_corner_attr) {   /* [0x2579] 改變 → 出口/warp */
            g_corner_new = corner;       /* [0x257a] */
            g_warp_pending = 1;          /* [0x257b]=1 */
        }
    }

apply_move:
    /* 真正套用位移(移動成立)*/
    g_pos_x += g_player_dx;              /* [0x4f33]+=[0x2572] */
    g_pos_y += g_player_dy;              /* [0x4f35]+=[0x2574] */
}

/* ---- player_move_collide 用到的 forward ---- */
extern i16  tile_interact(u8 tile);   /* sub 0x...19a(file ~0x4ed4)tile 互動,回 0=失敗 */
extern u8   g_border_tile;            /* DS:0x0b2d  邊界填充 tile */
extern u8   g_collide_sfx;            /* DS:0x2875 */
extern u8   g_event_code;             /* DS:0x258c */
extern u8   g_corner_new;             /* DS:0x257a */

/* ================================================================= */
/* scene_event_poll —— sub_2f01(file 0x4271)
 *
 * [0x4f46]&1 為真時呼叫:清掉 bit0,呼叫事件常式 0x488f(對話/腳本/戰鬥入口)。
 * 簡短常式:純旗標檢查 + 派發。
 */
void scene_event_poll(void)
{
    if (g_status_evt & 1) {              /* test al,1 of [0x4f46] */
        g_status_evt &= 0xfffe;         /* 清 bit0 */
        scene_event_run();              /* call 0x488f */
    }
}

extern void scene_event_run(void);     /* sub 0x488f  事件/腳本派發 */
