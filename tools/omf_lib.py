#!/usr/bin/env python3
"""OMF (.LIB/.OBJ) 解析器 — 用來把 SBCM.LIB 裡的驅動碼定位進 DQ3.EXE。

OMF library = LIBHDR(0xF0) + 對齊 page 的數個 OBJ 模組 + LIBEND(0xF1)。
每筆 record = type(1) + len(2, 含 data+checksum) + data + checksum(1)。
本檔只做「讀」:列模組、PUBDEF(符號:seg:off)、SEGDEF(段名/class)、LEDATA(碼/資料 bytes),
並可抽某符號所在段的 LEDATA + 標出 FIXUPP 動到的 byte(重定位,byte-match 時要遮罩)。

用法:
  python omf_lib.py dump  <lib>            # 列所有模組 + 符號 + 段
  python omf_lib.py bytes <lib> <module>   # 印某模組各段的 LEDATA(hex)+ fixup 遮罩
"""
import sys

REC = {0x80:"THEADR",0x88:"COMENT",0x8a:"MODEND",0x8b:"MODEND",0x8c:"EXTDEF",
       0x90:"PUBDEF",0x91:"PUBDEF",0x96:"LNAMES",0x98:"SEGDEF",0x99:"SEGDEF",
       0x9a:"GRPDEF",0x9c:"FIXUPP",0x9d:"FIXUPP",0xa0:"LEDATA",0xa1:"LEDATA",
       0xa2:"LIDATA",0xa3:"LIDATA",0xb0:"COMDEF",0x88:"COMENT",0xf0:"LIBHDR",0xf1:"LIBEND"}

def rd_idx(d,p):
    """OMF index: 1 byte if <0x80, else 2 bytes (high bit set)."""
    if d[p]<0x80: return d[p],p+1
    return ((d[p]&0x7f)<<8)|d[p+1],p+2

def rd_str(d,p):
    n=d[p]; return d[p+1:p+1+n].decode('latin1'),p+1+n

def parse(path):
    d=open(path,'rb').read()
    assert d[0]==0xf0, "not a LIB (no 0xF0 LIBHDR)"
    rl=d[1]|(d[2]<<8)
    page=rl+3                      # LIBHDR 記錄總長 = page size
    # 模組從第一個 page boundary 開始
    modules=[]; cur=None; lnames=[]; segdefs=[]
    i=page
    n=len(d)
    while i<n:
        t=d[i]
        if t==0xf1: break          # LIBEND
        ln=d[i+1]|(d[i+2]<<8)
        data=d[i+3:i+3+ln-1]       # 去掉尾端 checksum
        nm=REC.get(t,hex(t))
        if t==0x80:                # THEADR -> 新模組
            name,_=rd_str(data,0)
            cur={"name":name,"theadr_file":i,"pubdefs":[],"segdefs":[],"lnames":[],
                 "ledata":[],"fixupp":[]}
            modules.append(cur); lnames=cur["lnames"]; segdefs=cur["segdefs"]
        elif t==0x96 and cur:      # LNAMES
            p=0
            while p<len(data):
                s,p=rd_str(data,p); lnames.append(s)
        elif t in (0x98,0x99) and cur:  # SEGDEF
            acbp=data[0]; p=1
            if (acbp>>5)&7==0:      # alignment 0 has frame+offset
                p+=3
            seglen=data[p]|(data[p+1]<<8); p+=2
            if t==0x99: seglen+=0x10000
            segname,p=rd_idx(data,p)
            classname,p=rd_idx(data,p)
            overlay,p=rd_idx(data,p)
            segdefs.append({"len":seglen,"nameidx":segname,"classidx":classname})
        elif t in (0x90,0x91) and cur:  # PUBDEF
            p=0
            base_grp,p=rd_idx(data,p)
            base_seg,p=rd_idx(data,p)
            if base_seg==0: p+=2    # base frame if seg index 0
            while p<len(data):
                name,p=rd_str(data,p)
                off=data[p]|(data[p+1]<<8); p+=2
                typ,p=rd_idx(data,p)
                cur["pubdefs"].append((name,base_seg,off))
        elif t in (0xa0,0xa1) and cur:  # LEDATA
            p=0; seg,p=rd_idx(data,p)
            if t==0xa1:
                off=data[p]|(data[p+1]<<8)|(data[p+2]<<16)|(data[p+3]<<24); p+=4
            else:
                off=data[p]|(data[p+1]<<8); p+=2
            cur["ledata"].append({"seg":seg,"off":off,"data":data[p:],"rec_file":i})
        elif t in (0x9c,0x9d) and cur:  # FIXUPP (粗略:記下有 fixup 的 LEDATA)
            if cur["ledata"]:
                cur["fixupp"].append((cur["ledata"][-1]["rec_file"],len(data)))
        i+=3+ln
        # 模組結束(MODEND)後對齊到下一 page
        if t in (0x8a,0x8b):
            if i % page: i += page-(i%page)
    return d,modules,page

def segname(m,idx):
    try: return m["lnames"][idx-1]
    except: return "?"

def cmd_dump(path):
    d,mods,page=parse(path)
    print("page size:",page,"modules:",len(mods))
    for m in mods:
        tot=sum(len(l["data"]) for l in m["ledata"])
        print("\n== module %-14s  LEDATA bytes=%d  segs=%d ==" % (m["name"],tot,len(m["segdefs"])))
        for j,s in enumerate(m["segdefs"],1):
            print("   SEG#%d %-12s class=%-8s len=0x%x" % (j,segname(m,s["nameidx"]),segname(m,s["classidx"]),s["len"]))
        for nm,bs,off in m["pubdefs"]:
            print("   PUB %-22s seg#%d:0x%04x" % (nm,bs,off))

def cmd_bytes(path,modname):
    d,mods,page=parse(path)
    for m in mods:
        if modname.lower() not in m["name"].lower(): continue
        print("module",m["name"])
        for l in m["ledata"]:
            b=l["data"]
            print("  LEDATA seg#%d off 0x%04x len 0x%x" % (l["seg"],l["off"],len(b)))
            print("   "+" ".join("%02x"%x for x in b))

def cmd_find(libpath,modname,exepath,k=8):
    """把某模組各段 LEDATA 的 k-byte 視窗滑過 EXE,找命中位置叢集 = 驅動在 EXE 的位置。"""
    d,mods,page=parse(libpath)
    exe=open(exepath,'rb').read()
    # 建 EXE 的 k-mer 索引
    from collections import defaultdict
    idx=defaultdict(list)
    for p in range(len(exe)-k+1):
        idx[exe[p:p+k]].append(p)
    for m in mods:
        if modname.lower() not in m["name"].lower(): continue
        # 依 seg 串接 LEDATA(同段 off 連續)
        segbytes=defaultdict(bytearray)
        for l in m["ledata"]:
            segbytes[l["seg"]] += bytes(l["data"])
        for seg,b in segbytes.items():
            b=bytes(b)
            print("\n== module %s seg#%d  %d bytes ==" % (m["name"],seg,len(b)))
            # 對每個 lib 視窗,記錄它在 exe 的命中(可能多處);找「lib_off vs exe_off 差固定」的叢集
            votes=defaultdict(int); examples=defaultdict(list)
            for lo in range(0,len(b)-k+1):
                w=b[lo:lo+k]
                for eo in idx.get(w,()):
                    delta=eo-lo
                    votes[delta]+=1
                    if len(examples[delta])<3: examples[delta].append((lo,eo))
            best=sorted(votes.items(),key=lambda x:-x[1])[:5]
            for delta,cnt in best:
                if cnt<2: continue
                base=delta  # exe_file = lib_off + delta -> 段在 exe 的起點(file)
                print("   delta=0x%06x  matched_windows=%d  => seg base @ exe file 0x%05x" % (delta&0xffffff,cnt,base&0xffffff))
                for lo,eo in examples[delta]:
                    print("       lib_off 0x%03x -> exe 0x%05x" % (lo,eo))

if __name__=="__main__":
    cmd=sys.argv[1]
    if cmd=="dump": cmd_dump(sys.argv[2])
    elif cmd=="bytes": cmd_bytes(sys.argv[2],sys.argv[3])
    elif cmd=="find": cmd_find(sys.argv[2],sys.argv[3],sys.argv[4],int(sys.argv[5]) if len(sys.argv)>5 else 8)
