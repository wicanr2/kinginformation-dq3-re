#!/usr/bin/env python3
"""DQ3.EXE 戰鬥子系統反組譯偵察(capstone)。在 docker 內跑。

用法:
  re_battle_dis.py            # 反組譯戰鬥主迴圈鏈 + 指令/回合子函式
  re_battle_dis.py 0xNNNN n   # 任意 seg0 off 反組譯 n 條
  re_battle_dis.py d3mns      # dump D3MNS.DAT 結構嘗試
  re_battle_dis.py shp        # dump DQ3MNS.SHP 表頭
"""
import sys, struct
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

D = open("assets_raw/DQ3.EXE", "rb").read()
e_cparhdr = struct.unpack_from('<H', D, 8)[0]
HDR = e_cparhdr * 16
md = Cs(CS_ARCH_X86, CS_MODE_16)


def dis(off, cnt=60, label=""):
    print(f"\n;==== {label}  seg0:0x{off:04x}  (file 0x{HDR+off:x}) ====")
    foff = HDR + off
    code = D[foff:foff + cnt * 8]
    n = 0
    for ins in md.disasm(code, off):
        print(f"  {ins.address:06x}: {ins.bytes.hex():<14} {ins.mnemonic} {ins.op_str}")
        n += 1
        if n >= cnt:
            break


def dump_d3mns():
    m = open("assets_raw/D3MNS.DAT", "rb").read()
    print(f"; D3MNS.DAT size={len(m)} (0x{len(m):x})")
    for i in range(0, 128, 16):
        row = m[i:i+16]
        print(f"  {i:04x}: " + ' '.join(f'{b:02x}' for b in row) + "  " +
              ''.join(chr(b) if 32 <= b < 127 else '.' for b in row))


def dump_shp():
    m = open("assets_raw/DQ3MNS.SHP", "rb").read()
    print(f"; DQ3MNS.SHP size={len(m)} (0x{len(m):x})")
    print("; first 16 u32 offsets (LE):")
    offs = [struct.unpack_from('<I', m, i*4)[0] for i in range(16)]
    for i, o in enumerate(offs):
        print(f"  [{i}] = 0x{o:08x} ({o})")
    # try first sprite header (8 bytes per b19e: reads 8 bytes at offset)
    o0 = offs[0]
    if o0 < len(m):
        hdr = m[o0:o0+8]
        print(f"; sprite0 header @0x{o0:x}: " + ' '.join(f'{b:02x}' for b in hdr))
        w, h, x2, x3 = struct.unpack_from('<HHHH', m, o0)
        print(f";   as 4xu16: {w} {h} {x2} {x3}")


if len(sys.argv) > 1 and sys.argv[1] == "d3mns":
    dump_d3mns(); sys.exit(0)
if len(sys.argv) > 1 and sys.argv[1] == "shp":
    dump_shp(); sys.exit(0)
if len(sys.argv) > 2:
    dis(int(sys.argv[1], 16), int(sys.argv[2]), "manual"); sys.exit(0)

# 戰鬥進入點 + 主迴圈 + 子函式
dis(0xbd97, 50, "sub_bd97  battle-launch gate (called from update_hud 0x9530)")
dis(0xbddf, 80, "sub_bddf  BATTLE MAIN (calls encounter/bg/restore/turn subs)")
dis(0xc572, 25, "sub_c572  battle sub")
dis(0xc59b, 45, "sub_c59b  battle sub")
dis(0xc5f6, 30, "sub_c5f6  shared (also in update_hud)")
dis(0xc7d9, 70, "sub_c7d9  battle turn sub (also update_hud)")
dis(0xc8c6, 60, "sub_c8c6  battle sub")
dis(0xbcf2, 60, "sub_bcf2  battle sub (status/HUD redraw)")
