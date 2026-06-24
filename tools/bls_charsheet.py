#!/usr/bin/env python3
"""DQ3MAN.BLS 全角色 sprite 對照表(docs/sprites/dq3man_charsheet.png)。

格式(docs/27):6B 表頭 + 232 entry × 960B;每 entry = 32×24,4-bit plane-major(plane3→0)
+ AND 遮罩 @+0x180(bit=1 透明)。每角色 = 4 連續 entry(下/左/上/右 frame)。
色盤 DQ3.PAL 低 16 色。

在 docker 內跑(不污染 host):
  docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
    bash -c 'uv venv -q /tmp/v && . /tmp/v/bin/activate && uv pip install -q pillow && python tools/bls_charsheet.py'
"""
import sys
import os
sys.path.insert(0, 'tools')
from tile_lib import load_pal
from PIL import Image, ImageDraw

BLS = sys.argv[1] if len(sys.argv) > 1 else 'DQ3MAN.BLS'   # DQ3MAN(NPC)/ DQ3MST(職業)
body = open('assets_raw/' + BLS, 'rb').read()[6:]
pal = load_pal('assets_raw/DQ3.PAL')
NENTRY = len(body) // 960
BG = (0, 100, 0)


def decode(e):
    base = e * 960
    px = [[0] * 32 for _ in range(24)]
    for s, pl in enumerate((3, 2, 1, 0)):
        b0 = base + s * 96
        for r in range(24):
            for b in range(4):
                v = body[b0 + r * 4 + b]
                for bit in range(8):
                    if v & (0x80 >> bit):
                        px[r][b * 8 + bit] |= (1 << pl)
    msk = [[0] * 32 for _ in range(24)]
    mb = base + 0x180
    for r in range(24):
        for b in range(4):
            v = body[mb + r * 4 + b]
            for bit in range(8):
                msk[r][b * 8 + bit] = (v >> (7 - bit)) & 1
    return px, msk


nchar = NENTRY // 4
sc = 2
fw, fh = 32 * sc, 24 * sc
# 每列一個角色:char 標籤 + 4 方向 frame
char_per_col = 2                      # 一橫排放 2 個角色(各 4 frame)
rows = (nchar + char_per_col - 1) // char_per_col
col_w = 70 + 4 * (fw + 4)
row_h = fh + 8
img = Image.new('RGB', (col_w * char_per_col, row_h * rows + 20), (0, 0, 0))
pp = img.load()
d = ImageDraw.Draw(img)
d.text((6, 4), BLS + ' 角色 sprite(每角色 4 frame:下/左/上/右;entry=char*4)', fill=(255, 255, 0))
for ci in range(nchar):
    gx = (ci % char_per_col) * col_w
    gy = 20 + (ci // char_per_col) * row_h
    d.text((gx + 4, gy + fh // 2 - 6), 'c%d\ne%d' % (ci, ci * 4), fill=(255, 255, 0))
    for f in range(4):
        e = ci * 4 + f
        if e >= NENTRY:
            break
        px, msk = decode(e)
        ox = gx + 66 + f * (fw + 4)
        for r in range(24):
            for c in range(32):
                col = BG if msk[r][c] == 1 else pal[px[r][c]]
                for sy in range(sc):
                    for sx in range(sc):
                        pp[ox + c * sc + sx, gy + 4 + r * sc + sy] = col
os.makedirs('docs/sprites', exist_ok=True)
outp = 'docs/sprites/%s_charsheet.png' % BLS.split('.')[0].lower()
img.save(outp)
print('saved %s  (%d entries / %d chars)' % (outp, NENTRY, nchar))
