#!/usr/bin/env python3
"""對 sec0 指定字模索引範圍,並排多種重組變體,找出讓被混淆字(矮/愛/ㄞ)現形者。
變體:
 A 原始 row-interleave (byte0<<8|byte1)
 B 每列兩 byte 互換 (左右 8px 對調)
 C 左塊+右塊: row = left_block[y]<<8 | right_block[y]   (前16 byte=左欄)
 D 同 C 但左右互換
 E 每列 byte 互換 + 反轉 (nibble?) 先放著
"""
import os, sys
from PIL import Image, ImageDraw
import importlib.util
spec=importlib.util.spec_from_file_location("g","tools/font_greedy.py")
g=importlib.util.module_from_spec(spec); spec.loader.exec_module(g)
OUT="assets_out/font_probe"; os.makedirs(OUT,exist_ok=True)
si=0
glyphs=g.extract_section(si)
blob=g.data[g.offs[si]:g.offs[si+1]]
start=int(sys.argv[1]) if len(sys.argv)>1 else 38
end=int(sys.argv[2]) if len(sys.argv)>2 else 48

def variants(off):
    raw=[blob[off+i] for i in range(32)]
    A=[(raw[y*2]<<8)|raw[y*2+1] for y in range(16)]
    B=[(raw[y*2+1]<<8)|raw[y*2] for y in range(16)]
    C=[(raw[y]<<8)|raw[16+y] for y in range(16)]
    D=[(raw[16+y]<<8)|raw[y] for y in range(16)]
    return {"A":A,"B":B,"C":C,"D":D}

labels=["A","B","C","D"]; scale=5; pad=3
sub=glyphs[start:end]; N=len(sub)
cellW=16*scale+pad
sheet=Image.new("L",(N*cellW+pad, len(labels)*(16*scale+pad)+20),24)
d=ImageDraw.Draw(sheet)
def draw(rows):
    im=Image.new("L",(16,16),0); px=im.load()
    for y in range(16):
        for x in range(16):
            if (rows[y]>>(15-x))&1: px[x,y]=255
    return im.resize((16*scale,16*scale),Image.NEAREST)
for col,(off,_) in enumerate(sub):
    v=variants(off)
    for r,lab in enumerate(labels):
        sheet.paste(draw(v[lab]),(pad+col*cellW, 20+r*(16*scale+pad)))
    d.text((pad+col*cellW,2),str(start+col),fill=180)
for r,lab in enumerate(labels):
    d.text((1,20+r*(16*scale+pad)),lab,fill=255)
sheet.save(f"{OUT}/recombine_sec0_{start}_{end}.png")
print(f"sec0 glyph {start}..{end} -> recombine_sec0_{start}_{end}.png")
