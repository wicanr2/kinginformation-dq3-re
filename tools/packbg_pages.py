import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image, ImageDraw
d=open('assets_raw/PACKBG.SCR','rb').read()
pal=load_pal('assets_raw/MNSBK.PAL')
W,H=640,88; pb=W//8
def decode(page):
    pos=page*0x3d80+page; buf=d[pos:pos+0x6e00]
    rows=[];o=0
    for r in range(H):
        row=[0]*W
        for pl in range(4):
            for b in range(pb):
                v=buf[o] if o<len(buf) else 0;o+=1
                for bit in range(8):
                    if v&(0x80>>bit):row[b*8+bit]|=(1<<pl)
        rows.append(row)
    return rows
pages=[0,8,16,24,0x19,32,40]
sh=H//2
big=Image.new('RGB',(W,(sh+16)*len(pages)),(20,20,20));d2=ImageDraw.Draw(big)
for i,pg in enumerate(pages):
    rows=decode(pg)
    im=Image.new('RGB',(W,H));p=im.load()
    for r in range(H):
        for c in range(W):v=rows[r][c];p[c,r]=pal[v] if v<len(pal) else (255,0,255)
    im=im.resize((W,sh))
    d2.text((4,i*(sh+16)+2),'page %d (0x%x)'%(pg,pg),fill=(255,255,120))
    big.paste(im,(0,i*(sh+16)+14))
big.save('work/packbg_pages.png');print('saved')
