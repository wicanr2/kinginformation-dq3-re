#!/usr/bin/env python3
"""反組譯 DQ3.EXE 指定 file-offset 區段,雙標 file/logical(logical=file-0x1370)。
用法: dis_region.py <start_hex> <len_hex>  (start 為 file offset)"""
import sys, struct
from capstone import Cs, CS_ARCH_X86, CS_MODE_16
d = open("assets_raw/DQ3.EXE", "rb").read()
HDR = struct.unpack_from('<H', d, 8)[0] * 16   # e_cparhdr*16
start = int(sys.argv[1], 16)
length = int(sys.argv[2], 16)
md = Cs(CS_ARCH_X86, CS_MODE_16)
md.detail = False
for ins in md.disasm(d[start:start+length], start):
    logical = ins.address - 0x1370
    print(f"f{ins.address:05x} L{logical:05x}: {ins.mnemonic:<7} {ins.op_str}")
