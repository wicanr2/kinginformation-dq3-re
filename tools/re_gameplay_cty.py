#!/usr/bin/env python3
"""CTY 城鎮佈局解碼驗證(對照 re/render.c::cty_select_section / docs/11)。

依 DQ3.EXE sub_30cf 的程式路徑解 CTY 城鎮版面並貼成 PNG:
  檔首 u16 n + n×u16 section 偏移表;
  section_off + 0x0e → layout_ptr;版面起點 = section_off + layout_ptr;
  版面 = <u16 width><u16 height><spawn 4B> + width*height 個 u16(低 byte=BLK index)。

用法(docker uv venv 內,cwd=/work):
  tools/dockrun.sh tools/re_gameplay_cty.py [CTY 檔] [section] [BLK 圖庫]
  例:tools/dockrun.sh tools/re_gameplay_cty.py CTY00.DAT 0 DQ31.BLK
"""
import struct, sys
sys.path.insert(0, 'tools')
from tile_lib import load_pal, read_blk, decode_tile
from PIL import Image

cty_name = sys.argv[1] if len(sys.argv) > 1 else 'CTY00.DAT'
sect_idx = int(sys.argv[2]) if len(sys.argv) > 2 else 0
blk_name = sys.argv[3] if len(sys.argv) > 3 else 'DQ31.BLK'

cty = open('assets_raw/%s' % cty_name, 'rb').read()
u16 = lambda o: struct.unpack_from('<H', cty, o)[0]

n = u16(0)
offs = [u16(2 + 2 * i) for i in range(n)]
so = offs[sect_idx]
if so == 0xffff:
    sys.exit('section %d is empty (0xffff)' % sect_idx)

lp = u16(so + 0x0e)            # +0x0e: layout pointer (相對 section_off)
lay = lp + so                 # 版面資料起點
w = u16(lay)
h = u16(lay + 2)
base = lay + 4                # tile 陣列基底(spawn 4B 後)
print('n=%d sect%d @0x%x layout_ptr=0x%x w=%d h=%d tile_base=0x%x'
      % (n, sect_idx, so, lp, w, h, base))

pal = load_pal('assets_raw/DQ3.PAL')
count, tiles = read_blk('assets_raw/%s' % blk_name)
print('%s tiles=%d' % (blk_name, count))

TW, TH = 32, 24
img = Image.new('P', (w * TW, h * TH))
flat = []
for c in pal[:256]:
    flat += list(c)
while len(flat) < 768:
    flat += [255, 0, 255]
img.putpalette(flat)
px = img.load()
for ty in range(h):
    for tx in range(w):
        ti = cty[base + 2 * (ty * w + tx)]   # 低 byte = BLK tile index
        if ti >= count:
            ti = 0
        td = decode_tile(tiles[ti])
        for yy in range(TH):
            for xx in range(TW):
                px[tx * TW + xx, ty * TH + yy] = td[yy][xx]

out = 'work/%s_sect%d_verify.png' % (cty_name.split('.')[0].lower(), sect_idx)
img.convert('RGB').save(out)
print('saved %s  size %s' % (out, img.size))
