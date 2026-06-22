#!/usr/bin/env python3
"""共用:DQ3 地圖素材解碼基礎。

PAL: 80 色,每色 3 bytes RGB,值域 0-63 (VGA 6-bit DAC)。
BLK tile: header <HHH = (planar_bytes_per_row=4, height=24, count);
          body = count 個 tile,每 tile 384 bytes。
          格式 = 32x24 像素,4-bit 平面 (planar),4 個 plane 連續排列,
          每 plane 96 bytes (24 row x 4 byte/row),bit MSB-first。
MAP: header <HH = (width, height);body = width*height 個 byte,每 byte = tile index。
"""
import struct

TILE_W, TILE_H = 32, 24
PLANE_SZ = 96  # 24 row * 4 byte
TILE_SZ = 384

def load_pal(path):
    d = open(path, 'rb').read()
    cols = []
    for i in range(0, len(d) - 2, 3):
        r, g, b = d[i], d[i+1], d[i+2]
        cols.append(((r << 2) | (r >> 4), (g << 2) | (g >> 4), (b << 2) | (b >> 4)))
    return cols

def pal_flat(pal, fill=(255, 0, 255)):
    p = list(pal)
    while len(p) < 256:
        p.append(fill)
    flat = []
    for c in p[:256]:
        flat += list(c)
    return flat

def read_blk(path):
    d = open(path, 'rb').read()
    w, h, count = struct.unpack('<HHH', d[:6])
    body = d[6:]
    per = len(body) // count if count else 0
    tiles = [body[i*per:(i+1)*per] for i in range(count)]
    return count, tiles

def decode_tile(tile):
    """4-plane planar -> 32x24 list[list[idx]]"""
    px = [[0] * TILE_W for _ in range(TILE_H)]
    for p in range(4):
        base = p * PLANE_SZ
        for row in range(TILE_H):
            ro = base + row * 4
            for bx in range(4):
                byte = tile[ro + bx]
                if byte:
                    col0 = bx * 8
                    for bit in range(8):
                        if byte & (0x80 >> bit):
                            px[row][col0 + bit] |= (1 << p)
    return px

def read_map(path):
    d = open(path, 'rb').read()
    w, h = struct.unpack('<HH', d[:4])
    return w, h, d[4:4 + w * h]
