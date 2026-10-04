"""獨立核對原版正常入隊播放後返回；入口 docs/188。

來源未完整返回時失敗，不把199包前綴或計時旗標當作完成證據。
"""
import argparse
import json
from pathlib import Path
import sys

from verify_dosgolem_mother_return import digest, fields

PARENT = 'issue4-recruit-party-r2'
PREFIX = 'issue4-recruit-return-r2'
PARENT_HASH = 'd0f6428dbc6f66b0887c3991c4bd17cb00be2825df4a53e1cf5bc049d806ed32'
GO_HASH = 'aa3c7b92741325ccc69d7dd5c90f5a3ae2a36b5a4a557c9a56906a5a2a52b22f'
BINARY_HASH = '0ec5657bb4473dc5d7d70dc7d57157822b89f038a7a1c292979cb50d038ab0ba'
PRODUCER_HASH = 'a7eb2be1a4987af091fbb3ec52ef281cf385f8507df4d97181d6f753f1e3d092'


def validate(root, producer, assets):
    parent_path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(parent_path) == PARENT_HASH, 'unaccepted parent source'
    parent = json.loads(parent_path.read_text())
    assert parent['seed'] == '1357' and parent['meta']['seed_configured_before_execution']
    exe = assets / 'DQ3.EXE'
    assert exe.stat().st_size == parent['meta']['original_size'] == 115282
    assert digest(exe) == parent['original_sha256']
    old_log = root / (PARENT + '.log')
    assert digest(old_log) == parent['log_sha256'], 'parent log changed'
    names = set()
    for item in parent['artifacts']:
        name = item['path']
        assert Path(name).name == name and name not in names
        names.add(name)
        old, new = root / name, root / name.replace(PARENT, PREFIX, 1)
        assert old.stat().st_size == item['size'] and digest(old) == item['sha256']
        assert old.read_bytes() == new.read_bytes(), 'original199 canvas changed'
    assert len(names) == 398
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    changed = {'producer_sha256', 'probe_source_sha256', 'probe_sha256', 'args'}
    assert set(meta) == set(parent['meta'])
    assert all(value == parent['meta'][key] for key, value in meta.items() if key not in changed)
    args = [arg.replace(PARENT, PREFIX) for arg in parent['meta']['args']]
    assert args[args.index('-steps') + 1] == '9000000001'
    args[args.index('-steps') + 1] = '2500000001'
    assert meta['args'] == args
    assert digest(producer) == meta['producer_sha256'] == PRODUCER_HASH
    assert digest(root / (PREFIX + '-probe-source.go')) == meta['probe_source_sha256'] == GO_HASH
    assert digest(root / (PREFIX + '-probe')) == meta['probe_sha256'] == BINARY_HASH
    log_path = root / (PREFIX + '.log')
    lines = log_path.read_text().splitlines()
    entries = [i for i, line in enumerate(lines) if line.startswith('DQ3_JOIN_AUDIO_ENTRY ')]
    assert len(entries) == 1
    entry = fields(lines[entries[0]])
    assert entry == {'step': parent['done']['step'], 'packet': '199', 'irqs': '474', 'record': '538'}
    inherited = [line.replace(PREFIX, PARENT) for line in lines[:entries[0]]
                 if line.startswith('DQ3_') and not line.startswith('DQ3_JOIN_RETURN ')]
    prior = [line for line in old_log.read_text().splitlines()
             if line.startswith('DQ3_') and not line.startswith('DQ3_RECRUIT_DONE ')]
    assert inherited == prior, 'original199 events changed'
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT',
            'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(line) for line in lines if line.startswith(tag + ' ')]
              for key, tag in tags.items()}
    q, c, states, irqs = (groups[key] for key in tags)
    assert len(q) == len(c) == len(states) == 202 and len(irqs) == 480, 'normal return incomplete'
    for key in groups:
        assert groups[key][:len(parent[key])] == parent[key], key
    assert [row['scan'] for row in q[199:]] == ['4d', '1c', '1c']
    assert [row['kind'] for row in q[199:]] == ['recruit_decline_cursor', 'recruit_decline', 'recruit_wait']
    assert [row['phase'] for row in states[199:]] == ['choice', 'waiting', 'ready']
    assert [row['last_record'] for row in states[199:]] == ['540', '541', '541']
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(row['scan'], 16) for row in q]
    assert [int(row['port60'], 16) for row in irqs] == [value for scan in scans for value in (scan, scan | 128)]
    assert [int(row['count']) for row in irqs] == list(range(1, 481))
    for number, (queued, consumed, state) in enumerate(zip(q, c, states), 1):
        assert queued['packet'] == consumed['packet'] == state['packet'] == str(number)
        assert queued['scan'] == consumed['scan'] == state['scan']
        assert int(queued['step']) < int(consumed['step']) <= int(state['step'])
    for state in states[199:]:
        for key in ('player_x', 'player_y', 'raw0b24', 'gold_lo', 'gold_hi', 'actor', 'flags'):
            assert state[key] == states[198][key], 'join return changed ' + key
    observations = [fields(line) for line in lines if line.startswith('DQ3_RECRUIT_STATE ')]
    assert observations[:len(parent['recruit_observations'])] == parent['recruit_observations']
    for row in observations[len(parent['recruit_observations']):]:
        for key in ('roster_flags', 'slot1', 'raw4f1f', 'raw4f15', 'raw4f17', 'raw4f19', 'raw4f1b'):
            assert row[key] == parent['recruit_observations'][-1][key], 'return transaction changed ' + key
    join = [fields(line) for line in lines if line.startswith('DQ3_JOIN_OBSERVE ')]
    assert join == parent['join_observations'], 'original party/name observation changed'
    returns = [fields(line) for line in lines if line.startswith('DQ3_JOIN_RETURN ')]
    natural = [row for row in returns if row['packet'] == '199']
    assert [row['ida_linear'] for row in natural] == ['10459', '1045e', '10469', '10398', '103ae', '103b6']
    assert all(int(entry['step']) < int(row['step']) < int(q[199]['step'])
               and row['irqs'] == '474' and row['actual_DS'] == '15ed' for row in natural)
    done = [fields(line) for line in lines if line.startswith('DQ3_RECRUIT_DONE ')]
    assert len(done) == 1 and done[0]['packets'] == '202' and done[0]['irqs'] == '480'
    assert done[0]['step'] == states[-1]['step'] and done[0]['player_x'] == '2' and done[0]['player_y'] == '18'
    artifacts = []
    for number, state in enumerate(states, 1):
        base = root / f'{PREFIX}-packet-{number:03d}-{state["phase"]}'
        for suffix in ('.png', '.bin'):
            path = base.with_suffix(suffix)
            assert path.is_file()
            if suffix == '.bin':
                assert path.stat().st_size == 224000
            artifacts.append({'path': path.name, 'size': path.stat().st_size, 'sha256': digest(path)})
    assert len(artifacts) == 404
    return dict(original_sha256=parent['original_sha256'], original_size=115282,
                upstream_revision=parent['upstream_revision'], seed='1357', meta=meta,
                scope='normal single-member join music return and No farewell',
                **groups, recruit_observations=observations, join_observations=join,
                audio_entry=entry, return_observations=returns, done=done[0], artifacts=artifacts,
                parent_source_sha256=PARENT_HASH, prefix199_unchanged=True,
                normal_inputs=240, state_injection=False, emulator_snapshot_restore=False,
                complete_join=True, natural_music_return=True, audio_wall_clock_parity=False,
                original_save_load_parity=False, remake_parity=False,
                log_sha256=digest(log_path), checker_sha256=digest(Path(__file__)), python_version=sys.version)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', type=Path)
    parser.add_argument('producer', type=Path)
    parser.add_argument('output', type=Path)
    parser.add_argument('--assets', type=Path, required=True)
    args = parser.parse_args()
    assert not args.output.exists(), 'refuse to overwrite existing receipt'
    result = validate(args.root, args.producer, args.assets)
    with args.output.open('x') as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2)
        stream.write('\n')
    print('正常202包／480IRQ1的自然播放後返回來源接受；未驗wall-clock或remake parity')
