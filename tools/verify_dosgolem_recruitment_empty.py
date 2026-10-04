"""空Join／單人Leave的正常原版來源稽核；有界入口docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_recruitment_empty_view import validate as validate_view

IDENTITIES={
    'join':('issue4-empty-join-normal-r1','337e74015ede250277588cee0752bcc739fc9bd3a1ac50d5701ba0076103c11c','feff72522992b82988b5c5ca6898e09a05fd5c7b52f1a7b709e6568f6006dd7b'),
    'leave':('issue4-empty-leave-normal-r1','9c01f741b551bc786a2f0a7d87e82ff26bebfc3a4b8dcc20f24e8c57bb27181d','ec2d82e755e82757df95ea9bbb2b09d51c22a7478e242f5eaa2e395e8bac249a')}


def validate(root, action, producer):
    assert action in IDENTITIES
    prefix,go_hash,binary_hash=IDENTITIES[action]
    old_prefix='issue4-empty-view-normal-r1'
    prior_path=root/(old_prefix+'-source-r2-receipt.json')
    assert digest(prior_path)=='ff7a8abd5e93c867e5650f08a3af607feecf7013d7a781ff9a9ad2fd44eaeba7'
    prior=json.loads(prior_path.read_text())
    assert prior==validate_view(root,old_prefix,Path(__file__).with_name('dosgolem_recruitment_empty_view_probe.py'))
    meta=json.loads((root/(prefix+'-meta.json')).read_text())
    assert meta['original_size']==115282 and meta['original_sha256']==prior['original_sha256']
    assert meta['upstream_revision']=='2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    assert meta['seed']=='1357' and meta['seed_configured_before_execution'] is True
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['producer_sha256']==digest(producer)
    assert meta['generator_sha256']==digest(Path(__file__).with_name('dosgolem_newgame_probe.py'))
    assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
    assert meta['docker_image']=='dq3-ebiten-test:20260822-r1' and meta['build_flags']==['-trimpath','-p','2']
    assert meta['probe_source_sha256']==digest(root/(prefix+'-probe-source.go'))==go_hash
    assert meta['probe_sha256']==digest(root/(prefix+'-probe'))==binary_hash
    lines=(root/(prefix+'.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines and not any('找不到的檔（' in s for s in lines)
    seed=[s for s in lines if s.startswith('DQ3_CREATION_SEED ')]
    assert len(seed)==1 and 'fixed=1357' in seed[0]
    tags={'queued':'DQ3_QUIESCENT_QUEUED','consumed':'DQ3_QUIESCENT_INPUT','states':'DQ3_QUIESCENT_CAPTURE','actual_irq1_events':'DQ3_KEY_DELIVERED'}
    groups={k:[fields(s) for s in lines if s.startswith(tag+' ')] for k,tag in tags.items()}
    q,c,states,irq=(groups[k] for k in tags)
    assert len(q)==len(c)==len(states)==193 and len(irq)==462
    for key in ('queued','consumed','states'):assert groups[key][:188]==prior[key][:188]
    assert irq[:452]==prior['actual_irq1_events'][:452]
    assert len(meta['normal_prefix_inputs'])==38
    scans=[scan for _,scan in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in q]
    assert [int(s['port60'],16) for s in irq]==[v for scan in scans for v in (scan,scan|128)]
    assert [int(s['count']) for s in irq]==list(range(1,463))
    artifacts=[];phases={'ready':'1997c','inline_wait':'216d8','waiting':'21133','choice':'1f7b7','name':'11096'}
    for n,(a,b,s) in enumerate(zip(q,c,states),1):
        assert a['packet']==b['packet']==s['packet']==str(n)
        assert a['scan']==b['scan']==s['scan'] and a['kind']==s['kind']
        assert int(a['step'])==int(s['queued_step'])<int(b['step'])<int(s['step'])
        assert int(b['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(s['step'])
        assert s['ida_linear']==phases[s['phase']] and int(s['raw0013'],16)&0x4000==0
        assert len(bytes.fromhex(s['actor']))==128 and len(bytes.fromhex(s['flags']))==64
        if n>1:
            before=states[n-2];assert a['step']==before['step'] and a['phase']==before['phase'] and a['ida_linear']==before['ida_linear']
        if n>=150:assert all(s[k]==states[149][k] for k in ('actor','flags','gold_lo','gold_hi'))
        label=prefix+f'-packet-{n:03d}-'+s['phase'];image,indexed=root/(label+'.png'),root/(label+'.bin')
        w,h,indices,_=png(image);assert (w,h)==(640,350) and indices==indexed.read_bytes()
        for path in (image,indexed):
            artifacts.append({'path':path.name,'size':path.stat().st_size,'sha256':digest(path)})
            if n<=188:assert path.read_bytes()==(root/path.name.replace(prefix,old_prefix,1)).read_bytes()
    if action=='join':
        assert [s['scan'] for s in q[188:]]==['1c','1c','4d','1c','1c']
        assert [s['phase'] for s in states[188:]]==['inline_wait','choice','choice','waiting','ready']
        assert [s['last_record'] for s in states[188:]]==['316','540','540','541','541']
        points=['103d7','103e1','103ec','103f4','1046a','1046d','10472','103ae','103c2','103d6']
        texts_expected=[('187','020f'),('188','0210'),('189','0212'),('189','013c'),('190','021c'),('192','021d')]
    else:
        assert [s['scan'] for s in q[188:]]==['50','1c','4d','1c','1c']
        assert [s['phase'] for s in states[188:]]==['choice','choice','choice','waiting','ready']
        assert [s['last_record'] for s in states[188:]]==['528','540','540','541','541']
        points=['104c4','104c8','104cd','105bc','105bf','105c4','103ae','103c2','103d6']
        texts_expected=[('187','020f'),('188','0210'),('190','021e'),('190','021c'),('192','021d')]
    entries=[fields(s) for s in lines if s.startswith('DQ3_EMPTY_RECRUIT_ENTRY ')]
    assert [s['ida_linear'] for s in entries]==points
    assert all(s['DS']=='15ed' and s['raw5077']=='1' and s['roster_flags']=='020000000000000000000000' for s in entries)
    assert len({s['party_pointers'] for s in entries})==1 and len(bytes.fromhex(entries[0]['party_pointers']))==8
    if action=='join':assert all(s['raw5060']=='0' for s in entries)
    texts=[fields(s) for s in lines if s.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert [(s['packet'],s['DI']) for s in texts]==texts_expected
    assert all(s['text_segment']=='2826' for s in texts)
    assert all(s['player_x']=='2' and s['player_y']=='18' and s['raw0b24']=='000c' for s in states[186:])
    done=[fields(s) for s in lines if s.startswith('DQ3_QUIESCENT_DONE ')]
    assert len(done)==1 and done[0]['packets']=='193' and done[0]['irqs']=='462' and done[0]['step']==states[-1]['step']
    parent_path=root/'issue4-mother-finish-receipt.json';assert digest(parent_path)==meta['parent_receipt_sha256']=='9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    retained=0
    for a in json.loads(parent_path.read_text())['artifacts']:
        if a['path'].endswith(('.png','.bin')):
            path=root/a['path'].replace('issue4-mother-finish-',prefix+'-',1);assert path.stat().st_size==a['size'] and digest(path)==a['sha256'];retained+=1
    assert retained==174
    return {'scope':'正常空Join／單人Leave有限返回','action':action,'original_size':meta['original_size'],'original_sha256':meta['original_sha256'],'upstream_revision':meta['upstream_revision'],'seed':meta['seed'],'meta':meta,'normal_inputs':231,**groups,'entries':entries,'text_events':texts,'artifacts':artifacts,'prefix188_unchanged':True,'parent174_unchanged':True,'seed_control_once':True,'state_injection':False,'roster_and_party_invariant':'only hero at branch entry and caller return; pointers stable','done':done[0],'remake_parity':False,'full_rgb_parity':False,'original_save_load_parity':False,'audio_parity':False,'checker_sha256':digest(Path(__file__)),'log_sha256':digest(root/(prefix+'.log'))}


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root',type=Path);parser.add_argument('action',choices=tuple(IDENTITIES));parser.add_argument('producer',type=Path);parser.add_argument('receipt',type=Path)
    args=parser.parse_args();assert not args.receipt.exists(),'refuse to overwrite receipt'
    report=validate(args.root,args.action,args.producer)
    with args.receipt.open('x') as stream:json.dump(report,stream,ensure_ascii=False,indent=2);stream.write('\n')
    print('正常空清單來源接受',args.action,len(report['states']),len(report['actual_irq1_events']),digest(args.receipt))
