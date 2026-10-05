"""Docker-only normal394 -> search401 -> fresh Enter402 and movement404.

Source checks only. Scope, reproduction and READY ledger: docs/188.
"""
from pathlib import Path
import ast, gc, hashlib, json, sys
sys.path.insert(0, '/repo/tools')
from verify_dosgolem_mother_return import digest, fields, png
from verify_dosgolem_first_move import RUNTIME_HASHES

ROOT = Path('/work/dosgolem-opening')
CASES = {
    'first': ('issue4-search-first-normal-r1', 401, 'c4ceafced69fc7d07b1b2cfe859d420af4deb6917d5dd160ec6f8d34adae4762', '3fd48e48e4da2e5fdfd4cec58f994afe3c16fb47345d21bdc82054b1fdc94f88', '7520b598964de6f5a0b929a83661f3f0d16beadae1628c0d4f7acfaa958e430f'),
    'return': ('issue4-search-return-normal-r1', 404, '679ea70d4aeda29e5ee31116d2506a4cd5c069d175dec4adc20fe3950368f8ee', '0fea6b6e778dbfb7376da90607623b4042e92f6a9ce68e394c2e33005844f137', 'a88d7e6ba11c2821b1387e686300e01074b0b50f4ea5d577ce2378c36106bcba'),
}

def validate(case, root=ROOT):
    prefix, count, producer_hash, source_hash, binary_hash = CASES[case]
    pp = 'issue4-item-action-count-normal-r1'
    parent_path = root/(pp+'-source-r1-receipt.json')
    assert digest(parent_path) == 'c48788435ba495f97cc6626d4b83b6e657ea8c012de55d0b44c409ba74ec2e78'
    parent = json.loads(parent_path.read_text())
    meta = json.loads((root/(prefix+'-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'build_flags', 'docker_image'):
        assert meta[key] == parent['meta'][key], key
    assert all(meta[k] == v for k, v in RUNTIME_HASHES.items())
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    producer = Path('/repo/tools/dosgolem_field_search_'+case+'_probe.py')
    assert digest(producer) == meta['producer_sha256'] == producer_hash
    assert digest(root/(prefix+'-probe-source.go')) == meta['probe_source_sha256'] == source_hash
    assert digest(root/(prefix+'-probe')) == meta['probe_sha256'] == binary_hash
    exe = Path('/repo/assets_raw/DQ3.EXE')
    assert exe.stat().st_size == meta['original_size'] and digest(exe) == meta['original_sha256']
    tree = ast.parse(Path('/repo/tools/verify_dosgolem_field_equipment_wear.py').read_text())
    tags = next(ast.literal_eval(n.value) for n in tree.body if isinstance(n, ast.Assign) and any(isinstance(t, ast.Name) and t.id == 'tags' for t in n.targets))
    tags.update(wear_observations='DQ3_FIELD_EQUIP_WEAR_OBSERVE', native_events='DQ3_ARMOR_NATIVE')
    groups = {k: [] for k in tags}; seed = []; done = []; native = []; supported = False
    with (root/(prefix+'.log')).open() as log:
        for line in log:
            assert not line.startswith('DQ3_SAVE_BOUND ') and '找不到的檔（' not in line
            supported |= line.strip() == '沒實作的服務（0 種）：'
            for key, tag in tags.items():
                if line.startswith(tag+' '): groups[key].append(fields(line)); break
            if line.startswith('DQ3_CREATION_SEED '): seed.append(fields(line))
            if line.startswith('DQ3_SAVE_OBSERVED '): done.append(fields(line))
            if line.startswith('DQ3_SEARCH_NATIVE '): native.append(fields(line))
    assert supported and len(seed) == 1 and seed[0]['fixed'] == '1357'
    for key, rows in groups.items(): assert rows[:len(parent[key])] == parent[key], key
    q, consumed, states, irq = (groups[k] for k in ('queued', 'consumed', 'states', 'actual_irq1_events'))
    assert len(q) == len(consumed) == len(states) == count and len(irq) == (count+38)*2
    scans = [v for _, v in meta['normal_prefix_inputs']] + [int(s['scan'],16) for s in q]
    assert len(scans) == count+38 and [int(s['port60'],16) for s in irq] == [v for scan in scans for v in (scan,scan|128)]
    assert [int(s['count']) for s in irq] == list(range(1,len(irq)+1))
    expected_scans = ['39','50','50','50','50','50','39'] + (['1c','4b','4d'] if case == 'return' else [])
    assert [s['scan'] for s in q[394:]] == expected_scans
    assert [s['phase'] for s in states[394:]] == ['choice']*6+['waiting']+(['ready']*3 if case == 'return' else [])
    assert [(int(s['choice_count']),int(s['choice_cursor'])) for s in states[394:400]] == [(6,n) for n in range(1,7)]
    assert states[400]['last_record'] == '265' and states[400]['ida_linear'] == '21133'
    artifacts = []
    for a in parent['artifacts']:
        relative = a['path'].replace(pp,prefix,1); path = (root if '/' not in relative else root.parent)/relative
        assert path.stat().st_size == a['size'] and digest(path) == a['sha256'], relative
        artifacts.append(dict(path=relative,size=a['size'],sha256=a['sha256']))
    baseline = (root/(pp+'-persistent-394.bin')).read_bytes()
    for n, (queued,taken,state) in enumerate(zip(q,consumed,states),1):
        assert queued['packet'] == taken['packet'] == state['packet'] == str(n)
        assert queued['scan'] == taken['scan'] == state['scan'] and queued['kind'] == state['kind']
        assert int(queued['step']) == int(state['queued_step']) < int(taken['step']) < int(state['step'])
        assert int(taken['irqs']) == 76+n*2-1 and int(irq[76+n*2-1]['step']) <= int(state['step'])
        if n>1: assert all(queued[k] == states[n-2][k] for k in ('step','phase','ida_linear'))
        if n<=394: continue
        expected = bytearray(baseline)
        if n==403: expected[10]=2
        persistent = root/(prefix+f'-persistent-{n:03d}.bin')
        assert persistent.read_bytes() == expected, n
        assert bytes.fromhex(state['actor']) == expected[342:470] and state['flags'] == states[393]['flags']
        label = prefix+f'-packet-{n:03d}-'+state['phase']; im,idx = root/(label+'.png'),root/(label+'.bin')
        width,height,indices,_ = png(im)
        assert (width,height) == (640,350) and indices == idx.read_bytes()
        for path in (im,idx,persistent): artifacts.append(dict(path=path.name,size=path.stat().st_size,sha256=digest(path)))
    assert len(artifacts) == 992+(count-394)*3
    clocks = [s for s in groups['clocks'] if int(s['packet'])>=395]
    assert [s['packet'] for s in clocks] == [str(n) for n in range(395,count+1)]
    assert all(s['clock']=='0' and s['step']==states[int(s['packet'])-1]['step'] for s in clocks)
    ops_path = root/(prefix+'-fileops.json'); ops = json.loads(ops_path.read_text())
    parent_ops = json.loads((root/(pp+'-fileops.json')).read_text())
    assert ops[:len(parent_ops)] == parent_ops
    assert not any(o['Fn'] in (0x3c,0x40,0x41) for o in ops if o['Step']>=int(states[393]['step']))
    assert len(done)==1 and done[0]['packet']==str(count) and done[0]['step']==states[-1]['step']
    if case == 'return':
        first_path=root/(CASES['first'][0]+'-source-r1-receipt.json')
        first=json.loads(first_path.read_text())
        for key,rows in groups.items(): assert rows[:len(first[key])]==first[key],key
        for a in first['artifacts']:
            relative=a['path'].replace(CASES['first'][0],prefix,1); path=(root if '/' not in relative else root.parent)/relative
            assert digest(path)==a['sha256'] and path.stat().st_size==a['size'],relative
        expected = ['18966','18977','18986','18c75','18c83','15002','18c93','18c96','21414','18c9b','21414','18ca3','2111b','18ca8','18cac','1f604','1f604','1f604']
        assert [r['ida_linear'] for r in native]==expected
        assert all(r['raw01f7']=='0' and r['raw0b5d']=='0' and r['raw4f2d']=='1' and r['raw4f3b']=='0' and r['x']=='3' and r['y']=='18' and r['window3e6e']=='0b011300ee002c0060009401000000000000000000000c09' for r in native)
        assert [r['DI'] for r in native if r['ida_linear']=='21414']==['0108','0109']
        assert [r['packet'] for r in native]==['401']*13+['402']*5
        del first
    else: assert not native
    return dict(scope='normal394 single healthy hero, native records264/265, fresh Enter and field return; remake unverified',meta=meta,**groups,search_native_events=native,artifacts=artifacts,parent_source_sha256=digest(parent_path),prefix394_unchanged=True,normal_inputs=count+38,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,persistent_transaction_verified=True,original_saved_files_unchanged=True,native_clock_limit=parent['native_clock_limit'],fileops_sha256=digest(ops_path),log_sha256=digest(root/(prefix+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False)

if __name__ == '__main__':
    for case in ('first','return'):
        result=validate(case); path=ROOT/(CASES[case][0]+'-source-r1-receipt.json')
        assert not path.exists(); path.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
        print('SOURCE_PASS',case,len(result['artifacts']),digest(path),flush=True)
        del result; gc.collect()
