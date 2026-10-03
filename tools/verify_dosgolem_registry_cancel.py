"""登錄所正常取消來源的可重現稽核；用法與範圍見 docs/188。"""
import argparse
import collections
import json
from pathlib import Path
from verify_dosgolem_first_move import PREFIX, source_receipt as first_move_source_receipt
from verify_dosgolem_mother_return import digest, fields, png

RECEIPT_HASH = "0467c01bbf8218a0c21581cf200cc0b491284db53c72749ddb3631b79b2cf40e"

def source_receipt(root):
    prior = first_move_source_receipt(root)
    meta = json.loads((root/(PREFIX+'-meta.json')).read_text())
    lines = (root/(PREFIX+'.log')).read_text().splitlines()
    queues = [fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_QUEUED ')]
    inputs = [fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_INPUT ')]
    states = [fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_CAPTURE ')]
    irq = [fields(l) for l in lines if l.startswith('DQ3_KEY_DELIVERED ')]
    assert len(queues) == len(inputs) == len(states) == 164
    assert len(irq) == 404
    assert [int(r['count']) for r in irq] == list(range(1,405))
    assert len(meta['normal_prefix_inputs']) == 38
    scans = [scan for _, scan in meta['normal_prefix_inputs']]+[int(q['scan'],16) for q in queues]
    assert [int(r['port60'],16) for r in irq] == [v for scan in scans for v in (scan,scan|0x80)]
    for ordinal, (q,c,s) in enumerate(zip(queues,inputs,states),1):
        assert int(q['packet']) == int(c['packet']) == int(s['packet']) == ordinal
        assert q['scan'] == c['scan'] == s['scan']
        expected_stage = 0 if ordinal <= 43 else 1 if ordinal <= 52 else 2 if ordinal <= 137 else 3 if ordinal <= 150 else 4
        assert int(q['stage']) == int(c['stage']) == expected_stage
        assert int(s['stage']) == (1 if ordinal == 43 else expected_stage)
        assert q['kind'] == s['kind']
        assert int(q['step']) == int(s['queued_step']) < int(c['step']) < int(s['step'])
        assert int(c['irqs']) == 76+ordinal*2-1
        # Capture is gated on the observed keyup count in the pinned Go producer.
        assert int(irq[76+ordinal*2-1]['step']) <= int(s['step'])
        assert int(s['raw0013'],16)&0x4000 == 0
        assert len(bytes.fromhex(s['actor'])) == 128 and len(bytes.fromhex(s['flags'])) == 64
        assert s['phase'] in ('ready','inline_wait','waiting','choice','name')
        assert s['ida_linear'] == {'ready':'1997c','inline_wait':'216d8','waiting':'21133','choice':'1f7b7','name':'11096'}[s['phase']]
        if ordinal > 1:
            before = states[ordinal-2]
            assert q['phase'] == before['phase'] and int(q['step']) == int(before['step'])
            assert q['ida_linear'] == before['ida_linear']
    assert collections.Counter(s['kind'] for s in states) == {
        'king_up':43,'king_wait':9,'return_motion':85,'registry_approach':13,
        'registry_wait':4,'registry_accept':1,'registry_name':3,
        'registry_cancel_cursor':3,'registry_cancel':1,'registry_decline_cursor':1,'registry_decline':1}
    assert [int(q['scan'],16) for q in queues[137:150]] == [0x4b]*3+[0x50]*4+[0x4b]*2+[0x48,0x4b,0x48,0x1c]
    assert [int(q['scan'],16) for q in queues[150:]] == [0x1c,0x1c,0x1c,0x48,0x4b,0x1c,0x50,0x50,0x50,0x1c,0x1c,0x4d,0x1c,0x1c]
    assert states[51]['gold_lo'] == '0032' and states[51]['gold_hi'] == '0000'
    assert states[51]['player_x'] == '9' and states[51]['player_y'] == '7'
    assert states[136]['player_x'] == '8' and states[136]['player_y'] == '2'
    assert states[136]['return_ordinal'] == '85'
    registry = states[149:]
    assert [s['phase'] for s in registry] == ['inline_wait','inline_wait','choice','name','name','name','choice','choice','choice','choice','inline_wait','choice','choice','waiting','ready']
    assert [s['last_record'] for s in registry] == ['550']*3+['554']*7+['558']*3+['560']*2
    assert [s['choice_count'] for s in registry[6:10]] == ['5']*4
    assert [s['choice_cursor'] for s in registry[6:10]] == ['1','2','3','4']
    assert [s['name_cursor'] for s in registry[3:6]] == ['0','36','35']
    for s in registry:
        for key in ('player_x','player_y','raw0b24','gold_lo','gold_hi','actor','flags'):
            assert s[key] == registry[0][key], key
    assert registry[-1]['player_x'] == '2' and registry[-1]['player_y'] == '5'
    records = [fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_RECORD ') and fields(l)['stage'] in ('3','4')]
    assert [r['record'] for r in records] == ['550','554','558','560']
    done = [fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_DONE ')]
    assert len(done) == 1 and done[0]['cancelled'] == 'true'
    assert done[0]['packets'] == '164' and done[0]['irqs'] == '404'
    assert done[0]['step'] == states[-1]['step']
    artifacts = []
    for s in states:
        label = PREFIX+f"-packet-{int(s['packet']):03d}-"+s['phase']
        image = root/(label+'.png'); indexed = root/(label+'.bin')
        width,height,indices,_ = png(image)
        assert (width,height) == (640,350) and len(indexed.read_bytes()) == 640*350
        assert indices == indexed.read_bytes()
        for p in (image,indexed):
            artifacts.append({'path':p.name,'size':p.stat().st_size,'sha256':digest(p)})
    report = {'scope':'原版正常新遊戲至登錄所姓名取消、選否並返回行走；來源审查，不是remake parity',
        'original_sha256':meta['original_sha256'],'upstream_revision':meta['upstream_revision'],
        'normal_inputs':202,'seed':meta['seed'],'seed_control_once':True,'state_injection':False,
        'first_move_receipt_sha256':digest(root/(PREFIX+'-first-up-receipt.json')),
        'source_go_sha256':meta['probe_source_sha256'],'binary_sha256':meta['probe_sha256'],
        'log_sha256':digest(root/(PREFIX+'.log')),'meta_sha256':digest(root/(PREFIX+'-meta.json')),
        'queued':queues,'consumed':inputs,'states':states,'registry_records':records,
        'actual_irq1_events':irq,'artifacts':artifacts,'parent_174_png_bin_unchanged':True,
        'original_member_creation':False,'original_roster_invariant':'unknown',
        'audio_parity':False,'original_save_load_parity':False,'remake_parity':False,'full_rgb_parity':False}
    report['scope'] = report['scope'].replace('审查','審查')
    return report

def validate(path):
    assert path.name == PREFIX+"-cancel-source-r1-receipt.json"
    assert digest(path) == RECEIPT_HASH
    data = json.loads(path.read_text())
    assert digest(path.with_name(PREFIX+'-meta.json')) == data['meta_sha256']
    for record in data['artifacts']:
        p = path.with_name(record['path'])
        assert p.stat().st_size == record['size'] and digest(p) == record['sha256']
    expected = source_receipt(path.parent)
    assert data == expected
    return expected

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("receipt", type=Path)
    parser.add_argument("--emit", action="store_true", help="由固定原版來源重生，不覆寫既有收據")
    args = parser.parse_args()
    if args.emit:
        assert not args.receipt.exists()
        report = source_receipt(args.receipt.parent)
        text = json.dumps(report, ensure_ascii=False, indent=2)+"\n"
        import hashlib
        assert hashlib.sha256(text.encode()).hexdigest() == RECEIPT_HASH
        args.receipt.write_text(text)
    d = validate(args.receipt)
    print("原版正常取消來源 PASS",d["normal_inputs"],len(d["actual_irq1_events"]),len(d["artifacts"]))

if __name__ == "__main__":
    main()
