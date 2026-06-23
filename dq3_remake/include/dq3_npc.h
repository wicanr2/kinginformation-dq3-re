/* dq3_npc.h — 城鎮 NPC 槽 + per-frame 隨機走動步進器(docs/35 §九)。
 *
 * 鏡射 DQ3.EXE mover L02025(file 0x3395):逐 NPC、每幀節流式隨機走動。
 * 槽 = EXE [0xb66] 8-byte 槽;碰撞/玩家位置/地圖由 dq3_scene 提供。
 *
 * 移動控制在 ctrl(=EXE slot[3]):
 *   bits0-1 = 朝向(0下 1左 2上 3右,同 dq3_scene facing)
 *   bit2 (0x04) = 移動開關(set=隨機走動 / clear=靜止 NPC)  ← 「會動 vs 不會動」
 *   bit7 (0x80) = 暫時凍結(對話中等)
 */
#ifndef DQ3_NPC_H
#define DQ3_NPC_H

#include <stdint.h>
#include "dq3_rng.h"

typedef struct dq3_scene dq3_scene;   /* 前向宣告(避免與 dq3_scene.h 循環 include)*/

#define DQ3_NPC_MOVE   0x04
#define DQ3_NPC_FROZEN 0x80
#define DQ3_NPC_TILE_OCCUPIED 0x20   /* tile 高 byte:此格有 NPC(EXE 0x20|slot)*/

typedef struct {
    uint8_t x, y;        /* slot[0],[1] 地圖格座標 */
    uint8_t b2;          /* slot[2] sprite 群 id */
    uint8_t ctrl;        /* slot[3] 移動控制(見上)*/
    uint8_t b4;          /* slot[4] */
    uint8_t flags;       /* slot[5] 可見性旗標 id(低 3 bit)*/
    uint8_t base_hi;     /* slot[6] 被蓋住的底 tile 高 byte(供還原)*/
    uint8_t b7;          /* slot[7] */
} dq3_npc;

/* 把 NPC 蓋上場景 hi_map(存底 tile 到 base_hi;tile 高 byte = 底&0xc0 | 0x20|slot)。 */
void dq3_npc_stamp(dq3_scene *s, dq3_npc *npcs, int idx);

/* 走一步(鏡射 mover L02025):每幀 1/4 機率評估、靜止/凍結直接 return;
 * 多沿當前朝向走、1/20 轉向;落步閘:界內 / 不踏玩家格(否則反向)/ 不疊 NPC(0x20)/ 不撞牆(attr bit0)。
 * 成功移動回 1(並更新 hi_map stamp);否則 0。idx/n_npcs 供轉向 ±1 與佔位判定。 */
int  dq3_npc_step(dq3_scene *s, dq3_npc *npcs, int idx, int n_npcs, dq3_rng *rng);

#endif /* DQ3_NPC_H */
