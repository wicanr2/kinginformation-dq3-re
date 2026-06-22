#!/usr/bin/env python3
"""DQ3.EXE bug 分析用反組譯 / 搜尋輔助(capstone, docker 內跑)。

不污染他人 tools:本檔專供 docs/18 bug 分析。

用法:
  re_bug_dis.py dis <off_hex> [count]   反組譯(off 為 file offset;若帶 's:' 前綴則為 seg0 off)
  re_bug_dis.py seg <seg0off_hex> [n]   反組譯 seg0 邏輯 offset(自動 +0x1370)
  re_bug_dis.py bytes <off_hex> [n]     dump raw bytes
  re_bug_dis.py find <hexpat>           在整檔搜 byte pattern(?? 萬用), 回 file off + seg0 off
  re_bug_dis.py xref <imm_hex>          搜尋以該 16-bit 立即數為記憶體 disp 的指令(粗略)
"""
import sys, struct, re
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"
d = open(EXE, "rb").read()
e_cparhdr = struct.unpack_from('<H', d, 8)[0]
HDR = e_cparhdr * 16          # 0x1370
md = Cs(CS_ARCH_X86, CS_MODE_16)
md.detail = False


def seg0(fileoff):
    return fileoff - HDR


def dis(off, cnt):
    code = d[off:off + cnt * 8]
    n = 0
    for ins in md.disasm(code, off):
        s0 = ins.address - HDR
        print(f"  f{ins.address:06x} s{s0 & 0xffff:04x}: {ins.bytes.hex():<16} {ins.mnemonic} {ins.op_str}")
        n += 1
        if n >= cnt:
            break


def main():
    if len(sys.argv) < 2:
        print(__doc__); return
    cmd = sys.argv[1]
    if cmd == "dis":
        a = sys.argv[2]
        off = HDR + int(a[2:], 16) if a.startswith("s:") else int(a, 16)
        dis(off, int(sys.argv[3]) if len(sys.argv) > 3 else 60)
    elif cmd == "seg":
        off = HDR + int(sys.argv[2], 16)
        dis(off, int(sys.argv[3]) if len(sys.argv) > 3 else 60)
    elif cmd == "bytes":
        off = int(sys.argv[2], 16)
        n = int(sys.argv[3]) if len(sys.argv) > 3 else 64
        chunk = d[off:off + n]
        for i in range(0, len(chunk), 16):
            row = chunk[i:i + 16]
            print(f"  f{off+i:06x}: " + " ".join(f"{b:02x}" for b in row))
    elif cmd == "find":
        pat = sys.argv[2].replace(" ", "")
        rx = b""
        for i in range(0, len(pat), 2):
            t = pat[i:i+2]
            rx += b"." if t == "??" else re.escape(bytes([int(t, 16)]))
        for m in re.finditer(rx, d, re.DOTALL):
            o = m.start()
            print(f"  file 0x{o:06x}  seg0 0x{seg0(o)&0xffff:04x}")
    elif cmd == "xref":
        imm = int(sys.argv[2], 16)
        lo, hi = imm & 0xff, (imm >> 8) & 0xff
        # search 16-bit LE immediate appearing in code region
        needle = bytes([lo, hi])
        cnt = 0
        for m in re.finditer(re.escape(needle), d[HDR:HDR+0x10000], re.DOTALL):
            o = HDR + m.start()
            print(f"  file 0x{o:06x}  seg0 0x{seg0(o)&0xffff:04x}")
            cnt += 1
            if cnt > 60:
                print("  ... (truncated)"); break
    else:
        print(__doc__)


if __name__ == "__main__":
    main()
