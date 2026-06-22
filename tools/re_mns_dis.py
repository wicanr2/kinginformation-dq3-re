#!/usr/bin/env python3
"""DQ3.EXE 怪物 sprite 載入/解碼路徑反組譯偵察(capstone)。在 docker 內跑。
用法: re_mns_dis.py 0xNNNN n   # seg0 off 反組譯 n 條
       re_mns_dis.py b19e        # sub_b19e shp_load_sprite
"""
import sys, struct
from capstone import Cs, CS_ARCH_X86, CS_MODE_16
D = open("assets_raw/DQ3.EXE", "rb").read()
e_cparhdr = struct.unpack_from('<H', D, 8)[0]; HDR = e_cparhdr * 16
md = Cs(CS_ARCH_X86, CS_MODE_16)

def dis(off, cnt, label=""):
    print(f"\n;==== {label} seg0:0x{off:04x} (file 0x{HDR+off:x}) ====")
    code = D[HDR+off:HDR+off+cnt*8]
    n = 0
    for ins in md.disasm(code, off):
        print(f"  {ins.address:06x}: {ins.bytes.hex():<14} {ins.mnemonic} {ins.op_str}")
        n += 1
        if n >= cnt: break

if len(sys.argv) > 1 and sys.argv[1] == "b19e":
    dis(0xb19e, 160, "sub_b19e shp_load_sprite")
elif len(sys.argv) > 2:
    dis(int(sys.argv[1], 16), int(sys.argv[2]), "manual")
else:
    dis(0xb19e, 160, "sub_b19e")
