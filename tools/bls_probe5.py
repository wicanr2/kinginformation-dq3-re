#!/usr/bin/env python3
"""DQ3MAN.BLS 正確解碼:角色 blit 0x1ed0 反組譯 = 32×24 4-plane plane-major(384B 資料)
+ AND 遮罩(+0x180,384B),masked blit。frame stride=768。套遮罩→透明背景。
oracle 主角=金盔/髮+藍衣+膚色。試 DQ3.PAL/MNSBK 視窗 + 遮罩兩種極性。
"""
import sys
sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image

body=open('assets_raw/DQ3MAN.BLS','rb').read()[6:]
pal_dq3=load_pal('assets_raw/DQ3.PAL')
pal_mns=load_pal('assets_raw/MNSBK.PAL')

def planes(buf, base, wb=4, h=24, order=(3,2,1,0)):
    """回傳 [h][wb*8] 4-bit 值(plane-major,plane3 先)。"""
    W=wb*8; px=[[0]*W for _ in range(h)]; seg=wb*h
    for si,pl in enumerate(order):
        b0=base+si*seg
        for r in range(h):
            for b in range(wb):
                o=b0+r*wb+b
                if o<len(buf):
                    v=buf[o]
                    for bit in range(8):
                        if v&(0x80>>bit): px[r][b*8+bit]|=(1<<pl)
    return px

def mask_plane(buf, base, wb=4, h=24):
    """遮罩取第一個 plane(96B)的 bit:1bit/px。"""
    W=wb*8; m=[[0]*W for _ in range(h)]
    for r in range(h):
        for b in range(wb):
            o=base+r*wb+b
            if o<len(buf):
                v=buf[o]
                for bit in range(8):
                    m[r][b*8+bit]=1 if (v&(0x80>>bit)) else 0
    return m

def sheet(name, pal, win, stride, mask_transparent_bit, n=24, cols=8, scale=4):
    rows=(n+cols-1)//cols; tw,h=32,24
    img=Image.new('RGB',(cols*(tw*scale+4)+4, rows*(h*scale+4)+4),(0,128,0)); pp=img.load()
    for i in range(n):
        base=i*stride
        if base+0x300>len(body): break
        data=planes(body,base)
        msk=mask_plane(body,base+0x180)
        ox=4+(i%cols)*(tw*scale+4); oy=4+(i//cols)*(h*scale+4)
        for r in range(h):
            for c in range(tw):
                transparent = (msk[r][c]==mask_transparent_bit)
                if transparent:
                    col=(0,128,0)  # 透明=綠
                else:
                    v=data[r][c]; idx=win+v
                    col=pal[idx] if idx<len(pal) else (255,0,255)
                for sy in range(scale):
                    for sx in range(scale):
                        pp[ox+c*scale+sx, oy+r*scale+sy]=col
    img.save('work/%s.png'%name); print('saved work/%s.png'%name)

# stride 960(=222720/232 entries):資料@0、遮罩@0x180,遮罩 bit1=透明
sheet('bls_s960_dq3', pal_dq3, 0, 960, 1, n=32)
sheet('bls_s960_mns', pal_mns, 0, 960, 1, n=32)
sheet('bls_s960_dq3_m0', pal_dq3, 0, 960, 0, n=32)
