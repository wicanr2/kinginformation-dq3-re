#!/usr/bin/env python3
"""DQ3MAN.BLS 以 16 寬重解(對齊 oracle:地表主角為 ~16px 小 sprite)。
試 w_bytes=2(16px)、h=16/24,plane-major plane3-first;色0=透明(灰格)。
oracle 主角 = 金色頭盔/髮 + 藍衣 + 膚色;找出哪種尺寸+palette 跑出乾淨單隻。
"""
import sys
sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image

body = open('assets_raw/DQ3MAN.BLS','rb').read()[6:]
pal_dq3 = load_pal('assets_raw/DQ3.PAL')
pal_mns = load_pal('assets_raw/MNSBK.PAL')

def decode(buf, wb, h, order=(3,2,1,0)):
    W=wb*8; px=[[0]*W for _ in range(h)]; seg=wb*h
    for si,pl in enumerate(order):
        base=si*seg
        for r in range(h):
            for b in range(wb):
                o=base+r*wb+b
                if o>=len(buf): continue
                v=buf[o]
                for bit in range(8):
                    if v&(0x80>>bit): px[r][b*8+bit]|=(1<<pl)
    return px

def sheet(name, wb, h, pal, n=40, cols=10, scale=3):
    sz=wb*h*4
    rows=(n+cols-1)//cols; tw=wb*8
    img=Image.new('RGB',(cols*(tw*scale+4)+4, rows*(h*scale+4)+4),(60,60,60)); pp=img.load()
    for i in range(n):
        off=i*sz
        if off+sz>len(body): break
        px=decode(body[off:off+sz],wb,h)
        ox=4+(i%cols)*(tw*scale+4); oy=4+(i//cols)*(h*scale+4)
        for r in range(h):
            for c in range(tw):
                v=px[r][c]
                col=((110,110,110) if (r//2+c//2)%2 else (80,80,80)) if v==0 else (pal[v] if v<len(pal) else (255,0,255))
                for sy in range(scale):
                    for sx in range(scale):
                        pp[ox+c*scale+sx,oy+r*scale+sy]=col
    img.save('work/%s.png'%name); print('saved work/%s.png'%name, img.size)

sheet('bls16x24_dq3', 2, 24, pal_dq3)
sheet('bls16x16_dq3', 2, 16, pal_dq3)
sheet('bls16x24_mns', 2, 24, pal_mns)
