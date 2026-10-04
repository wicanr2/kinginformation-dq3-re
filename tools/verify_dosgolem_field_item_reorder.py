"""接受正常單人穿戴物給予自己的八格順序來源；入口 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_field_item_navigation import validate as validate_parent
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = 'issue4-field-item-reorder-r2'
PARENT = 'issue4-field-item-navigation-r2'
PARENT_HASH = '28995c8dd702f83d70c51f3f09212dc556356ed563d5b0b9d3977ff4b0d60eb4'
PRODUCER_HASH = '6da18ef2f8c620584b6b6dc482748b48be09ed20b058bece5ba307b64743d427'
SOURCE_HASH = 'd6c2caa9af49d440a1927b8fbff98bba996dbdac1ccd8b449bf192b3032a3b65'
BINARY_HASH = '8c090dd10dd878d35d1c246eb150eefa7a3fde074d40595dfd29c71835e94bcc'


def validate(root, producer):
    parent_path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(parent_path) == PARENT_HASH
    parent = json.loads(parent_path.read_text())
    assert parent == validate_parent(root, Path(__file__).with_name('dosgolem_field_item_navigation_probe.py'))
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'docker_image', 'build_flags'):
        assert meta[key] == parent['meta'][key]
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    assert '-load-state' not in meta['args']
    assert meta['producer_sha256'] == digest(producer) == PRODUCER_HASH
    assert all(meta[key] == value for key, value in RUNTIME_HASHES.items())
    assert meta['probe_source_sha256'] == digest(root / (PREFIX + '-probe-source.go')) == SOURCE_HASH
    assert meta['probe_sha256'] == digest(root / (PREFIX + '-probe')) == BINARY_HASH
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line or line.startswith('DQ3_SAVE_BOUND ') for line in lines)
    seeds = [fields(line) for line in lines if line.startswith('DQ3_CREATION_SEED ')]
    assert len(seeds) == 1 and seeds[0]['fixed'] == '1357'
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT', 'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(line) for line in lines if line.startswith(tag + ' ')] for key, tag in tags.items()}
    q, taken, states, irq = (groups[key] for key in tags)
    assert len(q) == len(taken) == len(states) == 230, 'normal packet count'
    assert len(irq) == 536, 'IRQ1 count'
    for key in groups:
        assert groups[key][:len(parent[key])] == parent[key]
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(row['scan'], 16) for row in q]
    assert [int(row['port60'], 16) for row in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(row['count']) for row in irq] == list(range(1, 537))
    expected = [(219, '39', 'choice', 6, 1), (220, '50', 'choice', 6, 2),
                (221, '4d', 'choice', 6, 5), (222, '39', 'choice', 7, 1),
                (223, '39', 'choice', 3, 1), (224, '50', 'choice', 3, 2),
                (225, '39', 'waiting', 3, 1), (226, '1c', 'ready', 6, 5),
                (227, '39', 'choice', 6, 1), (228, '50', 'choice', 6, 2),
                (229, '4d', 'choice', 6, 5), (230, '39', 'choice', 7, 1)]
    before_slots = bytes.fromhex('1e8000000100010003001f001f00ff00')
    after_slots = before_slots[2:] + before_slots[:2]
    actor_before = bytes.fromhex(states[217]['actor'])
    assert len(actor_before) == 128 and actor_before[0x3a:0x4a] == before_slots
    actor_after = actor_before[:0x3a] + after_slots + actor_before[0x4a:]
    for n, scan, phase, count, cursor in expected:
        assert q[n-1]['scan'] == scan
        state = states[n-1]
        assert state['phase'] == phase and state['choice_count'] == str(count) and state['choice_cursor'] == str(cursor)
        assert state['player_x'] == '3' and state['player_y'] == '18'
        assert all(state[key] == states[217][key] for key in ('flags', 'gold_lo', 'gold_hi', 'raw0b24'))
        assert bytes.fromhex(state['actor']) == (actor_before if n < 225 else actor_after)
        assert state['last_record'] == ('541' if n < 225 else '308')
    artifacts = []
    for n, (queued, consumed, state) in enumerate(zip(q, taken, states), 1):
        assert queued['packet'] == consumed['packet'] == state['packet'] == str(n)
        assert queued['scan'] == consumed['scan'] == state['scan'] and queued['kind'] == state['kind']
        assert int(queued['step']) == int(state['queued_step']) < int(consumed['step']) < int(state['step'])
        assert int(consumed['irqs']) == 76 + n * 2 - 1
        assert int(irq[76+n*2-1]['step']) <= int(state['step'])
        if n > 1:
            assert all(queued[key] == states[n-2][key] for key in ('step', 'phase', 'ida_linear'))
        label = PREFIX + f'-packet-{n:03d}-' + state['phase']
        image, indexed = root / (label + '.png'), root / (label + '.bin')
        width, height, indices, _ = png(image)
        assert (width, height) == (640, 350) and indices == indexed.read_bytes()
        for p in (image, indexed):
            if n <= 218:
                assert p.read_bytes() == (root / p.name.replace(PREFIX, PARENT, 1)).read_bytes()
            artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    clocks = [fields(line) for line in lines if line.startswith('DQ3_COMMAND_CLOCK ')]
    assert [row['packet'] for row in clocks] == list(map(str, range(193, 231)))
    assert clocks[:len(parent['clocks'])] == parent['clocks']
    for row in clocks:
        assert row['step'] == states[int(row['packet'])-1]['step'] and row['clock'] == '30' and row['raw526c'] == '1'
    baseline = (root / (PARENT + '-persistent-218.bin')).read_bytes()
    # 原版 DGROUP 4F29..57A4；角色八格在 DGROUP 50B9..50C8。
    slot_offset = 0x50b9 - 0x4f29
    assert len(baseline) == 2172 and baseline[slot_offset:slot_offset+16] == before_slots
    changed = baseline[:slot_offset] + after_slots + baseline[slot_offset+16:]
    for n in range(193, 231):
        p = root / (PREFIX + f'-persistent-{n:03d}.bin')
        if n <= 218:
            assert p.read_bytes() == (root / p.name.replace(PREFIX, PARENT, 1)).read_bytes()
        else:
            assert p.read_bytes() == (baseline if n < 225 else changed), f'persistent region n={n}'
        artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
    rasters = [fields(line) for line in lines if line.startswith('DQ3_COMMAND_RASTER ')]
    items = [fields(line) for line in lines if line.startswith('DQ3_ITEM_RASTER ')]
    assert [row['packet'] for row in rasters] == list(map(str, range(194, 231)))
    assert [row['packet'] for row in items] == list(map(str, range(206, 231)))
    assert rasters[:len(parent['rasters'])] == parent['rasters']
    assert items[:len(parent['item_rasters'])] == parent['item_rasters']
    for row in rasters + items:
        assert row['step'] == states[int(row['packet'])-1]['step']
    exe = root.parents[1] / 'assets_raw/DQ3.EXE'
    assert exe.stat().st_size == meta['original_size'] and digest(exe) == meta['original_sha256']
    raw = exe.read_bytes()
    list_window = bytearray(raw[0x1a118:0x1a118+96])
    list_window[8:10] = (144).to_bytes(2, 'little')
    list_window[12:14] = (7).to_bytes(2, 'little')
    action_window = raw[0x1a190:0x1a190+64]
    for row in items[len(parent['item_rasters']):]:
        assert row['owner062d'] == row['selected062f'] == '1' and row['raw2591'] == '0000'
        expected_window = bytearray(list_window)
        # 關閉清單後 raw count 保留7；開動作窗才清0，重開清單再寫7。
        expected_window[20:22] = (7 if int(row['packet']) <= 222 or int(row['packet']) == 230 else 0).to_bytes(2, 'little')
        assert row['window3fd8'] == expected_window.hex() and row['action4050'] == action_window.hex()
    writers = [fields(line) for line in lines if line.startswith('DQ3_ITEM_REORDER_WRITER ')]
    assert len(writers) == 4
    for row, n, address, slots in zip(writers, (225, 225, 225, 226), ('13a62', '13a87', '13a9f', '18197'), (before_slots, before_slots, after_slots, after_slots)):
        assert row['packet'] == str(n) and row['ida_linear'] == address
        assert row['owner'] == row['selected'] == '1'
        assert row['slots'] == slots.hex(), 'writer slots'
        assert row['raw2591'] == '0000'
        assert int(q[n-1]['step']) < int(row['step']) < int(states[n-1]['step'])
    assert writers[1]['AX'] == writers[2]['BX'] == '801e' and writers[1]['SI'] == '50b9' and writers[2]['SI'] == '50c7'
    texts = [fields(line) for line in lines if line.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert texts[:len(parent['text_events'])] == parent['text_events']
    assert len(texts) == len(parent['text_events']) + 1
    # 沿用的觀察 tag 也記錄新交易 record308；父路線七筆仍逐欄保持。
    added_text = texts[-1]
    assert (added_text['packet'], added_text['DI'], added_text['text_segment']) == ('225', '0134', '2826'), 'give text record'
    assert int(q[224]['step']) < int(added_text['step']) < int(writers[0]['step'])
    ops_path = root / (PREFIX + '-fileops.json')
    ops = json.loads(ops_path.read_text())
    assert ops and not list((root.parent / (PREFIX + '-scratch')).iterdir())
    assert not any(op['Fn'] in (0x3c, 0x40, 0x41) or op.get('Name', '').lower().startswith('dragon') for op in ops if op['Step'] >= int(states[192]['step']))
    done = [fields(line) for line in lines if line.startswith('DQ3_SAVE_OBSERVED ')]
    assert len(done) == 1 and done[0]['packet'] == '230' and done[0]['step'] == states[-1]['step']
    return {'scope': '正常單人第一件穿戴物給予自己、保留空格移至第八格、返回field後重開清單；有限230包',
            'meta': meta, **groups, 'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'],
            'upstream_revision': meta['upstream_revision'], 'seed': meta['seed'], 'seed_control_once': True,
            'parent_source_sha256': PARENT_HASH, 'prefix218_unchanged': True, 'normal_inputs': 268,
            'state_injection': False, 'emulator_snapshot_restore': False, 'clocks': clocks, 'rasters': rasters,
            'item_rasters': items, 'writer_events': writers, 'before_slots': before_slots.hex(), 'after_slots': after_slots.hex(),
            'text_events': texts, 'fileops': ops, 'fileops_sha256': digest(ops_path),
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
    print('正常單人穿戴物重排來源接受', digest(args.receipt))
