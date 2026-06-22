#!/usr/bin/env python3
"""組裝「精訊 DQ3 修正版(binary patch 對照組)」。
彙整各 agent 定位的高信心、同長度 binary patch + 重繪 sprite:
  #1 巴拉摩斯打不死        bug_patches.json   (EXE in-place)
  #2 彩虹水滴誤拿黃寶珠     bug_patches.json   (EXE in-place)
  #3 五頭龍/歐里狄加當機    重繪 sprite        (DQ3MNS_fixed.SHP,見 make_sprites.py)
  #4 勇者 MP+1            stat_patches.json  (EXE 靜態 DGROUP 成長表資料)
  #7a 隼劍雙擊            codecave_patches.json (EXE 同長度區段改寫,復用既有 re-attack)
未修(留 C 層 / SDL2,根因見 docs/18,23,22):#5 升級錯亂、#6 255溢位、#7b 魔甲抗魔。
  #7c 祈禱之戒:反組譯確認本就生效,不需修。
原始 EXE 唯讀;輸出 work/DQ3_fixed.EXE + work/dq3_fixed_game/。
"""
import json, os, shutil, struct, subprocess, sys

ORIG = "assets_raw/DQ3.EXE"
exe = bytearray(open(ORIG, "rb").read())
applied = []

def apply(off, orig_hex, new_hex, label):
    o = bytes.fromhex(orig_hex); n = bytes.fromhex(new_hex)
    assert len(o) == len(n), f"{label}: orig/new 長度不一 ({len(o)}/{len(n)})"
    cur = bytes(exe[off:off+len(o)])
    assert cur == o, f"{label} @{off:#x}: 原 bytes {cur.hex()} != 預期 {orig_hex}"
    exe[off:off+len(n)] = n
    applied.append(f"{label} @{off:#x} ({len(n)}B)")

# #1 #2  (bug_patches.json:bug 1/2;#3 用 sprite,跳過)
for p in json.load(open("docs/data/bug_patches.json"))["patches"]:
    if p["bug"] in (1, 2):
        apply(p["file_offset"], p["orig"], p["new"], f"Bug{p['bug']} {p['name']}")

# #4  (stat_patches.json:category exe-data)
for p in json.load(open("docs/data/stat_patches.json")).get("patches", []):
    if p.get("category") == "exe-data" and p.get("orig"):
        apply(p["file_offset"], p["orig"], p["new"], f"Bug{p['bug']} {p.get('name','勇者MP')}")

# #7a (codecave_patches.json:in-place-region-rewrite)
for p in json.load(open("docs/data/codecave_patches.json")).get("patches", []):
    if "orig_bytes" in p:
        apply(p["file_offset"], p["orig_bytes"], p["new_bytes"], f"Bug{p['bug']} {p['name']}")

os.makedirs("work", exist_ok=True)
open("work/DQ3_fixed.EXE", "wb").write(exe)
diff = sum(1 for a, b in zip(open(ORIG,'rb').read(), exe) if a != b)
print(f"EXE patch 套用 ({diff} bytes 變動):")
for a in applied: print("  ", a)

# #3 sprite:需 work/DQ3MNS_fixed.SHP(make_sprites.py)
if not os.path.exists("work/DQ3MNS_fixed.SHP"):
    subprocess.run([sys.executable, "tools/make_sprites.py"], check=True)

GD = "work/dq3_fixed_game"
if os.path.exists(GD): shutil.rmtree(GD)
shutil.copytree("assets_raw", GD)
shutil.copy("work/DQ3_fixed.EXE", os.path.join(GD, "DQ3.EXE"))
shutil.copy("work/DQ3MNS_fixed.SHP", os.path.join(GD, "DQ3MNS.SHP"))
print(f"\n完整修正版(對照組): {GD}/")
print("  EXE 修:#1 打不死 / #2 彩虹水滴 / #4 勇者MP / #7a 隼劍雙擊")
print("  SHP 修:#3 五頭龍大王 + 歐里狄加 sprite 復原")
print("  留 C 層:#5 升級錯亂 / #6 255溢位 / #7b 魔甲抗魔(#7c 不需修)")
