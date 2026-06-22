#!/usr/bin/env python3
"""reasm_ks.py — capstone 反組譯 → keystone 重組,逐 byte 比對(同一語法家族)。

keystone 與 capstone 同屬一家,接受 capstone 輸出的 Intel 助記符,
避免 NASM 語法翻譯歧義。對每條指令:capstone 解碼出助記符 → keystone 以
相同 16-bit 模式、相同位址(處理相對跳轉)組回 → 比對 bytes。

這證明「我們命名的指令(助記符)能無歧義組回原版機器碼」= 反組譯語意正確,
非只是 bytes 搬運。報整段是否 byte-identical + 逐指令比率 + 不符明細
(16-bit encoding 歧義點)。

用法(docker 內,需 capstone + keystone-engine):
  reasm_ks.py <file_off_hex> <size> [name]
  reasm_ks.py funcs <exe_funcs.json>     # 全 seg0 函式批次統計
"""
import sys, json
from capstone import Cs, CS_ARCH_X86, CS_MODE_16
from keystone import Ks, KS_ARCH_X86, KS_MODE_16

EXE = "assets_raw/DQ3.EXE"
_md = Cs(CS_ARCH_X86, CS_MODE_16)
_ks = Ks(KS_ARCH_X86, KS_MODE_16)


def reasm_segment(code, base):
    """回傳 (insn_total, insn_same, byte_total, byte_same, diffs[], covered)."""
    insn_total = insn_same = 0
    byte_same = 0
    diffs = []
    pos = 0
    for ins in _md.disasm(code, base):
        text = (ins.mnemonic + " " + ins.op_str).strip()
        orig = bytes(ins.bytes)
        try:
            enc, _ = _ks.asm(text, addr=ins.address)
            enc = bytes(enc) if enc else b""
        except Exception as e:
            enc = b""
        insn_total += 1
        if enc == orig:
            insn_same += 1
            byte_same += len(orig)
        else:
            diffs.append((ins.address, text, orig.hex(), enc.hex()))
        pos += ins.size
    return insn_total, insn_same, pos, byte_same, diffs, pos


def verify_one(off, size, name=""):
    code = open(EXE, "rb").read()[off:off + size]
    it, isame, covered, bsame, diffs, _ = reasm_segment(code, off)
    tag = f"sub_{name or off:>04}".replace("sub_0x", "sub_")
    print(f"=== {name or hex(off)} file={off:#x} size={size} ===")
    print(f"  insns {isame}/{it} byte-identical ({100*isame/max(it,1):.1f}%); "
          f"bytes {bsame}/{covered} ({100*bsame/max(covered,1):.1f}%); covered {covered}/{size}")
    for a, t, o, e in diffs[:20]:
        print(f"    {a:04x}: {t:<28} orig={o} ks={e}")
    return it, isame, covered, bsame


def batch(jsonpath):
    d = open(EXE, "rb").read()
    HDR = 0x1370
    funcs = json.load(open(jsonpath))["funcs"]
    T_i = T_is = T_bs = T_cov = 0
    full_match = 0
    worst = []
    for f in funcs:
        off = HDR + f["entry"]; size = f["size"]
        if size < 1:
            continue
        code = d[off:off + size]
        it, isame, covered, bsame, diffs, _ = reasm_segment(code, off)
        T_i += it; T_is += isame; T_bs += bsame; T_cov += covered
        if it and isame == it and bsame == covered == size:
            full_match += 1
        else:
            worst.append((f["entry"], size, it - isame, len(diffs)))
    print("===== seg0 batch reassembly (capstone->keystone) =====")
    print(f"functions: {len(funcs)}, fully byte-identical functions: {full_match} "
          f"({100*full_match/len(funcs):.1f}%)")
    print(f"instructions: {T_is}/{T_i} byte-identical ({100*T_is/max(T_i,1):.2f}%)")
    print(f"bytes (covered): {T_bs}/{T_cov} ({100*T_bs/max(T_cov,1):.2f}%)")
    # 列出最不對齊的函式 + 統計不符助記符
    from collections import Counter
    mism = Counter()
    for f in funcs:
        off = HDR + f["entry"]; size = f["size"]
        if size < 1: continue
        code = d[off:off+size]
        _,_,_,_,diffs,_ = reasm_segment(code, off)
        for a,t,o,e in diffs:
            mism[t.split()[0]] += 1
    print("top re-encode-mismatch mnemonics:")
    for m,c in mism.most_common(15):
        print(f"  {m:<10} {c}")


if __name__ == "__main__":
    if sys.argv[1] == "funcs":
        batch(sys.argv[2])
    else:
        off = int(sys.argv[1], 16); size = int(sys.argv[2])
        name = sys.argv[3] if len(sys.argv) > 3 else ""
        verify_one(off, size, name)
