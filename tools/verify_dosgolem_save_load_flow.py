"""正常 F5 拒絕／取消與 F5→移動→F6 來源稽核；入口 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_save_load import validate as validate_entry

ENTRY = 'issue4-save-f5-normal-r1'
ENTRY_HASH = '3350c30fa5c48fc690b144a6860cac155a4e02efb498649b7120a9bbd5f0973d'
IDENTITIES = {
    'decline': ('issue4-save-f5-decline-r1', 197, '94997502473015eb7449362ae998b9b2a698bec384d363e4e239af916a41dc1a', '43c3e56cbd77732455a75659b52ccf3df2955e86839a11cd23d0e3bec607f448'),
    'roundtrip': ('issue4-save-load-roundtrip-r1', 200, 'dd0e56f32dd0c162d1bc45426f2e0a986dada785e3eae6c23b8eaf9ddd8c8842', '709318601f3a3896eb4c51d76dc229084fd5f9088bc7420c43a597169bb72cc6'),
    'cancel': ('issue4-save-load-cancel-r1', 197, 'a1d167274208dedf99bc0f6f27300fe0024f0e3bdb5bf5372bf2db07ad5cf028', '7b36ed00901a50f41b543427f3f9b4f6bd7a542af3b0e089bfe2ea59bf4fece0'),
}


def validate(root, action, producer):
    prefix, count, go_hash, binary_hash = IDENTITIES[action]
    path = root / (ENTRY + '-source-r1-receipt.json')
    assert digest(path) == ENTRY_HASH
    prior = json.loads(path.read_text())
    assert prior == validate_entry(root, Path(__file__).with_name('dosgolem_save_load_probe.py'))
    meta = json.loads((root / (prefix + '-meta.json')).read_text())
    assert meta['original_size'] == 115282 and meta['original_sha256'] == prior['original_sha256']
    assert meta['upstream_revision'] == prior['upstream_revision']
    assert meta['seed'] == '1357' and meta['seed_configured_before_execution'] is True
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['producer_sha256'] == digest(producer)
    assert meta['generator_sha256'] == digest(Path(__file__).with_name('dosgolem_newgame_probe.py'))
    assert all(meta[k] == v for k, v in RUNTIME_HASHES.items())
    assert meta['docker_image'] == 'dq3-ebiten-test:20260822-r1' and meta['build_flags'] == ['-trimpath', '-p', '2']
    assert meta['probe_source_sha256'] == digest(root / (prefix + '-probe-source.go')) == go_hash
    assert meta['probe_sha256'] == digest(root / (prefix + '-probe')) == binary_hash
    assert meta['normal_prefix_inputs'] == prior['meta']['normal_prefix_inputs']
    lines = (root / (prefix + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines and not any('找不到的檔（' in s for s in lines)
    assert not any(s.startswith('DQ3_SAVE_BOUND ') for s in lines)
    seed = [s for s in lines if s.startswith('DQ3_CREATION_SEED ')]
    assert len(seed) == 1 and 'fixed=1357' in seed[0]
    scratch = [fields(s) for s in lines if s.startswith('DQ3_SAVE_SCRATCH ')]
    assert scratch == [{'step': '0', 'path': '/work/' + prefix + '-scratch'}]
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT',
            'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(s) for s in lines if s.startswith(tag + ' ')] for key, tag in tags.items()}
    q, c, states, irq = (groups[key] for key in tags)
    assert len(q) == len(c) == len(states) == count and len(irq) == 76 + count * 2
    for key in ('queued', 'consumed', 'states', 'actual_irq1_events'):
        assert groups[key][:len(prior[key])] == prior[key]
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(s['scan'], 16) for s in q]
    assert [int(s['port60'], 16) for s in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(s['count']) for s in irq] == list(range(1, len(irq) + 1))
    phases = {'ready': '1997c', 'inline_wait': '216d8', 'waiting': '21133', 'choice': '1f7b7', 'name': '11096'}
    artifacts = []
    for n, (a, b, s) in enumerate(zip(q, c, states), 1):
        assert a['packet'] == b['packet'] == s['packet'] == str(n)
        assert a['scan'] == b['scan'] == s['scan'] and a['kind'] == s['kind']
        assert int(a['step']) == int(s['queued_step']) < int(b['step']) < int(s['step'])
        assert int(b['irqs']) == 76 + n * 2 - 1 and int(irq[76 + n * 2 - 1]['step']) <= int(s['step'])
        assert s['ida_linear'] == phases[s['phase']] and int(s['raw0013'], 16) & 0x4000 == 0
        assert len(bytes.fromhex(s['actor'])) == 128 and len(bytes.fromhex(s['flags'])) == 64
        if n > 1:
            before = states[n - 2]
            assert a['step'] == before['step'] and a['phase'] == before['phase'] and a['ida_linear'] == before['ida_linear']
        if n >= 194:
            assert all(s[k] == states[192][k] for k in ('actor', 'flags', 'gold_lo', 'gold_hi', 'raw0b24'))
        label = prefix + f'-packet-{n:03d}-' + s['phase']
        image, indexed = root / (label + '.png'), root / (label + '.bin')
        width, height, indices, _ = png(image)
        assert (width, height) == (640, 350) and indices == indexed.read_bytes()
        for p in (image, indexed):
            artifacts.append({'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)})
            if n <= 194:
                assert p.read_bytes() == (root / p.name.replace(prefix, ENTRY, 1)).read_bytes()
    if action == 'decline':
        expected_scans, expected_phases, expected_records = ['3f', '4d', '1c', '1c'], ['choice', 'choice', 'waiting', 'ready'], ['253', '253', '252', '252']
        assert states[194]['choice_cursor'] == '2'
    elif action == 'cancel':
        expected_scans, expected_phases, expected_records = ['3f', '01', '01', '1c'], ['choice', 'choice', 'waiting', 'ready'], ['253', '250', '252', '252']
        assert states[194]['choice_cursor'] == '1' and states[194]['choice_count'] == '10'
    else:
        expected_scans = ['3f', '1c', '1c', '1c', '4b', '40', '1c']
        expected_phases = ['choice', 'choice', 'waiting', 'ready', 'ready', 'choice', 'ready']
        expected_records = ['253', '250', '252', '252', '252', '252', '252']
        assert states[194]['choice_count'] == states[198]['choice_count'] == '10'
        assert states[194]['choice_cursor'] == states[198]['choice_cursor'] == '1'
    assert [s['scan'] for s in q[193:]] == expected_scans
    assert [s['phase'] for s in states[193:]] == expected_phases
    assert [s['last_record'] for s in states[193:]] == expected_records
    assert all(s['player_y'] == '18' and s['player_x'] == ('1' if action == 'roundtrip' and int(s['packet']) in (198, 199) else '2') for s in states[193:])
    texts = [fields(s) for s in lines if s.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert texts[:len(prior['text_events'])] == prior['text_events']
    suffix = [('196', '00fc')] if action == 'decline' else [('195', '00fa'), ('196', '00fc')]
    assert [(s['packet'], s['DI']) for s in texts[len(prior['text_events']):]] == suffix
    assert all(s['text_segment'] == '2826' for s in texts)
    done = [fields(s) for s in lines if s.startswith('DQ3_QUIESCENT_DONE ')]
    assert len(done) == 1 and done[0]['packets'] == str(count) and done[0]['irqs'] == str(len(irq)) and done[0]['step'] == states[-1]['step']
    operations_path = root / (prefix + '-fileops.json')
    operations = json.loads(operations_path.read_text())
    assert operations and all(isinstance(op['Step'], int) and isinstance(op['Failed'], bool) for op in operations)
    local_scratch = root.parent / (prefix + '-scratch')
    native = [fields(s) for s in lines if s.startswith('DQ3_SAVE_LOAD_NATIVE ')]
    if action != 'roundtrip':
        assert local_scratch.is_dir() and not list(local_scratch.iterdir())
        assert not any(op['Fn'] in (0x3c, 0x40, 0x41) for op in operations if op['Step'] >= int(states[192]['step']))
        assert not native
    else:
        assert sorted(p.name for p in local_scratch.iterdir()) == ['dragon0.dat', 'player.dat']
        saved = (local_scratch / 'dragon0.dat').read_bytes()
        directory = (local_scratch / 'player.dat').read_bytes()
        assert len(saved) == 2172 and len(directory) == 200
        assert digest(local_scratch / 'dragon0.dat') == '9ef549b0ac969e8bb9a0138d4e339905c21f5abf8653ca718c09d4b4c5db5910'
        assert digest(local_scratch / 'player.dat') == '58a34d4793c2a28cbcc7f7dbe8da413d3a245798cf28e85208cef6343695ec25'
        actor = bytes.fromhex(states[195]['actor'])
        assert directory[:18] == actor[3:21] and directory[18] == actor[21] and directory[19] == actor[2]
        original_directory = Path('/repo/assets_raw/PLAYER.DAT').read_bytes()
        assert directory[20:] == original_directory[20:]
        persistent = []
        for n in range(193, 201):
            p = root / (prefix + f'-persistent-{n:03d}.bin')
            data = p.read_bytes()
            assert len(data) == 2172
            if n in (196, 197, 200):
                assert data == saved
            elif n in (198, 199):
                expected = bytearray(saved); expected[0x4f33 - 0x4f29] = 1
                assert data == expected
            else:
                assert data == (root / (prefix + '-persistent-193.bin')).read_bytes()
            persistent.append({'path': p.name, 'size': len(data), 'sha256': digest(p)})
        artifacts.extend(persistent)
        for p in sorted(local_scratch.iterdir()):
            artifacts.append({'path': str(p.relative_to(root.parent)), 'size': p.stat().st_size, 'sha256': digest(p)})
        created = [op for op in operations if op['Op'] == 'create']
        assert [op['Name'] for op in created] == ['player.dat', 'dragon0.dat'] and not any(op['Failed'] for op in created)
        reads = [op for op in operations if op['Op'] == 'read' and op['Name'] == 'dragon0.dat' and op['Step'] > int(states[198]['step'])]
        assert len(reads) == 1 and reads[0]['Len'] == reads[0]['Arg'] == 2172 and not reads[0]['Failed']
        assert [s['ida_linear'] for s in native] == ['11484', '114c8', '114d3', '114d9', '1157d', '1158b', '11591', '1165f']
        assert all(s['DS'] == '15ed' and s['raw0722'] == '1' and s['raw0726'] == '0' for s in native)
        assert native[2]['AX'] == '400b' and native[2]['CX'] == '087c' and native[2]['DX'] == '4f29'
        assert native[5]['AX'] == '3f0b' and native[5]['CX'] == '087c' and native[5]['DX'] == '4f29'
        assert all(s['raw01f0'] == ('0' if i < 4 else '1') for i, s in enumerate(native))
    return {'scope': '正常F5有限分支與第一槽Save→移動→F6 Load；只限實際觀測範圍', 'action': action,
            'original_size': meta['original_size'], 'original_sha256': meta['original_sha256'],
            'upstream_revision': meta['upstream_revision'], 'seed': meta['seed'], 'meta': meta,
            'normal_inputs': 38 + count, **groups, 'text_events': texts, 'artifacts': artifacts,
            'prefix194_unchanged': True, 'seed_control_once': True, 'state_injection': False,
            'emulator_snapshot_restore': False, 'native_game_load': action == 'roundtrip',
            'fileops_sha256': digest(operations_path), 'fileops': operations, 'native_events': native,
            'fileops_limit': '既有FileOps未記錄AH40；真正寫入由唯讀caller觀察、落地檔及Load後完整2172bytes共同核對',
            'done': done[0], 'remake_parity': False, 'full_rgb_parity': False, 'audio_parity': False,
            'checker_sha256': digest(Path(__file__)), 'log_sha256': digest(root / (prefix + '.log'))}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', type=Path)
    parser.add_argument('action', choices=tuple(IDENTITIES))
    parser.add_argument('producer', type=Path)
    parser.add_argument('receipt', type=Path)
    args = parser.parse_args()
    assert not args.receipt.exists()
    result = validate(args.root, args.action, args.producer)
    with args.receipt.open('x') as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2); stream.write('\n')
    print('正常存讀檔分支原版來源接受', args.action, digest(args.receipt))
