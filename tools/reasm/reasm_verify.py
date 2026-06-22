#!/usr/bin/env python3
"""reasm_verify.py — 用 nasm 把助記符 .asm 組回,逐 byte 比對原版該段。

流程:disasm_to_nasm 產 .asm + .meta.json → nasm 組成 flat binary →
與原版同段 bytes 比對。報:整段是否 byte-identical;若否,逐指令找出
哪幾條助記符 round-trip 後 encoding 不同(16-bit 編碼歧義常見點),
給 byte-identical 指令比率 + 不符明細。

用法(docker 內,需 nasm + capstone):
  reasm_verify.py <out.asm> <file_off_hex> <size>
"""
import sys, os, json, subprocess
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"


def main():
    asm = sys.argv[1]; off = int(sys.argv[2], 16); size = int(sys.argv[3])
    binout = asm + ".bin"
    # nasm:org 已在 asm 內;用 -f bin 產 flat。nasm 會從 org 起算,
    # 但 -f bin 輸出從 org 偏移開始填,故輸出檔開頭含 org 個 0 → 取 [off:off+size]。
    r = subprocess.run(["nasm", "-f", "bin", "-o", binout, asm],
                       capture_output=True, text=True)
    if r.returncode != 0:
        print("NASM ERROR:\n" + r.stderr)
        sys.exit(2)
    blob = open(binout, "rb").read()
    seg = blob[off:off + size]
    orig = open(EXE, "rb").read()[off:off + size]

    if seg == orig:
        print(f"RESULT: PASS — segment {off:#x}..{off+size:#x} byte-identical ({size} bytes) via nasm")
        return 0

    # 逐指令定位不符
    meta = json.load(open(asm + ".meta.json"))
    md = Cs(CS_ARCH_X86, CS_MODE_16)
    same = 0; diff = []
    for it in meta["insns"]:
        a = it["addr"]; n = it["size"]
        o = orig[a - off: a - off + n]
        s = seg[a - off: a - off + n]
        if o == s:
            same += 1
        else:
            diff.append((a, it["text"], o.hex(), s.hex()))
    tot = len(meta["insns"])
    print(f"RESULT: partial — {same}/{tot} insns byte-identical "
          f"({100*same/tot:.1f}%); {len(diff)} re-encode differently")
    for a, t, o, s in diff[:25]:
        print(f"  {a:04x}: {t:<28} orig={o} nasm={s}")
    # 整段 byte 比率
    bsame = sum(1 for i in range(size) if seg[i] == orig[i])
    print(f"byte-level: {bsame}/{size} ({100*bsame/size:.1f}%) identical")
    return 1


if __name__ == "__main__":
    sys.exit(main())
