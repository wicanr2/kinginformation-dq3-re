#!/usr/bin/env python3
"""把所有 CTY*.DAT 的每個 section 渲成全圖,標 CTY/section 號 → docs/maps/cty/CTYxx.png。

正確 section 解析(docs/34):word[i] = section i base offset(0xffff=空,無 count 前綴);
每 section:layout_ptr = [so+0x0e],layout = so+ptr,w=[lay],h=[lay+2],tiles 在 lay+4
(u16/tile,低 byte = BLK index)。每 CTY 的 BLK 由 map_blknum(DGROUP 0x0a04)決定。

docker:
  docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
    bash -c 'uv venv -q /tmp/v && . /tmp/v/bin/activate && uv pip install -q pillow && python tools/render_all_cty.py'
"""
import os
import struct
import sys
sys.path.insert(0, 'tools')
from tile_lib import load_pal, read_blk, decode_tile
from PIL import Image, ImageDraw

EXE = open('assets_raw/DQ3.EXE', 'rb').read()
MAP_BLKNUM = [EXE[0x16140 + 0x0a04 + i] for i in range(94)]
pal = load_pal('assets_raw/DQ3.PAL')
flat = []
for c in pal[:256]:
    flat += list(c)
while len(flat) < 768:
    flat += [255, 0, 255]
TW, TH = 32, 24
blk_cache = {}


def get_blk(n):
    if n not in blk_cache:
        blk_cache[n] = read_blk('assets_raw/DQ3%d.BLK' % n)
    return blk_cache[n]


def render_section(cty, so, count, tiles):
    u16 = lambda o: struct.unpack_from('<H', cty, o)[0]
    if so + 0x10 > len(cty):
        return None
    lp = u16(so + 0x0e)
    lay = so + lp
    if lay + 4 > len(cty):
        return None
    w, h = u16(lay), u16(lay + 2)
    base = lay + 4
    if not (0 < w <= 64 and 0 < h <= 64):
        return None
    bad = sum(1 for k in range(w * h) if base + 2 * k + 1 < len(cty) and cty[base + 2 * k] >= count)
    if bad > 0.30 * w * h:
        return None
    img = Image.new('P', (w * TW, h * TH))
    img.putpalette(flat)
    px = img.load()
    for ty in range(h):
        for tx in range(w):
            o = base + 2 * (ty * w + tx)
            ti = cty[o] if o + 1 < len(cty) else 0
            if ti >= count:
                ti = 0
            td = decode_tile(tiles[ti])
            for r in range(TH):
                for c in range(TW):
                    px[tx * TW + c, ty * TH + r] = td[r][c]
    return img.convert('RGB')


def render_cty(idx):
    name = 'CTY%02d.DAT' % idx
    path = 'work/' + name
    if not os.path.exists(path):
        return None
    cty = open(path, 'rb').read()
    if len(cty) < 4:
        return None
    bn = MAP_BLKNUM[idx] if idx < len(MAP_BLKNUM) else 1
    count, tiles = get_blk(bn)
    u16 = lambda o: struct.unpack_from('<H', cty, o)[0]
    # 掃 section 偏移表 word[0..]:到第一個 section base 為止
    secs = []
    first = None
    for s in range(24):
        if 2 * s + 1 >= len(cty):
            break
        off = u16(2 * s)
        if first is not None and 2 * s >= first:
            break
        if off == 0xffff:
            secs.append(None)
            continue
        if off == 0 and s > 0:
            break
        if off < len(cty):
            if first is None or off < first:
                first = off
            secs.append(off)
        else:
            secs.append(None)
    # 渲染每個 section,縱向堆疊 + 標籤
    rendered = []
    for s, so in enumerate(secs):
        if so is None:
            continue
        im = render_section(cty, so, count, tiles)
        if im:
            rendered.append((s, im))
    if not rendered:
        return None
    scale = 0.5
    margin = 18
    cols = max(int(im.width * scale) for _, im in rendered)
    total_h = sum(int(im.height * scale) + margin for _, im in rendered) + 6
    sheet = Image.new('RGB', (max(cols, 200) + 10, total_h), (20, 20, 30))
    d = ImageDraw.Draw(sheet)
    y = 4
    for s, im in rendered:
        small = im.resize((max(1, int(im.width * scale)), max(1, int(im.height * scale))))
        d.text((4, y), 'CTY%d  sec%d  %dx%d  BLK%d' % (idx, s, im.width // TW, im.height // TH, bn),
                fill=(255, 255, 0))
        sheet.paste(small, (6, y + 14))
        y += small.height + margin
    return sheet


def main():
    os.makedirs('docs/maps/cty', exist_ok=True)
    done = 0
    for idx in range(94):
        sheet = render_cty(idx)
        if sheet:
            sheet.save('docs/maps/cty/CTY%02d.png' % idx)
            done += 1
    print('rendered %d CTY → docs/maps/cty/' % done)


if __name__ == '__main__':
    main()
