#!/usr/bin/env python3
"""完整反組譯指定 byte4 的 sub2 handler(跳表 0x3bb4),標註 flag/item/rec/warp 語意。
用法: tools/dockrun(capstone) tools/dis_sub2_handler.py 49 [25 48 ...]
file = logical + 0x1370;DGROUP = 0x16140;rec = di - 0xbb8。
"""
import struct, sys
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"; DG = 0x16140; HDR = 0x1370; TBL = 0x3bb4
exe = open(EXE, "rb").read()
md = Cs(CS_ARCH_X86, CS_MODE_16); md.detail = False
hw = lambda off: struct.unpack_from("<H", exe, DG + off)[0]

CALLS = {0x8279: "test_flag(bx)->al", 0x8264: "set/clr_flag(bx)", 0x7bbe: "GIVE item[0x2593]",
         0x7c0c: "TAKE/has item[0x2593]->?", 0xd1f9: "WARP", 0x82a0: "?msg", 0x7d4e: "?"}

for arg in sys.argv[1:]:
    n = int(arg)
    h = hw(TBL + n * 2)
    fl = h + HDR
    print("\n==== byte4=%d  handler L0x%04x (file 0x%05x) ====" % (n, h, fl))
    last_bx = None; pend = None
    for ins in md.disasm(exe[fl:fl+0x180], fl):
        m, o = ins.mnemonic, ins.op_str
        note = ""
        if m == "mov" and o.startswith("bx, 0x"):
            try: last_bx = int(o.split(",")[1], 0); note = "; bx=flag 0x%02x" % last_bx
            except: pass
        elif m == "mov" and "[0x2593]" in o.split(",")[0]:
            try: pend = int(o.split(",")[1], 0); note = "; item=0x%02x" % pend
            except: pass
        elif m == "mov" and o.startswith("di, 0x"):
            v = int(o.split(",")[1], 0)
            if 0xbb8 <= v < 0xd00: note = "; rec %d" % (v - 0xbb8)
        elif m == "call":
            tgt = o.replace("0x", "")
            try: t = int(o, 0); note = "; " + CALLS.get(t, "call 0x%04x" % t)
            except: note = ""
        elif m == "cmp" and "[0x722]" in o:
            note = "; region check"
        elif m in ("je", "jne", "jz", "jnz", "jmp", "ja", "jb", "jbe", "jae"):
            note = "  <branch>"
        print("  0x%05x: %-8s %-24s%s" % (ins.address, m, o, note))
        # handler 單區塊邊界:jmp 0x6380(返回派發 trampoline)或 ret
        if m in ("ret", "retf"):
            break
        if m == "jmp":
            try:
                if int(o, 0) == 0x6380:
                    print("  --- block end (jmp dispatcher 0x6380) ---")
                    break
            except: pass
