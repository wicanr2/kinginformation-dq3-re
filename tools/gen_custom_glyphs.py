#!/usr/bin/env python3
"""自建字形:把 D3TXT00.FON 沒有的 CJK 字 rasterize 成同格式 16×16 點陣(MSB-first row-major,
字身置於 row0..13,對齊原版 16×14 + row14/15 留白),輸出 C 陣列供 remake 自建字形渲染。

用法(docker,掛 host CJK 字型):
  python tools/gen_custom_glyphs.py 設 選 項 ...
輸出:dq3_remake/src/dq3_customglyph.c(+ .h 由人工維護索引)。
字型優先 NotoSansCJK(黑體,點陣化清晰),退 uming。
"""
import sys, struct
from PIL import Image, ImageFont, ImageDraw

CHARS = sys.argv[1:] or ["設"]
FONTS = [
    "/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
    "/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc",
    "/usr/share/fonts/truetype/arphic/uming.ttc",
]
font = None
for fp in FONTS:
    try:
        font = ImageFont.truetype(fp, 15); FONT_USED = fp; break
    except Exception:
        continue
if font is None:
    print("// ERR: 找不到 CJK 字型", file=sys.stderr); sys.exit(1)


def raster(ch):
    """回 16 個 u16(row0..15,bit15=最左)。字身畫在 16×14 區(top-align),row14/15 留白。"""
    img = Image.new("L", (16, 16), 0)
    d = ImageDraw.Draw(img)
    # 量測置中(水平)+ 頂對齊到 14px 高字身
    bbox = d.textbbox((0, 0), ch, font=font)
    w = bbox[2] - bbox[0]; h = bbox[3] - bbox[1]
    ox = (16 - w) // 2 - bbox[0]
    oy = -bbox[1]  # top-align 到 row0
    d.text((ox, oy), ch, fill=255, font=font)
    rows = []
    for y in range(16):
        bits = 0
        for x in range(16):
            if y < 14 and img.getpixel((x, y)) >= 128:   # 只取字身 row0..13
                bits |= (1 << (15 - x))
        rows.append(bits)
    return rows


print("/* dq3_customglyph.c — 自建字形(D3TXT00.FON 缺的 CJK)。tools/gen_custom_glyphs.py 產;勿手改。")
print(" * 格式同原版:16×16 MSB-first row-major,字身 row0..13。索引見 dq3_customglyph.h。 */")
print('#include "dq3_customglyph.h"')
print("/* font: %s */" % FONT_USED)
print("const uint16_t dq3_customglyph_bits[][16] = {")
for ch in CHARS:
    rows = raster(ch)
    print("  { " + ", ".join("0x%04x" % r for r in rows) + " },  /* %s */" % ch)
print("};")
print("const int dq3_customglyph_count = %d;" % len(CHARS))
print("""
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
}""")
