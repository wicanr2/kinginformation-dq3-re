import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image, ImageDraw
body=open('assets_raw/DQ3MAN.BLS','rb').read()[6:]
pal=load_pal('assets_raw/DQ3.PAL')
def planes(base):
    px=[[0]*32 for _ in range(24)]
    for s,pl in enumerate((3,2,1,0)):
        b0=base+s*96
        for r in range(24):
            for b in range(4):
                v=body[b0+r*4+b]
                for bit in range(8):
                    if v&(0x80>>bit): px[r][b*8+bit]|=(1<<pl)
    msk=[[0]*32 for _ in range(24)]; mb=base+0x180
    for r in range(24):
        for b in range(4):
            v=body[mb+r*4+b]
            for bit in range(8): msk[r][b*8+bit]=(v>>(7-bit))&1
    return px,msk
sc=8; img=Image.new('RGB',(4*(32*sc+6)+6,24*sc+24),(0,150,0)); pp=img.load(); d=ImageDraw.Draw(img)
for i,e in enumerate((16,17,18,19)):
    px,msk=planes(e*960); ox=6+i*(32*sc+6)
    d.text((ox,2),'frame %d'%(e-16),fill=(255,255,255))
    for r in range(24):
        for c in range(32):
            col=(0,150,0) if msk[r][c]==1 else pal[px[r][c]]
            for sy in range(sc):
                for sx in range(sc): pp[ox+c*sc+sx,16+r*sc+sy]=col
img.save('work/hero16_frames.png'); print('saved work/hero16_frames.png')
