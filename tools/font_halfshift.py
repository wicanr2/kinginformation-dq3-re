#!/usr/bin/env python3
"""機制 = 左右半對調 + 一半垂直微移。
以對調為基底:left 8 欄 = 低 byte(odd), right 8 欄 = 高 byte(even);
再對 right 半或 left 半上下位移 dy,找出拼回正字者。"""
import os, sys
from PIL import Image, ImageDraw
import importlib.util
spec=importlib.util.spec_from_file_location("g","tools/font_greedy.py")
g=importlib.util.module_from_spec(spec); spec.loader.exec_module(g)
OUT="assets_out/font_probe"; os.makedirs(OUT,exist_ok=True)
si=0; glyphs=g.extract_section(si); blob=g.data[g.offs[si]:g.offs[si+1]]
idx=int(sys.argv[1]) if len(sys.argv)>1 else 43
off=glyphs[idx][0]
hi=[blob[off+y*2]   for y in range(16)]   # 高 byte
lo=[blob[off+y*2+1] for y in range(16)]   # 低 byte
# 對調基底:左欄=低byte, 右欄=高byte
left, right = lo, hi

def compose(dyR=0,dyL=0):
    out=[]
    for y in range(16):
        l = left[y-dyL] if 0<=y-dyL<16 else 0
        r = right[y-dyR] if 0<=y-dyR<16 else 0
        out.append((l<<8)|r)
    return out

def draw(rws,scale=7):
    im=Image.new("L",(16,16),0); px=im.load()
    for y in range(16):
        for x in range(16):
            if (rws[y]>>(15-x))&1: px[x,y]=255
    return im.resize((16*scale,16*scale),Image.NEAREST)

shifts=list(range(-4,5)); scale=7; pad=3; cellW=16*scale+pad
sheet=Image.new("L",(len(shifts)*cellW+pad, 2*(16*scale+pad)+24),24)
d=ImageDraw.Draw(sheet)
for c,dy in enumerate(shifts):
    sheet.paste(draw(compose(dyR=dy)),(pad+c*cellW,20))
    sheet.paste(draw(compose(dyL=dy)),(pad+c*cellW,20+16*scale+pad))
    d.text((pad+c*cellW,2),f"{dy:+d}",fill=200)
d.text((1,20),"R移",fill=255); d.text((1,20+16*scale+pad),"L移",fill=255)
sheet.save(f"{OUT}/halfshiftB_g{idx}.png")
print(f"glyph{idx} (對調基底) -> halfshiftB_g{idx}.png 上排=右半上下移, 下排=左半上下移")
