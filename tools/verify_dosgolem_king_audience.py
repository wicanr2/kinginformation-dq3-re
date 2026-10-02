"""核對正常王座接近來源；有限入口與限制見 docs/188。只在 Docker 執行。"""
import argparse
import hashlib
import json
from pathlib import Path
import re


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def fields(line):
    return dict(re.findall(r'(\w+)=(\S+)', line))


def validate(path):
    d = json.loads(path.read_text())
    assert d['scenario'] == 'king_audience' and d['game_state_injection'] is False
    assert d['upstream_revision_observed'] == '2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    exe = Path(d['original_path'])
    assert exe.stat().st_size == d['original_size'] == 115282
    assert digest(exe) == d['original_sha256'] == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
    old_path = path.with_name('issue4-king-idle-receipt.json')
    assert digest(old_path) == '5cd11f15beb7e8c42f89780da825e5802b5cb38033b683e000bc06e32fb021d1'
    old = json.loads(old_path.read_text())
    assert d['player_input'][:49] == old['player_input']
    expected = [{'queued_step': 2260000000, 'scan': '0x48'}]
    expected += [{'queued_step': 2268000000 + i*20000000, 'scan': '0x48'} for i in range(35)]
    expected += [{'queued_step': 2990000000 + i*20000000, 'scan': '0x1c'} for i in range(5)]
    assert d['player_input'][49:] == expected and len(d['player_input']) == 90
    assert d['test_rng_seed_control'] == old['test_rng_seed_control']
    assert d['creation_results'] == old['creation_results']
    assert d['actual_irq1_events'][:98] == old['actual_irq1_events']
    assert len(d['actual_irq1_events']) == 180
    for i, packet in enumerate(expected):
        make, release = [fields(x) for x in d['actual_irq1_events'][98+i*2:100+i*2]]
        scan = int(packet['scan'], 16)
        assert int(make['count']) == 99+i*2 and int(release['count']) == 100+i*2
        assert int(make['port60'], 16) == scan and int(release['port60'], 16) == scan|0x80
        a, b = int(make['step']), int(release['step'])
        assert packet['queued_step'] < a < b < packet['queued_step']+1100000
        assert b-a >= 500000
    log = path.with_name('issue4-king-audience.log').read_text().splitlines()
    assert '沒實作的服務（0 種）：' in log
    assert not any('找不到的檔（' in x for x in log)
    for label, key in [('DQ3_KEY_DELIVERED ', 'actual_irq1_events'),
                       ('DQ3_KING_AUDIENCE ', 'king_audience_events'),
                       ('DQ3_IDLE_WINDOW ', 'idle_window_events')]:
        assert [x for x in log if x.startswith(label)] == d[key]
    assert d['idle_window_events'][:5] == old['idle_window_events']
    artifacts = d['artifacts']
    assert len(artifacts) == len({a['path'] for a in artifacts})
    assert len(artifacts) in (386, 387)
    mapping = {a['path']: a for a in artifacts}
    # 初次來源的 Go source 由執行中容器另存；新版 producer 自動列入 manifest。
    if len(artifacts) == 387:
        assert 'issue4-king-audience-probe-source.go' in mapping
    for a in artifacts:
        p = path.parent/a['path']
        assert p.name == a['path'] and p.stat().st_size == a['size'] and digest(p) == a['sha256']
    matched = 0
    for a in old['artifacts']:
        if not a['path'].endswith(('.png', '.bin')):
            continue
        b = mapping[a['path'].replace('issue4-king-idle-', 'issue4-king-audience-', 1)]
        assert (a['size'], a['sha256']) == (b['size'], b['sha256'])
        matched += 1
    assert matched == 214
    frozen = path.with_name('issue4-king-audience-generation.py')
    assert digest(frozen) == d['generation_script_sha256']
    probe_source = path.with_name('issue4-king-audience-probe-source.go')
    assert digest(probe_source) == d['probe_source_sha256']
    events = [fields(x) for x in d['king_audience_events']]
    ready = [x for x in events if x['phase'] == 'ready']
    positions = [(int(x['input']), int(x['player_x']), int(x['player_y'])) for x in ready]
    assert positions[:21] == [(1,15,30)] + [(i+2,15,29-i) for i in range(19)] + [(21,9,22)]
    assert positions[21:36] == [(22,9,22)] + [(i+23,9,21-i) for i in range(14)]
    assert positions[36:] == [(i,9,8) for i in range(37,42)]
    assert all(int(x['pit_divisor']) == 12428 and x['dgroup'] == '15ed' for x in events)
    assert len({x['actor'] for x in events}) == len({x['flags'] for x in events}) == 1
    assert ready[20]['raw0b24'] == '08c7'
    return {'source_receipt_sha256': digest(path), 'normal_inputs': 90, 'irq1_events': 180,
            'artifacts_verified': len(artifacts), 'prior_png_bin_unchanged': matched,
            'first_floor': {'input_ordinal_after_prior49':21, 'cty':25, 'section':1, 'x':9, 'y':22},
            'final_position': {'x':9,'y':8}, 'actor_flags_unchanged': True,
            'king_audience_completed':False, 'full_rgb_parity':False, 'audio_parity':False,
            'stack_contract':'1991D兩個堆疊欄位只保留raw words；不得推定為far return IP／CS'}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--receipt', type=Path, default=Path('/work/dosgolem-opening/issue4-king-audience-receipt.json'))
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    result = validate(args.receipt)
    if args.output:
        args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(result, ensure_ascii=False))
