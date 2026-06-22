#!/usr/bin/env python3
"""精訊版 DQ3 道具特效修正 patch 產生器 (bugs.md #7)。

讀 assets_raw/DQ3.EXE(唯讀)→ 依 docs/data/codecave_patches.json 套用同長度
region rewrite(目前僅 #7a 隼劍雙擊)→ 輸出 work/DQ3_item_fixed.EXE。
套用前逐 byte 驗證原碼與 JSON 的 orig_bytes 相符,不符即中止(防錯位)。
不更動 MZ header / segment 佈局 / relocation table。

本檔專供 #7;不動他人 tools(re_bugpatch.py / build_fixed_version.py 屬 A)。

用法:
  python tools/re_codecave_patch.py            產生 work/DQ3_item_fixed.EXE
  python tools/re_codecave_patch.py --verify    僅驗證 orig_bytes,不輸出
  python tools/re_codecave_patch.py --base work/DQ3_fixed.EXE
        以 A 的卡關修正版為底再疊 #7a(輸出 work/DQ3_full_fixed.EXE)
"""
import sys, json, os

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
JSON = os.path.join(ROOT, "docs/data/codecave_patches.json")
SRC = os.path.join(ROOT, "assets_raw/DQ3.EXE")


def load_patches():
    with open(JSON, encoding="utf-8") as f:
        return json.load(f)["patches"]


def apply(base_path, out_path, verify_only=False):
    data = bytearray(open(base_path, "rb").read())
    patches = load_patches()
    for p in patches:
        off = p["file_offset"]
        orig = bytes.fromhex(p["orig_bytes"])
        new = bytes.fromhex(p["new_bytes"])
        assert len(orig) == len(new) == p["region_len"], \
            f"#{p['bug']} 長度不符 orig={len(orig)} new={len(new)} region={p['region_len']}"
        cur = bytes(data[off:off + len(orig)])
        if cur != orig:
            print(f"[中止] #{p['bug']} {p['name']} @ f0x{off:x}:")
            print(f"  期望 orig: {orig.hex()}")
            print(f"  實際     : {cur.hex()}")
            print("  → 底檔與預期不符(可能已被改過或選錯底檔),不套用。")
            return 1
        if not verify_only:
            data[off:off + len(new)] = new
        print(f"[OK] #{p['bug']} {p['name']} @ f0x{off:x} "
              f"({'verified' if verify_only else 'patched'} {len(new)} bytes)")
    if verify_only:
        print("驗證完成:所有 orig_bytes 相符。")
        return 0
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    open(out_path, "wb").write(data)
    print(f"輸出: {out_path} ({len(data)} bytes)")
    return 0


def main():
    args = sys.argv[1:]
    verify = "--verify" in args
    base = SRC
    out = os.path.join(ROOT, "work/DQ3_item_fixed.EXE")
    if "--base" in args:
        base = os.path.join(ROOT, args[args.index("--base") + 1])
        out = os.path.join(ROOT, "work/DQ3_full_fixed.EXE")
    return apply(base, out, verify)


if __name__ == "__main__":
    sys.exit(main())
