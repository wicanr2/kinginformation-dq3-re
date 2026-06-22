/* dq3_field.c — 地表場景引擎實作。
 *
 * 載入 → 預解 tile → 攝影機跟隨玩家 → 視窗貼圖 → bit0 碰撞走動。
 * 全部邏輯藏在此檔;對外只透過 dq3_field.h 的窄介面。
 *
 * 座標系(對齊 re/player.c):
 *   - 玩家以 tile 為單位(px,py),非像素;每步 ±1 tile。
 *   - 目標 tile = (px+dx, py+dy);出界視為不可走(地表此處原版會觸發換圖,
 *     remake 第一步先擋住,換圖留待事件系統)。
 *   - 屬性 = blkbm[map_tile_index](u16);bit0=1 阻擋。
 */
#include "dq3_field.h"
#include "dq3_assets.h"
#include "dq3_runtime.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* 視窗 tile 數:原版地表寬 0x14=20 tile(docs/11)。高取滿畫面。 */
#define TILE_W 32
#define TILE_H 24
#define VIEW_COLS (DQ3_SCREEN_W / TILE_W)               /* 640/32 = 20 */
#define VIEW_ROWS ((DQ3_SCREEN_H + TILE_H - 1) / TILE_H) /* ceil(350/24)=15 */

struct dq3_field {
    /* 素材原始 buffer(保留以便 body 指標有效) */
    uint8_t *blk_raw;  size_t blk_len;
    uint8_t *map_raw;  size_t map_len;
    uint8_t *pal_raw;  size_t pal_len;
    uint8_t *att_raw;  size_t att_len;

    dq3_blk  blk;                 /* tile 圖庫 header + body */
    int      map_w, map_h;
    const uint8_t *map;           /* map_raw + 4(row-major u8 tile index) */
    const uint16_t *attr;         /* 屬性表(u16/tile),little-endian */
    int      attr_count;

    dq3_color pal[16];            /* 地表 16 色(含海面修正) */
    int       pal_count;

    /* 預解 tile 像素:tiles[idx][row][col] = palette 索引 */
    uint8_t (*tiles)[TILE_H][TILE_W];
    int      tile_count;

    int px, py;                   /* 玩家 tile 座標 */
    int facing;                   /* 0下 1左 2上 3右 */
};

static uint8_t *load(const char *dir, const char *name, size_t *len)
{
    char path[2048];
    FILE *f; long sz; uint8_t *buf;
    snprintf(path, sizeof(path), "%s/%s", dir, name);
    f = fopen(path, "rb");
    if (!f) { *len = 0; return NULL; }
    fseek(f, 0, SEEK_END); sz = ftell(f); fseek(f, 0, SEEK_SET);
    if (sz <= 0) { fclose(f); *len = 0; return NULL; }
    buf = (uint8_t *)malloc((size_t)sz);
    if (!buf || fread(buf, 1, (size_t)sz, f) != (size_t)sz) {
        fclose(f); free(buf); *len = 0; return NULL;
    }
    fclose(f); *len = (size_t)sz;
    return buf;
}

/* tile 是否可走:界外不可走;屬性 bit0 阻擋。 */
static int walkable(const dq3_field *f, int tx, int ty)
{
    int idx; uint16_t a;
    if (tx < 0 || ty < 0 || tx >= f->map_w || ty >= f->map_h) return 0;
    idx = f->map[ty * f->map_w + tx];
    if (idx >= f->attr_count) return 1;          /* 無屬性資料 → 預設可走 */
    a = f->attr[idx];
    return (a & 0x0001) ? 0 : 1;                 /* bit0=阻擋 */
}

dq3_field *dq3_field_create(const char *assets_dir, char *err, int errcap)
{
    dq3_field *f = (dq3_field *)calloc(1, sizeof *f);
    int i;
    if (!f) return NULL;
    #define FAIL(msg) do { if (err) snprintf(err, errcap, "%s", msg); \
                           dq3_field_destroy(f); return NULL; } while (0)

    f->pal_raw = load(assets_dir, "DQ3.PAL", &f->pal_len);
    if (!f->pal_raw) FAIL("load DQ3.PAL failed");
    f->blk_raw = load(assets_dir, "DQ3.BLK", &f->blk_len);
    if (!f->blk_raw) FAIL("load DQ3.BLK failed");
    f->map_raw = load(assets_dir, "DQ3CON.MAP", &f->map_len);
    if (!f->map_raw) FAIL("load DQ3CON.MAP failed");
    f->att_raw = load(assets_dir, "BLKBM.DAT", &f->att_len);
    if (!f->att_raw) FAIL("load BLKBM.DAT failed");

    /* palette(地表 16 色取前 16;海面 idx2/10 由靜態沙色改藍,等價 runtime DAC) */
    f->pal_count = dq3_pal_decode(f->pal_raw, f->pal_len, f->pal, 16);
    if (f->pal_count >= 11) {
        dq3_color blue = { 0, 85, 223 };          /* = 海岸線 idx5,docs/04 */
        f->pal[2] = blue; f->pal[10] = blue;
    }

    /* map header:u16 w, u16 h */
    if (f->map_len < 4) FAIL("DQ3CON.MAP too small");
    f->map_w = f->map_raw[0] | (f->map_raw[1] << 8);
    f->map_h = f->map_raw[2] | (f->map_raw[3] << 8);
    f->map   = f->map_raw + 4;
    if ((size_t)f->map_w * f->map_h + 4 > f->map_len) FAIL("DQ3CON.MAP size mismatch");

    /* 屬性表:u16/tile */
    f->attr = (const uint16_t *)f->att_raw;       /* host LE == file LE */
    f->attr_count = (int)(f->att_len / 2);

    /* tile 圖庫 + 預解 */
    if (dq3_blk_open(f->blk_raw, f->blk_len, &f->blk) != 0) FAIL("DQ3.BLK open failed");
    f->tile_count = f->blk.count;
    f->tiles = malloc((size_t)f->tile_count * sizeof(*f->tiles));
    if (!f->tiles) FAIL("OOM tiles");
    for (i = 0; i < f->tile_count; i++)
        dq3_blk_tile(&f->blk, i, f->tiles[i]);

    /* 玩家起點:挑「9×9 視窗內可走鄰居最多」的可走 tile(落在開闊陸地,
     * 不被山林/海包死),讓走動可明顯捲動。掃全圖,取最佳。 */
    {
        int x, y, best = -1;
        f->px = f->map_w / 2; f->py = f->map_h / 2;
        for (y = 4; y < f->map_h - 4; y++)
            for (x = 4; x < f->map_w - 4; x++) {
                int wx, wy, open = 0;
                if (!walkable(f, x, y)) continue;
                for (wy = -4; wy <= 4; wy++)
                    for (wx = -4; wx <= 4; wx++)
                        open += walkable(f, x + wx, y + wy);
                if (open > best) { best = open; f->px = x; f->py = y; }
            }
    }
    f->facing = 0;
    return f;
    #undef FAIL
}

void dq3_field_destroy(dq3_field *f)
{
    if (!f) return;
    free(f->blk_raw); free(f->map_raw); free(f->pal_raw); free(f->att_raw);
    free(f->tiles);
    free(f);
}

int dq3_field_input(dq3_field *f, uint8_t sc)
{
    int dx = 0, dy = 0, tx, ty;
    switch (sc) {
        case 0x48: dy = -1; f->facing = 2; break;   /* 上 */
        case 0x50: dy =  1; f->facing = 0; break;   /* 下 */
        case 0x4b: dx = -1; f->facing = 1; break;   /* 左 */
        case 0x4d: dx =  1; f->facing = 3; break;   /* 右 */
        default: return 0;
    }
    tx = f->px + dx; ty = f->py + dy;
    if (!walkable(f, tx, ty)) return 0;             /* 撞牆:轉向但不移動 */
    f->px = tx; f->py = ty;
    return 1;
}

void dq3_field_render(dq3_field *f, uint8_t *fb, int fb_w, int fb_h)
{
    /* 攝影機左上角 tile:玩家置中,夾在地圖邊界內 */
    int cam_x = f->px - VIEW_COLS / 2;
    int cam_y = f->py - VIEW_ROWS / 2;
    int row, col, r, c;
    if (cam_x < 0) cam_x = 0;
    if (cam_y < 0) cam_y = 0;
    if (cam_x > f->map_w - VIEW_COLS) cam_x = f->map_w - VIEW_COLS;
    if (cam_y > f->map_h - VIEW_ROWS) cam_y = f->map_h - VIEW_ROWS;
    if (cam_x < 0) cam_x = 0;
    if (cam_y < 0) cam_y = 0;

    memset(fb, 0, (size_t)fb_w * fb_h);

    for (row = 0; row < VIEW_ROWS; row++) {
        int my = cam_y + row;
        int sy = row * TILE_H;
        if (my < 0 || my >= f->map_h) continue;
        for (col = 0; col < VIEW_COLS; col++) {
            int mx = cam_x + col;
            int sx = col * TILE_W;
            int idx;
            const uint8_t (*t)[TILE_W];
            if (mx < 0 || mx >= f->map_w) continue;
            idx = f->map[my * f->map_w + mx];
            if (idx < 0 || idx >= f->tile_count) continue;
            t = f->tiles[idx];
            for (r = 0; r < TILE_H; r++) {
                int yy = sy + r;
                if (yy < 0 || yy >= fb_h) continue;
                for (c = 0; c < TILE_W; c++) {
                    int xx = sx + c;
                    if (xx < 0 || xx >= fb_w) continue;
                    fb[yy * fb_w + xx] = t[r][c];
                }
            }
        }
    }

    /* 玩家標記(暫以方框+十字;真主角 sprite 待解 DQ3MAN.BLS)。
     * 螢幕位置 = (玩家 tile - 攝影機) × tile 尺寸。 */
    {
        int psx = (f->px - cam_x) * TILE_W;
        int psy = (f->py - cam_y) * TILE_H;
        int r2, c2;
        uint8_t mark = 15;   /* 高亮色 */
        for (r2 = 4; r2 < TILE_H - 4; r2++)
            for (c2 = 8; c2 < TILE_W - 8; c2++) {
                int yy = psy + r2, xx = psx + c2;
                int edge = (r2 == 4 || r2 == TILE_H-5 || c2 == 8 || c2 == TILE_W-9);
                int cross = (c2 == TILE_W/2 || r2 == TILE_H/2);
                if ((edge || cross) && yy >= 0 && yy < fb_h && xx >= 0 && xx < fb_w)
                    fb[yy * fb_w + xx] = mark;
            }
    }
}

const dq3_color *dq3_field_palette(const dq3_field *f, int *count)
{
    if (count) *count = f->pal_count;
    return f->pal;
}

void dq3_field_get_player(const dq3_field *f, int *tx, int *ty)
{
    if (tx) *tx = f->px;
    if (ty) *ty = f->py;
}

void dq3_field_set_player(dq3_field *f, int tx, int ty)
{
    f->px = tx; f->py = ty;
}
