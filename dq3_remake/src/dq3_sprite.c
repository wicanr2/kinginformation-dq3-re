/* dq3_sprite.c — DQ3MAN.BLS 角色 sprite 解碼實作。 */
#include "dq3_sprite.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* 解一個 frame:32×24,4-plane plane-major(plane3 先)資料 + 遮罩(+0x180)。
 * 對齊 tools/bls_probe5.py:plane 順序 (3,2,1,0)、遮罩 bit=1 透明。 */
static void decode_frame(const uint8_t *d, size_t n, size_t base,
                         uint8_t px[DQ3_CHAR_H][DQ3_CHAR_W],
                         uint8_t opaque[DQ3_CHAR_H][DQ3_CHAR_W])
{
    static const int order[4] = { 3, 2, 1, 0 };  /* segment s → plane bit order[s] */
    const int wb = 4, h = DQ3_CHAR_H, seg = 4 * DQ3_CHAR_H;  /* 96 bytes/plane */
    int s, r, b, bit;

    for (r = 0; r < h; r++)
        for (b = 0; b < DQ3_CHAR_W; b++) { px[r][b] = 0; opaque[r][b] = 0; }

    /* 4 個資料 plane */
    for (s = 0; s < 4; s++) {
        size_t b0 = base + (size_t)s * seg;
        int pl = order[s];
        for (r = 0; r < h; r++)
            for (b = 0; b < wb; b++) {
                size_t o = b0 + (size_t)r * wb + b;
                uint8_t v = (o < n) ? d[o] : 0;
                for (bit = 0; bit < 8; bit++)
                    if (v & (0x80 >> bit)) px[r][b * 8 + bit] |= (uint8_t)(1 << pl);
            }
    }

    /* 遮罩:資料 +0x180 取 1 plane(96B);bit=1 → 透明,bit=0 → 不透明 */
    {
        size_t mb = base + DQ3_BLS_MASKOFF;
        for (r = 0; r < h; r++)
            for (b = 0; b < wb; b++) {
                size_t o = mb + (size_t)r * wb + b;
                uint8_t v = (o < n) ? d[o] : 0xff;
                for (bit = 0; bit < 8; bit++)
                    opaque[r][b * 8 + bit] = (v & (0x80 >> bit)) ? 0 : 1;
            }
    }
}

int dq3_charsprite_load(dq3_charsprite *cs, const char *assets_dir, const char *bls_name,
                        int entry_base, char *err, int errcap)
{
    char path[2048]; FILE *f; long sz; uint8_t *d = NULL; size_t n;
    const size_t body0 = 6;   /* 6B 表頭後為 body */
    int fr;
    #define FAIL(m) do { if (err) snprintf(err, errcap, "%s", m); free(d); return -1; } while (0)

    memset(cs, 0, sizeof *cs);
    snprintf(path, sizeof path, "%s/%s", assets_dir, bls_name);   /* 主角=DQ3MST.BLS / NPC=DQ3MAN.BLS */
    f = fopen(path, "rb");
    if (!f) FAIL("open BLS failed");
    fseek(f, 0, SEEK_END); sz = ftell(f); fseek(f, 0, SEEK_SET);
    if (sz <= 0) { fclose(f); FAIL("BLS empty"); }
    d = (uint8_t *)malloc((size_t)sz);
    if (!d || fread(d, 1, (size_t)sz, f) != (size_t)sz) { fclose(f); FAIL("read BLS"); }
    fclose(f); n = (size_t)sz;

    for (fr = 0; fr < DQ3_CHAR_FRAMES; fr++) {
        size_t base = body0 + (size_t)(entry_base + fr) * DQ3_BLS_STRIDE;
        if (base + DQ3_BLS_MASKOFF + 96 > n) FAIL("entry out of range");
        decode_frame(d, n, base, cs->px[fr], cs->opaque[fr]);
    }
    cs->loaded = 1;
    free(d);
    return 0;
    #undef FAIL
}
