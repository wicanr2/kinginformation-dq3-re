#!/usr/bin/env python3
"""自動判斷字模是否被混淆:比較 [原始] vs [去混淆(左右對調+左半下移1)] 的筆畫連通性,
取連通性高者。連通性 = 平均每個亮點的 4-鄰亮點數(正字筆畫連續→高;亂點→低)。"""
import os, sys
import importlib.util
spec=importlib.util.spec_from_file_location("g","tools/font_greedy.py")
g=importlib.util.module_from_spec(spec); spec.loader.exec_module(g)
si=int(sys.argv[1]) if len(sys.argv)>1 else 0
glyphs=g.extract_section(si); blob=g.data[g.offs[si]:g.offs[si+1]]

def orig(off): return [(blob[off+y*2]<<8)|blob[off+y*2+1] for y in range(16)]
def deobf(off):
    hi=[blob[off+y*2] for y in range(16)]; lo=[blob[off+y*2+1] for y in range(16)]
    return [((lo[y-1] if y>=1 else 0)<<8)|hi[y] for y in range(16)]

def connectivity(rows):
    px=[[(rows[y]>>(15-x))&1 for x in range(16)] for y in range(16)]
    ink=sum(sum(r) for r in px)
    if ink==0: return 0,0
    nb=0
    for y in range(16):
        for x in range(16):
            if not px[y][x]: continue
            for dy,dx in ((1,0),(-1,0),(0,1),(0,-1)):
                ny,nx=y+dy,x+dx
                if 0<=ny<16 and 0<=nx<16 and px[ny][nx]: nb+=1
    return nb/ink, ink

if __name__=="__main__":
    a=int(sys.argv[2]) if len(sys.argv)>2 else 38
    b=int(sys.argv[3]) if len(sys.argv)>3 else 48
    print("idx | conn_orig | conn_deobf | 判定")
    for i in range(a,min(b,len(glyphs))):
        off=glyphs[i][0]
        co,_=connectivity(orig(off)); cd,_=connectivity(deobf(off))
        verdict="混淆(用解)" if cd>co+0.15 else "正常(用原)"
        print(f"{i:3} | {co:8.2f} | {cd:9.2f} | {verdict}")
