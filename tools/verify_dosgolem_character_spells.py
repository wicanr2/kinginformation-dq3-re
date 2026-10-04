"""正常僧侶登錄、觀看咒文頁與返回來源稽核；入口docs/188。"""
from pathlib import Path
import hashlib, json, sys
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png


def validate(root, prefix, producer, full=False, close_scan='01'):
    assert close_scan in ('01', '1e') and (full or close_scan == '01')
    meta = json.loads((root / (prefix + '-meta.json')).read_text())
    exe = Path('/repo/assets_raw/DQ3.EXE')
    assert meta['original_size'] == exe.stat().st_size == 115282
    assert meta['original_sha256'] == digest(exe) == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
    assert meta['upstream_revision'] == '2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    assert meta['seed'] == '1357' and meta['seed_configured_before_execution']
    assert not meta['state_restore'] and not meta['gameplay_state_injection']
    assert meta['producer_sha256'] == digest(producer)
    assert meta['generator_sha256'] == digest(Path('/repo/tools/dosgolem_newgame_probe.py'))
    assert all(meta[k] == value for k, value in RUNTIME_HASHES.items())
    for suffix, key in (('-probe-source.go','probe_source_sha256'), ('-probe','probe_sha256')):
        assert digest(root / (prefix + suffix)) == meta[key]
    lines = (root / (prefix + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines and not any('找不到的檔（' in s for s in lines)
    assert len([s for s in lines if s.startswith('DQ3_CREATION_SEED ')]) == 1
    tags = {'queued': 'DQ3_QUIESCENT_QUEUED', 'consumed': 'DQ3_QUIESCENT_INPUT', 'states': 'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events': 'DQ3_KEY_DELIVERED'}
    groups = {key: [fields(s) for s in lines if s.startswith(tag + ' ')] for key, tag in tags.items()}
    q, c, states, irq = (groups[k] for k in tags)
    expected = 209 if full else 203
    assert len(q) == len(c) == len(states) == expected
    assert len(irq) == 76 + 2 * expected
    scans = [scan for _, scan in meta['normal_prefix_inputs']] + [int(row['scan'],16) for row in q]
    assert len(meta['normal_prefix_inputs']) == 38
    assert [int(s['port60'],16) for s in irq] == [x for scan in scans for x in (scan, scan | 128)]
    assert [int(s['count']) for s in irq] == list(range(1, len(irq) + 1))
    prior_path = root / 'issue4-recruit-view-detail-r1-source-r1-receipt.json'
    assert digest(prior_path) == 'e61060c7db7007c5a76e3790f4d55cc2c252619fd04c5cf572f615bc8fb69ee0'
    prior = json.loads(prior_path.read_text())
    for key in ('queued','consumed','states'):
        assert groups[key][:165] == prior[key][:165]
    assert irq[:406] == prior['actual_irq1_events'][:406]
    artifacts = []
    for n, (a, b, s) in enumerate(zip(q,c,states),1):
        assert a['packet'] == b['packet'] == s['packet'] == str(n)
        assert a['scan'] == b['scan'] == s['scan'] and a['kind'] == s['kind']
        assert int(a['step']) == int(s['queued_step']) < int(b['step']) < int(s['step'])
        assert int(b['irqs']) == 76 + n * 2 - 1
        assert int(irq[76 + n * 2 - 1]['step']) <= int(s['step'])
        assert int(s['raw0013'],16) & 0x4000 == 0
        assert len(bytes.fromhex(s['actor'])) == 128 and len(bytes.fromhex(s['flags'])) == 64
        if n > 1:
            assert a['step'] == states[n-2]['step'] and a['phase'] == states[n-2]['phase']
        name = prefix + f'-packet-{n:03d}-' + s['phase']
        width,height,indices,_ = png(root / (name + '.png'))
        assert (width,height) == (640,350) and indices == (root / (name + '.bin')).read_bytes()
        for ext in ('.png','.bin'):
            p = root / (name + ext)
            artifacts.append({'path':p.name,'size':p.stat().st_size,'sha256':digest(p)})
            if n <= 165:
                assert p.read_bytes() == (root / (p.name.replace(prefix,'issue4-recruit-view-detail-r1',1))).read_bytes()
        if n >= 150:
            assert all(s[k] == states[149][k] for k in ('actor','flags','gold_lo','gold_hi'))
    parent_path = root / 'issue4-mother-finish-receipt.json'
    assert digest(parent_path) == meta['parent_receipt_sha256'] == '9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    parent = json.loads(parent_path.read_text()); unchanged = 0
    for a in parent['artifacts']:
        if a['path'].endswith(('.png','.bin')):
            p = root / a['path'].replace('issue4-mother-finish-',prefix+'-',1)
            assert p.stat().st_size == a['size'] and digest(p) == a['sha256']; unchanged += 1
    assert unchanged == 174
    assert [s['choice_cursor'] for s in states[164:168]] == ['1','2','3','1']
    assert [s['phase'] for s in states[168:175]] == ['waiting','waiting','choice','choice','choice','waiting','ready']
    assert [s['kind'] for s in states[165:169]] == ['registry_birth_class_cursor']*2 + ['registry_birth_class','registry_birth_gender']
    writers = [fields(s) for s in lines if s.startswith('DQ3_BIRTH_WRITER ')]
    assert [s['ida_linear'] for s in writers] == ['10924','10a9f','10816','1081c']
    assert [s['packet'] for s in writers] == ['169','172','172','172']
    assert writers[1]['candidate'] == writers[2]['candidate'] == writers[3]['candidate']
    candidate = bytes.fromhex(writers[1]['candidate']); assert len(candidate) == 128
    assert candidate[:5] == bytes.fromhex('0003010100') and candidate[0x15] == 1
    assert candidate[0x3a:0x3c] == bytes.fromhex('1e80') and sum(candidate[0x2e:0x32]) > 0
    assert writers[1]['roster_flags'] == writers[2]['roster_flags'] == '020000000000000000000000'
    assert writers[3]['roster_flags'] == '020100000000000000000000'
    observations = [fields(s) for s in lines if s.startswith('DQ3_RECRUIT_STATE ')]
    assert len(observations) == expected - 171
    assert all(s['roster_flags'] == '020100000000000000000000' and s['slot1'] == candidate[:97].hex() for s in observations)
    for key in ('raw4f1f','raw4f15','raw4f17','raw4f19','raw4f1b'):
        assert len({s[key] for s in observations}) == 1
    detail = [fields(s) for s in lines if s.startswith('DQ3_VIEW_DETAIL ')]
    assert [s['ida_linear'] for s in detail] == (['10668','10671','1834e','1068e'] if full else ['10668','10671','1834e'])
    assert all(s['DS'] == '15ed' and s['raw4f1d'] == '520b' and s['actor520b'] == candidate[:97].hex() for s in detail)
    done = [fields(s) for s in lines if s.startswith('DQ3_RECRUIT_DONE ')]
    assert len(done) == 1 and done[0]['step'] == states[-1]['step'] and done[0]['packets'] == str(expected)
    assert done[0]['irqs'] == str(len(irq))
    assert states[-1]['ida_linear'] == ('1997c' if full else '21133')
    if full:
        assert [s['phase'] for s in states[202:]] == ['waiting','waiting','view_detail_wait','choice','choice','waiting','ready']
        assert [s['last_record'] for s in states[202:]] == ['528','528','528','540','540','541','541']
        assert [s['scan'] for s in q[202:]] == ['1c','1c','1c',close_scan,'4d','1c','1c']
    windows = [fields(s) for s in lines if s.startswith('DQ3_VIEW_SPELL_WINDOW ')]
    if full:
        parent = json.loads((root / 'issue4-view-spells-class3-normal-r1-source-r1-receipt.json').read_text())
        for key in ('queued', 'consumed', 'states'):
            assert groups[key][:203] == parent[key]
        assert irq[:482] == parent['actual_irq1_events']
        union = (bytes(40) + b'\x01' + bytes(19)).hex()
        window = '120313002e002c003000d7010100d801d901000000000000'
        consumers = [s for s in windows if s['ida_linear'] != '213c4']
        assert [s['ida_linear'] for s in consumers] == ['184c0', '18567', '18573', '185c1', '185d7', '185dc', '10681']
        assert [s['packet'] for s in consumers] == ['204'] * 5 + ['205'] * 2
        assert all(s['window3f34'] == window and s['union232d'] == union for s in consumers)
        name = consumers[3]
        assert (name['DI'], name['text_x'], name['text_y']) == ('00a1', '21', '62')
        frame = [s for s in windows if s['packet'] == '204' and s['ida_linear'] == '213c4' and s['SI'] == '3f34']
        assert [(s['DI'], s['text_x'], s['text_y']) for s in frame] == [('01d7','19','46'), ('01d8','19','62'), ('01d9','19','78'), ('00a1','21','62')]
        assert candidate[0x2e:0x32] == bytes.fromhex('01000100')
        for ext in ('.png', '.bin'):
            assert (root / (prefix + '-packet-203-waiting' + ext)).read_bytes() == (root / (prefix + '-packet-205-view_detail_wait' + ext)).read_bytes()
        for n in range(1, 204):
            phase = states[n - 1]['phase']
            for ext in ('.png', '.bin'):
                name = f'-packet-{n:03d}-{phase}' + ext
                assert (root / (prefix + name)).read_bytes() == (root / ('issue4-view-spells-class3-normal-r1' + name)).read_bytes()
    return {'original_size':115282,'original_sha256':meta['original_sha256'],'upstream_revision':meta['upstream_revision'],'seed':'1357','seed_control_once':True,'scope':'正常第三職業男性登錄、能力與咒文等待、樓下觀看與有限返回；不含其他角色或原版存讀檔' if full else '正常第三職業男性登錄與觀看第一能力等待；咒文後續尚未接受','meta':meta,'normal_inputs':38+expected,'artifacts':artifacts,**groups,'writer_events':writers,'recruit_observations':observations,'view_detail':detail,'spell_window_observations':windows,'candidate_raw_128':candidate.hex(),'prefix165_unchanged':True,'parent174_unchanged':True,'done':done[0],'full_rgb_parity':False,'remake_parity':False,'spec_ready':False,'original_save_load_parity':False,'checker_sha256':digest(Path(__file__)),'log_sha256':digest(root / (prefix + '.log'))}


if __name__ == '__main__':
    import argparse
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root',type=Path)
    parser.add_argument('prefix')
    parser.add_argument('producer',type=Path)
    parser.add_argument('receipt',type=Path)
    parser.add_argument('--first-ability',action='store_true')
    args=parser.parse_args()
    assert not args.receipt.exists(), 'refuse to overwrite receipt'
    report=validate(args.root,args.prefix,args.producer,not args.first_ability)
    with args.receipt.open('x') as stream:
        json.dump(report,stream,ensure_ascii=False,indent=2);stream.write('\n')
    print('正常僧侶與咒文來源接受',len(report['states']),len(report['actual_irq1_events']),digest(args.receipt))
