#!/usr/bin/env python3
"""render FON glyph atlas with index labels, to identify low-index punctuation/control glyphs.
用法: text_atlas.py <start> <count> <out.png>
"""
import sys
from PIL import Image, ImageDraw

FON = 'assets_raw/D3TXT00.FON'
GB = 32
GW = GH = 16


def glyph_img(blob, idx, scale=2):
    base = idx * GB
    img = Image.new('L', (GW, GH), 255)
    px = img.load()
    if base + GB <= len(blob):
        for y in range(GH):
            row = (blob[base + y*2] << 8) | blob[base + y*2 + 1]
            for x in range(GW):
                if (row >> (15 - x)) & 1:
                    px[x, y] = 0
    return img.resize((GW*scale, GH*scale))


def main():
    start = int(sys.argv[1]); count = int(sys.argv[2]); out = sys.argv[3]
    blob = open(FON, 'rb').read()
    cols = 16
    cellw, cellh = 60, 60
    rows = (count + cols - 1) // cols
    canvas = Image.new('RGB', (cols*cellw, rows*cellh), (255, 255, 255))
    dr = ImageDraw.Draw(canvas)
    for i in range(count):
        idx = start + i
        cx = (i % cols) * cellw
        cy = (i // cols) * cellh
        g = glyph_img(blob, idx).convert('RGB')
        canvas.paste(g, (cx+12, cy+16))
        dr.text((cx+2, cy+2), f'{idx:02x}', fill=(200, 0, 0))
    canvas.save(out)
    print(f'atlas {start}..{start+count} -> {out}')


if __name__ == '__main__':
    main()
