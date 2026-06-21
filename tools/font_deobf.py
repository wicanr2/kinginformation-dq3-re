#!/usr/bin/env python3
"""驗證去混淆 transform:左右半對調 + 左半下移1。
對 sec0 指定範圍,並排顯示 [原始 A] 與 [去混淆],找出哪些字屬於被混淆類。"""
import os, sys
from PIL import Image, ImageDraw
import importlib.util
spec=importlib.util.spec_from_file_location("g","tools/font_greedy.py")
g=importlib.util.module_from_spec(spec); spec.loader.exec_module(g)
OUT="assets_out/font_probe"; os.makedirs(OUT,exist_ok=True)
si=0; glyphs=g.extract_section(si); blob=g.data[g.offs[si]:g.offs[si+1]]
start=int(sys.argv[1]) if len(sys.argv)>1 else 38
end=int(sys.argv[2]) if len(sys.argv)>2 else 48

def orig(off): return [(blob[off+y*2]<<8)|blob[off+y*2+1] for y in range(16)]
def deobf(off):
    hi=[blob[off+y*2] for y in range(16)]; lo=[blob[off+y*2+1] for y in range(16)]
    # 左欄=低byte 下移1, 右欄=高byte
    out=[]
    for y in range(16):
        l = lo[y-1] if 0<=y-1<16 else 0
        out.append((l<<8)|hi[y])
    return out

def draw(rws,scale=6):
    im=Image.new("L",(16,16),0); px=im.load()
    for y in range(16):
        for x in range(16):
            if (rws[y]>>(15-x))&1: px[x,y]=255
    return im.resize((16*scale,16*scale),Image.NEAREST)

sub=glyphs[start:end]; N=len(sub); scale=6; pad=3; cellW=16*scale+pad
sheet=Image.new("L",(N*cellW+pad,2*(16*scale+pad)+20),24)
d=ImageDraw.Draw(sheet)
for c,(off,_) in enumerate(sub):
    sheet.paste(draw(orig(off)),(pad+c*cellW,20))
    sheet.paste(draw(deobf(off)),(pad+c*cellW,20+16*scale+pad))
    d.text((pad+c*cellW,2),str(start+c),fill=180)
d.text((1,20),"原",fill=255); d.text((1,20+16*scale+pad),"解",fill=255)
sheet.save(f"{OUT}/deobf_{start}_{end}.png")
print(f"-> deobf_{start}_{end}.png (上=原始, 下=去混淆)")
