import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image
d=open('assets_raw/PACKBG.SCR','rb').read()
pal=load_pal('assets_raw/MNSBK.PAL')
def planar_rowint(buf, W, H):
    # row-interleaved planar: 每 row 4 planes 連續(plane0..3),每 plane W/8 byte
    px=[[0]*W for _ in range(H)]; pb=W//8
    o=0
    for r in range(H):
        for pl in range(4):
            for b in range(pb):
                if o>=len(buf): return px
                v=buf[o]; o+=1
                for bit in range(8):
                    if v&(0x80>>bit): px[r][b*8+bit]|=(1<<pl)
    return px
def planemajor(buf,W,H):
    px=[[0]*W for _ in range(H)]; pb=W//8; plane=pb*H
    for pl in range(4):
        base=pl*plane
        for r in range(H):
            for b in range(pb):
                o=base+r*pb+b
                if o>=len(buf):continue
                v=buf[o]
                for bit in range(8):
                    if v&(0x80>>bit): px[r][b*8+bit]|=(1<<(3-pl))
    return px
def save(px,W,H,name):
    im=Image.new('RGB',(W,H));p=im.load()
    for r in range(H):
        for c in range(W):
            v=px[r][c]; p[c,r]=pal[v] if v<len(pal) else (255,0,255)
    im.save(name);print('saved',name,W,H)
pos=0x13*0x3d80+0x13
buf=d[pos:pos+0x6e00]
# 0x6e00=28160=640*88*4/8 → 640×88
save(planar_rowint(buf,640,88),640,88,'work/packbg_ri_640x88.png')
save(planemajor(buf,640,88),640,88,'work/packbg_pm_640x88.png')
# 也試 page 0(可能是完整背景),640×88
buf0=d[0:0x6e00]
save(planar_rowint(buf0,640,88),640,88,'work/packbg0_ri.png')
