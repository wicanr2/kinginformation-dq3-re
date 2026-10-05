"""Docker-only: schema0.22.0/content0.1.94 -> native no-effect use.

Arguments: writable clean pack, read-only original assets, glyph Unicode map.
READY contract and routing: docs/188 and docs/84.
"""
from pathlib import Path
import hashlib
import json
import os
import re
import struct
import sys

root, assets, mapping = map(Path, sys.argv[1:])
paths = [root/'manifest.json', *sorted((root/'data').glob('*.json'))]
assert len(paths) == 9 and all(p.is_file() and p.stat().st_uid == os.getuid() == 1000 for p in paths)
raws = {p: p.read_text() for p in paths}
objects = {p: json.loads(s) for p, s in raws.items()}
assert all(d['schema_version'] == '0.22.0' for d in objects.values())
assert objects[root/'manifest.json']['content_version'] == '0.1.94'
def read(name, hash):
    data = (assets/name).read_bytes()
    assert hashlib.sha256(data).hexdigest() == hash, name
    return data
exe = read('DQ3.EXE', '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c')
items = read('ITEM.DAT', '7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d')
txt = read('D3TXT00.TXT', '38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b')
assert len(exe) == 115282 and len(items) == 896
assert exe[0x13989-0xec90:0x1398b-0xec90] == bytes.fromhex('b080')
assert exe[0x1399a-0xec90:0x1399f-0xec90] == bytes.fromhex('f687fe0101')
assert exe[0x13c7d-0xec90] == 8
evidence = dict(level='D3', source_kind='exe', source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; ITEM.DAT; dosgolem2f44a68 normal233 source c11efcbd', address_space='linear', address='IDA9.4 13942..139AA;13C6D..13D42;21414..215AE', consumer='healthy sole primary hero: native use introduction -> no-effect result -> fresh key -> field', doc='docs/188-opening-escort-to-castle-spec.md', note='正常木棒使用動態confirmed；全128項metadata逐record核對不等於每件玩家路線。非健康、多持有者、特殊效果與拒絕gate未由本切片驗收。')
characters = objects[root/'data/characters.json']
metadata = characters['item_storage']['items']
assert len(metadata) == len(items)//7 and all('single_hero_no_effect' not in x for x in metadata)
for code, d in enumerate(metadata):
    r = items[code*7:code*7+7]
    d['single_hero_no_effect'] = r[4]&8 == 0 and r[5]&1 == 0 and r[6]&0x80 != 0
characters['item_storage']['evidence']['note'] = characters['item_storage']['evidence'].get('note', '')+'; single_hero_no_effect: IDA9.4 1397C..139A7/13C77..13C7B，正常木棒233來源c11efcbd；docs/188。'
interface = objects[root/'data/interface.json']
s = interface['field_items']
assert 'use_no_effect' not in s
s['use_no_effect'] = dict(intro_text_id='dq3:text.field.items.use_intro', result_text_id='dq3:text.item_use.no_effect', actor_variable_code=65531, item_variable_code=65529, evidence=evidence)
definitions = objects[root/'data/texts.json']['definitions']
def record(n):
    start, end = struct.unpack_from('<HH', txt, n*2)
    values = list(struct.unpack('<'+'H'*((end-start)//2), txt[start:end]))
    assert values[-1] == 65535
    return values[:-1]
codes = record(273)
assert codes == [65531, 210, 412, 65529, 56]
assert record(341) == [508, 399, 435, 436, 494, 410, 147, 431, 273, 56]
assert not any(d['id'] == s['use_no_effect']['intro_text_id'] for d in definitions)
definitions.append(dict(id=s['use_no_effect']['intro_text_id'], value='{actor}使用{item}。', glyph_codes=codes, layout=dict(kind='dialogue', columns=20, lines_per_page=4), source=dict(kind='legacy_record', file='D3TXT00.TXT', record=273), evidence=evidence))
noeffect = next(d for d in definitions if d['id'] == s['use_no_effect']['result_text_id'])
assert noeffect['source']['record'] == 341 and noeffect['glyph_codes'] == record(341)
old_noeffect_value = noeffect['value']
unicode_map = json.loads(mapping.read_text())
noeffect['value'] = ''.join(unicode_map[str(c)] for c in record(341))
assert noeffect['value'] == '但是什麼也沒有發生。'

def replace_object(raw, name, value):
    key = json.dumps(name)+': '
    assert raw.count(key) == 1
    start = raw.index(key)+len(key)
    _, length = json.JSONDecoder().raw_decode(raw[start:])
    indent = start-len(key)-raw.rfind('\n', 0, start)-1
    serialized = json.dumps(value, ensure_ascii=False, indent=2).splitlines()
    replacement = serialized[0]+''.join('\n'+' '*indent+line for line in serialized[1:])
    return raw[:start]+replacement+raw[start+length:]

outputs = {}
for path, obj in objects.items():
    raw = raws[path]
    if path.name == 'characters.json':
        index = iter(metadata)
        def insert_metadata(match):
            return match.group(0)+', "single_hero_no_effect": '+json.dumps(next(index)['single_hero_no_effect'])
        raw, count = re.subn(r'"drop_forbidden": (?:true|false)', insert_metadata, raw)
        assert count == len(metadata)
        old_note = json.loads(raw)['item_storage']['evidence']['note']
        new_note = characters['item_storage']['evidence']['note']
        assert raw.count(json.dumps(old_note, ensure_ascii=False)) == 1
        raw = raw.replace(json.dumps(old_note, ensure_ascii=False), json.dumps(new_note, ensure_ascii=False))
    if path.name == 'interface.json': raw = replace_object(raw, 'field_items', s)
    if path.name == 'texts.json':
        old_value = '"value": '+json.dumps(old_noeffect_value, ensure_ascii=False)
        assert raw.count(old_value) == 1
        raw = raw.replace(old_value, '"value": '+json.dumps(noeffect['value'], ensure_ascii=False))
        start = raw.index('"definitions": ')+len('"definitions": ')
        _, length = json.JSONDecoder().raw_decode(raw[start:])
        end = start+length-1
        assert raw[end] == ']'
        fragment = json.dumps(definitions[-1], ensure_ascii=False, indent=2).splitlines()
        raw = raw[:end].rstrip()+',\n'+'\n'.join('    '+l for l in fragment)+'\n  '+raw[end:]
    obj['schema_version'] = '0.23.0'
    raw = raw.replace('"schema_version": "0.22.0"', '"schema_version": "0.23.0"')
    if path.name == 'manifest.json':
        obj['content_version'] = '0.1.95'
        raw = raw.replace('"content_version": "0.1.94"', '"content_version": "0.1.95"')
    assert json.loads(raw) == obj
    outputs[path] = raw
for path, value in outputs.items(): path.write_text(value)
print('Rebuilt nine JSON: schema0.23.0/content0.1.95; healthy sole hero no-effect use.')
