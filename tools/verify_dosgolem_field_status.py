"""接受正常狀況選單導航、取消與下一步的有限收據；入口 docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = 'issue4-field-status-normal-r3'
PARENT = 'issue4-item-drop-normal-r3'
PARENT_HASH = '5ca9fce6caff5bd1c1f27b489fba1ef9f687a808374d7fd544e94728d45f1971'
PRODUCER_HASH = 'a4b0d612f186753399af342bf32fcacc5ca235d415f2ac0eed22540c75bfd6ba'
SOURCE_HASH = '1289f44e9f9552dbb1b9622fe89ce6b0093e5d73e81c02ccce9453422404e848'
BINARY_HASH = '0dbd8a510513202522fb77f3ce7a8c876dc88c93d053451d4ab701839fea27c0'

def validate(root, producer):
    parent_path = root/(PARENT+'-source-r1-receipt.json')
    assert digest(parent_path) == PARENT_HASH
    parent = json.loads(parent_path.read_text())
    assert digest(root/(PARENT+'.log')) == parent['log_sha256']
    meta = json.loads((root/(PREFIX+'-meta.json')).read_text())
    for key in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','docker_image','build_flags'):
        assert meta[key] == parent['meta'][key], key
    assert all(meta[k] == v for k,v in RUNTIME_HASHES.items())
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    assert '-load-state' not in meta['args']
    assert digest(producer) == meta['producer_sha256'] == PRODUCER_HASH
    assert digest(root/(PREFIX+'-probe-source.go')) == meta['probe_source_sha256'] == SOURCE_HASH
    assert digest(root/(PREFIX+'-probe')) == meta['probe_sha256'] == BINARY_HASH
    lines = (root/(PREFIX+'.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in x or x.startswith('DQ3_SAVE_BOUND ') for x in lines)
    seeds = [fields(x) for x in lines if x.startswith('DQ3_CREATION_SEED ')]
    assert len(seeds) == 1 and seeds[0]['fixed'] == '1357'
    tags = dict(queued='DQ3_QUIESCENT_QUEUED', consumed='DQ3_QUIESCENT_INPUT', states='DQ3_QUIESCENT_CAPTURE', actual_irq1_events='DQ3_KEY_DELIVERED', clocks='DQ3_COMMAND_CLOCK', rasters='DQ3_COMMAND_RASTER', item_rasters='DQ3_ITEM_RASTER', writer_events='DQ3_ITEM_REORDER_WRITER', text_events='DQ3_EMPTY_RECRUIT_TEXT', observations='DQ3_ITEM_DROP_OBSERVE')
    groups = {k:[fields(x) for x in lines if x.startswith(tag+' ')] for k,tag in tags.items()}
    for key, rows in groups.items():
        assert rows[:len(parent[key])] == parent[key], 'prefix261 '+key
    q,taken,states,irq = (groups[k] for k in ('queued','consumed','states','actual_irq1_events'))
    assert len(q) == len(taken) == len(states) == 273
    scans = [v for _,v in meta['normal_prefix_inputs']] + [int(x['scan'],16) for x in q]
    assert len(scans) == 311 and len(irq) == 622
    assert [int(x['port60'],16) for x in irq] == [v for scan in scans for v in (scan,scan|128)]
    assert [int(x['count']) for x in irq] == list(range(1,623))
    assert [int(x['scan'],16) for x in q[261:]] == [0x39,0x50,0x39,0x50,0x50,0x50,0x48,0x48,0x48,0x01,0x4b,0x4d]
    baseline = (root/(PARENT+'-persistent-261.bin')).read_bytes()
    assert len(baseline) == 2172
    artifacts=[]
    for a in parent['artifacts']:
        old=root/a['path'];new=root/a['path'].replace(PARENT,PREFIX,1)
        assert old.stat().st_size == new.stat().st_size == a['size']
        assert digest(old)==digest(new)==a['sha256'],new.name
        artifacts.append(dict(path=new.name,size=a['size'],sha256=a['sha256']))
    for n,(queued,consumed,state) in enumerate(zip(q,taken,states),1):
        assert queued['packet']==consumed['packet']==state['packet']==str(n)
        assert queued['scan']==consumed['scan']==state['scan'] and queued['kind']==state['kind']
        assert int(queued['step'])==int(state['queued_step'])<int(consumed['step'])<int(state['step'])
        assert int(consumed['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
        if n>1: assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
        if n<=261:continue
        expected=bytearray(baseline)
        if n==272:expected[10]=2
        persistent=root/(PREFIX+f'-persistent-{n:03d}.bin')
        assert persistent.read_bytes()==bytes(expected),'persistent side effect '+str(n)
        assert bytes.fromhex(state['actor']) == baseline[342:470]
        assert all(state[k]==states[260][k] for k in ('flags','gold_lo','gold_hi','player_y','raw0b24'))
        assert state['player_x']==('2' if n==272 else '3')
        assert state['last_record']=='272'
        label=PREFIX+f'-packet-{n:03d}-'+state['phase']
        image,indexed=root/(label+'.png'),root/(label+'.bin')
        width,height,pixels,_=png(image)
        assert (width,height)==(640,350) and pixels==indexed.read_bytes()
        for p in (image,indexed,persistent):artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
    assert len(artifacts)==627
    expected=[('choice',6,1),('choice',6,2),('choice',3,1),('choice',3,2),('choice',3,3),('choice',3,1),('choice',3,3),('choice',3,2),('choice',3,1),('ready',6,2),('ready',6,2),('ready',6,2)]
    for state,(phase,count,cursor) in zip(states[261:],expected):
        assert (state['phase'],int(state['choice_count']),int(state['choice_cursor']))==(phase,count,cursor)
    assert [x['packet'] for x in groups['clocks']]==list(map(str,range(193,274)))
    assert all(x['clock']=='30' and x['raw526c']=='1' for x in groups['clocks'])
    obs=[fields(x) for x in lines if x.startswith('DQ3_STATUS_RASTER ')]
    exe=root.parents[1]/'assets_raw/DQ3.EXE'
    assert exe.stat().st_size==meta['original_size'] and digest(exe)==meta['original_sha256']
    window=exe.read_bytes()[0x19ff4:0x19ff4+42].hex()
    assert [x['packet'] for x in obs]==list(map(str,range(262,274)))
    assert all(x['window3eb4']==window and x['raw258f']=='8' for x in obs)
    assert all(x['step']==states[int(x['packet'])-1]['step'] for x in obs)
    ops_path=root/(PREFIX+'-fileops.json');ops=json.loads(ops_path.read_text())
    assert ops and not list((root.parent/(PREFIX+'-scratch')).iterdir())
    assert not any(x['Fn'] in (0x3c,0x40,0x41) or x.get('Name','').lower().startswith('dragon') for x in ops if x['Step']>=int(states[260]['step']))
    done=[fields(x) for x in lines if x.startswith('DQ3_SAVE_OBSERVED ')]
    assert len(done)==1 and done[0]['packet']=='273' and done[0]['step']==states[-1]['step']
    return dict(scope='正常261後狀況選單、單欄上下繞回、Esc及下一步；有限273包',meta=meta,**groups,status_rasters=obs,artifacts=artifacts,done=done[0],parent_source_sha256=PARENT_HASH,prefix261_unchanged=True,normal_inputs=311,original_size=meta['original_size'],original_sha256=meta['original_sha256'],upstream_revision=meta['upstream_revision'],seed=meta['seed'],seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,persistent_status_unchanged=True,fileops_sha256=digest(ops_path),log_sha256=digest(root/(PREFIX+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    for name in ('root','producer','receipt'):p.add_argument(name,type=Path)
    a=p.parse_args();assert not a.receipt.exists()
    report=validate(a.root,a.producer)
    with a.receipt.open('x') as f:json.dump(report,f,ensure_ascii=False,indent=2);f.write('\n')
    print('SOURCE_PASS',len(report['artifacts']),digest(a.receipt),flush=True)
