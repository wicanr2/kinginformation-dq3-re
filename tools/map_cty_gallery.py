#!/usr/bin/env python3
"""CTY 城鎮 gallery:把一個 CTY 檔的所有非空 section 各自貼成圖,並排成一張。
格式見 re/render.c::cty_select_section / docs/11。
用法: tools/dockrun.sh tools/map_cty_gallery.py CTY00.DAT DQ31.BLK
"""
import struct, sys
sys.path.insert(0, 'tools')
from tile_lib import load_pal, read_blk, decode_tile
from PIL import Image, ImageDraw

cty_name = sys.argv[1] if len(sys.argv) > 1 else 'CTY00.DAT'
blk_name = sys.argv[2] if len(sys.argv) > 2 else 'DQ31.BLK'
cty = open('assets_raw/%s' % cty_name, 'rb').read()
u16 = lambda o: struct.unpack_from('<H', cty, o)[0]
n = u16(0)
offs = [u16(2 + 2*i) for i in range(n)]

pal = load_pal('assets_raw/DQ3.PAL')
count, tiles = read_blk('assets_raw/%s' % blk_name)
flat = []
for c in pal[:256]: flat += list(c)
while len(flat) < 768: flat += [255,0,255]
TW, TH = 32, 24
SCALE = 0.5  # 縮小以放得下整鎮

def render_section(so):
    lp = u16(so + 0x0e); lay = lp + so
    w = u16(lay); h = u16(lay+2); base = lay+4
    if not (0 < w <= 64 and 0 < h <= 64): return None
    if base + 2*w*h > len(cty): return None          # 版面超出檔案 → 非此結構
    bad = sum(1 for k in range(w*h) if cty[base+2*k] >= count)
    if bad > 0.30 * w * h: return None               # 多數 tile 越界 → BLK 不對/解析錯
    img = Image.new('P', (w*TW, h*TH)); img.putpalette(flat); px = img.load()
    for ty in range(h):
        for tx in range(w):
            off = base + 2*(ty*w+tx)
            if off+1 >= len(cty): continue
            ti = cty[off]
            if ti >= count: ti = 0
            td = decode_tile(tiles[ti])
            for yy in range(TH):
                for xx in range(TW):
                    px[tx*TW+xx, ty*TH+yy] = td[yy][xx]
    return img.convert('RGB')

secs = []
for i, so in enumerate(offs):
    if so == 0xffff: continue
    try:
        im = render_section(so)
    except Exception:
        im = None
    if im: secs.append((i, im))
print('%s: %d sections rendered' % (cty_name, len(secs)))

# 直向排列, 每段縮 50%, 加標籤
pad = 8
cols = []
for i, im in secs:
    sw, sh = int(im.width*SCALE), int(im.height*SCALE)
    cols.append((i, im.resize((sw, sh))))
maxw = max(im.width for _, im in cols) if cols else 100
toth = sum(im.height+pad+14 for _, im in cols) + pad
sheet = Image.new('RGB', (maxw+pad*2, toth), (24,24,24))
d = ImageDraw.Draw(sheet)
y = pad
for i, im in cols:
    d.text((pad, y), 'section %d' % i, fill=(220,220,220)); y += 14
    sheet.paste(im, (pad, y)); y += im.height + pad
out = 'work/cty_gallery_%s.png' % cty_name.split('.')[0].lower()
sheet.save(out)
print('saved', out, sheet.size)
