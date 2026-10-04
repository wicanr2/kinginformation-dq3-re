"""獨立稽核正常202包的人物影格唯讀觀測；證據入口 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_after_load_move import validate as validate_parent
from verify_dosgolem_mother_return import digest, fields

PREFIX = 'issue4-field-pose-normal-r1'
PARENT = 'issue4-after-load-move-r1'
PARENT_HASH = '71a52768a26dabb174a3a96e53efe6aaebba78c7620b20cce39fe2f761548e10'
GO_HASH = 'f2eb1325fd3bc34c04dd2b3f4b0501c24ab1c165e21209550ffd50017456507f'
BINARY_HASH = 'ea91fc0e280be53023bec263df088d0294849c7c3547fb993e2f763f4ec69c1a'


def validate(root, producer):
    parent_path = root / (PARENT + '-source-r2-receipt.json')
    assert digest(parent_path) == PARENT_HASH
    parent = json.loads(parent_path.read_text())
    assert parent == validate_parent(root, Path(__file__).with_name('dosgolem_after_load_move_probe.py'))
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    changed = {'producer_sha256', 'probe_source_sha256', 'probe_sha256', 'args'}
    assert set(meta) == set(parent['meta'])
    assert all(value == parent['meta'][key] for key, value in meta.items() if key not in changed)
    assert meta['args'] == [arg.replace(PARENT, PREFIX) for arg in parent['meta']['args']]
    assert meta['producer_sha256'] == digest(producer)
    assert meta['probe_source_sha256'] == digest(root / (PREFIX + '-probe-source.go')) == GO_HASH
    assert meta['probe_sha256'] == digest(root / (PREFIX + '-probe')) == BINARY_HASH
    log = root / (PREFIX + '.log')
    lines = log.read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line or line.startswith('DQ3_SAVE_BOUND ') for line in lines)
    inherited = [line.replace(PREFIX, PARENT) for line in lines
                 if line.startswith('DQ3_') and not line.startswith('DQ3_FIELD_POSE_')]
    parent_lines = (root / (PARENT + '.log')).read_text().splitlines()
    assert inherited == [line for line in parent_lines if line.startswith('DQ3_')]
    artifacts = []
    for artifact in parent['artifacts']:
        old = root.parent / artifact['path'] if '-scratch/' in artifact['path'] else root / artifact['path']
        new = old.with_name(old.name.replace(PARENT, PREFIX, 1))
        if '-scratch/' in artifact['path']:
            new = root.parent / artifact['path'].replace(PARENT, PREFIX, 1)
        raw = new.read_bytes()
        assert raw == old.read_bytes()
        assert len(raw) == artifact['size'] and digest(new) == artifact['sha256']
        artifacts.append({**artifact, 'path': artifact['path'].replace(PARENT, PREFIX, 1)})
    scratch = root.parent / (PREFIX + '-scratch')
    assert sorted(p.name for p in scratch.iterdir()) == ['dragon0.dat', 'player.dat']
    assert json.loads((root / (PREFIX + '-fileops.json')).read_text()) == parent['fileops']
    boundaries = [fields(line) for line in lines if line.startswith('DQ3_FIELD_POSE_BOUNDARY ')]
    blits = [fields(line) for line in lines if line.startswith('DQ3_FIELD_POSE_BLIT ')]
    assert [row['packet'] for row in boundaries] == [str(n) for n in range(193, 203)]
    for row in boundaries:
        assert row['step'] == parent['states'][int(row['packet']) - 1]['step']
        assert row['raw0004'] in ('0', '1')
    assert blits and [int(row['step']) for row in blits] == sorted(int(row['step']) for row in blits)
    for row in blits:
        n, step = int(row['packet']), int(row['step'])
        assert 193 <= n <= 202 and row['ida_linear'] == '1e307' and row['DS'] == '15ed'
        assert row['raw0004'] in ('0', '1')
        assert int(row['BX'], 16) % 2 == 0
        assert (int(row['BX'], 16) // 2) % 2 == int(row['raw0004'])
        assert int(parent['queued'][n - 1]['step']) <= step <= int(parent['states'][n - 1]['step'])
    assert all(any(row['packet'] == str(n) for row in blits) for n in (198, 200, 201, 202))
    return {'scope': '正常202包人物影格的唯讀consumer觀測；不證明動畫時序parity',
            'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'],
            'upstream_revision': meta['upstream_revision'], 'seed': meta['seed'], 'meta': meta,
            'parent_source_sha256': PARENT_HASH, 'all202_inputs_states_irq_unchanged': True,
            'all202_png_bin_persistent_unchanged': True, 'native_files_unchanged': True,
            'normal_inputs': parent['normal_inputs'], 'actual_irq1_events': parent['actual_irq1_events'],
            'queued': parent['queued'], 'consumed': parent['consumed'], 'states': parent['states'],
            'boundaries': boundaries, 'blits': blits, 'artifacts': artifacts,
            'state_injection': False, 'emulator_snapshot_restore': False,
            'checker_sha256': digest(Path(__file__)), 'log_sha256': digest(log),
            'remake_parity': False, 'animation_timing': 'unknown'}


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
    print('正常人物影格來源接受', digest(args.receipt))
