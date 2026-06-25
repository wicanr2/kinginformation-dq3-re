/* dq3_ship.c — 船系統實作(#2)。海 tile attr 指紋見 docs/48。 */
#include "dq3_ship.h"

void dq3_ship_init(dq3_ship *sh)
{
    sh->owned = 0; sh->aboard = 0; sh->px = -1; sh->py = -1; sh->layer = 0;
}

/* tile 的 overworld attr(OOB→0xffff 表邊界)。 */
static unsigned tile_attr(const dq3_scene *s, int tx, int ty)
{
    int idx;
    if (tx < 0 || ty < 0 || tx >= s->map_w || ty >= s->map_h) return 0xffff;
    idx = s->index_map[ty * s->map_w + tx];
    if (!s->attr || idx >= s->attr_count) return 0;
    return s->attr[idx];
}

int dq3_ship_navigable(const dq3_scene *s, int tx, int ty)
{
    unsigned a = tile_attr(s, tx, ty);
    if (a == 0xffff) return 0;                  /* 邊界:不可航行 */
    return (a & 0x0001) && (a & 0x0020);        /* 海 0x25:BLK + bit5;山 0x07 無 bit5 */
}

/* 解方向掃描碼 → 位移 + facing(對齊 dq3_scene_input)。回 1 = 有效方向鍵。 */
static int decode_dir(uint8_t sc, int *dx, int *dy, int *facing)
{
    *dx = *dy = 0;
    switch (sc) {
        case 0x48: *dy = -1; *facing = 2; return 1;
        case 0x50: *dy =  1; *facing = 0; return 1;
        case 0x4b: *dx = -1; *facing = 1; return 1;
        case 0x4d: *dx =  1; *facing = 3; return 1;
        default: return 0;
    }
}

int dq3_ship_input(dq3_scene *s, dq3_ship *sh, uint8_t sc, int layer)
{
    int dx, dy, tx, ty;
    if (!decode_dir(sc, &dx, &dy, &s->facing)) return DQ3_SHIP_BLOCKED;
    tx = s->px + dx; ty = s->py + dy;

    if (sh->aboard) {
        if (dq3_ship_navigable(s, tx, ty)) {            /* 航行至水格 */
            s->px = tx; s->py = ty;
            return DQ3_SHIP_SAILED;
        }
        if (dq3_scene_walkable(s, tx, ty) && dq3_scene_npc_at(s, tx, ty) < 0) {
            sh->px = s->px; sh->py = s->py; sh->layer = layer;   /* 船停在原水格 */
            sh->aboard = 0;
            s->px = tx; s->py = ty;
            return DQ3_SHIP_DISEMBARK;
        }
        return DQ3_SHIP_BLOCKED;
    }

    /* 步行:踏上停泊船格 → 登船 */
    if (sh->owned && sh->layer == layer && tx == sh->px && ty == sh->py) {
        sh->aboard = 1;
        s->px = tx; s->py = ty;
        return DQ3_SHIP_BOARD;
    }
    if (dq3_scene_walkable(s, tx, ty) && dq3_scene_npc_at(s, tx, ty) < 0) {
        s->px = tx; s->py = ty;
        return DQ3_SHIP_WALKED;
    }
    return DQ3_SHIP_BLOCKED;
}
