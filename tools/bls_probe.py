#!/usr/bin/env python3
"""DQ3MAN.BLS sprite 佈局探針(地表/角色行走 sprite 還原偵察)。

DQ3MAN.BLS = 222726 bytes = 6B header + 58 × 0xf00 page(.bls 分頁,見 loaders.c
load_blk_page:base=0x3c0*bp+6, 每頁 4×0x3c0=0xf00)。header = 04 00 18 00 e8 00。

試幾種解碼假設,render 成 sheet 供肉眼判斷哪種出角色行走幀。
用法(docker uv venv,cwd=/work):tools/dockrun.sh tools/bls_probe.py
"""
import struct, sys
sys.path.insert(0, 'tools')
from tile_lib import load_pal
from PIL import Image

BLS = open('assets_raw/DQ3MAN.BLS', 'rb').read()
hdr = BLS[:6]
print('header bytes:', hdr.hex(), '=>',
      'w_bytes=%d h=%d c=%d' % struct.unpack('<HHH', hdr))
print('filesize=%d  (6 + %d*0xf00 = %d)' % (len(BLS), (len(BLS)-6)//0xf00, 6+((len(BLS)-6)//0xf00)*0xf00))

pal_dq3 = load_pal('assets_raw/DQ3.PAL')
pal_mns = load_pal('assets_raw/MNSBK.PAL')

def planar_planemajor(buf, wbytes, h, plane_order):
    """plane-major: 依 plane_order 順序,每段 h*wbytes byte,MSB-first。回傳 [h][wbytes*8]。"""
    W = wbytes * 8
    px = [[0]*W for _ in range(h)]
    seg = wbytes * h
    for si, plane in enumerate(plane_order):
        base = si * seg
        for r in range(h):
            for b in range(wbytes):
                if base + r*wbytes + b >= len(buf):
                    continue
                byte = buf[base + r*wbytes + b]
                for bit in range(8):
                    if byte & (0x80 >> bit):
                        px[r][b*8+bit] |= (1 << plane)
    return px, W

def sheet(sprites, pal, tw, th, cols, path, scale=2):
    rows = (len(sprites)+cols-1)//cols
    img = Image.new('RGB', (cols*tw*scale, rows*th*scale), (40,40,40))
    pp = img.load()
    for i, px in enumerate(sprites):
        ox = (i % cols)*tw*scale; oy = (i//cols)*th*scale
        for r in range(min(th, len(px))):
            for c in range(min(tw, len(px[0]))):
                col = pal[px[r][c]] if px[r][c] < len(pal) else (255,0,255)
                for sy in range(scale):
                    for sx in range(scale):
                        pp[ox+c*scale+sx, oy+r*scale+sy] = col
    img.save(path)
    print('saved', path, img.size)

body = BLS[6:]

# 假設 A:body 連續 384B 32×24 tile,plane-major plane3→0(SHP 順序)
spritesA = []
for i in range(48):
    off = i*384
    if off+384 > len(body): break
    px,_ = planar_planemajor(body[off:off+384], 4, 24, [3,2,1,0])
    spritesA.append(px)
sheet(spritesA, pal_mns, 32, 24, 8, 'work/bls_A_384_p3210_mns.png')
sheet(spritesA, pal_dq3, 32, 24, 8, 'work/bls_A_384_p3210_dq3.png')

# 假設 B:body 連續 384B,plane-major plane0→3(BLK 順序)
spritesB = []
for i in range(48):
    off = i*384
    if off+384 > len(body): break
    px,_ = planar_planemajor(body[off:off+384], 4, 24, [0,1,2,3])
    spritesB.append(px)
sheet(spritesB, pal_mns, 32, 24, 8, 'work/bls_B_384_p0123_mns.png')

# 假設 C:每頁 0xf00,內含 4 個 0x3c0(960B)子塊;960=4plane×240=4×(wbytes*h)
#   試 wbytes=10,h=24(80×24)與 wbytes=12,h=20(96×20)
for (wb,hh,tag) in [(10,24,'80x24'),(12,20,'96x20'),(8,30,'64x30')]:
    spr=[]
    seg=wb*hh
    if seg*4!=960:
        # 仍嘗試:取每 0x3c0 塊
        pass
    for p in range(16):
        blk_off = p*0x3c0
        if blk_off+wb*hh*4>len(body): break
        px,_=planar_planemajor(body[blk_off:blk_off+wb*hh*4], wb, hh, [3,2,1,0])
        spr.append(px)
    sheet(spr, pal_mns, wb*8, hh, 4, 'work/bls_C_%s_mns.png'%tag)

print('done')
