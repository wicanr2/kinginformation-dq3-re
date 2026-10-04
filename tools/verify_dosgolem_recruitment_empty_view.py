"""正常空名冊觀看來源稽核；Docker用法與有限範圍见docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_registry_cancel import validate as validate_cancel
from verify_dosgolem_mother_return import digest, fields, png


def validate(root, prefix, producer):
    old_prefix = 'issue4-registry-quiescent-r2'
    prior = validate_cancel(root / (old_prefix + '-cancel-source-r1-receipt.json'))
    meta = json.loads((root / (prefix + '-meta.json')).read_text())
    assert meta['original_size'] == 115282
    assert meta['original_sha256'] == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
    assert meta['upstream_revision'] == '2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    assert meta['seed'] == '1357' and meta['seed_configured_before_execution'] is True
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['producer_sha256'] == digest(producer)
    assert meta['generator_sha256'] == digest(Path(__file__).with_name('dosgolem_newgame_probe.py'))
    assert meta['docker_image'] == 'dq3-ebiten-test:20260822-r1'
    assert meta['build_flags'] == ['-trimpath', '-p', '2']
    assert all(meta[k] == value for k, value in RUNTIME_HASHES.items())
    assert digest(root / (prefix + '-probe-source.go')) == meta['probe_source_sha256'] == '899c1bb38d7e3cf5723ffb3c3af2e0fb25179000da8c066ff5bc0b82d71465fc'
    assert digest(root / (prefix + '-probe')) == meta['probe_sha256'] == 'eeb0737c9a3de5e280c65dbb969b04147a8de6a0f75ec6baa991ad5a424bf83d'
    lines = (root / (prefix + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line for line in lines)
    seeds = [line for line in lines if line.startswith('DQ3_CREATION_SEED ')]
    assert len(seeds) == 1 and 'fixed=1357' in seeds[0]
    tags = {'queued':'DQ3_QUIESCENT_QUEUED', 'consumed':'DQ3_QUIESCENT_INPUT',
            'states':'DQ3_QUIESCENT_CAPTURE', 'actual_irq1_events':'DQ3_KEY_DELIVERED'}
    groups = {key:[fields(line) for line in lines if line.startswith(tag + ' ')] for key,tag in tags.items()}
    q,c,states,irq = (groups[key] for key in tags)
    assert len(q) == len(c) == len(states) == 195 and len(irq) == 466
    for key in ('queued','consumed','states'):
        assert groups[key][:164] == prior[key]
    assert irq[:404] == prior['actual_irq1_events']
    assert len(meta['normal_prefix_inputs']) == 38
    scans = [scan for _,scan in meta['normal_prefix_inputs']] + [int(row['scan'],16) for row in q]
    assert [int(row['port60'],16) for row in irq] == [v for scan in scans for v in (scan,scan|128)]
    assert [int(row['count']) for row in irq] == list(range(1,467))
    artifacts = []
    phases = {'ready':'1997c','inline_wait':'216d8','waiting':'21133','choice':'1f7b7','name':'11096'}
    for n,(a,b,s) in enumerate(zip(q,c,states),1):
        assert a['packet'] == b['packet'] == s['packet'] == str(n)
        assert a['scan'] == b['scan'] == s['scan'] and a['kind'] == s['kind']
        assert int(a['step']) == int(s['queued_step']) < int(b['step']) < int(s['step'])
        assert int(b['irqs']) == 76 + n*2 - 1
        assert int(irq[76+n*2-1]['step']) <= int(s['step'])
        assert s['ida_linear'] == phases[s['phase']] and int(s['raw0013'],16)&0x4000 == 0
        assert len(bytes.fromhex(s['actor'])) == 128 and len(bytes.fromhex(s['flags'])) == 64
        if n > 1:
            before = states[n-2]
            assert a['step'] == before['step'] and a['phase'] == before['phase'] and a['ida_linear'] == before['ida_linear']
        if n >= 150:
            assert all(s[k] == states[149][k] for k in ('actor','flags','gold_lo','gold_hi'))
        label = prefix + f'-packet-{n:03d}-' + s['phase']
        image,indexed = root/(label+'.png'),root/(label+'.bin')
        width,height,indices,_ = png(image)
        assert (width,height) == (640,350) and indices == indexed.read_bytes()
        for path in (image,indexed):
            artifacts.append({'path':path.name,'size':path.stat().st_size,'sha256':digest(path)})
            if n <= 164:
                assert path.read_bytes() == (root/path.name.replace(prefix,old_prefix,1)).read_bytes()
    assert [a['scan'] for a in q[164:187]] == ['4d','50','4d','4d']+['48']*4+['4d']*3+['50']*4+['4b']*6+['48','1c']
    assert [a['scan'] for a in q[187:]] == ['1c','50','50','1c','1c','4d','1c','1c']
    assert [s['phase'] for s in states[186:]] == ['inline_wait','choice','choice','choice','inline_wait','choice','choice','waiting','ready']
    assert [s['last_record'] for s in states[186:]] == ['527','528','528','528','316','540','540','541','541']
    assert all(s['player_x'] == '2' and s['player_y'] == '18' and s['raw0b24'] == '000c' for s in states[186:])
    assert [s['choice_cursor'] for s in states[187:190]] == ['1','2','3']
    entries = [fields(line) for line in lines if line.startswith('DQ3_EMPTY_VIEW_ENTRY ')]
    assert [s['ida_linear'] for s in entries] == ['10624','10627','1062f','10696','10699','1069e']
    assert [s['packet'] for s in entries] == ['191']*5+['192']
    assert all(s['DS'] == '15ed' and s['raw5060'] == '0' and s['roster_flags'] == '020000000000000000000000' for s in entries)
    assert entries[2]['AX'] == '0000' and entries[4]['DI'] == '013c'
    texts = [fields(line) for line in lines if line.startswith('DQ3_EMPTY_VIEW_TEXT ')]
    assert [(s['packet'],s['DI']) for s in texts] == [('187','020f'),('188','0210'),('191','013c'),('192','021c'),('194','021d')]
    assert all(s['text_segment'] == '2826' for s in texts)
    assert int(entries[4]['step']) < int(texts[2]['step']) < int(states[190]['step']) < int(entries[5]['step']) < int(texts[3]['step']) < int(states[191]['step'])
    done = [fields(line) for line in lines if line.startswith('DQ3_QUIESCENT_DONE ')]
    assert len(done) == 1 and done[0]['packets'] == '195' and done[0]['irqs'] == '466'
    assert done[0]['step'] == states[-1]['step'] and done[0]['cancelled'] == 'true'
    parent_path = root/'issue4-mother-finish-receipt.json'
    assert digest(parent_path) == meta['parent_receipt_sha256'] == '9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    retained = 0
    for a in json.loads(parent_path.read_text())['artifacts']:
        if a['path'].endswith(('.png','.bin')):
            path = root/a['path'].replace('issue4-mother-finish-',prefix+'-',1)
            assert path.stat().st_size == a['size'] and digest(path) == a['sha256']
            retained += 1
    assert retained == 174
    return {'scope':'正常姓名取消後觀看空名冊、record316等待、540 No、541獨立等待、返回行走',
            'original_size':meta['original_size'],'original_sha256':meta['original_sha256'],
            'upstream_revision':meta['upstream_revision'], 'seed':meta['seed'],
            'meta':meta,'normal_inputs':233,**groups,'empty_view_entries':entries,'text_events':texts,
            'artifacts':artifacts,'prefix164_unchanged':True,'parent174_unchanged':True,
            'seed_control_once':True,'state_injection':False,'original_roster_invariant':'empty at View entry and return',
            'done':done[0],'remake_parity':False,'full_rgb_parity':False,'original_save_load_parity':False,
            'audio_parity':False,'checker_sha256':digest(Path(__file__)),'log_sha256':digest(root/(prefix+'.log'))}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root',type=Path)
    parser.add_argument('prefix')
    parser.add_argument('producer',type=Path)
    parser.add_argument('receipt',type=Path)
    args = parser.parse_args()
    assert not args.receipt.exists(), 'refuse to overwrite receipt'
    report = validate(args.root,args.prefix,args.producer)
    with args.receipt.open('x') as stream:
        json.dump(report,stream,ensure_ascii=False,indent=2)
        stream.write('\n')
    print('正常空名冊來源接受',len(report['states']),len(report['actual_irq1_events']),digest(args.receipt))
