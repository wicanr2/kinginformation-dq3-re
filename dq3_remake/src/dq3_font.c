/* dq3_font.c — D3TXT00.FON 遊戲內 16×16 點陣字型解碼。
 * 對應 docs/02:row-major、每列 2 byte、MSB-first、每字 32 byte、無 header。
 * 字身佔 row0..13(16×14),row14/15 為行距留白。日後 CJK 渲染畫在 hi-res 畫布。
 */
#include "dq3_assets.h"

int dq3_font_count(size_t n) { return (int)(n / 32); }

int dq3_font_glyph(const uint8_t *d, size_t n, int idx, uint8_t glyph[16][16])
{
    size_t base = (size_t)idx * 32;
    int y, x;
    if (idx < 0 || base + 32 > n) return -1;
    for (y = 0; y < 16; y++) {
        unsigned row = ((unsigned)d[base + y * 2] << 8) | d[base + y * 2 + 1];
        for (x = 0; x < 16; x++)
            glyph[y][x] = (uint8_t)((row >> (15 - x)) & 1);
    }
    return 0;
}
