#!/usr/bin/env python3
"""野外指令子系統 + 對話流程反組譯輔助(commands agent 用)。docker capstone 內跑。

涵蓋:
  - DS:0x3baa 12 項指令選單二級派發表(field command dispatch)
  - 各 handler(sub_988d..sub_51e9)與其呼叫的子函式(0x16dd / 0x1a4c / 0x7c50)
  - 對話流程 worker chain:sub_9cd6(Enter worker)/ sub_60c6 / sub_5b2d
  - 文字繪製器 111b:0264 入口取碼迴圈(對照 docs/03 控制碼)

用法:
  re_cmd_dis.py table          # 印 12 項派發表 + scancode 表
  re_cmd_dis.py all            # 印所有重點 handler / worker 反組譯
  re_cmd_dis.py 0xNNNN [C]     # 任意 file offset 反組譯 C 條
"""
import sys
import struct
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
        print("  %06x: %-16s %s %s" % (ins.address, ins.bytes.hex(),
                                       ins.mnemonic, ins.op_str))
        n += 1
        if ins.mnemonic in ("ret", "retf"):
            print("  --- %s ---" % ins.mnemonic)
        if n >= cnt:
            break
    print()


def table():
    print("# DS:0x3baa 12-entry field command dispatch (near ptr, [0x722]-1 index):")
    for i in range(12):
        w = struct.unpack_from("<H", d, DSBASE + 0x3baa + 2 * i)[0]
        print("#   idx=%2d near=0x%04x file=0x%05x" % (i, w, HDR + w if w else 0))
    print()
    print("# DS:0x19 scancode -> DS:0x25 handler (11 + 0xff term):")
    for i in range(12):
        sc = d[DSBASE + 0x19 + i]
        w = struct.unpack_from("<H", d, DSBASE + 0x25 + 2 * i)[0]
        print("#   idx=%2d sc=0x%02x near=0x%04x file=0x%05x" %
              (i, sc, w, HDR + w if w else 0))
        if sc == 0xff:
            break
    print()


# file offset = HDR + seg0_off
def fo(seg0):
    return HDR + seg0


TARGETS = [
    # field command dispatch handlers
    ("disp0  sub_988d", fo(0x988d), 24),
    ("disp1  sub_98b9", fo(0x98b9), 20),
    ("disp2  sub_98cd", fo(0x98cd), 16),
    ("disp3  sub_98e6", fo(0x98e6), 20),
    ("disp4  sub_9900", fo(0x9900), 30),
    ("disp5  sub_504e item/equip list", fo(0x504e), 120),
    ("disp6  sub_5112 -> call 0x16dd", fo(0x5112), 6),
    ("disp7  sub_5116", fo(0x5116), 30),
    ("disp8  sub_512f -> call 0x1a4c", fo(0x512f), 6),
    ("disp9  sub_5133 -> call 0x7c50", fo(0x5133), 6),
    ("disp10 sub_5137", fo(0x5137), 90),
    ("disp11 sub_51e9", fo(0x51e9), 90),
    # sub-call targets (unnamed)
    ("sub_16dd (disp6 target)", fo(0x16dd), 80),
    ("sub_1a4c (disp8 target)", fo(0x1a4c), 80),
    ("sub_7c50 (disp9 target)", fo(0x7c50), 60),
    # dialogue worker chain
    ("sub_9cd6 Enter worker", fo(0x9cd6), 140),
    ("sub_60c6 Enter stage2", fo(0x60c6), 120),
    ("sub_5b2d Enter stage3", fo(0x5b2d), 120),
    # menu helpers
    ("sub_10900 win_open", fo(0x900), 40),
    ("sub_10ae9 win_nav", fo(0xae9), 60),
    ("sub_10974 win_close", fo(0x974), 30),
    # flag get/set
    ("sub_8279 flag_get", fo(0x8279), 30),
    ("sub_8264 flag_set", fo(0x8264), 30),
    # text drawer entry (seg 0x111b off 0x264)
    ("text_draw 111b:0264", 0x12784, 120),
]


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "table":
        table()
    elif len(sys.argv) > 1 and sys.argv[1] == "all":
        table()
        for label, off, cnt in TARGETS:
            dis(off, cnt, label)
    elif len(sys.argv) > 1:
        dis(int(sys.argv[1], 16), int(sys.argv[2]) if len(sys.argv) > 2 else 60)
    else:
        table()
