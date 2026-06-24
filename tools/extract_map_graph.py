#!/usr/bin/env python3
"""從所有 CTY*.DAT 的 +0xc 轉場表抽出完整地圖連通圖 → docs/maps/map_graph.md + .json。

每 section header +0xc = 轉場表(4-byte 項 {destCTY,destSec,X,Y},by 轉場 tile subid);
轉場 tile = 屬性 attr&0xe000(BLKBMn.DAT)。把每個 CTY 每個 section 的「門 tile 座標 → 目的」
全抽出來,組成 CTYa.secX → CTYb.secY 的有向圖。資料驅動,非人工。

docker:
  docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
    python tools/extract_map_graph.py
"""
import json
import os
import struct

EXE = open('assets_raw/DQ3.EXE', 'rb').read()
MAP_BLKNUM = [EXE[0x16140 + 0x0a04 + i] for i in range(94)]
att_cache = {}


def attr_tab(n):
    if n not in att_cache:
        att_cache[n] = open('work/BLKBM%d.DAT' % n, 'rb').read()
    return att_cache[n]


def sections(cty):
    u = lambda o: struct.unpack_from('<H', cty, o)[0]
    offs, first = {}, None
    for s in range(24):
        if 2 * s + 1 >= len(cty):
            break
        o = u(2 * s)
        if first is not None and 2 * s >= first:
            break
        if o == 0xffff:
            continue
        if o == 0 and s > 0:
            break
        offs[s] = o
        first = o if first is None else min(first, o)
    return offs


def cty_links(idx):
    name = 'work/CTY%02d.DAT' % idx
    if not os.path.exists(name):
        return None
    cty = open(name, 'rb').read()
    if len(cty) < 4:
        return None
    u = lambda o: struct.unpack_from('<H', cty, o)[0]
    att = attr_tab(MAP_BLKNUM[idx] if idx < len(MAP_BLKNUM) else 1)
    attr = lambda ti: struct.unpack_from('<H', att, ti * 2)[0] if ti * 2 + 1 < len(att) else 0
    out = {}
    for sec, so in sections(cty).items():
        if so + 0x10 > len(cty):
            continue
        lp = u(so + 0x0e)
        lay = so + lp
        if lay + 4 > len(cty):
            continue
        w, h = u(lay), u(lay + 2)
        base = lay + 4
        tp = u(so + 0x0c)
        if tp == 0xffff:
            out[sec] = []
            continue
        tt = so + tp
        trans = [cty[tt + k * 4:tt + k * 4 + 4] for k in range(16)]
        doors = []
        if 0 < w <= 64 and 0 < h <= 64 and base + 2 * w * h <= len(cty):
            for y in range(h):
                for x in range(w):
                    o = base + 2 * (y * w + x)
                    ti, hi = cty[o], cty[o + 1]
                    if attr(ti) & 0xe000:
                        sub = hi & 0x1f
                        if sub < len(trans) and len(trans[sub]) == 4:
                            e = trans[sub]
                            doors.append((x, y, e[0], e[1], e[2], e[3]))
        # 去重(同目的只記一個代表門座標)
        seen, uniq = set(), []
        for x, y, dc, ds, dx, dy in doors:
            key = (dc, ds)
            if key not in seen:
                seen.add(key)
                uniq.append({'door': [x, y], 'dest_cty': dc, 'dest_sec': ds, 'spawn': [dx, dy]})
        out[sec] = uniq
    return out


def main():
    graph = {}
    for idx in range(94):
        links = cty_links(idx)
        if links is not None:
            graph[idx] = links
    os.makedirs('docs/maps', exist_ok=True)
    json.dump(graph, open('docs/maps/map_graph.json', 'w'), ensure_ascii=False, indent=0)

    lines = ['# 地圖連通圖(全 CTY 轉場 metadata,自動抽出)', '',
             '由各 `CTYxx.DAT` 的 section header `+0xc` 轉場表自動抽出(資料驅動,非人工)。',
             '每列 = 一個 section 的所有門 → 目的地。`門(x,y)` 為 sec0-relative tile 座標,',
             '`→ CTYa.secB @(x,y)` 為踩門後載入的目的 CTY/section + 落點座標。',
             '格式資料另見 `map_graph.json`。連結邏輯見 docs/35 §五(`+0xc` 轉場系統)。', '']
    cross = 0
    for idx in sorted(graph):
        lines.append('## CTY%d' % idx)
        for sec in sorted(graph[idx]):
            ds = graph[idx][sec]
            if not ds:
                lines.append('- sec%d:（無轉場）' % sec)
                continue
            parts = []
            for e in ds:
                tag = 'CTY%d.sec%d' % (e['dest_cty'], e['dest_sec'])
                if e['dest_cty'] != idx:
                    cross += 1
                parts.append('門(%d,%d)→%s@(%d,%d)' % (e['door'][0], e['door'][1], tag, e['spawn'][0], e['spawn'][1]))
            lines.append('- sec%d: %s' % (sec, '  '.join(parts)))
        lines.append('')
    open('docs/maps/map_graph.md', 'w').write('\n'.join(lines))
    print('graph: %d CTY, %d 跨 CTY 連結 → docs/maps/map_graph.{md,json}' % (len(graph), cross))


if __name__ == '__main__':
    main()
