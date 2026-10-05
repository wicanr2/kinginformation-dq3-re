"""正常381同部位甲胄換穿來源checker，限定範圍與容器契約見docs/188。"""
from pathlib import Path
import ast, hashlib, json, struct, sys
sys.path.insert(0,'/repo/tools')
from verify_dosgolem_mother_return import digest,fields,png
from verify_dosgolem_first_move import RUNTIME_HASHES

def validate(root=Path('/work/dosgolem-opening'),producer=Path('/repo/tools/dosgolem_field_equipment_armor_replace_probe.py')):
    prefix='issue4-equip-armor-replace-normal-r1';parent_prefix='issue4-equip-armor-normal-r1'
    parent_path=root/(parent_prefix+'-source-r1-receipt.json')
    assert digest(parent_path)=='377d6d3088cd1d1a71eb789ef67579b05656428bbd92bf4130cfdfeb2ea457d9'
    parent=json.loads(parent_path.read_text());meta=json.loads((root/(prefix+'-meta.json')).read_text())
    for k in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','build_flags','docker_image'):assert meta[k]==parent['meta'][k],k
    assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    assert digest(producer)==meta['producer_sha256']
    for suffix,key in [('-probe-source.go','probe_source_sha256'),('-probe','probe_sha256')]:assert digest(root/(prefix+suffix))==meta[key]
    exe=Path('/repo/assets_raw/DQ3.EXE');assert exe.stat().st_size==meta['original_size'] and digest(exe)==meta['original_sha256']
    lines=(root/(prefix+'.log')).read_text().splitlines();assert '沒實作的服務（0 種）：' in lines
    assert not any(s.startswith('DQ3_SAVE_BOUND ') or '找不到的檔（' in s for s in lines)
    seed=[fields(s) for s in lines if s.startswith('DQ3_CREATION_SEED ')];assert len(seed)==1 and seed[0]['fixed']=='1357'
    tree=ast.parse(Path('/repo/tools/verify_dosgolem_field_equipment_wear.py').read_text())
    tags=next(ast.literal_eval(n.value) for n in tree.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='tags' for t in n.targets))
    tags['wear_observations']='DQ3_FIELD_EQUIP_WEAR_OBSERVE';tags['native_events']='DQ3_ARMOR_NATIVE'
    groups={k:[fields(s) for s in lines if s.startswith(t+' ')] for k,t in tags.items()}
    for k,rows in groups.items():assert rows[:len(parent[k])]==parent[k],k
    q,c,states,irq=(groups[k] for k in ('queued','consumed','states','actual_irq1_events'))
    assert len(q)==len(c)==len(states)==381 and len(irq)==838
    scans=[v for _,v in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in q]
    assert len(scans)==419 and [int(s['port60'],16) for s in irq]==[v for scan in scans for v in (scan,scan|128)]
    assert [int(s['count']) for s in irq]==list(range(1,839))
    assert [s['scan'] for s in q[366:]]==['39','50','50','39','01','39','01','01','39','50','39','39','1c','4b','4d']
    assert [s['phase'] for s in states[366:]]==['choice']*7+['ready']+['choice']*3+['waiting']+['ready']*3
    choices=[(6,1),(6,2),(6,3),(3,1),(4,1),(1,1),(1,1),(6,3),(6,1),(6,2),(3,1),(3,1),(6,2),(6,2),(6,2)]
    assert [(int(s['choice_count']),int(s['choice_cursor'])) for s in states[366:]]==choices
    artifacts=[]
    for a in parent['artifacts']:
        relative=a['path'].replace(parent_prefix,prefix,1);p=(root if '/' not in relative else root.parent)/relative
        assert p.stat().st_size==a['size'] and digest(p)==a['sha256'],relative
        artifacts.append(dict(path=relative,size=a['size'],sha256=a['sha256']))
    base=(root/(parent_prefix+'-persistent-366.bin')).read_bytes()
    for n,(queued,taken,state) in enumerate(zip(q,c,states),1):
        assert queued['packet']==taken['packet']==state['packet']==str(n)
        assert queued['scan']==taken['scan']==state['scan'] and queued['kind']==state['kind']
        assert int(queued['step'])==int(state['queued_step'])<int(taken['step'])<int(state['step'])
        assert int(taken['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
        if n>1:assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
        if n<=366:continue
        expected=bytearray(base)
        if n>=372:
            struct.pack_into('<H',expected,342+0x42,0x801f)
            struct.pack_into('<H',expected,342+0x44,0x001f)
        if n==380:expected[10]=2
        persistent=root/(prefix+f'-persistent-{n:03d}.bin');assert persistent.read_bytes()==expected,n
        assert bytes.fromhex(state['actor'])==expected[342:470] and state['flags']==states[365]['flags']
        label=prefix+f'-packet-{n:03d}-'+state['phase'];im,idx=root/(label+'.png'),root/(label+'.bin')
        width,height,indices,_=png(im);assert (width,height)==(640,350) and indices==idx.read_bytes()
        for p in (im,idx,persistent):artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
    assert len(artifacts)==953
    clocks=[s for s in groups['clocks'] if int(s['packet'])>=367]
    assert [s['packet'] for s in clocks]==[str(n) for n in range(367,382)] and all(s['clock']=='0' and s['step']==states[int(s['packet'])-1]['step'] for s in clocks)
    native=groups['native_events'][len(parent['native_events']):]
    assert [(s['packet'],s['ida_linear']) for s in native]==[('372','1807b'),('372','18098'),('372','180ae'),('372','180ae'),('374','18197'),('374','1821d'),('378','18313'),('378','18338'),('378','1834e')]
    ops_path=root/(prefix+'-fileops.json');ops=json.loads(ops_path.read_text())
    parent_ops=json.loads((root/(parent_prefix+'-fileops.json')).read_text());assert ops[:len(parent_ops)]==parent_ops
    assert not any(o['Fn'] in (0x3c,0x40,0x41) for o in ops if o['Step']>=int(states[365]['step']))
    scratch=root.parent/(prefix+'-scratch');assert sorted(p.name for p in scratch.iterdir())==['dragon1.dat','player.dat']
    done=[fields(s) for s in lines if s.startswith('DQ3_SAVE_OBSERVED ')];assert len(done)==1 and done[0]['packet']=='381' and done[0]['step']==states[-1]['step']
    return dict(scope='正常新遊戲381，同部位physical5→4甲胄換穿、詳細狀況與返回；remake未驗證',meta=meta,**groups,artifacts=artifacts,parent_source_sha256=digest(parent_path),prefix366_unchanged=True,normal_inputs=419,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,persistent_transaction_verified=True,original_saved_files_unchanged=True,native_clock_limit=parent['native_clock_limit'],fileops_sha256=digest(ops_path),log_sha256=digest(root/(prefix+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False)

if __name__=='__main__':
    result=validate();p=Path('/work/dosgolem-opening/issue4-equip-armor-replace-normal-r1-source-r1-receipt.json');assert not p.exists();p.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('SOURCE_PASS',len(result['artifacts']),digest(p))
