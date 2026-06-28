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

int dq3_scene_npc_at(const dq3_scene *s, int tx, int ty)
{
    int i;
    if (tx < 0 || ty < 0 || tx >= s->map_w || ty >= s->map_h) return -1;
    for (i = 0; i < s->n_npcs; i++)
        if (s->npcs[i].x == tx && s->npcs[i].y == ty) return i;
    return -1;
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
    if (dq3_scene_npc_at(s, tx, ty) >= 0) return 0;   /* NPC 擋路(facing 已更新 → 撞即面向,可 Enter 對話/開店)*/
    {   /* 隊列 trail:把主角「移動前」的格 unshift 進歷史 → follower i 站 trail[i] */
        int k;
        for (k = 7; k > 0; k--) { s->trail_x[k]=s->trail_x[k-1]; s->trail_y[k]=s->trail_y[k-1]; s->trail_f[k]=s->trail_f[k-1]; }
        s->trail_x[0]=s->px; s->trail_y[0]=s->py; s->trail_f[0]=s->facing & 3;
        if (s->trail_len < 8) s->trail_len++;
    }
    s->px = tx; s->py = ty;
    return 1;
}

/* 在地圖格 (mx,my) 疊一個 BLK tile(用與 render 相同攝影機)。船 sprite overlay 用(docs/51)。
 * 整格不透明 blit;船 tile 自帶海背景,疊在海格上無縫。 */
void dq3_scene_draw_tile_at(const dq3_scene *s, uint8_t *fb, int fb_w, int fb_h,
                            int mx, int my, int tile_idx)
{
    int cam_x = s->px - VIEW_COLS / 2;
    int cam_y = s->py - VIEW_ROWS / 2;
    int sx, sy, r, c;
    const uint8_t (*t)[DQ3_TILE_W];
    if (cam_x > s->map_w - VIEW_COLS) cam_x = s->map_w - VIEW_COLS;
    if (cam_y > s->map_h - VIEW_ROWS) cam_y = s->map_h - VIEW_ROWS;
    if (cam_x < 0) cam_x = 0;
    if (cam_y < 0) cam_y = 0;
    if (tile_idx < 0 || tile_idx >= s->tile_count) return;
    sx = (mx - cam_x) * DQ3_TILE_W;
    sy = (my - cam_y) * DQ3_TILE_H;
    t = s->tiles[tile_idx];
    for (r = 0; r < DQ3_TILE_H; r++)
        for (c = 0; c < DQ3_TILE_W; c++) {
            int yy = sy + r, xx = sx + c;
            if (yy >= 0 && yy < fb_h && xx >= 0 && xx < fb_w)
                fb[yy * fb_w + xx] = t[r][c];
        }
}

/* 在地圖格 (mx,my) 疊「已開啟寶箱」標記:上 1/4 開蓋暗線 + 其餘棋盤 dither 變暗。
 * ★ remake 增強(非還原):原版寶箱取後不翻 tile(docs/31 handler 0x9ec1,re-examine 顯
 * 「可惜是空的」);使用者要求加開過回饋。程式疊繪、不動 BLK tile,變暗色取 bg 色盤最暗 index。 */
void dq3_scene_mark_opened_tile(const dq3_scene *s, uint8_t *fb, int fb_w, int fb_h, int mx, int my)
{
    int cam_x = s->px - VIEW_COLS / 2;
    int cam_y = s->py - VIEW_ROWS / 2;
    int sx, sy, r, c, di = 0, i, best = 1 << 30;
    if (cam_x > s->map_w - VIEW_COLS) cam_x = s->map_w - VIEW_COLS;
    if (cam_y > s->map_h - VIEW_ROWS) cam_y = s->map_h - VIEW_ROWS;
    if (cam_x < 0) cam_x = 0;
    if (cam_y < 0) cam_y = 0;
    for (i = 0; i < s->pal_count && i < 16; i++) {        /* 最暗 bg 色 index(luminance 最小)*/
        int lum = s->pal[i].r + s->pal[i].g + s->pal[i].b;
        if (lum < best) { best = lum; di = i; }
    }
    sx = (mx - cam_x) * DQ3_TILE_W;
    sy = (my - cam_y) * DQ3_TILE_H;
    for (r = 0; r < DQ3_TILE_H; r++)
        for (c = 0; c < DQ3_TILE_W; c++) {
            int yy = sy + r, xx = sx + c;
            if (yy < 0 || yy >= fb_h || xx < 0 || xx >= fb_w) continue;
            if (r < DQ3_TILE_H / 4) {                      /* 上 1/4:開蓋暗線(較密)*/
                if (((r ^ c) & 1) == 0) fb[yy * fb_w + xx] = (uint8_t)di;
            } else if (((r + c) & 1) == 0) {               /* 其餘:棋盤 dither 變暗 */
                fb[yy * fb_w + xx] = (uint8_t)di;
            }
        }
}

void dq3_scene_draw_charsprite_at(const dq3_scene *s, uint8_t *fb, int fb_w, int fb_h,
                                  int mx, int my, const dq3_charsprite *cs, int frame)
{
    int cam_x = s->px - VIEW_COLS / 2;
    int cam_y = s->py - VIEW_ROWS / 2;
    int sx, sy, r, c;
    if (cam_x > s->map_w - VIEW_COLS) cam_x = s->map_w - VIEW_COLS;
    if (cam_y > s->map_h - VIEW_ROWS) cam_y = s->map_h - VIEW_ROWS;
    if (cam_x < 0) cam_x = 0;
    if (cam_y < 0) cam_y = 0;
    if (frame < 0 || frame >= DQ3_CHAR_FRAMES) frame = 0;
    sx = (mx - cam_x) * DQ3_TILE_W;
    sy = (my - cam_y) * DQ3_TILE_H;
    for (r = 0; r < DQ3_CHAR_H; r++)
        for (c = 0; c < DQ3_CHAR_W; c++) {
            int yy = sy + r, xx = sx + c;
            if (!cs->opaque[frame][r][c]) continue;
            if (yy >= 0 && yy < fb_h && xx >= 0 && xx < fb_w)
                fb[yy * fb_w + xx] = (uint8_t)(DQ3_SPRITE_PAL_BASE + cs->px[frame][r][c]);
        }
}

/* 待機/行走動畫:全域計時器,每 DQ3_ANIM_PERIOD 次 render 切一次 walk 相位(0/1)。
 * DQ3 角色(主角 + NPC)即使站著手腳也持續左右擺動 → 連續交替兩個 walk sub-frame。 */
#define DQ3_ANIM_PERIOD 9
static unsigned g_anim_tick = 0;

void dq3_scene_render(const dq3_scene *s, uint8_t *fb, int fb_w, int fb_h)
{
    int cam_x = s->px - VIEW_COLS / 2;
    int cam_y = s->py - VIEW_ROWS / 2;
    int row, col, r, c;
    int walk = (++g_anim_tick / DQ3_ANIM_PERIOD) & 1;   /* 0/1 交替的 walk 相位 */

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
            int dir = s->frame_for_facing[s->facing & 3] & 3;
            int fr = dir * DQ3_CHAR_WALK + walk;          /* 方向×2 + walk 相位 → 手腳擺動 */
            if (fr < 0 || fr >= DQ3_CHAR_FRAMES) fr = 0;
            for (r2 = 0; r2 < DQ3_CHAR_H; r2++)
                for (c2 = 0; c2 < DQ3_CHAR_W; c2++) {
                    int yy = psy + r2, xx = psx + c2;
                    if (!s->hero.opaque[fr][r2][c2]) continue;   /* 透明像素跳過 */
                    if (yy >= 0 && yy < fb_h && xx >= 0 && xx < fb_w)
                        fb[yy * fb_w + xx] = (uint8_t)(DQ3_SPRITE_PAL_BASE + s->hero.px[fr][r2][c2]);
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

    /* 隊列 follower train:隊員跟在主角後面排成一列(follower j 站 trail[j]=主角 j+1 步前的格);
     * 陣亡者以棺材繪、排到隊尾(活的先佔近位)。 */
    if (s->n_followers > 0) {
        int order[3], no = 0, j, k2;
        for (k2 = 0; k2 < s->n_followers; k2++) if (!s->follower_dead[k2]) order[no++] = k2;  /* 活的先 */
        for (k2 = 0; k2 < s->n_followers; k2++) if ( s->follower_dead[k2]) order[no++] = k2;  /* 死的後 */
        for (j = 0; j < no; j++) {
            int fi = order[j];
            int tx = (s->trail_len > j) ? s->trail_x[j] : s->px;
            int ty = (s->trail_len > j) ? s->trail_y[j] : s->py;
            int tf = (s->trail_len > j) ? (s->trail_f[j] & 3) : (s->facing & 3);
            int fsx = (tx - cam_x) * DQ3_TILE_W, fsy = (ty - cam_y) * DQ3_TILE_H, r2, c2;
            if (s->follower_dead[fi]) {                       /* 棺材:暗箱 + 亮邊 + 十字 */
                for (r2 = 5; r2 < DQ3_CHAR_H - 2; r2++)
                    for (c2 = 9; c2 < DQ3_CHAR_W - 9; c2++) {
                        int yy = fsy + r2, xx = fsx + c2;
                        int edge   = (r2==5 || r2==DQ3_CHAR_H-3 || c2==9 || c2==DQ3_CHAR_W-10);
                        int crossv = (c2==DQ3_CHAR_W/2 && r2>=8 && r2<16);
                        int crossh = (r2==10 && c2>=DQ3_CHAR_W/2-3 && c2<=DQ3_CHAR_W/2+3);
                        if (yy>=0 && yy<fb_h && xx>=0 && xx<fb_w)
                            fb[yy*fb_w+xx] = (uint8_t)((edge||crossv||crossh) ? 15 : 0);
                    }
            } else if (s->follower_spr[fi].loaded) {          /* 活的:職業 sprite + 走路動畫 */
                int dir = s->frame_for_facing[tf] & 3;
                int fr = dir * DQ3_CHAR_WALK + walk;
                if (fr < 0 || fr >= DQ3_CHAR_FRAMES) fr = 0;
                for (r2 = 0; r2 < DQ3_CHAR_H; r2++)
                    for (c2 = 0; c2 < DQ3_CHAR_W; c2++) {
                        int yy = fsy + r2, xx = fsx + c2;
                        if (!s->follower_spr[fi].opaque[fr][r2][c2]) continue;
                        if (yy>=0 && yy<fb_h && xx>=0 && xx<fb_w)
                            fb[yy*fb_w+xx] = (uint8_t)(DQ3_SPRITE_PAL_BASE + s->follower_spr[fi].px[fr][r2][c2]);
                    }
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
                int dir = s->frame_for_facing[npc->ctrl & 3] & 3;
                int fr = dir * DQ3_CHAR_WALK + walk;       /* 方向×2 + walk 相位 → 站著也擺手 */
                const dq3_charsprite *cs = &s->npc_spr[ci];
                if (fr < 0 || fr >= DQ3_CHAR_FRAMES) fr = 0;
                for (r2 = 0; r2 < DQ3_CHAR_H; r2++)
                    for (c2 = 0; c2 < DQ3_CHAR_W; c2++) {
                        int yy = nsy + r2, xx = nsx + c2;
                        if (!cs->opaque[fr][r2][c2]) continue;
                        if (yy >= 0 && yy < fb_h && xx >= 0 && xx < fb_w)
                            fb[yy * fb_w + xx] = (uint8_t)(DQ3_SPRITE_PAL_BASE + cs->px[fr][r2][c2]);
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

void dq3_scene_reset_trail(dq3_scene *s)
{
    int i;
    for (i = 0; i < 8; i++) { s->trail_x[i] = s->px; s->trail_y[i] = s->py; s->trail_f[i] = s->facing & 3; }
    s->trail_len = 0;
}

void dq3_scene_set_followers(dq3_scene *s, const char *assets_dir,
                             const int *entries, const int *dead, int n)
{
    int i; char err[256];
    if (n < 0) n = 0; if (n > 3) n = 3;
    s->n_followers = n;
    for (i = 0; i < n; i++) {
        s->follower_dead[i] = dead ? dead[i] : 0;
        if (dq3_charsprite_load(&s->follower_spr[i], assets_dir, "DQ3MST.BLS", entries[i], err, sizeof err) != 0)
            s->follower_spr[i].loaded = 0;
    }
    dq3_scene_reset_trail(s);
}

/* 晝夜相位(0=白天 1=黃昏 2=黑夜 3=黎明)。overworld 步數驅動(main.c);set 後下次 apply 生效。 */
static int g_dn_phase = 0;
void dq3_scene_set_daynight(int phase) { g_dn_phase = phase & 3; }
int  dq3_scene_get_daynight(void) { return g_dn_phase; }
/* 相位 → RGB 縮放%(bg 用;sprite 用較淺版以保持角色可見)。夜暗偏藍、黃昏暗偏暖、黎明微暗偏冷。 */
static void dn_scale(int phase, int *r, int *g, int *b) {
    switch (phase) {
        case 2: *r=42; *g=44; *b=70; break;    /* 黑夜 */
        case 1: *r=82; *g=62; *b=58; break;    /* 黃昏 */
        case 3: *r=72; *g=74; *b=92; break;    /* 黎明 */
        default:*r=100;*g=100;*b=100; break;   /* 白天:原色 */
    }
}
static dq3_color dn_tint(dq3_color c, int rs, int gs, int bs) {
    dq3_color o;
    o.r = (uint8_t)(c.r * rs / 100); o.g = (uint8_t)(c.g * gs / 100); o.b = (uint8_t)(c.b * bs / 100);
    return o;
}

void dq3_scene_apply_palette(const dq3_scene *s)
{
    /* 修 bug #8(docs/28):場景進場/戰鬥返回必重套,DAC 不殘留前場景色盤。
     * slot 0-15 = 背景 tile 色盤;slot 16-31 = sprite 色盤 bank(無海面覆蓋,docs/51)。
     * 晝夜:bg 依相位調暗;sprite 用「淺一半」的暗化(角色夜間仍看得見)。 */
    dq3_color combined[32];
    int i, rs, gs, bs, srs, sgs, sbs;
    dn_scale(g_dn_phase, &rs, &gs, &bs);
    srs = (rs + 100) / 2; sgs = (gs + 100) / 2; sbs = (bs + 100) / 2;   /* sprite 較淺暗化 */
    for (i = 0; i < 16; i++) {
        dq3_color bg = (i < s->pal_count) ? s->pal[i] : (dq3_color){0,0,0};
        dq3_color sp = (i < s->spal_count) ? s->spal[i]
                     : ((i < s->pal_count) ? s->pal[i] : (dq3_color){0,0,0});
        combined[i]      = dn_tint(bg, rs, gs, bs);
        combined[16 + i] = dn_tint(sp, srs, sgs, sbs);
    }
    dq3_set_palette(combined, 32);
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
