"""獨立接受正常 Space 開窗原版來源；入口 docs/188，未宣稱 remake parity。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_recruitment_menu_cancel import validate as validate_parent

PREFIX = 'issue4-command-menu-r1'
PARENT = 'issue4-recruit-menu-esc-normal-r1'
PARENT_HASH = '137411c711e81684943b4cc8aac2d952b8c47c8981f70a79f0d308e369204c4c'
PRODUCER_HASH = '9ce24fb63da9385189ad537784bcdc99f4bbbdd1d3d5aaa15b4f03ba1484468d'
GO_HASH = 'db8a2e15894894f9a5b7e1aefd0232fbe3fb180f69ca50c649249ec8e38096f7'
BINARY_HASH = 'be27c97c0f56d7224e6f7bf0e911f3ecdce38da85207c0ea4d849c32eb7b5b1a'


def validate(root, producer):
    path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(path) == PARENT_HASH
    parent = json.loads(path.read_text())
    assert parent == validate_parent(root, Path(__file__).with_name('dosgolem_recruitment_menu_cancel_probe.py'))
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'docker_image', 'build_flags'):
        assert meta[key] == parent['meta'][key]
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    assert meta['producer_sha256'] == digest(producer) == PRODUCER_HASH
    assert all(meta[key] == value for key, value in RUNTIME_HASHES.items())
    assert meta['probe_source_sha256'] == digest(root / (PREFIX + '-probe-source.go')) == GO_HASH
    assert meta['probe_sha256'] == digest(root / (PREFIX + '-probe')) == BINARY_HASH
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line or line.startswith('DQ3_SAVE_BOUND ') for line in lines)
    seed = [line for line in lines if line.startswith('DQ3_CREATION_SEED ')]
    assert len(seed) == 1 and 'fixed=1357' in seed[0]
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT', 'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(line) for line in lines if line.startswith(tag + ' ')] for key, tag in tags.items()}
    q, taken, states, irq = (groups[key] for key in tags)
    assert len(q) == len(taken) == len(states) == 194 and len(irq) == 464
    for key in groups:
        assert groups[key][:len(parent[key])] == parent[key]
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(row['scan'], 16) for row in q]
    assert [int(row['port60'], 16) for row in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(row['count']) for row in irq] == list(range(1, 465))
    assert q[-1]['scan'] == '39' and q[-1]['kind'] == 'command_space_open'
    final = states[-1]
    assert final['phase'] == 'choice' and final['ida_linear'] == '1f7b7'
    assert final['choice_count'] == '6' and final['choice_cursor'] == '1'
    assert all(final[key] == states[-2][key] for key in ('player_x', 'player_y', 'actor', 'flags', 'gold_lo', 'gold_hi', 'raw0b24'))
    artifacts = []
    for n, (queued, consumed, state) in enumerate(zip(q, taken, states), 1):
        assert queued['packet'] == consumed['packet'] == state['packet'] == str(n)
        assert queued['scan'] == consumed['scan'] == state['scan'] and queued['kind'] == state['kind']
        assert int(queued['step']) == int(state['queued_step']) < int(consumed['step']) < int(state['step'])
        assert int(consumed['irqs']) == 76 + n * 2 - 1
        assert int(irq[76 + n * 2 - 1]['step']) <= int(state['step'])
        if n > 1:
            assert all(queued[key] == states[n - 2][key] for key in ('step', 'phase', 'ida_linear'))
        label = PREFIX + f'-packet-{n:03d}-' + state['phase']
        image, indexed = root / (label + '.png'), root / (label + '.bin')
        width, height, indices, _ = png(image)
        assert (width, height) == (640, 350) and indices == indexed.read_bytes()
        for p in (image, indexed):
            if n <= 193:
                assert p.read_bytes() == (root / p.name.replace(PREFIX, PARENT, 1)).read_bytes()
            artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    retained = 0
    for artifact in json.loads((root / 'issue4-mother-finish-receipt.json').read_text())['artifacts']:
        if artifact['path'].endswith(('.png', '.bin')):
            p = root / artifact['path'].replace('issue4-mother-finish-', PREFIX + '-', 1)
            assert p.stat().st_size == artifact['size'] and digest(p) == artifact['sha256']
            retained += 1
    assert retained == 174
    texts = [fields(line) for line in lines if line.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert texts == parent['text_events']
    clocks = [fields(line) for line in lines if line.startswith('DQ3_COMMAND_CLOCK ')]
    assert [row['packet'] for row in clocks] == ['193', '194']
    for row in clocks:
        assert row['step'] == states[int(row['packet']) - 1]['step'] and row['clock'] == '30' and row['raw526c'] == '1'
    initial = (root / (PREFIX + '-persistent-193.bin')).read_bytes()
    assert len(initial) == 2172
    for n in (193, 194):
        p = root / (PREFIX + f'-persistent-{n:03d}.bin')
        assert p.read_bytes() == initial
        artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    ops_path = root / (PREFIX + '-fileops.json')
    operations = json.loads(ops_path.read_text())
    assert operations and not list((root.parent / (PREFIX + '-scratch')).iterdir())
    later = [op for op in operations if op['Step'] >= int(states[-2]['step'])]
    assert not any(op['Fn'] in (0x3c, 0x40, 0x41) or op.get('Name', '').lower().startswith('dragon') for op in later)
    done = [fields(line) for line in lines if line.startswith('DQ3_SAVE_OBSERVED ')]
    assert len(done) == 1 and done[0]['packet'] == '194' and done[0]['step'] == final['step'] and done[0]['phase'] == 'choice'
    return {'scope': '正常Space開啟六指令窗；有限194包', 'meta': meta, **groups,
            'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'], 'upstream_revision': meta['upstream_revision'],
            'seed': meta['seed'], 'seed_control_once': True, 'parent_source_sha256': PARENT_HASH,
            'prefix193_unchanged': True, 'parent174_unchanged': True, 'normal_inputs': 232,
            'state_injection': False, 'emulator_snapshot_restore': False, 'clocks': clocks, 'text_events': texts,
            'fileops': operations, 'fileops_sha256': digest(ops_path), 'artifacts': artifacts, 'done': done[0],
            'checker_sha256': digest(Path(__file__)), 'log_sha256': digest(root / (PREFIX + '.log')),
            'remake_parity': False, 'full_rgb_parity': False}


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
    print('正常Space命令窗來源接受', digest(args.receipt))
