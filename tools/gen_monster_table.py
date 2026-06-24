import struct,json
exe=open('assets_raw/DQ3.EXE','rb').read(); DG=0x16140
dat=open('assets_raw/D3MNS.DAT','rb').read(); REC=0x29
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
# bit→spell_id remap(0x3930)
def bit2rec(bit):
    if bit<0x10: sid=bit
    elif bit<0x12: sid=bit+2
    else: sid=exe[DG+0x3930+(bit-0x12)]
    return sid+0x79
def u16(b,o): return b[o]|(b[o+1]<<8)
def spells(rec_bytes):
    out=[]
    for byi in range(6):
        bv=rec_bytes[0x0e+byi]
        for bit in range(8):
            if bv & (0x80>>bit):
                out.append(nm(bit2rec(byi*8+bit)))
    return out
rows=[]
for i in range(130):
    o=i*REC; b=dat[o:o+REC]
    name=nm(0x258+i)
    hp=u16(b,0)+u16(b,2); atk=b[0x08]; dfn=u16(b,0x09); agi=u16(b,0x0b)
    exp=u16(b,0x21); gold=u16(b,0x23)
    cast=b[0x0d]; fth=b[0x17]; frt=b[0x18]
    sp=spells(b)
    rows.append((i,name,hp,atk,dfn,agi,exp,gold,cast,fth,frt,sp))
# 輸出 markdown
import io
L=['# 精訊版 DQ3 完整怪物屬性表(130 隻)','',
   '各怪數值由 `D3MNS.DAT`(130 筆 × 0x29 byte,docs/16)直接抽出;法術由 AI 咒文 bitmask',
   '(+0x0e..+0x13)經 EXE remap(docs/37)解出真實咒名。HP = `+0x00 base + +0x02 rand` 上限;',
   '攻=`+0x08`、防=`+0x09`、速=`+0x0b`;經驗=`+0x21`、金錢=`+0x23`。AI 欄(docs/37):施咒%=`+0x0d`/256、',
   '逃閾=`+0x17`(我方平均等級門檻)、逃率=`+0x18`/256。怪物施法**不耗 MP**(DQ3 怪物自由施咒,',
   '無 MP 欄;施咒與否由施咒% 決定)。',
   '',
   '| id | 怪名 | HP | 攻 | 防 | 速 | 經驗 | 金錢 | 施咒% | 逃閾 | 逃率 | 會用的法術 |',
   '|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|']
for (i,name,hp,atk,dfn,agi,exp,gold,cast,fth,frt,sp) in rows:
    sps='、'.join(sp) if sp else '—'
    L.append('| %d | %s | %d | %d | %d | %d | %d | %d | %d | %d | %d | %s |'%(
        i,name,hp,atk,dfn,agi,exp,gold,cast,fth,frt,sps))
open('docs/38-monster-stats.md','w').write('\n'.join(L)+'\n')
print('wrote docs/38-monster-stats.md  (%d monsters)'%len(rows))
# 摘要幾隻
for i in [1,5,100,128,129]:
    r=rows[i]; print(i,r[1],'HP',r[2],'atk',r[3],'spells',r[11])
