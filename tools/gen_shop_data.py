#!/usr/bin/env python3
"""從 CTY*.DAT 直接萃取「設施表」(武器/防具店、道具店、旅社、教會、記錄點)。
RE 來源:設施 dispatcher file 0x839f → NPC byte4 索引 section header +6 的設施指標表
→ 設施 block[0]=type(跳表 0x623)。武防店/道具店 block = [type][count][item_id...]。
產出:dq3_remake/src/dq3_shopdata.c(+.h)與 docs 表格資料(stdout)。"""
import struct,glob,os,json,sys
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

TYPE={0:'inn',1:'weapon',2:'item',3:'church',4:'record'}
def parse(path):
    b=open(path,'rb').read(); cty=int(os.path.basename(path)[3:5])
    w0=u16(b,0)
    if w0==0 or w0>=len(b) or w0%2: return []
    nsec=w0//2; res=[]
    for s in range(nsec):
        sb=u16(b,s*2)
        if sb==0xffff or sb+8>len(b): continue
        ft_rel=u16(b,sb+6)
        if ft_rel==0 or ft_rel==0xffff or sb+ft_rel+2>len(b): continue
        ft=sb+ft_rel
        # 設施表長度 = (最小 block 偏移 - ft)/2;blocks 緊接表後
        blk0=sb+u16(b,ft)
        if blk0<=ft or blk0>=len(b): continue
        ncount=(blk0-ft)//2
        if ncount<1 or ncount>16: continue
        for k in range(ncount):
            blk=sb+u16(b,ft+k*2)
            if blk>=len(b): continue
            t=b[blk]
            if t>4: continue
            e={'cty':cty,'sec':s,'k':k,'type':t}
            if t in (1,2):
                cnt=b[blk+1]
                if cnt<=20 and blk+2+cnt<=len(b):
                    e['items']=list(b[blk+2:blk+2+cnt])
                else: e['items']=[]
            elif t==0:
                e['inn']=u16(b,blk+1)
            res.append(e)
    return res

allf=[]
for p in sorted(glob.glob(os.path.join(ROOT,'assets_raw/CTY*.DAT'))):
    allf+=parse(p)

# 產 C
lines=['/* dq3_shopdata.c — 自 CTY*.DAT 設施表萃取(tools/gen_shop_data.py 自動產生,勿手改)。',
       ' * RE:dispatcher file 0x839f,NPC byte4 → section header +6 設施表 → block。',
       ' * type: 0=旅社 1=武器/防具店 2=道具店 3=教會 4=記錄點。武防/道具 block=[type][count][item...]。*/',
       '#include "dq3_shopdata.h"','']
# item 清單池
pool=[]; offs={}
def itstr(items):
    key=tuple(items)
    if key in offs: return offs[key]
    o=len(pool); offs[key]=o; pool.extend(items); return o
recs=[]
for e in allf:
    items=e.get('items',[])
    o=itstr(items) if items else 0
    recs.append((e['cty'],e['sec'],e['k'],e['type'],len(items),o,e.get('inn',0)))
lines.append('const unsigned char dq3_shop_itempool[%d] = {'%max(1,len(pool)))
for i in range(0,len(pool),16):
    lines.append('  '+', '.join('0x%02x'%x for x in pool[i:i+16])+',')
lines.append('};')
lines.append('const int dq3_shop_itempool_len = %d;'%len(pool))
lines.append('')
lines.append('const dq3_facility dq3_facilities[%d] = {'%len(recs))
for cty,sec,k,t,n,o,inn in recs:
    lines.append('  { %2d, %d, %d, %d, %2d, %3d, %5d },'%(cty,sec,k,t,n,o,inn))
lines.append('};')
lines.append('const int dq3_facilities_len = %d;'%len(recs))
lines+=['',
 'int dq3_shop_items(int cty, int type, const unsigned char **items)','{',
 '    int i;',
 '    for (i = 0; i < dq3_facilities_len; i++) {',
 '        const dq3_facility *f = &dq3_facilities[i];',
 '        if (f->cty == cty && f->type == type && f->count > 0) {',
 '            if (items) *items = &dq3_shop_itempool[f->item_off];',
 '            return f->count;',
 '        }',
 '    }',
 '    if (items) *items = 0;',
 '    return 0;',
 '}','',
 'const dq3_facility *dq3_facility_at(int cty, int sec, int k)',
 '{',
 '    int i;',
 '    for (i = 0; i < dq3_facilities_len; i++) {',
 '        const dq3_facility *f = &dq3_facilities[i];',
 '        if (f->cty == cty && f->sec == sec && f->k == k) return f;',
 '    }',
 '    return 0;',
 '}']
open(os.path.join(ROOT,'dq3_remake/src/dq3_shopdata.c'),'w').write('\n'.join(lines)+'\n')

hdr='''/* dq3_shopdata.h — CTY 設施表(商店/旅社/教會/記錄點),自 CTY*.DAT 萃取。 */
#ifndef DQ3_SHOPDATA_H
#define DQ3_SHOPDATA_H

enum { DQ3_FAC_INN=0, DQ3_FAC_WEAPON=1, DQ3_FAC_ITEM=2, DQ3_FAC_CHURCH=3, DQ3_FAC_RECORD=4 };

typedef struct {
    unsigned char cty, sec, k;   /* CTY 號、section、設施索引(NPC byte4)*/
    unsigned char type;          /* DQ3_FAC_* */
    unsigned char count;         /* 商店品項數(非商店=0)*/
    unsigned short item_off;     /* 商店品項在 dq3_shop_itempool 的起始 */
    unsigned short inn_cost;     /* 旅社住宿費 raw(type=INN)*/
} dq3_facility;

extern const dq3_facility dq3_facilities[];
extern const int dq3_facilities_len;
extern const unsigned char dq3_shop_itempool[];
extern const int dq3_shop_itempool_len;

/* 取某 CTY 第一個指定類型商店的品項;回品項數,*items 指向 pool。找不到回 0。 */
int dq3_shop_items(int cty, int type, const unsigned char **items);

/* 依 (cty, sec, k=NPC byte4) 精確查設施;找不到回 NULL。供「面向店員 → 開該攤」。 */
const dq3_facility *dq3_facility_at(int cty, int sec, int k);

#endif
'''
open(os.path.join(ROOT,'dq3_remake/include/dq3_shopdata.h'),'w').write(hdr)

# 統計 + markdown 表(stdout 供文件)
nshop=sum(1 for e in allf if e['type'] in (1,2))
print('# 萃取 %d 設施(%d 商店),pool=%d bytes'%(len(allf),nshop,len(pool)))
print('| CTY | 類型 | 品項 |')
print('|---|---|---|')
for e in allf:
    if e['type'] in (1,2):
        nmt={1:'武防店',2:'道具店'}[e['type']]
        its=', '.join('%s'%iname(x) for x in e['items'])
        print('| %d | %s | %s |'%(e['cty'],nmt,its))
