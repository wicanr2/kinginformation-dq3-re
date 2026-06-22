#!/usr/bin/env python3
"""DQ3.EXE seg0 遞迴反組譯(recursive-descent)+ 函式圖。

線性反組譯會對齊錯位(資料夾雜在碼裡)。本工具從 entry 與已知進入點開始,
跟著 call / jmp / jcc 遞迴切函式邊界,只反組譯真正可達的指令,輸出:
  - 函式起點 / 大小 / caller-callee
  - 對主資料段(DS=0x14dd)位址與檔名池(DS:0x3d..)的 xref
  - 已知函式標註

位址空間 = seg0 邏輯 offset(= file_offset - HDR)。entry seg0 off = CS:IP = 0x9299。
在 docker 內跑(capstone)。

用法:
  re_recdis.py            # 完整跑,印摘要
  re_recdis.py json OUT   # 把函式圖寫成 json
"""
import sys, struct, json, collections
from capstone import Cs, CS_ARCH_X86, CS_MODE_16
from capstone import x86

EXE = "assets_raw/DQ3.EXE"
d = open(EXE, "rb").read()
e_cparhdr = struct.unpack_from('<H', d, 8)[0]
e_ip = struct.unpack_from('<H', d, 20)[0]
e_cs = struct.unpack_from('<H', d, 22)[0]
HDR = e_cparhdr * 16            # 0x1370

# seg0 邏輯位址空間:off 0..0xffff 對映 file HDR+off。
# entry seg0 off = e_cs*16 + e_ip (e_cs=0) = e_ip = 0x9299
SEG0_LO = 0x0000
SEG0_HI = 0x10000              # seg0 視為前 64KB
def seg0_bytes():
    return d[HDR:HDR + SEG0_HI]
CODE = seg0_bytes()

md = Cs(CS_ARCH_X86, CS_MODE_16)
md.detail = True

# 已知進入點(seg0 off)與名稱
ENTRY = e_cs * 16 + e_ip       # 0x9299
KNOWN = {
    ENTRY:   "entry",
    0xa660:  "startup_a660",    # entry 鏈下一棒
    0x82bb:  "init_82bb",       # 啟動鏈大函式
    0xa753:  "post_init_a753",
    0xeaf4:  "load_blk_eaf4",   # BLK tile 載入器(file 0xfe64)
}

# 主資料段檔名池(DS:off -> 檔名),來自 docs/06
FNPOOL = {
    0x003d:"setup.dat",0x0047:"player.dat",0x0052:"dragon0.dat",0x005e:"cty00.dat",
    0x0068:"d3txt00.fon",0x0074:"d3txt00.txt",0x0080:"item.dat",0x0089:"dq3con.map",
    0x0094:"dq3und.map",0x009f:"dq31.blk",0x00a8:"dq3.blk",0x00b0:"blkbm1.dat",
    0x00bb:"blkbm.dat",0x00c5:"dq3mst.bls",0x00d0:"dq3man.bls",0x00db:"dq3lin.bls",
    0x00e6:"dq3.pal",0x00ee:"mnsbk.pal",0x00f8:"d3mns.dat",0x0102:"dq3mns.shp",
    0x010d:"packbg.scr",0x0118:"mbg.mcx",0x0120:"ebg.mcx",
}

# 終止指令(到此函式流結束):ret/retf/iret/jmp(非條件)後不繼續直線
TERM = {x86.X86_INS_RET, x86.X86_INS_RETF, x86.X86_INS_IRET, x86.X86_INS_IRETD}

def in_seg0(off):
    return SEG0_LO <= off < SEG0_HI

# 遞迴反組譯:func_starts 為待處理函式入口佇列
visited_ins = {}        # off -> capstone insn
func_of = {}            # off(指令) -> 所屬函式入口
calls = collections.defaultdict(set)     # caller_func -> set(callee_off)
callers = collections.defaultdict(set)   # callee_func -> set(caller_func)
func_set = set()        # 已確認函式入口
data_xref = collections.defaultdict(set) # func -> set(ds offsets referenced)
far_calls = collections.defaultdict(collections.Counter)  # func -> Counter((seg,off))
func_ins = collections.defaultdict(list) # func -> [ins off...] (sorted later)

def disasm_one(off):
    """反組譯 single insn at seg0 off。回傳 capstone insn 或 None。"""
    if off in visited_ins:
        return visited_ins[off]
    if not in_seg0(off):
        return None
    chunk = CODE[off:off + 16]
    g = md.disasm(chunk, off)
    try:
        ins = next(g)
    except StopIteration:
        return None
    visited_ins[off] = ins
    return ins

def extract_imm_targets(ins):
    """從指令運算元抽出 (near_call_targets, jmp_targets, far_call(seg,off)|None, ds_refs)。"""
    near_call = None
    jmp_tgts = []
    far = None
    ds_refs = []
    mn = ins.mnemonic
    # call near (相對) -> capstone 已算好絕對 off(因為我們用 off 當 base)
    if mn == 'call':
        # 取 imm 運算元
        for op in ins.operands:
            if op.type == x86.X86_OP_IMM:
                near_call = op.imm & 0xffff
    elif mn == 'lcall':
        # lcall seg, off
        ops = ins.op_str
        if ',' in ops:
            try:
                s, o = ops.split(',')
                far = (int(s.strip(), 16), int(o.strip(), 16))
            except Exception:
                pass
    elif mn == 'jmp':
        for op in ins.operands:
            if op.type == x86.X86_OP_IMM:
                jmp_tgts.append(op.imm & 0xffff)
    elif mn.startswith('j'):   # jcc
        for op in ins.operands:
            if op.type == x86.X86_OP_IMM:
                jmp_tgts.append(op.imm & 0xffff)
    # DS 直接定址 [0xNNNN] 抓取(mem op,base/index 無,disp 在 ds 範圍)
    for op in ins.operands:
        if op.type == x86.X86_OP_MEM:
            m = op.mem
            if m.base == 0 and m.index == 0:
                disp = m.disp & 0xffff
                ds_refs.append(disp)
    return near_call, jmp_tgts, far, ds_refs

def walk_func(entry):
    """以 entry 為入口,線性+分支遞迴反組譯函式體,直到所有路徑遇 ret/jmp 終止。"""
    if entry in func_set:
        return
    func_set.add(entry)
    worklist = [entry]
    seen = set()
    while worklist:
        off = worklist.pop()
        while True:
            if off in seen or not in_seg0(off):
                break
            seen.add(off)
            ins = disasm_one(off)
            if ins is None:
                break
            # 該指令歸屬此函式(第一個碰到的函式擁有它)
            if off not in func_of:
                func_of[off] = entry
                func_ins[entry].append(off)
            near, jmps, far, ds_refs = extract_imm_targets(ins)
            for r in ds_refs:
                if r < 0x4000:    # 資料段前段(檔名池/旗標/座標)較有意義
                    data_xref[entry].add(r)
            if far is not None:
                far_calls[entry][far] += 1
            if near is not None and in_seg0(near):
                calls[entry].add(near)
                callers[near].add(entry)
                new_funcs.add(near)     # call 目標 = 新函式候選
            mn = ins.mnemonic
            nxt = off + ins.size
            if mn == 'jmp':
                # 非條件跳:跟過去(同函式 tail),不再直線
                for t in jmps:
                    if in_seg0(t):
                        worklist.append(t)
                break
            if mn.startswith('j') and mn != 'jmp':   # jcc:兩路
                for t in jmps:
                    if in_seg0(t):
                        worklist.append(t)
                off = nxt
                continue
            if ins.id in TERM:
                break
            # call / 一般指令:直線繼續
            off = nxt

# BFS:從已知入口擴展函式
new_funcs = set(KNOWN.keys())
processed = set()
rounds = 0
while new_funcs:
    rounds += 1
    batch = list(new_funcs)
    new_funcs = set()
    for f in batch:
        if f in processed or not in_seg0(f):
            continue
        processed.add(f)
        walk_func(f)
    if rounds > 5000:
        break

# 計算函式大小(該函式擁有的指令 span)
func_info = {}
for f in func_set:
    offs = func_ins.get(f, [])
    if not offs:
        continue
    lo = min(offs)
    # 函式末端 = 最後一指令 off + size
    last = max(offs)
    li = visited_ins.get(last)
    hi = last + (li.size if li else 0)
    func_info[f] = {
        "entry": f,
        "lo": lo, "hi": hi, "size": hi - lo,
        "n_ins": len(offs),
        "name": KNOWN.get(f, ""),
        "callees": sorted(calls.get(f, [])),
        "callers": sorted(callers.get(f, [])),
        "far": [[s, o, c] for (s, o), c in far_calls.get(f, {}).items()],
        "ds_xref": sorted(data_xref.get(f, [])),
        "fnpool": sorted(r for r in data_xref.get(f, []) if r in FNPOOL),
    }

def fname(f):
    n = KNOWN.get(f, "")
    return f"sub_{f:04x}" + (f"({n})" if n else "")

if len(sys.argv) > 1 and sys.argv[1] == "json":
    out = sys.argv[2] if len(sys.argv) > 2 else "docs/data/exe_funcs.json"
    json.dump({"hdr": HDR, "entry": ENTRY,
               "funcs": sorted(func_info.values(), key=lambda x: x["lo"])},
              open(out, "w"), indent=1)
    print(f"# wrote {out}: {len(func_info)} funcs")
    sys.exit(0)

# 摘要輸出
print(f"# 遞迴反組譯 seg0,entry off=0x{ENTRY:04x},HDR=0x{HDR:x}")
print(f"# 函式數={len(func_info)}  指令數={len(visited_ins)}  涵蓋 bytes≈{sum(fi['size'] for fi in func_info.values())}")

# 依被呼叫次數(in-degree)排熱點函式
hot = sorted(func_info.values(), key=lambda x: len(x["callers"]), reverse=True)
print("\n## 被呼叫最多的函式(in-degree top 25):")
for fi in hot[:25]:
    nm = fi["name"]
    fp = (" FILES:" + ",".join(FNPOOL[r] for r in fi["fnpool"])) if fi["fnpool"] else ""
    print(f"  {fname(fi['entry']):28s} size={fi['size']:5d} ins={fi['n_ins']:4d} "
          f"callers={len(fi['callers']):3d} callees={len(fi['callees']):3d}{fp}")

# 最大函式
print("\n## 最大函式(by size top 15):")
for fi in sorted(func_info.values(), key=lambda x: x["size"], reverse=True)[:15]:
    print(f"  {fname(fi['entry']):28s} size={fi['size']:5d} ins={fi['n_ins']:4d} "
          f"far_calls={sum(c for *_, c in fi['far'])}")

# 引用檔名池的函式 = 素材載入相關
print("\n## 引用主資料段檔名池的函式(素材載入相關):")
for fi in sorted(func_info.values(), key=lambda x: x["lo"]):
    if fi["fnpool"]:
        print(f"  {fname(fi['entry']):20s} -> " +
              ", ".join(f"{FNPOOL[r]}@{r:#05x}" for r in fi["fnpool"]))

# 啟動鏈追蹤
print("\n## 啟動鏈(entry 可達的呼叫樹,深度 3):")
def show_chain(f, depth, seen, pref=""):
    if depth < 0 or f in seen:
        return
    seen.add(f)
    fi = func_info.get(f)
    if not fi:
        return
    print(f"  {pref}{fname(f)} (size={fi['size']}, callees={len(fi['callees'])})")
    for c in fi["callees"][:8]:
        show_chain(c, depth - 1, seen, pref + "  ")
show_chain(ENTRY, 3, set())
