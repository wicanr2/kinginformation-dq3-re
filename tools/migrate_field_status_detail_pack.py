"""Docker-only: schema0.25.0/content0.1.97 -> native status detail.

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
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==os.getuid()==1000 for p in paths)
raws = {p:p.read_text() for p in paths}
objects = {p:json.loads(s) for p,s in raws.items()}
assert all(d['schema_version']=='0.25.0' for d in objects.values())
assert objects[root/'manifest.json']['content_version']=='0.1.97'
def read(name, sha):
    b=(assets/name).read_bytes()
    assert hashlib.sha256(b).hexdigest()==sha,name
    return b
exe=read('DQ3.EXE','5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c')
txt=read('D3TXT00.TXT','38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b')
assert len(exe)==115282
raw=exe[0x19ee8:0x19ee8+28]
assert raw.hex()=='010313002e002c00c00097010000000000004e830000000000000203'
assert exe[0x18498-0xec90:0x184a1-0xec90].hex()=='9adb000411e80100c3'
consumer=exe[0x213c4-0xec90:0x21414-0xec90]
assert consumer.count(bytes.fromhex('83c210'))==2
step=consumer[consumer.index(bytes.fromhex('83c210'))+2]
interface=objects[root/'data/interface.json']
menu=interface['field_status_menu']
assert 'detail' not in menu
window=interface['new_game_geometry']['raster']['ability']
record=struct.unpack_from('<H',raw,10)[0]
start,end=struct.unpack_from('<HH',txt,record*2)
text=objects[root/'data/texts.json']['definitions']
d=next(d for d in text if d['id']==window['text_id'])
assert d['source']['record']==record and d['glyph_codes']==list(struct.unpack('<'+'H'*((end-start)//2),txt[start:end]))[:-1]
detail=dict(window=window,equipment_row_step=step,scope='single_healthy_hero_without_spells',return_mode='fresh_key_to_field',evidence=dict(level='D3',source_kind='exe',source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; D3TXT00.TXT; dosgolem2f44a68 normal280 source edfad334',address_space='dgroup',address='0x3da8/file0x19ee8/IDAlinear0x28b78',consumer='IDA9.4 linear1834E -> 2111B -> 184A1 -> 1F4E3 restore -> normal278 field1997C',doc='docs/188-opening-escort-to-castle-spec.md',note='confirmed限定正常健康未學咒文主角277詳細頁、278新Enter返回與279..280左右行走；其他角色、裝備及咒文頁另驗。列距213C4 text consumer；原生能力幾何重用已驗契約。'))
menu['detail']=detail
outputs={}
for p,obj in objects.items():
    raw=raws[p]
    if p.name=='interface.json':
        start=raw.index('"field_status_menu": ')+len('"field_status_menu": ')
        _,length=json.JSONDecoder().raw_decode(raw[start:])
        end=start+length
        raw=raw[:start]+json.dumps(menu,ensure_ascii=False,indent=2).replace('\n','\n  ')+raw[end:]
    obj['schema_version']='0.26.0'
    raw=raw.replace('"schema_version": "0.25.0"','"schema_version": "0.26.0"')
    if p.name=='manifest.json':
        obj['content_version']='0.1.98'
        raw=raw.replace('"content_version": "0.1.97"','"content_version": "0.1.98"')
    assert json.loads(raw)==obj
    outputs[p]=raw
for p,raw in outputs.items():p.write_text(raw)
print('Rebuilt nine JSON: schema0.26.0/content0.1.98; native detail and fresh-key return.')
