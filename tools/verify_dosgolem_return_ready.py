"""冷啟動逐鍵回程收據稽核。入口、位址與限制見 docs/188。"""
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
    prefix = 'issue4-return-cold-ready-r4'
    assert path.name == prefix+'-receipt.json'
    parent_path = path.with_name('issue4-king-text-receipt.json')
    prior = runpy.run_path(str(Path(__file__).with_name('verify_dosgolem_king_text.py')))['validate'](parent_path)
    parent = json.loads(parent_path.read_text())
    d = json.loads(path.read_text())
    assert d['scenario'] == 'king_return_cold_ready' and d['game_state_injection'] is False
    for key in ['original_path', 'original_size', 'original_sha256', 'upstream_revision_observed',
                'test_rng_seed_control', 'creation_results', 'patched_files_sha256',
                'patched_bios_sha256', 'patched_vga_sha256', 'original_text']:
        assert d[key] == parent[key], key
    for key in ['king_approach_events', 'king_ready_events', 'king_camera_events', 'layer_tile_events',
                'idle_status_events', 'idle_window_events', 'king_audience_events', 'king_text_events']:
        assert d[key] == parent[key], key
    assert d['player_input'][:100] == parent['player_input']
    assert d['actual_irq1_events'][:200] == parent['actual_irq1_events']
    log = path.with_name(prefix+'.log').read_text().splitlines()
    assert '沒實作的服務（0 種）：' in log
    assert not any('找不到的檔（' in line for line in log)
    for label, key in [('DQ3_KEY_DELIVERED ', 'actual_irq1_events'),
                       ('DQ3_RETURN_READY ', 'return_events'),
                       ('DQ3_RETURN_CAMERA ', 'camera_events'),
                       ('DQ3_RETURN_QUEUED ', 'queued_events'),
                       ('DQ3_RETURN_IDLE ', 'return_idle_events')]:
        assert [line for line in log if line.startswith(label)] == d[key], key
    queue = [fields(line) for line in d['queued_events']]
    assert queue and [int(q['packet']) for q in queue] == list(range(1, len(queue)+1))
    tail = [{'queued_step': int(q['step']), 'scan': hex(int(q['scan'], 16))} for q in queue]
    assert d['player_input'] == parent['player_input']+tail
    assert all(a['queued_step'] < b['queued_step'] for a, b in zip(tail, tail[1:]))
    motions = [q for q in queue if q['kind'] == 'motion']
    directions = ['50']*54+['4b']*9+['48']+['4b']*7+['48']*11+['4d']*3
    assert len(motions) == 85 and [q['scan'] for q in motions] == directions
    assert [int(q['ordinal']) for q in motions] == list(range(1, 86))
    assert all(q['ida_linear'] == '1997c' for q in motions)
    idle_queue = [q for q in queue if q['kind'] == 'idle_dismiss']
    assert len(queue) == len(motions)+len(idle_queue)
    assert all(q['scan'] == '1c' and 0 <= int(q['ordinal']) <= 85 for q in idle_queue)
    idle_events = [fields(line) for line in d['return_idle_events']]
    assert len(idle_events) == len(idle_queue)
    for q, event in zip(idle_queue, idle_events):
        assert (q['step'], q['ordinal']) == (event['step'], event['ordinal'])
        assert (event['ida_linear'], event['return_ip'], event['return_cs']) == ('21133', '7e00', '0110')
        assert event['key_flag'] == '00', '等待函式尚未完成按鍵旗標初始化'
    assert len(d['actual_irq1_events']) == (100+len(queue))*2
    for i, packet in enumerate(tail):
        make, release = [fields(line) for line in d['actual_irq1_events'][200+i*2:202+i*2]]
        assert (int(make['count']), int(release['count'])) == (201+i*2, 202+i*2)
        scan = int(packet['scan'], 16)
        assert (int(make['port60'], 16), int(release['port60'], 16)) == (scan, scan | 0x80)
        a, b = int(make['step']), int(release['step'])
        assert packet['queued_step'] < a < b and b-a >= 500000
        if i+1 < len(tail):
            assert b < tail[i+1]['queued_step'], '前一鍵尚未送完就續送'
    artifacts = d['artifacts']
    assert len(artifacts) == len({a['path'] for a in artifacts})
    mapping = {a['path']: a for a in artifacts}
    for a in artifacts:
        p = path.parent/a['path']
        assert p.name == a['path'] and p.stat().st_size == a['size'] and digest(p) == a['sha256']
    for suffix, key in [('generation.py', 'generation_script_sha256'), ('probe-source.go', 'probe_source_sha256')]:
        name = prefix+'-'+suffix
        assert name in mapping and digest(path.with_name(name)) == d[key]
    matched = 0
    for a in parent['artifacts']:
        if a['path'].endswith(('.png', '.bin')):
            b = mapping[a['path'].replace('issue4-king-text-', prefix+'-', 1)]
            assert (a['size'], a['sha256']) == (b['size'], b['sha256'])
            matched += 1
    assert matched == 432
    events = [fields(line) for line in d['return_events']]
    assert [int(e['ordinal']) for e in events] == list(range(1, 86))
    assert all(int(a['step']) < int(b['step']) for a, b in zip(events, events[1:]))
    ready = next(fields(line) for line in parent['king_text_events'] if 'phase=normal_ready ' in line)
    positions, conservation = [], []
    cameras = [fields(line) for line in d['camera_events']]
    camera_groups = {}
    for camera in cameras:
        ordinal = int(camera['ordinal'])
        assert 1 <= ordinal <= 85 and camera['ida_linear'] == '11991'
        assert (int(camera['width']), int(camera['height'])) == (20, 15)
        assert int(camera['origin_x']) == int(camera['player_x'])-9
        assert int(camera['origin_y']) == int(camera['player_y'])-7
        camera_groups.setdefault(ordinal, []).append(camera)
    # 11991 is the full redraw path. Ordinary steps use incremental drawing.
    assert {15, 36, 85}.issubset(camera_groups), '轉場的繪圖前攝影機觀測缺失'
    for motion, event in zip(motions, events):
        ordinal = int(event['ordinal'])
        assert event['queued_step'] == motion['step']
        assert int(event['step']) > int(motion['step'])+11000000
        assert event['ida_linear'] == '1991d' and event['dgroup'] == '15ed'
        assert event['pit_divisor'] == '12428'
        assert len(bytes.fromhex(event['actor'])) == 128 and len(bytes.fromhex(event['flags'])) == 64
        for suffix in ('png', 'bin'):
            assert f'{prefix}-return-ready-{ordinal:02d}.{suffix}' in mapping
        camera = camera_groups.get(ordinal, [None])[-1]
        if camera is not None:
            assert int(motion['step']) < int(camera['step']) < int(event['step'])
            assert all(camera[key] == event[key] for key in ['player_x', 'player_y', 'raw0b24'])
        positions.append({key: int(event[key]) for key in ['ordinal', 'player_x', 'player_y', 'raw4f25', 'raw4f27']}
                         | {'raw0b24': event['raw0b24'], 'raw4f1f': event['raw4f1f'],
                            'camera_origin_x': int(camera['origin_x']) if camera else None,
                            'camera_origin_y': int(camera['origin_y']) if camera else None})
        conservation.append({'ordinal': ordinal,
                             'inventory_unchanged': event['actor'][0x3a*2:0x4a*2] == ready['actor'][0x3a*2:0x4a*2],
                             'flags_unchanged': event['flags'] == ready['flags'],
                             'gold_unchanged': (event['gold_lo'], event['gold_hi']) == (ready['gold_lo'], ready['gold_hi'])})
    for q in idle_queue:
        for suffix in ('png', 'bin'):
            assert f"{prefix}-idle-{int(q['packet']):02d}.{suffix}" in mapping
    done = [fields(line) for line in log if line.startswith('DQ3_RETURN_DONE ')]
    assert len(done) == 1 and done[0]['inputs'] == '85'
    assert int(done[0]['packets']) == len(queue) and int(done[0]['irqs']) == len(d['actual_irq1_events'])
    assert done[0]['step'] == events[-1]['step']
    return {'source_receipt_sha256': digest(path), 'prior_source': prior,
            'normal_inputs': len(d['player_input']), 'motion_inputs': 85, 'extra_idle_inputs': len(idle_queue),
            'irq1_events': len(d['actual_irq1_events']), 'artifacts_verified': len(artifacts),
            'prior_png_bin_unchanged': matched, 'ready_observations': len(events),
            'camera_observations': len(cameras), 'positions': positions, 'conservation': conservation,
            'camera_observed_ordinals': sorted(camera_groups),
            'inference_level': 'confirmed: this cold-boot original normal input trace and raw observations only',
            'camera_observation': '11991 full redraw only; unobserved incremental steps remain null; 1991D raw4F25/raw4F27 retained separately',
            'remake_parity': False, 'full_rgb_parity': False, 'audio_parity': False,
            'original_save_load_parity': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--receipt', type=Path, default=Path('/work/dosgolem-opening/issue4-return-cold-ready-r4-receipt.json'))
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    result = validate(args.receipt)
    if args.output:
        args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(result, ensure_ascii=False))
