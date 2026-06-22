#!/usr/bin/env python3
"""DQ3MAN.BLS 角色 sprite 細看:前 N 隻,色0=透明(灰格),DQ3.PAL,放大。
找出地表主角(下/左/上/右 4 方向幀)。
"""
import struct, sys
sys.path.insert(0, 'tools')
from tile_lib import load_pal
from PIL import Image

BLS = open('assets_raw/DQ3MAN.BLS', 'rb').read()
body = BLS[6:]
pal = load_pal('assets_raw/DQ3.PAL')

def decode(buf, wbytes=4, h=24, order=(3,2,1,0)):
    W = wbytes*8
    px=[[0]*W for _ in range(h)]
    seg=wbytes*h
    for si,plane in enumerate(order):
        base=si*seg
        for r in range(h):
            for b in range(wbytes):
                if base+r*wbytes+b>=len(buf): continue
                byte=buf[base+r*wbytes+b]
                for bit in range(8):
                    if byte&(0x80>>bit):
                        px[r][b*8+bit]|=(1<<plane)
    return px

N=32
cols=8; rows=(N+cols-1)//cols
tw,th=32,24; scale=3; pad=6
img=Image.new('RGB',(cols*(tw*scale+pad)+pad, rows*(th*scale+pad)+pad),(60,60,60))
pp=img.load()
for i in range(N):
    off=i*384
    if off+384>len(body): break
    px=decode(body[off:off+384])
    ox=pad+(i%cols)*(tw*scale+pad); oy=pad+(i//cols)*(th*scale+pad)
    for r in range(th):
        for c in range(tw):
            v=px[r][c]
            if v==0:
                col=(120,120,120) if (r//3+c//3)%2 else (90,90,90)  # 透明:灰格
            else:
                col=pal[v] if v<len(pal) else (255,0,255)
            for sy in range(scale):
                for sx in range(scale):
                    pp[ox+c*scale+sx, oy+r*scale+sy]=col
img.save('work/bls_chars_dq3.png')
print('saved work/bls_chars_dq3.png', img.size)
