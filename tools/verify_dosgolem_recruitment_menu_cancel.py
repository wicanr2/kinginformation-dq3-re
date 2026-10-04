"""正常招募主選單Esc來源稽核；有限範圍与重生入口docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_recruitment_empty_yes import validate as validate_yes

PREFIX='issue4-recruit-menu-esc-normal-r1'
PARENT='issue4-empty-join-yes-normal-r1'
PARENT_HASH='f56419c0113c6f895b71f9e287550c1c7d9e477e76172ddabc48038f3d3c3777'
GO_HASH='dcee6fe043459530c40a71453e202611c97e254649f174aecc8fa863afa6fb30'
BINARY_HASH='972b2dc313ddcff9a380ca466267ff7ee3993bf7e6ed50bf5abc073948d49613'


def validate(root, producer):
    parent_path=root/(PARENT+'-source-r1-receipt.json')
    assert digest(parent_path)==PARENT_HASH
    prior=json.loads(parent_path.read_text())
    assert prior==validate_yes(root,'join',Path(__file__).with_name('dosgolem_recruitment_empty_yes_probe.py'))
    meta=json.loads((root/(PREFIX+'-meta.json')).read_text())
    assert meta['original_size']==115282 and meta['original_sha256']==prior['original_sha256']
    assert meta['upstream_revision']=='2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    assert meta['seed']=='1357' and meta['seed_configured_before_execution'] is True
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    assert meta['producer_sha256']==digest(producer)
    assert meta['generator_sha256']==digest(Path(__file__).with_name('dosgolem_newgame_probe.py'))
    assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
    assert meta['docker_image']=='dq3-ebiten-test:20260822-r1' and meta['build_flags']==['-trimpath','-p','2']
    assert meta['probe_source_sha256']==digest(root/(PREFIX+'-probe-source.go'))==GO_HASH
    assert meta['probe_sha256']==digest(root/(PREFIX+'-probe'))==BINARY_HASH
    lines=(root/(PREFIX+'.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines and not any('找不到的檔（' in s for s in lines)
    seed=[s for s in lines if s.startswith('DQ3_CREATION_SEED ')];assert len(seed)==1 and 'fixed=1357' in seed[0]
    tags={'queued':'DQ3_QUIESCENT_QUEUED','consumed':'DQ3_QUIESCENT_INPUT','states':'DQ3_QUIESCENT_CAPTURE','actual_irq1_events':'DQ3_KEY_DELIVERED'}
    groups={k:[fields(s) for s in lines if s.startswith(tag+' ')] for k,tag in tags.items()}
    q,c,states,irq=(groups[k] for k in tags)
    assert len(q)==len(c)==len(states)==193 and len(irq)==462
    for key in ('queued','consumed','states'):assert groups[key][:191]==prior[key][:191]
    assert irq[:458]==prior['actual_irq1_events'][:458]
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
        label=PREFIX+f'-packet-{n:03d}-'+s['phase'];image,indexed=root/(label+'.png'),root/(label+'.bin')
        width,height,indices,_=png(image);assert (width,height)==(640,350) and indices==indexed.read_bytes()
        for path in (image,indexed):
            artifacts.append({'path':path.name,'size':path.stat().st_size,'sha256':digest(path)})
            if n<=191:assert path.read_bytes()==(root/path.name.replace(PREFIX,PARENT,1)).read_bytes()
    assert [s['scan'] for s in q[188:]]==['1c','1c','1c','01','1c']
    assert q[191]['kind']=='recruit_menu_cancel'
    assert [s['phase'] for s in states[188:]]==['inline_wait','choice','choice','waiting','ready']
    assert [s['last_record'] for s in states[188:]]==['316','540','528','541','541']
    routine=['103d7','103e1','103ec','103f4','1046a','1046d','10472','103ae']
    points=['10378','10384','10387','1038c']+routine+['103b9','103c0','10378','10384','10387','1038c','103c2','103d6']
    entries=[fields(s) for s in lines if s.startswith('DQ3_EMPTY_RECRUIT_ENTRY ')]
    assert [s['ida_linear'] for s in entries]==points
    assert all(s['DS']=='15ed' and s['raw5077']=='1' and s['raw5060']=='0' and s['roster_flags']=='020000000000000000000000' for s in entries)
    assert len({s['party_pointers'] for s in entries})==1 and len(bytes.fromhex(entries[0]['party_pointers']))==8
    cancel=[s for s in entries if s['packet']=='192']
    assert [s['ida_linear'] for s in cancel]==['10387','1038c','103c2'] and all(s['raw0726']=='1' for s in cancel)
    initial=[s for s in entries if s['ida_linear'] in ('10387','1038c') and s['packet']=='189']
    assert len(initial)==2 and all(s['raw0726']=='0' for s in initial)
    texts=[fields(s) for s in lines if s.startswith('DQ3_EMPTY_RECRUIT_TEXT ')]
    assert [(s['packet'],s['DI']) for s in texts]==[('187','020f'),('188','0210'),('189','0212'),('189','013c'),('190','021c'),('191','0210'),('192','021d')]
    assert all(s['text_segment']=='2826' for s in texts)
    assert all(s['player_x']=='2' and s['player_y']=='18' and s['raw0b24']=='000c' for s in states[186:])
    done=[fields(s) for s in lines if s.startswith('DQ3_QUIESCENT_DONE ')]
    assert len(done)==1 and done[0]['packets']=='193' and done[0]['irqs']=='462' and done[0]['step']==states[-1]['step']
    mother_path=root/'issue4-mother-finish-receipt.json'
    assert digest(mother_path)==meta['parent_receipt_sha256']=='9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    retained=0
    for a in json.loads(mother_path.read_text())['artifacts']:
        if a['path'].endswith(('.png','.bin')):
            path=root/a['path'].replace('issue4-mother-finish-',PREFIX+'-',1)
            assert path.stat().st_size==a['size'] and digest(path)==a['sha256'];retained+=1
    assert retained==174
    return {'scope':'正常空Join Yes返回招募主選單Esc、541獨立等待及field返回','original_size':meta['original_size'],'original_sha256':meta['original_sha256'],'upstream_revision':meta['upstream_revision'],'seed':meta['seed'],'meta':meta,'normal_inputs':231,**groups,'entries':entries,'text_events':texts,'artifacts':artifacts,'prefix191_unchanged':True,'parent174_unchanged':True,'seed_control_once':True,'state_injection':False,'done':done[0],'remake_parity':False,'full_rgb_parity':False,'original_save_load_parity':False,'audio_parity':False,'checker_sha256':digest(Path(__file__)),'log_sha256':digest(root/(PREFIX+'.log'))}


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root',type=Path);parser.add_argument('producer',type=Path);parser.add_argument('receipt',type=Path)
    args=parser.parse_args();assert not args.receipt.exists()
    result=validate(args.root,args.producer)
    args.receipt.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print('正常招募主選單Esc原版來源接受',digest(args.receipt))
