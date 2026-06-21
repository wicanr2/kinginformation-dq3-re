#!/usr/bin/env python3
"""CHINA.FON 解碼:結構 = 22 sections,每 section = 8-byte header + N×(16x16 row-major MSB 字模)。
先驗證 (size-8)%32==0,再 render 各 section 字模網格。"""
import struct, os
from PIL import Image

SRC="assets_raw/CHINA.FON"; OUT="assets_out/font_probe"
os.makedirs(OUT,exist_ok=True)
data=open(SRC,"rb").read()
first=struct.unpack_from("<I",data,0)[0]; n=first//4
offs=[struct.unpack_from("<I",data,i*4)[0] for i in range(n)]+[len(data)]

HDR=8; GB=32  # 8-byte header, 16x16=32 bytes/glyph
print("sec | size | (size-8)%32 | glyphs | header(hex)")
total=0
for i in range(n):
    size=offs[i+1]-offs[i]
    rem=(size-HDR)%GB; ng=(size-HDR)//GB
    hdr=data[offs[i]:offs[i]+HDR].hex()
    total+=ng
    print(f"{i:3} | {size:6} | {rem:3} | {ng:5} | {hdr}")
print("總字數(推估):",total)

def render_sec(si,cols=24,scale=3):
    o=offs[si]; size=offs[si+1]-o
    ng=(size-HDR)//GB
    rows=(ng+cols-1)//cols; pad=1
    img=Image.new("L",(cols*(16+pad)+pad,rows*(16+pad)+pad),48); px=img.load()
    for g in range(ng):
        gx=(g%cols)*(16+pad)+pad; gy=(g//cols)*(16+pad)+pad
        base=o+HDR+g*GB
        for yy in range(16):
            r=(data[base+yy*2]<<8)|data[base+yy*2+1]
            for xx in range(16):
                if (r>>(15-xx))&1: px[gx+xx,gy+yy]=255
    img.resize((img.width*scale,img.height*scale),Image.NEAREST).save(f"{OUT}/china_sec{si:02d}.png")

for si in (0,2,5):
    render_sec(si)
print("rendered sec0/2/5")
