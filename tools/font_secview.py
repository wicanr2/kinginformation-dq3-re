#!/usr/bin/env python3
"""完整檢視單一 section 的所有 16x16 字模 (HDR=8),找出 table 區與字模區邊界。"""
import struct, os, sys
from PIL import Image

SRC="assets_raw/CHINA.FON"; OUT="assets_out/font_probe"
os.makedirs(OUT,exist_ok=True)
data=open(SRC,"rb").read()
first=struct.unpack_from("<I",data,0)[0]; n=first//4
offs=[struct.unpack_from("<I",data,i*4)[0] for i in range(n)]+[len(data)]
HDR=8

def render(si,cols=32,scale=3):
    o=offs[si]; size=offs[si+1]-o; ng=(size-HDR)//32
    rows=(ng+cols-1)//cols; pad=1
    img=Image.new("L",(cols*(16+pad)+pad,rows*(16+pad)+pad),48); px=img.load()
    for g in range(ng):
        gx=(g%cols)*(16+pad)+pad; gy=(g//cols)*(16+pad)+pad; base=o+HDR+g*32
        for yy in range(16):
            r=(data[base+yy*2]<<8)|data[base+yy*2+1]
            for xx in range(16):
                if (r>>(15-xx))&1: px[gx+xx,gy+yy]=255
    img.resize((img.width*scale,img.height*scale),Image.NEAREST).save(f"{OUT}/secview_{si:02d}.png")
    print(f"sec{si}: {ng} glyphs, {rows} rows x {cols} cols")

render(int(sys.argv[1]) if len(sys.argv)>1 else 0)
