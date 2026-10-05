"""正常288全體狀況來源稽核；入口與執行契約見docs/188。"""
from pathlib import Path
import json,sys
sys.path.insert(0,'/repo/tools')
from verify_dosgolem_mother_return import digest,fields,png
from verify_dosgolem_first_move import RUNTIME_HASHES
w=Path('/work/dosgolem-opening')
prefix='issue4-field-summary-normal-r2'
parent_prefix='issue4-field-detail-normal-r5'
parent_file=w/(parent_prefix+'-source-r1-receipt.json')
assert digest(parent_file)=='edfad33432fb8b0dc6ee5bd59536e826af9b27e4c3b91ad4eb861365983469ff'
parent=json.loads(parent_file.read_text())
meta=json.loads((w/(prefix+'-meta.json')).read_text())
for k in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','docker_image','build_flags'):
 assert meta[k]==parent['meta'][k],k
assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
assert '-load-state' not in meta['args']
producer=Path('/repo/tools/dosgolem_field_party_summary_probe.py')
assert digest(producer)==meta['producer_sha256']
for suffix,key in [('-probe-source.go','probe_source_sha256'),('-probe','probe_sha256')]:assert digest(w/(prefix+suffix))==meta[key]
lines=(w/(prefix+'.log')).read_text().splitlines()
assert '沒實作的服務（0 種）：' in lines and not any('DQ3_SAVE_BOUND ' in s or '找不到的檔（' in s for s in lines)
seeds=[fields(s) for s in lines if s.startswith('DQ3_CREATION_SEED ')]
assert len(seeds)==1 and seeds[0]['fixed']=='1357'
tags=dict(queued='DQ3_QUIESCENT_QUEUED',consumed='DQ3_QUIESCENT_INPUT',states='DQ3_QUIESCENT_CAPTURE',actual_irq1_events='DQ3_KEY_DELIVERED',clocks='DQ3_COMMAND_CLOCK',rasters='DQ3_COMMAND_RASTER',item_rasters='DQ3_ITEM_RASTER',writer_events='DQ3_ITEM_REORDER_WRITER',text_events='DQ3_EMPTY_RECRUIT_TEXT',observations='DQ3_ITEM_DROP_OBSERVE',status_rasters='DQ3_STATUS_RASTER')
groups={k:[fields(s) for s in lines if s.startswith(t+' ')] for k,t in tags.items()}
for k,rows in groups.items():assert rows[:len(parent[k])]==parent[k],k
q,c,states,irq=(groups[k] for k in ['queued','consumed','states','actual_irq1_events'])
assert len(q)==len(c)==len(states)==288 and len(irq)==652
scans=[v for _,v in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in q]
assert len(scans)==326 and [int(s['port60'],16) for s in irq]==[v for scan in scans for v in (scan,scan|128)]
assert [int(s['count']) for s in irq]==list(range(1,653))
assert [int(s['scan'],16) for s in q[280:]]==[0x39,0x50,0x39,0x50,0x39,0x1c,0x4b,0x4d]
artifacts=[]
for a in parent['artifacts']:
 old=w/a['path'];new=w/a['path'].replace(parent_prefix,prefix,1)
 assert old.stat().st_size==new.stat().st_size==a['size'] and digest(old)==digest(new)==a['sha256']
 artifacts.append(dict(path=new.name,size=a['size'],sha256=a['sha256']))
baseline=(w/(parent_prefix+'-persistent-280.bin')).read_bytes()
for n,(queued,consumed,state) in enumerate(zip(q,c,states),1):
 assert queued['packet']==consumed['packet']==state['packet']==str(n)
 assert queued['scan']==consumed['scan']==state['scan'] and queued['kind']==state['kind']
 assert int(queued['step'])==int(state['queued_step'])<int(consumed['step'])<int(state['step'])
 assert int(consumed['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
 if n>1:assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
 if n<=280:continue
 expected=bytearray(baseline)
 if n==287:expected[10]=2
 assert (w/(prefix+f'-persistent-{n:03d}.bin')).read_bytes()==expected
 assert bytes.fromhex(state['actor'])==baseline[342:470]
 assert all(state[k]==states[279][k] for k in ('flags','gold_lo','gold_hi','player_y','raw0b24','last_record'))
 assert int(state['player_x'])==(2 if n==287 else 3)
 label=prefix+f'-packet-{n:03d}-'+state['phase']
 image,indexed=w/(label+'.png'),w/(label+'.bin')
 width,height,pixels,_=png(image)
 assert (width,height)==(640,350) and pixels==indexed.read_bytes()
 for p in (image,indexed,w/(prefix+f'-persistent-{n:03d}.bin')):artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
assert len(artifacts)==672
assert [(s['phase'],int(s['choice_count']),int(s['choice_cursor'])) for s in states[280:]]==[('choice',6,1),('choice',6,2),('choice',3,1),('choice',3,2),('waiting',3,2),('ready',6,2),('ready',6,2),('ready',6,2)]
assert all(s['clock']=='30' and s['raw526c']=='1' for s in groups['clocks'])
detail=[fields(s) for s in lines if s.startswith('DQ3_FIELD_DETAIL_OBSERVE ')]
assert detail[:len(parent['detail_observations'])]==parent['detail_observations']
assert all(int(s['packet'])>280 for s in detail[len(parent['detail_observations']):])
obs=[fields(s) for s in lines if s.startswith('DQ3_FIELD_SUMMARY_OBSERVE ')]
assert [int(s['ida_linear'],16) for s in obs]==[0x185ef,0x1f590,0x1f4e3,0x2111b]
raw=Path('/repo/assets_raw/DQ3.EXE').read_bytes()
assert len(raw)==meta['original_size'] and digest(Path('/repo/assets_raw/DQ3.EXE'))==meta['original_sha256']
assert all(s['packet']=='285' for s in obs)
assert obs[1]['SI']=='3e84' and obs[1]['window_si'][:48]==raw[0x19fc4:0x19fc4+24].hex()
body=bytearray(raw[0x1a03c:0x1a03c+28])
body[6:8]=bytes.fromhex('0e00');body[12:14]=bytes.fromhex('0100')
assert obs[2]['SI']=='3efc' and obs[2]['window_si']==body.hex()
assert int(c[284]['step'])<int(obs[0]['step'])<int(obs[-1]['step'])<int(states[284]['step'])
opsfile=w/(prefix+'-fileops.json');ops=json.loads(opsfile.read_text())
assert not list((w.parent/(prefix+'-scratch')).iterdir())
assert not any(s['Fn'] in (0x3c,0x40,0x41) or s.get('Name','').lower().startswith('dragon') for s in ops if s['Step']>=int(states[279]['step']))
done=[fields(s) for s in lines if s.startswith('DQ3_SAVE_OBSERVED ')]
assert len(done)==1 and done[0]['packet']=='288' and done[0]['step']==states[-1]['step']
report=dict(scope='正常新遊戲至288單人全體狀況、新Enter返回場景與左右行走；remake尚未驗收',meta=meta,**groups,detail_observations=detail,summary_observations=obs,artifacts=artifacts,parent_source_sha256=digest(parent_file),prefix280_unchanged=True,persistent_unchanged_except_left_step=True,normal_inputs=326,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,fileops_sha256=digest(opsfile),log_sha256=digest(w/(prefix+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False,return_verified=True)
p=w/(prefix+'-source-r1-receipt.json');assert not p.exists()
with p.open('x') as f:json.dump(report,f,ensure_ascii=False,indent=2);f.write('\n')
print('SOURCE_PASS',len(artifacts),digest(p),meta['producer_sha256'],meta['probe_source_sha256'],meta['probe_sha256'])
