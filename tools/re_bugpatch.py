#!/usr/bin/env python3
"""re_bugpatch.py — 精訊版 DQ3 bug 修正版 binary patch 產生器。

讀 assets_raw/DQ3.EXE(唯讀,不動原檔)→ 套用一組「同長度、不位移」的 in-place
patch → 輸出 work/DQ3_fixed.EXE(work/ 已 gitignore)。

每個 patch 都先驗證該 file offset 上的原 bytes 與預期相符(防錯位),不符即中止。
所有改碼均為同長度,不更動 MZ header / segment 佈局 / relocation table。

分析依據:docs/18-bug-analysis.md。對照青衫先生攻略
references/勇者鬥惡龍3:傳說的終章.html 內「修改:DQ3.EXE」官方修正碼,以反組譯逐點覆核。

用法:
  python tools/re_bugpatch.py            # 套用全部高信心 patch → work/DQ3_fixed.EXE
  python tools/re_bugpatch.py --verify   # 只驗證原 bytes,不輸出
"""
import sys, os

SRC = "assets_raw/DQ3.EXE"
OUT = "work/DQ3_fixed.EXE"

# 每筆 patch:(file_offset, expect_bytes, new_bytes, bug, why)
# expect/new 同長度;套用前先比對 expect 防錯位。
PATCHES = [
    # ---- Bug 1:巴拉摩斯(魔王)打不死 ----
    # sub_aa69(file 0xbdd9)的「敵方累積值 clamp」:
    #   cmp word[bx+0x2336],ax / mov ax,word[bx+0x2336]  把 [+0x2336] 改成 [+0x2334]。
    # 對應青衫官方碼「2.魔王打不死修正法」:39 87 36 23 .. 8B 87 36 23 → 把兩個 36→34。
    # 效果:修正魔王在被擊敗判定上的累積值參照欄位(原參照欄使勝負判定錯亂,全隊被吹飛卻判贏)。
    (0x00be04, b"\x36", b"\x34", "Bug1 魔王打不死",
     "cmp word[bx+0x2336]→[bx+0x2334](官方魔王打不死修正法 第1處)"),
    (0x00be0a, b"\x36", b"\x34", "Bug1 魔王打不死",
     "mov ax,word[bx+0x2336]→[bx+0x2334](官方魔王打不死修正法 第2處)"),

    # ---- Bug 2:彩虹水滴合成事件給錯成品 ----
    # 合成 handler(file 0x77e7):mov word[si],0x6b(銀寶珠)→ 0x75(彩虹水滴)。
    # 太陽之石(0x72)+ 雲雨之杖(0x73)應合成彩虹水滴(0x75),原碼成品 item code 寫錯。
    (0x0077e9, b"\x6b", b"\x75", "Bug2 彩虹水滴",
     "合成成品 item code 0x6b(銀寶珠)→ 0x75(彩虹水滴)"),

    # ---- Bug 3:五頭龍 / 歐里狄加 戰當機(缺 sprite)----
    # boss 遭遇腳本表(file 0x1b02c..)含 monster_id 0x80(歐里狄加)/0x81(五頭龍大王);
    # 此二怪的 DQ3MNS.SHP sprite header 為未完成的填充值(w=0xfff8/h=0xfff8)→ blit
    # 讀到天文數字 w/h 越界 → 當機。把這兩個 boss 換成有效 sprite 的怪(0x7c/0x7d/0x7e)。
    # 對應青衫官方碼「4.五頭龍大王當機修正」:80→7E、81→7D(第一場)、81→7C(第二場)。
    (0x01b02c, b"\x80", b"\x7e", "Bug3 五頭龍當機",
     "boss 表 歐里狄加 0x80(缺 sprite)→ 0x7e 冰河魔人(有效 sprite),第一場"),
    (0x01b02e, b"\x81", b"\x7d", "Bug3 五頭龍當機",
     "boss 表 五頭龍大王 0x81(缺 sprite)→ 0x7d 魔文(有效 sprite),第一場"),
    (0x01b038, b"\x81", b"\x7c", "Bug3 五頭龍當機",
     "boss 表 五頭龍大王 0x81(缺 sprite)→ 0x7c 索瑪(有效 sprite),第二場"),
]


def main():
    verify_only = "--verify" in sys.argv
    data = bytearray(open(SRC, "rb").read())
    print(f"# 讀入 {SRC}  size=0x{len(data):x}")

    ok = True
    for off, expect, new, bug, why in PATCHES:
        cur = bytes(data[off:off + len(expect)])
        status = "OK" if cur == expect else "MISMATCH!"
        if cur != expect:
            ok = False
        print(f"  [{status:9}] {bug:18} @0x{off:06x}: "
              f"{cur.hex()} (expect {expect.hex()}) -> {new.hex()}  # {why}")
        if cur == expect and not verify_only:
            data[off:off + len(new)] = new

    if not ok:
        print("\n!! 有 patch 原 bytes 不符預期(可能錯位 / 檔案版本不同),已中止,未輸出。")
        sys.exit(1)

    if verify_only:
        print("\n# --verify:全部原 bytes 符合預期。未輸出檔案。")
        return

    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    with open(OUT, "wb") as f:
        f.write(data)
    print(f"\n# 已輸出修正版 {OUT}  size=0x{len(data):x}(與原檔同長度)")
    print(f"# patch 數:{len(PATCHES)}  (全部同長度 in-place,未動 MZ / segment / reloc)")


if __name__ == "__main__":
    main()
