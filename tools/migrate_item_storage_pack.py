"""Docker-only: rebuild schema0.20.0 ordered item data from a clean0.19.0 pack.

Arguments: writable pack copy, read-only original asset directory. Entry: docs/84
and docs/188. Checks every input before any write; original assets remain private.
"""
from pathlib import Path
import hashlib
import json
import os
import sys

root, assets = map(Path, sys.argv[1:])
paths = [root / 'manifest.json', *sorted((root / 'data').glob('*.json'))]
assert len(paths) == 9 and all(p.is_file() and p.stat().st_uid == os.getuid() == 1000 for p in paths)
original = [p.read_text() for p in paths]
objects = [json.loads(raw) for raw in original]
assert all(obj['schema_version'] == '0.19.0' for obj in objects)
assert objects[0]['content_version'] == '0.1.91'
exe = (assets / 'DQ3.EXE').read_bytes()
item = (assets / 'ITEM.DAT').read_bytes()
assert len(exe) == 115282 and hashlib.sha256(exe).hexdigest() == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert len(item) == 896 and hashlib.sha256(item).hexdigest() == '7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d'
word = lambda offset: int.from_bytes(exe[offset:offset + 2], 'little')
# Original file offsets, reviewed IDA9.4 instruction bytes, not rebased bytes.
for offset, expected in (
    (0x4c99, '3dff00'), (0x4cab, '81269125ff00'),
    (0x4d78, 'a900e0'), (0x9408, 'b80080'),
    (0x940d, 'ba000e'), (0x9416, 'b80040'), (0x941e, 'b8ff7f2104'),
    (0x1c19, 'b90800c704ff0083c602e2f7'),
    (0x1c2d, 'b81e000d008089443a'),
    (0x1c94, 'b922008d360b5283c616b000880446e2fb8d360b52c6441501b81e000d00805683c63ab90800c704ff0083c602e2f75e89443a'),
):
    expected = bytes.fromhex(expected)
    assert exe[offset:offset + len(expected)] == expected, hex(offset)
empty, capacity = word(0x4c9a), word(0x1c1a)
assert word(0x1c1e) == empty
initial = [word(0x1c2e) | word(0x1c31)] + [empty] * (capacity - 1)
characters = objects[paths.index(root / 'data/characters.json')]
events = objects[paths.index(root / 'data/events.json')]
assert events['item_actions']['personal_inventory_slots'] == capacity
assert 'item_storage' not in characters
roles = characters['default_refs']
assert len(characters['defaults']) == len(roles) == 2
for default in characters['defaults']:
    assert default['id'] in roles.values()
    assert default['equipment'] == {'weapon': None, 'armor': initial[0] & word(0x4caf), 'shield': None, 'head': None}
    del default['equipment']
    default['item_words'] = initial.copy()
    default['evidence']['note'] = 'Canonical physical words; explicit empty slots retained; equipment is a derived view.'
    if default['id'] == roles['new_game_player']:
        default['evidence']['address'] = '0x1c19..0x1c36'
    else:
        default['evidence']['address'] = '0x1c94..0x1cc7'
metadata = []
for offset in range(0, len(item), 7):
    group = item[offset + 4] >> 5
    metadata.append({
        'equipment_part': {1: 0, 2: 1, 3: 3, 4: 2}.get(group, -1),
        'cursed_when_worn': bool(int.from_bytes(item[offset + 4:offset + 6], 'little') & word(0x940e)),
    })
characters['item_storage'] = {
    'encoding': {'empty': empty, 'code_mask': word(0x4caf), 'worn_mask': word(0x9409),
                 'curse_mask': word(0x9417), 'transfer_blocked_mask': word(0x4d79)},
    'part_count': 4, 'items': metadata,
    'evidence': {'level': 'D2', 'source_kind': 'exe', 'source': 'DQ3.EXE + ITEM.DAT',
                 'address_space': 'file', 'address': '0x4c99/0x4cab/0x4d78/0x9408..0x9423',
                 'consumer': 'IDA9.4 linear13929/1393B/13A08/18098..180B1; original item decoder; physical owner words',
                 'doc': 'docs/188-opening-escort-to-castle-spec.md',
                 'note': 'EXE sha2565178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; ITEM sha2567f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d; count validated against actual archive.'},
}
outputs = []
for path, obj, raw in zip(paths, objects, original):
    obj['schema_version'] = '0.20.0'
    out = raw.replace('"schema_version": "0.19.0"', '"schema_version": "0.20.0"')
    if path.name == 'manifest.json':
        obj['content_version'] = '0.1.92'
        out = out.replace('"content_version": "0.1.91"', '"content_version": "0.1.92"')
    if path.name == 'characters.json':
        out = json.dumps(obj, ensure_ascii=False, indent=2) + '\n'
    assert json.loads(out) == obj, path
    outputs.append(out)
for path, out in zip(paths, outputs):
    path.write_text(out)
    assert path.stat().st_uid == os.getuid() == 1000
print('Rebuilt nine JSON files: schema0.20.0/content0.1.92; original word encoding, archive metadata and explicit initial words.')
