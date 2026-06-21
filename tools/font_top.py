#!/usr/bin/env python3
import struct,os,sys
from PIL import Image,ImageDraw
data=open("assets_raw/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',data,i*4)[0] for i in range(22)]+[len(data)]
OUT="assets_out/font_probe"
si=int(sys.argv[1]) if len(sys.argv)>1 else 2
s=offs[si]; blob=data[s:offs[si+1]]
ROWS=192; scale=7
im=Image.new("L",(16,ROWS),0); px=im.load()
for y in range(ROWS):
    r=(blob[y*2]<<8)|blob[y*2+1]
    for x in range(16):
        if (r>>(15-x))&1: px[x,y]=255
big=im.resize((16*scale,ROWS*scale),Image.NEAREST)
canvas=Image.new("L",(16*scale+50,ROWS*scale),24)
canvas.paste(big,(50,0))
d=ImageDraw.Draw(canvas)
for r in range(0,ROWS,8):
    y=r*scale
    d.line([(46,y),(50,y)],fill=120)
    d.text((2,y),f"{r}/{r*2}B",fill=170)
canvas.save(f"{OUT}/top_sec{si:02d}.png")
print(f"sec{si} 前{ROWS}列  (列號/byte)")
