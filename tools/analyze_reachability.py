#!/usr/bin/env python3
"""世界可達性 / 可破關結構分析。
★ 用 docs/maps/map_graph.json(extract_map_graph.py 以**轉場 tile subid 正確界定**抽出的轉場圖)
為權威來源 —— 不可盲讀固定 N 個 entry(會溢出表尾讀到 garbage 假邊,踩過雷,見 docs/43 教訓)。
組合 cty_loc(地表/下層進城)算從 CTY0 經轉場可達的 CTY。未模型化地形/船 gate。"""
import json,glob,os,re
ROOT=os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
exist=set(int(os.path.basename(p)[3:5]) for p in glob.glob(os.path.join(ROOT,'assets_raw/CTY*.DAT')))
graph=json.load(open(os.path.join(ROOT,'docs/maps/map_graph.json')))
src=open(os.path.join(ROOT,'dq3_remake/src/dq3_exedata.c')).read()
m=re.search(r'dq3x_cty_loc\[DQ3X_CTYLOC_N\]\[3\] = \{(.*?)\};',src,re.S)
rows=re.findall(r'\{(\d+),(\d+),0x([0-9a-fA-F]+)\}',m.group(1))
loc={i:(int(x),int(y),int(mp,16)) for i,(x,y,mp) in enumerate(rows)}
SURF=set(i for i in exist if i<len(loc) and loc[i][2]==0)
UNDER=set(i for i in exist if i<len(loc) and loc[i][2]==1)
DUNG=set(i for i in exist if i<len(loc) and loc[i][2]==0xff)
edges={}
for cty,secs in graph.items():
    cty=int(cty); ds=set()
    for sec,doors in secs.items():
        for dr in doors:
            dc=dr['dest_cty']
            if dc!=cty and dc<100 and dc in exist: ds.add(dc)
    edges[cty]=ds
# 地表 hub(出城回地表 = 地表城互通,樂觀未扣地形/船);BFS 轉場邊
reach=set([0])|SURF
ch=True
while ch:
    ch=False
    for c in list(reach):
        for dd in edges.get(c,()):
            if dd not in reach: reach.add(dd); ch=True
unreach=sorted(exist-reach)
downlinks=[(c,dd) for c in (reach&SURF) for dd in edges.get(c,()) if dd in UNDER]
print('# 世界可達性(權威:map_graph.json,正確界定)\n')
print('現存 CTY %d(地表 %d / 下層 %d / 純迷宮 %d)'%(len(exist),len(SURF),len(UNDER),len(DUNG)))
print('從 CTY0 經轉場可達:%d / %d'%(len(reach&exist),len(exist)))
print('不可達(%d):%s'%(len(unreach),unreach))
print('\n地表→下層 跨 CTY 轉場邊:%s'%(downlinks if downlinks else '(無 → 下層需「下降事件」進入,非轉場圖)'))
print('下層內部跨 CTY 邊:%s'%([(c,dd) for c in (reach|UNDER) for dd in edges.get(c,()) if dd in UNDER and c in UNDER]))
