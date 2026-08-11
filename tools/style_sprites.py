#!/usr/bin/env python3
"""以可重現、非破壞方式整理 128／129 的 MNS 像素風格。

這個工具不重新猜角色輪廓，也不改變 sprite 的透明遮罩或邏輯尺寸；它只在
既有回補 SHP 上做很小的 palette alias、輪廓群組與抖點調整，然後重建：

* ``work/DQ3MNS_fixed.SHP``
* ``dq3_remake_ebitan/mobile/assets/DQ3MNS.SHP``（embedded 執行版副本）
* ``dq3_remake/src/dq3_restored_sprites.c``
* ``dq3_remake_ebitan/internal/dq3data/restored_sprites.go``（fallback）
* ``docs/monsters/spr_128.png``、``spr_129.png``
* ``docs/monsters/monster_sheet.png``、``restored_128_129.png``

SHP 的 128／129 尾端也會寫入與 Go loader 相容的逐列 RLE AND-mask，讓有修正版
SHP 的執行版直接載入，不必依賴 fallback；`assets_raw/DQ3MNS.SHP` 永遠不會被覆寫。

本檔刻意只使用 Python 標準函式庫，讓沒有 Pillow 的 Docker 驗證 image 也能
重播。參考來源、外部重繪的限制與推論等級見 ``docs/121-monster-128-129-style-audit.md``。
"""

from __future__ import annotations

import binascii
import os
import struct
import zlib
from collections import Counter


SHP_IN = "work/DQ3MNS_fixed.SHP"
SHP_OUT = "work/DQ3MNS_fixed.SHP"
PAL_PATH = "assets_raw/MNSBK.PAL"
C_OUT = "dq3_remake/src/dq3_restored_sprites.c"
GO_OUT = "dq3_remake_ebitan/internal/dq3data/restored_sprites.go"
MOBILE_SHP_OUT = "dq3_remake_ebitan/mobile/assets/DQ3MNS.SHP"


def read_offsets(blob: bytes) -> list[int]:
    first = struct.unpack_from("<I", blob, 0)[0]
    count = first // 4
    return [struct.unpack_from("<I", blob, i * 4)[0] for i in range(count)]


def decode(blob: bytes, offsets: list[int], idx: int):
    data = blob[offsets[idx] : offsets[idx + 1]]
    if len(data) < 4:
        raise ValueError(f"sprite {idx} is empty")
    w_bytes, height = struct.unpack_from("<HH", data, 0)
    w_bytes &= 0x7FFF
    width = w_bytes * 8
    if not w_bytes or not height:
        raise ValueError(f"sprite {idx} has invalid size")
    grid = [[0] * width for _ in range(height)]
    pos = 4
    for segment in range(4):
        plane = 3 - segment
        for y in range(height):
            for byte_x in range(w_bytes):
                value = data[pos]
                pos += 1
                for bit in range(8):
                    if value & (0x80 >> bit):
                        grid[y][byte_x * 8 + bit] |= 1 << plane
    return width, height, grid


def encode(width: int, height: int, grid: list[list[int]]) -> bytes:
    w_bytes = width // 8
    out = bytearray(struct.pack("<HH", w_bytes, height))
    for segment in range(4):
        plane = 3 - segment
        for y in range(height):
            for byte_x in range(w_bytes):
                value = 0
                for bit in range(8):
                    if (grid[y][byte_x * 8 + bit] >> plane) & 1:
                        value |= 0x80 >> bit
                out.append(value)
    out.extend(encode_mask(grid))
    return bytes(out)


def encode_mask(grid: list[list[int]]) -> bytes:
    """寫出與原版 consumer 相容的逐列 run-length AND-mask。

    mask bit=1 代表保留背景（透明），bit=0 代表清除背景後寫入色彩 plane。
    128／129 的補圖沒有原始 mask bytes，因此以色彩 plane 的非零像素建立同一
    遮罩；每列依連續 byte 選擇 0x40／0x80／0xC0 run，避免猜測原版壓縮選擇。
    """
    height = len(grid)
    width = len(grid[0])
    row_bytes = width // 8
    if not 0 < row_bytes < 0x40:
        raise ValueError(f"mask row width out of RLE range: {row_bytes}")
    out = bytearray()
    for row in grid:
        mask_row = []
        for byte_x in range(row_bytes):
            value = 0
            for bit in range(8):
                if row[byte_x * 8 + bit] == 0:
                    value |= 0x80 >> bit
            mask_row.append(value)
        start = 0
        while start < len(mask_row):
            end = start + 1
            while end < len(mask_row) and mask_row[end] == mask_row[start] and end - start < 0x3f:
                end += 1
            count, value = end - start, mask_row[start]
            if value == 0xff:
                out.append(0x40 | count)
            elif value == 0x00:
                out.append(0x80 | count)
            elif value < 0x40:
                # 單 byte literal；高兩位不是 run marker。
                if count != 1:
                    out.extend((0xc0 | count, value))
                else:
                    out.append(value)
            else:
                out.extend((0xc0 | count, value))
            start = end
    return bytes(out)


def neighbours(grid, x: int, y: int):
    height = len(grid)
    width = len(grid[0])
    for dy in (-1, 0, 1):
        for dx in (-1, 0, 1):
            if not dx and not dy:
                continue
            nx, ny = x + dx, y + dy
            if 0 <= nx < width and 0 <= ny < height:
                yield grid[ny][nx]


def style_native_clusters(grid: list[list[int]], monster_id: int) -> list[list[int]]:
    """只改色塊，不改 opaque mask；結果保持原回補角色的 silhouette。

    1--127 的原生圖常以重複 palette alias、短抖點與深色外框組成像素群組。
    這裡不創造新部位，只在已有明暗交界附近加入低密度、可重播的 cluster。
    """

    height = len(grid)
    width = len(grid[0])
    before = [row[:] for row in grid]

    if monster_id == 128:
        # MNSBK index 8 與 15 都是白色；原生圖較常使用 15 作高光。
        aliases = {8: 15, 4: 13}
        light = {1, 15}
        dark = {2, 5, 10, 11, 13}
        outline = 11
    elif monster_id == 129:
        aliases = {}
        light = {6, 7}
        dark = {2, 3, 10, 11}
        outline = 11
    else:
        return grid

    # 先建立不可變的 alias 基線；後續所有候選都只讀這份基線，避免同一輪
    # 的掃描順序或重播次數改變結果。
    base = [[aliases.get(value, value) if value else 0 for value in row] for row in before]
    grid = [row[:] for row in base]

    # 依透明遮罩計算距離外框一、二格的位置；只使用幾何遮罩，不依賴
    # 鄰居的顏色，故同一工具重播時不會因前一輪已改色而產生漂移。
    boundary = [[False] * width for _ in range(height)]
    near_edge = [[False] * width for _ in range(height)]
    for y in range(height):
        for x in range(width):
            if not base[y][x]:
                continue
            ns = list(neighbours(base, x, y))
            boundary[y][x] = any(v == 0 for v in ns) or x in (0, width - 1) or y in (0, height - 1)
    for y in range(height):
        for x in range(width):
            if boundary[y][x]:
                near_edge[y][x] = True
                continue
            near_edge[y][x] = any(
                boundary[ny][nx]
                for ny in range(max(0, y - 1), min(height, y + 2))
                for nx in range(max(0, x - 1), min(width, x + 2))
            )

    # 在已有外框內側加入稀疏 2x2 抖點。候選位置只由 mask 與座標決定，
    # 避免把細節／輪廓整片染平，也保證這個轉換是冪等的。
    for y in range(height):
        for x in range(width):
            value = base[y][x]
            if value not in light:
                continue
            if not near_edge[y][x]:
                continue
            if ((x + 2 * y + monster_id) & 3) == 0:
                grid[y][x] = outline

    # 強化既有透明邊界的局部深色節奏，但只處理原本為亮色的邊界像素，
    # 並保留大部分高光，故不會改變角色外形。
    for y in range(height):
        for x in range(width):
            value = base[y][x]
            if value not in light:
                continue
            if not boundary[y][x]:
                continue
            if ((3 * x + y + monster_id) % 7) != 0:
                continue
            grid[y][x] = outline

    return grid


def load_palette(path: str) -> list[tuple[int, int, int]]:
    data = open(path, "rb").read()
    palette = []
    for i in range(16):
        r, g, b = data[i * 3 : i * 3 + 3]
        palette.append(((r << 2) | (r >> 4), (g << 2) | (g >> 4), (b << 2) | (b >> 4)))
    return palette


def png_chunk(kind: bytes, data: bytes) -> bytes:
    return struct.pack(">I", len(data)) + kind + data + struct.pack(">I", binascii.crc32(kind + data) & 0xFFFFFFFF)


def write_png(path: str, rgba: list[list[tuple[int, int, int, int]]]):
    height = len(rgba)
    width = len(rgba[0])
    raw = bytearray()
    for row in rgba:
        raw.append(0)
        for r, g, b, a in row:
            raw.extend((r, g, b, a))
    blob = b"\x89PNG\r\n\x1a\n"
    blob += png_chunk(b"IHDR", struct.pack(">IIBBBBB", width, height, 8, 6, 0, 0, 0))
    blob += png_chunk(b"IDAT", zlib.compress(bytes(raw), 9))
    blob += png_chunk(b"IEND", b"")
    os.makedirs(os.path.dirname(path), exist_ok=True)
    open(path, "wb").write(blob)


def render(width, height, grid, palette, scale=3):
    rows = []
    for y in range(height):
        row = []
        for x in range(width):
            value = grid[y][x]
            colour = (0, 0, 0, 0) if not value else (*palette[value], 255)
            row.extend([colour] * scale)
        rows.extend([row] * scale)
    return rows


FONT = {
    "0": ("111", "101", "101", "101", "111"),
    "1": ("010", "110", "010", "010", "111"),
    "2": ("110", "001", "010", "100", "111"),
    "3": ("110", "001", "010", "001", "110"),
    "4": ("101", "101", "111", "001", "001"),
    "5": ("111", "100", "110", "001", "110"),
    "6": ("011", "100", "111", "101", "111"),
    "7": ("111", "001", "010", "010", "010"),
    "8": ("111", "101", "111", "101", "111"),
    "9": ("111", "101", "111", "001", "110"),
}


def put_label(canvas, x, y, text, colour=(255, 255, 120, 255)):
    for char in text:
        glyph = FONT[char]
        for gy, line in enumerate(glyph):
            for gx, bit in enumerate(line):
                if bit == "1" and 0 <= y + gy < len(canvas) and 0 <= x + gx < len(canvas[0]):
                    canvas[y + gy][x + gx] = colour
        x += 4


def blank_canvas(width, height, colour):
    return [[colour for _ in range(width)] for _ in range(height)]


def alpha_composite(dst, src, ox, oy):
    for y, row in enumerate(src):
        for x, pixel in enumerate(row):
            if pixel[3] and 0 <= oy + y < len(dst) and 0 <= ox + x < len(dst[0]):
                dst[oy + y][ox + x] = pixel


def nearest_resize(src, new_width, new_height):
    old_height = len(src)
    old_width = len(src[0])
    return [
        [src[(y * old_height) // new_height][(x * old_width) // new_width] for x in range(new_width)]
        for y in range(new_height)
    ]


def write_overviews(blob, offsets, palette, styled):
    # 單隻 PNG：和既有圖鑑一樣使用 3x nearest-neighbour。
    for idx in (128, 129):
        width, height, grid = styled[idx]
        write_png(f"docs/monsters/spr_{idx}.png", render(width, height, grid, palette, 3))

    # 完整 0..129 圖鑑，保留 mns_sprite.py 的 13 欄、90x130 cell 契約。
    cols, cell_w, cell_h = 13, 90, 130
    rows = (len(offsets) - 1 + cols - 1) // cols
    sheet = blank_canvas(cols * cell_w, rows * cell_h, (40, 40, 60, 255))
    for idx in range(len(offsets) - 1):
        cx, cy = (idx % cols) * cell_w, (idx // cols) * cell_h
        for x in range(cell_w):
            sheet[cy][cx + x] = sheet[cy + cell_h - 1][cx + x] = (80, 80, 100, 255)
        for y in range(cell_h):
            sheet[cy + y][cx] = sheet[cy + y][cx + cell_w - 1] = (80, 80, 100, 255)
        put_label(sheet, cx + 2, cy + 1, str(idx))
        width, height, grid = styled.get(idx) or decode(blob, offsets, idx)
        src = render(width, height, grid, palette, 1)
        max_w, max_h = cell_w, cell_h - 14
        scale = min(max_w / width, max_h / height, 1.0)
        nw, nh = max(1, int(width * scale)), max(1, int(height * scale))
        if (nw, nh) != (width, height):
            src = nearest_resize(src, nw, nh)
        alpha_composite(sheet, src, cx + (cell_w - nw) // 2, cy + 14 + (max_h - nh) // 2)
    write_png("docs/monsters/monster_sheet.png", sheet)

    # 128／129 兩格對照圖：固定 400x340 面板，避免只看局部 PNG。
    panel_w, panel_h = 400, 340
    contact = blank_canvas(panel_w * 2, panel_h, (40, 40, 60, 255))
    for panel, idx in enumerate((128, 129)):
        ox = panel * panel_w
        for x in range(panel_w):
            contact[0][ox + x] = contact[panel_h - 1][ox + x] = (80, 80, 100, 255)
        for y in range(panel_h):
            contact[y][ox] = contact[y][ox + panel_w - 1] = (80, 80, 100, 255)
        put_label(contact, ox + 8, 5, str(idx))
        width, height, grid = styled[idx]
        src = render(width, height, grid, palette, 3)
        max_w, max_h = panel_w - 20, panel_h - 35
        if len(src[0]) > max_w or len(src) > max_h:
            scale = min(max_w / len(src[0]), max_h / len(src))
            src = nearest_resize(src, max(1, int(len(src[0]) * scale)), max(1, int(len(src) * scale)))
        alpha_composite(contact, src, ox + (panel_w - len(src[0])) // 2, 25 + (max_h - len(src)) // 2)
    write_png("docs/monsters/restored_128_129.png", contact)


def write_c(blobs: dict[int, bytes]):
    lines = [
        "/* dq3_restored_sprites.c — bug #3 復原 boss sprite(128 歐里狄加 / 129 五頭龍大王)。",
        " * 生成檔(tools/style_sprites.py);格式 = MNSBK.PAL plane-major,同 DQ3MNS.SHP",
        " * 單隻 sprite:u16 w_bytes、u16 h、4 plane(plane3→0),尾端為逐列 RLE AND-mask。",
        " * C fallback 仍以色彩 plane 建立 opaque；Go SHP loader 會直接消費此 mask。",
        " * 來源與推論等級見 docs/121-monster-128-129-style-audit.md；本工具只做局部像素風格整理。 */",
        '#include "dq3_restored_sprites.h"',
        "#include <string.h>",
        "",
    ]
    for idx in (128, 129):
        data = blobs[idx]
        lines.append(f"static const unsigned char SPR_{idx}[{len(data)}] = {{")
        for start in range(0, len(data), 20):
            lines.append("    " + ", ".join(str(v) for v in data[start : start + 20]) + ",")
        lines.append("};")
        lines.append("")
    lines.extend(
        [
            "int dq3_restored_sprite(int id, dq3_monster_sprite *out)",
            "{",
            "    const unsigned char *d; int n;",
            "    if (id == 128) { d = SPR_128; n = (int)sizeof SPR_128; }",
            "    else if (id == 129) { d = SPR_129; n = (int)sizeof SPR_129; }",
            "    else return -1;",
            "    {",
            "        int wb = d[0] | (d[1] << 8); int h = d[2] | (d[3] << 8);",
            "        int W = (wb & 0x7fff) * 8, s, r, b, bit; int plane_sz = (wb & 0x7fff) * h;",
            "        wb &= 0x7fff;",
            "        if (W <= 0 || h <= 0 || W > DQ3_SHP_MAXW || h > DQ3_SHP_MAXH) return -1;",
            "        if (4 + plane_sz * 4 > n) return -1;",
            "        memset(out, 0, sizeof *out); out->w = W; out->h = h;",
            "        for (s = 0; s < 4; s++) {",
            "            int pl = 3 - s; int b0 = 4 + s * plane_sz;",
            "            for (r = 0; r < h; r++)",
            "                for (b = 0; b < wb; b++) {",
            "                    unsigned char v = d[b0 + r * wb + b];",
            "                    for (bit = 0; bit < 8; bit++)",
            "                        if (v & (0x80 >> bit)) out->px[r][b*8+bit] |= (unsigned char)(1 << pl);",
            "                }",
            "        }",
            "        for (r = 0; r < h; r++)",
            "            for (b = 0; b < W; b++)",
            "                out->opaque[r][b] = out->px[r][b] ? 1 : 0;",
            "    }",
            "    return 0;",
            "}",
            "",
        ]
    )
    open(C_OUT, "w", encoding="utf-8").write("\n".join(lines))


def write_go(blobs: dict[int, bytes]):
    lines = [
        "// restored_sprites.go — 由 tools/style_sprites.py 產生。",
        "// 復原 boss sprite (128 歐里狄加 / 129 五頭龍大王)。",
        "// 原版 DQ3MNS.SHP 這兩格為空；本檔提供符合 MNSBK.PAL 且含 RLE AND-mask 的回補資料。",
        "// 本輪只做局部像素群組／palette alias 整理，原始來源與推論等級見 docs/121。",
        "package dq3data",
        "",
        "// restoredSprite 依怪物 ID 提供復原 sprite byte 資料。",
        "// 格式：u16 w_bytes、u16 h、4 plane (plane3→0)、逐列 RLE AND-mask。",
        "var restoredSprite = map[int][]byte{",
    ]
    for idx in (128, 129):
        lines.append(f"\t{idx}: {{")
        data = blobs[idx]
        for start in range(0, len(data), 16):
            lines.append("\t\t" + ", ".join(f"0x{v:02x}" for v in data[start : start + 16]) + ",")
        lines.append("\t},")
    lines.extend(["}", ""])
    open(GO_OUT, "w", encoding="utf-8").write("\n".join(lines))


def main():
    blob = open(SHP_IN, "rb").read()
    offsets = read_offsets(blob)
    palette = load_palette(PAL_PATH)
    styled = {}
    for idx in (128, 129):
        width, height, grid = decode(blob, offsets, idx)
        styled[idx] = (width, height, style_native_clusters(grid, idx))

    out = bytearray(blob[: offsets[128]])
    blobs = {}
    new_offsets = list(offsets)
    for idx in (128, 129):
        width, height, grid = styled[idx]
        blobs[idx] = encode(width, height, grid)
        new_offsets[idx] = len(out)
        out.extend(blobs[idx])
        new_offsets[idx + 1] = len(out)
    # Keep the original offset table at the front, changing only 128/129/sentinel.
    for idx in (128, 129, 130):
        struct.pack_into("<I", out, idx * 4, new_offsets[idx])
    open(SHP_OUT, "wb").write(out)
    # mobile/assets 是被 embed 的執行版素材目錄；它是 ignored 的使用者素材副本，
    # 不會覆寫 assets_raw，也不會把版權資料加入 Git。
    if os.path.isdir(os.path.dirname(MOBILE_SHP_OUT)):
        open(MOBILE_SHP_OUT, "wb").write(out)
    write_c(blobs)
    write_go(blobs)
    write_overviews(out, new_offsets, palette, styled)
    for idx in (128, 129):
        width, height, grid = styled[idx]
        counts = Counter(v for row in grid for v in row if v)
        print(f"styled {idx}: {width}x{height} opaque={sum(counts.values())} colors={dict(sorted(counts.items()))}")
    print(f"wrote {SHP_OUT} ({len(out)} bytes), {MOBILE_SHP_OUT}, {C_OUT}, {GO_OUT}, and monster overviews")


if __name__ == "__main__":
    main()
