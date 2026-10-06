"""Docker-only independent checker; docs/188. Explicit injected component only."""
from pathlib import Path
import hashlib, json, os

root = Path('/work/dosgolem-opening')
pre = 'issue4-npc-turn-layers-component-r1'
def digest(p):
    return hashlib.sha256(p.read_bytes()).hexdigest()
meta_path = root / (pre + '-meta.json')
log_path = root / (pre + '.log')
meta = json.loads(meta_path.read_text())
pins = {
    'producer_sha256': '428a968bca5d49cb4a50c200585d636b2f3514ce4c01578c7bf67b14c6e83e85',
    'probe_source_sha256': '066964d6da584229f00017a466a1f080486251d88786991c7ace11b13539058f',
    'probe_sha256': 'e740410cd63cf1e952e15ba3b63ab469913c79f493f8908e3e8cb7b071cbc809',
    'frozen_parent_go_sha256': 'f4ec495afa57d001d181c165bc9141643dc246fd781eca8a6e646bd494186321',
    'generator_sha256': '778f39eefc1c8070caf6e5a341186d131bffb18c41b892c7b420f80a5e6ed6dd',
}
actual = {
    'producer_sha256': Path('/repo/tools/dosgolem_npc_turn_layers_component_probe.py'),
    'probe_source_sha256': root / (pre + '-probe-source.go'),
    'probe_sha256': root / (pre + '-probe'),
    'frozen_parent_go_sha256': root / 'issue4-npc-move-state-normal-r1-probe-source.go',
    'generator_sha256': Path('/repo/tools/dosgolem_newgame_probe.py'),
}
for key, expected in pins.items():
    assert meta[key] == digest(actual[key]) == expected, key
native_meta = json.loads((root / 'issue4-npc-move-state-normal-r1-meta.json').read_text())
for key in native_meta:
    if key.startswith(('patched_', 'upstream_')):
        assert meta[key] == native_meta[key], key
exe = Path('/repo/assets_raw/DQ3.EXE')
assert exe.stat().st_size == meta['original_size'] == 115282
assert digest(exe) == meta['original_sha256'] == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert meta['upstream_revision'] == '2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
assert meta['warm_seed'] == '1357' and meta['component_seed'] == '1e2c'
assert meta['component_seed_fixed_before_each_call'] is True
assert meta['layers'] == [0, 1, 2, 3]
assert meta['state_restore'] is True and meta['cpu_register_reentry'] is True
assert meta['gameplay_state_injection'] is True
assert meta['emulator_snapshot_restore'] is False and meta['normal_player_path'] is False
def parse(line):
    return dict(v.split('=', 1) for v in line.strip().split()[1:])
rows = []
with log_path.open() as stream:
    for line in stream:
        if line.startswith('DQ3_TURN_COMPONENT_'):
            rows.append((line.split()[0], parse(line)))
assert len(rows) == 30
assert rows[0][0] == 'DQ3_TURN_COMPONENT_WARM' and rows[-1][0] == 'DQ3_TURN_COMPONENT_DONE'
warm = rows[0][1]
assert warm['step'] == '2282686367' and warm['packet'] == '424'
assert warm['seed'] == '1e2c' and warm['slot'] == '6' and warm['count'] == '14'
assert warm['original_word'] == '2650' and len(bytes.fromhex(warm['npc_slots'])) == 112
parent_path = root / 'issue4-npc-move-state-normal-r1-component-r1.json'
assert digest(parent_path) == '6586abb198648dc27859901813f2c40cad156579b511c101db820eb7023b69b4'
parent = json.loads(parent_path.read_text())
donors = [c for c in parent['cases'] if c['entry']['step'] == warm['step']]
assert len(donors) == 1
donor = donors[0]
assert donor['entry']['npc_slots'] == warm['npc_slots']
cases = []
for layer in range(4):
    group = rows[1 + layer * 7:1 + (layer + 1) * 7]
    assert group[0][0] == 'DQ3_TURN_COMPONENT_ENTRY'
    entry = group[0][1]
    word = (0x2650 & 0x3fff) | layer << 14
    assert entry['case'] == entry['layer'] == str(layer)
    assert int(entry['word'], 16) == word and entry['seed'] == '1e2c'
    assert entry['ctrl'] == '05' and entry['x'] == '5' and entry['y'] == '29'
    assert entry['player_x'] == '5' and entry['player_y'] == '23'
    assert entry['state_injection'] == 'true'
    assert entry['cpu_reentry'] == ('false' if layer == 0 else 'true')
    events = [v for tag, v in group[1:]]
    assert all(tag == 'DQ3_TURN_COMPONENT_NATIVE' for tag, v in group[1:])
    assert [v['ida_linear'] for v in events] == ['12043', '12065', '12074', '1207f', '12098', '120a6']
    assert [v['seed'] for v in events] == ['7225', '11e8', '1005', '1005', '1005', '1005']
    assert [v['DX'] for v in events[:3]] == ['0001', '0000', '0001']
    assert all(int(v['DI'], 16) == word for v in events[:5])
    signed_word = word if word < 0x8000 else word - 0x10000
    assert events[2]['AX'] == events[3]['AX'] == '00cd'
    assert events[3]['CX'] == '0007'
    expected_ctrl = 0x04 if signed_word >= 7 else 0x06
    assert int(events[4]['AX'], 16) == expected_ctrl
    assert events[-1]['ctrl'] == f'{expected_ctrl:02x}'
    assert all(v['x'] == '5' and v['y'] == '29' for v in events)
    assert all(v['layer'] == v['case'] == str(layer) for v in events)
    assert all(int(a['step']) < int(b['step']) for a, b in zip([entry, *events], events))
    if layer == 0:
        assert events[-1]['seed'] == donor['return']['seed']
        assert int(events[-1]['step']) == int(donor['return']['step'])
        assert bytes.fromhex(donor['return']['npc_slots'])[6 * 8 + 3] == expected_ctrl
    else:
        assert entry['step'] == cases[-1]['return_state']['step']
    cases.append(dict(entry=entry, events=events, return_state=events[-1]))
assert rows[-1][1]['cases'] == '4' and rows[-1][1]['no_normal_path_claim'] == 'true'
assert int(rows[-1][1]['step']) - int(warm['step']) < 50000
fixture = dict(normal_player_path=False, state_injection=True, cpu_register_reentry=True,
               emulator_snapshot_restore=False, warm_source_sha256=parent['source_sha256'],
               component_seed='1e2c', npc_slots=warm['npc_slots'], slot=6, cases=cases)
fixture_path = root / (pre + '-fixture-r1.json')
receipt_path = root / (pre + '-receipt-r1.json')
assert not fixture_path.exists() and not receipt_path.exists()
assert root.stat().st_uid == os.getuid() == 1000
fixture_path.write_text(json.dumps(fixture, ensure_ascii=False, indent=2) + '\n')
receipt = dict(scope='original four-layer automatic-turn component only', cases=4,
               normal_player_path=False, state_injection=True, cpu_register_reentry=True,
               emulator_snapshot_restore=False, seed='1e2c', seed_fixed_before_each_call=True,
               original_sha256=digest(exe), original_size=exe.stat().st_size,
               meta_sha256=digest(meta_path), log_sha256=digest(log_path),
               checker_sha256=digest(Path(__file__)), fixture_sha256=digest(fixture_path),
               tools=pins, result='PASS', ctrl_by_layer=['04','04','06','06'],
               coordinate_changes=0, step_roll_observations=0)
receipt_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + '\n')
assert fixture_path.stat().st_uid == receipt_path.stat().st_uid == 1000
print(json.dumps(dict(receipt_sha256=digest(receipt_path), fixture_sha256=digest(fixture_path), result='PASS')))
