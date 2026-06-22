/* dq3_pal.c — DQ3.PAL / MNSBK.PAL 解碼(VGA 6-bit DAC → 8-bit RGB)。
 * 對應 tools/tile_lib.py load_pal。每色 3 byte R,G,B,值域 0..63。
 */
#include "dq3_assets.h"

int dq3_pal_decode(const uint8_t *d, size_t n, dq3_color *pal, int max)
{
    int count = 0;
    size_t i;
    for (i = 0; i + 2 < n && count < max; i += 3) {
        uint8_t r = d[i], g = d[i + 1], b = d[i + 2];
        pal[count].r = (uint8_t)((r << 2) | (r >> 4));
        pal[count].g = (uint8_t)((g << 2) | (g >> 4));
        pal[count].b = (uint8_t)((b << 2) | (b >> 4));
        count++;
    }
    return count;
}
