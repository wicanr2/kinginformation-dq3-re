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

    /* 玩家標記(暫:方框+十字;真主角 sprite 待解 DQ3MAN.BLS)。 */
    {
        int psx = (s->px - cam_x) * DQ3_TILE_W;
        int psy = (s->py - cam_y) * DQ3_TILE_H;
        int r2, c2;
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
    free(s->tiles);
    for (i = 0; i < s->nowned; i++) free(s->owned[i]);
    free(s);
}
