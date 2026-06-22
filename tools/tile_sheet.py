#!/usr/bin/env python3
"""把一個 BLK 的全部 tile 渲染成 tile sheet (含 index 編號用的格線)。
用法: tile_sheet.py <blk> <pal> <out.png> [scale]
"""
import sys
from PIL import Image
from tile_lib import load_pal, pal_flat, read_blk, decode_tile, TILE_W, TILE_H

blk = sys.argv[1]
palf = sys.argv[2]
out = sys.argv[3]
scale = int(sys.argv[4]) if len(sys.argv) > 4 else 2

pal = load_pal(palf)
flat = pal_flat(pal)
count, tiles = read_blk(blk)

cols = 16
rows = (count + cols - 1) // cols
gap = 2
cw, ch = (TILE_W + gap) * scale, (TILE_H + gap) * scale
sheet = Image.new('P', (cols * cw, rows * ch), 0)
sheet.putpalette(flat)
for i in range(count):
    px = decode_tile(tiles[i])
    ti = Image.new('P', (TILE_W, TILE_H))
    ti.putpalette(flat)
    ti.putdata([px[r][c] for r in range(TILE_H) for c in range(TILE_W)])
    ti = ti.resize((TILE_W * scale, TILE_H * scale), Image.NEAREST)
    sheet.paste(ti, ((i % cols) * cw, (i // cols) * ch))
sheet.convert('RGB').save(out)
print('wrote', out, 'tiles', count)
