#!/usr/bin/env python3
"""CHINA.FON 偵察:讀 offset table、把各 section 當點陣字模 render 成 PNG 網格。
目的是肉眼確認字模尺寸 (16x16 / 24x24) 與 bit 排列方向。"""
import struct, os, sys
from PIL import Image

SRC = "assets_raw/CHINA.FON"
OUT = "assets_out/font_probe"
os.makedirs(OUT, exist_ok=True)

data = open(SRC, "rb").read()
print(f"檔案大小: {len(data)}")

# 第一個 32-bit 值 = table 結束 offset → table 長度
first = struct.unpack_from("<I", data, 0)[0]
n = first // 4
offs = [struct.unpack_from("<I", data, i*4)[0] for i in range(n)]
offs_end = offs + [len(data)]
print(f"offset table: {n} entries (table 結束於 {first:#x})")
for i in range(n):
    size = offs_end[i+1] - offs[i]
    print(f"  sec[{i:2}] off={offs[i]:#08x} size={size:6d}  /72={size/72:.2f} /32={size/32:.2f} /18={size/18:.2f}")

def render_grid(blob, w, h, cols, path, max_glyphs=256):
    """把 blob 視為連續字模, 每字 w x h, MSB-first, 橫向 bytes_per_row = ceil(w/8)."""
    bpr = (w + 7) // 8
    glyph_bytes = bpr * h
    count = min(len(blob) // glyph_bytes, max_glyphs)
    if count == 0:
        return 0
    rows = (count + cols - 1) // cols
    pad = 2
    img = Image.new("L", (cols*(w+pad)+pad, rows*(h+pad)+pad), 64)
    px = img.load()
    for g in range(count):
        gx = (g % cols) * (w+pad) + pad
        gy = (g // cols) * (h+pad) + pad
        base = g * glyph_bytes
        for yy in range(h):
            for xx in range(w):
                byte = blob[base + yy*bpr + xx//8]
                bit = (byte >> (7 - (xx & 7))) & 1
                if bit:
                    px[gx+xx, gy+yy] = 255
    img.save(path)
    return count

# 對幾個代表性 section 試 24x24 與 16x16
for si in [0, 1, 2, 3, 10]:
    if si >= n:
        continue
    blob = data[offs[si]:offs_end[si+1]]
    for (w, h) in [(24, 24), (16, 16), (16, 15)]:
        c = render_grid(blob, w, h, 16, f"{OUT}/sec{si:02d}_{w}x{h}.png")
        print(f"  render sec{si:02d} {w}x{h}: {c} glyphs -> sec{si:02d}_{w}x{h}.png")
