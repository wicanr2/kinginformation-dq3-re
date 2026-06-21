#!/usr/bin/env python3
"""放大檢視 CHINA.FON 單一 section 的前 N 個字模,並嘗試多種 (width,stride,bitorder)
組合,找出正確的點陣排列。"""
import struct, os
from PIL import Image

SRC = "assets_raw/CHINA.FON"
OUT = "assets_out/font_probe"
os.makedirs(OUT, exist_ok=True)
data = open(SRC, "rb").read()
first = struct.unpack_from("<I", data, 0)[0]
n = first // 4
offs = [struct.unpack_from("<I", data, i*4)[0] for i in range(n)] + [len(data)]

def render(blob, w, h, bpr, n_glyphs, cols, scale, msb=True):
    glyph_bytes = bpr * h
    n_glyphs = min(n_glyphs, len(blob)//glyph_bytes)
    rows = (n_glyphs + cols - 1)//cols
    pad = 1
    img = Image.new("L", (cols*(w+pad)+pad, rows*(h+pad)+pad), 64)
    px = img.load()
    for g in range(n_glyphs):
        gx=(g%cols)*(w+pad)+pad; gy=(g//cols)*(h+pad)+pad
        base=g*glyph_bytes
        for yy in range(h):
            for xx in range(w):
                byte=blob[base+yy*bpr+xx//8]
                bit=(byte>>(7-(xx&7)))&1 if msb else (byte>>(xx&7))&1
                if bit: px[gx+xx,gy+yy]=255
    return img.resize((img.width*scale, img.height*scale), Image.NEAREST)

# sec2: 81*72, 試 24x24 (bpr=3) 標準
blob = data[offs[2]:offs[3]]
render(blob, 24,24,3, 32,8,6,msb=True ).save(f"{OUT}/zoom_sec2_24x24_msb.png")
render(blob, 24,24,3, 32,8,6,msb=False).save(f"{OUT}/zoom_sec2_24x24_lsb.png")
# 試 16 寬 (bpr=2) 高 24 — 萬一是 16x24
render(blob, 16,24,2, 32,8,6,msb=True ).save(f"{OUT}/zoom_sec2_16x24_msb.png")
print("done")
