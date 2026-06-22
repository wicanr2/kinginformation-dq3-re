import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image
d=open('assets_raw/PACKBG.SCR','rb').read()
pal=load_pal('assets_raw/MNSBK.PAL')
def planar_rowint(buf,W,H):
    px=[[0]*W for _ in range(H)];pb=W//8;o=0
    for r in range(H):
        for pl in range(4):
            for b in range(pb):
                if o>=len(buf):return px
                v=buf[o];o+=1
                for bit in range(8):
                    if v&(0x80>>bit):px[r][b*8+bit]|=(1<<pl)
    return px
def save(px,W,H,name,scale=2):
    im=Image.new('RGB',(W*scale,H*scale));p=im.load()
    for r in range(H):
        for c in range(W):
            v=px[r][c];col=pal[v] if v<len(pal) else (255,0,255)
            for sy in range(scale):
                for sx in range(scale):p[c*scale+sx,r*scale+sy]=col
    im.save(name);print('saved',name)
pos=0x13*0x3d80+0x13
save(planar_rowint(d[pos:pos+0x6e00],320,176),320,176,'work/packbg_320x176.png')
