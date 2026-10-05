from pathlib import Path
import hashlib,json,struct,sys
sys.path.insert(0,'/repo/tools')
from verify_dosgolem_mother_return import fields,png
from verify_dosgolem_first_move import RUNTIME_HASHES
w=Path('/work/dosgolem-opening');prefix='issue4-equip-wear-normal-r1';parent_prefix='issue4-equip-entry-normal-r3'
digest=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
parent_file=w/(parent_prefix+'-source-r1-receipt.json')
assert digest(parent_file)=='234fc67e75ac7a4ca87fb07c5221e4c6726c44cb40f1f72e2db3be6f868bbeb7'
parent=json.loads(parent_file.read_text());meta=json.loads((w/(prefix+'-meta.json')).read_text())
for k in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','build_flags'):assert meta[k]==parent['meta'][k],k
assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
assert digest(Path('/repo/tools/dosgolem_field_equipment_wear_probe.py'))==meta['producer_sha256']
for suffix,key in [('-probe-source.go','probe_source_sha256'),('-probe','probe_sha256')]:assert digest(w/(prefix+suffix))==meta[key]
lines=(w/(prefix+'.log')).read_text().splitlines()
assert '沒實作的服務（0 種）：' in lines
seed=[fields(s) for s in lines if s.startswith('DQ3_CREATION_SEED ')];assert len(seed)==1 and seed[0]['fixed']=='1357'
tags={'queued':'DQ3_QUIESCENT_QUEUED','consumed':'DQ3_QUIESCENT_INPUT','states':'DQ3_QUIESCENT_CAPTURE','actual_irq1_events':'DQ3_KEY_DELIVERED','clocks':'DQ3_COMMAND_CLOCK','rasters':'DQ3_COMMAND_RASTER','item_rasters':'DQ3_ITEM_RASTER','writer_events':'DQ3_ITEM_REORDER_WRITER','text_events':'DQ3_EMPTY_RECRUIT_TEXT','observations':'DQ3_ITEM_DROP_OBSERVE','status_rasters':'DQ3_STATUS_RASTER','detail_observations':'DQ3_FIELD_DETAIL_OBSERVE','summary_observations':'DQ3_FIELD_SUMMARY_OBSERVE','reorder_observations':'DQ3_FIELD_REORDER_OBSERVE','spell_empty_observations':'DQ3_FIELD_SPELL_EMPTY_OBSERVE','equip_observations':'DQ3_FIELD_EQUIP_ENTRY_OBSERVE'}
groups={k:[fields(s) for s in lines if s.startswith(t+' ')] for k,t in tags.items()}
for k,rows in groups.items():assert rows[:len(parent[k])]==parent[k],k
q,c,states,irq=(groups[k] for k in ['queued','consumed','states','actual_irq1_events'])
assert len(q)==len(c)==len(states)==339 and len(irq)==754
scans=[v for _,v in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in q]
assert len(scans)==377 and [int(s['port60'],16) for s in irq]==[v for scan in scans for v in (scan,scan|128)]
assert [int(s['count']) for s in irq]==list(range(1,755))
assert [s['scan'] for s in q[314:]]==['39','50','50','39','50','39','50','50','39','01','01','4b','4d','39','50','50','39','48','39','48','39','01','01','4b','4d']
artifacts=[]
for a in parent['artifacts']:
    p=w/a['path'].replace(parent_prefix,prefix,1)
    assert p.stat().st_size==a['size'] and digest(p)==a['sha256']
    artifacts.append(dict(path=p.name,size=a['size'],sha256=a['sha256']))
base=(w/(parent_prefix+'-persistent-314.bin')).read_bytes();items=Path('/repo/assets_raw/ITEM.DAT').read_bytes()
assert hashlib.sha256(items).hexdigest()=='7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d'
for n,(queued,consumed,state) in enumerate(zip(q,c,states),1):
    assert queued['packet']==consumed['packet']==state['packet']==str(n)
    assert queued['scan']==consumed['scan']==state['scan'] and queued['kind']==state['kind']
    assert int(queued['step'])==int(state['queued_step'])<int(consumed['step'])<int(state['step'])
    assert int(consumed['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
    if n>1:assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
    if n<=314:continue
    expected=bytearray(base)
    if 320<=n<333:struct.pack_into('<H',expected,342+0x40,0x8003)
    if n>=335:struct.pack_into('<H',expected,342+0x48,30)
    if 325<=n<337:struct.pack_into('<H',expected,342+0x1c,8+items[3*7])
    if n>=337:struct.pack_into('<H',expected,342+0x20,4)
    if n in (326,338):expected[10]=2
    assert (w/(prefix+f'-persistent-{n:03d}.bin')).read_bytes()==expected,n
    assert bytes.fromhex(state['actor'])==expected[342:470],n
    label=prefix+f'-packet-{n:03d}-'+state['phase'];im,idx=w/(label+'.png'),w/(label+'.bin')
    width,height,pixels,_=png(im);assert (width,height)==(640,350) and pixels==idx.read_bytes()
    for p in (im,idx,w/(prefix+f'-persistent-{n:03d}.bin')):artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
assert len(artifacts)==825
obs=[fields(s) for s in lines if s.startswith('DQ3_FIELD_EQUIP_WEAR_OBSERVE ')]
heads=[o for o in obs if o['ida_linear']=='17feb'];assert len(heads)==8
assert [bytes.fromhex(o['window40dc'])[10:12].hex() for o in heads]==['a901','aa01','ab01','ac01']*2
assert [o['raw0638'] for o in heads]==['2','3','0','0']*2
assert all(o['raw259c']=='1' for o in heads)
assert bytes.fromhex(heads[0]['candidates'])[:6].hex()=='010002030003'
assert bytes.fromhex(heads[1]['candidates'])[:9].hex()=='1f00041f00051e8007'
assert bytes.fromhex(heads[4]['candidates'])[:6].hex()=='010002038003'
assert any(o['ida_linear']=='18098' and o['packet']=='320' for o in obs)
assert any(o['ida_linear']=='180d9' and o['packet']=='333' for o in obs)
assert any(o['ida_linear']=='180d9' and o['packet']=='335' for o in obs)
opsfile=w/(prefix+'-fileops.json');ops=json.loads(opsfile.read_text())
assert not list((w.parent/(prefix+'-scratch')).iterdir())
assert not any(o['Fn'] in (0x3c,0x40,0x41) or o.get('Name','').lower().startswith('dragon') for o in ops if o['Step']>=int(states[313]['step']))
done=[fields(s) for s in lines if s.startswith('DQ3_SAVE_OBSERVED ')];assert len(done)==1 and done[0]['packet']=='339' and done[0]['step']==states[-1]['step']
report=dict(scope='normal new-game339 single hero equipment selection, same armor, none/unwear, four-slot return and next moves; remake not verified',meta=meta,**groups,artifacts=artifacts,wear_observations=obs,parent_source_sha256=digest(parent_file),prefix314_unchanged=True,normal_inputs=377,return_verified=True,persistent_transaction_verified=True,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,fileops_sha256=digest(opsfile),log_sha256=digest(w/(prefix+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False)
p=w/(prefix+'-source-r1-receipt.json');assert not p.exists();p.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print('SOURCE_PASS',len(artifacts),digest(p))
