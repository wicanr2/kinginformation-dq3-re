"""Docker-only: schema0.24.0/content0.1.96 -> reviewed first status menu.

Arguments: writable pack, read-only original assets. READY: docs/188; JSON: docs/84.
"""
from pathlib import Path
import hashlib
import json
import os
import struct
import sys

root, assets = map(Path, sys.argv[1:])
paths = [root/'manifest.json', *sorted((root/'data').glob('*.json'))]
assert len(paths) == 9 and all(p.is_file() and p.stat().st_uid == os.getuid() == 1000 for p in paths)
raws = {p:p.read_text() for p in paths}
objects = {p:json.loads(s) for p,s in raws.items()}
assert all(d['schema_version']=='0.24.0' for d in objects.values())
assert objects[root/'manifest.json']['content_version']=='0.1.96'
def read(name, sha):
    data=(assets/name).read_bytes()
    assert hashlib.sha256(data).hexdigest()==sha,name
    return data
exe=read('DQ3.EXE','5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c')
txt=read('D3TXT00.TXT','38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b')
assert len(exe)==115282
assert exe[0x1830b-0xec90:0x18313-0xec90].hex()=='8d36b43ee8d171c3'
raw=exe[0x19ff4:0x19ff4+42]
word=lambda offset:struct.unpack_from('<H',raw,offset)[0]
assert list(struct.unpack('<21H',raw))==[782,15,62,20,80,405,0,0,0,0,3,1,17,78,33555,17,94,34287,17,110,34437]
evidence=dict(level='D3',source_kind='exe',source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; D3TXT00.TXT; dosgolem2f44a68 normal273 sourcea30e50ed',address_space='dgroup',address='0x3eb4/file0x19ff4/IDAlinear0x28c84',consumer='IDA9.4 linear18301 -> 1830B -> 1F4E3 -> 1F779/1F908; normal sole healthy hero first status selector',doc='docs/188-opening-escort-to-castle-spec.md',note='正常264..270三列游標與上下繞回、271 Esc、272..273下一步confirmed；選取詳細內容／全體／排序與其他角色分支未驗。原始callbacks只作證據，不執行。')
interface=objects[root/'data/interface.json']
assert 'field_status_menu' not in interface
menu=dict(id='dq3:window.field_status_menu',raw_window=dict(id='dq3:raw_window.field_status_menu',flags=word(0)>>8,x=word(2),y=word(4),width=word(6),height=word(8),address='0x3eb4'),text_id='dq3:text.field.status_menu',navigation='cyclic_single_column',cursor_glyph=interface['field_command_menu']['cursor_glyph'],font_index=interface['field_command_menu']['font_index'],entries=[dict(role=role,x=word(24+6*i)*8,y=word(26+6*i),callback_raw=word(28+6*i)) for i,role in enumerate(('detail','party_summary','reorder'))],evidence=evidence)
interface['field_status_menu']=menu
start,end=struct.unpack_from('<HH',txt,word(10)*2)
codes=list(struct.unpack('<'+'H'*((end-start)//2),txt[start:end]))
assert len(codes)==55 and codes[-1]==65535
codes=codes[:-1]
assert [codes[i] for i in (10,21,32,43)]==[65534]*4
definition=dict(id=menu['text_id'],value='┌┐┐┐狀┐況┐┐└\n│ 看各人的狀況 │\n│ 看全體的情形 │\n│ 重 新 排 序│\n┘┐┐┐┐┐┐┐┐├',glyph_codes=codes,layout=dict(kind='menu_record',columns=word(6)//2,lines_per_page=word(8)//16),source=dict(kind='legacy_record',file='D3TXT00.TXT',record=word(10)),evidence=evidence)
definitions=objects[root/'data/texts.json']['definitions']
assert not any(d['id']==definition['id'] for d in definitions)
definitions.append(definition)
outputs={}
for p,obj in objects.items():
    raw=raws[p]
    if p.name=='interface.json':
        assert raw.rstrip().endswith('}')
        parts=json.dumps(menu,ensure_ascii=False,indent=2).splitlines()
        raw=raw.rstrip()[:-1].rstrip()+',\n  "field_status_menu": '+parts[0]+''.join('\n  '+line for line in parts[1:])+'\n}\n'
    if p.name=='texts.json':
        start=raw.index('"definitions": ')+len('"definitions": ')
        _,length=json.JSONDecoder().raw_decode(raw[start:]);end=start+length-1
        assert raw[end]==']'
        addition='\n'.join('    '+line for line in json.dumps(definition,ensure_ascii=False,indent=2).splitlines())
        raw=raw[:end].rstrip()+',\n'+addition+'\n  '+raw[end:]
    obj['schema_version']='0.25.0';raw=raw.replace('"schema_version": "0.24.0"','"schema_version": "0.25.0"')
    if p.name=='manifest.json':
        obj['content_version']='0.1.97';raw=raw.replace('"content_version": "0.1.96"','"content_version": "0.1.97"')
    assert json.loads(raw)==obj;outputs[p]=raw
for p,raw in outputs.items():p.write_text(raw)
print('Rebuilt nine JSON: schema0.25.0/content0.1.97; first status selector only.')
