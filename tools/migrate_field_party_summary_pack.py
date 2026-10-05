"""Docker-only: schema0.26.0/content0.1.98 -> reviewed party summary.

Arguments: writable pack, read-only original assets. READY: docs/188; JSON: docs/84.
"""
from pathlib import Path
import hashlib, json, os, struct, sys
root,assets=map(Path,sys.argv[1:])
paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==os.getuid()==1000 for p in paths)
raws={p:p.read_text() for p in paths};objects={p:json.loads(s) for p,s in raws.items()}
assert all(o['schema_version']=='0.26.0' for o in objects.values())
assert objects[root/'manifest.json']['content_version']=='0.1.98'
exe=(assets/'DQ3.EXE').read_bytes();txt=(assets/'D3TXT00.TXT').read_bytes()
assert len(exe)==115282 and hashlib.sha256(exe).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(txt).hexdigest()=='38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
def code(a,b):return exe[a-0xec90:b-0xec90]
assert code(0x17cdf,0x17cec).hex()=='b30af6e3050400a3a23ea3023f'
assert code(0x1fc57,0x1fc5f).hex()=='8a4401a8087401c3'
assert code(0x1867f,0x18685).hex()=='9adb000411c3'
evidence=dict(level='D3',source_kind='exe',source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; D3TXT00.TXT; dosgolem2f44a68 normal288 source fac03247',address_space='dgroup',address='0x3e84/file0x19fc4 and 0x3efc/file0x1a03c',consumer='IDA9.4 linear17C83 -> 17CE9 width; 185EF -> 1F590/18847 -> 1F4E3 -> 18610 -> 2111B; normal285..288',doc='docs/188-opening-escort-to-castle-spec.md',note='confirmed限定正常單人285全體狀況、286新Enter與287..288行走。多人物欄位有原始loop與width writer證據，尚無正常多人同狀態畫面收據。')
def raw(offset,id):
    b=exe[0x16140+offset:0x16140+offset+24];v=struct.unpack('<12H',b)
    return dict(id=id,flags=v[0]>>8,x=v[1],y=v[2],width=v[3],height=v[4],address=hex(offset)),v
money,m=raw(0x3e84,'dq3:raw_window.party_summary_money')
body,b=raw(0x3efc,'dq3:raw_window.party_summary')
assert tuple(m[:6])==(2316,54,14,24,48,406)
assert tuple(b[:12])==(784,19,62,44,144,408,4,409,410,34320,254,10)
base=code(0x17ce3,0x17ce6)[1];step=b[11]*8;capacity=(body['width']-base)//b[11]
number=lambda x,y,d:dict(x=x,y=y,digits=d)
anchor=lambda x,y,dx,dy:dict(x=x,y=y,step_x=dx,step_y=dy)
ids={n:'dq3:text.field.party_summary.'+role for n,role in [(406,'money'),(408,'left'),(409,'column'),(410,'right')]}
summary=dict(money_window=money,money_text_id=ids[m[5]],money_value=number((m[1]+4)*8,m[2]+16,8),raw_window=body,left_text_id=ids[b[5]],column_text_id=ids[b[7]],right_text_id=ids[b[8]],width_base=base,column_step=step,max_columns=capacity,column_origin=anchor((b[1]+2)*8,b[2],step,0),name=anchor((b[1]+2)*8,b[2]+16,16,0),name_limit=4,hp=number(b[1]*8,b[2]+32,5),max_hp=number(b[1]*8,b[2]+64,5),mp=number(b[1]*8,b[2]+80,5),max_mp=number(b[1]*8,b[2]+112,5),return_mode='fresh_key_to_field',evidence=evidence)
menu=objects[root/'data/interface.json']['field_status_menu'];assert 'summary' not in menu;menu['summary']=summary
definitions=objects[root/'data/texts.json']['definitions'];new=[]
glyphs={12:' ',21:'G',22:'H',27:'M',47:'┌',48:'─',49:'┐',50:'└',51:'┘',54:'│',65534:'\n'}
for n in ids:
    start,end=struct.unpack_from('<HH',txt,n*2);codes=list(struct.unpack('<'+'H'*((end-start)//2),txt[start:end]));assert codes.pop()==65535
    rows=[[]]
    for c in codes:
        if c==65534:rows.append([])
        else:rows[-1].append(c)
    assert len(set(map(len,rows)))==1
    new.append(dict(id=ids[n],value=''.join(glyphs[c] for c in codes),glyph_codes=codes,layout=dict(kind='menu_record',columns=len(rows[0]),lines_per_page=len(rows)),source=dict(kind='legacy_record',file='D3TXT00.TXT',record=n),evidence=evidence))
assert not any(d['id'] in ids.values() for d in definitions);definitions.extend(new)
for p,obj in objects.items():
    s=raws[p]
    if p.name=='interface.json':
        start=s.index('"field_status_menu": ')+len('"field_status_menu": ');_,length=json.JSONDecoder().raw_decode(s[start:]);s=s[:start]+json.dumps(menu,ensure_ascii=False,indent=2).replace('\n','\n  ')+s[start+length:]
    if p.name=='texts.json':
        start=s.index('"definitions": ')+len('"definitions": ');_,length=json.JSONDecoder().raw_decode(s[start:]);end=start+length-1;assert s[end]==']'
        addition=',\n'.join('\n'.join('    '+l for l in json.dumps(d,ensure_ascii=False,indent=2).splitlines()) for d in new)
        s=s[:end].rstrip()+',\n'+addition+'\n  '+s[end:]
    obj['schema_version']='0.27.0';s=s.replace('"schema_version": "0.26.0"','"schema_version": "0.27.0"')
    if p.name=='manifest.json':obj['content_version']='0.1.99';s=s.replace('"content_version": "0.1.98"','"content_version": "0.1.99"')
    assert json.loads(s)==obj;p.write_text(s)
print('Rebuilt nine JSON: schema0.27.0/content0.1.99; native party summary.')
