#!/usr/bin/env python3
"""全檔線性反組譯 → stdout(file offset)。在 docker 內跑。"""
import struct
from capstone import Cs, CS_ARCH_X86, CS_MODE_16
d=open("assets_raw/DQ3.EXE","rb").read()
e_cparhdr=struct.unpack_from('<H',d,8)[0]
HDR=e_cparhdr*16
md=Cs(CS_ARCH_X86, CS_MODE_16)
# 反組譯 code 區(header 後到檔尾)
code=d[HDR:]
for ins in md.disasm(code, HDR):
    print(f"{ins.address:06x}: {ins.mnemonic} {ins.op_str}")
