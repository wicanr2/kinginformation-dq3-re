"""Docker-only: upgrade clean schema0.21.0/content0.1.93 native give prompt.

Arguments: writable pack directory and read-only original assets directory.
Contract, READY evidence and routing: docs/84 and docs/188.
"""
from pathlib import Path
import copy, hashlib, json, os, struct, sys

root, assets = map(Path, sys.argv[1:])
paths = [root / 'manifest.json', *sorted((root / 'data').glob('*.json'))]
assert len(paths) == 9 and all(p.is_file() and p.stat().st_uid == os.getuid() == 1000 for p in paths)
old = [p.read_text() for p in paths]
objects = [json.loads(s) for s in old]
assert all(d['schema_version'] == '0.21.0' for d in objects) and objects[0]['content_version'] == '0.1.93'
exe = (assets / 'DQ3.EXE').read_bytes()
assert len(exe) == 115282 and hashlib.sha256(exe).hexdigest() == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
text = (assets / 'D3TXT00.TXT').read_bytes()
assert hashlib.sha256(text).hexdigest() == '38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
raw = exe[0x19fae:0x19fae+30]
assert raw.hex() == '0b011300ee002c0060009401000000000000000000000c0936000e001800'
w = struct.unpack('<15H', raw)
interface = objects[paths.index(root / 'data/interface.json')]
assert 'give_presentation' not in interface['field_items']
ev = {'level': 'D3', 'source_kind': 'exe', 'source': 'DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; dosgolem2f44a68正常230包只讀來源3312cc33', 'address_space': 'linear', 'address': 'IDA9.4 139AB..139B6;15008..15019;21414..214FE;21133', 'consumer': 'field item self-give: preserved parent windows -> DGROUP3E6E / record308 -> physical-word rotation -> fresh read-key', 'doc': 'docs/188-opening-escort-to-castle-spec.md', 'note': '限定單持有者正常225/226；geometry/glyph/preserved UI confirmed。PIT字模hold另為hardware-spec approximation；不外推動畫時鐘或多人。'}
p = interface['opening_prelude']
window = copy.deepcopy(p['window'])
window.update(id='dq3:window.field_item_give_prompt', x=w[1]*8, y=w[2], width=w[3]*8, height=w[4], evidence=ev)
assert (window['text_inset_x'],window['text_inset_y'],p['glyph_step_x'],p['variable_code_words']) == (16,16,24,1)
shadow = copy.deepcopy(interface['field_save_load']['shadow'])
shadow['evidence'] = dict(ev, address='IDA9.4 1FC57..1FCC6', consumer='native window VGA word-latch AND shadow over pre-transaction UI')
prompt = {'raw_window': {'id': 'dq3:raw_window.field_item_give_prompt', 'flags': w[0]>>8, 'x': w[1], 'y': w[2], 'width': w[3], 'height': w[4], 'address': 'DGROUP:0x3e6e'}, 'window': window, 'frame_text_id': p['frame_text_id'], 'glyph_step_x': p['glyph_step_x'], 'variable_code_words': p['variable_code_words'], 'foreground_rgb': p['foreground_rgb'], 'backdrop_rgb': p['backdrop_rgb'], 'shadow': shadow, 'evidence': ev}
texts = objects[paths.index(root/'data/texts.json')]['definitions']
for id, record in [(prompt['frame_text_id'],w[5]),(interface['field_items']['text_ids']['give_prompt'],308)]:
 d = next(d for d in texts if d['id'] == id)
 assert d['source']['record'] == record
 start,end = struct.unpack_from('<HH',text,record*2)
 assert list(struct.unpack('<'+'H'*((end-start)//2),text[start:end])) == d['glyph_codes']+[65535]
interface['field_items']['give_presentation'] = prompt
outputs=[]
for path,obj,raw in zip(paths,objects,old):
 obj['schema_version']='0.22.0'
 out=raw.replace('"schema_version": "0.21.0"','"schema_version": "0.22.0"')
 if path.name=='manifest.json':
  obj['content_version']='0.1.94'
  out=out.replace('"content_version": "0.1.93"','"content_version": "0.1.94"')
 if path.name=='interface.json':
  end=out.rindex('\n  }')
  fragment=json.dumps(prompt,ensure_ascii=False,indent=2).splitlines()
  out=out[:end]+',\n    "give_presentation": '+fragment[0]+'\n'+'\n'.join('    '+line for line in fragment[1:])+out[end:]
 assert json.loads(out)==obj
 outputs.append(out)
for path,out in zip(paths,outputs): path.write_text(out)
print('Rebuilt nine JSON: schema0.22.0/content0.1.94; reviewed single-owner native give prompt.')
