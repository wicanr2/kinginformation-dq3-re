/* dq3_town.c — 城鎮(CTY)場景載入器(CTY section → dq3_scene)。 */
#include "dq3_town.h"
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

static uint16_t u16le(const uint8_t *d, size_t o) { return (uint16_t)(d[o] | (d[o+1] << 8)); }

dq3_scene *dq3_town_load(const char *dir, const char *cty_name,
                         int section, int blk_n, char *err, int errcap)
{
    uint8_t *cty = NULL, *pal_raw = NULL, *blk_raw = NULL, *att_raw = NULL;
    size_t cty_len, pal_len, blk_len, att_len;
    char blk_name[32], att_name[32];
    dq3_scene *s = NULL;
    dq3_blk blk;
    int n, w, h, i, sx, sy;
    size_t so, lp, lay, tbase;

    #define FAIL(m) do { if (err) snprintf(err, errcap, "%s", m); goto fail; } while (0)

    if (blk_n < 1 || blk_n > 5) blk_n = 1;
    snprintf(blk_name, sizeof blk_name, "DQ3%d.BLK", blk_n);
    snprintf(att_name, sizeof att_name, "BLKBM%d.DAT", blk_n);

    cty     = load(dir, cty_name,  &cty_len); if (!cty)     FAIL("load CTY");
    pal_raw = load(dir, "DQ3.PAL", &pal_len); if (!pal_raw) FAIL("load DQ3.PAL");
    blk_raw = load(dir, blk_name,  &blk_len); if (!blk_raw) FAIL("load DQ3N.BLK");
    att_raw = load(dir, att_name,  &att_len); if (!att_raw) FAIL("load BLKBMN.DAT");

    /* section 偏移表 */
    if (cty_len < 2) FAIL("CTY too small");
    n = u16le(cty, 0);
    if (section < 0 || section >= n) FAIL("section out of range");
    so = u16le(cty, 2 + 2 * (size_t)section);
    if (so == 0xffff || so + 0x10 > cty_len) FAIL("section empty/oob");

    lp  = u16le(cty, so + 0x0e);          /* layout_ptr(相對 section_off) */
    lay = so + lp;
    if (lay + 8 > cty_len) FAIL("layout oob");
    w  = u16le(cty, lay);
    h  = u16le(cty, lay + 2);
    sx = cty[lay + 4];                     /* spawn_x */
    sy = cty[lay + 6];                     /* spawn_y */
    tbase = lay + 8;                       /* tile 陣列(spawn 4B 後) */
    if (w <= 0 || h <= 0 || tbase >= cty_len) FAIL("tile array oob");
    /* 容忍 tile 陣列略超檔尾(部分 section 末尾差幾 byte):逐 tile bounds-safe 讀。 */

    if (dq3_blk_open(blk_raw, blk_len, &blk) != 0) FAIL("DQ3N.BLK open");

    s = (dq3_scene *)calloc(1, sizeof *s);
    if (!s) FAIL("OOM scene");
    s->map_w = w; s->map_h = h;
    s->index_map = (uint8_t *)malloc((size_t)w * h);
    s->hi_map    = (uint8_t *)malloc((size_t)w * h);   /* 高 byte = 事件 subid 來源 */
    if (!s->index_map || !s->hi_map) FAIL("OOM map");
    for (i = 0; i < w * h; i++) {
        size_t o = tbase + 2 * (size_t)i;
        uint16_t tw = (o + 1 < cty_len) ? u16le(cty, o) : 0;   /* bounds-safe */
        s->index_map[i] = (uint8_t)tw;          /* 低 byte = BLK index */
        s->hi_map[i]    = (uint8_t)(tw >> 8);    /* 高 byte = 屬性/事件 subid */
    }

    /* section 事件表(docs/31):指標在 section+8;0xffff=無。
     * 表 = count byte + 4-byte 項[type, param u16, p2],項 subid 在 +1+subid*4。 */
    {
        uint16_t evptr = u16le(cty, so + 8);
        if (evptr != 0xffff && so + evptr < cty_len) {
            size_t et = so + evptr;
            int cnt = cty[et], k;
            if (cnt > 32) cnt = 32;
            for (k = 0; k < cnt; k++) {
                size_t o = et + 1 + (size_t)k * 4;
                if (o + 4 > cty_len) break;
                s->events[k].type  = cty[o];
                s->events[k].param = u16le(cty, o + 1);
                s->events[k].p2    = cty[o + 3];
                s->n_events = k + 1;
            }
        }
    }

    s->tile_count = blk.count;
    s->tiles = malloc((size_t)blk.count * sizeof(*s->tiles));
    if (!s->tiles) FAIL("OOM tiles");
    for (i = 0; i < blk.count; i++) dq3_blk_tile(&blk, i, s->tiles[i]);

    s->attr = (const uint16_t *)att_raw;
    s->attr_count = (int)(att_len / 2);
    s->owned[s->nowned++] = att_raw; att_raw = NULL;

    s->pal_count = dq3_pal_decode(pal_raw, pal_len, s->pal, 16);

    /* 起點:版面 spawn;不可走則挑開闊地 */
    s->px = sx; s->py = sy; s->facing = 0;
    if (sx < 0 || sx >= w || sy < 0 || sy >= h || !dq3_scene_walkable(s, sx, sy))
        dq3_scene_pick_open_start(s);

    free(cty); free(pal_raw); free(blk_raw);
    return s;

fail:
    free(cty); free(pal_raw); free(blk_raw); free(att_raw);
    if (s) dq3_scene_free(s);
    return NULL;
    #undef FAIL
}
