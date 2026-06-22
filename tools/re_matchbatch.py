#!/usr/bin/env python3
"""re_matchbatch — 批次比對「MSC 5.1 編出的 .obj」與「原版 DQ3.EXE 函式 bytes」。

對每個 target:抽 OBJ code bytes、抽原版該函式 bytes,逐 byte 比對並輸出 match%。
同時輸出「遮罩後」match%:把 16-bit 絕對位址運算元 (data reloc) 遮成 00,
因為全域/陣列的 linker 指派位址必然與原版不同,屬資料重定位差異而非 codegen 差異;
遮罩後仍能看出「指令序列/opcode/結構」是否 byte-identical。

用法:
  re_matchbatch.py <seg0_off_hex> <size> <obj_path>
  re_matchbatch.py manifest tools/re_match_manifest.json

在 docker (dq3-recap / capstone) 內跑。
"""
import sys, json, struct

EXE = "assets_raw/DQ3.EXE"
HDR = 0x1370


def orig_bytes(off, size):
    d = open(EXE, "rb").read()
    return d[HDR + off: HDR + off + size]


def parse_omf_text(path):
    data = open(path, "rb").read()
    i = 0
    segs = {}
    while i < len(data):
        rectype = data[i]
        rlen = struct.unpack_from("<H", data, i + 1)[0]
        body = data[i + 3: i + 3 + rlen - 1]
        if rectype in (0xA0, 0xA1):
            p = 0
            segidx = body[p]
            if segidx & 0x80:
                segidx = ((segidx & 0x7F) << 8) | body[p + 1]; p += 2
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


def obj_code(path):
    """回傳 code segment (index 最小的非空段) 的 bytes。"""
    segs = parse_omf_text(path)
    if not segs:
        return b""
    # 取最小 index 段 (MSC small model _TEXT 通常是 seg 1)
    return segs[min(segs)]


def disp_addr_mask(code, base=0):
    """用 capstone 找出每條指令的『16-bit 絕對位址 disp/imm 運算元』位元組位置,
    回傳一個 set(byte offset),供遮罩。只遮真正屬於絕對位址的 2 bytes。"""
    from capstone import Cs, CS_ARCH_X86, CS_MODE_16
    md = Cs(CS_ARCH_X86, CS_MODE_16)
    md.detail = True
    mask = set()
    for ins in md.disasm(code, base):
        # disp_offset / disp_size 來自 capstone detail (x86)
        try:
            dx = ins.disp_offset
            dz = ins.disp_size
        except Exception:
            dx = dz = 0
        # 直接定址 (mod=00, rm=110) 與 a1/a3 (mov ax,[mem]) 的 16-bit 絕對位址
        if dz == 2 and dx > 0:
            for k in range(dx, dx + 2):
                mask.add(ins.address - base + k)
    return mask


def cmp_one(off, size, obj_path, show=True):
    a = orig_bytes(off, size)
    b = obj_code(obj_path)
    n = min(len(a), len(b))
    total = max(len(a), len(b))
    raw = sum(1 for i in range(n) if a[i] == b[i])
    # 遮罩兩邊各自的絕對位址運算元位元組
    ma = disp_addr_mask(a)
    mb = disp_addr_mask(b)
    masked_match = 0
    for i in range(n):
        if i in ma or i in mb:
            masked_match += 1  # 遮罩位元組視為相符 (屬 data reloc 差異)
        elif a[i] == b[i]:
            masked_match += 1
    rawpct = 100.0 * raw / total if total else 0.0
    mpct = 100.0 * masked_match / total if total else 0.0
    samelen = len(a) == len(b)
    if show:
        print(f"sub_{off:04x}: size orig={len(a)} obj={len(b)} samelen={samelen}")
        print(f"  raw match    {raw}/{total} = {rawpct:.1f}%")
        print(f"  masked match {masked_match}/{total} = {mpct:.1f}%  (絕對位址運算元遮罩後)")
        print(f"  orig: {a.hex()}")
        print(f"  obj : {b.hex()}")
    return dict(off=off, size=len(a), objsize=len(b), samelen=samelen,
               raw=rawpct, masked=mpct)


def main():
    if len(sys.argv) < 2:
        print(__doc__); sys.exit(1)
    if sys.argv[1] == "manifest":
        man = json.load(open(sys.argv[2]))
        results = []
        for t in man:
            results.append(cmp_one(int(t["off"], 16), t["size"], t["obj"]))
            print()
        print("=== summary ===")
        for r in results:
            tag = "EXACT" if r["masked"] >= 99.99 and r["samelen"] else \
                  ("STRUCT" if r["masked"] >= 99.99 else f"{r['masked']:.0f}%")
            print(f"  sub_{r['off']:04x}: raw={r['raw']:.0f}% masked={r['masked']:.0f}% [{tag}]")
    else:
        off = int(sys.argv[1], 16); size = int(sys.argv[2]); obj = sys.argv[3]
        cmp_one(off, size, obj)


if __name__ == "__main__":
    main()
