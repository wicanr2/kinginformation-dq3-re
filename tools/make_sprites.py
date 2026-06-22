#!/usr/bin/env python3
"""把未完成的 sprite(128 歐里狄加 / 129 五頭龍大王)用參考圖重繪成精訊 DQ3 風格,
注入 DQ3MNS.SHP。sprite 格式同 mns_sprite.decode_sprite,色 = mnsbk.pal 前 16 色,0=透明。"""
import struct, os
from collections import deque
from PIL import Image

SHP = "assets_raw/DQ3MNS.SHP"; PAL = "assets_raw/MNSBK.PAL"
pd = open(PAL, "rb").read()
pal = [(pd[i*3] << 2, pd[i*3+1] << 2, pd[i*3+2] << 2) for i in range(16)]

def nearest(rgb):
    best, bi = 1e9, 1
    for i in range(1, 16):
        d = sum((rgb[k]-pal[i][k])**2 for k in range(3))
        if d < best: best, bi = d, i
    return bi

def near(a, b, t): return all(abs(a[i]-b[i]) <= t for i in range(3))

def build_grid(ref, Wb, H, bg):
    """bg: ('flood', crop_frac) 角落同色去背 / ('black', thr) 近黑去背(先裁黑框)。"""
    W = Wb*8
    src = Image.open(ref).convert("RGB")
    sw, sh = src.size
    if bg[0] == 'flood':
        src = src.crop((0,0,sw,int(sh*bg[1])))
    elif bg[0] == 'black':
        # 裁到「近黑大方框」bbox(去掉外圍灰邊)
        thr = bg[1]; px=src.load()
        xs=[x for x in range(sw) for y in (0,sh//2,sh-1) if max(px[x,y])<thr] or [0,sw-1]
        ys=[y for y in range(sh) for x in (0,sw//2,sw-1) if max(px[x,y])<thr] or [0,sh-1]
        # 用全掃 bbox 較穩
        minx=miny=10**9; maxx=maxy=-1
        for y in range(sh):
            for x in range(sw):
                if max(px[x,y])<thr:
                    minx=min(minx,x);maxx=max(maxx,x);miny=min(miny,y);maxy=max(maxy,y)
        if maxx>minx: src=src.crop((minx,miny,maxx+1,maxy+1))
    img = src.resize((W, H), Image.LANCZOS); px = img.load()
    isbg = [[False]*W for _ in range(H)]
    if bg[0] == 'flood':
        refc = px[0,0]; q=deque()
        for s in [(0,0),(W-1,0),(0,H-1),(W-1,H-1)]:
            isbg[s[1]][s[0]]=True; q.append(s)
        while q:
            x,y=q.popleft()
            for dx,dy in ((1,0),(-1,0),(0,1),(0,-1)):
                nx,ny=x+dx,y+dy
                if 0<=nx<W and 0<=ny<H and not isbg[ny][nx] and near(px[nx,ny],refc,44):
                    isbg[ny][nx]=True; q.append((nx,ny))
    else:  # black
        thr=bg[1]
        for y in range(H):
            for x in range(W):
                if max(px[x,y])<thr: isbg[y][x]=True
    grid=[[0]*W for _ in range(H)]; ink=0
    for y in range(H):
        for x in range(W):
            if not isbg[y][x]: grid[y][x]=nearest(px[x,y]); ink+=1
    return Wb,H,grid,ink

def encode(Wb,H,grid):
    out=bytearray(struct.pack("<HH",Wb,H))
    for seg in range(4):
        pb=3-seg
        for row in range(H):
            for bx in range(Wb):
                b=0
                for bit in range(8):
                    if (grid[row][bx*8+bit]>>pb)&1: b|=(0x80>>bit)
                out.append(b)
    return bytes(out)

JOBS = {
    128: ("dq3_org_pic/128-sprite.jpeg", 12, 88, ('black', 50)),
    129: ("dq3_org_pic/九頭蛇.jpg",      16, 96, ('flood', 0.91)),
}
sprites={}
for mid,(ref,Wb,H,bg) in JOBS.items():
    _,_,grid,ink = build_grid(ref,Wb,H,bg)
    sprites[mid]=encode(Wb,H,grid)
    print(f"id {mid}: {Wb*8}x{H} ink={ink} bytes={len(sprites[mid])}")

# 注入:128,129 連續(其後即 sentinel 130)。重建 offset table 尾段。
shp=bytearray(open(SHP,"rb").read())
offs=[struct.unpack_from("<I",shp,i*4)[0] for i in range(131)]
new=bytearray(shp[:offs[128]])
o128=len(new); new+=sprites[128]
o129=len(new); new+=sprites[129]
struct.pack_into("<I",new,128*4,o128)
struct.pack_into("<I",new,129*4,o129)
struct.pack_into("<I",new,130*4,len(new))
os.makedirs("work",exist_ok=True)
open("work/DQ3MNS_fixed.SHP","wb").write(new)
print(f"injected 128@{o128} 129@{o129} -> work/DQ3MNS_fixed.SHP ({len(new)} bytes)")
