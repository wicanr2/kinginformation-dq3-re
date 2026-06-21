#!/usr/bin/env python3
"""CHINA.FON 字模抽取(run-based 對齊版)。

歷史:早期版本用「貪婪非重疊 +32/+1」掃描,並以為某些字被精訊「左右半對調混淆」。
後來程式驗證:那不是混淆,是 1-byte 對齊漂移(字模間插入 1 byte 使部分字落在奇相位,
讀晚 1 byte 就像左右半對調 + 下移)。deobf_rows(b) 數學上等於 read@(b-1)。
本版改用 run-based 對齊(見 tools/font_align.py):維持相位連成 run,偵測 1-byte 插入翻相位,
以 row14/row15 底部留白乾淨度為強訊號,讓正確相位的字優先成形。

用法:
  tools/dockrun.sh tools/font_greedy.py            # 全 22 段 -> CHINA.greedy.atlas.png
  tools/dockrun.sh tools/font_greedy.py 0          # 單段 -> greedy_sec00.png
  tools/dockrun.sh tools/font_greedy.py 0 1 2      # 指定數段,各自單段 render
"""
import os
import sys
from PIL import Image
import font_align as A

OUT = "assets_out/font"
os.makedirs(OUT, exist_ok=True)

data, offs = A.load()


def extract_section(si):
    """回傳 [(byte_off, rows) ...](相容舊呼叫端的 2-tuple 介面)。"""
    return A.extract_section(data, offs, si)


def atlas(glyphs, cols, scale, path):
    rows_n = (len(glyphs) + cols - 1) // cols
    pad = 1
    img = Image.new("L", (cols * (16 + pad) + pad, rows_n * (16 + pad) + pad), 40)
    px = img.load()
    for i, item in enumerate(glyphs):
        g = item[1]
        gx = (i % cols) * (16 + pad) + pad
        gy = (i // cols) * (16 + pad) + pad
        for y in range(16):
            for x in range(16):
                if (g[y] >> (15 - x)) & 1:
                    px[gx + x, gy + y] = 255
    img.resize((img.width * scale, img.height * scale), Image.NEAREST).save(path)


if __name__ == "__main__":
    secs = [int(x) for x in sys.argv[1:]] or list(range(22))
    allg = []
    for si in secs:
        gl = extract_section(si)
        allg += gl
        if len(secs) <= 3:
            atlas(gl, 24, 4, f"{OUT}/greedy_sec{si:02d}.png")
        print(f"sec{si:2}: {len(gl)} 字")
    print(f"合計: {len(allg)} 字")
    if len(secs) > 3:
        atlas(allg, 48, 3, f"{OUT}/CHINA.greedy.atlas.png")
        print("-> CHINA.greedy.atlas.png")
