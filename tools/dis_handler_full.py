#!/usr/bin/env python3
"""完整反組譯一個 sub2 handler,跟進所有分支(je/jne/jmp 區域內),標出戰鬥啟動線索。

dis_sub2_handler.py 在第一個 je/ret 就停(單區塊),會漏掉 je-target 之後的 give/battle 分支。
本工具用 worklist 跟進函式內分支(同段近距離 jmp/je/jne target),標註:
  - [0x2321]/[0x231f] 敵群表寫入(boss sprite id / 隻數)— 戰鬥啟動關鍵
  - call 0xbfd1(battle_enter_screen,sub_bfd1)
  - 各種 call 語意(test/set flag, give/take item, warp)
file = logical + 0x1370;DGROUP = 0x16140;rec = di - 0xbb8;跳表 0x3bb4。

用法: tools/dis_handler_full.py 58 [起始上限位元組數]
"""
import struct, sys
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"
DG = 0x16140; HDR = 0x1370; TBL = 0x3bb4
exe = open(EXE, "rb").read()
md = Cs(CS_ARCH_X86, CS_MODE_16); md.detail = True
hw = lambda off: struct.unpack_from("<H", exe, DG + off)[0]

CALLS = {
    0x8279: "test_flag(bx)->al", 0x8264: "set/clr_flag(bx)",
    0x7bbe: "GIVE item[0x2593]", 0x7c0c: "TAKE/has item->[0x726]",
    0xd1f9: "WARP", 0xbfd1: "★battle_enter_screen(sub_bfd1)",
    0x6372: "?prep", 0x09ac: "?msg/anim",
}

byte4 = int(sys.argv[1])
limit = int(sys.argv[2]) if len(sys.argv) > 2 else 400
log_off = hw(TBL + byte4 * 2)
start = log_off + HDR
print("==== byte4=%d  handler L0x%04x (file 0x%05x) ====" % (byte4, log_off, start))

seen = set()
worklist = [start]
# 反組譯整段範圍(start .. start+limit),線性印,但標出分支 target 供人追
end = start + limit
code = exe[start:end]
targets = set()
for ins in md.disasm(code, log_off):
    fileoff = ins.address + HDR
    mn, op = ins.mnemonic, ins.op_str
    note = ""
    # call 語意
    if mn == "call":
        try:
            tgt = int(op, 0)
            if tgt in CALLS:
                note = "; " + CALLS[tgt]
            else:
                note = "; call 0x%x" % tgt
        except ValueError:
            note = "; call %s" % op
    elif mn == "lcall":
        note = "; lcall %s" % op
    # 敵群表寫入(boss 戰關鍵)
    if "0x2321" in op:
        note += "  ◀◀ 敵群 sprite id 表 [0x2321]"
    if "0x231f" in op:
        note += "  ◀◀ 敵群隻數 [0x231f]"
    if "0x2593" in op:
        note += "  (item slot)"
    # rec(di 載入)
    if mn == "mov" and op.startswith("di,"):
        try:
            di = int(op.split(",")[1], 0)
            note += "  ; rec %d" % (di - 0xbb8)
        except ValueError:
            pass
    # 旗標 bx
    if mn == "mov" and op.startswith("bx,"):
        try:
            bx = int(op.split(",")[1], 0)
            note += "  ; flag 0x%x" % bx
        except ValueError:
            pass
    # 分支 target 記錄
    if mn in ("je", "jne", "jmp", "jz", "jnz", "jb", "jbe", "ja", "jae"):
        try:
            t = int(op, 0)
            targets.add(t)
            note += "  →0x%04x" % t
        except ValueError:
            pass
    print("  0x%05x  %-7s %-22s%s" % (fileoff, mn, op, note))
    if mn == "jmp" and op == "0x6380":
        print("  --- block end (jmp dispatcher 0x6380) ---")

print()
print("分支 targets:", " ".join("0x%04x" % t for t in sorted(targets)))
