"""Docker-only: clean schema0.23.0/content0.1.95 -> reviewed native drop.

Arguments: writable pack, read-only original assets. READY: docs/188; JSON: docs/84.
"""
from pathlib import Path
import hashlib
import json
import os
import re
import struct
import sys

root, assets = map(Path, sys.argv[1:])
paths = [root/'manifest.json', *sorted((root/'data').glob('*.json'))]
assert len(paths) == 9 and all(p.is_file() and p.stat().st_uid == os.getuid() == 1000 for p in paths)
raws = {p:p.read_text() for p in paths}
objects = {p:json.loads(s) for p,s in raws.items()}
assert all(d['schema_version']=='0.23.0' for d in objects.values())
assert objects[root/'manifest.json']['content_version']=='0.1.95'
def read(name, sha):
    data=(assets/name).read_bytes()
    assert hashlib.sha256(data).hexdigest()==sha,name
    return data
exe=read('DQ3.EXE','5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c')
items=read('ITEM.DAT','7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d')
txt=read('D3TXT00.TXT','38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b')
assert len(exe)==115282 and len(items)==896
for linear, raw in ((0x13af3,'3d0000'),(0x13af8,'c704ff00'),(0x13afc,'bf1501'),(0x13b09,'bf1001')):
    expected=bytes.fromhex(raw);assert exe[linear-0xec90:linear-0xec90+len(expected)]==expected
evidence=dict(level='D3',source_kind='exe',source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; ITEM.DAT; dosgolem2f44a68 normal261 source5ca9fce6',address_space='linear',address='IDA9.4 13ABC..13B0F;13919;15023;21414;21133;18197',consumer='healthy sole hero: drop physical item -> native result or worn rejection -> fresh key -> field',doc='docs/188-opening-escort-to-castle-spec.md',note='正常兩次成功及穿戴拒絕confirmed；空格不壓縮。其他metadata及零價gate為原始資料D2／strong，不外推所有物品玩家流程。')
metadata=objects[root/'data/characters.json']['item_storage']['items']
assert len(metadata)==128 and all('drop_has_value' not in x for x in metadata)
for code,d in enumerate(metadata):d['drop_has_value']=struct.unpack_from('<H',items,code*7+2)[0]!=0
s=objects[root/'data/interface.json']['field_items'];assert 'drop' not in s
s['drop']=dict(success_text_id='dq3:text.field.items.drop_success',blocked_text_id='dq3:text.field.items.drop_blocked',actor_variable_code=65531,item_variable_code=65529,evidence=evidence)
definitions=objects[root/'data/texts.json']['definitions']
def record(n):
    start,end=struct.unpack_from('<HH',txt,n*2)
    values=list(struct.unpack('<'+'H'*((end-start)//2),txt[start:end]));assert values[-1]==65535
    return values[:-1]
assert record(277)==[65531,530,65529,65534,557,560,423,56]
assert record(272)==[398,546,539,380,228,408,557,56]
for key,n,value in (('success_text_id',277,'{actor}把{item}丟掉了。'),('blocked_text_id',272,'這個東西不能丟。')):
    assert not any(d['id']==s['drop'][key] for d in definitions)
    definitions.append(dict(id=s['drop'][key],value=value,glyph_codes=record(n),layout=dict(kind='dialogue',columns=20,lines_per_page=4),source=dict(kind='legacy_record',file='D3TXT00.TXT',record=n),evidence=evidence))
def replace_object(raw,name,value):
    key=json.dumps(name)+': ';assert raw.count(key)==1
    start=raw.index(key)+len(key);_,length=json.JSONDecoder().raw_decode(raw[start:])
    indent=start-len(key)-raw.rfind('\n',0,start)-1
    parts=json.dumps(value,ensure_ascii=False,indent=2).splitlines()
    replacement=parts[0]+''.join('\n'+' '*indent+line for line in parts[1:])
    return raw[:start]+replacement+raw[start+length:]
outputs={}
for p,obj in objects.items():
    raw=raws[p]
    if p.name=='characters.json':
        iterator=iter(metadata)
        def add(m):return m.group(0)+', "drop_has_value": '+json.dumps(next(iterator)['drop_has_value'])
        raw,count=re.subn(r'"drop_forbidden": (?:true|false)',add,raw);assert count==128
    if p.name=='interface.json':raw=replace_object(raw,'field_items',s)
    if p.name=='texts.json':
        start=raw.index('"definitions": ')+len('"definitions": ');_,length=json.JSONDecoder().raw_decode(raw[start:]);end=start+length-1
        assert raw[end]==']'
        additions=['\n'.join('    '+line for line in json.dumps(d,ensure_ascii=False,indent=2).splitlines()) for d in definitions[-2:]]
        raw=raw[:end].rstrip()+',\n'+',\n'.join(additions)+'\n  '+raw[end:]
    obj['schema_version']='0.24.0';raw=raw.replace('"schema_version": "0.23.0"','"schema_version": "0.24.0"')
    if p.name=='manifest.json':obj['content_version']='0.1.96';raw=raw.replace('"content_version": "0.1.95"','"content_version": "0.1.96"')
    assert json.loads(raw)==obj;outputs[p]=raw
for p,raw in outputs.items():p.write_text(raw)
print('Rebuilt nine JSON: schema0.24.0/content0.1.96; normal drop and worn rejection.')
