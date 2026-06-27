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

/* 把 NPC idx 移到 (nx,ny):unstamp 舊格 → 移動 → stamp 新格(供倉庫番推石 B-7)。 */
void dq3_npc_move(dq3_scene *s, dq3_npc *npcs, int idx, int nx, int ny)
{
    unstamp(s, &npcs[idx]);
    npcs[idx].x = (uint8_t)nx; npcs[idx].y = (uint8_t)ny;
    dq3_npc_stamp(s, npcs, idx);
}

/* B-6 新城鎮視覺:移除可見性 flag == visflag 的 NPC(原版「建城後才顯示」的商店/居民),
 * 但保留 (keep_x,keep_y) 位的(老人)。unstamp + 陣列壓縮;保留者位置不動,hi_map OCCUPIED 仍正確。 */
void dq3_scene_hide_unbuilt(dq3_scene *s, int visflag, int keep_x, int keep_y)
{
    int i, j = 0;
    for (i = 0; i < s->n_npcs; i++) {
        if ((s->npcs[i].flags >> 3) == visflag &&
            !(s->npcs[i].x == keep_x && s->npcs[i].y == keep_y)) {
            unstamp(s, &s->npcs[i]);          /* 還原 hi_map(消失)*/
            continue;
        }
        if (j != i) s->npcs[j] = s->npcs[i];  /* 壓縮(位置不變,cell stamp 仍有效)*/
        j++;
    }
    s->n_npcs = j;
}

/* 嘗試沿當前朝向走一步;成功回 1。 */
static int try_step(dq3_scene *s, dq3_npc *npcs, int idx)
{
    dq3_npc *npc = &npcs[idx];
    int dir = npc->ctrl & 3;
    int tx = npc->x + DDX[dir], ty = npc->y + DDY[dir];
    int cell, ni;

    if (tx < 0 || ty < 0 || tx >= s->map_w || ty >= s->map_h) return 0;   /* 界外 */
    /* 近玩家閘(L020e1/L02111,test bl,1 分軸):沿移動軸與玩家距離 <3 → 不走
     * (左右看 X、上下看 Y;NPC 不擠進玩家 2 格內,留空間讓玩家靠近對話)。 */
    if (dir & 1) { int d = s->px - tx; if (d < 3 && d > -3) return 0; }   /* 左/右 → X */
    else         { int d = s->py - ty; if (d < 3 && d > -3) return 0; }   /* 上/下 → Y */
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
