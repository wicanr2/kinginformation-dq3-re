"""原版正常 F5 開窗來源稽核；有限範圍及重生入口 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_recruitment_menu_cancel import validate as validate_parent

PREFIX = 'issue4-save-f5-normal-r1'
PARENT = 'issue4-recruit-menu-esc-normal-r1'
PARENT_HASH = '137411c711e81684943b4cc8aac2d952b8c47c8981f70a79f0d308e369204c4c'
GO_HASH = '7b5274e4c26c533806285f3a32f3713f219b7bb237c349c49c73288c02fafcf5'
BINARY_HASH = 'a47c7a8783c7056f9312f9c4df6e7559f16c44becd4a78f73377444914f7ffb7'


def validate(root, producer):
    prior_path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(prior_path) == PARENT_HASH
    prior = json.loads(prior_path.read_text())
    assert prior == validate_parent(root, Path(__file__).with_name('dosgolem_recruitment_menu_cancel_probe.py'))
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    assert meta['original_size'] == 115282 and meta['original_sha256'] == prior['original_sha256']
    assert meta['upstream_revision'] == '2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    assert meta['seed'] == '1357' and meta['seed_configured_before_execution'] is True
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['producer_sha256'] == digest(producer)
    assert meta['generator_sha256'] == digest(Path(__file__).with_name('dosgolem_newgame_probe.py'))
    assert all(meta[k] == v for k, v in RUNTIME_HASHES.items())
    assert meta['docker_image'] == 'dq3-ebiten-test:20260822-r1' and meta['build_flags'] == ['-trimpath', '-p', '2']
    assert meta['probe_source_sha256'] == digest(root / (PREFIX + '-probe-source.go')) == GO_HASH
    assert meta['probe_sha256'] == digest(root / (PREFIX + '-probe')) == BINARY_HASH
    assert meta['normal_prefix_inputs'] == prior['meta']['normal_prefix_inputs']
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in s or s.startswith('DQ3_SAVE_BOUND ') for s in lines)
    seed = [s for s in lines if s.startswith('DQ3_CREATION_SEED ')]
    assert len(seed) == 1 and 'fixed=1357' in seed[0]
    scratch = [fields(s) for s in lines if s.startswith('DQ3_SAVE_SCRATCH ')]
    assert scratch == [{'step': '0', 'path': '/work/' + PREFIX + '-scratch'}]
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT',
            'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(s) for s in lines if s.startswith(tag + ' ')] for key, tag in tags.items()}
    q, c, states, irq = (groups[key] for key in tags)
    assert len(q) == len(c) == len(states) == 194 and len(irq) == 464
    for key in ('queued', 'consumed', 'states'):
        assert groups[key][:193] == prior[key]
    assert irq[:462] == prior['actual_irq1_events']
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(s['scan'], 16) for s in q]
    assert [int(s['port60'], 16) for s in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(s['count']) for s in irq] == list(range(1, 465))
    phases = {'ready': '1997c', 'inline_wait': '216d8', 'waiting': '21133', 'choice': '1f7b7', 'name': '11096'}
    artifacts = []
    for n, (a, b, s) in enumerate(zip(q, c, states), 1):
        assert a['packet'] == b['packet'] == s['packet'] == str(n)
        assert a['scan'] == b['scan'] == s['scan'] and a['kind'] == s['kind']
        assert int(a['step']) == int(s['queued_step']) < int(b['step']) < int(s['step'])
        assert int(b['irqs']) == 76 + n * 2 - 1
        assert int(irq[76 + n * 2 - 1]['step']) <= int(s['step'])
        assert s['ida_linear'] == phases[s['phase']] and int(s['raw0013'], 16) & 0x4000 == 0
        assert len(bytes.fromhex(s['actor'])) == 128 and len(bytes.fromhex(s['flags'])) == 64
        if n > 1:
            before = states[n - 2]
            assert a['step'] == before['step'] and a['phase'] == before['phase'] and a['ida_linear'] == before['ida_linear']
        label = PREFIX + f'-packet-{n:03d}-' + s['phase']
        image, indexed = root / (label + '.png'), root / (label + '.bin')
        width, height, indices, _ = png(image)
        assert (width, height) == (640, 350) and indices == indexed.read_bytes()
        for path in (image, indexed):
            artifacts.append({'path': path.name, 'size': path.stat().st_size, 'sha256': digest(path)})
            if n <= 193:
                assert path.read_bytes() == (root / path.name.replace(PREFIX, PARENT, 1)).read_bytes()
    last = states[-1]
    assert q[-1]['kind'] == 'save_f5' and q[-1]['scan'] == '3f'
    assert last['phase'] == 'choice' and last['last_record'] == '253'
    assert last['choice_count'] == '2' and last['choice_cursor'] == '1'
    for key in ('actor', 'flags', 'gold_lo', 'gold_hi', 'player_x', 'player_y', 'raw0b24'):
        assert last[key] == states[-2][key]
    texts = [fields(s) for s in lines if s.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert texts[:-2] == prior['text_events']
    assert [(s['packet'], s['DI'], s['text_segment']) for s in texts[-2:]] == [
        ('194', '00fb', '2826'), ('194', '00fd', '2826')]
    done = [fields(s) for s in lines if s.startswith('DQ3_QUIESCENT_DONE ')]
    assert len(done) == 1 and done[0]['packets'] == '194' and done[0]['irqs'] == '464'
    assert done[0]['step'] == last['step']
    operations_path = root / (PREFIX + '-fileops.json')
    operations = json.loads(operations_path.read_text())
    assert operations and all(isinstance(op['Step'], int) and isinstance(op['Failed'], bool) for op in operations)
    assert not any(op['Fn'] in (0x3c, 0x40, 0x41) for op in operations if op['Step'] >= int(states[-2]['step']))
    local_scratch = root.parent / (PREFIX + '-scratch')
    assert local_scratch.is_dir() and not list(local_scratch.iterdir())
    parent_path = root / 'issue4-mother-finish-receipt.json'
    assert digest(parent_path) == meta['parent_receipt_sha256'] == '9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    retained = 0
    for a in json.loads(parent_path.read_text())['artifacts']:
        if a['path'].endswith(('.png', '.bin')):
            path = root / a['path'].replace('issue4-mother-finish-', PREFIX + '-', 1)
            assert path.stat().st_size == a['size'] and digest(path) == a['sha256']
            retained += 1
    assert retained == 174
    return {'scope': '正常招募返回後F5顯示253確認，尚未選Yes或存檔',
            'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'],
            'upstream_revision': meta['upstream_revision'], 'seed': meta['seed'], 'meta': meta,
            'normal_inputs': 232, **groups, 'text_events': texts, 'artifacts': artifacts,
            'prefix193_unchanged': True, 'parent174_unchanged': True, 'seed_control_once': True,
            'state_injection': False, 'scratch': scratch[0], 'fileops_sha256': digest(operations_path),
            'fileops': operations, 'done': done[0], 'remake_parity': False, 'full_rgb_parity': False,
            'original_save_load_parity': False, 'audio_parity': False,
            'checker_sha256': digest(Path(__file__)), 'log_sha256': digest(root / (PREFIX + '.log'))}


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
    print('正常F5確認原版來源接受', digest(args.receipt))
