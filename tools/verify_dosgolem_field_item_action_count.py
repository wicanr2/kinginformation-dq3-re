"""正常394單人道具三動作繞回來源checker，限定範圍與容器契約見docs/188。"""
from pathlib import Path
import ast, hashlib, json, struct, sys
sys.path.insert(0,'/repo/tools')
from verify_dosgolem_mother_return import digest,fields,png
from verify_dosgolem_first_move import RUNTIME_HASHES

def validate(root=Path('/work/dosgolem-opening'),producer=Path('/repo/tools/dosgolem_field_item_action_count_probe.py')):
    prefix='issue4-item-action-count-normal-r1';parent_prefix='issue4-equip-armor-replace-normal-r1'
    parent_path=root/(parent_prefix+'-source-r1-receipt.json')
    assert digest(parent_path)=='d13ac42535c7f8b0bf4844ef8820f1f8d9e3b25b08095478fe4237ebdd43771f'
    parent=json.loads(parent_path.read_text());meta=json.loads((root/(prefix+'-meta.json')).read_text())
    for k in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','build_flags','docker_image'):assert meta[k]==parent['meta'][k],k
    assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    assert digest(producer)==meta['producer_sha256']=='be351a1ec1acd4a047d05f5d2db639edeefd81dc114e32b82ef3c2886d721c78'
    for suffix,key in [('-probe-source.go','probe_source_sha256'),('-probe','probe_sha256')]:assert digest(root/(prefix+suffix))==meta[key]
    assert meta['probe_source_sha256']=='66e318bcb54d8046db15d72e9a74e0faa85abaa7a8bdbdb3d8b91e17c7dae41f'
    assert meta['probe_sha256']=='c7f327f035b4a2ab6fb1882669d966f7ed3dd3420b3c0945ee4988de3c1d2b42'
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
    assert len(q)==len(c)==len(states)==394 and len(irq)==864
    scans=[v for _,v in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in q]
    assert len(scans)==432 and [int(s['port60'],16) for s in irq]==[v for scan in scans for v in (scan,scan|128)]
    assert [int(s['count']) for s in irq]==list(range(1,865))
    assert [s['scan'] for s in q[381:]]==['39','50','50','50','50','39','39','50','50','50','01','4b','4d']
    assert [s['phase'] for s in states[381:]]==['choice']*10+['ready']*3
    choices=[(6,1),(6,2),(6,3),(6,4),(6,5),(5,1),(3,1),(3,2),(3,3),(3,1),(6,5),(6,5),(6,5)]
    assert [(int(s['choice_count']),int(s['choice_cursor'])) for s in states[381:]]==choices
    artifacts=[]
    for a in parent['artifacts']:
        relative=a['path'].replace(parent_prefix,prefix,1);p=(root if '/' not in relative else root.parent)/relative
        assert p.stat().st_size==a['size'] and digest(p)==a['sha256'],relative
        artifacts.append(dict(path=relative,size=a['size'],sha256=a['sha256']))
    base=(root/(parent_prefix+'-persistent-381.bin')).read_bytes()
    for n,(queued,taken,state) in enumerate(zip(q,c,states),1):
        assert queued['packet']==taken['packet']==state['packet']==str(n)
        assert queued['scan']==taken['scan']==state['scan'] and queued['kind']==state['kind']
        assert int(queued['step'])==int(state['queued_step'])<int(taken['step'])<int(state['step'])
        assert int(taken['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
        if n>1:assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
        if n<=381:continue
        expected=bytearray(base)
        if n==393:expected[10]=2
        persistent=root/(prefix+f'-persistent-{n:03d}.bin');assert persistent.read_bytes()==expected,n
        assert bytes.fromhex(state['actor'])==expected[342:470] and state['flags']==states[380]['flags']
        label=prefix+f'-packet-{n:03d}-'+state['phase'];im,idx=root/(label+'.png'),root/(label+'.bin')
        width,height,indices,_=png(im);assert (width,height)==(640,350) and indices==idx.read_bytes()
        for p in (im,idx,persistent):artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
    assert len(artifacts)==992
    clocks=[s for s in groups['clocks'] if int(s['packet'])>=382]
    assert [s['packet'] for s in clocks]==[str(n) for n in range(382,395)] and all(s['clock']=='0' and s['step']==states[int(s['packet'])-1]['step'] for s in clocks)
    item_rows=[v for v in groups['item_rasters'] if int(v['packet'])>=382]
    assert [v['packet'] for v in item_rows]==[str(n) for n in range(382,395)]
    for row in item_rows:
        assert row['step']==states[int(row['packet'])-1]['step']
        assert row['owner062d']=='1' and bytes.fromhex(row['action4050'])==exe.read_bytes()[0x1a190:0x1a190+64]
    native=groups['native_events'][len(parent['native_events']):]
    assert native==[],native
    ops_path=root/(prefix+'-fileops.json');ops=json.loads(ops_path.read_text())
    parent_ops=json.loads((root/(parent_prefix+'-fileops.json')).read_text());assert ops[:len(parent_ops)]==parent_ops
    assert not any(o['Fn'] in (0x3c,0x40,0x41) for o in ops if o['Step']>=int(states[380]['step']))
    scratch=root.parent/(prefix+'-scratch');assert sorted(p.name for p in scratch.iterdir())==['dragon1.dat','player.dat']
    done=[fields(s) for s in lines if s.startswith('DQ3_SAVE_OBSERVED ')];assert len(done)==1 and done[0]['packet']=='394' and done[0]['step']==states[-1]['step']
    return dict(scope='正常新遊戲394，五物品三動作1→2→3→1繞回、Esc返回與行走；remake未驗證',meta=meta,**groups,artifacts=artifacts,parent_source_sha256=digest(parent_path),prefix381_unchanged=True,normal_inputs=432,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,persistent_transaction_verified=True,original_saved_files_unchanged=True,native_clock_limit=parent['native_clock_limit'],fileops_sha256=digest(ops_path),log_sha256=digest(root/(prefix+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False)

if __name__=='__main__':
    result=validate();p=Path('/work/dosgolem-opening/issue4-item-action-count-normal-r1-source-r1-receipt.json');assert not p.exists();p.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('SOURCE_PASS',len(result['artifacts']),digest(p))
