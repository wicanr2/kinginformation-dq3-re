import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image, ImageDraw
d=open('assets_raw/PACKBG.SCR','rb').read()
pal=load_pal('assets_raw/MNSBK.PAL')
W,H=640,88;pb=W//8
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
n=12; sh=H//2
big=Image.new('RGB',(W,(sh+14)*n),(20,20,20));dr=ImageDraw.Draw(big)
for k in range(n):
    off=k*0x6e00
    rows=decode_at(off)
    im=Image.new('RGB',(W,H));p=im.load()
    for r in range(H):
        for c in range(W):v=rows[r][c];p[c,r]=pal[v] if v<len(pal) else (255,0,255)
    im=im.resize((W,sh))
    dr.text((4,k*(sh+14)+1),'idx %d off 0x%x'%(k,off),fill=(255,255,120))
    big.paste(im,(0,k*(sh+14)+12))
big.save('work/packbg_arr.png');print('saved',n,'backgrounds')
