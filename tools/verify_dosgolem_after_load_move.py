"""獨立稽核正常 F6 後兩步行走；沿 docs/188 的固定200包父來源。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_save_load_flow import validate as validate_roundtrip

PREFIX = 'issue4-after-load-move-r1'
PARENT = 'issue4-save-load-roundtrip-r1'
PARENT_HASH = '0a88466eab206d7a4b346b8229bbcebd10683044aa363cb346bb81e173e3d238'
GO_HASH = '80dc4674639d14d390e8900c5775160500e675938054b0a43e6b69f872f3096a'
BINARY_HASH = 'b68bb2e0402b3e119e8d134149a4f236c517c2a994a109113034936e1c1b5e89'


def validate(root, producer):
    parent_path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(parent_path) == PARENT_HASH
    parent = json.loads(parent_path.read_text())
    assert parent == validate_roundtrip(root, 'roundtrip', Path(__file__).with_name('dosgolem_save_load_roundtrip_probe.py'))
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'docker_image', 'build_flags'):
        assert meta[key] == parent['meta'][key]
    assert meta['seed_configured_before_execution'] is True
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['producer_sha256'] == digest(producer)
    assert all(meta[key] == value for key, value in RUNTIME_HASHES.items())
    assert meta['probe_source_sha256'] == digest(root / (PREFIX + '-probe-source.go')) == GO_HASH
    assert meta['probe_sha256'] == digest(root / (PREFIX + '-probe')) == BINARY_HASH
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines and not any('找不到的檔（' in line or line.startswith('DQ3_SAVE_BOUND ') for line in lines)
    seed = [line for line in lines if line.startswith('DQ3_CREATION_SEED ')]
    assert len(seed) == 1 and 'fixed=1357' in seed[0]
    assert [fields(line) for line in lines if line.startswith('DQ3_SAVE_SCRATCH ')] == [{'step': '0', 'path': '/work/' + PREFIX + '-scratch'}]
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT', 'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(line) for line in lines if line.startswith(tag + ' ')] for key, tag in tags.items()}
    q, c, states, irq = (groups[key] for key in tags)
    assert len(q) == len(c) == len(states) == 202 and len(irq) == 480
    for key in groups:
        assert groups[key][:len(parent[key])] == parent[key]
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(row['scan'], 16) for row in q]
    assert [int(row['port60'], 16) for row in irq] == [value for scan in scans for value in (scan, scan | 128)]
    assert [int(row['count']) for row in irq] == list(range(1, 481))
    assert [row['scan'] for row in q[200:]] == ['4b', '4d']
    assert [row['kind'] for row in q[200:]] == ['after_load_left', 'after_load_right']
    artifacts = []
    for n, (queued, consumed, state) in enumerate(zip(q, c, states), 1):
        assert queued['packet'] == consumed['packet'] == state['packet'] == str(n)
        assert queued['scan'] == consumed['scan'] == state['scan']
        assert int(queued['step']) == int(state['queued_step']) < int(consumed['step']) < int(state['step'])
        assert int(consumed['irqs']) == 76 + n * 2 - 1
        assert int(irq[76 + n * 2 - 1]['step']) <= int(state['step'])
        if n > 1:
            before = states[n - 2]
            assert queued['step'] == before['step'] and queued['phase'] == before['phase'] and queued['ida_linear'] == before['ida_linear']
        if n > 200:
            assert state['phase'] == 'ready' and state['ida_linear'] == '1997c'
            assert int(state['raw0013'], 16) & 0x4000 == 0
            assert state['player_y'] == '18' and state['player_x'] == ('1' if n == 201 else '2')
            assert all(state[key] == states[199][key] for key in ('actor', 'flags', 'gold_lo', 'gold_hi', 'raw0b24'))
        label = PREFIX + f'-packet-{n:03d}-' + state['phase']
        image, indexed = root / (label + '.png'), root / (label + '.bin')
        width, height, indices, _ = png(image)
        assert (width, height) == (640, 350) and indices == indexed.read_bytes()
        for p in (image, indexed):
            if n <= 200:
                assert p.read_bytes() == (root / p.name.replace(PREFIX, PARENT, 1)).read_bytes()
            artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    clocks = [fields(line) for line in lines if line.startswith('DQ3_AFTER_LOAD_CLOCK ')]
    assert [row['packet'] for row in clocks] == [str(n) for n in range(193, 203)]
    for row in clocks:
        assert row['step'] == states[int(row['packet']) - 1]['step']
        assert 0 <= int(row['clock']) < 240
    assert all(row['clock'] == '0' and row['raw526c'] == '1' for row in clocks[7:])
    texts = [fields(line) for line in lines if line.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert texts == parent['text_events']
    native = [fields(line) for line in lines if line.startswith('DQ3_SAVE_LOAD_NATIVE ')]
    assert native == parent['native_events']
    operations_path = root / (PREFIX + '-fileops.json')
    operations = json.loads(operations_path.read_text())
    assert operations == parent['fileops']
    scratch = root.parent / (PREFIX + '-scratch')
    assert sorted(p.name for p in scratch.iterdir()) == ['dragon0.dat', 'player.dat']
    for p in scratch.iterdir():
        assert p.read_bytes() == (root.parent / (PARENT + '-scratch') / p.name).read_bytes()
        artifacts.append({'path': str(p.relative_to(root.parent)), 'size': p.stat().st_size, 'sha256': digest(p)})
    saved = (scratch / 'dragon0.dat').read_bytes()
    for n in range(193, 203):
        p = root / (PREFIX + f'-persistent-{n:03d}.bin')
        data = p.read_bytes()
        assert len(data) == 2172
        if n <= 200:
            assert data == (root / p.name.replace(PREFIX, PARENT, 1)).read_bytes()
        else:
            expected = bytearray(saved)
            if n == 201:
                expected[0x4f33 - 0x4f29] = 1
            assert data == expected
        artifacts.append({'path': p.name, 'size': len(data), 'sha256': digest(p)})
    done = [fields(line) for line in lines if line.startswith('DQ3_QUIESCENT_DONE ')]
    assert len(done) == 1 and done[0]['packets'] == '202' and done[0]['irqs'] == '480'
    assert done[0]['step'] == states[-1]['step']
    return {'scope': '正常冷啟動F5/F6後左移與右移；有限202包', 'meta': meta, **groups,
            'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'],
            'upstream_revision': meta['upstream_revision'], 'seed': meta['seed'],
            'clocks': clocks, 'text_events': texts, 'native_events': native, 'fileops': operations,
            'parent_source_sha256': PARENT_HASH, 'prefix200_unchanged': True,
            'normal_inputs': 240, 'seed_control_once': True, 'state_injection': False,
            'emulator_snapshot_restore': False, 'native_game_load': True,
            'artifacts': artifacts, 'done': done[0], 'checker_sha256': digest(Path(__file__)),
            'log_sha256': digest(root / (PREFIX + '.log')), 'remake_parity': False, 'full_rgb_parity': False}


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
    print('正常F6後行走來源接受', digest(args.receipt))
