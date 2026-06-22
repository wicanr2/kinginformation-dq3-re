/* mainloop.c — 精訊 DQ3.EXE 主迴圈骨架(real-mode large-model C)
 *
 * 反編譯來源:tools/re_recdis.py 遞迴反組譯 sub_93e3(seg0:93e3,file 0xa753);
 * 對照 docs/08-exe-functions.md「主迴圈」一節。
 *
 * 結構:每幀 讀鍵盤 → 方向鍵改玩家座標/朝向 → 其他鍵走狀態機跳表 →
 *        繪場景 → 更新玩家(移動+碰撞)→ 更新 HUD/子系統 → 節拍等待 → 回圈。
 * sub_93e3 是 game_body 的最後一個 call,return 後整個遊戲結束。
 *
 * 求結構正確、可審閱;非位元精準。far runtime 以 exe.h wrapper 表達。
 */
#include "exe.h"

/* DOS / BIOS 方向鍵 scancode(AH 來自 kbd_read_scancode) */
#define SC_DOWN  0x50
#define SC_LEFT  0x4b
#define SC_UP    0x48
#define SC_RIGHT 0x4d
#define SC_MENU  0x3e   /* 0x3e 鍵 → 選單/特殊(lcall 11dd:0c) */

#define DIR_DOWN  0
#define DIR_LEFT  1
#define DIR_UP    2
#define DIR_RIGHT 3
#define DIR_NONE  0xff

/* 朝向(8 朝向系統,精靈方向用):下/左/上/右 = 0/2/4/6 */
#define FACE_DOWN  0
#define FACE_LEFT  2
#define FACE_UP    4
#define FACE_RIGHT 6

/* sub_93e3:主迴圈 */
void main_loop(void)
{
    int i;

    /* 進入前重置玩家狀態並繪首幀 */
    g_player_dx = 0;               /* [0x2572] */
    g_player_dy = 0;               /* [0x2574] */
    g_player_face = FACE_DOWN;     /* [0x26ad] */

    mouse_op(0x166);               /* lcall 10b6:166 */
    draw_scene();                  /* sub_255b */
    update_player();               /* sub_19b8 */
    mouse_op(0x19c);               /* lcall 10b6:19c */

    for (;;) {                     /* loop @ 0x9405 */
        u8 sc;

        /* 每幀先清本幀位移與方向 */
        g_player_dx = 0;
        g_player_dy = 0;
        g_move_dir  = DIR_NONE;    /* [0x4f1f]=0xff */

        sc = kbd_read_scancode();  /* lcall 1104:8a → AH = scancode(0=無) */

        if (sc == 0) {
            /* 無輸入 → 直接繪製 */
        }
        else if (sc == SC_DOWN) {
            g_player_dy++;                 /* inc [0x2574] */
            g_player_face = FACE_DOWN;     /* [0x26ad]=0 */
            g_move_dir    = DIR_DOWN;      /* [0x4f1f]=0 */
        }
        else if (sc == SC_LEFT) {
            g_player_dx--;                 /* dec [0x2572] */
            g_player_face = FACE_LEFT;     /* [0x26ad]=2 */
            g_move_dir    = DIR_LEFT;      /* [0x4f1f]=1 */
        }
        else if (sc == SC_UP) {
            g_player_dy--;                 /* dec [0x2574] */
            g_player_face = FACE_UP;       /* [0x26ad]=4 */
            g_move_dir    = DIR_UP;        /* [0x4f1f]=2 */
        }
        else if (sc == SC_RIGHT) {
            g_player_dx++;                 /* inc [0x2572] */
            g_player_face = FACE_RIGHT;    /* [0x26ad]=6 */
            g_move_dir    = DIR_RIGHT;     /* [0x4f1f]=3 */
        }
        else if (sc == SC_MENU) {
            menu_special();                /* lcall 11dd:0c —— 選單/指令入口 */
        }
        else {
            /* 其他鍵:14-entry 狀態機跳表。
             * scancode 表在 DS:0x19,far 函式指標表在 DS:0x25(2 bytes/entry)。 */
            for (i = 0; i < DQ3_CMD_TABLE_LEN; i++) {     /* cx=0xe; bx=0..; cmp [bx+0x19],ah */
                if (g_cmd_scancode[i] == sc) {
                    if (g_cmd_handler[i] != 0)             /* cmp [bx*2+0x25],0; je skip */
                        g_cmd_handler[i]();                /* call word ptr [bx+0x25] */
                    break;
                }
            }
        }

        /* draw @ 0x94b0 —— 每幀繪製/更新管線 */
        draw_scene();              /* sub_255b:地圖捲動繪製 */
        update_player();           /* sub_19b8:玩家精靈+移動(jump table [bx+0x255a])+碰撞 */
        update_hud();              /* sub_9530:狀態/HUD(test [0x4f46]&4) */
        update_sub();              /* sub_991d:子系統更新 */
        frame_wait();              /* sub_ee23:節拍/計時等待 */
    }                              /* jmp 0x9405 */
}

/* sub_94c3:自動移動一步。
 * 不讀鍵盤,改用上次方向 g_move_dir([0x4f1f])重播一格位移(碰撞解算後續步 /
 * 腳本移動 / 過場用)。繪製管線較主迴圈精簡。 */
void auto_step(void)
{
    g_player_dx = 0;
    g_player_dy = 0;
    g_player_face = FACE_DOWN;

    switch (g_move_dir) {          /* ax = [0x4f1f] */
    case DIR_DOWN:                 /* 0 */
        g_player_dy++;  g_player_face = FACE_DOWN;  break;
    case DIR_LEFT:                 /* 1 */
        g_player_dx--;  g_player_face = FACE_LEFT;  break;
    case DIR_UP:                   /* 2 */
        g_player_dy--;  g_player_face = FACE_UP;    break;
    case DIR_RIGHT:                /* 3 */
        g_player_dx++;  g_player_face = FACE_RIGHT; break;
    default:
        break;
    }

    draw_scene();                  /* sub_255b */
    update_player();               /* sub_19b8 */
    update_sub();                  /* sub_991d */
    g_move_dir = DIR_NONE;         /* [0x4f1f]=0xff */
}
