#!/usr/bin/env python3
"""生成 dq3_remake/src/dq3_encounter.c:overworld region map(0x4966)+ 遭遇表(0x4a56)。
docs/39。位置→region→候選怪。"""
exe=open('assets_raw/DQ3.EXE','rb').read(); DG=0x16140
RMAP=0x4966; ETAB=0x4a56
rmap=[exe[DG+RMAP+i] for i in range(256)]
NREG=32
etab=[exe[DG+ETAB+i] for i in range(NREG*0x20)]
def carr(name,arr,perline=16):
    out=['static const unsigned char %s[%d] = {'%(name,len(arr))]
    for i in range(0,len(arr),perline):
        out.append('    '+', '.join(str(b) for b in arr[i:i+perline])+',')
    out.append('};')
    return out
L=['/* dq3_encounter.c — 遭遇區資料(生成檔,tools/gen_encounter_data.py;docs/39)。',
   ' * overworld region map(0x4966,16×16)+ 遭遇表(0x4a56,region×0x20=4 子表×8 byte)。 */',
   '#include "dq3_encounter.h"','']
L+=carr('REGION_MAP',rmap)
L.append('')
L+=carr('ENC_TABLE',etab)
L+=['',
'/* 地表座標 → region(file 0xbb45):Y>=0x12c 用固定 region;否則 region map。 */',
'int dq3_encounter_region(int x, int y)',
'{',
'    if (y >= 0x12c) return (x >= 0x7a) ? 0x12 : 0x11;',
'    { int cell = ((x >> 4) & 0xf) + (y & 0xf0); return REGION_MAP[cell & 0xff]; }',
'}',
'',
'/* region + rng → 候選怪 id(挑 1 子表→非 0xff 候選隨機一個)。無候選回 -1。 */',
'int dq3_encounter_pick(int region, unsigned rng)',
'{',
'    const unsigned char *e; int cand[4], nc = 0, k;',
'    if (region < 0 || region >= 32) return -1;',
'    e = &ENC_TABLE[region * 0x20 + (rng & 3) * 8];',
'    for (k = 0; k < 4; k++) if (e[4 + k] != 0xff) cand[nc++] = e[4 + k];',
'    if (nc == 0) return -1;',
'    return cand[(rng >> 2) % (unsigned)nc];',
'}',
'',
'/* region 子表的戰鬥背景頁(byte2)。 */',
'int dq3_encounter_bgpage(int region, unsigned rng)',
'{ if (region<0||region>=32) return 0; return ENC_TABLE[region*0x20 + (rng&3)*8 + 2]; }',
'']
open('dq3_remake/src/dq3_encounter.c','w').write('\n'.join(L))
print('wrote dq3_remake/src/dq3_encounter.c')
