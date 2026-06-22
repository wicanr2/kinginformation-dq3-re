#!/usr/bin/env python3
"""TIT*.P 解碼 = 標準 ZSoft PCX(640×350, 4 plane, 1bpp/plane, RLE)。

表頭 128 byte 即標準 PCX header:
  0a 05 01 01 | xmin ymin xmax ymax(u16×4) | hdpi vdpi |
  offset 16: 16 色 EGA palette(48 byte) |
  offset 64: reserved | nplanes(=4) | bytesPerLine(u16,=80) | ...
RLE:若 (b & 0xC0)==0xC0 → count=b&0x3F, 下一 byte 為值, 重複 count 次;否則 b 為單一 literal。
每掃描列 = nplanes(4) × bytesPerLine(80) = 320 byte;4 個 plane 依序排(plane-sequential per row),
組回 4bpp:pixel = bit(p0)|bit(p1)<<1|bit(p2)<<2|bit(p3)<<3。
調色盤取自 header offset 16 的 16 色(各檔自帶,不用 DQ3.PAL)。

用法: tools/dockrun.sh tools/title_pcx.py <IN.P> <OUT.png>
"""
import os, sys, struct
from PIL import Image


def decode_pcx(path):
    d = open(path, 'rb').read()
    assert d[0] == 0x0A, 'not PCX'
    xmin, ymin, xmax, ymax = struct.unpack_from('<4H', d, 4)
    W = xmax - xmin + 1
    H = ymax - ymin + 1
    nplanes = d[65]
    bpl = struct.unpack_from('<H', d, 66)[0]
    # 16 色 EGA palette @ offset 16
    pal = [tuple(d[16 + i * 3: 16 + i * 3 + 3]) for i in range(16)]

    row_bytes = nplanes * bpl
    total = H * row_bytes
    out = bytearray()
    i = 128  # data after header
    n = len(d)
    while len(out) < total and i < n:
        b = d[i]; i += 1
        if (b & 0xC0) == 0xC0:
            cnt = b & 0x3F
            if i >= n:
                break
            val = d[i]; i += 1
            out.extend([val] * cnt)
        else:
            out.append(b)
    out = bytes(out[:total])

    im = Image.new('P', (W, H))
    flat = []
    for c in pal:
        flat += list(c)
    while len(flat) < 768:
        flat += [0, 0, 0]
    im.putpalette(flat)
    px = im.load()
    for y in range(H):
        base = y * row_bytes
        for xb in range(bpl):
            planes = [out[base + p * bpl + xb] for p in range(nplanes)]
            for bit in range(8):
                v = 0
                for p in range(nplanes):
                    v |= ((planes[p] >> (7 - bit)) & 1) << p
                x = xb * 8 + bit
                if x < W:
                    px[x, y] = v
    return im, (W, H), pal


if __name__ == '__main__':
    inp = sys.argv[1]
    outp = sys.argv[2]
    im, size, pal = decode_pcx(inp)
    os.makedirs(os.path.dirname(outp) or '.', exist_ok=True)
    im.convert('RGB').save(outp)
    print('saved', outp, size)
