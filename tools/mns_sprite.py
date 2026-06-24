#!/usr/bin/env python3
"""DQ3MNS.SHP 怪物 sprite 解碼 + 圖鑑 render。

格式(由 DQ3.EXE sub_b19e / sub_b2af 反組譯確認):
  檔首:131 個 u32 LE = offset table(每筆 = 該 sprite 資料在檔內的絕對位移)。
        sprite_id 資料 = file[ offs[id] : offs[id+1] ]。
  sprite 資料區(sub_b2af 色彩 blit):
    +0  u16 w   (每 plane 每 row 的 byte 數;像素寬 = w*8;top bit 0x8000 為旗標)
    +2  u16 h   (row 數 = 像素高)
    +4  4 個 plane 連續(plane-major):VGA Map Mask ah 由 0x08→0x01,
        故第 1 段對應 bit-plane 3、最後段對應 bit-plane 0。每段 h 行 × w byte,
        每 byte = 8 像素的該 plane bit(MSB-first)。色號 = Σ bit<<plane。
"""
import sys, struct
from PIL import Image, ImageDraw

SHP = "assets_raw/DQ3MNS.SHP"
PAL = "assets_raw/MNSBK.PAL"


def load_pal(path):
    d = open(path, 'rb').read()
    cols = []
    for i in range(0, len(d) - 2, 3):
        r, g, b = d[i], d[i+1], d[i+2]
        cols.append(((r << 2) | (r >> 4), (g << 2) | (g >> 4), (b << 2) | (b >> 4)))
    while len(cols) < 256:
        cols.append((0, 0, 0))
    return cols


def read_offsets(m):
    o0 = struct.unpack_from('<I', m, 0)[0]
    n = o0 // 4
    return [struct.unpack_from('<I', m, i * 4)[0] for i in range(n)]


def decode_sprite(m, offs, idx, plane_order="hi2lo"):
    start = offs[idx]
    end = offs[idx + 1]
    data = m[start:end]
    if len(data) < 4:
        return None
    w_raw, h = struct.unpack_from('<HH', data, 0)
    w = w_raw & 0x7fff
    if w == 0 or h == 0 or w > 64 or h > 256:
        return None
    W = w * 8
    grid = [[0] * W for _ in range(h)]
    pos = 4
    for seg in range(4):
        if plane_order == "hi2lo":
            plane_bit = 3 - seg
        else:
            plane_bit = seg
        for row in range(h):
            for bx in range(w):
                if pos >= len(data):
                    return W, h, grid, w_raw
                byte = data[pos]; pos += 1
                if byte:
                    col0 = bx * 8
                    for bit in range(8):
                        if byte & (0x80 >> bit):
                            grid[row][col0 + bit] |= (1 << plane_bit)
    return W, h, grid, w_raw


def render_one(m, offs, pal, idx, transparent=0, plane_order="hi2lo"):
    r = decode_sprite(m, offs, idx, plane_order)
    if r is None:
        return None
    W, h, grid, w_raw = r
    img = Image.new('RGBA', (W, h), (0, 0, 0, 0))
    px = img.load()
    for y in range(h):
        for x in range(W):
            c = grid[y][x]
            if c == transparent:
                continue
            r8, g8, b8 = pal[c]
            px[x, y] = (r8, g8, b8, 255)
    return img


def main():
    # 若已生成含復原 128/129(歐里狄加/五頭龍大王)的 fixed SHP,優先用它,補完圖鑑空槽。
    import os.path
    src = "work/DQ3MNS_fixed.SHP" if os.path.exists("work/DQ3MNS_fixed.SHP") else SHP
    m = open(src, 'rb').read()
    offs = read_offsets(m)
    pal = load_pal(PAL)
    nspr = len(offs) - 1
    po = "hi2lo"
    args = sys.argv[1:]
    if "lo2hi" in args:
        po = "lo2hi"; args = [a for a in args if a != "lo2hi"]
    print(f"; SHP sprites = {nspr}  plane_order={po}")

    if args and args[0] == "one":
        idx = int(args[1])
        img = render_one(m, offs, pal, idx, plane_order=po)
        if img:
            img = img.resize((img.width * 3, img.height * 3), Image.NEAREST)
            out = args[2] if len(args) > 2 else f"docs/monsters/spr_{idx:03d}.png"
            img.save(out)
            print(f"saved {out} ({img.width}x{img.height})")
        return

    cols = 13
    cell_w, cell_h = 90, 130
    rows = (nspr + cols - 1) // cols
    sheet = Image.new('RGBA', (cols * cell_w, rows * cell_h), (40, 40, 60, 255))
    dr = ImageDraw.Draw(sheet)
    ok = 0
    for idx in range(nspr):
        img = render_one(m, offs, pal, idx, plane_order=po)
        cx = (idx % cols) * cell_w
        cy = (idx // cols) * cell_h
        dr.rectangle([cx, cy, cx + cell_w - 1, cy + cell_h - 1], outline=(80, 80, 100))
        dr.text((cx + 2, cy + 1), str(idx), fill=(255, 255, 120))
        if img:
            ok += 1
            if img.width > cell_w or img.height > cell_h - 14:
                sc = min(cell_w / img.width, (cell_h - 14) / img.height)
                img = img.resize((max(1, int(img.width * sc)), max(1, int(img.height * sc))), Image.NEAREST)
            ox = cx + max(0, (cell_w - img.width) // 2)
            oy = cy + 14 + max(0, (cell_h - 14 - img.height) // 2)
            sheet.alpha_composite(img, (ox, oy))
    sheet.save("docs/monsters/monster_sheet.png")
    print(f"saved docs/monsters/monster_sheet.png  ({ok}/{nspr} decoded)")


if __name__ == "__main__":
    main()
