#!/usr/bin/env python3
"""byte-match PoC:把候選編譯器產出的 16-bit .obj/.exe 反組譯,
與原版 DQ3.EXE 內某函式的 bytes 逐位元比對,輸出 match%。

用法:
  re_match.py orig <seg0_off_hex> <size>          # dump 原版某函式 bytes
  re_match.py obj  <path.obj> <symbol|@off> [size] # dump OBJ 內某段 code bytes
  re_match.py diff <orig_hex> <cand_hex>           # 比對兩段 hex
  re_match.py cmp  <seg0_off_hex> <size> <cand_hex># 直接比對原版函式 vs 候選 hex

在 docker 內跑(需要 capstone)。OBJ 解析為簡化 OMF (Intel Object Module Format)。
"""
import sys, struct

EXE = "assets_raw/DQ3.EXE"
HDR = 0x1370  # seg0 file base = e_cparhdr*16


def orig_bytes(off, size):
    d = open(EXE, "rb").read()
    return d[HDR + off: HDR + off + size]


def parse_omf_text(path):
    """極簡 OMF 解析:抽出所有 LEDATA(0xA0/0xA1) 紀錄的 code bytes,
    依 segment index + offset 串接。回傳 dict[segidx] -> bytes(以 offset 排)。"""
    data = open(path, "rb").read()
    i = 0
    segs = {}
    while i < len(data):
        rectype = data[i]
        rlen = struct.unpack_from("<H", data, i + 1)[0]
        body = data[i + 3: i + 3 + rlen - 1]  # 末位是 checksum
        if rectype in (0xA0, 0xA1):  # LEDATA / LEDATA32
            # segidx(1 or 2 bytes index) + enum offset(2 or 4)
            p = 0
            segidx = body[p]
            if segidx & 0x80:
                segidx = ((segidx & 0x7F) << 8) | body[p + 1]
                p += 2
            else:
                p += 1
            if rectype == 0xA1:
                eoff = struct.unpack_from("<I", body, p)[0]; p += 4
            else:
                eoff = struct.unpack_from("<H", body, p)[0]; p += 2
            chunk = body[p:]
            buf = segs.setdefault(segidx, bytearray())
            end = eoff + len(chunk)
            if len(buf) < end:
                buf.extend(b"\x00" * (end - len(buf)))
            buf[eoff:end] = chunk
        i += 3 + rlen
    return {k: bytes(v) for k, v in segs.items()}


def disasm(code, base=0):
    from capstone import Cs, CS_ARCH_X86, CS_MODE_16
    md = Cs(CS_ARCH_X86, CS_MODE_16)
    for ins in md.disasm(code, base):
        print(f"  {ins.address:04x}: {ins.bytes.hex():<14} {ins.mnemonic} {ins.op_str}")


def match_pct(a, b):
    """逐 byte 比對(對齊到較短者),回傳 (matched, total, pct)。"""
    n = min(len(a), len(b))
    m = sum(1 for i in range(n) if a[i] == b[i])
    total = max(len(a), len(b))
    return m, total, (100.0 * m / total if total else 0.0)


def show_diff(a, b):
    n = max(len(a), len(b))
    for i in range(0, n, 16):
        ra = a[i:i + 16]; rb = b[i:i + 16]
        marks = "".join("  " if (j < len(ra) and j < len(rb) and ra[j] == rb[j]) else "^^"
                        for j in range(16))
        print(f"  {i:04x} A: {ra.hex():<32}")
        print(f"       B: {rb.hex():<32}")
        print(f"          {marks}")


def main():
    if len(sys.argv) < 2:
        print(__doc__); sys.exit(1)
    cmd = sys.argv[1]
    if cmd == "orig":
        off = int(sys.argv[2], 16); size = int(sys.argv[3])
        b = orig_bytes(off, size)
        print(f"# orig sub_{off:04x} size={size}")
        print(b.hex())
        disasm(b, off)
    elif cmd == "obj":
        segs = parse_omf_text(sys.argv[2])
        for idx, code in sorted(segs.items()):
            print(f"# OMF seg index {idx}  ({len(code)} bytes)")
            print(code.hex())
            disasm(code)
    elif cmd == "diff":
        a = bytes.fromhex(sys.argv[2]); b = bytes.fromhex(sys.argv[3])
        m, t, p = match_pct(a, b)
        print(f"match {m}/{t} = {p:.1f}%")
        show_diff(a, b)
    elif cmd == "cmp":
        off = int(sys.argv[2], 16); size = int(sys.argv[3])
        a = orig_bytes(off, size); b = bytes.fromhex(sys.argv[4])
        m, t, p = match_pct(a, b)
        print(f"orig sub_{off:04x} vs candidate: match {m}/{t} = {p:.1f}%")
        show_diff(a, b)
    else:
        print(__doc__); sys.exit(1)


if __name__ == "__main__":
    main()
