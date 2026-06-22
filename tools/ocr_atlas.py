#!/usr/bin/env python3
"""把 D3TXT00.FON 全部字模 render 成帶 index 標籤的分塊大圖,供 vision 逐塊辨識。

每塊預設 64 字(8 列 x 8 行),scale 大,每格上方標 index(十進位)。
用法: ocr_atlas.py [start] [count] [out.png] [cols] [scale]
不帶參數則一次輸出全部塊到 assets_out/ocr/block_NNNN.png (每塊 64 字)。
"""
import sys, os
from PIL import Image, ImageDraw, ImageFont

FON = 'assets_raw/D3TXT00.FON'
GB = 32
GW = GH = 16


def glyph_img(blob, idx, scale):
    base = idx * GB
    img = Image.new('L', (GW, GH), 255)
    px = img.load()
    if base + GB <= len(blob):
        for y in range(GH):
            row = (blob[base + y*2] << 8) | blob[base + y*2 + 1]
            for x in range(GW):
                if (row >> (15 - x)) & 1:
                    px[x, y] = 0
    return img.resize((GW*scale, GH*scale), Image.NEAREST)


def get_font(size):
    for p in ['/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf',
              '/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf']:
        if os.path.exists(p):
            return ImageFont.truetype(p, size)
    return ImageFont.load_default()


def render_block(blob, start, count, out, cols=8, scale=5):
    gpx = GW*scale
    labelh = 18
    pad = 6
    cellw = gpx + pad
    cellh = gpx + labelh + pad
    rows = (count + cols - 1) // cols
    canvas = Image.new('RGB', (cols*cellw + pad, rows*cellh + pad), (255, 255, 255))
    dr = ImageDraw.Draw(canvas)
    fnt = get_font(13)
    for i in range(count):
        idx = start + i
        if idx >= len(blob)//GB:
            break
        cx = pad + (i % cols) * cellw
        cy = pad + (i // cols) * cellh
        dr.text((cx, cy), str(idx), fill=(200, 0, 0), font=fnt)
        g = glyph_img(blob, idx, scale).convert('RGB')
        canvas.paste(g, (cx, cy + labelh))
        # cell border
        dr.rectangle([cx-1, cy+labelh-1, cx+gpx, cy+labelh+gpx], outline=(200, 200, 200))
    canvas.save(out)
    return out


def main():
    blob = open(FON, 'rb').read()
    nglyph = len(blob) // GB
    if len(sys.argv) >= 4:
        start = int(sys.argv[1]); count = int(sys.argv[2]); out = sys.argv[3]
        cols = int(sys.argv[4]) if len(sys.argv) > 4 else 8
        scale = int(sys.argv[5]) if len(sys.argv) > 5 else 5
        render_block(blob, start, count, out, cols, scale)
        print(f'block {start}..{start+count} -> {out}')
    else:
        os.makedirs('assets_out/ocr', exist_ok=True)
        per = 64
        for start in range(0, nglyph, per):
            cnt = min(per, nglyph - start)
            out = f'assets_out/ocr/block_{start:04d}.png'
            render_block(blob, start, cnt, out, cols=8, scale=5)
        print(f'glyphs={nglyph} blocks={ (nglyph+per-1)//per } -> assets_out/ocr/')


if __name__ == '__main__':
    main()
