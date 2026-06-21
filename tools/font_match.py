#!/usr/bin/env python3
"""地面真值對齊:用已驗證正確的 D3TXT00.FON 字模當範本,在 CHINA.FON 逐 byte
搜尋完全相符的 32-byte 片段。比中的 byte offset 揭露 CHINA 字模真實對齊(起點/相位)。"""
import struct
d3=open("assets_raw/D3TXT00.FON","rb").read()
china=open("assets_raw/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',china,i*4)[0] for i in range(22)]+[len(china)]

# D3TXT00 全部 32-byte 字模 -> set
tmpls={}
for i in range(len(d3)//32):
    g=d3[i*32:(i+1)*32]
    # 跳過全空/過簡
    if sum(bin(b).count('1') for b in g) < 24: continue
    tmpls.setdefault(g, i)
print("D3TXT00 範本數(去重後非空):", len(tmpls))

# 逐 byte 搜尋
hits=[]
for off in range(0, len(china)-32):
    seg=china[off:off+32]
    if seg in tmpls:
        hits.append((off, tmpls[seg]))
print("CHINA.FON 完全相符片段數:", len(hits))

# 統計每個 hit 落在哪個 section, 相對 offset, mod 32
def sec_of(off):
    for s in range(22):
        if offs[s]<=off<offs[s+1]: return s
    return -1
from collections import defaultdict
permod=defaultdict(int); persec=defaultdict(list)
for off,idx in hits:
    s=sec_of(off); rel=off-offs[s]
    permod[rel%32]+=1
    persec[s].append(rel)
print("\nhit 的 (相對 offset mod 32) 分佈:", dict(sorted(permod.items())))
print("\n每段最早 hit 的相對 offset(可推 glyph 起點/相位):")
for s in range(22):
    if persec[s]:
        v=sorted(persec[s])
        print(f"  sec{s:2}: 最早 rel={v[0]:5} (mod32={v[0]%32})  hit數={len(v)}  最早幾個={v[:5]}")
