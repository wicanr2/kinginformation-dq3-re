#!/usr/bin/env python3
"""cty_loc 表(DGROUP 0x748)→ 世界地圖標記。攻略反推地點用。
   輸出 docs/maps/world_{con,und}_cty.png。詳見 docs/32。
   執行:tools/dockrun.sh tools/overlay_cty_map.py(PIL 在 docker uv venv)。"""
import struct
from PIL import Image, ImageDraw, ImageFont

d = open('assets_raw/DQ3.EXE', 'rb').read(); B = 0x16140 + 0x748
ent = [(i, d[B+i*4], d[B+i*4+1], struct.unpack_from('<H', d, B+i*4+2)[0]) for i in range(98)]

def font(sz):
    try: return ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf', sz)
    except Exception: return ImageFont.load_default()

def mark(img, tw, th, pts, out):
    im = Image.open(img).convert('RGB'); W, H = im.size; sx, sy = W/tw, H/th
    dr = ImageDraw.Draw(im); F = font(26); n = 0
    for (i, x, yo) in pts:
        px, py = x*sx, yo*sy
        if not (0 <= px < W and 0 <= py < H): continue
        dr.ellipse([px-16, py-16, px+16, py+16], outline=(255, 0, 0), width=4)
        dr.text((px+18, py-16), str(i), fill=(255, 255, 0), font=F, stroke_width=3, stroke_fill=(0, 0, 0)); n += 1
    im.save(out); print('%s: %d 點' % (out, n))

# 統一垂直座標:上層 y<205,下層 269<=y<436(local_y=y-269)
over = [(i, x, y) for (i, x, b1, y) in ent if 0 < x < 244 and 0 < y < 205 and b1 == 0]
und  = [(i, x, y-269) for (i, x, b1, y) in ent if 0 < x < 244 and 269 <= y < 436 and b1 == 0]
mark('docs/maps/world_con_full.png', 244, 205, over, 'docs/maps/world_con_cty.png')
mark('docs/maps/world_und_full.png', 244, 167, und, 'docs/maps/world_und_cty.png')
