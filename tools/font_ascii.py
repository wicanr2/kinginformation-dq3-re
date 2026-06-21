#!/usr/bin/env python3
"""解 ASCII 字型 D3TXT00.FON / SETTXT.FON 當羅塞塔石碑:
已知拉丁字母長相,用來鎖定 bit 排列 (寬度/row-or-col/MSB)。"""
import os
from PIL import Image

OUT="assets_out/font_probe"; os.makedirs(OUT,exist_ok=True)

def grid_rowmajor(blob,w,h,n_glyphs,cols,scale,start=0,msb=True):
    bpr=(w+7)//8; gb=bpr*h
    n_glyphs=min(n_glyphs,(len(blob)-start)//gb)
    rows=(n_glyphs+cols-1)//cols; pad=2
    img=Image.new("L",(cols*(w+pad)+pad,rows*(h+pad)+pad),48); px=img.load()
    for g in range(n_glyphs):
        gx=(g%cols)*(w+pad)+pad; gy=(g//cols)*(h+pad)+pad; base=start+g*gb
        for yy in range(h):
            for xx in range(w):
                b=blob[base+yy*bpr+xx//8]
                bit=(b>>(7-(xx&7)))&1 if msb else (b>>(xx&7))&1
                if bit: px[gx+xx,gy+yy]=255
    return img.resize((img.width*scale,img.height*scale),Image.NEAREST)

for name in ("D3TXT00.FON","SETTXT.FON"):
    blob=open(f"assets_raw/{name}","rb").read()
    print(name, len(blob), "bytes  /32=",len(blob)/32," /16=",len(blob)/16)
    # 試 16x16 與 8x16
    grid_rowmajor(blob,16,16,128,16,4).save(f"{OUT}/ascii_{name}_16x16.png")
    grid_rowmajor(blob,8,16,128,16,4).save(f"{OUT}/ascii_{name}_8x16.png")
print("done")
