#!/usr/bin/env python3
"""DQ3.EXE 素材載入函式反組譯輔助。在 docker 內跑(需 capstone)。

用法:
  re_loaders.py disasm <file_offset_hex> [count]   # 線性反組譯一段
  re_loaders.py findopen                            # 掃描所有 mov ah,0x3d / int 21h 開檔點
  re_loaders.py xref <ds_offset_hex>                # 找對檔名池 offset 的 lea/mov 引用
  re_loaders.py seg0 <seg_off_hex> [count]          # 以 seg0 內位址反組譯(file = HDR + off)

座標系: header HDR=0x1370;主程式段 seg0 file base = HDR(seg0:off -> file HDR+off)。
主資料段 DS=0x14dd,file base 0x16140。
"""
import sys, struct
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"
d = open(EXE, "rb").read()
e_cparhdr = struct.unpack_from('<H', d, 8)[0]
HDR = e_cparhdr * 16  # 0x1370
md = Cs(CS_ARCH_X86, CS_MODE_16)
md.detail = False


def disasm(off, cnt):
    code = d[off:off + cnt * 8]
    n = 0
    for ins in md.disasm(code, off):
        print(f"  {ins.address:06x}: {ins.bytes.hex():<16} {ins.mnemonic} {ins.op_str}")
        n += 1
        if n >= cnt:
            break


def findopen():
    """掃 seg0 程式碼段找 mov ah,0x3d (b4 3d) 開檔點,印前後文。"""
    # seg0 程式碼: file HDR .. HDR+0x10000
    start = HDR
    end = HDR + 0x10000
    hits = []
    i = start
    while i < end - 1:
        # mov ah, 0x3d = B4 3D ; int 21h = CD 21
        if d[i] == 0xB4 and d[i+1] == 0x3D:
            hits.append(i)
        i += 1
    print(f"# mov ah,0x3d (開檔) 出現 {len(hits)} 處 (file off):")
    for h in hits:
        seg0 = h - HDR
        print(f"\n=== file 0x{h:x}  (seg0:{seg0:04x}) ===")
        # 反組譯前 24 bytes 上下文 (回看找 lea dx)
        ctx_start = max(start, h - 40)
        disasm(ctx_start, 24)


def xref(dsoff):
    """找對 DS:dsoff 的引用 (lea dx,[imm] = 8d 16 lo hi ; 或 mov word/byte imm)。"""
    lo = dsoff & 0xff
    hi = (dsoff >> 8) & 0xff
    start = HDR
    end = HDR + 0x10000
    print(f"# 尋找對 DS:0x{dsoff:x} 的引用 (lea dx,[0x{dsoff:x}] = 8d 16 {lo:02x} {hi:02x}):")
    i = start
    found = 0
    while i < end - 3:
        # lea dx,[imm16]  : 8D 16 lo hi
        if d[i] == 0x8D and d[i+1] == 0x16 and d[i+2] == lo and d[i+3] == hi:
            print(f"\n=== lea dx @ file 0x{i:x} (seg0:{i-HDR:04x}) ===")
            disasm(i - 12, 12)
            found += 1
        i += 1
    if not found:
        print("  (無 lea dx 直接命中;可能用 mov dx,imm 或經暫存器)")
        # 再掃 mov dx,imm16 = BA lo hi
        i = start
        while i < end - 2:
            if d[i] == 0xBA and d[i+1] == lo and d[i+2] == hi:
                print(f"\n=== mov dx,imm @ file 0x{i:x} (seg0:{i-HDR:04x}) ===")
                disasm(i - 10, 10)
            i += 1


if __name__ == "__main__":
    cmd = sys.argv[1] if len(sys.argv) > 1 else "disasm"
    if cmd == "disasm":
        off = int(sys.argv[2], 16)
        cnt = int(sys.argv[3]) if len(sys.argv) > 3 else 60
        disasm(off, cnt)
    elif cmd == "seg0":
        off = int(sys.argv[2], 16)
        cnt = int(sys.argv[3]) if len(sys.argv) > 3 else 60
        disasm(HDR + off, cnt)
    elif cmd == "findopen":
        findopen()
    elif cmd == "xref":
        xref(int(sys.argv[2], 16))
