#!/usr/bin/env python3
"""滑動起始位移:對 sec2,以不同 header 大小 (0..40) 取前 16 個 16x16 字模,
找出讓字模乾淨對齊的位移。並放大顯示。"""
import struct, os
from PIL import Image, ImageDraw

SRC="assets_raw/CHINA.FON"; OUT="assets_out/font_probe"
os.makedirs(OUT,exist_ok=True)
data=open(SRC,"rb").read()
first=struct.unpack_from("<I",data,0)[0]; n=first//4
offs=[struct.unpack_from("<I",data,i*4)[0] for i in range(n)]

def glyph_img(base):
    im=Image.new("L",(16,16),0); px=im.load()
    for yy in range(16):
        r=(data[base+yy*2]<<8)|data[base+yy*2+1]
        for xx in range(16):
            if (r>>(15-xx))&1: px[xx,yy]=255
    return im

o=offs[2]
scale=5; cols=16; ng=16
rows=len(range(0,41,2))
pad=3
sheet=Image.new("L",(40+cols*(16+pad)+pad, rows*(16+pad)+pad), 24)
px=sheet
d=ImageDraw.Draw(sheet)
ry=0
for hdr in range(0,41,2):
    gy=ry*(16+pad)+pad
    d.text((2,gy+4),str(hdr),fill=200)
    for g in range(ng):
        im=glyph_img(o+hdr+g*32)
        sheet.paste(im,(40+g*(16+pad)+pad, gy))
    ry+=1
sheet=sheet.resize((sheet.width*scale,sheet.height*scale),Image.NEAREST)
sheet.save(f"{OUT}/slide_sec2.png")
print("saved slide_sec2.png  (左欄=header 位移 bytes)")
