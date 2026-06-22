#!/usr/bin/env python3
"""DQ3.EXE 16-bit 反組譯偵察(capstone)。在 docker 內跑。
用法: re_disasm.py <file_offset_hex> [count]
也可: re_disasm.py entry   # 反組譯 entry point
"""
import sys, struct
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

d=open("assets_raw/DQ3.EXE","rb").read()
e_cparhdr=struct.unpack_from('<H',d,8)[0]
e_ip=struct.unpack_from('<H',d,20)[0]
e_cs=struct.unpack_from('<H',d,22)[0]
HDR=e_cparhdr*16
arg=sys.argv[1] if len(sys.argv)>1 else "entry"
if arg=="entry":
    off=HDR + e_cs*16 + e_ip
    print(f"# entry CS:IP={e_cs:04x}:{e_ip:04x}  load_base(file)=0x{HDR:x}  entry file off=0x{off:x}")
else:
    off=int(arg,16)
cnt=int(sys.argv[2]) if len(sys.argv)>2 else 60

md=Cs(CS_ARCH_X86, CS_MODE_16)
code=d[off:off+cnt*8]
n=0
for ins in md.disasm(code, off):
    print(f"  {ins.address:06x}: {ins.bytes.hex():<14} {ins.mnemonic} {ins.op_str}")
    n+=1
    if n>=cnt: break
