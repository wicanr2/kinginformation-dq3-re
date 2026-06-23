/* dq3_npc.c — NPC 隨機走動步進器(鏡射 DQ3.EXE mover L02025 / file 0x3395)。 */
#include "dq3_npc.h"
#include "dq3_scene.h"   /* 完整 dq3_scene 定義(npc.h 只前向宣告)*/

/* 方向位移(EXE [0xb35]/[0xb37] 表,stride-4 {dx,dy};0下 1左 2上 3右)。 */
static const int DDX[4] = { 0, -1, 0, 1 };
static const int DDY[4] = { 1,  0, -1, 0 };

void dq3_npc_stamp(dq3_scene *s, dq3_npc *npcs, int idx)
{
    dq3_npc *npc = &npcs[idx];
    int cell;
    if (!s->hi_map) return;
    cell = npc->y * s->map_w + npc->x;
    npc->base_hi = s->hi_map[cell];                       /* 存底 tile 高 byte */
    s->hi_map[cell] = (uint8_t)((npc->base_hi & 0xc0)     /* 保留頂 2 bit(轉場類)*/
                       | DQ3_NPC_TILE_OCCUPIED | (idx & 0x1f));  /* 0x20 | slot */
}

static void unstamp(dq3_scene *s, dq3_npc *npc)
{
    if (s->hi_map) s->hi_map[npc->y * s->map_w + npc->x] = npc->base_hi;  /* 還原底 tile */
}

/* 嘗試沿當前朝向走一步;成功回 1。 */
static int try_step(dq3_scene *s, dq3_npc *npcs, int idx)
{
    dq3_npc *npc = &npcs[idx];
    int dir = npc->ctrl & 3;
    int tx = npc->x + DDX[dir], ty = npc->y + DDY[dir];
    int cell, ni;

    if (tx < 0 || ty < 0 || tx >= s->map_w || ty >= s->map_h) return 0;   /* 界外 */
    if (tx == s->px && ty == s->py) {                     /* 踏到玩家格 → 反向、不走(L02146)*/
        npc->ctrl = (uint8_t)((npc->ctrl & 0xfc) | ((dir + 2) & 3));
        return 0;
    }
    cell = ty * s->map_w + tx;
    if (s->hi_map && (s->hi_map[cell] & DQ3_NPC_TILE_OCCUPIED)) return 0;  /* 已有 NPC(L0217e)*/
    if (s->attr) {                                        /* 牆(attr bit0,L0218c)*/
        ni = s->index_map[cell];
        if (ni < s->attr_count && (s->attr[ni] & 0x0001)) return 0;
    }
    /* 成立:移動 + 重蓋 */
    unstamp(s, npc);
    npc->x = (uint8_t)tx; npc->y = (uint8_t)ty;
    dq3_npc_stamp(s, npcs, idx);
    return 1;
    /* 註:EXE 另有「離玩家 ≥3 格」近距閘(L020e1/L02111),語意/軸別待再追,尚未鏡射。 */
}

int dq3_npc_step(dq3_scene *s, dq3_npc *npcs, int idx, int n_npcs, dq3_rng *rng)
{
    dq3_npc *npc = &npcs[idx];
    int cur, rdir;

    if (dq3_rng_next(rng, 4) != 1) return 0;       /* ★ 節流:每幀 1/4 才評估(L02040)*/
    if (npc->ctrl & DQ3_NPC_FROZEN) return 0;      /* 凍結(L02048)*/
    if (!(npc->ctrl & DQ3_NPC_MOVE)) return 0;     /* ★ bit2 清 = 靜止 NPC(L0204e)*/

    cur  = npc->ctrl & 3;
    rdir = dq3_rng_next(rng, 4);                   /* 隨機方向(L02062)*/
    if (rdir == cur) {
        return try_step(s, npcs, idx);             /* 同朝向 → 直接走(L02057)*/
    }
    if (dq3_rng_next(rng, 20) == 1) {              /* 1/20 才轉向(L02071)*/
        int nd = (idx < n_npcs / 2) ? ((cur + 1) & 3) : ((cur + 3) & 3);  /* ±1 by index(L0207f)*/
        npc->ctrl = (uint8_t)((npc->ctrl & 0xfc) | nd);
        return try_step(s, npcs, idx);
    }
    return 0;                                       /* 本幀不動 */
}
