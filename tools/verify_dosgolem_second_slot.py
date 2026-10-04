"""獨立稽核正常第二槽 F5/F6 及行走；有限來源入口 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_save_load import validate as validate_parent

PREFIX = 'issue4-second-slot-r1'
PARENT = 'issue4-save-f5-normal-r1'
PARENT_HASH = '3350c30fa5c48fc690b144a6860cac155a4e02efb498649b7120a9bbd5f0973d'
PRODUCER_HASH = '527189545249de4a41d97807a22bee609909feb520f6c666a1c613296b518b8a'
GO_HASH = 'fd44bee0ca105634932a33259d762ec75cdcdd7be1c135961176fbc0b838dbbc'
BINARY_HASH = 'dafce59163f33c2838a60cc82a63edb8db10131a4d26df18745d1fa55207aed4'


def validate(root, producer, assets):
    parent_path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(parent_path) == PARENT_HASH
    parent = json.loads(parent_path.read_text())
    assert parent == validate_parent(root, Path(__file__).with_name('dosgolem_save_load_probe.py'))
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'docker_image', 'build_flags'):
        assert meta[key] == parent['meta'][key]
    assert meta['seed_configured_before_execution'] is True
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['producer_sha256'] == digest(producer) == PRODUCER_HASH
    assert all(meta[key] == value for key, value in RUNTIME_HASHES.items())
    assert meta['probe_source_sha256'] == digest(root / (PREFIX + '-probe-source.go')) == GO_HASH
    assert meta['probe_sha256'] == digest(root / (PREFIX + '-probe')) == BINARY_HASH
    assert (assets / 'DQ3.EXE').stat().st_size == meta['original_size'] and digest(assets / 'DQ3.EXE') == meta['original_sha256']
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line or line.startswith('DQ3_SAVE_BOUND ') for line in lines)
    seed = [line for line in lines if line.startswith('DQ3_CREATION_SEED ')]
    assert len(seed) == 1 and 'fixed=1357' in seed[0]
    assert [fields(line) for line in lines if line.startswith('DQ3_SAVE_SCRATCH ')] == [{'step': '0', 'path': '/work/' + PREFIX + '-scratch'}]
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT', 'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(line) for line in lines if line.startswith(tag + ' ')] for key, tag in tags.items()}
    queued, consumed, states, irq = (groups[key] for key in tags)
    assert len(queued) == len(consumed) == len(states) == 204 and len(irq) == 484
    for key in groups:
        assert groups[key][:len(parent[key])] == parent[key]
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(row['scan'], 16) for row in queued]
    assert [int(row['port60'], 16) for row in irq] == [value for scan in scans for value in (scan, scan | 128)]
    assert [int(row['count']) for row in irq] == list(range(1, 485))
    assert [row['scan'] for row in queued[193:]] == ['3f', '1c', '50', '1c', '1c', '4b', '40', '50', '1c', '4b', '4d']
    assert [row['kind'] for row in queued[194:]] == ['save_accept', 'save_second_slot_cursor', 'save_second_slot', 'save_close', 'after_save_move', 'load_f6', 'load_second_slot_cursor', 'load_second_slot', 'after_load_left', 'after_load_right']
    artifacts = []
    phases = {'ready': '1997c', 'inline_wait': '216d8', 'waiting': '21133', 'choice': '1f7b7', 'name': '11096'}
    for n, (q, taken, state) in enumerate(zip(queued, consumed, states), 1):
        assert q['packet'] == taken['packet'] == state['packet'] == str(n)
        assert q['scan'] == taken['scan'] == state['scan'] and q['kind'] == state['kind']
        assert int(q['step']) == int(state['queued_step']) < int(taken['step']) < int(state['step'])
        assert int(taken['irqs']) == 76 + n * 2 - 1 and int(irq[76 + n * 2 - 1]['step']) <= int(state['step'])
        assert state['ida_linear'] == phases[state['phase']] and int(state['raw0013'], 16) & 0x4000 == 0
        if n > 1:
            before = states[n - 2]
            assert all(q[key] == before[key] for key in ('step', 'phase', 'ida_linear'))
        if n >= 194:
            assert all(state[key] == states[192][key] for key in ('actor', 'flags', 'gold_lo', 'gold_hi', 'raw0b24'))
            assert state['player_y'] == '18' and state['player_x'] == ('1' if n in (199, 200, 201, 203) else '2')
        label = PREFIX + f'-packet-{n:03d}-' + state['phase']
        image, indexed = root / (label + '.png'), root / (label + '.bin')
        width, height, indices, _ = png(image)
        assert (width, height) == (640, 350) and indices == indexed.read_bytes()
        for path in (image, indexed):
            if n <= 194:
                assert path.read_bytes() == (root / path.name.replace(PREFIX, PARENT, 1)).read_bytes()
            artifacts.append({'path': path.name, 'size': path.stat().st_size, 'sha256': digest(path)})
    assert [s['phase'] for s in states[193:]] == ['choice', 'choice', 'choice', 'waiting', 'ready', 'ready', 'choice', 'choice', 'ready', 'ready', 'ready']
    assert [s['last_record'] for s in states[193:]] == ['253', '250', '250'] + ['252'] * 8
    assert all(states[n - 1]['choice_count'] == '10' for n in (195, 196, 200, 201))
    assert all(states[n - 1]['choice_cursor'] == str(cursor) for n, cursor in ((195, 1), (196, 2), (200, 1), (201, 2)))
    texts = [fields(line) for line in lines if line.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert texts[:len(parent['text_events'])] == parent['text_events']
    assert [(row['packet'], row['DI']) for row in texts[len(parent['text_events']):]] == [('195', '00fa'), ('197', '00fc')]
    assert all(row['text_segment'] == '2826' for row in texts)
    clocks = [fields(line) for line in lines if line.startswith('DQ3_AFTER_LOAD_CLOCK ')]
    assert [row['packet'] for row in clocks] == [str(n) for n in range(193, 205)]
    for row in clocks:
        n = int(row['packet'])
        assert row['step'] == states[n - 1]['step'] and row['clock'] == ('0' if n >= 202 else '30') and row['raw526c'] == '1'
    scratch = root.parent / (PREFIX + '-scratch')
    assert sorted(p.name for p in scratch.iterdir()) == ['dragon1.dat', 'player.dat']
    saved, directory = (scratch / 'dragon1.dat').read_bytes(), (scratch / 'player.dat').read_bytes()
    assert len(saved) == 2172 and len(directory) == 200
    assert digest(scratch / 'dragon1.dat') == '9ef549b0ac969e8bb9a0138d4e339905c21f5abf8653ca718c09d4b4c5db5910'
    original_directory = (assets / 'PLAYER.DAT').read_bytes()
    assert digest(assets / 'PLAYER.DAT') == 'a445a11f52a6711aba1433d9107d10284253e08d27aa6be2b82dc7aec91376dc'
    assert directory[:20] == original_directory[:20] and directory[40:] == original_directory[40:]
    actor = bytes.fromhex(states[196]['actor'])
    assert directory[20:38] == actor[3:21] and directory[38] == actor[21] and directory[39] == actor[2]
    initial = (root / (PREFIX + '-persistent-193.bin')).read_bytes()
    for n in range(193, 205):
        path = root / (PREFIX + f'-persistent-{n:03d}.bin')
        expected = bytearray(initial if n <= 196 else saved)
        if n in (199, 200, 201, 203):
            expected[0x4f33 - 0x4f29] = 1
        assert path.read_bytes() == expected
        artifacts.append({'path': path.name, 'size': path.stat().st_size, 'sha256': digest(path)})
    for path in sorted(scratch.iterdir()):
        artifacts.append({'path': str(path.relative_to(root.parent)), 'size': path.stat().st_size, 'sha256': digest(path)})
    operations_path = root / (PREFIX + '-fileops.json')
    operations = json.loads(operations_path.read_text())
    created = [op for op in operations if op['Op'] == 'create']
    assert [op['Name'] for op in created] == ['player.dat', 'dragon1.dat'] and not any(op['Failed'] for op in created)
    reads = [op for op in operations if op['Op'] == 'read' and op['Name'] == 'dragon1.dat' and op['Step'] > int(states[200]['step'])]
    assert len(reads) == 1 and reads[0]['Len'] == reads[0]['Arg'] == 2172 and not reads[0]['Failed']
    assert not any(op['Name'] == 'dragon0.dat' for op in operations if op['Step'] >= int(states[192]['step']))
    native = [fields(line) for line in lines if line.startswith('DQ3_SAVE_LOAD_NATIVE ')]
    assert [row['ida_linear'] for row in native] == ['11484', '114c8', '114d3', '114d9', '1157d', '1158b', '11591', '1165f']
    assert all(row['DS'] == '15ed' and row['raw0722'] == '2' and row['raw0726'] == '0' for row in native)
    assert native[2]['AX'] == '400b' and native[2]['CX'] == '087c' and native[2]['DX'] == '4f29'
    assert native[5]['AX'] == '3f0b' and native[5]['CX'] == '087c' and native[5]['DX'] == '4f29'
    retained = 0
    for artifact in json.loads((root / 'issue4-mother-finish-receipt.json').read_text())['artifacts']:
        if artifact['path'].endswith(('.png', '.bin')):
            path = root / artifact['path'].replace('issue4-mother-finish-', PREFIX + '-', 1)
            assert path.stat().st_size == artifact['size'] and digest(path) == artifact['sha256']
            retained += 1
    assert retained == 174
    done = [fields(line) for line in lines if line.startswith('DQ3_QUIESCENT_DONE ')]
    assert len(done) == 1 and done[0]['packets'] == '204' and done[0]['irqs'] == '484' and done[0]['step'] == states[-1]['step']
    return {'scope': '正常F5/F6第二槽保存讀回及左右行走；有限204包', 'meta': meta, **groups,
            'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'],
            'upstream_revision': meta['upstream_revision'], 'seed': meta['seed'],
            'parent_source_sha256': PARENT_HASH, 'prefix194_unchanged': True, 'parent174_unchanged': True,
            'seed_control_once': True, 'state_injection': False, 'emulator_snapshot_restore': False,
            'normal_inputs': 242, 'clocks': clocks, 'text_events': texts, 'native_events': native,
            'fileops': operations, 'fileops_sha256': digest(operations_path),
            'fileops_limit': 'AH40由原生caller、落地2172bytes及完整讀回共同核對，FileOps不是單独writer證據',
            'artifacts': artifacts, 'done': done[0], 'checker_sha256': digest(Path(__file__)),
            'log_sha256': digest(root / (PREFIX + '.log')), 'remake_parity': False, 'full_rgb_parity': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', type=Path)
    parser.add_argument('producer', type=Path)
    parser.add_argument('receipt', type=Path)
    parser.add_argument('--assets', type=Path, required=True)
    args = parser.parse_args()
    assert not args.receipt.exists()
    result = validate(args.root, args.producer, args.assets)
    with args.receipt.open('x') as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2)
        stream.write('\n')
    print('正常第二槽來源接受', digest(args.receipt))
