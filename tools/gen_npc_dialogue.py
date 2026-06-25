#!/usr/bin/env python3
"""NPC 對話逐句 mapping(獨立資料)。
RE(docs/42):NPC 記錄 7-byte {X,Y,sprite,b3,b4,flags,b6}(section header +0/+2 清單)。
互動子型 = (b3>>3)&7:0/1=對話、2=特殊、3-7=設施(dispatcher 0x839f)。
對話 NPC:對話記錄 = b4(該城 D3TXT bank 本地 rec;printer 全域加 0xbb8)。
flags 低 3 bit = 可見性旗標(區分劇情前/後 NPC 變體)。
多城共用同一 D3TXTnn(各佔不同 b4 區段);精確 CTY→bank 表在 overlay,靜態封死,
此處 file→地名 區域索引 + 早期區(D3TXT01)已驗證。
產出:docs/data/npc_dialogue.json(全城 NPC 記錄)+ stdout(markdown:逐句已驗證區)。"""
import struct,json,glob,os
def u16(b,o): return b[o]|(b[o+1]<<8)
ROOT=os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
gm=json.load(open(os.path.join(ROOT,'docs/data/glyph_unicode_map.json')))
VARNAME={0xfffb:'{名}',0xfffa:'{道具}',0xfff9:'{數}',0xfff6:'{名}',0xfff5:'{名}',0xffed:'{字}'}
HASP={0xfffb,0xfffa,0xfff9,0xfff6,0xfff5,0xffed}
def load(fn):
    b=open(fn,'rb').read(); n=u16(b,0)//2
    return b,struct.unpack_from('<%dH'%n,b,0)
def render(bp,r):
    b,pt=bp
    if r>=len(pt): return None
    o=pt[r];s=''
    while o+1<len(b):
        v=u16(b,o);o+=2
        if v==0xffff:break
        if v==0xfffe or v==0xfffd: s+=' / ';continue
        if v==0xfffc: s+=' // ';continue
        if v>=0xffed:
            s+=VARNAME.get(v,'{?}')
            if v in HASP: o+=2
            continue
        s+=gm.get(str(v),'·')
    return s
def npcs(path):
    b=open(path,'rb').read(); w0=u16(b,0); out=[]
    if w0==0 or w0>=len(b) or w0%2: return out
    for s in range(w0//2):
        sb=u16(b,s*2)
        if sb==0xffff or sb+0x10>len(b): continue
        rel=u16(b,sb+0)
        if rel==0 or rel==0xffff or sb+rel>=len(b): continue
        p=sb+rel; cnt=b[p]
        if cnt>60: continue
        for i in range(cnt):
            o=p+1+i*7
            if o+7>len(b): break
            X,Y,spr,b3,b4,fl,b6=b[o:o+7]
            sub=(b3>>3)&7
            kind='facility' if sub>=3 else ('talk' if sub<2 else 'special')
            out.append({'sec':s,'x':X,'y':Y,'sprite':spr,'sub':sub,'kind':kind,
                        'dlg':b4,'vis_flag':fl&7,'b3':b3,'b6':b6})
    return out
data={}
for p in sorted(glob.glob(os.path.join(ROOT,'assets_raw/CTY*.DAT'))):
    cty=int(os.path.basename(p)[3:5])
    ns=npcs(p)
    if ns: data['CTY%02d'%cty]=ns
json.dump(data, open(os.path.join(ROOT,'docs/data/npc_dialogue.json'),'w'), ensure_ascii=False, indent=0)

# stdout:已驗證早期區逐句(D3TXT01 = CTY00 阿里阿罕 / CTY01 雷貝)
d01=load(os.path.join(ROOT,'assets_raw/D3TXT01.TXT'))
tot=sum(len(v) for v in data.values())
talk=sum(1 for v in data.values() for x in v if x['kind']=='talk')
print('# NPC 對話逐句 mapping(獨立資料)\n')
print('全 %d 城共 %d NPC(對話型 %d)。完整記錄見 `docs/data/npc_dialogue.json`。\n'%(len(data),tot,talk))
for cty,nm in [(0,'阿里阿罕'),(1,'雷貝')]:
    print('## CTY%02d %s(對話檔 D3TXT01,已驗證)\n'%(cty,nm))
    print('| sec | 位置 | sprite | dlg(b4)| 可見旗標 | 對話 |')
    print('|---|---|---|---|---|---|')
    for x in data['CTY%02d'%cty]:
        if x['kind']!='talk': continue
        line=render(d01,x['dlg']) or '(超範圍)'
        print('| %d | (%d,%d) | %d | 0x%02x | 0x%02x | %s |'%(x['sec'],x['x'],x['y'],x['sprite'],x['dlg'],x['vis_flag'],line[:60]))
    print()
