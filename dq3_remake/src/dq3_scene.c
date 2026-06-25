/* dq3_scene.c — 場景核心實作(攝影機/貼圖/碰撞/走動)。 */
#include "dq3_scene.h"
#include "dq3_runtime.h"
#include <stdlib.h>
#include <string.h>

#define VIEW_COLS (DQ3_SCREEN_W / DQ3_TILE_W)                 /* 20 */
#define VIEW_ROWS ((DQ3_SCREEN_H + DQ3_TILE_H - 1) / DQ3_TILE_H) /* 15 */

int dq3_scene_walkable(const dq3_scene *s, int tx, int ty)
{
    int idx; uint16_t a;
    if (tx < 0 || ty < 0 || tx >= s->map_w || ty >= s->map_h) return 0;
    idx = s->index_map[ty * s->map_w + tx];
    if (!s->attr || idx >= s->attr_count) return 1;   /* 無屬性 → 可走 */
    a = s->attr[idx];
    return (a & 0x0001) ? 0 : 1;
}

int dq3_scene_tile_event(const dq3_scene *s, int tx, int ty, int *type, int *param)
{
    return dq3_scene_tile_event_p2(s, tx, ty, type, param, 0);
}

int dq3_scene_tile_event_p2(const dq3_scene *s, int tx, int ty, int *type, int *param, int *p2)
{
    int i, idx, subid;
    if (!s->hi_map || s->n_events <= 0) return 0;
    if (tx < 0 || ty < 0 || tx >= s->map_w || ty >= s->map_h) return 0;
    i = ty * s->map_w + tx;
    idx = s->index_map[i];
    /* 屬性 attr&8 = 事件格(docs/31) */
    if (!s->attr || idx >= s->attr_count || !(s->attr[idx] & 0x0008)) return 0;
    subid = s->hi_map[i] & 0x1f;          /* 高 byte 低 5 bit = 事件 subid */
    if (subid >= s->n_events) return 0;
    if (type)  *type  = s->events[subid].type;
    if (param) *param = s->events[subid].param;
    if (p2)    *p2    = s->events[subid].p2;
    return 1;
}

int dq3_scene_tile_transition(const dq3_scene *s, int tx, int ty,
                              int *dest_cty, int *dest_sec, int *dx, int *dy)
{
    int i, idx, subid;
    if (!s->hi_map || s->n_transitions <= 0) return 0;
    if (tx < 0 || ty < 0 || tx >= s->map_w || ty >= s->map_h) return 0;
    i = ty * s->map_w + tx;
    idx = s->index_map[i];
    /* 屬性 attr 高 3 bit(0xe000)= 轉場 tile(docs/35;move handler 0x2e8c)。 */
    if (!s->attr || idx >= s->attr_count || !(s->attr[idx] & 0xe000)) return 0;
    subid = s->hi_map[i] & 0x1f;          /* 高 byte 低 5 bit = 轉場 subid */
    if (subid >= s->n_transitions) return 0;
    if (dest_cty) *dest_cty = s->transitions[subid].dest_cty;
    if (dest_sec) *dest_sec = s->transitions[subid].dest_sec;
    if (dx)       *dx       = s->transitions[subid].x;
    if (dy)       *dy       = s->transitions[subid].y;
    return 1;
}

int dq3_scene_door_tier(const dq3_scene *s, int tx, int ty)
{
    int idx; uint16_t a;
    if (tx < 0 || ty < 0 || tx >= s->map_w || ty >= s->map_h) return 0;
    idx = s->index_map[ty * s->map_w + tx];
    if (!s->attr || idx >= s->attr_count) return 0;
    a = s->attr[idx];
    if (!(a & 0x00c0)) return 0;          /* 0x4906: test attr,0xc0 → bits6-7=0 非鎖門 */
    return (a >> 6) & 3;                   /* rol bl,1 ×2; and 3 → 所需鑰匙等級 1/2/3 */
}

int dq3_scene_open_door(dq3_scene *s, int tx, int ty)
{
    int i; uint8_t hi;
    if (!dq3_scene_door_tier(s, tx, ty)) return 0;
    i = ty * s->map_w + tx;
    hi = s->hi_map ? s->hi_map[i] : 0;
    s->index_map[i] = (uint8_t)(hi & 0x1f);   /* 0x4981: [di] = high & 0x1f(門→通行 tile)*/
    if (s->hi_map) s->hi_map[i] = (uint8_t)(hi & 0xe0); /* 0x4985: [di+1] &= 0xe0(清事件位、留轉場位)*/
    return 1;
}

/* 面向位移(同 dq3_scene_input:0下 1左 2上 3右)。 */
static void facing_delta(int facing, int *dx, int *dy)
{
    *dx = 0; *dy = 0;
    switch (facing & 3) {
        case 0: *dy =  1; break;   /* 下 */
        case 1: *dx = -1; break;   /* 左 */
        case 2: *dy = -1; break;   /* 上 */
        case 3: *dx =  1; break;   /* 右 */
    }
}

int dq3_scene_try_open_facing_door(dq3_scene *s, int key_tier)
{
    int dx, dy, tx, ty, need;
    facing_delta(s->facing, &dx, &dy);
    tx = s->px + dx; ty = s->py + dy;
    need = dq3_scene_door_tier(s, tx, ty);
    if (need == 0 || key_tier < need) return 0;   /* 非門 / 鑰匙不夠 */
    return dq3_scene_open_door(s, tx, ty);
}

int dq3_scene_load_npcs(dq3_scene *s, const uint8_t *cty, size_t cty_len, size_t so)
{
    size_t base; int cnt, i;
    uint16_t np;
    s->n_npcs = 0;
    dq3_rng_seed(&s->npc_rng, (uint16_t)(so * 2654 + 1));   /* 確定性種子(依 section)*/
    if (so + 4 > cty_len) return 0;
    np = (uint16_t)(cty[so] | (cty[so + 1] << 8));          /* +0 NPC 清單 */
    if (np == 0xffff) np = (uint16_t)(cty[so + 2] | (cty[so + 3] << 8));  /* 退 +2 */
    if (np == 0xffff || so + np >= cty_len) return 0;
    base = so + np;
    cnt = cty[base];                                        /* {count, records*7} */
    for (i = 0; i < cnt && s->n_npcs < DQ3_SCENE_MAX_NPC; i++) {
        const uint8_t *r = cty + base + 1 + (size_t)i * 7;
        dq3_npc *n;
        if (base + 1 + (size_t)i * 7 + 7 > cty_len) break;
        if (r[0] >= s->map_w || r[1] >= s->map_h) continue; /* 座標越界跳過 */
        n = &s->npcs[s->n_npcs];
        n->x = r[0]; n->y = r[1]; n->b2 = r[2]; n->ctrl = r[3];
        n->b4 = r[4]; n->flags = r[5]; n->b7 = r[6]; n->base_hi = 0;
        dq3_npc_stamp(s, s->npcs, s->n_npcs);               /* 蓋上 hi_map(0x20|slot)*/
        s->n_npcs++;
    }
    return s->n_npcs;
}

int dq3_scene_npc_tick(dq3_scene *s)
{
    int i, moved = 0;
    for (i = 0; i < s->n_npcs; i++)
        moved += dq3_npc_step(s, s->npcs, i, s->n_npcs, &s->npc_rng) ? 1 : 0;
    return moved;
}

/* 找 b2 在快取的索引;無則回 -1。 */
static int npc_spr_find(const dq3_scene *s, int b2)
{
    int i;
    for (i = 0; i < s->n_npc_spr; i++) if (s->npc_spr_b2[i] == b2) return i;
    return -1;
}

int dq3_scene_load_npc_sprites(dq3_scene *s, const char *assets_dir)
{
    int i; char err[128];
    s->n_npc_spr = 0;
    for (i = 0; i < s->n_npcs && s->n_npc_spr < 8; i++) {
        int b2 = s->npcs[i].b2, entry;
        if (npc_spr_find(s, b2) >= 0) continue;                 /* 已快取 */
        if (b2 < 4) continue;                                   /* key<4 保留(無對應 BLS 角色)*/
        entry = (b2 - 4) * 4;   /* BLS offset=(key-4)*0xf00+6 → entry_base=(b2-4)*4(RE file 0xff99/0xffc3)*/
        if (dq3_charsprite_load(&s->npc_spr[s->n_npc_spr], assets_dir, "DQ3MAN.BLS",
                                entry, err, sizeof err) == 0) {
            s->npc_spr_b2[s->n_npc_spr] = b2;
            s->n_npc_spr++;
        }
    }
    return s->n_npc_spr;
}

int dq3_scene_input(dq3_scene *s, uint8_t sc)
{
    int dx = 0, dy = 0, tx, ty;
    switch (sc) {
        case 0x48: dy = -1; s->facing = 2; break;
        case 0x50: dy =  1; s->facing = 0; break;
        case 0x4b: dx = -1; s->facing = 1; break;
        case 0x4d: dx =  1; s->facing = 3; break;
        default: return 0;
    }
    tx = s->px + dx; ty = s->py + dy;
    if (!dq3_scene_walkable(s, tx, ty)) return 0;
    s->px = tx; s->py = ty;
    return 1;
}

void dq3_scene_render(const dq3_scene *s, uint8_t *fb, int fb_w, int fb_h)
{
    int cam_x = s->px - VIEW_COLS / 2;
    int cam_y = s->py - VIEW_ROWS / 2;
    int row, col, r, c;

    if (cam_x > s->map_w - VIEW_COLS) cam_x = s->map_w - VIEW_COLS;
    if (cam_y > s->map_h - VIEW_ROWS) cam_y = s->map_h - VIEW_ROWS;
    if (cam_x < 0) cam_x = 0;
    if (cam_y < 0) cam_y = 0;

    memset(fb, 0, (size_t)fb_w * fb_h);

    for (row = 0; row < VIEW_ROWS; row++) {
        int my = cam_y + row, sy = row * DQ3_TILE_H;
        if (my < 0 || my >= s->map_h) continue;
        for (col = 0; col < VIEW_COLS; col++) {
            int mx = cam_x + col, sx = col * DQ3_TILE_W;
            int idx;
            const uint8_t (*t)[DQ3_TILE_W];
            if (mx < 0 || mx >= s->map_w) continue;
            idx = s->index_map[my * s->map_w + mx];
            if (idx < 0 || idx >= s->tile_count) continue;
            t = s->tiles[idx];
            for (r = 0; r < DQ3_TILE_H; r++) {
                int yy = sy + r;
                if (yy < 0 || yy >= fb_h) continue;
                for (c = 0; c < DQ3_TILE_W; c++) {
                    int xx = sx + c;
                    if (xx < 0 || xx >= fb_w) continue;
                    fb[yy * fb_w + xx] = t[r][c];
                }
            }
        }
    }

    /* 玩家:有真主角 sprite 則透明 blit,否則退回佔位方框。 */
    {
        int psx = (s->px - cam_x) * DQ3_TILE_W;
        int psy = (s->py - cam_y) * DQ3_TILE_H;
        int r2, c2;
        if (s->has_hero) {
            int fr = s->frame_for_facing[s->facing & 3];
            if (fr < 0 || fr >= DQ3_CHAR_FRAMES) fr = 0;
            for (r2 = 0; r2 < DQ3_CHAR_H; r2++)
                for (c2 = 0; c2 < DQ3_CHAR_W; c2++) {
                    int yy = psy + r2, xx = psx + c2;
                    if (!s->hero.opaque[fr][r2][c2]) continue;   /* 透明像素跳過 */
                    if (yy >= 0 && yy < fb_h && xx >= 0 && xx < fb_w)
                        fb[yy * fb_w + xx] = s->hero.px[fr][r2][c2];
                }
        } else {
            for (r2 = 4; r2 < DQ3_TILE_H - 4; r2++)
                for (c2 = 8; c2 < DQ3_TILE_W - 8; c2++) {
                    int yy = psy + r2, xx = psx + c2;
                    int edge  = (r2 == 4 || r2 == DQ3_TILE_H-5 || c2 == 8 || c2 == DQ3_TILE_W-9);
                    int cross = (c2 == DQ3_TILE_W/2 || r2 == DQ3_TILE_H/2);
                    if ((edge || cross) && yy >= 0 && yy < fb_h && xx >= 0 && xx < fb_w)
                        fb[yy * fb_w + xx] = 15;
                }
        }
    }

    /* NPC:有 sprite 快取則透明 blit(frame=朝向),否則小佔位點。 */
    {
        int n;
        for (n = 0; n < s->n_npcs; n++) {
            const dq3_npc *npc = &s->npcs[n];
            int nx = npc->x - cam_x, ny = npc->y - cam_y;
            int nsx, nsy, ci, r2, c2;
            if (nx < 0 || nx >= VIEW_COLS || ny < 0 || ny >= VIEW_ROWS) continue;  /* 視窗外 */
            nsx = nx * DQ3_TILE_W; nsy = ny * DQ3_TILE_H;
            ci = npc_spr_find(s, npc->b2);
            if (ci >= 0) {
                int fr = s->frame_for_facing[npc->ctrl & 3];
                const dq3_charsprite *cs = &s->npc_spr[ci];
                if (fr < 0 || fr >= DQ3_CHAR_FRAMES) fr = 0;
                for (r2 = 0; r2 < DQ3_CHAR_H; r2++)
                    for (c2 = 0; c2 < DQ3_CHAR_W; c2++) {
                        int yy = nsy + r2, xx = nsx + c2;
                        if (!cs->opaque[fr][r2][c2]) continue;
                        if (yy >= 0 && yy < fb_h && xx >= 0 && xx < fb_w)
                            fb[yy * fb_w + xx] = cs->px[fr][r2][c2];
                    }
            } else {
                for (r2 = 8; r2 < DQ3_TILE_H - 8; r2++)
                    for (c2 = 12; c2 < DQ3_TILE_W - 12; c2++) {
                        int yy = nsy + r2, xx = nsx + c2;
                        if (yy >= 0 && yy < fb_h && xx >= 0 && xx < fb_w)
                            fb[yy * fb_w + xx] = 14;
                    }
            }
        }
    }
}

int dq3_scene_load_hero(dq3_scene *s, const char *assets_dir, int entry_base,
                        const int facing_order[4])
{
    int i;
    char err[128];
    for (i = 0; i < 4; i++) s->frame_for_facing[i] = facing_order ? facing_order[i] : i;
    if (dq3_charsprite_load(&s->hero, assets_dir, "DQ3MST.BLS", entry_base, err, sizeof err) != 0) {
        s->has_hero = 0;
        return -1;
    }
    s->has_hero = 1;
    return 0;
}

void dq3_scene_apply_palette(const dq3_scene *s)
{
    /* 修 bug #8(docs/28):場景進場/戰鬥返回必重套,DAC 不殘留前場景色盤。 */
    dq3_set_palette(s->pal, s->pal_count);
}

void dq3_scene_pick_open_start(dq3_scene *s)
{
    int x, y, best = -1;
    s->px = s->map_w / 2; s->py = s->map_h / 2;
    for (y = 4; y < s->map_h - 4; y++)
        for (x = 4; x < s->map_w - 4; x++) {
            int wx, wy, open = 0;
            if (!dq3_scene_walkable(s, x, y)) continue;
            for (wy = -4; wy <= 4; wy++)
                for (wx = -4; wx <= 4; wx++)
                    open += dq3_scene_walkable(s, x + wx, y + wy);
            if (open > best) { best = open; s->px = x; s->py = y; }
        }
}

void dq3_scene_free(dq3_scene *s)
{
    int i;
    if (!s) return;
    free(s->index_map);
    free(s->hi_map);
    free(s->tiles);
    for (i = 0; i < s->nowned; i++) free(s->owned[i]);
    free(s);
}
