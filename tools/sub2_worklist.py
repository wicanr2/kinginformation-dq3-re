#!/usr/bin/env python3
"""剩餘 sub2 handler 接線 worklist:交叉比對 handler 行為(decode_sub2_struct)× 真實 NPC 存在
(sweep_sub2_npcs,修正後 parser)× 已接清單。輸出每個「有 give/take 且有真 NPC 但未接」的項。
用法:capstone docker。
"""
import struct, glob, os
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE="assets_raw/DQ3.EXE"; DG=0x16140; HDR=0x1370; TBL=0x3bb4
exe=open(EXE,"rb").read(); md=Cs(CS_ARCH_X86,CS_MODE_16)
hw=lambda o:struct.unpack_from("<H",exe,DG+o)[0]

ITEM={0x33:"皇冠",0x55:"盜賊鑰匙",0x56:"魔法鑰匙",0x58:"魔法玉",0x59:"?59",0x5b:"船",
      0x5c:"黑胡椒",0x62:"?62",0x65:"光之珠",0x66:"?66",0x6b:"銀寶珠",0x72:"太陽之石",
      0x73:"雲雨之杖",0x10:"誘惑之劍",0x1f:"?1f",0x4b:"水槍"}
WIRED={7,12,31,52,84,9,26}  # 已接(9=Romaly 特例,26=Portoga 特例)

def block(start,limit=0x80):
    out=[]
    for ins in md.disasm(exe[start:start+limit],start):
        out.append(ins)
        if ins.mnemonic in("ret","retf"):break
        if ins.mnemonic=="jmp":
            try:
                if int(ins.op_str,0)==0x6380:break
            except:pass
    return out

def scan(ins):
    r={"give":None,"take":None,"sflag":[]}; bx=None;pend=None
    for i in ins:
        m,o=i.mnemonic,i.op_str
        if m=="mov" and o.startswith("bx, 0x"):
            try:bx=int(o.split(",")[1],0)
            except:bx=None
        elif m=="mov" and "[0x2593]" in o.split(",")[0]:
            try:pend=int(o.split(",")[1],0)
            except:pend=None
        elif m=="call":
            try:t=int(o,0)
            except:continue
            if t==0x7bbe and pend is not None:r["give"]=pend;pend=None
            elif t==0x7c0c and pend is not None:r["take"]=pend;pend=None
            elif t==0x8264 and bx is not None:r["sflag"].append(bx)
    return r

def behavior(n):
    h=hw(TBL+n*2)
    if h==0 or h==0xffff:return None
    fl=h+HDR; first=block(fl)
    pre=None;je=None;kind=None
    seen=False
    for i in first:
        m,o=i.mnemonic,i.op_str
        if m=="mov" and o.startswith("bx, 0x"):
            try:pre=int(o.split(",")[1],0)
            except:pass
        if m=="call" and o=="0x8279":seen=True
        if m in("je","jne") and seen and je is None:
            try:je=int(o,0);kind=m
            except:pass
    b=scan(first)
    gb=scan(block(je)) if je else {"give":None,"take":None,"sflag":[]}
    give=gb["give"] if gb["give"] is not None else b["give"]
    take=gb["take"] if gb["take"] is not None else b["take"]
    return dict(h=h,pre=pre,kind=kind,give=give,take=take)

# 真實 NPC:byte4 → [(cty,sect,x,y)]
npcs={}
for path in sorted(glob.glob("assets_raw/CTY*.DAT")):
    d=open(path,"rb").read()
    u16=lambda o:struct.unpack_from("<H",d,o)[0] if o+2<=len(d) else 0xffff
    cty=int(os.path.basename(path).replace("CTY","").replace(".DAT","").lstrip("0") or "0")
    for si in range(64):
        so=u16(2*si)
        if so==0xffff:continue
        if so+0x16>len(d):break
        np=u16(so)
        if np==0xffff:np=u16(so+2)
        if np==0xffff or so+np>=len(d):continue
        base=so+np;cnt=d[base]
        for i in range(cnt):
            r=base+1+i*7
            if r+7>len(d):break
            x,y,b2,ctrl,b4,fl,b7=d[r:r+7]
            if x>=64 or y>=64:continue
            if ((ctrl>>3)&7)==2:
                npcs.setdefault(b4,[]).append((cty,si,x,y))

print("byte4 | 行為 | 道具 | prereq | 真實 NPC 位置 | 狀態")
print("|---|---|---|---|---|---|")
for n in sorted(npcs):
    bh=behavior(n)
    if not bh or (bh["give"] is None and bh["take"] is None):continue
    act=[]
    if bh["give"] is not None:act.append("GIVE 0x%02x(%s)"%(bh["give"],ITEM.get(bh["give"],"?")))
    if bh["take"] is not None:act.append("TAKE 0x%02x(%s)"%(bh["take"],ITEM.get(bh["take"],"?")))
    pre="%s0x%02x"%("!" if bh["kind"]=="jne" else "",bh["pre"]) if bh["pre"] is not None else "-"
    locs=";".join("CTY%d s%d(%d,%d)"%l for l in npcs[n][:3])
    st="✅已接" if n in WIRED else "○未接"
    print("| %d | %s | - | %s | %s | %s |"%(n," / ".join(act),pre,locs,st))
