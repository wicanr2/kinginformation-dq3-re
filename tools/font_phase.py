#!/usr/bin/env python3
"""相位測試:sec2 某段以 16-wide 連續 render,疊兩種 16 列網格 (phase 0 / phase 8),
看哪個相位把整字框住 (阿 不被切半)。"""
import struct,os,sys
from PIL import Image,ImageDraw
data=open("assets_raw/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',data,i*4)[0] for i in range(22)]+[len(data)]
OUT="assets_out/font_probe"
si=int(sys.argv[1]) if len(sys.argv)>1 else 2
r0=int(sys.argv[2]) if len(sys.argv)>2 else 96
nrows=160
s=offs[si]; blob=data[s:offs[si+1]]
scale=6
def strip(phase):
    im=Image.new("L",(16,nrows),0); px=im.load()
    for y in range(nrows):
        rr=r0+y
        r=(blob[rr*2]<<8)|blob[rr*2+1]
        for x in range(16):
            if (r>>(15-x))&1: px[x,y]=255
    big=im.resize((16*scale,nrows*scale),Image.NEAREST).convert("RGB")
    d=ImageDraw.Draw(big)
    # 每 16 列畫一條紅線, 起始相位 = phase
    for gy in range((-phase)%16, nrows, 16):
        d.line([(0,gy*scale),(16*scale,gy*scale)],fill=(255,60,60))
    return big
a=strip(0); b=strip(8)
canvas=Image.new("RGB",(a.width+b.width+60,a.height+24),(20,20,20))
canvas.paste(a,(40,20)); canvas.paste(b,(a.width+60,20))
d=ImageDraw.Draw(canvas)
d.text((40,4),"phase 0",fill=(255,255,255)); d.text((a.width+60,4),"phase 8",fill=(255,255,255))
for r in range(0,nrows,16):
    d.text((2,20+r*scale),str(r0+r),fill=(180,180,180))
canvas.save(f"{OUT}/phase_sec{si:02d}_r{r0}.png")
print(f"sec{si} rows {r0}..{r0+nrows}  -> phase_sec{si:02d}_r{r0}.png")
