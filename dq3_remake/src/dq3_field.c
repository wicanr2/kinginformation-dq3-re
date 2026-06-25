/* dq3_field.c — 地表場景載入器(素材 → dq3_scene)。 */
#include "dq3_field.h"
#include "dq3_assets.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static uint8_t *load(const char *dir, const char *name, size_t *len)
{
    char path[2048]; FILE *f; long sz; uint8_t *buf;
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

dq3_scene *dq3_field_load(const char *dir, char *err, int errcap)
{
    return dq3_field_load_map(dir, "DQ3CON.MAP", err, errcap);   /* 地表 */
}

dq3_scene *dq3_field_load_map(const char *dir, const char *map_name, char *err, int errcap)
{
    uint8_t *pal_raw = NULL, *blk_raw = NULL, *map_raw = NULL, *att_raw = NULL;
    size_t pal_len, blk_len, map_len, att_len;
    dq3_scene *s = NULL;
    dq3_blk blk;
    int map_w, map_h, i, n;
    const uint8_t *map;

    #define FAIL(m) do { if (err) snprintf(err, errcap, "%s", m); goto fail; } while (0)

    pal_raw = load(dir, "DQ3.PAL",    &pal_len); if (!pal_raw) FAIL("load DQ3.PAL");
    blk_raw = load(dir, "DQ3.BLK",    &blk_len); if (!blk_raw) FAIL("load DQ3.BLK");
    map_raw = load(dir, map_name,     &map_len); if (!map_raw) FAIL("load overworld map");
    att_raw = load(dir, "BLKBM.DAT",  &att_len); if (!att_raw) FAIL("load BLKBM.DAT");

    if (map_len < 4) FAIL("overworld map too small");
    map_w = map_raw[0] | (map_raw[1] << 8);
    map_h = map_raw[2] | (map_raw[3] << 8);
    map   = map_raw + 4;
    if ((size_t)map_w * map_h + 4 > map_len) FAIL("DQ3CON.MAP size mismatch");
    if (dq3_blk_open(blk_raw, blk_len, &blk) != 0) FAIL("DQ3.BLK open");

    s = (dq3_scene *)calloc(1, sizeof *s);
    if (!s) FAIL("OOM scene");

    s->map_w = map_w; s->map_h = map_h;
    n = map_w * map_h;
    s->index_map = (uint8_t *)malloc((size_t)n);
    if (!s->index_map) FAIL("OOM index_map");
    memcpy(s->index_map, map, (size_t)n);          /* 世界地圖已是 u8 index */

    s->tile_count = blk.count;
    s->tiles = malloc((size_t)blk.count * sizeof(*s->tiles));
    if (!s->tiles) FAIL("OOM tiles");
    for (i = 0; i < blk.count; i++) dq3_blk_tile(&blk, i, s->tiles[i]);

    /* 屬性表(scene 接管 att_raw 以保持指標有效) */
    s->attr = (const uint16_t *)att_raw;
    s->attr_count = (int)(att_len / 2);
    s->owned[s->nowned++] = att_raw; att_raw = NULL;

    /* palette(海面 idx2/10 沙色→藍,等價 runtime DAC) */
    s->pal_count = dq3_pal_decode(pal_raw, pal_len, s->pal, 16);
    /* sprite 色盤 bank = 海面動畫覆蓋「之前」的靜態色盤(idx2=膚色,docs/51)。
     * 主角/NPC 用此 bank,海藍不會染到膚色。 */
    { int i; for (i = 0; i < s->pal_count; i++) s->spal[i] = s->pal[i]; s->spal_count = s->pal_count; }
    if (s->pal_count >= 11) {
        dq3_color blue = { 0, 85, 223 };
        s->pal[2] = blue; s->pal[10] = blue;   /* 只覆蓋背景 tile 色盤(slot 2/10),不動 spal */
    }

    dq3_scene_pick_open_start(s);
    s->facing = 0;

    free(pal_raw); free(blk_raw); free(map_raw);
    return s;

fail:
    free(pal_raw); free(blk_raw); free(map_raw); free(att_raw);
    if (s) dq3_scene_free(s);
    return NULL;
    #undef FAIL
}
