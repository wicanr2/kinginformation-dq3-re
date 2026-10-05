"""接受正常丟掉的有限原版收據。入口及證據範圍見 docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = 'issue4-item-drop-normal-r3'
PARENT = 'issue4-item-use-normal-r2'
PARENT_HASH = 'c11efcbdb9ff928af1d3a9c8d31c7703b5a0b4c9ded3cc55cc0949b9cf2173ee'
PRODUCER_HASH = '69eedc7ac129246b59de3ab084d5162279b31a3a81e305da1f8cf950001d33c1'
SOURCE_HASH = 'eeeca052297f6456632323854283d268f3fe746526cbb24d18d8faea0a395221'
BINARY_HASH = 'f4bc03616e90dcd944da13780ecc44375c795055f061e60334817d39fd69e1fa'

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
    tags = dict(queued='DQ3_QUIESCENT_QUEUED', consumed='DQ3_QUIESCENT_INPUT', states='DQ3_QUIESCENT_CAPTURE', actual_irq1_events='DQ3_KEY_DELIVERED', clocks='DQ3_COMMAND_CLOCK', rasters='DQ3_COMMAND_RASTER', item_rasters='DQ3_ITEM_RASTER', writer_events='DQ3_ITEM_REORDER_WRITER', text_events='DQ3_EMPTY_RECRUIT_TEXT')
    groups = {k:[fields(x) for x in lines if x.startswith(tag+' ')] for k,tag in tags.items()}
    for key, rows in groups.items():
        assert rows[:len(parent[key])] == parent[key], 'prefix233 '+key
    q,taken,states,irq = (groups[k] for k in ('queued','consumed','states','actual_irq1_events'))
    assert len(q) == len(taken) == len(states) == 261
    scans = [v for _,v in meta['normal_prefix_inputs']] + [int(x['scan'],16) for x in q]
    assert len(scans) == 299 and len(irq) == 598
    assert [int(x['port60'],16) for x in irq] == [v for scan in scans for v in (scan,scan|128)]
    assert [int(x['count']) for x in irq] == list(range(1,599))
    expected_scans = [0x39,0x50,0x4d,0x39,0x39,0x50,0x50,0x39,0x1c,0x39,0x50,0x4d,0x39,0x39,0x50,0x50,0x39,0x1c,0x39,0x50,0x4d,0x39,0x48,0x39,0x50,0x50,0x39,0x1c]
    assert [int(x['scan'],16) for x in q[233:]] == expected_scans
    baseline = (root/(PARENT+'-persistent-233.bin')).read_bytes()
    assert len(baseline)==2172 and baseline[342:470].hex()==states[232]['actor']
    artifacts=[]
    for a in parent['artifacts']:
        old=root/a['path'];new=root/a['path'].replace(PARENT,PREFIX,1)
        assert old.stat().st_size == new.stat().st_size == a['size']
        assert digest(old)==digest(new)==a['sha256'],new.name
        artifacts.append(dict(path=new.name,size=a['size'],sha256=a['sha256']))
    expected_words=['0000','0001','0001','0003','001f','001f','00ff','801e']
    for n,(queued,consumed,state) in enumerate(zip(q,taken,states),1):
        assert queued['packet']==consumed['packet']==state['packet']==str(n)
        assert queued['scan']==consumed['scan']==state['scan'] and queued['kind']==state['kind']
        assert int(queued['step'])==int(state['queued_step'])<int(consumed['step'])<int(state['step'])
        assert int(consumed['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
        if n>1: assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
        if n<=233:continue
        if n==241:expected_words[0]='00ff'
        if n==250:expected_words[1]='00ff'
        words=b''.join(int(w,16).to_bytes(2,'little') for w in expected_words)
        actor=bytes.fromhex(state['actor'])
        assert actor[0x3a:0x4a]==words
        expected=bytearray(baseline);expected[400:416]=words
        persistent=root/(PREFIX+f'-persistent-{n:03d}.bin')
        assert persistent.read_bytes()==bytes(expected),'exact persistent transaction '+str(n)
        assert all(state[k]==states[232][k] for k in ('flags','gold_lo','gold_hi','player_x','player_y','raw0b24'))
        label=PREFIX+f'-packet-{n:03d}-'+state['phase']
        image,indexed=root/(label+'.png'),root/(label+'.bin')
        width,height,pixels,_=png(image)
        assert (width,height)==(640,350) and pixels==indexed.read_bytes()
        for p in (image,indexed,persistent):artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
    assert len(artifacts)==591
    for n,phase,count,cursor,record in ((237,'choice',7,1,'341'),(241,'waiting',3,3,'277'),(242,'ready',6,5,'277'),(246,'choice',6,1,'277'),(250,'waiting',3,3,'277'),(251,'ready',6,5,'277'),(255,'choice',5,1,'277'),(256,'choice',5,5,'277'),(260,'waiting',3,3,'272'),(261,'ready',6,5,'272')):
        s=states[n-1]
        assert s['phase']==phase and s['choice_count']==str(count) and s['choice_cursor']==str(cursor) and s['last_record']==record,('native state',n,s)
    assert [x['packet'] for x in groups['clocks']]==list(map(str,range(193,262)))
    assert [x['packet'] for x in groups['rasters']]==list(map(str,range(194,262)))
    assert [x['packet'] for x in groups['item_rasters']]==list(map(str,range(206,262)))
    assert all(x['clock']=='30' and x['raw526c']=='1' for x in groups['clocks'])
    for key in ('clocks','rasters','item_rasters'):
        assert all(x['step']==states[int(x['packet'])-1]['step'] for x in groups[key])
    text=groups['text_events'][len(parent['text_events']):]
    assert [(x['packet'],x['DI'],x['text_segment']) for x in text]==[('241','0115','2826'),('250','0115','2826'),('260','0110','2826')]
    obs=[fields(x) for x in lines if x.startswith('DQ3_ITEM_DROP_OBSERVE ')]
    exe=root.parents[1]/'assets_raw/DQ3.EXE'
    assert exe.stat().st_size==meta['original_size'] and digest(exe)==meta['original_sha256']
    window=exe.read_bytes()[0x19fae:0x19fae+30].hex()
    assert obs and all(x['window3e6e']==window and x['owner']=='1' for x in obs)
    for n,code,offset in ((241,0,0x50b9),(250,1,0x50bb),(260,30,0x50c7)):
        rows=[x for x in obs if x['packet']==str(n)]
        read=[x for x in rows if x['ida_linear']=='13ad5']
        assert len(read)==1 and int(read[0]['SI'],16)==offset and int(read[0]['raw2591'],16)==code
        assert read[0]['selected']==('5' if n==260 else '1')
        assert len([x for x in rows if x['ida_linear']=='13af8'])==(0 if n==260 else 1)
        calls=[x for x in rows if x['ida_linear']=='15023']
        assert len(calls)==1 and calls[0]['DI']==('0110' if n==260 else '0115')
        waits=[x for x in rows if x['ida_linear']=='21133']
        assert len(waits)==1 and waits[0]['step']==states[n-1]['step']
    assert [(x['packet'],x['ida_linear']) for x in obs if x['ida_linear']=='18197']==[('242','18197'),('251','18197')]
    ops_path=root/(PREFIX+'-fileops.json');ops=json.loads(ops_path.read_text())
    assert ops and not list((root.parent/(PREFIX+'-scratch')).iterdir())
    assert not any(x['Fn'] in (0x3c,0x40,0x41) or x.get('Name','').lower().startswith('dragon') for x in ops if x['Step']>=int(states[192]['step']))
    done=[fields(x) for x in lines if x.startswith('DQ3_SAVE_OBSERVED ')]
    assert len(done)==1 and done[0]['packet']=='261' and done[0]['step']==states[-1]['step']
    return dict(scope='正常233後丟兩件、保留空格、拒絕穿戴物品、新Enter返回場景；有限261包',meta=meta,**groups,observations=obs,artifacts=artifacts,done=done[0],parent_source_sha256=PARENT_HASH,prefix233_unchanged=True,normal_inputs=299,original_size=meta['original_size'],original_sha256=meta['original_sha256'],upstream_revision=meta['upstream_revision'],seed=meta['seed'],seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,persistent_slot_transaction_only=True,fileops_sha256=digest(ops_path),log_sha256=digest(root/(PREFIX+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    for name in ('root','producer','receipt'):p.add_argument(name,type=Path)
    a=p.parse_args();assert not a.receipt.exists()
    report=validate(a.root,a.producer)
    with a.receipt.open('x') as f:json.dump(report,f,ensure_ascii=False,indent=2);f.write('\n')
    print('SOURCE_PASS',len(report['artifacts']),digest(a.receipt),flush=True)
