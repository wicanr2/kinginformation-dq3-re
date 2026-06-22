#!/usr/bin/env python3
"""注音姓名輸入子系統反組譯輔助(nameinput agent 用)。在 docker capstone 內跑。

用法:
  re_name_dis.py            # 印出注音輸入子系統各 handler 反組譯 + 版面 / 對照表
  re_name_dis.py layout     # 只印 D3TXT00 rec 451-456 的注音鍵盤版面與 cell→碼表
  re_name_dis.py 0xNNNN C   # 反組譯任意 file offset，C 條指令

子系統(seg0):
  sub_0854 角色建立 modal → sub_0d17 dispatcher
    sub_0dc8 功能列 / sub_0e8e 英數 / sub_0f5b 注音
    sub_23f7 grid 導航 / sub_2623 組字 / sub_2654 查國字
"""
import sys
import struct
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"
HDR = 0x1370
DSBASE = 0x16140
md = Cs(CS_ARCH_X86, CS_MODE_16)
d = open(EXE, "rb").read()


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


def codetab():
    """DS:0x2758 注音 cell→(group<<5)|code 對照表(45 格)。"""
    GRP = {0: "聲母", 1: "介音", 2: "韻母", 3: "聲調"}
    print("# 注音 cell→碼對照表 DS:0x2758（每格 1 byte = (group<<5)|code）:")
    for i in range(45):
        b = d[DSBASE + 0x2758 + i]
        if i >= 42:
            note = ("OK", "切英數", "離開")[i - 42]
            print("#   cell %2d: 0x%02x  -> %s" % (i, b, note))
        else:
            print("#   cell %2d: 0x%02x  -> %s code=%d" %
                  (i, b, GRP[b >> 5], b & 0x1f))
    print()


if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "layout":
        codetab()
    elif len(sys.argv) > 1 and sys.argv[1] != "all":
        dis(int(sys.argv[1], 16),
            int(sys.argv[2]) if len(sys.argv) > 2 else 60)
    else:
        codetab()
        for label, off, cnt in [
            ("ni_dispatch sub_0d17", 0x2087, 50),
            ("ni_count_name sub_0da4", 0x2114, 16),
            ("ni_fn_list sub_0dc8", 0x2138, 45),
            ("ni_redraw_cursor sub_0e55", 0x21c5, 30),
            ("ni_alnum sub_0e8e", 0x21fe, 60),
            ("ni_zhuyin sub_0f5b", 0x22cb, 95),
            ("ni_grid_nav sub_23f7", 0x23f7, 80),
            ("ni_compose_put sub_2623", 0x2623, 25),
            ("ni_compose_resolve sub_2654", 0x2654, 50),
        ]:
            dis(off, cnt, label)
