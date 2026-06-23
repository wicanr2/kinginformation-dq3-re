#!/usr/bin/env python3
"""cty_loc(0x748)→ 世界地圖標記。下層 local_y = raw_Y - 300(經建築 sprite 驗證)。
   輸出 docs/maps/world_{con,und}_cty.png。詳見 docs/32。"""
import struct
from PIL import Image, ImageDraw, ImageFont
d = open('assets_raw/DQ3.EXE', 'rb').read(); B = 0x16140 + 0x748
ent = [(i, d[B+i*4], struct.unpack_from('<H', d, B+i*4+2)[0]) for i in range(100)]
def font(sz):
    try: return ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf', sz)
    except Exception: return ImageFont.load_default()
def mark(img, tw, th, pts, out):
    im = Image.open(img).convert('RGB'); W, H = im.size; sx, sy = W/tw, H/th
    dr = ImageDraw.Draw(im); F = font(26); n = 0
    for (i, x, yy) in pts:
        px, py = x*sx, yy*sy
        if not (0 <= px < W and 0 <= py < H): continue
        dr.ellipse([px-14, py-14, px+14, py+14], outline=(255, 0, 0), width=4)
        dr.text((px+15, py-15), str(i), fill=(255, 255, 0), font=F, stroke_width=3, stroke_fill=(0, 0, 0)); n += 1
    im.save(out); print('%s: %d 點' % (out, n))
over = [(i, x, y) for (i, x, y) in ent if 0 < x < 244 and 0 < y < 205]          # 地面 raw Y
und  = [(i, x, y-300) for (i, x, y) in ent if 0 < x < 244 and 300 < y < 467]    # 下層 Y-300
mark('docs/maps/world_con_full.png', 244, 205, over, 'docs/maps/world_con_cty.png')
mark('docs/maps/world_und_full.png', 244, 167, und, 'docs/maps/world_und_cty.png')
