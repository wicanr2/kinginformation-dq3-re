#!/usr/bin/env python3
"""從 CTY*.DAT 的 examine 事件表(section header +8)萃取寶箱 / 隱藏物品 / 調査點。
RE 來源 docs/31/35 §三:ev=section_base+word[base+8];ev[0]=count;之後 4-byte entry
{type, param(u16), flag}。type 0=寶箱/調査(param=道具,0=空)、1/3=寶箱給道具、2=傳送、4=特殊。
產出 dq3_remake/src/dq3_treasure.{c,h} + docs 表(stdout)。"""
import struct,glob,os,json
def u16(b,o): return b[o]|(b[o+1]<<8)
ROOT=os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
gm=json.load(open(os.path.join(ROOT,'docs/data/glyph_unicode_map.json')))
txt=open(os.path.join(ROOT,'assets_raw/D3TXT00.TXT'),'rb').read(); ptab=struct.unpack_from('<760H',txt,0)
def nm(rec):
    if rec>=len(ptab): return '?'
    o=ptab[rec];s=''
    while o+1<len(txt):
        v=struct.unpack_from('<H',txt,o)[0];o+=2
        if v==0xffff or v>=0xffed:break
        s+=gm.get(str(v),'·')
    return s
def iname(it): return nm(it+1)
def parse(path):
    b=open(path,'rb').read(); cty=int(os.path.basename(path)[3:5])
    w0=u16(b,0)
    if w0==0 or w0>=len(b) or w0%2: return []
    res=[]
    for s in range(w0//2):
        sb=u16(b,s*2)
        if sb==0xffff or sb+0x10>len(b): continue
        ev_rel=u16(b,sb+8)
        if ev_rel==0 or ev_rel==0xffff or sb+ev_rel>=len(b): continue
        ev=sb+ev_rel; cnt=b[ev]
        if cnt==0 or cnt>40: continue
        for i in range(cnt):
            eo=ev+1+i*4
            if eo+4>len(b): break
            t=b[eo]; param=u16(b,eo+1); flag=b[eo+3]
            # 只收給道具型(type 0/1/3),param 為有效道具 id 且非 0
            if t in (0,1,3) and 0<param<0x90:
                res.append({'cty':cty,'sec':s,'sub':i,'type':t,'item':param,'flag':flag})
    return res
allt=[]
for p in sorted(glob.glob(os.path.join(ROOT,'assets_raw/CTY*.DAT'))): allt+=parse(p)

# C 產出
L=['/* dq3_treasure.c — 寶箱 / 隱藏物品(自 CTY*.DAT examine 事件表萃取;勿手改)。',
   ' * RE docs/31/35 §三:section header +8 事件表,type 0/1/3 給道具,param=道具 id。',
   ' * flag = 一次性旗標 id([0x4f70] bit 陣列):set=未取、清=已取。*/',
   '#include "dq3_treasure.h"','',
   'const dq3_treasure dq3_treasures[%d] = {'%len(allt)]
for e in allt:
    L.append('  { %2d, %d, %2d, %d, 0x%02x, 0x%02x },'%(e['cty'],e['sec'],e['sub'],e['type'],e['item'],e['flag']))
L+=['};','const int dq3_treasures_len = %d;'%len(allt),'',
    'int dq3_treasures_for(int cty, const dq3_treasure **out)','{',
    '    int i, n=0, first=-1;',
    '    for (i = 0; i < dq3_treasures_len; i++) {',
    '        if (dq3_treasures[i].cty == cty) { if (first<0) first=i; n++; }',
    '    }',
    '    if (out) *out = (first>=0) ? &dq3_treasures[first] : 0;',
    '    return n;','}']
open(os.path.join(ROOT,'dq3_remake/src/dq3_treasure.c'),'w').write('\n'.join(L)+'\n')
open(os.path.join(ROOT,'dq3_remake/include/dq3_treasure.h'),'w').write('''/* dq3_treasure.h — CTY 寶箱 / 隱藏物品表(自 CTY*.DAT examine 事件表萃取)。 */
#ifndef DQ3_TREASURE_H
#define DQ3_TREASURE_H

typedef struct {
    unsigned char cty, sec, sub;   /* CTY 號、section、事件 subid */
    unsigned char type;            /* 0=調査點/寶箱 1/3=寶箱給道具 */
    unsigned char item;            /* 道具 id(= ITEM.DAT index;名 = D3TXT00 rec id+1)*/
    unsigned char flag;            /* 一次性旗標 id([0x4f70] bit;set=未取)*/
} dq3_treasure;

extern const dq3_treasure dq3_treasures[];
extern const int dq3_treasures_len;
/* 某 CTY 的寶箱清單;回數量,*out 指向首筆(連續)。 */
int dq3_treasures_for(int cty, const dq3_treasure **out);

#endif
''')
# 統計 + 文件表
from collections import defaultdict
by=defaultdict(list)
for e in allt: by[e['cty']].append(e['item'])
print('# 萃取 %d 寶箱/隱藏物品,橫跨 %d 個 CTY'%(len(allt),len(by)))
print('| CTY | 寶箱 / 隱藏物品 |')
print('|---|---|')
for cty in sorted(by):
    print('| %d | %s |'%(cty,'、'.join(iname(x) for x in by[cty])))
