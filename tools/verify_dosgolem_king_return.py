"""正常謁見後回程來源核對；有限範圍與入口見 docs/188。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import runpy


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def fields(line):
    return dict(re.findall(r'(\w+)=(\S+)', line))


def validate(path):
    parent_path = path.with_name('issue4-king-text-receipt.json')
    prior_validator = runpy.run_path(str(Path(__file__).with_name('verify_dosgolem_king_text.py')))['validate']
    prior = prior_validator(parent_path)
    parent = json.loads(parent_path.read_text())
    d = json.loads(path.read_text())
    assert d['scenario'] == 'king_return' and d['game_state_injection'] is False
    for key in ['original_path', 'original_size', 'original_sha256', 'upstream_revision_observed',
                'test_rng_seed_control', 'creation_results', 'patched_files_sha256',
                'patched_bios_sha256', 'patched_vga_sha256', 'original_text']:
        assert d[key] == parent[key], key
    directions = [0]*54+[2]*9+[1]+[2]*7+[1]*11+[3]*3
    scans = ['0x50', '0x48', '0x4b', '0x4d']
    tail = [{'queued_step': 3324000000+i*20000000, 'scan': scans[direction]}
            for i, direction in enumerate(directions)]
    assert len(tail) == 85 and d['player_input'] == parent['player_input']+tail
    assert d['actual_irq1_events'][:200] == parent['actual_irq1_events']
    assert len(d['actual_irq1_events']) == 370
    for i, packet in enumerate(tail):
        make, release = [fields(x) for x in d['actual_irq1_events'][200+i*2:202+i*2]]
        assert (int(make['count']), int(release['count'])) == (201+i*2, 202+i*2)
        scan = int(packet['scan'], 16)
        assert (int(make['port60'], 16), int(release['port60'], 16)) == (scan, scan | 0x80)
        a, b = int(make['step']), int(release['step'])
        assert packet['queued_step'] < a < b < packet['queued_step']+11000000
        assert b-a >= 500000
    for key in ['king_approach_events', 'king_ready_events', 'king_camera_events', 'layer_tile_events']:
        assert d[key] == parent[key], key
    for key in ['idle_status_events', 'idle_window_events', 'king_audience_events', 'king_text_events']:
        # New normal inputs may add observations; the entire prior trace must survive.
        assert d[key][:len(parent[key])] == parent[key], key
    log = path.with_name('issue4-king-return.log').read_text().splitlines()
    assert '沒實作的服務（0 種）：' in log
    assert not any('找不到的檔（' in x for x in log)
    for label, key in [('DQ3_KEY_DELIVERED ', 'actual_irq1_events'),
                       ('DQ3_IDLE_STATUS ', 'idle_status_events'),
                       ('DQ3_IDLE_WINDOW ', 'idle_window_events'),
                       ('DQ3_KING_AUDIENCE ', 'king_audience_events'),
                       ('DQ3_KING_TEXT ', 'king_text_events'),
                       ('DQ3_KING_RETURN ', 'king_return_events')]:
        assert [x for x in log if x.startswith(label)] == d[key], key
    artifacts = d['artifacts']
    assert len(artifacts) == len({a['path'] for a in artifacts})
    mapping = {a['path']: a for a in artifacts}
    for a in artifacts:
        p = path.parent/a['path']
        assert p.name == a['path'] and p.stat().st_size == a['size'] and digest(p) == a['sha256']
    for name, key in [('issue4-king-return-generation.py', 'generation_script_sha256'),
                      ('issue4-king-return-probe-source.go', 'probe_source_sha256')]:
        assert name in mapping and digest(path.with_name(name)) == d[key]
    matched = 0
    for a in parent['artifacts']:
        if a['path'].endswith(('.png', '.bin')):
            b = mapping[a['path'].replace('issue4-king-text-', 'issue4-king-return-', 1)]
            assert (a['size'], a['sha256']) == (b['size'], b['sha256'])
            matched += 1
    assert matched == 432
    for ordinal in range(1, 86):
        for suffix in ('png', 'bin'):
            assert f'issue4-king-return-king-return-step-{ordinal:02d}.{suffix}' in mapping
    events = [fields(x) for x in d['king_return_events']]
    assert events, '未觀測到回程正常runner'
    ordinals = [int(x['ordinal']) for x in events]
    steps = [int(x['step']) for x in events]
    ready_artifacts = {int(match.group(1)) for name in mapping
                       if (match := re.fullmatch(r'issue4-king-return-return-ready-(\d{2})\.png', name))}
    assert ready_artifacts == set(ordinals), 'normal runner images and events must agree'
    assert all(a < b for a, b in zip(ordinals, ordinals[1:]))
    assert all(a < b for a, b in zip(steps, steps[1:]))
    ready = next(fields(x) for x in parent['king_text_events'] if 'phase=normal_ready ' in x)
    conservation = []
    positions = []
    for event, ordinal in zip(events, ordinals):
        assert 1 <= ordinal <= 85
        queued = tail[ordinal-1]['queued_step']
        assert int(event['queued_step']) == queued
        assert queued+11000000 < int(event['step']) < queued+20000000
        assert event['ida_linear'] == '1991d' and event['dgroup'] == '15ed'
        assert event['pit_divisor'] == '12428'
        assert len(bytes.fromhex(event['actor'])) == 128 and len(bytes.fromhex(event['flags'])) == 64
        for suffix in ('png', 'bin'):
            assert f'issue4-king-return-return-ready-{ordinal:02d}.{suffix}' in mapping
        # Retain raw identities; do not infer a CTY or section from a field name.
        positions.append({key: int(event[key]) for key in ['ordinal', 'player_x', 'player_y']}
                         | {'raw0b24': event['raw0b24'], 'raw4f1f': event['raw4f1f'],
                            'raw4f25': int(event['origin_x']), 'raw4f27': int(event['origin_y'])})
        conservation.append({'ordinal': ordinal,
                             'inventory_unchanged': event['actor'][0x3a*2:0x4a*2] == ready['actor'][0x3a*2:0x4a*2],
                             'flags_unchanged': event['flags'] == ready['flags'],
                             'gold_unchanged': (event['gold_lo'], event['gold_hi']) == (ready['gold_lo'], ready['gold_hi'])})
    return {'source_receipt_sha256': digest(path), 'prior_source': prior,
            'normal_inputs': 185, 'irq1_events': 370, 'artifacts_verified': len(artifacts),
            'prior_png_bin_unchanged': matched, 'ready_observations': len(events),
            'missing_ready_ordinals': sorted(set(range(1, 86))-set(ordinals)),
            'positions': positions, 'conservation': conservation,
            'inference_level': 'confirmed: this original input trace and raw observations only',
            'scene_identity': 'raw DGROUP fields retained; CTY source/consumer closure required',
            'camera_label_correction': 'old origin_y stores post-drawing DGROUP4F27; keep raw value, do not call it camera origin',
            'remake_parity': False, 'full_rgb_parity': False, 'audio_parity': False,
            'original_save_load_parity': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--receipt', type=Path, default=Path('/work/dosgolem-opening/issue4-king-return-receipt.json'))
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    result = validate(args.receipt)
    if args.output:
        args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(result, ensure_ascii=False))
