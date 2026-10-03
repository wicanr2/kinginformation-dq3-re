"""母親帶路後首次北行的有限原版來源核對；入口 docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = 'issue4-registry-quiescent-r2'
SOURCE_HASH = 'ecc9714025c26789ad7ed32e7645833a96528393a6fe9ea1bf72fe2271b49405'
LOG_HASH = 'b98fc4d8c622c69544b7ebe3a32929fc3dcd4e270d72f7ac673d67b70b608239'
RUNTIME_HASHES = {
    'upstream_files_sha256': '463a8a83315a3054af75d6220a8d5bcc1464021ea28bad15659e6b9d5a0ce69e',
    'patched_files_sha256': '3d0f45b34f9ea88dfef28b7da8a7663303579551c559eeb7dc0cad390a5d1e13',
    'upstream_bios_sha256': '5b8e4bf8f8996af35c135cc0cedae628011108d058ddb64726c0623c1d9b8512',
    'patched_bios_sha256': '1de55aa13479ecc27de1fdd1436a79482967a2e53dd36e73d14ed7715f94c604',
    'upstream_vga_sha256': 'ba076659ae7da9eddbe01be2118f1ffdd518912d56edec123ab969216a294f86',
    'patched_vga_sha256': 'a9203b643e1b81160b0930279f3895b0473f15efc4e802e220b91531dd3d7cac',
}
EXPECTED_IMAGES = {
    'event-19530': '26c026a0828220cf66e81e3be256e228dc11f0816c901da5941738261b6f9356',
    'event-196d2': 'ebd60e01da883de93d8999afb3400f78396b21666495986ff207753cafe033bf',
    'event-1020b': 'ebd60e01da883de93d8999afb3400f78396b21666495986ff207753cafe033bf',
    'event-10232': '47a942e6d560d47b5411cfc57163dc61c39c2431fdf2f9d4f5d53cd5e56de271',
    'event-10245': 'e4cc65b5ea83ad62f70c05efdd5d1b49e24358b385b343eb5ebe2901e8ce101e',
    'packet-001-ready': 'e4cc65b5ea83ad62f70c05efdd5d1b49e24358b385b343eb5ebe2901e8ce101e',
    'packet-002-ready': '1e032767f4be32c0b1169e0e8973367a478ab5af5b3f60f39e6acf78a0c8dda4',
    'packet-003-ready': '37245124a75ded76a3abb5a0d4e70e87fb12a4e826a4950d8a961bf2cedf2446',
}


def source_receipt(root):
    parent_path = root/'issue4-mother-finish-receipt.json'
    assert digest(parent_path) == '9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    parent = json.loads(parent_path.read_text())
    meta_path = root/(PREFIX+'-meta.json')
    meta = json.loads(meta_path.read_text())
    assert meta['original_size'] == 115282 and meta['original_sha256'] == parent['original_sha256']
    assert meta['upstream_revision'] == '2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    assert meta['scenario'] == 'registry_quiescent_cold_draft'
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['seed'] == '1357' and meta['seed_configured_before_execution'] is True
    assert meta['parent_receipt_sha256'] == digest(parent_path)
    assert meta['build_flags'] == ['-trimpath', '-p', '2']
    assert meta['producer_sha256'] == digest(Path(__file__).with_name('dosgolem_registry_quiescent_probe.py'))
    assert meta['generator_sha256'] == digest(Path(__file__).with_name('dosgolem_newgame_probe.py'))
    assert meta['probe_source_sha256'] == digest(root/(PREFIX+'-probe-source.go')) == SOURCE_HASH
    assert meta['probe_sha256'] == digest(root/(PREFIX+'-probe')) == '8411a2709038fbe21dad7a64e12f89c4704e3fe9c6fa4b40b46941934721c9a8'
    assert meta['docker_image'] == 'dq3-ebiten-test:20260822-r1'
    for key, value in RUNTIME_HASHES.items():
        assert meta[key] == value
    retained = []
    for a in parent['artifacts']:
        if a['path'].endswith(('.png', '.bin')):
            p = root/a['path'].replace('issue4-mother-finish-', PREFIX+'-', 1)
            assert p.stat().st_size == a['size'] and digest(p) == a['sha256']
            retained.append(p)
    assert len(retained) == 174
    log_path = root/(PREFIX+'.log')
    assert digest(log_path) == LOG_HASH
    lines = log_path.read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in l for l in lines)
    seeds = [l for l in lines if l.startswith('DQ3_CREATION_SEED ')]
    assert len(seeds) == 1 and 'fixed=1357' in seeds[0]
    irq = [l for l in lines if l.startswith('DQ3_KEY_DELIVERED ')][:82]
    assert len(irq) == 82 and irq[:76] == parent['actual_irq1_events']
    assert [int(fields(l)['count']) for l in irq] == list(range(1, 83))
    assert [fields(l)['port60'] for l in irq[76:]] == ['48', 'c8']*3
    queued = [fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_QUEUED ')][:3]
    consumed = [fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_INPUT ')][:3]
    states = [fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_CAPTURE ')][:3]
    assert len(queued) == len(consumed) == len(states) == 3
    for i, (q, c, s) in enumerate(zip(queued, consumed, states), 1):
        assert q['packet'] == c['packet'] == s['packet'] == str(i)
        assert q['scan'] == c['scan'] == s['scan'] == '48'
        assert q['kind'] == s['kind'] == 'king_up' and s['phase'] == 'ready'
        assert int(q['step']) < int(c['step']) < int(fields(irq[75+i*2])['step']) < int(s['step'])
        assert (s['player_x'], s['player_y']) == ('21', str(16-i))
        assert s['raw0b24'] == '000c' and s['gold_lo'] == s['gold_hi'] == '0000'
        assert s['flags'] == states[0]['flags'] and s['raw0013'] == '0002'
    events = [fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_EVENT ')]
    assert [e['ida_linear'] for e in events] == ['19530', '196d2', '1020b', '10232', '10245']
    assert [e['raw4f46'] for e in events] == ['0800', '0800', '0000', '0000', '0000']
    assert [e['player_y'] for e in events] == ['17', '16', '16', '16', '15']
    assert all(e['player_x'] == '21' and e['raw258c'] == '0001' for e in events)
    first_record = next(fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_RECORD '))
    assert first_record['packet'] == '1' and first_record['record'] == '3079'
    assert int(events[2]['step']) < int(first_record['step']) < int(events[3]['step'])
    captures = []
    for label, expected in EXPECTED_IMAGES.items():
        p = root/(PREFIX+'-'+label+'.png')
        assert digest(p) == expected
        w, h, indices, _ = png(p)
        b = p.with_suffix('.bin')
        assert (w, h) == (640, 350) and indices == b.read_bytes()
        captures += [p, b]
    artifacts = [
        {'path': p.name, 'size': p.stat().st_size, 'sha256': digest(p)}
        for p in retained+captures+[meta_path, log_path, root/(PREFIX+'-probe-source.go'), root/(PREFIX+'-probe')]
    ]
    return {
        'scope': '正常母親返回後首三次Up；首次消費帶路殘留handler55，後兩步不重複',
        'original_sha256': parent['original_sha256'], 'parent_receipt_sha256': digest(parent_path),
        'normal_inputs': 41, 'actual_irq1_events': irq, 'events': events, 'states': states,
        'first_record': first_record, 'artifacts': artifacts, 'parent_174_png_bin_unchanged': True,
        'seed': '1357', 'seed_control_once': True, 'no_state_restore': True,
        'registration_parity': False, 'remake_parity': False, 'audio_parity': False,
        'original_save_load_parity': False,
    }


def validate(path):
    assert path.name == PREFIX+'-first-up-receipt.json'
    expected = source_receipt(path.parent)
    assert json.loads(path.read_text()) == expected
    return expected


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('receipt', type=Path)
    parser.add_argument('--write', action='store_true')
    args = parser.parse_args()
    if args.write:
        assert not args.receipt.exists()
        args.receipt.write_text(json.dumps(source_receipt(args.receipt.parent), ensure_ascii=False, indent=2)+'\n')
    validate(args.receipt)
    print('原版首三次Up／41正常輸入／82IRQ1／174父PNG與bin PASS；收據SHA-256 '+digest(args.receipt))


if __name__ == '__main__':
    main()
