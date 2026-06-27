/* dq3_customglyph.c — 自建字形(D3TXT00.FON 缺的 CJK)。tools/gen_custom_glyphs.py 產;勿手改。
 * 格式同原版:16×16 MSB-first row-major,字身 row0..13。索引見 dq3_customglyph.h。 */
#include "dq3_customglyph.h"
/* font: /usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc */
const uint16_t dq3_customglyph_bits[][16] = {
  { 0x7df0, 0x0190, 0x7d10, 0x0114, 0x7f1c, 0x0200, 0x7ff8, 0x0108, 0x7c98, 0x44d0, 0x4460, 0x4470, 0x7d9c, 0x4204, 0x0000, 0x0000 },  /* 設 */
  { 0x0800, 0x29fc, 0x2900, 0x2900, 0x2900, 0x3ffc, 0x21c4, 0x2148, 0x7d48, 0x4528, 0x4530, 0x4530, 0x4538, 0x464c, 0x0000, 0x0000 },  /* 版 */
};
const int dq3_customglyph_count = 2;

void dq3_customglyph_draw(int idx, uint8_t *fb, int fb_w, int fb_h, int x, int y, uint8_t fg)
{
    int r, c;
    if (idx < 0 || idx >= dq3_customglyph_count) return;
    for (r = 0; r < 16; r++) {
        uint16_t bits = dq3_customglyph_bits[idx][r];
        for (c = 0; c < 16; c++)
            if (bits & (1 << (15 - c))) {
                int xx = x + c, yy = y + r;
                if (xx >= 0 && xx < fb_w && yy >= 0 && yy < fb_h) fb[yy * fb_w + xx] = fg;
            }
    }
}
