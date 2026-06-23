#!/usr/bin/env python3
"""從 D3TXT00.TXT(全名稱表)+ D3TXT00.FON(16x16 字型)render 名稱記錄區間。
   用法:render_names.py <lo> <hi> [out.png]。glyph 碼 16-bit + 0xffff 終結。"""
import struct,sys
from PIL import Image, ImageDraw
d=open("assets_raw/D3TXT00.TXT","rb").read()
fon=open("assets_raw/D3TXT00.FON","rb").read()
def u16(o): return struct.unpack_from("<H",d,o)[0]
N=u16(0)//2
offs=[u16(i*2) for i in range(N+1)]
def glyph(c):
    base=c*32; im=Image.new("1",(16,16)); px=im.load()
    for row in range(16):
        w=(fon[base+row*2]<<8|fon[base+row*2+1]) if base+row*2+1<len(fon) else 0
        for col in range(16):
            if w&(0x8000>>col): px[col,row]=1
    return im
def rec_img(r):
    if r>=len(offs)-1: return None
    s=offs[r]; ws=[]; o=s
    while o+1<len(d) and u16(o)!=0xffff and len(ws)<14: ws.append(u16(o)); o+=2
    if not ws: ws=[0]
    im=Image.new("1",(16*len(ws),16))
    for i,c in enumerate(ws): im.paste(glyph(c),(i*16,0))
    return im
lo,hi=int(sys.argv[1]),int(sys.argv[2])
out=sys.argv[3] if len(sys.argv)>3 else "/out/r_%d_%d.png"%(lo,hi)
rows=[]
for r in range(lo,hi):
    im=rec_img(r)
    if im is None: break
    lab=Image.new("1",(46,16)); ImageDraw.Draw(lab).text((0,3),str(r),fill=1)
    row=Image.new("1",(46+im.width,16)); row.paste(lab,(0,0)); row.paste(im,(46,0))
    rows.append(row)
W=max(r.width for r in rows); H=18*len(rows)
c=Image.new("1",(W,H),0); y=0
for r in rows: c.paste(r,(0,y)); y+=18
c.resize((W*2,H*2)).save(out); print("saved %s rows %d..%d (N=%d)"%(out,lo,hi,N))
