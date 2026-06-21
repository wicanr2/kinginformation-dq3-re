#!/usr/bin/env python3
"""把 D3TXT*.TXT 的記錄(2-byte glyph index 序列)用 D3TXT00.FON 字模 render 成圖。

每筆記錄 render 成一行(每字 16x16),整檔輸出成一張縱向拼接圖。
用法: text_render.py <TXT檔> <輸出PNG> [起始記錄] [記錄數]
"""
import struct, sys
from PIL import Image

FON = 'assets_raw/D3TXT00.FON'
GLYPH_BYTES = 32
GW = GH = 16


def load_glyphs():
    blob = open(FON, 'rb').read()
    n = len(blob) // GLYPH_BYTES
    return blob, n


def render_glyph(blob, idx):
    """回傳 16x16 的 0/1 list,超出範圍回傳空白。"""
    base = idx * GLYPH_BYTES
    img = Image.new('1', (GW, GH), 0)
    px = img.load()
    if base + GLYPH_BYTES > len(blob):
        return img  # 空白
    for y in range(GH):
        row = (blob[base + y*2] << 8) | blob[base + y*2 + 1]
        for x in range(GW):
            if (row >> (15 - x)) & 1:
                px[x, y] = 1
    return img


def parse_records(txt):
    d = open(txt, 'rb').read()
    first = struct.unpack('<H', d[0:2])[0]
    ptrs = [struct.unpack('<H', d[i:i+2])[0] for i in range(0, first, 2)]
    recs = []
    for i in range(len(ptrs) - 1):
        s, e = ptrs[i], ptrs[i+1]
        raw = d[s:e]
        codes = [struct.unpack('<H', raw[j:j+2])[0] for j in range(0, len(raw) - (len(raw) % 2), 2)]
        # 去尾端 0xffff 終止碼
        while codes and codes[-1] == 0xffff:
            codes.pop()
        recs.append(codes)
    return recs


def main():
    txt = sys.argv[1]
    out = sys.argv[2]
    start = int(sys.argv[3]) if len(sys.argv) > 3 else 0
    count = int(sys.argv[4]) if len(sys.argv) > 4 else 60
    blob, nglyph = load_glyphs()
    recs = parse_records(txt)
    sel = recs[start:start+count]
    maxlen = max((len(r) for r in sel), default=1)
    scale = 2
    W = maxlen * GW * scale + 4
    rowh = GH * scale + 2
    H = len(sel) * rowh + 4
    canvas = Image.new('L', (W, H), 255)
    for ri, codes in enumerate(sel):
        for ci, code in enumerate(codes):
            g = render_glyph(blob, code).resize((GW*scale, GH*scale))
            # paste: glyph 1 = black
            gg = g.point(lambda v: 0 if v else 255, 'L')
            canvas.paste(gg, (2 + ci*GW*scale, 2 + ri*rowh))
    canvas.save(out)
    print(f'glyphs={nglyph} records={len(recs)} rendered={len(sel)} start={start} -> {out}')


if __name__ == '__main__':
    main()
