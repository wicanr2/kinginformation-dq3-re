#!/usr/bin/env python3
"""DQ3.EXE code cave 探勘 + 攻擊/傷害 handler 反組譯輔助 (capstone, docker 內跑)。
本檔專供 docs/22 道具特效修正 (#7),不動他人 tools。

用法:
  re_codecave_scan.py dis <file_off_hex> [count]   反組譯 (file offset)
  re_codecave_scan.py seg <seg0off_hex> [count]    反組譯 seg0 邏輯 off (+0x1370)
  re_codecave_scan.py bytes <file_off_hex> [n]     dump raw bytes
  re_codecave_scan.py caves                         掃 code 段 NOP/重複 padding
  re_codecave_scan.py find <hexpat>                 byte pattern 搜尋 (?? 萬用)
"""
import sys, struct, re
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"
d = open(EXE, "rb").read()
HDR = struct.unpack_from('<H', d, 8)[0] * 16
md = Cs(CS_ARCH_X86, CS_MODE_16)


def dis(off, cnt):
    for n, ins in enumerate(md.disasm(d[off:off + cnt * 8], off)):
        if n >= cnt:
            break
        print(f"  f{ins.address:06x} s{(ins.address-HDR)&0xffff:04x}: {ins.bytes.hex():<18} {ins.mnemonic} {ins.op_str}")


def caves():
    # NOP runs and identical-byte runs in resident code region HDR..0x12000
    lo, hi = HDR, 0x12000
    print("== code 段 (f%06x..f%06x) padding 探勘 ==" % (lo, hi))
    for val in (0x90, 0x00):
        i = lo
        while i < hi:
            if d[i] == val:
                j = i
                while j < hi and d[j] == val:
                    j += 1
                if j - i >= 3:
                    print(f"  byte 0x{val:02x}: f{i:06x} s{(i-HDR)&0xffff:04x} len {j-i}")
                i = j
            else:
                i += 1


def main():
    c = sys.argv[1]
    if c == "dis":
        dis(int(sys.argv[2], 16), int(sys.argv[3]) if len(sys.argv) > 3 else 40)
    elif c == "seg":
        dis(HDR + int(sys.argv[2], 16), int(sys.argv[3]) if len(sys.argv) > 3 else 40)
    elif c == "bytes":
        off = int(sys.argv[2], 16)
        n = int(sys.argv[3]) if len(sys.argv) > 3 else 64
        for i in range(0, n, 16):
            row = d[off + i:off + i + 16]
            print(f"  f{off+i:06x}: " + " ".join(f"{b:02x}" for b in row))
    elif c == "caves":
        caves()
    elif c == "find":
        pat = sys.argv[2].replace(" ", "")
        rx = b"".join((b"." if pat[i:i+2] == "??" else re.escape(bytes([int(pat[i:i+2], 16)]))) for i in range(0, len(pat), 2))
        for m in re.finditer(rx, d, re.DOTALL):
            print(f"  file 0x{m.start():06x}  seg0 0x{(m.start()-HDR)&0xffff:04x}")


main()
