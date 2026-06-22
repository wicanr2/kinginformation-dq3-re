#!/usr/bin/env python3
"""splice_lib.py — matching-decompilation 的 splice 框架共用核心。

正路 (b) 的地基:把「已 byte-match 的 C 函式機器碼」splice 進 docs/17 的 byte-identical
重組產物(原版 seg0 bytes),取代該函式對應的 db bytes,其餘維持,組成 EXE。

設計不變量(inv):
  inv.1  matching-decomp 的判準 = 編出來的函式 bytes 與原版該函式 [lo,hi) 逐 byte 相同。
         因此一個「已 match」的 C 函式 splice 進去,**不會改變** seg0 內容 → 整檔 sha256 不變。
  inv.2  0 個 C 函式時,seg0 = 原版 seg0 → 整檔重組 sha256 == 原版(框架正確性的基準)。
  inv.3  函式邊界 [lo,hi) 來自 docs/data/exe_funcs.json(seg0 邏輯 offset)。
         splice 寫入位置 = seg0 buffer 的 [lo:hi];長度必須吻合(C obj 函式 byte 數 == hi-lo)。

本模組只提供「在記憶體裡建 spliced seg0 / 整檔」與「比對」的純函式;
不跑 docker、不寫檔(那是 match_check.py / rebuild.sh 的事)。OMF 解析重用 re_match.py 的
parse_omf_text(抽 LEDATA code bytes)。
"""
import os
import sys
import json
import struct
import hashlib

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
EXE = os.path.join(ROOT, "assets_raw", "DQ3.EXE")
FUNCS_JSON = os.path.join(ROOT, "docs", "data", "exe_funcs.json")

# 讓 re_match.parse_omf_text 可被 import(tools/re_match.py)
sys.path.insert(0, os.path.join(ROOT, "tools"))


def sha(b):
    return hashlib.sha256(b).hexdigest()


def load_exe():
    """回傳原版 DQ3.EXE 全部 bytes。"""
    with open(EXE, "rb") as f:
        return f.read()


def parse_mz(d):
    """解析 MZ header 關鍵欄位,回傳 (hdr_bytes, reloc_off, reloc_cnt)。"""
    e_crlc = struct.unpack_from("<H", d, 6)[0]
    e_cparhdr = struct.unpack_from("<H", d, 8)[0]
    e_lfarlc = struct.unpack_from("<H", d, 24)[0]
    return e_cparhdr * 16, e_lfarlc, e_crlc


def seg0_view(d):
    """從整檔切出 seg0(前 64KB load image,或不足則到 EOF)。回傳 (hdr_bytes, seg0_bytes)。"""
    hdr_bytes, _, _ = parse_mz(d)
    img = d[hdr_bytes:]
    seg0_len = 0x10000 if len(img) >= 0x10000 else len(img)
    return hdr_bytes, img[:seg0_len]


def load_funcs():
    """讀 exe_funcs.json,回傳 {entry: {lo, hi, size, ...}} 與 list。"""
    with open(FUNCS_JSON) as f:
        data = json.load(f)
    funcs = data["funcs"]
    by_entry = {fi["entry"]: fi for fi in funcs}
    return data, by_entry, funcs


def func_orig_bytes(d, lo, hi):
    """原版 seg0 內 [lo,hi) 的 bytes(seg0 邏輯 offset)。"""
    hdr_bytes, _, _ = parse_mz(d)
    return d[hdr_bytes + lo: hdr_bytes + hi]


def extract_obj_code(obj_path):
    """從 MSC 5.1 產出的 .obj(OMF)抽 code segment bytes。

    重用 re_match.parse_omf_text(抽 LEDATA,依 segidx 串接)。
    DQ3 主碼是單一 _TEXT segment;回傳「最大的 code 段」bytes
    (CONST/DATA 段通常較小;若需精準可用 symbol,但對單函式 .obj,_TEXT 即最大段)。
    回傳 dict[segidx] -> bytes 以及 (best_idx, best_bytes)。
    """
    from re_match import parse_omf_text
    segs = parse_omf_text(obj_path)
    if not segs:
        return segs, None, b""
    best_idx = max(segs, key=lambda k: len(segs[k]))
    return segs, best_idx, segs[best_idx]


def check_func_match(d, fi, cand_bytes):
    """把候選函式機器碼 cand_bytes 與原版 [lo,hi) 逐 byte 比對。

    回傳 dict: matched / total / pct / size_ok / first_diff / orig_hex / cand_hex。
    matching-decomp 判準 = matched == total 且 size_ok(長度吻合)。
    """
    lo, hi = fi["lo"], fi["hi"]
    orig = func_orig_bytes(d, lo, hi)
    size_ok = (len(cand_bytes) == len(orig))
    n = min(len(orig), len(cand_bytes))
    matched = sum(1 for i in range(n) if orig[i] == cand_bytes[i])
    total = max(len(orig), len(cand_bytes))
    first_diff = None
    for i in range(n):
        if orig[i] != cand_bytes[i]:
            first_diff = i
            break
    if first_diff is None and len(orig) != len(cand_bytes):
        first_diff = n
    return {
        "entry": fi["entry"], "lo": lo, "hi": hi,
        "matched": matched, "total": total,
        "pct": (100.0 * matched / total) if total else 0.0,
        "size_ok": size_ok,
        "exact": (size_ok and matched == total),
        "first_diff": first_diff,
        "orig": orig, "cand": cand_bytes,
    }


def build_spliced_seg0(d, splices):
    """建 spliced seg0:從原版 seg0 出發,把每個 (lo, hi, cand_bytes) 覆蓋上去。

    splices: list of (lo, hi, cand_bytes)。要求 cand_bytes 長度 == hi-lo。
    inv.1:若每段 cand_bytes 都與原版 [lo,hi) 相同(已 match),結果 == 原版 seg0。
    回傳 spliced seg0 bytes。
    """
    _, seg0 = seg0_view(d)
    buf = bytearray(seg0)
    for lo, hi, cand in splices:
        if len(cand) != hi - lo:
            raise ValueError(
                f"splice 長度不符 @ lo={lo:#x}: cand={len(cand)} != region={hi - lo}")
        if hi > len(buf):
            raise ValueError(f"splice 超出 seg0 範圍 @ hi={hi:#x} > {len(buf):#x}")
        buf[lo:hi] = cand
    return bytes(buf)


def emit_asm(d, spliced_seg0, out_path):
    """產生 nasm 原始檔(與 exe_to_nasm.py 同協議),但 seg0 用 spliced 版本。

    MZ header 欄位 / reloc / padding / runtime+data 全部用原版 db;seg0 用 spliced bytes
    (每條附助記符反組譯註解,解不出退化為 1-byte db,與 exe_to_nasm.py 一致)。
    inv.2:spliced_seg0 == 原版 seg0 時,輸出 .asm 經 nasm 組回 == 原版。
    """
    from capstone import Cs, CS_ARCH_X86, CS_MODE_16

    MZ_FIELDS = ['e_magic', 'e_cblp', 'e_cp', 'e_crlc', 'e_cparhdr', 'e_minalloc',
                 'e_maxalloc', 'e_ss', 'e_sp', 'e_csum', 'e_ip', 'e_cs',
                 'e_lfarlc', 'e_ovno']
    hdr = dict(zip(MZ_FIELDS, struct.unpack_from('<14H', d, 0)))
    hdr_bytes = hdr['e_cparhdr'] * 16
    reloc_off = hdr['e_lfarlc']
    reloc_cnt = hdr['e_crlc']
    reloc_end = reloc_off + reloc_cnt * 4

    def db_line(bs, comment=""):
        body = ",".join(f"0x{b:02x}" for b in bs)
        return f"db {body}" + (f"   ; {comment}" if comment else "")

    L = ["; DQ3.EXE spliced 重組原始檔(splice_lib 自動產生)— seg0 含已 match 的 C 函式 bytes",
         f"; 原版 size={len(d)} header_bytes={hdr_bytes:#x} reloc={reloc_cnt}@{reloc_off:#x}",
         "BITS 16", "", "; ==== MZ header (28 bytes) ===="]
    for i, k in enumerate(MZ_FIELDS):
        L.append(db_line(d[i * 2:i * 2 + 2], f"{k} = {hdr[k]:#06x}"))
    L.append("")
    if reloc_off > 0x1c:
        L.append("; ==== header gap ====")
        L.append(db_line(d[0x1c:reloc_off]))
        L.append("")
    L.append(f"; ==== reloc table ({reloc_cnt} entries) ====")
    for i in range(reloc_cnt):
        o, s = struct.unpack_from('<HH', d, reloc_off + i * 4)
        L.append(db_line(d[reloc_off + i * 4: reloc_off + i * 4 + 4],
                         f"[{i}] {s:#06x}:{o:#06x}"))
    L.append("")
    if hdr_bytes > reloc_end:
        L.append("; ==== header padding ====")
        L.append(db_line(d[reloc_end:hdr_bytes]))
        L.append("")

    img = d[hdr_bytes:]
    L.append(f"; ==== load image ({len(img)} bytes, file {hdr_bytes:#x}..{len(d):#x}) ====")
    md = Cs(CS_ARCH_X86, CS_MODE_16)
    seg0 = spliced_seg0
    seg0_len = len(seg0)
    insns = {ins.address: ins for ins in md.disasm(seg0, 0)}
    pos = 0
    while pos < seg0_len:
        ins = insns.get(pos)
        if ins and ins.size > 0:
            bs = seg0[pos:pos + ins.size]
            L.append(db_line(bs, f"{pos + hdr_bytes:#07x} {ins.mnemonic} {ins.op_str}".strip()))
            pos += ins.size
        else:
            L.append(db_line(seg0[pos:pos + 1], f"{pos + hdr_bytes:#07x} (data)"))
            pos += 1
    # 其餘 image(runtime + data)維持原版 db
    rest = img[seg0_len:]
    if rest:
        L.append(f"; ==== runtime segments + data ({len(rest)} bytes) ====")
        for i in range(0, len(rest), 16):
            L.append(db_line(rest[i:i + 16]))
    L.append("")
    with open(out_path, "w") as f:
        f.write("\n".join(L))
    return len(L)


def coverage_report(d, exact_funcs, all_funcs):
    """計算 byte-match 覆蓋率:已 exact-match 的函式 bytes / seg0 總 bytes。

    exact_funcs: list of fi(已通過 exact match 的函式)。
    回傳 dict: n_matched_funcs / n_total_funcs / matched_bytes / seg0_bytes / pct。
    """
    _, seg0 = seg0_view(d)
    matched_bytes = sum(fi["hi"] - fi["lo"] for fi in exact_funcs)
    # seg0 內被 exe_funcs.json 函式覆蓋的總 bytes(分母可選 seg0 全長或函式覆蓋)
    func_bytes = sum(fi["hi"] - fi["lo"] for fi in all_funcs)
    return {
        "n_matched_funcs": len(exact_funcs),
        "n_total_funcs": len(all_funcs),
        "matched_bytes": matched_bytes,
        "func_bytes": func_bytes,
        "seg0_bytes": len(seg0),
        "pct_of_funcs": (100.0 * len(exact_funcs) / len(all_funcs)) if all_funcs else 0.0,
        "pct_of_func_bytes": (100.0 * matched_bytes / func_bytes) if func_bytes else 0.0,
        "pct_of_seg0": (100.0 * matched_bytes / len(seg0)) if seg0 else 0.0,
    }
