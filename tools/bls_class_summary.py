#!/usr/bin/env python3
"""DQ3MST.BLS 職業 sprite 摘要圖(docs/sprites/dq3_class_sprites.png)。

對映(實機/使用者確認):char = 職業×2 + 性別,entry_base = char×4。
8 職業 × 2 性別 = c0..c15。PIL 預設字型不支援中文 → 標籤用英文 + 性別 + c/e 號。

docker 內跑:
  docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
    bash -c 'uv venv -q /tmp/v && . /tmp/v/bin/activate && uv pip install -q pillow && python tools/bls_class_summary.py'
"""
import os
import sys
sys.path.insert(0, 'tools')
from tile_lib import load_pal
from PIL import Image, ImageDraw

body = open('assets_raw/DQ3MST.BLS', 'rb').read()[6:]
pal = load_pal('assets_raw/DQ3.PAL')
CLASSES = ['Hero(勇者)', 'Warrior(戰士)', 'Fighter(武鬥家)', 'Priest(僧侶)',
           'Mage(魔法使者)', 'Sage(賢者)', 'Merchant(商人)', 'Gambler(遊玩者)']
# PIL 預設字型只畫得出 ASCII,實際輸出取括號前英文
CLASSES = [c.split('(')[0] for c in CLASSES]


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


sc = 3
fw, fh = 32 * sc, 24 * sc
BG = (30, 30, 40)
cell_w = fw + 10
cell_h = fh + 30
img = Image.new('RGB', (cell_w * 8 + 10, cell_h * 2 + 30), BG)
pp = img.load()
d = ImageDraw.Draw(img)
d.text((8, 6), 'DQ3MST.BLS class sprites — entry = (class*2 + gender)*4   [front frame]', fill=(255, 255, 0))
for cls in range(8):
    d.text((10 + cls * cell_w, 24), CLASSES[cls], fill=(120, 220, 255))
    for g in range(2):                        # 0=M(top) 1=F(bottom)
        char = cls * 2 + g
        e = char * 4
        px, msk = decode(e)
        gx = 10 + cls * cell_w
        gy = 38 + g * cell_h
        d.text((gx, gy), 'c%d e%d %s' % (char, e, 'M' if g == 0 else 'F'), fill=(255, 255, 0))
        for r in range(24):
            for c in range(32):
                col = BG if msk[r][c] == 1 else pal[px[r][c]]
                for sy in range(sc):
                    for sx in range(sc):
                        pp[gx + c * sc + sx, gy + 14 + r * sc + sy] = col
os.makedirs('docs/sprites', exist_ok=True)
img.save('docs/sprites/dq3_class_sprites.png')
print('saved docs/sprites/dq3_class_sprites.png')
