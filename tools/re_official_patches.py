#!/usr/bin/env python3
"""青衫攻略「修改:DQ3.EXE」8 段官方 EXE patch 的解析、定位與驗證工具。

來源:references/勇者鬥惡龍3:傳說的終章.html「修改:DQ3.EXE」一節(本機,git 排除)。
每段 patch 的格式是「search pattern(原 bytes)」+「mask 行(-- 保留 / 兩碼 = 新值)」。
本工具:
  1. 把 8 段 patch 硬編為 (pattern_bytes, new_bytes_with_None) 結構。
  2. 在 assets_raw/DQ3.EXE 內搜 pattern(唯一命中)→ 換算每個被改 byte 的精確 file_offset。
  3. 印出每段:命中 file_offset、原 byte、新 byte、長度檢查(同長度 in-place)。
  4. dump official_patches.json(只收 in-place 同長度 patch)。

不污染他人 tools:本檔專供 docs/21 官方 patch 分析,不改 re_bug*/build_fixed_version。
跑法(docker uv venv):見 docs/21-official-patches.md。
"""
import json
import struct
import sys

EXE = "assets_raw/DQ3.EXE"
OUT_JSON = "docs/data/official_patches.json"

d = open(EXE, "rb").read()
e_cparhdr = struct.unpack_from("<H", d, 8)[0]
HDR = e_cparhdr * 16  # 0x1370


def parse_pat(orig_line, mask_lines):
    """orig_line: 'BE E8 80';  mask_lines: 連接後與 orig 等長, '--' = 保留, 兩碼 = 新值。
    回 (pat[list of int], new[list of int|None])。"""
    o = orig_line.split()
    m = " ".join(mask_lines).split()
    assert len(o) == len(m), f"len mismatch: {o} vs {m}"
    pat = [int(x, 16) for x in o]
    new = [None if x == "--" else int(x, 16) for x in m]
    return pat, new


# 8 段官方 patch。多行 pattern 的(4/5/6/7)各段分別搜尋(原 HTML 把 pattern 拆成兩塊,
# 兩塊在 EXE 內相鄰但各自是獨立 byte 串,逐塊定位最穩)。
PATCHES = [
    {
        "id": 1, "title": "聲霸卡防當法",
        "bug": None, "note": "Sound Blaster 偵測防當機(開機/音效初始化路徑,非遊戲性 bug)",
        "blocks": [("BE E8 80", ["-- F0 77"])],
    },
    {
        "id": 2, "title": "魔王打不死修正法",
        "bug": 1, "note": "巴拉摩斯勝負判定:clamp 參照欄 [bx+0x2336]->[bx+0x2334],兩處 36->34",
        "blocks": [("39 87 36 23 73 04 8B 87 36 23",
                    ["-- -- 34 -- -- -- -- -- 34 --"])],
    },
    {
        "id": 3, "title": "敵人消失法",
        "bug": None, "note": "金手指類:把某遞減+條件跳轉 NOP 掉使敵人不出現(作弊,非修 bug)",
        "blocks": [("FE 0E F4 52 75 30 F7", ["90 90 90 90 EB -- --"])],
    },
    {
        "id": 4, "title": "五頭龍大王當機修正",
        "bug": 3, "note": "boss 遭遇表換怪:0x80->0x7e / 0x81->0x7d(第一場)、0x81->0x7c(第二場)",
        "blocks": [
            ("80 01 81 01 01 2D", ["7E -- 7D -- -- --"]),
            ("81 01 01 2A 08", ["7C -- -- -- --"]),
        ],
    },
    {
        "id": 5, "title": "修正某些機器上跑會當機的問題",
        "bug": None, "note": "相容性:某些機器 lcall/分支當機,改 NOP / 短跳(環境相容,非遊戲性 bug)",
        "blocks": [
            ("3D 00 00 74 1D 9A", ["-- -- -- 90 90 --"]),
            ("3D 00 00 74 02 EB EA", ["-- -- -- EB -- -- --"]),
        ],
    },
    {
        "id": 6, "title": "穿牆",
        "bug": None, "note": "金手指:把碰撞旗標檢查改成永遠通過(作弊,非修 bug)",
        "blocks": [
            ("F7 C3 01 00 75 71", ["-- -- -- -- -- 00"]),
            ("F7 C3 01 00 75 6D", ["-- -- -- -- -- 00"]),
            ("F6 C4 0F 74 06", ["-- -- -- EB --"]),
        ],
    },
    {
        "id": 7, "title": "致命一擊",
        "bug": None, "note": "金手指:改攻擊判定使必出致命一擊(作弊,非修 bug)",
        "blocks": [
            ("3C 0A 77 39 53 57 56", ["-- -- -- 00 -- -- --"]),
            ("32 FF 8B 6D 07 2B EB", ["8B 5D 07 8B EB -- --"]),
        ],
    },
    {
        "id": 8, "title": "生命不減",
        "bug": None, "note": "金手指:把 HP 寫回 NOP 掉使生命不減(作弊,非修 bug)",
        "blocks": [("89 6D 16 EB 04", ["90 90 90 -- --"])],
    },
]


def find_unique(pat):
    """在 EXE 找 pat 的所有命中位置(list of file_offset)。"""
    hits = []
    n = len(pat)
    pb = bytes(pat)
    start = 0
    while True:
        i = d.find(pb, start)
        if i < 0:
            break
        hits.append(i)
        start = i + 1
    return hits


def main():
    json_patches = []
    for p in PATCHES:
        print(f"\n=== Patch {p['id']}: {p['title']}  (bug={p['bug']}) ===")
        print(f"    {p['note']}")
        for orig_line, mask_lines in p["blocks"]:
            pat, new = parse_pat(orig_line, mask_lines)
            hits = find_unique(pat)
            if len(hits) == 0:
                print(f"  [MISS] pattern not found: {orig_line}")
                continue
            if len(hits) > 1:
                print(f"  [WARN] {len(hits)} 命中 (非唯一): " +
                      ", ".join(f"{h:#x}" for h in hits))
            base = hits[0]
            print(f"  pattern @ file {base:#x} (seg0 {base - HDR:#x})  唯一={len(hits)==1}")
            for k, (o, nn) in enumerate(zip(pat, new)):
                if nn is None:
                    continue
                off = base + k
                cur = d[off]
                ok = "OK" if cur == o else f"MISMATCH(實際 {cur:#04x})"
                print(f"    file {off:#06x}: {o:#04x} -> {nn:#04x}  [{ok}]")
                if p["bug"] is not None and cur == o:
                    json_patches.append({
                        "official_id": p["id"],
                        "title": p["title"],
                        "bug": p["bug"],
                        "file_offset": off,
                        "file_offset_hex": f"{off:#x}",
                        "orig": f"{o:02x}",
                        "new": f"{nn:02x}",
                        "kind": "in-place",
                        "source": "青衫官方",
                        "desc": p["note"],
                    })

    out = {
        "source_exe": EXE,
        "exe_size": len(d),
        "hdr": HDR,
        "note": ("青衫攻略「修改:DQ3.EXE」8 段官方 EXE patch,經反組譯交叉驗證。"
                 "只收『修 bug』類同長度 in-place patch(對映已知 bug);"
                 "金手指/相容性類(敵人消失/穿牆/致命一擊/生命不減/聲霸/相容)不收進修正版,"
                 "完整 8 段見 docs/21-official-patches.md。file_offset = EXE 內絕對位移。"),
        "patches": json_patches,
    }
    if len(sys.argv) > 1 and sys.argv[1] == "write":
        json.dump(out, open(OUT_JSON, "w", encoding="utf-8"),
                  ensure_ascii=False, indent=2)
        print(f"\n寫出 {OUT_JSON}({len(json_patches)} 個 in-place 修 bug patch)")
    else:
        print("\n(加 'write' 參數可寫出 docs/data/official_patches.json)")
        print(json.dumps(out, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
