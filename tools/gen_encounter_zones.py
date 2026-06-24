#!/usr/bin/env python3
"""遭遇區系統 → docs/39-encounter-zones.md。RE:encounter_build_group(file 0xbb45,docs/13)。
overworld region map @DGROUP 0x4966(16×16 grid,cell=(X>>4)+(Y&0xf0)→region);
遭遇表 @0x4a56(region×0x20 = 4 子表×8 byte;子表 byte2=戰鬥背景頁、byte4-7=候選怪 0xff 終止)。"""
import struct,json
exe=open('assets_raw/DQ3.EXE','rb').read(); DG=0x16140
gm=json.load(open('docs/data/glyph_unicode_map.json'))
txt=open('assets_raw/D3TXT00.TXT','rb').read(); ptab=struct.unpack_from('<760H',txt,0)
def nm(rec):
    if rec>=len(ptab):return '?'
    o=ptab[rec];s=''
    while o+1<len(txt):
        v=struct.unpack_from('<H',txt,o)[0];o+=2
        if v==0xffff or v>=0xffed:break
        s+=gm.get(str(v),'·')
    return s
def mon(mid): return '' if mid==0xff else nm(0x258+mid)
RMAP=0x4966; ETAB=0x4a56
# region map 16x16
grid=[[exe[DG+RMAP+y*16+x] for x in range(16)] for y in range(16)]
used=sorted(set(v for row in grid for v in row) | set([0xf,0x10,0x11,0x12]))
def region_cands(reg):
    o=DG+ETAB+reg*0x20; subs=[]
    for s in range(4):
        e=exe[o+s*8:o+s*8+8]
        cands=[mon(e[4+k]) for k in range(4) if e[4+k]!=0xff]
        subs.append((e[2],cands))   # (bg page, candidates)
    return subs
L=['# 精訊版 DQ3 遭遇區系統(哪個區域出哪些怪)','',
   '反組譯 `encounter_build_group`(file 0xbb45,docs/13)。玩家每走一步在非安全場景擲遭遇;',
   '觸發時依**位置 → region → 遭遇表**生成敵群。',
   '',
   '## region 怎麼決定(file 0xbb45)','',
   '- **地表(overworld)**:grid cell = `(X>>4) + (Y & 0xf0)`,`region = [cell + 0x4966]`',
   '  (16×16 region map,見下)。Y≥0x12c(南界)用固定 region 0x11/0x12。',
   '- **城鎮/洞窟(CTY)**:`region = [0xd77]`(該 section 自帶;0 = 安全,無遭遇)。',
   '',
   '## 遭遇表結構(DGROUP 0x4a56)','',
   '每 region = 0x20 byte = **4 子表 × 8 byte**;觸發時 rng(4) 挑一子表。子表:',
   '`byte2` = 戰鬥背景頁、`byte4..7` = 候選怪 id(0xff 終止,最多 4)。再依 spawn_weight(+0x28)',
   '預算決定同種隻數(docs/13)。',
   '',
   '## 地表 region map(16×16,cell→region)','',
   '行 = Y>>4(0..15)、欄 = X>>4(0..15);值 = region 號(0=安全)。','',
   '```']
L.append('     '+' '.join('%2X'%x for x in range(16)))
for y in range(16):
    L.append('Y%X:  '%y + ' '.join('%2X'%grid[y][x] for x in range(16)))
L.append('```')
L.append('')
L.append('## 各 region 候選怪')
L.append('')
for reg in used:
    subs=region_cands(reg)
    if all(not c for _,c in subs): continue
    L.append('### region 0x%X'%reg)
    seen=set()
    for s,(bg,cands) in enumerate(subs):
        if not cands: continue
        key=tuple(cands)
        tag='(背景頁%d)'%bg
        L.append('- 子表%d %s:%s'%(s,tag,'、'.join(cands)))
    L.append('')
open('docs/39-encounter-zones.md','w').write('\n'.join(L)+'\n')
print('wrote docs/39-encounter-zones.md  regions used:',['0x%X'%r for r in used])
