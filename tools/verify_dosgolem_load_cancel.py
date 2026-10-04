"""獨立稽核正常 F6 選槽 Esc 取消與行走；有限來源入口 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_recruitment_menu_cancel import validate as validate_parent

PREFIX = 'issue4-load-cancel-r1'
PARENT = 'issue4-recruit-menu-esc-normal-r1'
PARENT_HASH = '137411c711e81684943b4cc8aac2d952b8c47c8981f70a79f0d308e369204c4c'
PRODUCER_HASH = 'f76a4df7704e1cfdfbb62a3c78752703098a68aec2a37fbf8b2f9e335b205767'
GO_HASH = 'c997de82e459c30a4fb59a64c62e7625bf626d70df8975363d8af7677c77baf7'
BINARY_HASH = 'fb7da8e2f6b985c7209522a623494ccba36c2ac3c0ea045e6227551efaa31771'


def validate(root, producer):
    path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(path) == PARENT_HASH
    parent = json.loads(path.read_text())
    assert parent == validate_parent(root, Path(__file__).with_name('dosgolem_recruitment_menu_cancel_probe.py'))
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'docker_image', 'build_flags'):
        assert meta[key] == parent['meta'][key]
    assert meta['seed_configured_before_execution'] is True
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['producer_sha256'] == digest(producer) == PRODUCER_HASH
    assert all(meta[key] == value for key, value in RUNTIME_HASHES.items())
    assert meta['probe_source_sha256'] == digest(root / (PREFIX + '-probe-source.go')) == GO_HASH
    assert meta['probe_sha256'] == digest(root / (PREFIX + '-probe')) == BINARY_HASH
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line or line.startswith('DQ3_SAVE_BOUND ') for line in lines)
    seed = [line for line in lines if line.startswith('DQ3_CREATION_SEED ')]
    assert len(seed) == 1 and 'fixed=1357' in seed[0]
    assert [fields(line) for line in lines if line.startswith('DQ3_SAVE_SCRATCH ')] == [{'step': '0', 'path': '/work/' + PREFIX + '-scratch'}]
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT', 'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(line) for line in lines if line.startswith(tag + ' ')] for key, tag in tags.items()}
    q, consumed, states, irq = (groups[key] for key in tags)
    assert len(q) == len(consumed) == len(states) == 197 and len(irq) == 470
    for key in groups:
        assert groups[key][:len(parent[key])] == parent[key]
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(row['scan'], 16) for row in q]
    assert [int(row['port60'], 16) for row in irq] == [value for scan in scans for value in (scan, scan | 128)]
    assert [int(row['count']) for row in irq] == list(range(1, 471))
    assert [row['scan'] for row in q[193:]] == ['40', '01', '4b', '4d']
    assert [row['kind'] for row in q[193:]] == ['load_f6_cancel_entry', 'load_slot_escape', 'after_load_cancel_left', 'after_load_cancel_right']
    artifacts = []
    for n, (queued, taken, state) in enumerate(zip(q, consumed, states), 1):
        assert queued['packet'] == taken['packet'] == state['packet'] == str(n)
        assert queued['scan'] == taken['scan'] == state['scan'] and queued['kind'] == state['kind']
        assert int(queued['step']) == int(state['queued_step']) < int(taken['step']) < int(state['step'])
        assert int(taken['irqs']) == 76 + n * 2 - 1
        assert int(irq[76 + n * 2 - 1]['step']) <= int(state['step'])
        if n > 1:
            before = states[n - 2]
            assert all(queued[key] == before[key] for key in ('step', 'phase', 'ida_linear'))
        if n >= 194:
            assert state['phase'] == ('choice' if n == 194 else 'ready')
            assert state['ida_linear'] == ('1f7b7' if n == 194 else '1997c')
            assert int(state['raw0013'], 16) & 0x4000 == 0
            assert state['player_y'] == '18' and state['player_x'] == ('1' if n == 196 else '2')
            assert all(state[key] == states[192][key] for key in ('actor', 'flags', 'gold_lo', 'gold_hi', 'raw0b24'))
        label = PREFIX + f'-packet-{n:03d}-' + state['phase']
        image, indexed = root / (label + '.png'), root / (label + '.bin')
        width, height, indices, _ = png(image)
        assert (width, height) == (640, 350) and indices == indexed.read_bytes()
        for p in (image, indexed):
            if n <= 193:
                assert p.read_bytes() == (root / p.name.replace(PREFIX, PARENT, 1)).read_bytes()
            artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    assert states[193]['choice_count'] == '10' and states[193]['choice_cursor'] == '1'
    retained = 0
    for artifact in json.loads((root / 'issue4-mother-finish-receipt.json').read_text())['artifacts']:
        if artifact['path'].endswith(('.png', '.bin')):
            p = root / artifact['path'].replace('issue4-mother-finish-', PREFIX + '-', 1)
            assert p.stat().st_size == artifact['size'] and digest(p) == artifact['sha256']
            retained += 1
    assert retained == 174
    texts = [fields(line) for line in lines if line.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert texts == parent['text_events'], 'cancel added dialogue'
    clocks = [fields(line) for line in lines if line.startswith('DQ3_LOAD_CANCEL_CLOCK ')]
    assert [row['packet'] for row in clocks] == [str(n) for n in range(193, 198)]
    for row in clocks:
        assert row['step'] == states[int(row['packet']) - 1]['step']
        assert row['clock'] == clocks[0]['clock'] and row['raw526c'] == clocks[0]['raw526c']
    initial = (root / (PREFIX + '-persistent-193.bin')).read_bytes()
    assert len(initial) == 2172
    for n in range(193, 198):
        p = root / (PREFIX + f'-persistent-{n:03d}.bin')
        expected = bytearray(initial)
        if n == 196:
            expected[0x4f33 - 0x4f29] = 1
        assert p.read_bytes() == expected
        artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    operations_path = root / (PREFIX + '-fileops.json')
    operations = json.loads(operations_path.read_text())
    assert operations and not list((root.parent / (PREFIX + '-scratch')).iterdir())
    later = [op for op in operations if op['Step'] >= int(states[192]['step'])]
    assert not any(op['Fn'] in (0x3c, 0x40, 0x41) or op.get('Name', '').lower().startswith('dragon') for op in later)
    native = [fields(line) for line in lines if line.startswith('DQ3_LOAD_CANCEL_NATIVE ')]
    done = [fields(line) for line in lines if line.startswith('DQ3_QUIESCENT_DONE ')]
    assert len(done) == 1 and done[0]['packets'] == '197' and done[0]['irqs'] == '470' and done[0]['step'] == states[-1]['step']
    return {'scope': '正常F6十槽Esc取消及左右行走；有限197包', 'meta': meta, **groups,
            'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'],
            'upstream_revision': meta['upstream_revision'], 'seed': meta['seed'],
            'parent_source_sha256': PARENT_HASH, 'prefix193_unchanged': True, 'parent174_unchanged': True,
            'seed_control_once': True, 'state_injection': False, 'emulator_snapshot_restore': False,
            'normal_inputs': 235, 'clocks': clocks, 'text_events': texts, 'native_events': native,
            'fileops': operations, 'fileops_sha256': digest(operations_path),
            'fileops_limit': '既有FileOps未記錄AH40；Scratch為空且完整持久區保持共同核對',
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
    print('正常F6取消来源接受', digest(args.receipt))
