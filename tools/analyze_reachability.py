#!/usr/bin/env python3
"""世界可達性 / 可破關結構分析(資料驅動)。
組合:① map_graph.json(各 CTY section +0xc 轉場 {dest_cty})② cty_loc(地表進城表:
map 0=地表/1=下層/0xff=純迷宮)③ 現存 CTY*.DAT 清單。
從 CTY0 起算可達 CTY,找:不可達 CTY、壞掉的轉場目標(指向不存在的 CTY)、地表↔下層連接點。
注意:未模型化地形/船 gate(地表內非全可走;早期被海阻隔),屬下一層分析。"""
import json,glob,os,re,sys
ROOT=os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
graph=json.load(open(os.path.join(ROOT,'docs/maps/map_graph.json')))
# 現存 CTY 檔
exist=set()
for p in glob.glob(os.path.join(ROOT,'assets_raw/CTY*.DAT')):
    exist.add(int(os.path.basename(p)[3:5]))
# cty_loc:從 dq3_exedata.c 解析
src=open(os.path.join(ROOT,'dq3_remake/src/dq3_exedata.c')).read()
m=re.search(r'dq3x_cty_loc\[DQ3X_CTYLOC_N\]\[3\] = \{(.*?)\};',src,re.S)
rows=re.findall(r'\{(\d+),(\d+),0x([0-9a-fA-F]+)\}',m.group(1))
cty_loc={i:(int(x),int(y),int(mp,16)) for i,(x,y,mp) in enumerate(rows)}
SURF=[i for i,(x,y,mp) in cty_loc.items() if mp==0 and i in exist]
UNDER=[i for i,(x,y,mp) in cty_loc.items() if mp==1 and i in exist]
DUNGEON=[i for i in exist if i<len(cty_loc) and cty_loc[i][2]==0xff]  # 無地表進城=純迷宮

# 轉場邊:CTY a → dest_cty
edges={}
broken=[]
for cty,secs in graph.items():
    cty=int(cty); dests=set()
    for sec,doors in secs.items():
        for dr in doors:
            dc=dr['dest_cty']
            if dc==cty or dc>=100: continue       # 同城/特殊(出城 0xff 等已在 dest_sec)
            dests.add(dc)
            if dc not in exist: broken.append((cty,dc))
    edges[cty]=dests

# 可達性 BFS:CTY0 + 地表 hub(假設地表互通,標註限制)
reach=set([0]); 
def add_hub(hub): 
    for c in hub: reach.add(c)
add_hub(SURF)                 # 地表 hub(樂觀:未扣地形/船)
changed=True
while changed:
    changed=False
    for c in list(reach):
        for d in edges.get(c,()):
            if d not in reach and d in exist: reach.add(d); changed=True
    # 若已達任一下層 CTY → 開放下層 hub
    if any(c in reach for c in UNDER) and not all(c in reach for c in UNDER):
        add_hub(UNDER); changed=True
# 下層連接點:哪些(地表/迷宮)CTY 轉場到下層 CTY
down_links=[(c,d) for c in reach for d in edges.get(c,()) if d in UNDER]

unreach=sorted(c for c in exist if c not in reach)
print('# 世界可達性 / 可破關結構分析\n')
print('現存 CTY 檔:%d(地表進城 %d、下層進城 %d、純迷宮 %d)'%(len(exist),len(SURF),len(UNDER),len(DUNGEON)))
print('從 CTY0 可達:%d / %d'%(len(reach&exist),len(exist)))
print()
print('## 不可達 CTY(%d)'%len(unreach))
print(' ', unreach if unreach else '(無 — 全可達)')
print('\n## 壞掉的轉場(目標 CTY 無檔案,%d)'%len(broken))
for c,d in broken[:20]: print('  CTY%d → CTY%d(不存在)'%(c,d))
if not broken: print('  (無)')
print('\n## 地表 → 下層 連接點(%d)'%len(down_links))
for c,d in down_links[:20]: print('  CTY%d → CTY%d(下層)'%(c,d))
if not down_links: print('  ⚠ 無 — 下層世界從地表無轉場入口(可能靠特殊 destSec/地形洞)')
