"""獨立稽核正常366的第二件甲胄、詳細狀況與穿戴後第二槽存讀檔；入口docs/188。"""
import ast
import json
import struct
from pathlib import Path

from verify_dosgolem_first_move import RUNTIME_HASHES
from verify_dosgolem_mother_return import digest, fields, png

PREFIX = 'issue4-equip-armor-normal-r1'
PARENT = 'issue4-equip-wear-normal-r1'
PARENT_HASH = '721d93f6de2123f03ad82ce991185f7329fa918898c1b1cfbf25d2870f3cf414'
PRODUCER_HASH = '5bc2ff07e9294be634921ee4cea8b94d3f4faaa10570fa9429f3837d70c166fb'


def validate(root, producer, assets):
    parent_path = root / (PARENT + '-source-r1-receipt.json')
    assert digest(parent_path) == PARENT_HASH
    parent = json.loads(parent_path.read_text())
    meta = json.loads((root / (PREFIX + '-meta.json')).read_text())
    for key in ('original_size', 'original_sha256', 'upstream_revision', 'seed', 'normal_prefix_inputs', 'generator_sha256', 'docker_image', 'build_flags'):
        assert meta[key] == parent['meta'][key], key
    assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
    assert digest(producer) == meta['producer_sha256'] == PRODUCER_HASH
    assert all(meta[key] == value for key, value in RUNTIME_HASHES.items())
    for suffix, key in [('-probe-source.go', 'probe_source_sha256'), ('-probe', 'probe_sha256')]:
        assert digest(root / (PREFIX + suffix)) == meta[key]
    assert (assets / 'DQ3.EXE').stat().st_size == meta['original_size'] and digest(assets / 'DQ3.EXE') == meta['original_sha256']
    assert digest(assets / 'ITEM.DAT') == '7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d'
    lines = (root / (PREFIX + '.log')).read_text().splitlines()
    assert '沒實作的服務（0 種）：' in lines
    assert not any(line.startswith('DQ3_SAVE_BOUND ') or '找不到的檔（' in line for line in lines)
    seed = [fields(line) for line in lines if line.startswith('DQ3_CREATION_SEED ')]
    assert len(seed) == 1 and seed[0]['fixed'] == '1357'
    # Read only the parent's literal tag contract, never execute its checker.
    tree = ast.parse(Path(__file__).with_name('verify_dosgolem_field_equipment_wear.py').read_text())
    tags = next(ast.literal_eval(n.value) for n in tree.body if isinstance(n, ast.Assign) and any(isinstance(t, ast.Name) and t.id == 'tags' for t in n.targets))
    tags['wear_observations'] = 'DQ3_FIELD_EQUIP_WEAR_OBSERVE'
    groups = {key: [fields(line) for line in lines if line.startswith(tag + ' ')] for key, tag in tags.items()}
    for key, rows in groups.items():
        assert rows[:len(parent[key])] == parent[key], key
    q, c, states, irq = (groups[key] for key in ('queued', 'consumed', 'states', 'actual_irq1_events'))
    assert len(q) == len(c) == len(states) == 366 and len(irq) == 808
    scans = [v for _, v in meta['normal_prefix_inputs']] + [int(s['scan'], 16) for s in q]
    assert len(scans) == 404 and [int(s['port60'], 16) for s in irq] == [v for scan in scans for v in (scan, scan | 128)]
    assert [int(s['count']) for s in irq] == list(range(1, 809))
    assert [s['scan'] for s in q[339:]] == ['39','50','50','39','01','50','39','01','01','39','50','39','39','1c','4b','4d','3f','1c','50','1c','1c','4b','40','50','1c','4b','4d']
    phases = ['choice']*8 + ['ready'] + ['choice']*3 + ['waiting','ready','ready','ready'] + ['choice']*3 + ['waiting','ready','ready','choice','choice','ready','ready','ready']
    assert [s['phase'] for s in states[339:]] == phases
    choices = [(6,1),(6,2),(6,3),(3,1),(4,1),(4,2),(1,1),(1,1),(6,3),(6,1),(6,2),(3,1),(3,1),(6,2),(6,2),(6,2),(2,1),(10,1),(10,2),(10,2),(10,2),(10,2),(10,1),(10,2),(10,2),(10,2),(10,2)]
    assert [(int(s['choice_count']), int(s['choice_cursor'])) for s in states[339:]] == choices
    artifacts = []
    for a in parent['artifacts']:
        p = root / a['path'].replace(PARENT, PREFIX, 1)
        assert p.stat().st_size == a['size'] and digest(p) == a['sha256'], p.name
        artifacts.append(dict(path=p.name, size=a['size'], sha256=a['sha256']))
    base = (root / (PARENT + '-persistent-339.bin')).read_bytes()
    for n, (queued, consumed, state) in enumerate(zip(q, c, states), 1):
        assert queued['packet'] == consumed['packet'] == state['packet'] == str(n)
        assert queued['scan'] == consumed['scan'] == state['scan'] and queued['kind'] == state['kind']
        assert int(queued['step']) == int(state['queued_step']) < int(consumed['step']) < int(state['step'])
        assert int(consumed['irqs']) == 76+n*2-1 and int(irq[76+n*2-1]['step']) <= int(state['step'])
        if n > 1:
            assert all(queued[k] == states[n-2][k] for k in ('step','phase','ida_linear'))
        if n <= 339:
            continue
        expected = bytearray(base)
        if n >= 346:
            struct.pack_into('<H', expected, 342+0x44, 0x801f)
        if n >= 348:
            struct.pack_into('<H', expected, 342+0x20, 8)
        # Existing F5 contract copies current coordinates into its respawn header.
        if n >= 359:
            expected[31:41] = base[6:14] + base[0:2]
        if n in (354,361,362,363,365):
            expected[10] = 2
        persistent = root / (PREFIX + f'-persistent-{n:03d}.bin')
        assert persistent.read_bytes() == expected, n
        assert bytes.fromhex(state['actor']) == expected[342:470], n
        assert state['flags'] == states[338]['flags'] and state['gold_lo'] == '0032' and state['gold_hi'] == '0000'
        label = PREFIX + f'-packet-{n:03d}-' + state['phase']
        im, idx = root / (label+'.png'), root / (label+'.bin')
        width, height, pixels, _ = png(im)
        assert (width,height) == (640,350) and pixels == idx.read_bytes()
        for p in (im,idx,persistent):
            artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
    clocks = [s for s in groups['clocks'] if int(s['packet']) >= 340]
    assert [s['packet'] for s in clocks] == [str(n) for n in range(340,367)]
    assert all(s['clock'] == ('0' if int(s['packet']) >= 364 else '30') and s['step'] == states[int(s['packet'])-1]['step'] for s in clocks)
    scratch = root.parent / (PREFIX+'-scratch')
    assert sorted(p.name for p in scratch.iterdir()) == ['dragon1.dat','player.dat']
    saved, directory = (scratch/'dragon1.dat').read_bytes(), (scratch/'player.dat').read_bytes()
    assert len(saved) == 2172 and len(directory) == 200
    assert saved == (root/(PREFIX+'-persistent-359.bin')).read_bytes() == (root/(PREFIX+'-persistent-364.bin')).read_bytes()
    original = assets/'PLAYER.DAT'
    assert digest(original) == 'a445a11f52a6711aba1433d9107d10284253e08d27aa6be2b82dc7aec91376dc'
    assert directory[:20] == original.read_bytes()[:20] and directory[40:] == original.read_bytes()[40:]
    actor = bytes.fromhex(states[358]['actor'])
    assert directory[20:38] == actor[3:21] and directory[38] == actor[21] and directory[39] == actor[2]
    for p in sorted(scratch.iterdir()):
        artifacts.append(dict(path=str(p.relative_to(root.parent)),size=p.stat().st_size,sha256=digest(p)))
    assert len(artifacts) == 908
    ops_path = root/(PREFIX+'-fileops.json');ops = json.loads(ops_path.read_text())
    creates = [o for o in ops if o['Op'] == 'create']
    assert [o['Name'] for o in creates] == ['player.dat','dragon1.dat'] and not any(o['Failed'] for o in creates)
    reads = [o for o in ops if o['Op'] == 'read' and o['Name'] == 'dragon1.dat']
    assert len(reads) == 1 and reads[0]['Len'] == reads[0]['Arg'] == 2172 and reads[0]['Step'] > int(states[362]['step']) and not reads[0]['Failed']
    assert not any(o['Name'] == 'dragon0.dat' for o in ops if o['Step'] >= int(states[338]['step']))
    native = [fields(line) for line in lines if line.startswith('DQ3_ARMOR_NATIVE ')]
    save_native = [s for s in native if s['ida_linear'].startswith(('114','115','116'))]
    assert [s['ida_linear'] for s in save_native] == ['11484','114c8','114d3','114d9','1157d','1158b','11591','1165f']
    assert all(s['DS'] == '15ed' and s['raw0722'] == '2' and s['raw0726'] == '0' for s in save_native)
    assert all(s['packet'] == ('359' if i < 4 else '364') for i,s in enumerate(save_native))
    assert [(save_native[i]['AX'],save_native[i]['CX'],save_native[i]['DX']) for i in (2,5)] == [('400b','087c','4f29'),('3f0b','087c','4f29')]
    assert [(s['packet'],s['ida_linear']) for s in native[:9]] == [('346','1807b'),('346','180ae'),('346','18098'),('346','180ae'),('348','18197'),('348','1821d'),('352','18313'),('352','18338'),('352','1834e')]
    done = [fields(line) for line in lines if line.startswith('DQ3_SAVE_OBSERVED ')]
    assert len(done) == 1 and done[0]['packet'] == '366' and done[0]['step'] == states[-1]['step']
    return dict(scope='正常新遊戲366，第二件甲胄物理格5、詳細狀況、穿戴後F5/F6第二槽及下一步；remake尚未驗證',meta=meta,**groups,artifacts=artifacts,native_events=native,native_clock_limit='DQ3_ARMOR_NATIVE.clock為DGROUP001F未解欄位；世界時鐘只用DQ3_COMMAND_CLOCK的DGROUP251D',parent_source_sha256=PARENT_HASH,prefix339_unchanged=True,normal_inputs=404,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,persistent_transaction_verified=True,normal_worn_save_load=True,other_native_slots_unchanged=True,fileops_sha256=digest(ops_path),log_sha256=digest(root/(PREFIX+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False)


if __name__ == '__main__':
    root = Path('/work/dosgolem-opening')
    result = validate(root, Path('/repo/tools/dosgolem_field_equipment_armor_probe.py'), Path('/repo/assets_raw'))
    path = root/(PREFIX+'-source-r1-receipt.json')
    assert not path.exists()
    with path.open('x') as stream:
        json.dump(result,stream,ensure_ascii=False,indent=2);stream.write('\n')
    print('SOURCE_PASS',len(result['artifacts']),digest(path))
