#!/usr/bin/env python3
"""從 DQ3.EXE 抽咒文習得表 → dq3_remake/src/dq3_spells_table.c。

RE 依據(docs/18):升級學咒 sub_db5f(file 0xeecf)。bx=咒文系基底(0 勇者系/8 僧侶系/
0x10 魔法系),每系在指標表 0x36f9 起 stride 8 = {spellA,levelA,spellB,levelB}(DGROUP off)。
平行清單 level[i]/spell[i]:角色到 level[i] 級時學會 spell[i];咒文名 rec = spell_id + 0x79。
清單長度由下一個指標界定(原版越界 bug 即掃過此界 = #5)。

職業→系:勇者0→勇者系、僧侶3→僧侶系、魔法使4→魔法系、賢者5→僧侶+魔法;其餘無咒。

docker: tools/dockrun.sh tools/gen_spells.py
"""
import struct, json

d = open('assets_raw/DQ3.EXE', 'rb').read(); DG = 0x16140
def w(off): return struct.unpack_from('<H', d, DG + off)[0]
def b(off): return d[DG + off]

SCHOOLS = {0: 'HERO', 8: 'PRIEST', 0x10: 'MAGE'}
ptrs = {}
for base, nm in SCHOOLS.items():
    ptrs[(nm, 'spA')] = w(0x36f9 + base); ptrs[(nm, 'lvA')] = w(0x36fb + base)
    ptrs[(nm, 'spB')] = w(0x36fd + base); ptrs[(nm, 'lvB')] = w(0x36ff + base)
alloff = sorted(set(ptrs.values()))
def length(off):
    nxt = [o for o in alloff if o > off]
    return (min(nxt) - off) if nxt else 8

# 咒文名(D3TXT00 rec → unicode,僅供註解可讀)
gm = json.load(open('docs/data/glyph_unicode_map.json'))
txt = open('assets_raw/D3TXT00.TXT', 'rb').read()
ptab = struct.unpack_from('<760H', txt, 0)
def recname(rec):
    if rec >= len(ptab): return '?'
    o = ptab[rec]; s = ''
    while o + 1 < len(txt):
        v = struct.unpack_from('<H', txt, o)[0]; o += 2
        if v == 0xffff or v >= 0xffed: break
        s += gm.get(str(v), '·')
    return s

def school_list(nm):
    seen = {}; out = []
    for tag in ('A', 'B'):
        sp = ptrs[(nm, 'sp' + tag)]; lv = ptrs[(nm, 'lv' + tag)]
        n = min(length(sp), length(lv))
        for i in range(n):
            L = b(lv + i); rec = b(sp + i) + 0x79
            if not (1 <= L <= 43): continue
            if rec in seen: continue        # 同咒只記最早習得級
            seen[rec] = L; out.append((L, rec))
    out.sort()
    return out

lines = [
    '/* dq3_spells_table.c — 咒文習得表(生成檔,tools/gen_spells.py)。',
    ' * 由 DQ3.EXE 升級學咒 sub_db5f 的習得表(DGROUP 0x36f9 起)抽出;每筆 {習得等級, 咒文名 rec}。',
    ' * 職業→系見 dq3_spell.c。請勿手改。 */',
    '#include "dq3_spell.h"', '',
]
counts = {}
for base, nm in SCHOOLS.items():
    lst = school_list(nm); counts[nm] = len(lst)
    lines.append('/* %s:%d 咒 */' % (nm, len(lst)))
    lines.append('const dq3_spell_learn dq3_school_%s[%d] = {' % (nm.lower(), len(lst)))
    for L, rec in lst:
        lines.append('    { %2d, %3d },  /* %s */' % (L, rec, recname(rec)))
    lines.append('};')
    lines.append('const int dq3_school_%s_n = %d;' % (nm.lower(), len(lst)))
    lines.append('')
open('dq3_remake/src/dq3_spells_table.c', 'w').write('\n'.join(lines))
print('wrote dq3_remake/src/dq3_spells_table.c  HERO=%d PRIEST=%d MAGE=%d'
      % (counts['HERO'], counts['PRIEST'], counts['MAGE']))
