#!/usr/bin/env python3
"""FIRST.SCR 開場/標題畫面 render。
格式:640×350、4bpp、row-interleaved planar(每列 = plane0..3 各 80 byte = 320 B/row),
16 色取自 DQ3.PAL 前 16 色。輸出 docs/title/first_opening.png。
用法: tools/dockrun.sh tools/title_render.py
"""
import os, sys
sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image

d=open('assets_raw/FIRST.SCR','rb').read()
W,H=640,350; BPR=W//8                 # 80 byte/plane/row
pal=load_pal('assets_raw/DQ3.PAL')
im=Image.new('P',(W,H)); fl=[]
for c in pal[:16]: fl+=list(c)
while len(fl)<768: fl+=[0,0,0]
im.putpalette(fl); px=im.load()
for y in range(H):
    for xb in range(BPR):
        b=[d[y*4*BPR + p*BPR + xb] for p in range(4)]   # row-interleaved planes
        for bit in range(8):
            v=0
            for p in range(4): v|=((b[p]>>(7-bit))&1)<<p
            px[xb*8+bit,y]=v
os.makedirs('docs/title',exist_ok=True)
im.convert('RGB').save('docs/title/first_opening.png')
print('saved docs/title/first_opening.png', im.size)
