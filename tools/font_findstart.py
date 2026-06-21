#!/usr/bin/env python3
"""量化找字模起點:對每個 section,掃描 header 長度 H,計算 H 之後前 K 個
32-byte 區塊的『字模相似度』。表格/索引區=很多垂直常數欄(條紋);字模=列變化多、
密度適中。回報每個 section 最可能的字模起始相對位移。"""
import struct
data=open("assets_raw/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',data,i*4)[0] for i in range(22)]+[len(data)]

def rows16(base):
    return [ (data[base+y*2]<<8)|data[base+y*2+1] for y in range(16) ]

def glyph_score(base):
    """越像真實 CJK 字模分數越高 (0..1)。"""
    if base+32>len(data): return -1
    r=rows16(base)
    ink=sum(bin(x).count('1') for x in r)
    density=ink/256
    # 垂直常數欄數:某一 bit 欄在 16 列中全 0 或全 1
    colconst=0
    for c in range(16):
        col=[(r[y]>>(15-c))&1 for y in range(16)]
        if all(v==col[0] for v in col): colconst+=1
    full_rows=sum(1 for x in r if x in (0x0000,0xFFFF))
    distinct=len(set(r))
    # 字模特徵:density 0.10~0.45、colconst 少、distinct 多、不要太多 full row
    s=0.0
    if 0.08<=density<=0.50: s+=1
    s+= max(0,(12-colconst))/12        # 垂直常數欄越少越好
    s+= min(distinct,14)/14            # 列變化越多越好
    s-= full_rows/16
    return s

print(f"{'sec':>3} | best H | run-score | 開頭8byte")
for i in range(22):
    s=offs[i]; size=offs[i+1]-s
    best=(None,-9)
    for H in range(0, min(2200,size-32*8), 2):
        # 評估從 H 起連續 8 個字模的平均分
        sc=sum(glyph_score(s+H+g*32) for g in range(8))/8
        if sc>best[1]: best=(H,sc)
    print(f"{i:3} | {best[0]:5} | {best[1]:8.3f} | {data[s:s+8].hex()}")
