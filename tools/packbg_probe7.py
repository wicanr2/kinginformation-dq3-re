import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image
d=open('assets_raw/PACKBG.SCR','rb').read()
pal=load_pal('assets_raw/MNSBK.PAL')
def decode_page(page,nrows1=88,W=640):
    pos=page*0x3d80+page
    buf=d[pos:pos+0x6e00]
    pb=W//8; rows=[]; o=0
    for r in range(nrows1):
        row=[0]*W
        for pl in range(4):
            for b in range(pb):
                if o>=len(buf): break
                v=buf[o]; o+=1
                for bit in range(8):
                    if v&(0x80>>bit): row[b*8+bit]|=(1<<pl)
        rows.append(row)
    return rows
def save(rows,name,W=640):
    H=len(rows); im=Image.new('RGB',(W,H));p=im.load()
    for r in range(H):
        for c in range(W): v=rows[r][c]; p[c,r]=pal[v] if v<len(pal) else (255,0,255)
    im.resize((W,H*2),Image.NEAREST).save(name);print('saved',name)
for pg in (0,8,0x19):
    save(decode_page(pg),'work/pbpage_%02x.png'%pg)
