import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image
d=open('assets_raw/PACKBG.SCR','rb').read()
pal=load_pal('assets_raw/MNSBK.PAL')
W,H=320,176; pb=W//8; plane=pb*H  # 40*176=7040
def planemajor(buf,order):
    px=[[0]*W for _ in range(H)]
    for s in range(4):
        pl=order[s]; base=s*plane
        for r in range(H):
            for b in range(pb):
                o=base+r*pb+b
                if o>=len(buf):continue
                v=buf[o]
                for bit in range(8):
                    if v&(0x80>>bit):px[r][b*8+bit]|=(1<<pl)
    return px
def save(px,name,scale=2):
    im=Image.new('RGB',(W*scale,H*scale));p=im.load()
    for r in range(H):
        for c in range(W):
            v=px[r][c];col=pal[v] if v<len(pal) else (255,0,255)
            for sy in range(scale):
                for sx in range(scale):p[c*scale+sx,r*scale+sy]=col
    im.save(name);print('saved',name)
pos=0x13*0x3d80+0x13
buf=d[pos:pos+0x6e00]
save(planemajor(buf,(0,1,2,3)),'work/packbg_pm0123.png')
save(planemajor(buf,(3,2,1,0)),'work/packbg_pm3210.png')
