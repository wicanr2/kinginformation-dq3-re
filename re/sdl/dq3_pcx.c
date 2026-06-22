/* dq3_pcx.c — TIT*.P 標題畫面解碼(ZSoft PCX,640x350,4 plane,RLE)。
 * 對應 tools/title_pcx.py(已驗證解碼)。把 planar 4bpp 組回 packed 索引,
 * 並取 header 內建的 16 色 EGA palette(各檔自帶)。
 */
#include "dq3_pcx.h"
#include <stdlib.h>
#include <string.h>

int dq3_pcx_decode(const uint8_t *d, size_t n,
                   uint8_t *out, int out_w, int out_h, dq3_color pal16[16])
{
    int xmin, ymin, xmax, ymax, W, H, nplanes, bpl, row_bytes;
    int i, row, plane, x;
    size_t ip;

    if (n < 128 || d[0] != 0x0A) return -1;          /* PCX magic */
    xmin = d[4] | (d[5] << 8);
    ymin = d[6] | (d[7] << 8);
    xmax = d[8] | (d[9] << 8);
    ymax = d[10] | (d[11] << 8);
    W = xmax - xmin + 1;
    H = ymax - ymin + 1;
    nplanes = d[65];
    bpl = d[66] | (d[67] << 8);
    if (W <= 0 || H <= 0 || nplanes != 4) return -2;
    if (W > out_w || H > out_h) return -3;

    /* header offset 16:16 色 EGA palette(每色 6-bit DAC 已是 0..255 範圍) */
    for (i = 0; i < 16; i++) {
        pal16[i].r = d[16 + i * 3 + 0];
        pal16[i].g = d[16 + i * 3 + 1];
        pal16[i].b = d[16 + i * 3 + 2];
    }

    row_bytes = nplanes * bpl;
    {
        uint8_t *line = (uint8_t *)malloc(row_bytes);
        if (!line) return -4;
        memset(out, 0, (size_t)out_w * out_h);
        ip = 128;
        for (row = 0; row < H; row++) {
            int filled = 0;
            /* RLE 解一整列(4 planes 連續 plane-sequential) */
            while (filled < row_bytes && ip < n) {
                uint8_t b = d[ip++];
                if ((b & 0xC0) == 0xC0) {
                    int cnt = b & 0x3F;
                    uint8_t v = (ip < n) ? d[ip++] : 0;
                    while (cnt-- > 0 && filled < row_bytes) line[filled++] = v;
                } else {
                    line[filled++] = b;
                }
            }
            /* 4 planes 組回 4bpp packed:pixel = p0 | p1<<1 | p2<<2 | p3<<3 */
            for (plane = 0; plane < nplanes; plane++) {
                const uint8_t *pb = &line[plane * bpl];
                uint8_t *orow = &out[row * out_w];
                for (x = 0; x < W; x++) {
                    int bit = (pb[x >> 3] >> (7 - (x & 7))) & 1;
                    orow[x] |= (uint8_t)(bit << plane);
                }
            }
        }
        free(line);
    }
    return 0;
}
