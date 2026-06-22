#!/usr/bin/env python3
"""精訊 DQ3 數值類 bug(#4 勇者 MP / #5 升級錯亂 / #6 數值溢位)修正分析 + patch 產生器。

不污染他人 tools:本檔專供 docs/23 數值修正,只動 stat_patches.json / work/。

關鍵反組譯定位(本 agent 自行核出,修正 docs/18 的兩處誤判):
- 成長表「不在 dragon0.dat」:dragon0.dat 是存檔(2172 byte,多為 0xff)。
  成長表是 EXE 靜態初始化資料,在 DGROUP(seg 0x14dd)DS:0x4366,
  → file 0x1a4a6,8 個職業 × 14 byte/列。
- 每職業 14 byte 列(成長公式 sub_d9cc file 0xed3c 解出):
    +0 HPbase +1 HPslope  +2 MPbase +3 MPslope
    +4 b4base +5 b4slope   +6 b6base +7 b6slope
    +8 b8base +9 b8slope   +0xa bAbase +0xb bAslope  +0xc,+0xd 其他
  公式:target = base + (slope*level)>>1(全程 16-bit:mul→ax, shr ax,1, add al/adc ah,0),
        實得增量 = sub_fa57(target-current) = rng%(target-current),下限 1。
- 職業序(D3TXT00 0x1c9..):0勇者 1戰士 2武鬥家 3僧侶 4魔法使者 5賢者 6商人 7遊玩者。
- #6 修正 docs/18:成長公式 target 計算是 16-bit 正確(add al + adc ah 正確進位),
  屬性欄位 [si+0x24]/[si+0x26] 全程 word 讀寫(無 8a/88 byte 截斷)。255-wrap 不在此處,
  較可能是顯示寬度或與 #5 越界垃圾互動,需 runtime 觀察 → 留 C 層。
- #5:升級門檻表 8 職業各 44 entry(lv0..43),MAX_LEVEL=43;lv44 越界讀到下一職業
  entry[0]=0 → exp 永遠 ≥0 → 連升 + 學錯職業咒。clamp 需 code cave;
  re_codecave_scan 全檔最大 cave 僅 8 byte,放不下安全 trampoline(需 ≥11~19 byte),
  依安全優先原則不冒崩潰風險做 → 留 C 層。

用法:
  re_stat_patch.py table          dump 成長表 8 職業列
  re_stat_patch.py thresholds     dump 各職業升級門檻表長度(證 MAX_LEVEL=43)
  re_stat_patch.py build          產 work/DQ3_stat_fixed.EXE(只套 #4 安全 in-place patch)
"""
import json, os, struct, sys

EXE = "assets_raw/DQ3.EXE"
HDR = 0x1370
DGROUP = 0x14dd
GROWTH_DS = 0x4366                       # 成長表 DS off
GROWTH_FILE = HDR + DGROUP * 16 + GROWTH_DS   # 0x1a4a6
THR_PTR_DS = 0x43d6

# #4 勇者(class0)MP:base@列+2、slope@列+3。原 base=3 slope=5 → 提到 base=8 slope=10。
HERO_MP_BASE_OFF = GROWTH_FILE + 0 * 0xe + 2   # 0x1a4a8
HERO_MP_SLOPE_OFF = GROWTH_FILE + 0 * 0xe + 3  # 0x1a4a9
STAT_PATCHES = [
    {"bug": 4, "name": "勇者MaxMP成長偏低", "file_offset": HERO_MP_BASE_OFF,
     "orig": "03", "new": "08", "kind": "exe-data",
     "desc": "成長表 勇者(class0)MP base 0x03->0x08;EXE 靜態 DGROUP 表,非 dragon0.dat"},
    {"bug": 4, "name": "勇者MaxMP成長偏低", "file_offset": HERO_MP_SLOPE_OFF,
     "orig": "05", "new": "0a", "kind": "exe-data",
     "desc": "成長表 勇者(class0)MP slope 0x05->0x0a;lv43 MaxMP ~107->~217,8-bit 無 wrap"},
]


def _d():
    return open(EXE, "rb").read()


def table():
    d = _d()
    names = "勇者 戰士 武鬥家 僧侶 魔法使者 賢者 商人 遊玩者".split()
    print("成長表 file 0x%05x (DS:0x%04x, 8 職業 × 14 byte):" % (GROWTH_FILE, GROWTH_DS))
    print("列格式: HPb HPs MPb MPs b4b b4s b6b b6s b8b b8s bAb bAs c d")
    for i in range(8):
        rec = d[GROWTH_FILE + i * 0xe: GROWTH_FILE + i * 0xe + 0xe]
        print("  %d %-8s: %s" % (i, names[i], " ".join("%02x" % b for b in rec)))


def thresholds():
    d = _d()
    pt = HDR + DGROUP * 16 + THR_PTR_DS
    ptrs = [struct.unpack_from('<H', d, pt + i * 2)[0] for i in range(8)]
    print("升級門檻表 8 職業(證 MAX_LEVEL=43):")
    for i in range(7):
        n = (ptrs[i + 1] - ptrs[i]) // 4
        print("  class%d: %d entries (lv0..%d)" % (i, n, n - 1))
    print("  class7: 44 entries (lv0..43) [尾,長度同前]")
    print("→ lv44 越界讀下一職業 entry[0]=0 → 連升(#5)。clamp 需 code cave。")


def build():
    exe = bytearray(_d())
    applied = []
    for p in STAT_PATCHES:
        off = p["file_offset"]
        ob = bytes.fromhex(p["orig"]); nb = bytes.fromhex(p["new"])
        cur = bytes(exe[off:off + len(ob)])
        assert cur == ob, "offset %#x 原 byte %s != 預期 %s" % (off, cur.hex(), p["orig"])
        exe[off:off + len(nb)] = nb
        applied.append("Bug%s %s @%#x %s->%s" % (p["bug"], p["name"], off, p["orig"], p["new"]))
    os.makedirs("work", exist_ok=True)
    open("work/DQ3_stat_fixed.EXE", "wb").write(exe)
    print("已套用 stat patch -> work/DQ3_stat_fixed.EXE:")
    for a in applied:
        print("  ", a)


def main():
    cmd = sys.argv[1] if len(sys.argv) > 1 else "table"
    {"table": table, "thresholds": thresholds, "build": build}.get(cmd, table)()


if __name__ == "__main__":
    main()
