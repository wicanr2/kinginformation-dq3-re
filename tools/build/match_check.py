#!/usr/bin/env python3
"""match_check.py — byte-match 覆蓋率工具(matching-decompilation)。

流程:
  1. 收集 re/match/*.c(或指定檔)。每個 .c 用 marker comment 宣告它對應哪個原版函式:
        // @match entry=0xNNNN          (seg0 邏輯 offset,= exe_funcs.json 的 entry)
     可附 model:   // @match entry=0xfa29 model=/AS
  2. 對每個 .c 用 MSC 5.1(tools/build/msc_compile.sh)編成 .obj(docker 內,不污染 host)。
  3. 從 .obj 抽出該函式機器碼(OMF LEDATA,取最大 code 段)。
  4. 與原版 DQ3.EXE 該函式 [lo,hi) bytes 逐一比對。
  5. 報告:match / total、各函式 match%、size 是否吻合、不符處 first diff + hexdump。

判準:exact match = 長度吻合 且 逐 byte 全同。只有 exact match 的函式才會被 rebuild.sh
splice(見 splice_lib.inv.1)。

用法:
  match_check.py                       # 掃 re/match/*.c
  match_check.py re/match/foo.c ...    # 指定檔
  match_check.py --json work/match_report.json   # 額外輸出 json
  match_check.py --no-compile          # 不編譯,直接用已存在的 .obj(除錯用)

docker 內編譯;比對在 host(純 Python + capstone via re_match)。
無 @match marker 的 .c 會被略過並警告。
"""
import os
import re
import sys
import json
import glob
import subprocess

import splice_lib as S

ROOT = S.ROOT
MATCH_DIR = os.path.join(ROOT, "re", "match")
COMPILE_SH = os.path.join(ROOT, "tools", "build", "msc_compile.sh")

MARKER = re.compile(r"@match\s+entry=(0x[0-9a-fA-F]+|\d+)(?:\s+model=(/A[SMCLH]))?")
# fallback:由 sub_XXXX 命名(函式名 / 檔名)推 entry,XXXX 為 seg0 offset 的 hex。
SUBNAME = re.compile(r"\bsub_([0-9a-fA-F]{3,5})\b")
MODELRE = re.compile(r"model=(/A[SMCLH])")


def parse_marker(c_path):
    """決定這個 .c 對應哪個原版函式與 memory model。回傳 (entry, model) 或 (None, None)。

    優先序:
      1) 明確 marker `// @match entry=0xNNNN [model=/AS]`。
      2) fallback:檔名或檔頭出現 `sub_XXXX`(XXXX=seg0 offset hex)→ 推 entry。
         此 fallback 讓 match C 既有命名慣例(sub_e6b9.c)免改即可被檢查。
    model 預設 /AS;若檔頭有 `model=/A?` 則採用。
    """
    base = os.path.basename(c_path)
    with open(c_path, encoding="utf-8", errors="replace") as f:
        head = f.read(4096)
    mm = MODELRE.search(head)
    model = mm.group(1) if mm else "/AS"
    m = MARKER.search(head)
    if m:
        return int(m.group(1), 0), (m.group(2) or model)
    # fallback:先看檔名,再看檔頭
    sn = SUBNAME.search(base) or SUBNAME.search(head)
    if sn:
        return int(sn.group(1), 16), model
    return None, None


def compile_obj(c_path, model):
    """用 msc_compile.sh 編出 .obj。回傳產出的 .obj 路徑或 None。"""
    flags = f"/c {model} /Ox"
    name = os.path.splitext(os.path.basename(c_path))[0]
    tag = model.lstrip("/").lower()
    out_obj = os.path.join(os.path.dirname(c_path), f"{name}.{tag}.msc.obj")
    if os.path.exists(out_obj):
        os.remove(out_obj)
    proc = subprocess.run(
        ["bash", COMPILE_SH, c_path] + flags.split(),
        capture_output=True, text=True)
    if proc.returncode != 0:
        sys.stderr.write(proc.stdout + proc.stderr)
    return out_obj if os.path.exists(out_obj) else None


def main():
    argv = sys.argv[1:]
    json_out = None
    no_compile = False
    files = []
    i = 0
    while i < len(argv):
        a = argv[i]
        if a == "--json":
            json_out = argv[i + 1]; i += 2; continue
        if a == "--no-compile":
            no_compile = True; i += 1; continue
        files.append(a); i += 1

    if not files:
        files = sorted(glob.glob(os.path.join(MATCH_DIR, "*.c")))
    if not files:
        print(f"[match_check] 沒有 .c 可檢查(re/match/ 為空)。")
        print(f"[match_check] 覆蓋率: 0 函式 — 框架就緒,等待 match C source。")
        if json_out:
            os.makedirs(os.path.dirname(os.path.abspath(json_out)), exist_ok=True)
            json.dump({"results": [], "exact": [], "summary": {"n_total_funcs": 0}},
                      open(json_out, "w"), indent=1)
        return 0

    d = S.load_exe()
    _, by_entry, all_funcs = S.load_funcs()

    results = []
    exact_entries = []
    for c_path in files:
        entry, model = parse_marker(c_path)
        rel = os.path.relpath(c_path, ROOT)
        if entry is None:
            print(f"[skip] {rel}: 無 @match marker(需 `// @match entry=0xNNNN`)")
            continue
        fi = by_entry.get(entry)
        if fi is None:
            print(f"[FAIL] {rel}: entry {entry:#x} 不在 exe_funcs.json")
            results.append({"file": rel, "entry": entry, "error": "entry not in exe_funcs.json"})
            continue

        if no_compile:
            tag = (model or "/AS").lstrip("/").lower()
            obj = os.path.join(os.path.dirname(c_path),
                               f"{os.path.splitext(os.path.basename(c_path))[0]}.{tag}.msc.obj")
            obj = obj if os.path.exists(obj) else None
        else:
            obj = compile_obj(c_path, model)
        if not obj:
            print(f"[FAIL] {rel}: 編譯失敗(無 .obj 產出)")
            results.append({"file": rel, "entry": entry, "error": "compile failed"})
            continue

        _, _, code = S.extract_obj_code(obj)
        chk = S.check_func_match(d, fi, code)
        status = "MATCH" if chk["exact"] else "DIFF"
        print(f"[{status}] {rel}  entry={entry:#06x}  "
              f"match {chk['matched']}/{chk['total']} = {chk['pct']:.1f}%  "
              f"size {len(code)} vs {fi['hi'] - fi['lo']}"
              + ("" if chk["size_ok"] else "  ⚠ size 不符"))
        if not chk["exact"]:
            fd = chk["first_diff"]
            if fd is not None:
                lo = max(0, fd - (fd % 16))
                oa = chk["orig"][lo:lo + 32]
                cb = chk["cand"][lo:lo + 32]
                print(f"        first diff @ +{fd:#x}")
                print(f"        orig: {oa.hex()}")
                print(f"        cand: {cb.hex()}")
        results.append({
            "file": rel, "entry": entry, "model": model,
            "matched": chk["matched"], "total": chk["total"], "pct": chk["pct"],
            "size_ok": chk["size_ok"], "exact": chk["exact"],
            "first_diff": chk["first_diff"], "obj": os.path.relpath(obj, ROOT),
        })
        if chk["exact"]:
            exact_entries.append(entry)

    exact_funcs = [by_entry[e] for e in exact_entries]
    cov = S.coverage_report(d, exact_funcs, all_funcs)
    print("\n=== byte-match 覆蓋率 ===")
    print(f"  exact-match 函式: {cov['n_matched_funcs']} / {cov['n_total_funcs']} "
          f"({cov['pct_of_funcs']:.2f}%)")
    print(f"  match bytes: {cov['matched_bytes']} / {cov['func_bytes']} 函式 bytes "
          f"({cov['pct_of_func_bytes']:.2f}%)  | seg0 {cov['seg0_bytes']} bytes "
          f"({cov['pct_of_seg0']:.2f}% of seg0)")

    if json_out:
        os.makedirs(os.path.dirname(os.path.abspath(json_out)) or ".", exist_ok=True)
        json.dump({"results": results,
                   "exact": exact_entries,
                   "summary": cov},
                  open(json_out, "w"), indent=1)
        print(f"  → json: {os.path.relpath(json_out, ROOT)}")

    # 任何宣告了 @match 卻沒 exact 的 → 視為「尚未 match」,非錯誤(進行中);
    # 但編譯失敗 / entry 不存在 → 回非 0,讓 CI 抓得到。
    hard_fail = any("error" in r for r in results)
    return 1 if hard_fail else 0


if __name__ == "__main__":
    sys.exit(main())
