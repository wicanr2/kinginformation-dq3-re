#!/usr/bin/env python3
"""把所有 D3TXT*.TXT 解析成結構化資料 + 全檔驗證圖。

輸出:
  - assets_out/txtNN_full.png      : 整檔逐記錄 render(驗證用)
  - docs/data/d3txt_codes.json      : 每檔每記錄的 glyph index 序列(可追溯)
用法: text_dump.py
"""
import struct, json, os
from PIL import Image

RAW = 'assets_raw'
OUT = 'assets_out'
DATA = 'docs/data'
FON = f'{RAW}/D3TXT00.FON'
GB = 32
GW = GH = 16


def load_fon():
    return open(FON, 'rb').read()


def glyph_img(blob, idx):
    base = idx * GB
    img = Image.new('L', (GW, GH), 255)
    px = img.load()
    if base + GB > len(blob):
        return img
    for y in range(GH):
        row = (blob[base + y*2] << 8) | blob[base + y*2 + 1]
        for x in range(GW):
            if (row >> (15 - x)) & 1:
                px[x, y] = 0
    return img


def parse(txt):
    d = open(txt, 'rb').read()
    first = struct.unpack('<H', d[0:2])[0]
    ptrs = [struct.unpack('<H', d[i:i+2])[0] for i in range(0, first, 2)]
    recs = []
    for i in range(len(ptrs) - 1):
        s, e = ptrs[i], ptrs[i+1]
        raw = d[s:e]
        n = len(raw) - (len(raw) % 2)
        codes = [struct.unpack('<H', raw[j:j+2])[0] for j in range(0, n, 2)]
        term = codes and codes[-1] == 0xffff
        while codes and codes[-1] == 0xffff:
            codes.pop()
        recs.append({'i': i, 'off': s, 'len': len(raw), 'term': term, 'codes': codes})
    return recs, first


def render_full(blob, recs, outpath, scale=1):
    sel = [r['codes'] for r in recs]
    maxlen = max((len(c) for c in sel), default=1)
    W = maxlen * GW * scale + 2
    rowh = GH * scale
    H = len(sel) * rowh + 2
    canvas = Image.new('L', (W, H), 255)
    for ri, codes in enumerate(sel):
        for ci, code in enumerate(codes):
            g = glyph_img(blob, code)
            if scale != 1:
                g = g.resize((GW*scale, GH*scale))
            canvas.paste(g, (1 + ci*GW*scale, 1 + ri*rowh))
    canvas.save(outpath)


def main():
    os.makedirs(OUT, exist_ok=True)
    os.makedirs(DATA, exist_ok=True)
    blob = load_fon()
    allrec = {}
    for n in range(10):
        txt = f'{RAW}/D3TXT{n:02d}.TXT'
        if not os.path.exists(txt):
            continue
        recs, first = parse(txt)
        render_full(blob, recs, f'{OUT}/txt{n:02d}_full.png')
        allrec[f'D3TXT{n:02d}'] = {
            'table_bytes': first,
            'num_records': len(recs),
            'records': [{'i': r['i'], 'off': r['off'], 'term': r['term'], 'codes': r['codes']} for r in recs],
        }
        print(f'D3TXT{n:02d}: {len(recs)} records, table={first}B -> txt{n:02d}_full.png')
    with open(f'{DATA}/d3txt_codes.json', 'w') as f:
        json.dump(allrec, f, ensure_ascii=False)
    print(f'codes -> {DATA}/d3txt_codes.json')


if __name__ == '__main__':
    main()
