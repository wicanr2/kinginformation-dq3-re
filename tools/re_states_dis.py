#!/usr/bin/env python3
"""狀態機 handler 反組譯輔助(states agent 用)。在 docker capstone 內跑。
用法: re_states_dis.py            # 反組譯 14-entry 跳表的 6 個 handler
      re_states_dis.py 0xNNNN C   # 反組譯任意 file offset，C 條指令
讀 DS:0x19 scancode 表與 DS:0x25 near 指標表並印出來。"""
import sys, struct
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

d = open("assets_raw/DQ3.EXE", "rb").read()
HDR = 0x1370
DSBASE = 0x16140
md = Cs(CS_ARCH_X86, CS_MODE_16)


def dis(off, cnt, label=""):
    print("==== %s file=0x%05x seg0=0x%04x ====" % (label, off, off - HDR))
    code = d[off:off + cnt * 8]
    n = 0
    for ins in md.disasm(code, off):
        print("  %06x: %-14s %s %s" % (ins.address, ins.bytes.hex(),
                                        ins.mnemonic, ins.op_str))
        n += 1
        if ins.mnemonic in ("ret", "retf"):
            print("  --- %s ---" % ins.mnemonic)
        if n >= cnt:
            break
    print()


def table():
    sc = []
    i = 0x19
    while d[DSBASE + i] != 0xff:
        sc.append((i - 0x19, d[DSBASE + i]))
        i += 1
    print("# scancode table @DS:0x19 (0xff terminated):")
    for idx, b in sc:
        w = struct.unpack_from("<H", d, DSBASE + 0x25 + 2 * idx)[0]
        fo = HDR + w if w else 0
        print("#   idx=%2d sc=0x%02x -> near off=0x%04x file=0x%05x" %
              (idx, b, w, fo))
    print("# terminator 0xff at idx %d\n" % (i - 0x19))
    return sc


# --- extra batch invoked via: re_states_dis.py extra ---
def extra():
    for label, off, cnt in [
        ("subcmd5_504e", 0x63be, 60),
        ("subcmd6_5112", 0x6482, 10),
        ("subcmd8_512f", 0x649f, 50),
        ("subcmd11_51e9", 0x6559, 50),
        ("menu_helper_10900(file)", 0x900 + 0x1370, 40),
        ("menu_helper_10ae9(file)", 0xae9 + 0x1370, 40),
        ("F5_ec59 (file 0xffc9)", 0xffc9, 40),
    ]:
        dis(off, cnt, label)


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "extra":
        extra()
    elif len(sys.argv) > 1:
        dis(int(sys.argv[1], 16), int(sys.argv[2]) if len(sys.argv) > 2 else 60)
    else:
        table()
        for label, off, cnt in [
            ("h_9842 (sc 0x10/0x1f/0x3b)", 0xabb2, 90),
            ("h_7c83 (sc 0x39 Space)", 0x8ff3, 70),
            ("h_7c43 (sc 0x1c Enter)", 0x8fb3, 70),
            ("h_13a9 (sc 0x3f F5)", 0x2719, 80),
            ("h_14da (sc 0x40 F6)", 0x284a, 80),
            ("h_e915 (sc 0x43 F9)", 0xfc85, 90),
        ]:
            dis(off, cnt, label)
