"""獨立稽核原版正常 K 改名交易；固定來源、IRQ1、全畫布與角色，入口 docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_mother_return import digest,fields,png

def validate(root,prefix,run,producer):
    assert run in (2,4,5)
    parent_path=root/'issue4-view-rename-normal-r1-source-r2-receipt.json'
    assert digest(parent_path)=='fe3df1165b4d7e27c1f5f2701ec59bfe3d078f9b20ea0f8ab9c4b2556d805f9c'
    parent=json.loads(parent_path.read_text());meta=json.loads((root/(prefix+'-meta.json')).read_text())
    for k,v in parent['meta'].items():
        if k in ('probe_source_sha256','probe_sha256','producer_sha256'):continue
        if k=='args':v=[a.replace('issue4-view-rename-normal-r1',prefix) for a in v]
        assert meta[k]==v,k
    assert meta['producer_sha256']==digest(producer)
    for suffix,key in [('-probe-source.go','probe_source_sha256'),('-probe','probe_sha256')]:assert digest(root/(prefix+suffix))==meta[key]
    lines=(root/(prefix+'.log')).read_text().splitlines()
    groups={k:[fields(x) for x in lines if x.startswith(tag+' ')] for k,tag in [('queued','DQ3_QUIESCENT_QUEUED'),('consumed','DQ3_QUIESCENT_INPUT'),('states','DQ3_QUIESCENT_CAPTURE'),('actual_irq1_events','DQ3_KEY_DELIVERED')]}
    q,c,s,irq=[groups[k] for k in ['queued','consumed','states','actual_irq1_events']]
    for k in ['queued','consumed','states']:assert groups[k][:202]==parent[k],k
    assert irq[:480]==parent['actual_irq1_events']
    name_scans={2:[0x48,0x4b,0x1c,0x1c,0x4d,0x50,0x1c,0x48,0x4b,0x1c,0x48,0x1c],5:[0x48,0x4b,0x1c,0x1c,0x4d,0x50,0x4d,0x1c,0x48,0x4b,0x4b,0x1c,0x48,0x1c],4:[0x48,0x4b,0x1c,0x50,0x50,0x50,0x1c]}[run]
    want=202+len(name_scans)+3
    assert len(q)==len(c)==len(s)==want and len(irq)==(38+want)*2
    assert [int(x['scan'],16) for x in q[202:]]==name_scans+[0x4d,0x1c,0x1c]
    assert [x['kind'] for x in q[202:]]==['recruit_rename_name']*len(name_scans)+['recruit_rename_decline_cursor','recruit_rename_decline','recruit_wait']
    assert [x['phase'] for x in s[-4:]]==['choice','choice','waiting','ready']
    assert [x['last_record'] for x in s[-4:]]==['540','540','541','541']
    assert [x['ida_linear'] for x in s[-4:]]==['1f7b7','1f7b7','21133','1997c']
    assert [x['choice_cursor'] for x in s[-4:]]==['1','2','2','2']
    scans=[scan for _,scan in meta['normal_prefix_inputs']]+[int(x['scan'],16) for x in q]
    assert [int(x['port60'],16) for x in irq]==[v for scan in scans for v in [scan,scan|128]]
    assert [int(x['count']) for x in irq]==list(range(1,len(irq)+1))
    parent_artifacts={x['path']:x for x in parent['artifacts']};artifacts=[]
    before=bytes.fromhex(s[201]['actor'])
    after=bytearray(before)
    if run==5:after[7:9]=b'\x05\x00' # 原始英數格raw1的字模為5，沿已驗證的rec453。
    for number,(a,b,d) in enumerate(zip(q,c,s),1):
        assert a['packet']==b['packet']==d['packet']==str(number)
        assert a['scan']==b['scan']==d['scan'] and a['kind']==d['kind']
        assert int(a['step'])==int(d['queued_step'])<int(b['step'])<int(d['step'])
        assert int(b['irqs'])==76+2*number-1 and int(irq[76+2*number-1]['step'])<=int(d['step'])
        assert int(d['raw0013'],16)&0x4000==0
        if number>1:assert a['step']==s[number-2]['step'] and a['phase']==s[number-2]['phase']
        name=prefix+f'-packet-{number:03d}-'+d['phase']
        w,h,indices,_=png(root/(name+'.png'))
        assert (w,h)==(640,350) and indices==(root/(name+'.bin')).read_bytes()
        for ext in ['.png','.bin']:
            path=root/(name+ext);artifacts.append({'path':path.name,'size':path.stat().st_size,'sha256':digest(path)})
            if number<=202:
                prior=root/('issue4-view-rename-normal-r1'+f'-packet-{number:03d}-'+d['phase']+ext)
                assert path.read_bytes()==prior.read_bytes() and digest(prior)==parent_artifacts[prior.name]['sha256']
        if number>=202:
            for k in ['player_x','player_y','raw0b24','gold_lo','gold_hi','flags']:assert d[k]==s[201][k]
            assert bytes.fromhex(d['actor'])==(bytes(after) if number>=202+len(name_scans) else before)
    obs=[fields(x) for x in lines if x.startswith('DQ3_RENAME_OBSERVE ')]
    assert obs[:3]==parent['rename_observations']
    tail=['106c2','1068e'] if run==4 else ['106c2','106cb','106d8','106da','1068e']
    assert [x['ida_linear'] for x in obs[3:]]==tail
    if run==4:assert obs[3]['cancel']=='1'
    assert all(x['DS']=='15ed' and x['pointers']=='507f50e2514551a8520b' and x['actor520b']==x['slot1']==obs[0]['slot1'] for x in obs)
    if run==4:assert all(x['hero']==obs[0]['hero'] for x in obs)
    else:assert all(x['hero']==obs[0]['hero'] for x in obs[:-2])
    if run==5:
        assert obs[-2]['hero']==obs[-1]['hero']==bytes(after[:97]).hex()
        assert obs[-3]['DI']=='5082' and obs[-2]['DI']=='5094'
    recruit=[fields(x) for x in lines if x.startswith('DQ3_RECRUIT_STATE ')]
    assert len(recruit)==want-171
    assert all(x['slot1']==obs[0]['slot1'] and x['roster_flags']=='020100000000000000000000' for x in recruit)
    done=[fields(x) for x in lines if x.startswith('DQ3_RECRUIT_DONE ')]
    assert len(done)==1 and done[0]['step']==s[-1]['step'] and done[0]['packets']==str(want) and done[0]['irqs']==str(len(irq))
    report={'scope':{2:'相同姓名成功，正常No返回正對照',5:'不同姓名成功，主角單字0→5，正常No返回',4:'姓名功能列取消，不寫角色，正常No返回'}[run],'meta':meta,'parent_sha256':digest(parent_path),'log_sha256':digest(root/(prefix+'.log')),'prefix202_unchanged':True,'normal_inputs':want+38,'artifacts':artifacts,**groups,'rename_observations':obs,'recruit_observations':recruit,'done':done[0],'hero_before':before.hex(),'hero_after':bytes(after).hex(),'roster_unchanged':True,'spec_ready':False,'remake_parity':False,'original_save_load_parity':False}
    report['scope']={2:'相同姓名成功，正常No返回正對照',5:'不同姓名成功，主角單字0→5，正常No返回',4:'姓名功能列取消，不寫角色，正常No返回'}[run]
    report['original_sha256']=meta['original_sha256']
    report['upstream_revision']=meta['upstream_revision']
    report['seed']=meta['seed']
    report['validator_sha256']=digest(Path(__file__))
    return report

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('root',type=Path)
    p.add_argument('prefix')
    p.add_argument('route',type=int,choices=(2,4,5))
    p.add_argument('producer',type=Path)
    p.add_argument('receipt',type=Path)
    a=p.parse_args()
    assert not a.receipt.exists(),'refuse to overwrite accepted evidence'
    report=validate(a.root,a.prefix,a.route,a.producer)
    with a.receipt.open('x') as f:json.dump(report,f,ensure_ascii=False,indent=2);f.write('\n')
    print('原版正常改名來源 PASS',a.route,len(report['states']),len(report['artifacts']),digest(a.receipt))

if __name__=='__main__':main()
