#!/usr/bin/env python3
"""disasm_to_nasm.py — 把 DQ3.EXE 一段反組譯成 NASM 助記符 .asm,供 nasm 重組驗證。

兩個層次的 RE 證明:
  L1 (coverage):每個 byte 都被正確切分/對齊成指令(或保留為 data)。
  L2 (fidelity):把指令以**助記符**寫出,獨立 nasm 組回,bytes 與原版相同。
L2 才是「RE 成功」的硬證據(L1 用 db 會恆等,意義有限)。

本檔產生 L2 的助記符 asm(`org` 對齊以正確解析相對位移/絕對位址),
並對每條指令附上原始 bytes 註解供逐條比對。重組由 reasm_verify.py 逐指令
比對「nasm 組出的 bytes vs 原版 bytes」,報 byte-identical 比率與不符清單。

用法: disasm_to_nasm.py <file_off_hex> <size> <out.asm>
"""
import sys, json
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"


def gen(off, size):
    d = open(EXE, "rb").read()
    code = d[off:off + size]
    md = Cs(CS_ARCH_X86, CS_MODE_16)
    insns = []
    pos = 0
    for ins in md.disasm(code, off):
        insns.append({
            "addr": ins.address, "size": ins.size,
            "bytes": bytes(ins.bytes).hex(),
            "mnemonic": ins.mnemonic, "op_str": ins.op_str,
            "text": (ins.mnemonic + " " + ins.op_str).strip(),
        })
        pos += ins.size
    tail = code[pos:].hex() if pos < size else ""
    return insns, tail, pos


def main():
    off = int(sys.argv[1], 16); size = int(sys.argv[2]); outp = sys.argv[3]
    insns, tail, covered = gen(off, size)
    # asm:BITS 16 + org=off,讓相對跳轉/絕對位址在重組時落回原值
    lines = ["BITS 16", f"org 0x{off:x}", ""]
    for it in insns:
        lines.append(f"{it['text']}")
    if tail:
        lines.append("db " + ",".join("0x" + tail[i:i+2] for i in range(0, len(tail), 2)))
    open(outp, "w").write("\n".join(lines) + "\n")
    # 旁出 meta(per-instruction 原始 bytes)供逐條驗證
    json.dump({"off": off, "size": size, "covered": covered,
               "insns": insns, "tail": tail},
              open(outp + ".meta.json", "w"))
    print(f"wrote {outp}: {len(insns)} insns, covered {covered}/{size} bytes")


if __name__ == "__main__":
    main()
