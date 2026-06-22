#!/usr/bin/env python3
"""用 MAP 的 tile index 陣列 + BLK tile 圖庫拼出地圖。
用法: map_render.py <map> <blk> <pal> <out.png> [scale]
"""
import sys
from PIL import Image
from tile_lib import load_pal, pal_flat, read_blk, decode_tile, read_map, TILE_W, TILE_H

mapf = sys.argv[1]
blk = sys.argv[2]
palf = sys.argv[3]
out = sys.argv[4]
scale = float(sys.argv[5]) if len(sys.argv) > 5 else 1.0

pal = load_pal(palf)
flat = pal_flat(pal)
count, tiles = read_blk(blk)
W, H, data = read_map(mapf)
print('map', W, 'x', H, 'tiles in blk', count, 'max idx', max(data) if data else 0)

# 預先 decode 全部 tile 成 Image
timgs = []
for i in range(count):
    px = decode_tile(tiles[i])
    ti = Image.new('P', (TILE_W, TILE_H))
    ti.putpalette(flat)
    ti.putdata([px[r][c] for r in range(TILE_H) for c in range(TILE_W)])
    timgs.append(ti)

canvas = Image.new('P', (W * TILE_W, H * TILE_H), 0)
canvas.putpalette(flat)
for y in range(H):
    row = data[y * W:(y + 1) * W]
    for x, idx in enumerate(row):
        if idx < count:
            canvas.paste(timgs[idx], (x * TILE_W, y * TILE_H))
canvas = canvas.convert('RGB')
if scale != 1.0:
    canvas = canvas.resize((int(W * TILE_W * scale), int(H * TILE_H * scale)), Image.NEAREST)
canvas.save(out)
print('wrote', out, canvas.size)
