#!/usr/bin/env python3
"""splice_rebuild.py — 把已 exact-match 的 C 函式 splice 進 byte-identical 重組,產 .asm。

在 docker 內被 rebuild.sh 呼叫(需 capstone)。流程:
  1. 讀 re/match/*.c 的 @match marker → 編譯(可選)→ 抽函式 bytes → 與原版比對。
     只有 exact-match 的函式才會被 splice(splice_lib.inv.1:exact 的 bytes == 原版,
     splice 不改變 seg0;這正是「splice 框架正確 + 函式真的 match」的雙重保證)。
  2. build_spliced_seg0:把這些函式 bytes 覆蓋進原版 seg0(0 函式 → == 原版 seg0)。
  3. emit_asm:產生 nasm 原始檔(seg0 用 spliced 版,其餘維持原版 db)。
  4. 印出將被 splice 的函式清單 + 覆蓋率。

之後 rebuild.sh 用 nasm 組 .asm,sha256 與原版比。

用法(docker 內):
  splice_rebuild.py <out.asm> [--no-compile]
"""
import os
import sys
import glob

import splice_lib as S
from match_check import parse_marker, compile_obj

ROOT = S.ROOT
MATCH_DIR = os.path.join(ROOT, "re", "match")


def main():
    args = [a for a in sys.argv[1:]]
    no_compile = "--no-compile" in args
    args = [a for a in args if a != "--no-compile"]
    out_asm = args[0] if args else os.path.join(ROOT, "work", "csplice", "dq3.asm")
    os.makedirs(os.path.dirname(out_asm), exist_ok=True)

    d = S.load_exe()
    _, by_entry, all_funcs = S.load_funcs()

    c_files = sorted(glob.glob(os.path.join(MATCH_DIR, "*.c")))
    splices = []
    exact_funcs = []
    for c_path in c_files:
        entry, model = parse_marker(c_path)
        if entry is None or entry not in by_entry:
            continue
        fi = by_entry[entry]
        if no_compile:
            tag = (model or "/AS").lstrip("/").lower()
            obj = os.path.join(os.path.dirname(c_path),
                               f"{os.path.splitext(os.path.basename(c_path))[0]}.{tag}.msc.obj")
            obj = obj if os.path.exists(obj) else None
        else:
            obj = compile_obj(c_path, model)
        if not obj:
            print(f"[skip] {os.path.relpath(c_path, ROOT)}: 無 .obj")
            continue
        _, _, code = S.extract_obj_code(obj)
        chk = S.check_func_match(d, fi, code)
        if chk["exact"]:
            splices.append((fi["lo"], fi["hi"], code))
            exact_funcs.append(fi)
            print(f"[splice] {os.path.relpath(c_path, ROOT)}  entry={entry:#06x}  "
                  f"[{fi['lo']:#06x},{fi['hi']:#06x}) {fi['hi'] - fi['lo']} bytes")
        else:
            print(f"[skip-nomatch] {os.path.relpath(c_path, ROOT)}  entry={entry:#06x}  "
                  f"{chk['matched']}/{chk['total']} ({chk['pct']:.1f}%) — 不 splice")

    spliced = S.build_spliced_seg0(d, splices)
    nlines = S.emit_asm(d, spliced, out_asm)
    cov = S.coverage_report(d, exact_funcs, all_funcs)
    print(f"\n[splice_rebuild] {len(splices)} 個 exact-match 函式 splice 進 seg0")
    print(f"[splice_rebuild] 覆蓋率: {cov['n_matched_funcs']}/{cov['n_total_funcs']} 函式 "
          f"({cov['pct_of_funcs']:.2f}%), {cov['matched_bytes']}/{cov['func_bytes']} 函式 bytes "
          f"({cov['pct_of_func_bytes']:.2f}%)")
    print(f"[splice_rebuild] 產出 {os.path.relpath(out_asm, ROOT)} ({nlines} lines)")
    # 把覆蓋率寫成小檔供 rebuild.sh 顯示
    cov_path = os.path.join(os.path.dirname(out_asm), "coverage.txt")
    with open(cov_path, "w") as f:
        f.write(f"matched_funcs={cov['n_matched_funcs']}\n"
                f"total_funcs={cov['n_total_funcs']}\n"
                f"pct_of_funcs={cov['pct_of_funcs']:.4f}\n"
                f"matched_bytes={cov['matched_bytes']}\n"
                f"func_bytes={cov['func_bytes']}\n"
                f"pct_of_func_bytes={cov['pct_of_func_bytes']:.4f}\n")


if __name__ == "__main__":
    main()
