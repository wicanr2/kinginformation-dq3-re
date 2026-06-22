#!/usr/bin/env python3
"""組裝「精訊 DQ3 修正版」:
- EXE: 套用 bug agent 的 Bug1(巴拉摩斯打不死)+ Bug2(彩虹水滴)高信心 in-place patch。
  Bug3(九頭龍當機)**不用** EXE reroute,改用本專案重繪的真實 sprite(見下),更忠實。
- SHP: 用 tools/make_sprites.py 產的 work/DQ3MNS_fixed.SHP(復原 128 歐里狄加 / 129 五頭龍大王)。
- 輸出可在 DOSBox 跑的 work/dq3_fixed_game/(原素材 + 修正 EXE/SHP)。
"""
import json, os, shutil, struct, subprocess, sys

ORIG = "assets_raw/DQ3.EXE"
exe = bytearray(open(ORIG, "rb").read())
patches = json.load(open("docs/data/bug_patches.json"))["patches"]

applied = []
for p in patches:
    if p["bug"] not in (1, 2):       # Bug3 用 SHP sprite 修復,跳過 EXE reroute
        continue
    off = p["file_offset"]
    orig_b = bytes.fromhex(p["orig"]); new_b = bytes.fromhex(p["new"])
    cur = bytes(exe[off:off+len(orig_b)])
    assert cur == orig_b, f"offset {off:#x} 原 bytes {cur.hex()} != 預期 {p['orig']}"
    exe[off:off+len(new_b)] = new_b
    applied.append(f"Bug{p['bug']} {p['name']} @{off:#x} {p['orig']}->{p['new']}")

os.makedirs("work", exist_ok=True)
open("work/DQ3_fixed.EXE", "wb").write(exe)
print("EXE patch 套用:")
for a in applied: print("  ", a)

# SHP:需先有 work/DQ3MNS_fixed.SHP
if not os.path.exists("work/DQ3MNS_fixed.SHP"):
    print("  (產生 DQ3MNS_fixed.SHP …)")
    subprocess.run([sys.executable, "tools/make_sprites.py"], check=True)

# 組可跑的遊戲目錄:原素材 + 修正 EXE/SHP
GD = "work/dq3_fixed_game"
if os.path.exists(GD): shutil.rmtree(GD)
shutil.copytree("assets_raw", GD)
shutil.copy("work/DQ3_fixed.EXE", os.path.join(GD, "DQ3.EXE"))
shutil.copy("work/DQ3MNS_fixed.SHP", os.path.join(GD, "DQ3MNS.SHP"))
print(f"\n修正版遊戲目錄: {GD}/  (DQ3.EXE patched + DQ3MNS.SHP 復原 sprite)")
print("Bug3 處理:SHP 內 id128 歐里狄加 / id129 五頭龍大王 已補有效 sprite,blit 不再爆走")
