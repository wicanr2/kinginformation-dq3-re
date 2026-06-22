#!/usr/bin/env python3
"""探測 BLK tile 像素格式。每 tile 384 bytes,header w=4 h=24。
假設候選:
  A) 32x24 4-plane planar (EGA-style): 4 planes, each 24 rows x 4 bytes/row, bit-plane combine
  B) 32x24 linear 4bpp (nibbles): 384 bytes -> 32*24/2=384 ✓
  C) 16x24 8bpp linear: 16*24=384 ✓
  D) 32x24 planar but row-interleaved (4 bytes plane0,4 bytes plane1...)
渲染前幾個 tile 成 sheet,放大 4x,各假設一張圖。
"""
import struct, sys
from PIL import Image
from tile_lib import load_pal, read_blk

pal = load_pal('assets_raw/DQ3.PAL')
# 補滿 256 色避免 index 超界
while len(pal) < 256:
    pal.append((255, 0, 255))

def putpal(img):
    flat = []
    for c in pal[:256]:
        flat += list(c)
    img.putpalette(flat)

w, h, count, per, tiles = read_blk(sys.argv[1] if len(sys.argv) > 1 else 'assets_raw/DQ31.BLK')

N = min(64, count)
PW, PH = 4, 24  # bytes-per-row-per-plane, rows

def decode_A(tile):
    # 4 plane planar, planes contiguous: plane size = 4*24=96
    px = [[0]*32 for _ in range(24)]
    for p in range(4):
        base = p*96
        for row in range(24):
            for bx in range(4):
                byte = tile[base + row*4 + bx]
                for bit in range(8):
                    if byte & (0x80 >> bit):
                        px[row][bx*8+bit] |= (1 << p)
    return 32, 24, px

def decode_D(tile):
    # planar row-interleaved: per row 4 planes x 4 bytes =16 bytes; 24 rows=384
    px = [[0]*32 for _ in range(24)]
    for row in range(24):
        rb = row*16
        for p in range(4):
            for bx in range(4):
                byte = tile[rb + p*4 + bx]
                for bit in range(8):
                    if byte & (0x80 >> bit):
                        px[row][bx*8+bit] |= (1 << p)
    return 32, 24, px

def decode_B(tile):
    # linear 4bpp nibbles, 32 wide
    px = [[0]*32 for _ in range(24)]
    for row in range(24):
        for col in range(16):
            byte = tile[row*16 + col]
            px[row][col*2] = byte >> 4
            px[row][col*2+1] = byte & 0xF
    return 32, 24, px

def decode_C(tile):
    # 16x24 8bpp
    px = [[tile[row*16+col] for col in range(16)] for row in range(24)]
    return 16, 24, px

def make_sheet(decoder, tw, th):
    cols = 16
    rows = (N + cols - 1)//cols
    scale = 3
    sheet = Image.new('P', (cols*(tw+2)*scale, rows*(th+2)*scale), 0)
    putpal(sheet)
    for i in range(N):
        tw2, th2, px = decoder(tiles[i])
        ti = Image.new('P', (tw2, th2))
        putpal(ti)
        ti.putdata([px[r][c] for r in range(th2) for c in range(tw2)])
        ti = ti.resize((tw2*scale, th2*scale), Image.NEAREST)
        cx = (i % cols)*(tw+2)*scale
        cy = (i // cols)*(th+2)*scale
        sheet.paste(ti, (cx, cy))
    return sheet.convert('RGB')

for name, dec, tw, th in [('A',decode_A,32,24),('D',decode_D,32,24),('B',decode_B,32,24),('C',decode_C,16,24)]:
    sheet = make_sheet(dec, tw, th)
    out = f'docs/maps/probe_{name}.png'
    sheet.save(out)
    print('wrote', out)
