import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image, ImageDraw
d=open('assets_raw/PACKBG.SCR','rb').read()
pal=load_pal('assets_raw/MNSBK.PAL')
W,H=640,88;pb=W//8
n_total=len(d)//0x6e00
def decode_at(off):
    buf=d[off:off+0x6e00];rows=[];o=0
    for r in range(H):
        row=[0]*W
        for pl in range(4):
            for b in range(pb):
                v=buf[o] if o<len(buf) else 0;o+=1
                for bit in range(8):
                    if v&(0x80>>bit):row[b*8+bit]|=(1<<pl)
        rows.append(row)
    return rows
idxs=list(range(0,n_total,10))[:14]
sh=H//3
big=Image.new('RGB',(W,(sh+12)*len(idxs)),(20,20,20));dr=ImageDraw.Draw(big)
for i,k in enumerate(idxs):
    rows=decode_at(k*0x6e00)
    im=Image.new('RGB',(W,H));p=im.load()
    for r in range(H):
        for c in range(W):v=rows[r][c];p[c,r]=pal[v] if v<len(pal) else (255,0,255)
    im=im.resize((W,sh))
    dr.text((4,i*(sh+12)+1),'idx %d'%k,fill=(255,255,120))
    big.paste(im,(0,i*(sh+12)+10))
big.save('work/packbg_scan.png');print('total bg',n_total)
