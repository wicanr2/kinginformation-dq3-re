"""正常國王文字與交易來源核對；Docker 與證據範圍見 docs/188。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import runpy
import struct


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def fields(line):
    return dict(re.findall(r'(\w+)=(\S+)', line))


def inventory(event):
    raw = bytes.fromhex(event['actor'])
    assert len(raw) == 128
    return list(struct.unpack_from('<8H', raw, 0x3a))


def validate(path):
    parent_path = path.with_name('issue4-king-audience-receipt.json')
    prior_validator = runpy.run_path(str(Path(__file__).with_name('verify_dosgolem_king_audience.py')))['validate']
    prior = prior_validator(parent_path)
    parent = json.loads(parent_path.read_text())
    d = json.loads(path.read_text())
    assert d['scenario'] == 'king_text' and d['game_state_injection'] is False
    for key in ['original_path', 'original_size', 'original_sha256', 'upstream_revision_observed',
                'test_rng_seed_control', 'creation_results', 'patched_files_sha256',
                'patched_bios_sha256', 'patched_vga_sha256']:
        assert d[key] == parent[key], key
    expected = [{'queued_step': 3102000000, 'scan': '0x48'}]
    expected += [{'queued_step': 3130000000+i*20000000, 'scan': '0x1c'} for i in range(9)]
    assert d['player_input'] == parent['player_input']+expected
    assert d['actual_irq1_events'][:180] == parent['actual_irq1_events']
    assert len(d['actual_irq1_events']) == 200
    for i, packet in enumerate(expected):
        make, release = [fields(x) for x in d['actual_irq1_events'][180+i*2:182+i*2]]
        assert (int(make['count']), int(release['count'])) == (181+i*2, 182+i*2)
        scan = int(packet['scan'], 16)
        assert (int(make['port60'], 16), int(release['port60'], 16)) == (scan, scan | 0x80)
        a, b = int(make['step']), int(release['step'])
        assert packet['queued_step'] < a < b < packet['queued_step']+11000000
        assert b-a >= 500000
    for key in ['king_approach_events', 'king_ready_events', 'king_camera_events', 'layer_tile_events']:
        assert d[key] == parent[key]
    assert d['idle_status_events'][:len(parent['idle_status_events'])] == parent['idle_status_events']
    assert d['idle_window_events'] == parent['idle_window_events']
    log = path.with_name('issue4-king-text.log').read_text().splitlines()
    assert '沒實作的服務（0 種）：' in log
    assert not any('找不到的檔（' in x for x in log)
    for label, key in [('DQ3_KEY_DELIVERED ', 'actual_irq1_events'),
                       ('DQ3_IDLE_STATUS ', 'idle_status_events'),
                       ('DQ3_KING_TEXT ', 'king_text_events')]:
        assert [x for x in log if x.startswith(label)] == d[key]
    artifacts = d['artifacts']
    assert len(artifacts) == len({a['path'] for a in artifacts})
    mapping = {a['path']: a for a in artifacts}
    for a in artifacts:
        p = path.parent/a['path']
        assert p.name == a['path'] and p.stat().st_size == a['size'] and digest(p) == a['sha256']
    for name, key in [('issue4-king-text-generation.py', 'generation_script_sha256'),
                      ('issue4-king-text-probe-source.go', 'probe_source_sha256')]:
        assert name in mapping and digest(path.with_name(name)) == d[key]
    matched = 0
    for a in parent['artifacts']:
        if a['path'].endswith(('.png', '.bin')):
            b = mapping[a['path'].replace('issue4-king-audience-', 'issue4-king-text-', 1)]
            assert (a['size'], a['sha256']) == (b['size'], b['sha256'])
            matched += 1
    assert matched == 384
    events = [fields(x) for x in d['king_text_events']]
    steps = [int(x['step']) for x in events]
    assert all(a < b for a, b in zip(steps, steps[1:])), 'native event steps must be strictly increasing'
    by_phase = lambda phase: [x for x in events if x['phase'] == phase]
    entry, text, returned, gold, cleared, setflag, handler_return, ready = [by_phase(phase) for phase in
        ['handler_entry', 'text_entry', 'text_return', 'gold_return', 'clear_flag_return', 'set_flag_return', 'handler_return', 'normal_ready']]
    assert all(len(x) == 1 for x in [entry, text, returned, gold, cleared, setflag, handler_return, ready])
    entry, text, returned, gold, cleared, setflag, handler_return, ready = [x[0] for x in
        [entry, text, returned, gold, cleared, setflag, handler_return, ready]]
    assert all(x['player_x'] == '9' and x['player_y'] == '7' and x['raw0b24'] == '08c7'
               and x['dgroup'] == '15ed' and x['pit_divisor'] == '12428' for x in events)
    assert text['DI'] == '0c06'
    waiting, wait_entries, items = by_phase('text_waiting'), by_phase('wait_entry'), by_phase('item_return')
    assert len(waiting) == len(wait_entries) == 9
    assert [int(x['wait']) for x in waiting] == list(range(1, 10))
    assert [int(x['wait']) for x in wait_entries] == list(range(1, 10))
    before_slots = inventory(entry)
    for event in [text, returned]+waiting:
        assert inventory(event) == before_slots and event['flags'] == entry['flags']
        assert event['gold_lo'] == event['gold_hi'] == '0000'
    for i, waiting_event in enumerate(waiting):
        assert int(waiting_event['step']) < expected[i+1]['queued_step']
        assert int(wait_entries[i]['step']) < int(waiting_event['step'])
    assert len(items) == 6
    slots = before_slots[:]
    for event, code in zip(items, [0, 1, 1, 3, 31, 31]):
        index = slots.index(0x00ff)
        slots[index] = code
        assert inventory(event) == slots and int(event['raw_item2593'], 16) == code
        assert event['flags'] == entry['flags'] and event['gold_lo'] == '0000'
    assert int(returned['step']) < int(items[0]['step']) < int(items[-1]['step']) < int(gold['step'])
    assert int(gold['step']) < int(cleared['step']) < int(setflag['step']) < int(handler_return['step']) < int(ready['step'])
    assert inventory(ready) == slots
    assert gold['gold_lo'] == ready['gold_lo'] == '0032' and gold['gold_hi'] == ready['gold_hi'] == '0000'
    prior_flags = bytearray.fromhex(entry['flags'])
    # Native flags use MSB-first numbering. Keep the raw bytes as the oracle.
    assert prior_flags[2] & 1 and not prior_flags[3] & 0x80
    prior_flags[2] &= 0xfe
    assert bytes.fromhex(cleared['flags']) == prior_flags
    prior_flags[3] |= 0x80
    assert bytes.fromhex(setflag['flags']) == bytes.fromhex(ready['flags']) == prior_flags
    assert text['raw0b34'] == returned['raw0b34'] == '01' and ready['raw0b34'] == '00'
    return {'source_receipt_sha256': digest(path), 'prior_source': prior,
            'normal_inputs': 100, 'irq1_events': 200, 'artifacts_verified': len(artifacts),
            'prior_png_bin_unchanged': matched, 'normal_position': {'x': 9, 'y': 7},
            'native_waits': len(waiting), 'inventory_before': before_slots,
            'inventory_after': slots, 'gold_after': 50,
            'dialogue_before_rewards': True, 'king_audience_completed': True,
            'remake_parity': False, 'full_rgb_parity': False, 'audio_parity': False,
            'stack_contract': '兩個stack_word保留raw，不推定far caller'}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--receipt', type=Path, default=Path('/work/dosgolem-opening/issue4-king-text-receipt.json'))
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    result = validate(args.receipt)
    if args.output:
        args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(result, ensure_ascii=False))
