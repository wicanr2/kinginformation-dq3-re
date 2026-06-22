#!/usr/bin/env python3
"""DQ3MAN.BLS 32×24 固定,掃 plane 順序 × MNSBK palette 16色視窗,找乾淨主角配色。
oracle 主角 = 金盔/髮 + 藍衣 + 膚色臉。形狀已對(32×24 plane-major),只差色彩對應。
"""
import sys, itertools
sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image

body = open('assets_raw/DQ3MAN.BLS','rb').read()[6:]
pal_mns = load_pal('assets_raw/MNSBK.PAL')   # 160 色
pal_dq3 = load_pal('assets_raw/DQ3.PAL')

def decode(buf, order, wb=4, h=24):
    W=wb*8; px=[[0]*W for _ in range(h)]; seg=wb*h
    for si,pl in enumerate(order):
        base=si*seg
        for r in range(h):
            for b in range(wb):
                o=base+r*wb+b
                if o<len(buf):
                    v=buf[o]
                    for bit in range(8):
                        if v&(0x80>>bit): px[r][b*8+bit]|=(1<<pl)
    return px

def render_row(sprites, pal, win, tw=32, h=24, scale=3):
    cols=len(sprites)
    img=Image.new('RGB',(cols*(tw*scale+4)+4, h*scale+8),(50,50,50)); pp=img.load()
    for i,px in enumerate(sprites):
        ox=4+i*(tw*scale+4)
        for r in range(h):
            for c in range(tw):
                v=px[r][c]
                if v==0: col=(100,100,100)
                else:
                    idx=win+v
                    col=pal[idx] if idx<len(pal) else (255,0,255)
                for sy in range(scale):
                    for sx in range(scale):
                        pp[ox+c*scale+sx, r*scale+sy+4]=col
    return img

# 取前 8 隻 sprite
N=8
# 掃 plane 順序(主要兩種)× MNSBK 視窗 0,16,..,144
orders={'p3210':(3,2,1,0),'p0123':(0,1,2,3)}
out=Image.new('RGB',(8*(32*3+4)+4, 0))
panels=[]
for oname,order in orders.items():
    sprites=[decode(body[i*384:i*384+384],order) for i in range(N)]
    for win in range(0,160,16):
        img=render_row(sprites, pal_mns, win)
        panels.append((f'{oname}_mns_win{win}', img))
    # 也試 dq3 pal 視窗 0/16/32/48/64
    for win in (0,16,32,48,64):
        img=render_row(sprites, pal_dq3, win)
        panels.append((f'{oname}_dq3_win{win}', img))

# 疊成一張長圖,左側標籤
from PIL import ImageDraw
lblw=150
totalh=sum(p[1].height+18 for p in panels)
W=panels[0][1].width
big=Image.new('RGB',(W+lblw, totalh),(20,20,20)); d=ImageDraw.Draw(big)
y=0
for name,img in panels:
    d.text((4,y+4), name, fill=(230,230,230))
    big.paste(img,(lblw,y))
    y+=img.height+18
big.save('work/bls_palsweep.png'); print('saved work/bls_palsweep.png', big.size)
