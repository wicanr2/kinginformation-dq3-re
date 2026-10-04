"""接受本輪正常七列道具導覽與取消來源；入口 docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_command_navigation import validate as validate_parent
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = 'issue4-field-item-navigation-r2'
PARENT = 'issue4-command-navigation-r1'
PARENT_HASH = '5b2a78c152e77f1ddbfd850af0cc129837bf206cd9a15731529927d21b12978a'
PRODUCER_HASH = '60774b10b9986ac8cb842a161eba296429fa50b1a00cf9733b6769758e377d51'
SOURCE_HASH = 'df80f760b5cdb95d4b729fdce5a24f8afa104ae0b039f5d056ffb86eb5d3897a'
BINARY_HASH = '3f825a6918c9571ab4160a9478d1628e8d127aa469f3d1eb02b571023b0c5807'


def validate(root, producer):
    parent_path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(parent_path) == PARENT_HASH
    parent = json.loads(parent_path.read_text())
    assert parent == validate_parent(root, Path(__file__).with_name('dosgolem_command_menu_navigation_probe.py'))
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'docker_image', 'build_flags'):
        assert meta[key] == parent['meta'][key]
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    assert '-load-state' not in meta['args']
    assert meta['producer_sha256'] == digest(producer) == PRODUCER_HASH
    assert all(meta[key] == value for key, value in RUNTIME_HASHES.items())
    assert meta['probe_source_sha256'] == digest(root / (PREFIX + '-probe-source.go')) == SOURCE_HASH
    assert meta['probe_sha256'] == digest(root / (PREFIX + '-probe')) == BINARY_HASH
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line or line.startswith('DQ3_SAVE_BOUND ') for line in lines)
    seeds = [fields(line) for line in lines if line.startswith('DQ3_CREATION_SEED ')]
    assert len(seeds) == 1 and seeds[0]['fixed'] == '1357'
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT', 'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(line) for line in lines if line.startswith(tag + ' ')] for key, tag in tags.items()}
    q, taken, states, irq = (groups[key] for key in tags)
    assert len(q) == len(taken) == len(states) == 218 and len(irq) == 512
    for key in groups:
        assert groups[key][:len(parent[key])] == parent[key]
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(row['scan'], 16) for row in q]
    assert [int(row['port60'], 16) for row in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(row['count']) for row in irq] == list(range(1, 513))
    expected = [(207, '39', 'choice', 3, 1), (208, '01', 'ready', 6, 5),
                (209, '39', 'choice', 6, 1), (210, '50', 'choice', 6, 2),
                (211, '4d', 'choice', 6, 5), (212, '39', 'choice', 7, 1),
                (213, '50', 'choice', 7, 2), (214, '48', 'choice', 7, 1),
                (215, '48', 'choice', 7, 7), (216, '50', 'choice', 7, 1),
                (217, '01', 'ready', 6, 5), (218, '4d', 'ready', 6, 5)]
    for n, scan, phase, count, cursor in expected:
        assert q[n-1]['scan'] == scan
        assert states[n-1]['phase'] == phase
        assert states[n-1]['choice_count'] == str(count) and states[n-1]['choice_cursor'] == str(cursor)
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
            assert all(state[key] == states[192][key] for key in ('actor', 'flags', 'gold_lo', 'gold_hi', 'raw0b24'))
            assert state['player_y'] == '18' and state['player_x'] == ('3' if n == 218 else '2')
        label = PREFIX + f'-packet-{n:03d}-' + state['phase']
        image, indexed = root / (label+'.png'), root / (label+'.bin')
        width, height, indices, _ = png(image)
        assert (width, height) == (640, 350) and indices == indexed.read_bytes()
        for p in (image, indexed):
            if n <= 206:
                assert p.read_bytes() == (root / p.name.replace(PREFIX, PARENT, 1)).read_bytes()
            artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    clocks = [fields(line) for line in lines if line.startswith('DQ3_COMMAND_CLOCK ')]
    assert [row['packet'] for row in clocks] == list(map(str, range(193, 219)))
    for row in clocks:
        assert row['step'] == states[int(row['packet'])-1]['step'] and row['clock'] == '30' and row['raw526c'] == '1'
    baseline = (root / (PARENT+'-persistent-193.bin')).read_bytes()
    assert len(baseline) == 2172
    moved = bytearray(baseline)
    moved[0x4f33-0x4f29:0x4f35-0x4f29] = (3).to_bytes(2, 'little')
    for n in range(193, 219):
        p = root / (PREFIX+f'-persistent-{n:03d}.bin')
        assert p.read_bytes() == (bytes(moved) if n == 218 else baseline)
        artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    rasters = [fields(line) for line in lines if line.startswith('DQ3_COMMAND_RASTER ')]
    assert [row['packet'] for row in rasters] == list(map(str, range(194, 219)))
    assert rasters[:len(parent['rasters'])] == parent['rasters']
    items = [fields(line) for line in lines if line.startswith('DQ3_ITEM_RASTER ')]
    assert [row['packet'] for row in items] == list(map(str, range(206, 219)))
    exe = root.parents[1] / 'assets_raw/DQ3.EXE'
    assert exe.stat().st_size == meta['original_size'] and digest(exe) == meta['original_sha256']
    raw = exe.read_bytes()
    list_window = bytearray(raw[0x1a118:0x1a118+96])
    list_window[8:10] = (144).to_bytes(2, 'little')
    list_window[12:14] = (7).to_bytes(2, 'little')
    list_window[20:22] = (7).to_bytes(2, 'little')
    action_window = raw[0x1a190:0x1a190+64]
    for row in items:
        assert row['step'] == states[int(row['packet'])-1]['step']
        assert row['owner062d'] == '1' and row['raw2591'] == '0000'
        expected_window = bytearray(list_window)
        if int(row['packet']) >= 207 and int(row['packet']) <= 211:
            expected_window[20:22] = (0).to_bytes(2, 'little')
        assert row['window3fd8'] == expected_window.hex()
        assert row['action4050'] == action_window.hex()
    texts = [fields(line) for line in lines if line.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert texts == parent['text_events']
    ops_path = root / (PREFIX+'-fileops.json')
    ops = json.loads(ops_path.read_text())
    assert ops and not list((root.parent/(PREFIX+'-scratch')).iterdir())
    assert not any(op['Fn'] in (0x3c, 0x40, 0x41) or op.get('Name', '').lower().startswith('dragon') for op in ops if op['Step'] >= int(states[192]['step']))
    done = [fields(line) for line in lines if line.startswith('DQ3_SAVE_OBSERVED ')]
    assert len(done) == 1 and done[0]['packet'] == '218' and done[0]['step'] == states[-1]['step']
    return {'scope': '正常單人七列、穿戴列開action後Esc返回、導覽繞回、清單Esc及下一步；有限218包',
            'meta': meta, **groups, 'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'],
            'upstream_revision': meta['upstream_revision'], 'seed': meta['seed'], 'seed_control_once': True,
            'parent_source_sha256': PARENT_HASH, 'prefix206_unchanged': True, 'normal_inputs': 256,
            'state_injection': False, 'emulator_snapshot_restore': False, 'clocks': clocks, 'rasters': rasters,
            'item_rasters': items, 'text_events': texts, 'fileops': ops, 'fileops_sha256': digest(ops_path),
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
    print('正常七列道具導覽來源接受', digest(args.receipt))
