#!/usr/bin/env python3
"""在 DOS MZ code image 逐 byte 尋找含指定 immediate 的 16-bit x86 指令。

Ghidra 對 real-mode 跳表後 code 可能漏分析；此工具用 Capstone 每一個 file
offset 各解一條指令，不依賴既有函式邊界。輸出是候選清單，仍須以 caller、
鄰近控制流及原版實機交叉驗證。

用法（在 tools/dockrun_cap.sh 內）：
  tools/re_find_imm.py 0x79
"""

import struct
import sys

from capstone import CS_ARCH_X86, CS_MODE_16, Cs
from capstone.x86 import X86_OP_IMM


raw = open("assets_raw/DQ3.EXE", "rb").read()
hdr = struct.unpack_from("<H", raw, 8)[0] * 16
want = int(sys.argv[1], 0)
md = Cs(CS_ARCH_X86, CS_MODE_16)
md.detail = True

seen = set()
for off in range(hdr, len(raw)):
    ins = next(md.disasm(raw[off : off + 15], off, count=1), None)
    if ins is None or not any(op.type == X86_OP_IMM and op.imm == want for op in ins.operands):
        continue
    key = (ins.address, ins.size)
    if key in seen:
        continue
    seen.add(key)
    logical = ins.address - hdr
    print(
        f"f{ins.address:05x} L{logical:05x} "
        f"{ins.bytes.hex():<20} {ins.mnemonic:<7} {ins.op_str}"
    )
