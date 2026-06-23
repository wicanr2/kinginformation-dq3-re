/* test_npc.c — NPC 隨機走動步進器 + EXE RNG 驗證(docs/35 §九)。
 * 步進器只讀 dq3_scene 結構欄位,不呼叫 scene 函式 → 測試自建 scene、免連 SDL。 */
#include "dq3_npc.h"
#include "dq3_scene.h"   /* 完整 dq3_scene 定義(建 scene 物件用)*/
#include "dq3_rng.h"
#include <stdio.h>
#include <string.h>

static int g_fail = 0;
#define CHECK(c,msg) do{ if(c)printf("  [PASS] %s\n",msg); else {printf("  [FAIL] %s\n",msg); g_fail++;} }while(0)

/* 9×9 全可走場景(tile index 0 = 地板 attr 0;index 1 = 牆 attr bit0)。 */
static uint8_t  idx[81];
static uint8_t  himap[81];
static uint16_t attr2[2] = { 0x0000, 0x0001 };  /* 0=地板 1=牆 */

static void build(dq3_scene *s)
{
    memset(s, 0, sizeof *s);
    s->map_w = s->map_h = 9;
    s->index_map = idx; s->hi_map = himap;
    s->attr = attr2; s->attr_count = 2;
    memset(idx, 0, sizeof idx);
    memset(himap, 0, sizeof himap);
    s->px = 0; s->py = 0;            /* 玩家放角落,遠離測試 NPC */
}

int main(void)
{
    dq3_scene s;
    dq3_rng rng;

    printf("== EXE RNG(0xfa39:+0x9018 / rol3 / mod)==\n");
    dq3_rng_seed(&rng, 0);
    CHECK(dq3_rng_next(&rng, 100) == 64, "seed0 → next(100)=64(手算 0x80c4%100)");
    CHECK(dq3_rng_next(&rng, 0) == 0, "n=0 → 0");
    {
        int i, inrange = 1; dq3_rng_seed(&rng, 12345);
        for (i = 0; i < 1000; i++) { int v = dq3_rng_next(&rng, 4); if (v < 0 || v >= 4) inrange = 0; }
        CHECK(inrange, "next(4) 永遠 ∈ [0,4)");
    }

    printf("== 靜止 / 凍結 NPC 不動 ==\n");
    {
        dq3_npc npc; int i; memset(&npc, 0, sizeof npc);
        build(&s); npc.x = 4; npc.y = 4; npc.ctrl = 0x00;   /* bit2 清 = 靜止 */
        dq3_rng_seed(&rng, 7);
        for (i = 0; i < 500; i++) dq3_npc_step(&s, &npc, 0, 1, &rng);
        CHECK(npc.x == 4 && npc.y == 4, "靜止 NPC(bit2=0)500 步不動");

        build(&s); npc.x = 4; npc.y = 4; npc.ctrl = DQ3_NPC_MOVE | DQ3_NPC_FROZEN;  /* 移動但凍結 */
        dq3_rng_seed(&rng, 7);
        for (i = 0; i < 500; i++) dq3_npc_step(&s, &npc, 0, 1, &rng);
        CHECK(npc.x == 4 && npc.y == 4, "凍結 NPC(bit7)500 步不動");
    }

    printf("== 移動 NPC:軌跡合法(界內/相鄰/不踏玩家/可走)==\n");
    {
        dq3_npc npc; int i, moved = 0, ok = 1; int px, py;
        build(&s); npc.x = 4; npc.y = 4; npc.ctrl = DQ3_NPC_MOVE;  /* bit2 set */
        dq3_npc_stamp(&s, &npc, 0);
        dq3_rng_seed(&rng, 999);
        px = npc.x; py = npc.y;
        for (i = 0; i < 4000; i++) {
            int r = dq3_npc_step(&s, &npc, 0, 1, &rng);
            if (r) {
                int man = (npc.x>px?npc.x-px:px-npc.x) + (npc.y>py?npc.y-py:py-npc.y);
                if (man != 1) ok = 0;                                 /* 必相鄰一步 */
                if (npc.x>=s.map_w||npc.y>=s.map_h) ok = 0;           /* 界內 */
                if (npc.x==s.px && npc.y==s.py) ok = 0;               /* 不踏玩家 */
                moved++; px = npc.x; py = npc.y;
            }
        }
        printf("  4000 步中移動 %d 次,末位置(%d,%d)\n", moved, npc.x, npc.y);
        CHECK(moved > 0, "移動 NPC 會走動");
        CHECK(ok, "每次移動都合法(相鄰/界內/非玩家格)");
    }

    printf("== 全牆包圍 → 不動 ==\n");
    {
        dq3_npc npc; int i;
        build(&s); npc.x = 4; npc.y = 4; npc.ctrl = DQ3_NPC_MOVE;
        idx[3*9+4]=idx[5*9+4]=idx[4*9+3]=idx[4*9+5]=1;  /* 上下左右設牆 */
        dq3_npc_stamp(&s, &npc, 0);
        dq3_rng_seed(&rng, 4242);
        for (i = 0; i < 2000; i++) dq3_npc_step(&s, &npc, 0, 1, &rng);
        CHECK(npc.x == 4 && npc.y == 4, "四周牆包圍 → NPC 不移動");
    }

    printf("== 兩 NPC 不互疊 ==\n");
    {
        dq3_npc npc[2]; int i, overlap = 0;
        build(&s);
        npc[0].x=2; npc[0].y=4; npc[0].ctrl=DQ3_NPC_MOVE; memset(&npc[0].b2,0,5);
        npc[1].x=6; npc[1].y=4; npc[1].ctrl=DQ3_NPC_MOVE; memset(&npc[1].b2,0,5);
        dq3_npc_stamp(&s, npc, 0); dq3_npc_stamp(&s, npc, 1);
        dq3_rng_seed(&rng, 55);
        for (i = 0; i < 4000; i++) {
            dq3_npc_step(&s, npc, 0, 2, &rng);
            dq3_npc_step(&s, npc, 1, 2, &rng);
            if (npc[0].x==npc[1].x && npc[0].y==npc[1].y) overlap = 1;
        }
        CHECK(!overlap, "兩 NPC 不會走到同一格(0x20 佔位閘)");
    }

    printf("\n%s (%d failures)\n", g_fail?"== 有測試未通過 ==":"== 全部通過 ==", g_fail);
    return g_fail ? 1 : 0;
}
