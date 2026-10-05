"""接受正常木棒使用的有限原版收據；證據與限制見 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = 'issue4-item-use-normal-r2'
PARENT = 'issue4-field-item-reorder-r2'
PARENT_HASH = 'aad971bb8cbf08956384b6964b4a893251cb3b4ff316263953f7563a7e5b2fc4'
PRODUCER_HASH = 'e5e5d61aa5f85a1dd0f29e6b2629f05594c1c9684e73d28479b6d03497259e4a'
SOURCE_HASH = '508927c5e17d22a4c9c1cf82390a90019aa0f476ae8d00536f48184a6f417840'
BINARY_HASH = 'ed5dcc0675b2bbb2cccd55ca5c29b61d598c473bb01aad4b0b4619e8ab642117'


def validate(root, producer):
    parent_path = root / (PARENT + '-source-r3-receipt.json')
    assert digest(parent_path) == PARENT_HASH, 'accepted parent identity'
    parent = json.loads(parent_path.read_text())
    assert digest(root / (PARENT + '.log')) == parent['log_sha256']
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed',
                'normal_prefix_inputs', 'generator_sha256', 'docker_image', 'build_flags'):
        assert meta[key] == parent['meta'][key], key
    assert all(meta[k] == v for k, v in RUNTIME_HASHES.items())
    assert meta['seed_configured_before_execution']
    assert not meta['state_restore'] and not meta['gameplay_state_injection']
    assert '-load-state' not in meta['args']
    assert digest(producer) == meta['producer_sha256'] == PRODUCER_HASH
    assert digest(root / (PREFIX + '-probe-source.go')) == meta['probe_source_sha256'] == SOURCE_HASH
    assert digest(root / (PREFIX + '-probe')) == meta['probe_sha256'] == BINARY_HASH
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in x or x.startswith('DQ3_SAVE_BOUND ') for x in lines)
    seeds = [fields(x) for x in lines if x.startswith('DQ3_CREATION_SEED ')]
    assert len(seeds) == 1 and seeds[0]['fixed'] == '1357'
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT',
            'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED',
            'clocks': 'DQ3_COMMAND_CLOCK', 'rasters': 'DQ3_COMMAND_RASTER',
            'item_rasters': 'DQ3_ITEM_RASTER', 'writer_events': 'DQ3_ITEM_REORDER_WRITER',
            'text_events': 'DQ3_EMPTY_RECRUIT_TEXT'}
    groups = {k: [fields(x) for x in lines if x.startswith(tag+' ')] for k, tag in tags.items()}
    for key, rows in groups.items():
        assert rows[:len(parent[key])] == parent[key], 'prefix230 '+key
    q, taken, states, irq = (groups[k] for k in ('queued', 'consumed', 'states', 'actual_irq1_events'))
    assert len(q) == len(taken) == len(states) == 233, 'normal packet count'
    assert len(irq) == 542, 'IRQ1 count'
    scans = [v for _, v in meta['normal_prefix_inputs']] + [int(x['scan'], 16) for x in q]
    assert len(scans) == 271
    assert [int(x['port60'], 16) for x in irq] == [v for scan in scans for v in (scan, scan | 128)], 'IRQ1 make/break'
    assert [int(x['count']) for x in irq] == list(range(1, 543))
    before = states[229]
    for n, scan, phase, count, cursor, record in (
            (231, '39', 'choice', 3, 1, '308'),
            (232, '39', 'waiting', 3, 1, '341'),
            (233, '1c', 'ready', 6, 5, '341')):
        state = states[n-1]
        assert q[n-1]['scan'] == scan and state['phase'] == phase
        assert state['choice_count'] == str(count) and state['choice_cursor'] == str(cursor)
        assert state['last_record'] == record
        assert all(state[k] == before[k] for k in ('actor', 'flags', 'gold_lo', 'gold_hi', 'player_x', 'player_y', 'raw0b24')), 'unchanged player state'
    artifacts = []
    for a in parent['artifacts']:
        old = root / a['path']
        new = root / a['path'].replace(PARENT, PREFIX, 1)
        assert old.stat().st_size == new.stat().st_size == a['size']
        assert digest(old) == digest(new) == a['sha256'], 'prefix230 artifact '+new.name
        artifacts.append(dict(path=new.name, size=a['size'], sha256=a['sha256']))
    for n, (queued, consumed, state) in enumerate(zip(q, taken, states), 1):
        assert queued['packet'] == consumed['packet'] == state['packet'] == str(n)
        assert queued['scan'] == consumed['scan'] == state['scan']
        assert queued['kind'] == state['kind']
        assert int(queued['step']) == int(state['queued_step']) < int(consumed['step']) < int(state['step'])
        assert int(consumed['irqs']) == 76+n*2-1
        assert int(irq[76+n*2-1]['step']) <= int(state['step'])
        if n > 1:
            assert all(queued[k] == states[n-2][k] for k in ('step', 'phase', 'ida_linear'))
        if n <= 230:
            continue
        label = PREFIX+f'-packet-{n:03d}-'+state['phase']
        image, indexed = root/(label+'.png'), root/(label+'.bin')
        width, height, pixels, _ = png(image)
        assert (width, height) == (640, 350) and pixels == indexed.read_bytes(), 'new PNG/index identity'
        persistent = root/(PREFIX+f'-persistent-{n:03d}.bin')
        baseline = root/(PREFIX+'-persistent-230.bin')
        assert len(persistent.read_bytes()) == 2172 and persistent.read_bytes() == baseline.read_bytes(), 'no-effect persistent region'
        for p in (image, indexed, persistent):
            artifacts.append(dict(path=p.name, size=p.stat().st_size, sha256=digest(p)))
    assert len(artifacts) == 507
    assert [x['packet'] for x in groups['clocks']] == list(map(str, range(193, 234)))
    assert [x['packet'] for x in groups['rasters']] == list(map(str, range(194, 234)))
    assert [x['packet'] for x in groups['item_rasters']] == list(map(str, range(206, 234)))
    for key in ('clocks', 'rasters', 'item_rasters'):
        for row in groups[key]:
            assert row['step'] == states[int(row['packet'])-1]['step']
    assert all(x['clock'] == '30' and x['raw526c'] == '1' for x in groups['clocks'])
    assert len(groups['writer_events']) == len(parent['writer_events'])
    text = groups['text_events'][len(parent['text_events']):]
    assert [(x['packet'], x['DI'], x['text_segment']) for x in text] == [('232', '0111', '2826'), ('232', '0155', '2826')], 'use introduction/result records'
    obs = [fields(x) for x in lines if x.startswith('DQ3_ITEM_USE_OBSERVE ')]
    raw_window = (root.parents[1]/'assets_raw/DQ3.EXE').read_bytes()[0x19fae:0x19fae+30].hex()
    assert obs and all(x['window3e6e'] == raw_window and x['owner'] == x['selected'] == '1' for x in obs)
    glyphs = [x for x in obs if x['ida_linear'] == '214b9']
    expected = [(210, 192, 254), (412, 216, 254), (56, 312, 254)]
    expected += [(code, 168+i*24, 270) for i, code in enumerate((508, 399, 435, 436, 494, 410, 147, 431, 273, 56))]
    assert [(int(x['BX'], 16), int(x['BP'], 16)*8, int(x['DX'], 16)) for x in glyphs] == expected, 'ordinary glyph placement'
    records = [x for x in obs if x['ida_linear'] == '21414']
    assert [(x['DI'], x['raw259b']) for x in records] == [('0111', '0'), ('0155', '1')]
    callbacks = [x['ida_linear'] for x in obs if x['ida_linear'] not in ('214b9', '21414', '21133')]
    assert callbacks == ['13942', '13c6d', '15002', '2111b'], 'use callback/order'
    assert all(x['raw259b'] == '2' for x in obs if x['ida_linear'] == '21133')
    exe = root.parents[1]/'assets_raw/DQ3.EXE'
    assert exe.stat().st_size == meta['original_size'] and digest(exe) == meta['original_sha256']
    ops_path = root/(PREFIX+'-fileops.json')
    ops = json.loads(ops_path.read_text())
    assert ops and not list((root.parent/(PREFIX+'-scratch')).iterdir())
    assert not any(x['Fn'] in (0x3c, 0x40, 0x41) or x.get('Name', '').lower().startswith('dragon') for x in ops if x['Step'] >= int(states[192]['step']))
    done = [fields(x) for x in lines if x.startswith('DQ3_SAVE_OBSERVED ')]
    assert len(done) == 1 and done[0]['packet'] == '233' and done[0]['step'] == states[-1]['step']
    return dict(scope='正常單人木棒使用：原版record273/341、保留父畫面、新Enter返回場景；有限233包',
                meta=meta, **groups, observations=obs, artifacts=artifacts, done=done[0],
                parent_source_sha256=PARENT_HASH, prefix230_unchanged=True, normal_inputs=271,
                original_size=meta['original_size'], original_sha256=meta['original_sha256'],
                upstream_revision=meta['upstream_revision'], seed=meta['seed'], seed_control_once=True,
                state_injection=False, emulator_snapshot_restore=False, no_effect_persistent_unchanged=True,
                fileops_sha256=digest(ops_path), log_sha256=digest(root/(PREFIX+'.log')),
                checker_sha256=digest(Path(__file__)), remake_parity=False, full_rgb_parity=False)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', type=Path)
    parser.add_argument('producer', type=Path)
    parser.add_argument('receipt', type=Path)
    args = parser.parse_args()
    assert not args.receipt.exists()
    report = validate(args.root, args.producer)
    with args.receipt.open('x') as f:
        json.dump(report, f, ensure_ascii=False, indent=2)
        f.write('\n')
    print('SOURCE_PASS', len(report['artifacts']), digest(args.receipt), flush=True)
