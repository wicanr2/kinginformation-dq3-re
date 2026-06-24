#!/usr/bin/env python3
"""產生注音→字模候選表(dq3_remake/src/dq3_zhuyin_table.c)。

原版 DQ3 用 EXE 內建注音引擎(seg 0x11c4)做注音→國字;remake 改用此表:
把 D3TXT00.FON 的 1476 字模經 glyph_unicode_map 取 unicode,再用 pypinyin 取 BOPOMOFO,
反建「注音音節 → 候選字模 glyph」表(721 音節桶 / 1338 漢字)。

在 docker uv venv 內跑(不污染 host):
  docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
    bash -c 'uv venv -q /tmp/v && . /tmp/v/bin/activate && uv pip install -q pypinyin && python tools/gen_zhuyin.py'
"""
import json
from pypinyin import pinyin, Style

SH = "ㄅㄆㄇㄈㄉㄊㄋㄌㄍㄎㄏㄐㄑㄒㄓㄔㄕㄖㄗㄘㄙ"   # 聲母 21(code 1-21)
JIE = "ㄧㄨㄩ"                                        # 介音 3(1-3)
YUN = "ㄚㄛㄜㄝㄞㄟㄠㄡㄢㄣㄤㄥㄦ"                    # 韻母 13(1-13)


def decompose(bpmf):
    sh = ji = yu = tone = 0
    for ch in bpmf:
        if ch in SH: sh = SH.index(ch) + 1
        elif ch in JIE: ji = JIE.index(ch) + 1
        elif ch in YUN: yu = YUN.index(ch) + 1
        elif ch == "ˊ": tone = 1
        elif ch == "ˇ": tone = 2
        elif ch == "ˋ": tone = 3
        elif ch == "˙": tone = 4
    return (sh, ji, yu, tone)


def pack(sh, ji, yu, tone):
    return ((sh * 4 + ji) * 14 + yu) * 5 + tone


def main():
    gm = json.load(open("docs/data/glyph_unicode_map.json"))
    buckets = {}
    for g in range(1476):
        u = gm.get(str(g))
        if not u or len(u) != 1 or not ('一' <= u <= '鿿'):
            continue
        b = pinyin(u, style=Style.BOPOMOFO, errors='ignore')
        if not b or not b[0]:
            continue
        k = decompose(b[0][0])
        if k == (0, 0, 0, 0):
            continue
        buckets.setdefault(pack(*k), []).append(g)

    items = sorted(buckets.items())
    keys, offs, pool = [], [0], []
    for pk, gl in items:
        keys.append(pk); pool.extend(gl); offs.append(len(pool))

    out = [
        '/* dq3_zhuyin_table.c — 注音→字模候選表(生成檔;tools/gen_zhuyin.py via pypinyin)。',
        ' * key = ((聲母*4+介音)*14+韻母)*5+聲調;聲母1-21/介音1-3/韻母1-13/聲調0-4(0=無)。',
        ' * 由 D3TXT00.FON 1476 字模的 unicode(glyph_unicode_map)經 pypinyin BOPOMOFO 反建。 */',
        '#include "dq3_zhuyin.h"', '',
        'const int dq3_zh_nbucket = %d;' % len(keys),
        'const uint16_t dq3_zh_key[%d] = {%s};' % (len(keys), ",".join(map(str, keys))),
        'const uint16_t dq3_zh_off[%d] = {%s};' % (len(offs), ",".join(map(str, offs))),
        'const uint16_t dq3_zh_pool[%d] = {%s};' % (len(pool), ",".join(map(str, pool))),
    ]
    open("dq3_remake/src/dq3_zhuyin_table.c", "w").write("\n".join(out) + "\n")
    print("buckets=%d pool=%d" % (len(keys), len(pool)))


if __name__ == "__main__":
    main()
