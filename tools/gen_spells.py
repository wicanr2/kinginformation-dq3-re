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

# ---- 咒文施放 descriptor:DQ3.EXE 0x37c4 表(3 byte/咒,docs/13)----
# b0=base 威力、b1=目標類別(bits0x18:0x08單/0x18組/0x10全)、b2 未用。
# 施放公式(0xc22e)= base/2 + rng(base/2)。MP 用 DQ3 標準值(EXE MP 來源未定位)。
DESC = 0x37c4
HEAL_IDS = {40, 41, 42, 43, 44}          # 荷依米系
# DQ3 標準 MP(BBS 佐證);spell_id→MP
MP = {0:2,1:6,2:10, 3:4,4:6,5:10, 6:5,7:8,8:15, 9:3,10:5,11:8,12:11,
      13:4,14:8,15:12, 16:4,17:15, 40:3,41:5,42:7,43:18,44:36}
def tgt(b1):
    c = b1 & 0x18
    return {0x08:'DQ3_TG_ENEMY1', 0x18:'DQ3_TG_GROUP', 0x10:'DQ3_TG_ALL'}.get(c, 'DQ3_TG_ENEMY1')

defs = []
for sid in range(48):
    o = DG + DESC + sid * 3
    b0, b1 = d[o], d[o + 1]
    if b0 == 0:                # 輔助/狀態咒(無傷害威力)— 暫不列為可施放
        continue
    rec = sid + 0x79
    if sid in HEAL_IDS:
        kind, target = 'DQ3_SK_HEAL', ('DQ3_TG_ALLYALL' if sid in (43, 44) else 'DQ3_TG_ALLY1')
    else:
        kind, target = 'DQ3_SK_DMG', tgt(b1)
    mp = MP.get(sid, 6)
    defs.append((rec, mp, b0, kind, target, recname(rec)))

dl = ['/* dq3_spelldef.c — 咒文施放 descriptor(生成檔,tools/gen_spells.py)。',
      ' * base 威力 / 目標 由 DQ3.EXE 0x37c4 表抽出(3 byte/咒:b0=base、b1=目標類別 bits0x18);',
      ' * 公式 base/2+rng(base/2)(docs/13 file 0xc22e)、魔甲減半 #7b。MP 為 DQ3 標準值',
      ' * (EXE MP 來源未定位)。spell_id = rec - 0x79。請勿手改,改 tools/gen_spells.py。 */',
      '#include "dq3_spell.h"', '#include <stddef.h>', '',
      'static const dq3_spell_def DEFS[] = {']
for rec, mp, base, kind, target, nm in defs:
    dl.append('    { %3d, %2d, %3d, %s, %s },  /* %s */' % (rec, mp, base, kind, target, nm))
dl += ['};', '',
       'const dq3_spell_def *dq3_spell_def_get(unsigned short rec)',
       '{', '    int i, n = (int)(sizeof DEFS / sizeof DEFS[0]);',
       '    for (i = 0; i < n; i++) if (DEFS[i].rec == rec) return &DEFS[i];',
       '    return NULL;', '}', '']

# 怪物咒文 bit(0..47)→ spell_id(DQ3.EXE remap,docs/37):bit<0x10→bit、[0x10,0x12)→+2、
# ≥0x12→0x3930[bit-0x12]。再 +0x79 = 咒名 rec。
REMAP = 0x3930
bit2rec = []
for bit in range(48):
    if bit < 0x10: sid = bit
    elif bit < 0x12: sid = bit + 2
    else: sid = d[DG + REMAP + (bit - 0x12)]
    bit2rec.append(sid + 0x79)
dl.append('/* 怪物咒文 bit(0..47)→ 咒名 rec(EXE 0x3930 remap,docs/37)。 */')
dl.append('const unsigned short dq3_monster_spell_rec[48] = {')
for r in range(0, 48, 12):
    dl.append('    ' + ', '.join('%d' % bit2rec[r + j] for j in range(12) if r + j < 48) + ',')
dl.append('};')
dl.append('')
open('dq3_remake/src/dq3_spelldef.c', 'w').write('\n'.join(dl))
print('wrote dq3_remake/src/dq3_spelldef.c  (%d 可施放咒,EXE base + 怪物 bit remap)' % len(defs))
