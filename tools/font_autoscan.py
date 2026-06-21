#!/usr/bin/env python3
"""每段:試所有相位(起始 byte 0..30 偶數),數整段合格字模,選最佳相位;
再於最佳相位下用滑動視窗找字模密集區起點(連續視窗合格率>0.6)。"""
import struct, sys
data=open("assets_raw/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',data,i*4)[0] for i in range(22)]+[len(data)]

def rows16(blob,base):
    return [(blob[base+y*2]<<8)|blob[base+y*2+1] for y in range(16)]

def is_glyph(blob,base):
    if base+32>len(blob): return False
    r=rows16(blob,base)
    dens=sum(bin(x).count('1') for x in r)/256
    if not (0.08<=dens<=0.47): return False
    if sum(1 for x in r if x==0)<1: return False
    colconst=sum(1 for c in range(16) if len({(rr>>(15-c))&1 for rr in r})==1)
    if colconst>6: return False
    if len(set(r))<7: return False
    return True

def analyze(si):
    blob=data[offs[si]:offs[si+1]]; L=len(blob)
    best=(0,-1)  # (phase, count)
    for p in range(0,32,2):
        cnt=sum(1 for b in range(p,L-31,32) if is_glyph(blob,b))
        if cnt>best[1]: best=(p,cnt)
    phase,cnt=best
    # 在該相位找起點:第一個「之後 16 視窗合格率>0.6」的 glyph index
    idxs=list(range(phase,L-31,32))
    flags=[is_glyph(blob,b) for b in idxs]
    start=None
    for i in range(len(idxs)):
        w=flags[i:i+16]
        if len(w)>=8 and sum(w)/len(w)>0.6:
            start=idxs[i]; break
    ng=(L-start)//32 if start is not None else 0
    return phase,cnt,start,ng,L

print("sec | phase | good | start_byte | mod32 | glyphs | size")
for si in range(22):
    phase,cnt,start,ng,L=analyze(si)
    sb = start if start is not None else -1
    print(f"{si:3} | {phase:5} | {cnt:4} | {sb:9} | {sb%32 if sb>=0 else -1:5} | {ng:5} | {L}")
