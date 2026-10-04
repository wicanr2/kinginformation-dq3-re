"""於 Docker 從乾淨0.18.0 pack重建正常指令窗資料；入口 docs/84及docs/188。"""
from pathlib import Path
import hashlib
import json
import os
import struct
import sys

root, assets, mapping = map(Path, sys.argv[1:])
paths = [root / 'manifest.json', *sorted((root / 'data').glob('*.json'))]
assert len(paths) == 9 and all(p.is_file() and p.stat().st_uid == os.getuid() == 1000 for p in paths)
old = [p.read_text() for p in paths]
objects = [json.loads(s) for s in old]
assert all(d['schema_version'] == '0.18.0' for d in objects) and objects[0]['content_version'] == '0.1.90'
exe = (assets / 'DQ3.EXE').read_bytes()
text = (assets / 'D3TXT00.TXT').read_bytes()
assert len(exe) == 115282 and hashlib.sha256(exe).hexdigest() == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(text).hexdigest() == '38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
end = int.from_bytes(text[:2], 'little')
ptrs = struct.unpack('<' + 'H' * (end // 2), text[:end])
assert len(ptrs) - 1 == 759
w = struct.unpack_from('<30H', exe, 0x19eac)
assert w[:12] == (1792, 19, 30, 24, 80, 400, 0, 0, 0, 0, 6, 2)
record = text[ptrs[w[5]]:ptrs[w[5] + 1]]
codes = list(struct.unpack('<' + 'H' * (len(record) // 2), record))
assert codes[-1] == 65535
codes = codes[:-1]
assert len(codes) == 64 and codes.count(65534) == 4
glyphs = json.loads(mapping.read_text())
value = ''.join('\n' if c == 65534 else glyphs[str(c)] for c in codes)
interface = objects[paths.index(root / 'data/interface.json')]
texts = objects[paths.index(root / 'data/texts.json')]
assert 'field_command_menu' not in interface
ev = {'level': 'D3', 'source_kind': 'exe',
      'source': 'DQ3.EXE + dosgolem2f44a68正常194..205，source5b2a78c152e77f1ddbfd850af0cc129837bf206cd9a15731529927d21b12978a',
      'address_space': 'dgroup', 'address': '0x3d6c/file0x19eac/IDAlinear0x28b3c',
      'consumer': 'IDAlinear17C83 → 1F4E3 → 1F779/1F908；正常Space六指令窗',
      'doc': 'docs/188-opening-escort-to-castle-spec.md',
      'note': '導航正常1→2→3→4→3→6→3→6；raw258F8，cursor glyph11；Enter靜態strong，Space與Esc正常閉合。'}
entries = []
for i in range(w[10]):
    x, y, callback = w[12+i*3:15+i*3]
    row, col = (y-w[2]) // 16, (x-w[1]) // 2 + 1
    start = row * 13 + col
    matches = [role for role, label in interface['field_command_labels'].items()
               if role != 'title' and codes[start] == label['primary_glyph'] and codes[start+2] == label['secondary_glyph']]
    assert len(matches) == 1
    entries.append({'command': matches[0], 'x': x*8, 'y': y, 'callback_raw': callback})
s = {'id': 'dq3:window.field_command_menu',
     'raw_window': {'id': 'dq3:raw_window.field_command_menu', 'flags': w[0] >> 8,
                    'x': w[1], 'y': w[2], 'width': w[3], 'height': w[4], 'address': '0x3d6c'},
     'text_id': 'dq3:text.field.command_frame', 'navigation': 'linear_two_columns',
     'cursor_glyph': 11, 'font_index': 8, 'entries': entries, 'evidence': ev}
d = {'id': s['text_id'], 'value': value, 'glyph_codes': codes,
     'layout': {'kind': 'menu_record', 'columns': 12, 'lines_per_page': 5},
     'source': {'kind': 'legacy_record', 'file': 'D3TXT00.TXT', 'record': w[5]},
     'evidence': dict(ev, source_kind='data_file', source='D3TXT00.TXT + D3TXT00.FON + 正常194畫面', address_space='record', address='400')}
assert not any(t['id'] == d['id'] for t in texts['definitions'])
interface['field_command_menu'] = s
texts['definitions'].append(d)
outputs = []
for p, obj, raw in zip(paths, objects, old):
    obj['schema_version'] = '0.19.0'
    out = raw.replace('"schema_version": "0.18.0"', '"schema_version": "0.19.0"')
    if p.name == 'manifest.json':
        obj['content_version'] = '0.1.91'
        out = out.replace('"content_version": "0.1.90"', '"content_version": "0.1.91"')
    if p.name == 'interface.json':
        end = out.rindex('\n}')
        fragment = json.dumps(s, ensure_ascii=False, indent=2).splitlines()
        out = out[:end] + ',\n  "field_command_menu": ' + fragment[0] + '\n' + '\n'.join('  '+line for line in fragment[1:]) + out[end:]
    if p.name == 'texts.json':
        end = out.rindex('\n  ]')
        out = out[:end] + ',\n' + '\n'.join('    '+line for line in json.dumps(d, ensure_ascii=False, indent=2).splitlines()) + out[end:]
    assert json.loads(out) == obj, p
    outputs.append(out)
for p, out in zip(paths, outputs):
    p.write_text(out)
    assert p.stat().st_uid == 1000
print('九份JSON重建：schema0.19.0／content0.1.91；正常六指令原始窗口／文字／順序')
