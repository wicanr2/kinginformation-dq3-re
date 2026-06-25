#!/usr/bin/env python3
"""地表「無船」真實可達性(gate 2)= 步行 flood-fill + 轉場 traversal(洞穴/旅の扉:出口 spawn 可能在別洲)。
模型:走連續地形(attr bit0=0)→ 進城/洞(cty_loc)→ 走轉場到別 CTY → 該 CTY 出城(0xff)到某 overworld
位置(可能別洲)→ 從該位置再 flood。無船時海(bit0=1)阻斷步行。資料:DQ3CON.MAP+BLKBM+cty_loc+map_graph。"""
import struct,re,os,json,glob
from collections import deque
ROOT=os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
b=open(os.path.join(ROOT,'assets_raw/DQ3CON.MAP'),'rb').read()
W=b[0]|(b[1]<<8); H=b[2]|(b[3]<<8); mp=b[4:4+W*H]
att=open(os.path.join(ROOT,'assets_raw/BLKBM.DAT'),'rb').read()
def attr(t): return struct.unpack_from('<H',att,t*2)[0] if t*2+1<len(att) else 1
def walk(x,y): return 0<=x<W and 0<=y<H and (attr(mp[y*W+x])&1)==0
src=open(os.path.join(ROOT,'dq3_remake/src/dq3_exedata.c')).read()
m=re.search(r'dq3x_cty_loc\[DQ3X_CTYLOC_N\]\[3\] = \{(.*?)\};',src,re.S)
rows=re.findall(r'\{(\d+),(\d+),0x([0-9a-fA-F]+)\}',m.group(1))
loc={i:(int(x),int(y),int(mp_,16)) for i,(x,y,mp_) in enumerate(rows)}
exist=set(int(os.path.basename(p)[3:5]) for p in glob.glob(os.path.join(ROOT,'assets_raw/CTY*.DAT')))
SURF=set(i for i in exist if i<len(loc) and loc[i][2]==0)
graph=json.load(open(os.path.join(ROOT,'docs/maps/map_graph.json')))
# CTY → 轉場目的 CTY(跨 CTY);CTY → 0xff 出口 spawn(地表落點)
cross={}; exits={}
for c,secs in graph.items():
    c=int(c); cross[c]=set(); exits[c]=[]
    for sec,doors in secs.items():
        for dr in doors:
            dc=dr['dest_cty']
            if dc!=c and dc<100 and dc in exist: cross[c].add(dc)
            if dr['dest_sec']==0xff:
                sp=dr['spawn']; 
                if sp!=[0,0]: exits[c].append(tuple(sp))
seen=[[False]*W for _ in range(H)]
reach_cty=set()
def flood(sx,sy):
    if not (0<=sx<W and 0<=sy<H): return
    if seen[sy][sx] or not walk(sx,sy):
        # 起點本身可能是城 tile(可走);仍從鄰格擴
        if not walk(sx,sy): 
            for dx,dy in((0,1),(0,-1),(1,0),(-1,0)):
                if walk(sx+dx,sy+dy) and not seen[sy+dy][sx+dx]: flood(sx+dx,sy+dy)
            return
    dq=deque([(sx,sy)]); seen[sy][sx]=True
    while dq:
        x,y=dq.popleft()
        # 命中城入口?
        for i in SURF:
            lx,ly,_=loc[i]
            if abs(lx-x)<=1 and abs(ly-y)<=1: reach_cty.add(i)
        for dx,dy in((0,1),(0,-1),(1,0),(-1,0)):
            nx,ny=x+dx,y+dy
            if walk(nx,ny) and not seen[ny][nx]: seen[ny][nx]=True; dq.append((nx,ny))
# 起點 flood
sx,sy,_=loc[0]; flood(sx,sy)
# fixpoint:可達 CTY 的轉場目的 + 出口落點 再 flood
changed=True
while changed:
    changed=False
    for c in list(reach_cty):
        for dc in cross.get(c,()):
            if dc not in reach_cty: reach_cty.add(dc); changed=True
            for (ex,ey) in exits.get(dc,()):
                if 0<=ex<W and 0<=ey<H and not seen[ey][ex]:
                    flood(ex,ey); changed=True
        for (ex,ey) in exits.get(c,()):
            if 0<=ex<W and 0<=ey<H and not seen[ey][ex]:
                flood(ex,ey); changed=True
footcells=sum(r.count(True) for r in seen)
rs=sorted(reach_cty&SURF)
locked=sorted(SURF-reach_cty)
print('# 無船真實可達(步行 + 轉場 traversal)\n')
print('步行可走格(含轉場後各洲):%d / %d'%(footcells,W*H))
print('地表城 %d:無船可達 %d、仍需船 %d'%(len(SURF),len(rs),len(locked)))
print('\n## 無船可達城:\n ',rs)
print('\n## 仍需船(隔海未連)城:\n ',locked)

# --- with-ship:船在海上移動(海 tile attr 有海旗標;近似=可走 land∪sea,排除山等真阻擋)---
# 海 tile = 最常見阻擋 tile(index 88);ship 可走 land(bit0=0)或 sea(該 tile)
from collections import Counter
seatile=Counter(mp).most_common(1)[0][0]
def walk_ship(x,y):
    if not(0<=x<W and 0<=y<H): return False
    t=mp[y*W+x]
    return (attr(t)&1)==0 or t==seatile
seen2=[[False]*W for _ in range(H)]
dq=deque([(loc[0][0],loc[0][1])]); seen2[loc[0][1]][loc[0][0]]=True
while dq:
    x,y=dq.popleft()
    for dx,dy in((0,1),(0,-1),(1,0),(-1,0)):
        nx,ny=x+dx,y+dy
        if walk_ship(nx,ny) and not seen2[ny][nx]: seen2[ny][nx]=True; dq.append((nx,ny))
ship_cty=[i for i in SURF if any(0<=loc[i][0]+dx<W and 0<=loc[i][1]+dy<H and seen2[loc[i][1]+dy][loc[i][0]+dx] for dx in(-1,0,1) for dy in(-1,0,1))]
print('\n## 有船(海亦可走)可達地表城:%d / %d'%(len(ship_cty),len(SURF)))
print('  有船仍不可達:', sorted(SURF-set(ship_cty)) or '(無 — 船解鎖全地表)')
