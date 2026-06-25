#!/usr/bin/env python3
"""NPC 對話逐句 mapping(獨立資料)。
RE(docs/42):NPC 記錄 7-byte {X,Y,sprite,b3,b4,flags,b6}(section header +0/+2 清單)。
互動子型 = (b3>>3)&7:0/1=對話、2=特殊、3-7=設施(dispatcher 0x839f)。
對話 NPC:對話記錄 = b4(printer 全域加 0xbb8;對 D3TXT 檔即本地 rec)。
flags 低 3 bit = 可見性旗標(區分劇情前/後 NPC 變體)。
★ 對話檔 = D3TXT0<bank>,**bank = section header +0x17 的 byte**(寫在 CTY 檔裡;
load_cty 0x4526 ax=[0xb58]=section+0x17 → 0x2bd5 載 d3txtNN.txt)。純靜態、資料驅動,
無需 overlay/DOSBox。每 section 各自有自己的 bank。
產出:docs/data/npc_dialogue.json(全城 NPC 記錄 + 渲染對話)+ stdout(markdown 樣本)。"""
import struct,json,glob,os
def u16(b,o): return b[o]|(b[o+1]<<8)
ROOT=os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
gm=json.load(open(os.path.join(ROOT,'docs/data/glyph_unicode_map.json')))
VARNAME={0xfffb:'{名}',0xfffa:'{道具}',0xfff9:'{數}',0xfff6:'{名}',0xfff5:'{名}',0xffed:'{字}'}
HASP={0xfffb,0xfffa,0xfff9,0xfff6,0xfff5,0xffed}
_cache={}
def bank_file(bank):
    if bank<1 or bank>9: return None
    if bank not in _cache:
        fn=os.path.join(ROOT,'assets_raw/D3TXT0%d.TXT'%bank)
        b=open(fn,'rb').read(); n=u16(b,0)//2
        _cache[bank]=(b,struct.unpack_from('<%dH'%n,b,0))
    return _cache[bank]
def render(bank,r):
    bp=bank_file(bank)
    if bp is None: return None
    b,pt=bp
    if r>=len(pt): return '(rec 超範圍)'
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
def parse(path):
    b=open(path,'rb').read(); w0=u16(b,0); out=[]
    if w0==0 or w0>=len(b) or w0%2: return out
    for s in range(w0//2):
        sb=u16(b,s*2)
        if sb==0xffff or sb+0x18>len(b): continue
        bank=b[sb+0x17]                      # ★ section header +0x17 = 對話 bank
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
            rec={'sec':s,'bank':bank,'x':X,'y':Y,'sprite':spr,'sub':sub,'kind':kind,
                 'dlg':b4,'vis_flag':fl&7}
            if kind=='talk':
                rec['text']=render(bank,b4)
            out.append(rec)
    return out
data={}
for p in sorted(glob.glob(os.path.join(ROOT,'assets_raw/CTY*.DAT'))):
    cty=int(os.path.basename(p)[3:5]); ns=parse(p)
    if ns: data['CTY%02d'%cty]=ns
json.dump(data, open(os.path.join(ROOT,'docs/data/npc_dialogue.json'),'w'), ensure_ascii=False, indent=0)

tot=sum(len(v) for v in data.values())
talk=sum(1 for v in data.values() for x in v if x['kind']=='talk')
print('# NPC 對話逐句 mapping(獨立資料)\n')
print('全 %d 城共 %d NPC(對話型 %d)。對話檔 = D3TXT0<bank>,bank=section header +0x17。'%(len(data),tot,talk))
print('完整記錄(含逐句對話)見 `docs/data/npc_dialogue.json`。\n')
SAMPLE=[(0,'阿里阿罕'),(2,'羅馬利亞'),(17,'達瑪神殿(轉職)')]
for cty,nm in SAMPLE:
    key='CTY%02d'%cty
    if key not in data: continue
    bk=data[key][0]['bank']
    print('## CTY%02d %s(對話檔 D3TXT0%d)\n'%(cty,nm,bk))
    print('| sec | 位置 | dlg | 可見 | 對話 |')
    print('|---|---|---|---|---|')
    seen=0
    for x in data[key]:
        if x['kind']!='talk': continue
        seen+=1
        if seen>8: break
        print('| %d | (%d,%d) | 0x%02x | 0x%02x | %s |'%(x['sec'],x['x'],x['y'],x['dlg'],x['vis_flag'],(x['text'] or '')[:54]))
    print()
