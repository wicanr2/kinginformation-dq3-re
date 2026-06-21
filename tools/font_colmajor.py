#!/usr/bin/env python3
"""測試 column-major(直行)字模排列。
24x24 col-major: 每字 24 直行,每行 3 bytes(24 bits, MSB=最上),共 72 bytes。
同時試 16x16 col-major(每行 2 bytes,32 bytes)。輸出前 64 字網格對照。"""
import struct, os
from PIL import Image

SRC="assets_raw/CHINA.FON"; OUT="assets_out/font_probe"
os.makedirs(OUT,exist_ok=True)
data=open(SRC,"rb").read()
first=struct.unpack_from("<I",data,0)[0]; n=first//4
offs=[struct.unpack_from("<I",data,i*4)[0] for i in range(n)]+[len(data)]

def grid_colmajor(blob, w, h, n_glyphs, cols, scale, start=0):
    bpc=(h+7)//8            # bytes per column
    gb=bpc*w                # glyph bytes
    n_glyphs=min(n_glyphs,(len(blob)-start)//gb)
    rows=(n_glyphs+cols-1)//cols; pad=2
    img=Image.new("L",(cols*(w+pad)+pad,rows*(h+pad)+pad),48); px=img.load()
    for g in range(n_glyphs):
        gx=(g%cols)*(w+pad)+pad; gy=(g//cols)*(h+pad)+pad
        base=start+g*gb
        for xx in range(w):
            for yy in range(h):
                b=blob[base+xx*bpc+yy//8]
                if (b>>(7-(yy&7)))&1: px[gx+xx,gy+yy]=255
    return img.resize((img.width*scale,img.height*scale),Image.NEAREST)

blob=data[offs[2]:offs[3]]
grid_colmajor(blob,24,24,64,8,5).save(f"{OUT}/col_sec2_24x24.png")
grid_colmajor(blob,16,16,64,8,5).save(f"{OUT}/col_sec2_16x16.png")
# 也試帶一點 header skip(若 section 前有 header)
grid_colmajor(blob,24,24,64,8,5,start=0).save(f"{OUT}/col_sec2_24x24_s0.png")
print("done")
