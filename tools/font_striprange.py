#!/usr/bin/env python3
"""指定 section 與 byte 範圍,render 連續 16-wide 點陣長條(附 byte 標尺),
看被混淆字的真實結構與邊界。"""
import struct,os,sys
from PIL import Image,ImageDraw
data=open("assets_raw/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',data,i*4)[0] for i in range(22)]+[len(data)]
OUT="assets_out/font_probe"
si=int(sys.argv[1]); b0=int(sys.argv[2]); b1=int(sys.argv[3])
s=offs[si]; blob=data[s:offs[si+1]]
nrows=(b1-b0)//2; scale=9
im=Image.new("L",(16,nrows),0); px=im.load()
for y in range(nrows):
    r=(blob[b0+y*2]<<8)|blob[b0+y*2+1]
    for x in range(16):
        if (r>>(15-x))&1: px[x,y]=255
big=im.resize((16*scale,nrows*scale),Image.NEAREST)
canvas=Image.new("L",(16*scale+70,nrows*scale),24)
canvas.paste(big,(70,0))
d=ImageDraw.Draw(canvas)
for r in range(0,nrows,2):
    y=r*scale; by=b0+r*2
    d.line([(66,y),(70,y)],fill=110)
    d.text((2,y),f"{by}",fill=170)
canvas.save(f"{OUT}/striprange_sec{si}_{b0}_{b1}.png")
print(f"sec{si} byte {b0}..{b1} -> striprange_sec{si}_{b0}_{b1}.png")
