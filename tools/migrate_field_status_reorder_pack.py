"""Docker-only: schema0.27.0/content0.1.99 -> reviewed single-member response.

Arguments: writable pack, read-only original assets. READY: docs/188; JSON: docs/84.
"""
from pathlib import Path
import copy, hashlib, json, os, struct, sys
root, assets = map(Path, sys.argv[1:])
paths = [root/'manifest.json', *sorted((root/'data').glob('*.json'))]
assert len(paths) == 9 and all(p.is_file() and p.stat().st_uid == os.getuid() == 1000 for p in paths)
raws = {p: p.read_text() for p in paths}
objects = {p: json.loads(s) for p, s in raws.items()}
assert all(o['schema_version'] == '0.27.0' for o in objects.values())
assert objects[root/'manifest.json']['content_version'] == '0.1.99'
exe = (assets/'DQ3.EXE').read_bytes()
txt = (assets/'D3TXT00.TXT').read_bytes()
assert len(exe) == 115282 and hashlib.sha256(exe).hexdigest() == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(txt).hexdigest() == '38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
def code(a, b): return exe[a-0xec90:b-0xec90]
assert code(0x18685, 0x1869b).hex() == '803e77500174088d36143fe8506ec3bf0a02e889c9c3'
assert code(0x15010, 0x15015).hex() == '9adb000411'
assert code(0x21286, 0x21298).hex() == '81ffb80b7d0cd1e71ea12e258ed88b351fc3'
record = struct.unpack_from('<H', code(0x18694, 0x18697), 1)[0]
raw_window = exe[0x19fae:0x19fae+24]
assert raw_window.hex() == '0b011300ee002c0060009401000000000000000000000c09'
words = struct.unpack('<12H', raw_window)
evidence = dict(level='D3', source_kind='exe',
 source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; D3TXT00.TXT SHA256 38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b; dosgolem2f44a68 normal298 source e06a5e419d255303140c1516af0aa09021bbb967e94b0db1e0db89d1aca90370',
 address_space='linear', address='IDA9.4 18685..1869B;15002..15037;21286;21414..215EE;DGROUP3E6E/file19FAE',
 consumer='status third entry -> single-member DI020A -> 15023 -> text table/21414 -> inline216D8 -> final21133 -> field1997C',
 doc='docs/188-opening-escort-to-castle-spec.md',
 note='confirmed限定正常294..298單人信件兩頁、返回與下一步；原版引用此文字的原因unknown，多人排序未驗收。字模／捲動hold與等待閃爍時間沿既有hardware-spec approximation。')
interface = objects[root/'data/interface.json']
menu = interface['field_status_menu']
assert 'reorder' not in menu
presentation = copy.deepcopy(interface['field_items']['give_presentation'])
presentation['raw_window'] = dict(id='dq3:raw_window.field_status_reorder', flags=words[0]>>8,
 x=words[1], y=words[2], width=words[3], height=words[4], address='DGROUP:0x3e6e')
presentation['window']['id'] = 'dq3:window.field_status_reorder'
presentation['evidence'] = copy.deepcopy(evidence)
presentation['window']['evidence'] = copy.deepcopy(evidence)
presentation['shadow']['evidence'] = copy.deepcopy(evidence)
w = presentation['window']
assert (w['x'],w['y'],w['width'],w['height']) == (words[1]*8,words[2],words[3]*8,words[4])
assert (w['text_inset_x'],w['text_inset_y'],w['columns'],w['lines_per_page'],presentation['glyph_step_x']) == (16,16,20,4,24)
flow = copy.deepcopy(interface['opening_prelude']['text_flow'])
indicator = copy.deepcopy(interface['opening_prelude']['wait_indicator'])
assert (flow['mode'],flow['scroll_step_pixels'],flow['scroll_steps'],flow['scroll_hold_frames']) == ('retained_rows',4,4,1)
flow['evidence'] = copy.deepcopy(evidence)
flow['evidence']['address'] = 'IDA9.4 2149A..214BA;21501..21593;219FE..21AA6'
flow['evidence']['consumer'] = 'shared21414/219F4 retained four-row region; normal294/295 complete RGB0 DRAFT'
indicator['evidence'] = copy.deepcopy(evidence)
indicator['evidence']['address'] = 'IDA9.4 21558..21593;216C3..21726'
indicator['evidence']['consumer'] = 'inline216D8 cursor-row indicator; final21133 has no inline indicator'
text_id = 'dq3:text.field.status_reorder.single_member'
menu['reorder'] = dict(single_member_text_id=text_id, presentation=presentation, text_flow=flow,
 wait_indicator=indicator, scope='single_member', return_mode='fresh_key_to_field', evidence=evidence)
start,end = struct.unpack_from('<HH',txt,record*2)
codes = list(struct.unpack('<'+'H'*((end-start)//2),txt[start:end]))
assert codes.pop() == 65535 and codes.count(65532) == 1
glyphs = json.loads(Path('/repo/docs/data/glyph_unicode_map.json').read_text())
value = ''.join('\n' if c in (65534,65533) else '\n\n' if c == 65532 else glyphs[str(c)] for c in codes)
new = dict(id=text_id, value=value, glyph_codes=codes,
 layout=dict(kind='dialogue',columns=w['columns'],lines_per_page=w['lines_per_page']),
 source=dict(kind='legacy_record',file='D3TXT00.TXT',record=record),evidence=evidence)
definitions = objects[root/'data/texts.json']['definitions']
assert not any(d['id'] == text_id for d in definitions)
definitions.append(new)
for path,obj in objects.items():
 s = raws[path]
 if path.name == 'interface.json':
  start = s.index('"field_status_menu": ') + len('"field_status_menu": ')
  _,length = json.JSONDecoder().raw_decode(s[start:])
  s = s[:start] + json.dumps(menu,ensure_ascii=False,indent=2).replace('\n','\n  ') + s[start+length:]
 if path.name == 'texts.json':
  start = s.index('"definitions": ') + len('"definitions": ')
  _,length = json.JSONDecoder().raw_decode(s[start:]); end = start+length-1
  assert s[end] == ']'
  addition = '\n'.join('    '+l for l in json.dumps(new,ensure_ascii=False,indent=2).splitlines())
  s = s[:end].rstrip() + ',\n'+addition+'\n  '+s[end:]
 obj['schema_version'] = '0.28.0'
 s = s.replace('"schema_version": "0.27.0"','"schema_version": "0.28.0"')
 if path.name == 'manifest.json':
  obj['content_version'] = '0.1.100'
  s = s.replace('"content_version": "0.1.99"','"content_version": "0.1.100"')
 assert json.loads(s) == obj
 path.write_text(s)
print('Rebuilt nine JSON: schema0.28.0/content0.1.100; single-member reorder response.')
