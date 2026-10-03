"""單一正常登錄出生的原版來源稽核；範圍與用法見docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png

PREFIX="issue4-registry-birth-normal-r2"
RECEIPT_HASH="de818064f9bcaf3f36dcd6a5d8b9259908781356a27c128cf1e8b41218c10ce3"

def source_receipt(root):
    prefix=PREFIX
    meta=json.loads((root/(prefix+'-meta.json')).read_text())
    assert meta['original_size']==115282
    assert meta['original_sha256']=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
    assert meta['upstream_revision']=='2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    assert meta['scenario']=='registry_birth_cold_draft' and meta['seed']=='1357'
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    assert meta['producer_sha256']==digest(Path(__file__).with_name('dosgolem_registry_birth_probe.py'))=='7cf6ec7a735863e0abdc07d225305846e51ea1150e793f90478981caf8f4310e'
    assert meta['probe_source_sha256']==digest(root/(prefix+'-probe-source.go'))=='a9bd3f16fee319c45b1780bbf35efd9995cb98b874dd60ef854ff260c5811baa'
    assert meta['probe_sha256']==digest(root/(prefix+'-probe'))=='a67ca4d91363308b80c4231aa3885926ad045c312664077accdbdae6ff55c6b4'
    assert meta['generator_sha256']==digest(Path('/repo/tools/dosgolem_newgame_probe.py'))
    assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
    log=root/(prefix+'.log')
    assert digest(log)=='7530db5727fb3673a88de69e4739c947fd674e804c0c70ee0940c6121f0021fa'
    lines=log.read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines and not any('找不到的檔（' in l for l in lines)
    assert len([l for l in lines if l.startswith('DQ3_CREATION_SEED ')])==1
    queues=[fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_QUEUED ')]
    consumed=[fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_INPUT ')]
    states=[fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_CAPTURE ')]
    irq=[fields(l) for l in lines if l.startswith('DQ3_KEY_DELIVERED ')]
    assert len(queues)==len(consumed)==len(states)==172 and len(irq)==420
    assert [int(i['count']) for i in irq]==list(range(1,421))
    scans=[s for _,s in meta['normal_prefix_inputs']]+[int(q['scan'],16) for q in queues]
    assert len(scans)==210 and [int(i['port60'],16) for i in irq]==[v for s in scans for v in (s,s|0x80)]
    for n,(q,c,s) in enumerate(zip(queues,consumed,states),1):
        assert int(q['packet'])==int(c['packet'])==int(s['packet'])==n
        assert q['scan']==c['scan']==s['scan'] and q['kind']==s['kind']
        assert int(q['step'])==int(s['queued_step'])<int(c['step'])<int(s['step'])
        assert int(c['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(s['step'])
        assert int(s['raw0013'],16)&0x4000==0
        assert len(bytes.fromhex(s['actor']))==128 and len(bytes.fromhex(s['flags']))==64
        if n>1:assert q['step']==states[n-2]['step'] and q['phase']==states[n-2]['phase']
    cancel_path=root/'issue4-registry-quiescent-r2-cancel-source-r1-receipt.json'
    assert digest(cancel_path)=='0467c01bbf8218a0c21581cf200cc0b491284db53c72749ddb3631b79b2cf40e'
    cancel=json.loads(cancel_path.read_text())
    assert states[:153]==cancel['states'][:153] and queues[:153]==cancel['queued'][:153]
    assert irq[:382]==cancel['actual_irq1_events'][:382]
    parent_path=root/'issue4-mother-finish-receipt.json'
    assert digest(parent_path)==meta['parent_receipt_sha256']=='9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    parent=json.loads(parent_path.read_text());unchanged=0
    for a in parent['artifacts']:
        if a['path'].endswith(('.png','.bin')):
            p=root/a['path'].replace('issue4-mother-finish-',prefix+'-',1)
            assert p.stat().st_size==a['size'] and digest(p)==a['sha256'];unchanged+=1
    assert unchanged==174
    artifacts=[]
    for s in states:
        label=prefix+f"-packet-{int(s['packet']):03d}-"+s['phase']
        image=root/(label+'.png');indexed=root/(label+'.bin')
        w,h,indices,_=png(image)
        assert (w,h)==(640,350) and indices==indexed.read_bytes() and len(indices)==224000
        for p in (image,indexed):artifacts.append({'path':p.name,'size':p.stat().st_size,'sha256':digest(p)})
        if int(s['packet'])<=153:
            for p in (image,indexed):
                prior=root/p.name.replace(prefix,'issue4-registry-quiescent-r2',1)
                assert digest(prior)==digest(p)
    assert [s['kind'] for s in states[153:165]]==['registry_birth_name']*12
    assert [s['phase'] for s in states[164:]]==['choice','choice','waiting','choice','choice','choice','waiting','ready']
    assert [s['choice_count'] for s in states[164:]]==['6','2','2','2','2','2','2','2']
    records=[fields(l) for l in lines if l.startswith('DQ3_QUIESCENT_RECORD ') and fields(l)['stage'] in ('3','4')]
    assert [r['record'] for r in records]==['550','554','559','560']
    writers=[fields(l) for l in lines if l.startswith('DQ3_BIRTH_WRITER ')]
    assert [w['ida_linear'] for w in writers]==['10924','10a9f','10816','1081c']
    assert [w['packet'] for w in writers]==['167','169','169','169']
    assert writers[1]['candidate']==writers[2]['candidate']==writers[3]['candidate']
    assert writers[1]['dx']==writers[2]['dx']==writers[3]['dx']=='0001'
    assert writers[1]['roster_flags']==writers[2]['roster_flags']=='020000000000000000000000'
    assert writers[3]['roster_flags']=='020100000000000000000000'
    candidate=bytes.fromhex(writers[1]['candidate'])
    assert len(candidate)==128 and candidate[:5]==bytes.fromhex('0001010100')
    assert candidate[0x15]==1 and candidate[0x3a:0x3c]==bytes.fromhex('1e80')
    assert writers[2]['si']==writers[3]['si']=='526c'
    done=[fields(l) for l in lines if l.startswith('DQ3_BIRTH_DONE ')]
    assert len(done)==1 and done[0]['packets']=='172' and done[0]['irqs']=='420' and done[0]['cancelled']=='false'
    assert done[0]['step']==states[-1]['step'] and states[-1]['phase']=='ready'
    for s in states[149:]:
        for k in ('actor','flags','gold_lo','gold_hi','player_x','player_y','raw0b24'):assert s[k]==states[149][k]
    report={'scope':'原版正常新遊戲至單一戰士男性登錄、接受能力、選否並返回行走；來源審查',
        'original_size':115282,'original_sha256':meta['original_sha256'],'upstream_revision':meta['upstream_revision'],
        'normal_inputs':210,'actual_irq1_events':irq,'seed':'1357','seed_control_once':True,'state_injection':False,
        'source_go_sha256':meta['probe_source_sha256'],'binary_sha256':meta['probe_sha256'],'log_sha256':digest(log),
        'meta_sha256':digest(root/(prefix+'-meta.json')),'producer_sha256':meta['producer_sha256'],
        'queued':queues,'consumed':consumed,'states':states,'records':records,'writer_events':writers,'artifacts':artifacts,
        'parent_174_png_bin_unchanged':True,'prior_153_states_png_bin_unchanged':True,
        'registration_slot':1,'registration_status_before':0,'registration_status_after':1,
        'candidate_raw_128':writers[1]['candidate'],'copy_97_bytes':'IDA9.4 linear10A9F..10ABD；DS520B→DS52FD+DX*61h；DX=1',
        'destination_record_bytes_observed':False,'remake_parity':False,'full_rgb_parity':False,
        'audio_parity':False,'original_save_load_parity':False,'other_class_gender_paths':'unknown'}
    return report

def validate(path):
    assert path.name==PREFIX+"-source-r1-receipt.json"
    assert digest(path)==RECEIPT_HASH
    data=json.loads(path.read_text())
    assert digest(path.with_name(PREFIX+"-meta.json"))==data["meta_sha256"]
    for a in data["artifacts"]:
        p=path.with_name(a["path"])
        assert p.stat().st_size==a["size"] and digest(p)==a["sha256"]
    assert source_receipt(path.parent)==data
    return data

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument("receipt",type=Path)
    parser.add_argument("--emit",action="store_true",help="重生固定來源，不覆寫既有收據")
    args=parser.parse_args()
    if args.emit:
        assert not args.receipt.exists()
        d=source_receipt(args.receipt.parent)
        text=json.dumps(d,ensure_ascii=False,indent=2)+"\n"
        import hashlib
        assert hashlib.sha256(text.encode()).hexdigest()==RECEIPT_HASH
        args.receipt.write_text(text)
    d=validate(args.receipt)
    print("原版正常單一出生來源 PASS",d["normal_inputs"],len(d["actual_irq1_events"]),len(d["artifacts"]))

if __name__=="__main__":
    main()
