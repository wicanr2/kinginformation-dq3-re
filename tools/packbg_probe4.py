import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image
d=open('assets_raw/PACKBG.SCR','rb').read()
pal=load_pal('assets_raw/MNSBK.PAL')
W=640; pb=W//8  # 80
def rows_from(buf, off, nrows):
    px=[]; o=off
    for r in range(nrows):
        row=[0]*W
        for pl in range(4):
            for b in range(pb):
                if o>=len(buf): return px
                v=buf[o]; o+=1
                for bit in range(8):
                    if v&(0x80>>bit): row[b*8+bit]|=(1<<pl)
        px.append(row)
    return px,o
pos=0x13*0x3d80+0x13
buf=d[pos:pos+0x6e00+0xcf80]
top,o1=rows_from(buf,0,88)
bot,_=rows_from(buf,0x6e00,158)
px=top+bot; H=len(px)
im=Image.new('RGB',(W,H));p=im.load()
for r in range(H):
    for c in range(W):
        v=px[r][c];p[c,r]=pal[v] if v<len(pal) else (255,0,255)
im.save('work/packbg_full.png');print('saved work/packbg_full.png',W,H)
