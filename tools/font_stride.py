#!/usr/bin/env python3
"""前 N 列窗格:sec2 在 w=16 與 w=24 並排,放大,標列號,找 header 邊界與字模對齊。"""
import struct, os
from PIL import Image, ImageDraw

SRC = "assets_raw/CHINA.FON"
OUT = "assets_out/font_probe"
os.makedirs(OUT, exist_ok=True)
data = open(SRC, "rb").read()
first = struct.unpack_from("<I", data, 0)[0]
n = first // 4
offs = [struct.unpack_from("<I", data, i*4)[0] for i in range(n)] + [len(data)]

def strip(blob, w_bits, rows, start_byte=0, msb=True):
    bpr = (w_bits + 7)//8
    img = Image.new("L", (w_bits, rows), 0)
    px = img.load()
    for y in range(rows):
        for x in range(w_bits):
            idx = start_byte + y*bpr + x//8
            if idx >= len(blob): continue
            b = blob[idx]
            bit = (b>>(7-(x&7)))&1 if msb else (b>>(x&7))&1
            if bit: px[x,y]=255
    return img

blob = data[offs[2]:offs[3]]
ROWS = 240
scale = 6
panels = []
for w in (16, 24):
    im = strip(blob, w, ROWS)
    big = im.resize((im.width*scale, im.height*scale), Image.NEAREST)
    lab = Image.new("L", (big.width+44, big.height), 24)
    lab.paste(big, (44,0))
    d = ImageDraw.Draw(lab)
    for r in range(0, ROWS, 8):
        y=r*scale
        d.line([(40,y),(44,y)], fill=110)
        d.text((2,y), str(r), fill=170)
    d.text((46,2), f"w={w}", fill=255)
    panels.append(lab)
gap=20
W=sum(p.width for p in panels)+gap*(len(panels)+1)
H=max(p.height for p in panels)
sheet=Image.new("L",(W,H),16)
x=gap
for p in panels:
    sheet.paste(p,(x,0)); x+=p.width+gap
sheet.save(f"{OUT}/window_sec2.png")
print("saved window_sec2.png  rows:", ROWS)
