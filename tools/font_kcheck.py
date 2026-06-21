#!/usr/bin/env python3
"""測 sec2 字模起點 = 8+32*k,k=13..17,各 render 前 24 字,找乾淨對齊的 k。"""
import struct,os
from PIL import Image,ImageDraw
data=open("assets_raw/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',data,i*4)[0] for i in range(22)]+[len(data)]
OUT="assets_out/font_probe"
s=offs[2]; blob=data[s:offs[3]]
def glyph(base):
    im=Image.new("L",(16,16),0); px=im.load()
    for y in range(16):
        r=(blob[base+y*2]<<8)|blob[base+y*2+1]
        for x in range(16):
            if (r>>(15-x))&1: px[x,y]=255
    return im
scale=5; cols=24; pad=2
ks=list(range(13,18))
sheet=Image.new("L",(60+cols*(16+pad), len(ks)*(16+pad)+pad),24)
d=ImageDraw.Draw(sheet)
for ri,k in enumerate(ks):
    H=8+32*k
    gy=ri*(16+pad)+pad
    d.text((2,gy+4),f"k={k} H={H}",fill=200)
    for g in range(cols):
        base=H+g*32
        if base+32>len(blob): break
        sheet.paste(glyph(base),(60+g*(16+pad),gy))
sheet=sheet.resize((sheet.width*scale,sheet.height*scale),Image.NEAREST)
sheet.save(f"{OUT}/kcheck_sec2.png")
print("saved kcheck_sec2.png")
