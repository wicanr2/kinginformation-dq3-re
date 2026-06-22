import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image
d=open('assets_raw/PACKBG.SCR','rb').read()
pal=load_pal('assets_raw/MNSBK.PAL')
pos=0x13*0x3d80+0x13
buf=d[pos:pos+0x6e00+0xcf80]
# 每 row = 4 plane × 80 byte。試:取每 plane 後 40 byte → 320 寬影像
def decode(off_in_plane, W):
    pb=80  # plane span in buffer
    rows=[]; o=0
    while o+4*pb<=len(buf):
        row=[0]*W
        for pl in range(4):
            base=o+pl*pb+off_in_plane
            for b in range(W//8):
                v=buf[base+b]
                for bit in range(8):
                    if v&(0x80>>bit): row[b*8+bit]|=(1<<pl)
        rows.append(row); o+=4*pb
        if len(rows)>=254: break
    return rows
def save(rows,W,name):
    H=len(rows); im=Image.new('RGB',(W,H));p=im.load()
    for r in range(H):
        for c in range(W): v=rows[r][c]; p[c,r]=pal[v] if v<len(pal) else (255,0,255)
    im.save(name);print('saved',name,W,H)
save(decode(40,320),320,'work/pb_back40.png')   # 後 40 byte
save(decode(0,320),320,'work/pb_front40.png')    # 前 40 byte
