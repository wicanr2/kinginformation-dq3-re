#!/usr/bin/env python3
"""掃全部 CTY 的 sub2(scripted-event)NPC,交叉比對 sub2 跳表 handler 的 GIVE/TAKE/flag。
section 偏移表無 count 前綴(對齊 dq3_town.c)。handler 解碼同 gen_sub2_full。
輸出每個實際存在的 sub2 NPC:CTY/section/座標 + byte4 + handler 行為摘要。
用法: tools/dockrun.sh tools/sweep_sub2_npcs.py
"""
import struct, glob, os
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"; DG = 0x16140; HDR = 0x1370; TBL = 0x3bb4
exe = open(EXE, "rb").read()
md = Cs(CS_ARCH_X86, CS_MODE_16)
hw = lambda off: struct.unpack_from("<H", exe, DG + off)[0]


def decode_handler(n):
    h = hw(TBL + n * 2)
    if h == 0 or h == 0xffff:
        return None
    fl = h + HDR
    out = {"give": [], "take": [], "tflag": [], "sflag": [], "recs": [], "region": [], "warp": False}
    last_bx = None; pend = None
    for ins in md.disasm(exe[fl:fl+0x140], fl):
        m, o = ins.mnemonic, ins.op_str
        if m == "mov" and o.startswith("bx, 0x"):
            try: last_bx = int(o.split(",")[1], 0)
            except: last_bx = None
        elif m == "call":
            if "0x8279" in o and last_bx is not None: out["tflag"].append(last_bx)
            elif "0x8264" in o and last_bx is not None: out["sflag"].append(last_bx)
            elif "0xd1f9" in o: out["warp"] = True
            elif "0x7bbe" in o and pend is not None: out["give"].append(pend); pend = None
            elif "0x7c0c" in o and pend is not None: out["take"].append(pend); pend = None
        elif m == "mov" and o.startswith("di, 0x"):
            v = int(o.split(",")[1], 0)
            if 0xbb8 <= v < 0xd00: out["recs"].append(v - 0xbb8)
        elif m == "mov" and "[0x2593]" in o.split(",")[0]:
            try: pend = int(o.split(",")[1], 0)
            except: pend = None
        elif m == "cmp" and "[0x722]" in o:
            rv = o.split(",")[-1].strip()
            if rv.startswith("0x") or rv.isdigit(): out["region"].append(int(rv, 0))
        # 單區塊邊界:ret 或 jmp 0x6380(返回派發 trampoline)。掃過界會混入下個 handler。
        if m in ("ret", "retf"): break
        if m == "jmp":
            try:
                if int(o, 0) == 0x6380: break
            except: pass
    return out


def npcs_of(path):
    d = open(path, "rb").read()
    u16 = lambda o: struct.unpack_from("<H", d, o)[0] if o + 2 <= len(d) else 0xffff
    res = []
    for si in range(64):
        so = u16(2 * si)
        if so == 0xffff:
            continue
        if so + 0x16 > len(d):
            break
        np = u16(so)
        if np == 0xffff: np = u16(so + 2)
        if np == 0xffff or so + np >= len(d): continue
        base = so + np; cnt = d[base]
        for i in range(cnt):
            r = base + 1 + i * 7
            if r + 7 > len(d): break
            x, y, b2, ctrl, b4, fl, b7 = d[r:r+7]
            if x >= 64 or y >= 64: continue          # 越界=雜訊
            if ((ctrl >> 3) & 7) == 2:
                res.append((si, x, y, b4))
    return res


seen = set()
print("CTY | sect | (x,y) | byte4 | GIVE | TAKE | tflag | sflag | region | warp | recs")
print("|---|---|---|---|---|---|---|---|---|---|---|")
for path in sorted(glob.glob("assets_raw/CTY*.DAT")):
    cty = os.path.basename(path).replace("CTY", "").replace(".DAT", "").lstrip("0") or "0"
    for si, x, y, b4 in npcs_of(path):
        key = (cty, si, x, y, b4)
        if key in seen: continue
        seen.add(key)
        hd = decode_handler(b4)
        if hd is None:
            g = t = tf = sf = rg = "-"; w = "-"; rc = "-"
        else:
            f = lambda l: ",".join("0x%02x" % v for v in l) if l else "-"
            g = f(hd["give"]); t = f(hd["take"]); tf = f(hd["tflag"]); sf = f(hd["sflag"])
            rg = f(hd["region"]); w = "Y" if hd["warp"] else "-"; rc = ",".join(map(str, hd["recs"])) or "-"
        print("| %s | %d | (%d,%d) | %d | %s | %s | %s | %s | %s | %s | %s |" % (cty, si, x, y, b4, g, t, tf, sf, rg, w, rc))
