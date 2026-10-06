"""相同422冷啟動的唯讀分支觀測：全部事件與產物須等於已接受來源。"""
from pathlib import Path
import ast, json, sys
sys.path.insert(0, '/repo/tools')
from verify_dosgolem_mother_return import digest, fields
from verify_dosgolem_first_move import RUNTIME_HASHES
root = Path('/work/dosgolem-opening')
pre, pp = 'issue4-room-npc-layer-normal-r1', 'issue4-field-room-door-normal-r1'
parent_path = root / (pp + '-source-r1-receipt.json')
assert digest(parent_path) == 'b7d0b1534a8b1055b57d78de8c343d12b736f355cf17584211b24dc2b84bcf3e'
parent = json.loads(parent_path.read_text())
meta = json.loads((root / (pre + '-meta.json')).read_text())
for k in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'build_flags', 'docker_image'):
    assert meta[k] == parent['meta'][k], k
assert all(meta[k] == v for k, v in RUNTIME_HASHES.items())
assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
assert digest(Path('/repo/tools/dosgolem_room_npc_layers_probe.py')) == meta['producer_sha256']
for suffix, key in [('-probe-source.go', 'probe_source_sha256'), ('-probe', 'probe_sha256')]: assert digest(root / (pre + suffix)) == meta[key]
assert meta['producer_sha256']=='e9d4d1551f4ad3e779e2e596e520d4900910dbf39ae3f31946552f54a8c270df'
assert meta['probe_source_sha256']=='d2d5d8a2a27695bd37da2c1e2ee2ab80d8a631aa20935322f2ade8eadd09ed82'
assert meta['probe_sha256']=='2dda379ccc9171cac8967c1f1e0827f14e102accab823641634f18a0678cb7cc'
exe = Path('/repo/assets_raw/DQ3.EXE')
assert exe.stat().st_size == meta['original_size'] and digest(exe) == meta['original_sha256']
tree = ast.parse(Path('/repo/tools/verify_dosgolem_field_equipment_wear.py').read_text())
tags = next(ast.literal_eval(n.value) for n in tree.body if isinstance(n, ast.Assign) and any(isinstance(t, ast.Name) and t.id == 'tags' for t in n.targets))
tags.update(wear_observations='DQ3_FIELD_EQUIP_WEAR_OBSERVE', native_events='DQ3_ARMOR_NATIVE', search_native_events='DQ3_SEARCH_NATIVE', talk_native_events='DQ3_TALK_NATIVE')
groups = {k: [] for k in tags}
layers, seeds, done = [], [], []
supported = False
with (root / (pre + '.log')).open() as f:
    for line in f:
        assert not line.startswith('DQ3_SAVE_BOUND ') and '找不到的檔（' not in line
        supported |= line.strip() == '沒實作的服務（0 種）：'
        for k, tag in tags.items():
            if line.startswith(tag + ' '): groups[k].append(fields(line)); break
        if line.startswith('DQ3_NPC_LAYER_NATIVE '): layers.append(fields(line))
        if line.startswith('DQ3_CREATION_SEED '): seeds.append(fields(line))
        if line.startswith('DQ3_SAVE_OBSERVED '): done.append(fields(line))
assert supported and len(seeds) == 1 and seeds[0]['fixed'] == '1357'
for k, rows in groups.items(): assert rows == parent[k], k
assert len(groups['states']) == 422 and len(groups['actual_irq1_events']) == 920
assert len(done) == 1 and done[0]['packet'] == '422' and done[0]['step'] == parent['states'][-1]['step']
artifacts = []
for a in parent['artifacts']:
    rel = a['path'].replace(pp, pre, 1)
    path = (root if '/' not in rel else root.parent) / rel
    assert path.stat().st_size == a['size'] and digest(path) == a['sha256'], rel
    artifacts.append(dict(path=rel, size=a['size'], sha256=a['sha256']))
assert len(artifacts) == 1076
ops_path = root / (pre + '-fileops.json')
assert json.loads(ops_path.read_text()) == json.loads((root / (pp + '-fileops.json')).read_text())
witnesses = []
expected = [(418,19,11,26,'11e19','80',0),(420,21,2,28,'11e19','80',0),(421,22,5,29,'11e19','80',0),(422,23,2,16,'11e25','00',0x80),(422,23,8,21,'11e25','00',0x80),(422,23,11,26,'11e33','00',0),(422,23,2,28,'11e33','00',0),(422,23,5,29,'11e33','00',0)]
assert len(layers)==len(expected)==8
for packet, py, x, y, pc, selector, cell_layer in expected:
    rows=[r for r in layers if r['packet']==str(packet) and r['player_x']=='4' and r['player_y']==str(py) and r['ida_linear']==pc and r['selector']==selector and r['tile_x']==str(x) and r['tile_y']==str(y)]
    assert len(rows)==1 and all((int(r['AX'],16)>>8)&0xe0==cell_layer|0x20 for r in rows),(packet,x,y)
    witnesses.append(dict(packet=packet,tile_x=x,tile_y=y,branch=pc,selector=selector,rows=len(rows),first=rows[0]))
report = dict(scope='normal422 unchanged; confirmed native selector -> NPC map-cell C0 mask -> display/skip branches; both room/field directions; animation clock not claimed', meta=meta, original422_source_sha256=digest(parent_path), full422_prefix_unchanged=True, artifacts=artifacts, native_layer_events=layers, witnesses=witnesses, normal_inputs=460, irq1_events=920, seed_control_once=True, state_injection=False, emulator_snapshot_restore=False, fileops_sha256=digest(ops_path), log_sha256=digest(root / (pre + '.log')), checker_sha256=digest(Path(__file__)), remake_parity=False, full_rgb_parity=False)
out = root / (pre + '-public-check-r1.json')
assert not out.exists()
out.write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
print('NATIVE_LAYER_SOURCE_PASS', digest(out), len(layers), [(r['packet'], r['tile_x'], r['tile_y'], r['branch'], r['rows']) for r in witnesses])
