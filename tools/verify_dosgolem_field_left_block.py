"""Issue4 DRAFT來源稽核；不预設轉場、遭遇或remake parity。"""
from pathlib import Path
import ast, json, struct, sys
sys.path.insert(0, '/repo/tools')
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_first_move import RUNTIME_HASHES

root = Path('/work/dosgolem-opening')
pre = 'issue4-field-left-transition-normal-r2'
pp = 'issue4-command-enter-first-normal-r1'
parent_path = root / (pp + '-source-r1-receipt.json')
assert digest(parent_path) == '62e403190b4787e33b96d28ab81891f249ac0815900859eb706a411722f299e6'
parent = json.loads(parent_path.read_text())
meta = json.loads((root / (pre + '-meta.json')).read_text())
for k in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'build_flags', 'docker_image'):
    assert meta[k] == parent['meta'][k], k
assert all(meta[k] == v for k, v in RUNTIME_HASHES.items())
assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
assert digest(Path('/repo/tools/dosgolem_field_left_block_probe.py')) == meta['producer_sha256']
for suffix, key in [('-probe-source.go', 'probe_source_sha256'), ('-probe', 'probe_sha256')]:
    assert digest(root / (pre + suffix)) == meta[key]
assert meta['producer_sha256']=='c1564d7c58502a5810a14986c25a22b857e082cb20fe88223b75ffdd93a2088d'
assert meta['probe_source_sha256']=='39f249e16266c8831f9bcfea53b7d3c3b8c66dab6b66ce69098ccc5ea4fcc8bb'
assert meta['probe_sha256']=='b6e90badfb60e5885a7a6f398b9a39ee254b21386827aa6106af6203e71b7249'
exe = Path('/repo/assets_raw/DQ3.EXE')
assert exe.stat().st_size == meta['original_size'] and digest(exe) == meta['original_sha256']
tree = ast.parse(Path('/repo/tools/verify_dosgolem_field_equipment_wear.py').read_text())
tags = next(ast.literal_eval(n.value) for n in tree.body if isinstance(n, ast.Assign) and any(isinstance(t, ast.Name) and t.id == 'tags' for t in n.targets))
tags.update(wear_observations='DQ3_FIELD_EQUIP_WEAR_OBSERVE', native_events='DQ3_ARMOR_NATIVE', search_native_events='DQ3_SEARCH_NATIVE')
groups = {k: [] for k in tags}
talk, seed, done = [], [], []
supported = False
with (root / (pre + '.log')).open() as f:
    for line in f:
        assert not line.startswith('DQ3_SAVE_BOUND ') and '找不到的檔（' not in line
        supported |= line.strip() == '沒實作的服務（0 種）：'
        for k, tag in tags.items():
            if line.startswith(tag + ' '):
                groups[k].append(fields(line))
                break
        if line.startswith('DQ3_TALK_NATIVE '): talk.append(fields(line))
        if line.startswith('DQ3_CREATION_SEED '): seed.append(fields(line))
        if line.startswith('DQ3_SAVE_OBSERVED '): done.append(fields(line))
assert supported and len(seed) == 1 and seed[0]['fixed'] == '1357'
for k, rows in groups.items(): assert rows[:len(parent[k])] == parent[k], k
q, c, states, irq = (groups[k] for k in ('queued', 'consumed', 'states', 'actual_irq1_events'))
count = len(states)
assert count == 414 and len(q) == len(c) == count and len(irq) == 76 + count * 2
scans = [v for _, v in meta['normal_prefix_inputs']] + [int(s['scan'], 16) for s in q]
assert len(scans) == 38 + count and [int(s['port60'], 16) for s in irq] == [v for scan in scans for v in (scan, scan | 128)]
assert [int(s['count']) for s in irq] == list(range(1, len(irq) + 1))
assert [s['scan'] for s in q[410:]] == ['4b'] * (count - 410)
artifacts = []
for a in parent['artifacts']:
    rel = a['path'].replace(pp, pre, 1)
    path = (root if '/' not in rel else root.parent) / rel
    assert path.stat().st_size == a['size'] and digest(path) == a['sha256'], rel
    artifacts.append(dict(path=rel, size=a['size'], sha256=a['sha256']))
base = (root / (pp + '-persistent-410.bin')).read_bytes()
observed = []
for n, (queued, taken, state) in enumerate(zip(q, c, states), 1):
    assert queued['packet'] == taken['packet'] == state['packet'] == str(n)
    assert queued['scan'] == taken['scan'] == state['scan'] and queued['kind'] == state['kind']
    assert int(queued['step']) == int(state['queued_step']) < int(taken['step']) < int(state['step'])
    assert int(taken['irqs']) == 76 + n * 2 - 1 and int(irq[76 + n * 2 - 1]['step']) <= int(state['step'])
    if n > 1: assert all(queued[k] == states[n-2][k] for k in ('step', 'phase', 'ida_linear'))
    if n <= 410: continue
    path = root / (pre + f'-persistent-{n:03d}.bin')
    data = path.read_bytes()
    assert len(data) == len(base) == 2172
    expected = bytearray(base)
    struct.pack_into('<HH',expected,10,int(state['player_x']),int(state['player_y']))
    assert data == expected, n
    assert bytes.fromhex(state['actor']) == data[342:470]
    assert struct.unpack_from('<HH', data, 10) == (int(state['player_x']), int(state['player_y']))
    label = pre + f'-packet-{n:03d}-' + state['phase']
    im, idx = root / (label + '.png'), root / (label + '.bin')
    width, height, indices, _ = png(im)
    assert (width, height) == (640, 350) and indices == idx.read_bytes()
    observed.append(dict(packet=n, phase=state['phase'], ida_linear=state['ida_linear'], player_x=state['player_x'], player_y=state['player_y'], changed_persistent_offsets=[i for i, (a, b) in enumerate(zip(base, data)) if a != b]))
    for path in (im, idx, path): artifacts.append(dict(path=path.name, size=path.stat().st_size, sha256=digest(path)))
assert len(artifacts) == 1040 + 3 * (count - 410)
clocks = [s for s in groups['clocks'] if int(s['packet']) > 410]
assert [int(s['packet']) for s in clocks] == list(range(411, count + 1))
assert all(s['step'] == states[int(s['packet'])-1]['step'] for s in clocks)
ops_path = root / (pre + '-fileops.json')
ops = json.loads(ops_path.read_text())
old = json.loads((root / (pp + '-fileops.json')).read_text())
assert ops[:len(old)] == old
assert not any(o['Fn'] in (0x3c, 0x40, 0x41) for o in ops if o['Step'] >= int(states[409]['step']))
assert len(done) == 1 and done[0]['packet'] == str(count) and done[0]['step'] == states[-1]['step']
assert talk[:len(parent['talk_native_events'])] == parent['talk_native_events']
report = dict(talk_native_events=talk, scope='normal410 -> up to four Left keys; stop first non-town/non-field stable result; source only', meta=meta, **groups, artifacts=artifacts, parent_source_sha256=digest(parent_path), prefix410_unchanged=True, normal_inputs=len(scans), seed_control_once=True, state_injection=False, emulator_snapshot_restore=False, observed_results=observed, original_saved_files_unchanged=True, native_clock_limit=parent['native_clock_limit'], fileops_sha256=digest(ops_path), log_sha256=digest(root / (pre + '.log')), checker_sha256=digest(Path(__file__)), remake_parity=False, full_rgb_parity=False)
out = root / (pre + '-public-check-r1.json')
assert not out.exists()
out.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
print('SOURCE_ONLY_PASS', len(artifacts), digest(out), observed)
