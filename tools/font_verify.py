#!/usr/bin/env python3
import struct,sys
from PIL import Image
data=open("assets_raw/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',data,i*4)[0] for i in range(22)]+[len(data)]
# (start_byte) per section from autoscan
START={0:358,1:384,2:392,3:374,4:362,5:390,6:386,7:362,8:384,9:368,10:342,
       11:352,12:418,13:410,14:448,15:342,16:364,17:360,18:340,19:366,20:368,21:392}
def render(si,cols=24,scale=4):
    blob=data[offs[si]:offs[si+1]]; S=START[si]; ng=(len(blob)-S)//32
    rows=(ng+cols-1)//cols; pad=1
    img=Image.new("L",(cols*(16+pad)+pad,rows*(16+pad)+pad),48); px=img.load()
    for g in range(ng):
        gx=(g%cols)*(16+pad)+pad; gy=(g//cols)*(16+pad)+pad; base=S+g*32
        for y in range(16):
            r=(blob[base+y*2]<<8)|blob[base+y*2+1]
            for x in range(16):
                if (r>>(15-x))&1: px[gx+x,gy+y]=255
    img.resize((img.width*scale,img.height*scale),Image.NEAREST).save(f"assets_out/font_probe/verify_sec{si:02d}.png")
    print(f"sec{si}: start={S} glyphs={ng}")
for si in (int(x) for x in (sys.argv[1:] or [0,2])):
    render(si)
