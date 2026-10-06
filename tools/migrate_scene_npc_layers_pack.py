"""Docker-only: content0.1.104 -> 0.1.105; schema and gameplay data unchanged.

Arguments: writable pack directory, read-only original assets. READY docs/188.
Only the manifest version and CTY00 scene-layer evidence are updated.
"""
from pathlib import Path
import hashlib, json, os, sys

root, assets = map(Path, sys.argv[1:])
paths = [root / 'manifest.json', *sorted((root / 'data').glob('*.json'))]
assert len(paths) == 9 and all(p.is_file() and p.stat().st_uid == os.getuid() == 1000 for p in paths)
raws = {p: p.read_text() for p in paths}
objects = {p: json.loads(s) for p, s in raws.items()}
assert all(o['schema_version'] == '0.32.0' for o in objects.values())
manifest, interface = root / 'manifest.json', root / 'data/interface.json'
assert objects[manifest]['content_version'] == '0.1.104'
exe, cty = (assets / 'DQ3.EXE').read_bytes(), (assets / 'CTY00.DAT').read_bytes()
assert len(exe) == 115282 and hashlib.sha256(exe).hexdigest() == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(cty).hexdigest() == 'ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836'
for start, end, expected in ((0x11943, 0x1196b, 'c60679250090c6067a250090c6067b25009080e3c080fb00740e881e7925881e7a25c6067b250090'), (0x11e19, 0x11e1d, '8a1e570b'), (0x11e25, 0x11e29, '8a1e560b'), (0x11e33, 0x11e36, 'e87100')):
    assert exe[start-0xec90:end-0xec90].hex() == expected
layers = objects[interface]['scene_tile_layers']
entries = [s for s in layers if s['cty'] == 0 and s['section'] == 0]
assert len(entries) == 1 and entries[0]['mode'] == 'player_cell_layer'
evidence = entries[0]['evidence']
assert evidence['level'] == 'D3'
evidence['source'] += '; dosgolem2f44a68 normal422 source b7d0b1534a8b1055b57d78de8c343d12b736f355cf17584211b24dc2b84bcf3e; readonly native layer source 8d01dbfd57b81e5a077678d42dc9fff2fb49d8b25aad4c26aa455c6736def8ec'
evidence['address'] += ';0x11943..0x11965;0x11E07..0x11E37;0x11EA7..0x11ECF'
evidence['consumer'] += ';同一cell layer比較先於NPC adapter／BLS consumer，異層人物只畫替代背景'
evidence['note'] = '原正常回程186來源347600d38f062e111a08a852a0cdf2a8937b959827c37ff984f3dca44a732a75保留；新增正常418..422的八筆native分支限定確認內外圖層NPC遮蔽。正常422屋內兩NPC略過，完整動畫與戶外NPC位置仍未知。11943..11965 writer為static strong；不冒稱每次移動均經此入口。'
objects[manifest]['content_version'] = '0.1.105'
updated = raws[manifest].replace('"content_version": "0.1.104"', '"content_version": "0.1.105"')
assert json.loads(updated) == objects[manifest]
manifest.write_text(updated)
s = raws[interface]
at = s.index('[', s.index('"scene_tile_layers":'))
_, length = json.JSONDecoder().raw_decode(s[at:])
updated = s[:at] + json.dumps(layers, ensure_ascii=False, indent=2).replace('\n', '\n  ') + s[at+length:]
assert json.loads(updated) == objects[interface]
interface.write_text(updated)
for p in paths:
    if p not in (manifest, interface): assert p.read_text() == raws[p]
print('Content0.1.105 rebuilt; schema0.32.0; seven other JSON unchanged.')
