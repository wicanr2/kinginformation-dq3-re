#!/usr/bin/env python3
"""忠實解每個 sub2 handler 的完整結構(雙區塊:flag-not-set 與 give 分支)。
模式(實證 byte4=7/12/31/49/52):
  call 0x6372(preamble)
  [mov di,recA; lcall]            ; always_rec(進門先顯示;可無)
  mov bx,FLAG; call 0x8279(test)  ; prereq_flag
  cmp al,1; je/jne GIVE           ; 條件分支
  mov di,recB; lcall; jmp 0x6380  ; before(條件未滿足)
GIVE:
  mov di,recC; lcall              ; give_rec
  [mov [0x2593],item; call 0x7bbe/0x7c0c]  ; give/take
  [bx=FLAG2; call 0x8264]         ; set_flag
輸出 markdown 表:byte4 | handler | always_rec | prereq_flag(je極性)| before_rec | give_rec | give/take item | set_flag | region | warp
用法:capstone docker。
"""
import struct, sys
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"; DG = 0x16140; HDR = 0x1370; TBL = 0x3bb4
exe = open(EXE, "rb").read()
md = Cs(CS_ARCH_X86, CS_MODE_16)
hw = lambda off: struct.unpack_from("<H", exe, DG + off)[0]


def lin_block(start, limit=0x80):
    """線性解一段到 jmp 0x6380 / ret。回 (insns, end_targets)。"""
    out = []
    for ins in md.disasm(exe[start:start+limit], start):
        out.append(ins)
        if ins.mnemonic in ("ret", "retf"):
            break
        if ins.mnemonic == "jmp":
            try:
                if int(ins.op_str, 0) == 0x6380:
                    break
            except: pass
    return out


def scan(insns):
    """從一串指令抽 give/take/setflag/recs。"""
    r = {"give": None, "take": None, "sflag": [], "recs": []}
    last_bx = None; pend = None
    for ins in insns:
        m, o = ins.mnemonic, ins.op_str
        if m == "mov" and o.startswith("bx, 0x"):
            try: last_bx = int(o.split(",")[1], 0)
            except: last_bx = None
        elif m == "mov" and "[0x2593]" in o.split(",")[0]:
            try: pend = int(o.split(",")[1], 0)
            except: pend = None
        elif m == "mov" and o.startswith("di, 0x"):
            v = int(o.split(",")[1], 0)
            if 0xbb8 <= v < 0xd00: r["recs"].append(v - 0xbb8)
        elif m == "call":
            try: t = int(o, 0)
            except: continue
            if t == 0x7bbe and pend is not None: r["give"] = pend; pend = None
            elif t == 0x7c0c and pend is not None: r["take"] = pend; pend = None
            elif t == 0x8264 and last_bx is not None: r["sflag"].append(last_bx)
    return r


def decode(n):
    h = hw(TBL + n * 2)
    if h == 0 or h == 0xffff: return None
    fl = h + HDR
    first = lin_block(fl)
    # 找 always_rec(test_flag 前的 di rec)、prereq_flag、條件分支目標
    always_rec = None; prereq = None; je_target = None; je_kind = None; region = None
    seen_test = False
    for ins in first:
        m, o = ins.mnemonic, ins.op_str
        if m == "mov" and o.startswith("di, 0x") and not seen_test:
            v = int(o.split(",")[1], 0)
            if 0xbb8 <= v < 0xd00: always_rec = v - 0xbb8
        if m == "mov" and o.startswith("bx, 0x"):
            try: prereq = int(o.split(",")[1], 0)
            except: pass
        if m == "call" and o == "0x8279": seen_test = True
        if m == "cmp" and "[0x722]" in o:
            rv = o.split(",")[-1].strip()
            if rv.startswith("0x") or rv.isdigit(): region = int(rv, 0)
        if m in ("je", "jne") and seen_test and je_target is None:
            try: je_target = int(o, 0); je_kind = m
            except: pass
    before = scan(first)
    give_branch = scan(lin_block(je_target)) if je_target else {"give": None, "take": None, "sflag": [], "recs": []}
    return {
        "h": h, "always_rec": always_rec, "prereq_flag": prereq, "je_kind": je_kind,
        "before_recs": [x for x in before["recs"] if x != always_rec],
        "give": give_branch["give"], "take": give_branch["take"],
        "give_recs": give_branch["recs"], "set_flag": give_branch["sflag"],
        "region": region,
        "first_give": before["give"], "first_take": before["take"],  # 無分支型(52)直接給
    }


targets = [int(a) for a in sys.argv[1:]] if len(sys.argv) > 1 else range(0, 90)
print("byte4 | hdr | always_rec | prereq_flag/je | before_rec | give_rec | GIVE | TAKE | set_flag | region")
print("|---|---|---|---|---|---|---|---|---|---|")
for n in targets:
    d = decode(n)
    if d is None: continue
    g = d["give"] if d["give"] is not None else d["first_give"]
    t = d["take"] if d["take"] is not None else d["first_take"]
    # 只印有意義(有 give/take/prereq 的給予型)
    if g is None and t is None:
        continue
    pf = ("%s0x%02x" % ("!" if d["je_kind"] == "jne" else "", d["prereq_flag"])) if d["prereq_flag"] is not None else "-"
    sf = ",".join("0x%02x" % x for x in d["set_flag"]) or "-"
    print("| %d | L0x%04x | %s | %s | %s | %s | %s | %s | %s | %s |" % (
        n, d["h"],
        d["always_rec"] if d["always_rec"] is not None else "-",
        pf,
        ",".join(map(str, d["before_recs"])) or "-",
        ",".join(map(str, d["give_recs"])) or "-",
        "0x%02x" % g if g is not None else "-",
        "0x%02x" % t if t is not None else "-",
        sf,
        "0x%02x" % d["region"] if d["region"] is not None else "-"))
