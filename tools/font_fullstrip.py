#!/usr/bin/env python3
"""整個 section 攤成 16-wide 連續點陣長條(整段),橫向切成多欄並排以便一眼看完,
找『條紋(表格)→成字(字模)』轉折列。"""
import struct, os, sys
from PIL import Image, ImageDraw
data=open("assets_raw/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',data,i*4)[0] for i in range(22)]+[len(data)]
OUT="assets_out/font_probe"; os.makedirs(OUT,exist_ok=True)

si=int(sys.argv[1]) if len(sys.argv)>1 else 2
s=offs[si]; e=offs[si+1]; blob=data[s:e]
nrows=len(blob)//2
# 16 寬點陣
full=Image.new("L",(16,nrows),0); px=full.load()
for y in range(nrows):
    r=(blob[y*2]<<8)|blob[y*2+1]
    for x in range(16):
        if (r>>(15-x))&1: px[x,y]=255
# 切成每欄 256 列,並排
COLH=256; scale=3
ncols=(nrows+COLH-1)//COLH
pad=14
sheet=Image.new("L",(ncols*(16*scale+pad)+pad, COLH*scale+16),20)
d=ImageDraw.Draw(sheet)
for c in range(ncols):
    seg=full.crop((0,c*COLH,16,min((c+1)*COLH,nrows)))
    seg=seg.resize((16*scale,seg.height*scale),Image.NEAREST)
    x=pad+c*(16*scale+pad)
    sheet.paste(seg,(x,16))
    d.text((x,2),f"r{c*COLH}",fill=200)
sheet.save(f"{OUT}/fullstrip_sec{si:02d}.png")
print(f"sec{si}: {nrows} rows, {ncols} 欄並排 (每欄 256 列), byte/列=2")
