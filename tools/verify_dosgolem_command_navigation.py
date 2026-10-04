"""接受正常指令窗導覽、取消及道具入口原版來源；入口 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_command_menu import validate as validate_open
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = 'issue4-command-navigation-r1'
PARENT = 'issue4-command-menu-r1'
PARENT_HASH = 'cdb0a4b69c83613681f6f3572665b1b1a8bc745fab5a8c959b23854c081af6fa'


def validate(root, producer):
    parent_path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(parent_path) == PARENT_HASH
    parent = json.loads(parent_path.read_text())
    assert parent == validate_open(root, Path(__file__).with_name('dosgolem_command_menu_probe.py'))
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'docker_image', 'build_flags'):
        assert meta[key] == parent['meta'][key]
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    assert '-load-state' not in meta['args']
    assert meta['producer_sha256'] == digest(producer) == '83f8ef19b873c9380d73ce3903f5bc98252bf7c87290247d1044d1a18dc16a37'
    assert all(meta[key] == value for key, value in RUNTIME_HASHES.items())
    assert meta['probe_source_sha256'] == digest(root / (PREFIX + '-probe-source.go')) == '822bc46f169c97f0e08306f5958aeb85857b3b168c04f4da91881a067e543337'
    assert meta['probe_sha256'] == digest(root / (PREFIX + '-probe')) == '4ae2e365313f53294977ac4d12eaf6d103bc9634cf5177bfb9d23186d365d93c'
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line or line.startswith('DQ3_SAVE_BOUND ') for line in lines)
    seeds = [fields(line) for line in lines if line.startswith('DQ3_CREATION_SEED ')]
    assert len(seeds) == 1 and seeds[0]['fixed'] == '1357'
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT', 'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(line) for line in lines if line.startswith(tag + ' ')] for key, tag in tags.items()}
    q, taken, states, irq = (groups[key] for key in tags)
    assert len(q) == len(taken) == len(states) == 206 and len(irq) == 488
    for key in groups:
        assert groups[key][:len(parent[key])] == parent[key]
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(row['scan'], 16) for row in q]
    assert [int(row['port60'], 16) for row in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(row['count']) for row in irq] == list(range(1, 489))
    expected = [('50', '2'), ('50', '3'), ('50', '4'), ('48', '3'), ('4b', '6'), ('4d', '3'), ('4d', '6'), ('01', '6'), ('39', '1'), ('50', '2'), ('4d', '5'), ('39', '1')]
    for n, (scan, cursor) in enumerate(expected, 195):
        assert q[n-1]['scan'] == scan and states[n-1]['choice_cursor'] == cursor
        assert states[n-1]['phase'] == ('ready' if n == 202 else 'choice')
        if n != 202:
            assert states[n-1]['choice_count'] == ('7' if n == 206 else '6')
    artifacts = []
    for n, (queued, consumed, state) in enumerate(zip(q, taken, states), 1):
        assert queued['packet'] == consumed['packet'] == state['packet'] == str(n)
        assert queued['scan'] == consumed['scan'] == state['scan'] and queued['kind'] == state['kind']
        assert int(queued['step']) == int(state['queued_step']) < int(consumed['step']) < int(state['step'])
        assert int(consumed['irqs']) == 76 + n * 2 - 1
        assert int(irq[76+n*2-1]['step']) <= int(state['step'])
        if n > 1:
            assert all(queued[key] == states[n-2][key] for key in ('step', 'phase', 'ida_linear'))
        if n >= 194:
            assert all(state[key] == states[192][key] for key in ('player_x', 'player_y', 'actor', 'flags', 'gold_lo', 'gold_hi', 'raw0b24'))
        label = PREFIX + f'-packet-{n:03d}-' + state['phase']
        image, indexed = root / (label+'.png'), root / (label+'.bin')
        width, height, indices, _ = png(image)
        assert (width, height) == (640, 350) and indices == indexed.read_bytes()
        for p in (image, indexed):
            if n <= 194:
                assert p.read_bytes() == (root / p.name.replace(PREFIX, PARENT, 1)).read_bytes()
            artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    clocks = [fields(line) for line in lines if line.startswith('DQ3_COMMAND_CLOCK ')]
    assert [row['packet'] for row in clocks] == list(map(str, range(193, 207)))
    for row in clocks:
        assert row['step'] == states[int(row['packet'])-1]['step'] and row['clock'] == '30' and row['raw526c'] == '1'
    persistent = (root / (PARENT+'-persistent-193.bin')).read_bytes()
    assert len(persistent) == 2172
    for n in range(193, 207):
        p = root / (PREFIX+f'-persistent-{n:03d}.bin')
        assert p.read_bytes() == persistent
        artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    exe = root.parents[1] / 'assets_raw/DQ3.EXE'
    assert exe.stat().st_size == meta['original_size'] and digest(exe) == meta['original_sha256']
    raw = exe.read_bytes()
    window = raw[0x19eac:0x19eac+64].hex()
    hud = bytearray(raw[0x19fdc:0x19fdc+24])
    hud[6:8] = (14).to_bytes(2, 'little')
    hud[12:14] = (1).to_bytes(2, 'little')
    rasters = [fields(line) for line in lines if line.startswith('DQ3_COMMAND_RASTER ')]
    assert [row['packet'] for row in rasters] == list(map(str, range(194, 207)))
    for row in rasters:
        assert row['step'] == states[int(row['packet'])-1]['step']
        assert row['window3d6c'] == window and row['hud3e9c'] == hud.hex()
        for key, value in {'raw258f': '8', 'raw0727': '5', 'raw2784': '255', 'raw5077': '1', 'raw071c': '3' if row['packet'] == '206' else '2'}.items():
            assert row[key] == value
    texts = [fields(line) for line in lines if line.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert texts == parent['text_events']
    ops_path = root / (PREFIX+'-fileops.json')
    ops = json.loads(ops_path.read_text())
    assert ops and not list((root.parent/(PREFIX+'-scratch')).iterdir())
    assert not any(op['Fn'] in (0x3c, 0x40, 0x41) or op.get('Name', '').lower().startswith('dragon') for op in ops if op['Step'] >= int(states[192]['step']))
    done = [fields(line) for line in lines if line.startswith('DQ3_SAVE_OBSERVED ')]
    assert len(done) == 1 and done[0]['packet'] == '206' and done[0]['step'] == states[-1]['step']
    return {'scope': '正常六指令導覽、取消、重開與道具入口；有限206包', 'meta': meta, **groups,
            'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'], 'upstream_revision': meta['upstream_revision'],
            'seed': meta['seed'], 'seed_control_once': True, 'parent_source_sha256': PARENT_HASH,
            'prefix194_unchanged': True, 'normal_inputs': 244, 'state_injection': False, 'emulator_snapshot_restore': False,
            'clocks': clocks, 'rasters': rasters, 'text_events': texts, 'fileops': ops, 'fileops_sha256': digest(ops_path),
            'artifacts': artifacts, 'done': done[0], 'checker_sha256': digest(Path(__file__)),
            'log_sha256': digest(root/(PREFIX+'.log')), 'remake_parity': False, 'full_rgb_parity': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', type=Path)
    parser.add_argument('producer', type=Path)
    parser.add_argument('receipt', type=Path)
    args = parser.parse_args()
    assert not args.receipt.exists()
    result = validate(args.root, args.producer)
    with args.receipt.open('x') as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2)
        stream.write('\n')
    print('正常指令導覽來源接受', digest(args.receipt))
