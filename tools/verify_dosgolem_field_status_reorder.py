"""正常298單人重新排序來源稽核；入口與執行契約見docs/188。"""
from pathlib import Path
import json,sys
sys.path.insert(0,'/repo/tools')
from verify_dosgolem_mother_return import digest,fields,png
from verify_dosgolem_first_move import RUNTIME_HASHES
w=Path('/work/dosgolem-opening')
prefix='issue4-field-reorder-normal-r2'
parent_prefix='issue4-field-summary-normal-r2'
parent_file=w/(parent_prefix+'-source-r1-receipt.json')
assert digest(parent_file)=='fac03247817ebd9ac9f6aa80c8bba97bc5d9d9ff0f59aba7385575a370a731e2'
parent=json.loads(parent_file.read_text())
meta=json.loads((w/(prefix+'-meta.json')).read_text())
for k in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','docker_image','build_flags'):
 assert meta[k]==parent['meta'][k],k
assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
assert '-load-state' not in meta['args']
producer=Path('/repo/tools/dosgolem_field_status_reorder_probe.py')
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
assert len(q)==len(c)==len(states)==298 and len(irq)==672
scans=[v for _,v in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in q]
assert len(scans)==336 and [int(s['port60'],16) for s in irq]==[v for scan in scans for v in (scan,scan|128)]
assert [int(s['count']) for s in irq]==list(range(1,673))
assert [int(s['scan'],16) for s in q[288:]]==[0x39,0x50,0x39,0x50,0x50,0x39,0x1c,0x1c,0x4b,0x4d]
artifacts=[]
for a in parent['artifacts']:
 old=w/a['path'];new=w/a['path'].replace(parent_prefix,prefix,1)
 assert old.stat().st_size==new.stat().st_size==a['size'] and digest(old)==digest(new)==a['sha256']
 artifacts.append(dict(path=new.name,size=a['size'],sha256=a['sha256']))
baseline=(w/(parent_prefix+'-persistent-288.bin')).read_bytes()
for n,(queued,consumed,state) in enumerate(zip(q,c,states),1):
 assert queued['packet']==consumed['packet']==state['packet']==str(n)
 assert queued['scan']==consumed['scan']==state['scan'] and queued['kind']==state['kind']
 assert int(queued['step'])==int(state['queued_step'])<int(consumed['step'])<int(state['step'])
 assert int(consumed['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
 if n>1:assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
 if n<=288:continue
 expected=bytearray(baseline)
 if n==297:expected[10]=2
 assert (w/(prefix+f'-persistent-{n:03d}.bin')).read_bytes()==expected
 assert bytes.fromhex(state['actor'])==baseline[342:470]
 assert all(state[k]==states[287][k] for k in ('flags','gold_lo','gold_hi','player_y','raw0b24'))
 assert state['last_record']==('522' if n>=294 else states[287]['last_record'])
 assert int(state['player_x'])==(2 if n==297 else 3)
 label=prefix+f'-packet-{n:03d}-'+state['phase']
 image,indexed=w/(label+'.png'),w/(label+'.bin')
 width,height,pixels,_=png(image)
 assert (width,height)==(640,350) and pixels==indexed.read_bytes()
 for p in (image,indexed,w/(prefix+f'-persistent-{n:03d}.bin')):artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
assert len(artifacts)==702
assert [(s['phase'],int(s['choice_count']),int(s['choice_cursor'])) for s in states[288:]]==[('choice',6,1),('choice',6,2),('choice',3,1),('choice',3,2),('choice',3,3),('inline_wait',3,3),('waiting',3,3),('ready',6,2),('ready',6,2),('ready',6,2)]
assert all(s['clock']=='30' and s['raw526c']=='1' for s in groups['clocks'])
detail=[fields(s) for s in lines if s.startswith('DQ3_FIELD_DETAIL_OBSERVE ')]
assert detail[:len(parent['detail_observations'])]==parent['detail_observations']
assert all(int(s['packet'])>288 for s in detail[len(parent['detail_observations']):])
summary=[fields(s) for s in lines if s.startswith('DQ3_FIELD_SUMMARY_OBSERVE ')]
assert summary[:len(parent['summary_observations'])]==parent['summary_observations']
assert all(int(s['packet'])>288 for s in summary[len(parent['summary_observations']):])
obs=[fields(s) for s in lines if s.startswith('DQ3_FIELD_REORDER_OBSERVE ')]
assert [int(s['ida_linear'],16) for s in obs]==[0x18685,0x18694,0x15023,0x21414,0x2111b]
raw=Path('/repo/assets_raw/DQ3.EXE').read_bytes()
assert len(raw)==meta['original_size'] and digest(Path('/repo/assets_raw/DQ3.EXE'))==meta['original_sha256']
assert all(s['party_count']=='1' and s['window3e6e']==raw[0x19fae:0x19fae+24].hex() for s in obs)
assert all(s['packet']=='294' for s in obs[:4]) and obs[4]['packet']=='295'
assert obs[2]['DI']==obs[3]['DI']=='020a' and obs[3]['SI']=='3e6e'
assert int(c[293]['step'])<int(obs[0]['step'])<int(obs[3]['step'])<int(states[293]['step'])
assert int(c[294]['step'])<int(obs[4]['step'])<int(states[294]['step'])
assert states[293]['ida_linear']=='216d8' and states[294]['ida_linear']=='21133'
assert all(s['ida_linear']=='1997c' for s in states[295:])
opsfile=w/(prefix+'-fileops.json');ops=json.loads(opsfile.read_text())
assert not list((w.parent/(prefix+'-scratch')).iterdir())
assert not any(s['Fn'] in (0x3c,0x40,0x41) or s.get('Name','').lower().startswith('dragon') for s in ops if s['Step']>=int(states[287]['step']))
done=[fields(s) for s in lines if s.startswith('DQ3_SAVE_OBSERVED ')]
assert len(done)==1 and done[0]['packet']=='298' and done[0]['step']==states[-1]['step']
report=dict(scope='正常新遊戲至298單人重新排序TXT00/522兩頁、新Enter換頁與返回場景及左右行走；remake尚未驗收',meta=meta,**groups,detail_observations=detail,summary_observations=summary,reorder_observations=obs,artifacts=artifacts,parent_source_sha256=digest(parent_file),prefix288_unchanged=True,persistent_unchanged_except_left_step=True,normal_inputs=336,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,fileops_sha256=digest(opsfile),log_sha256=digest(w/(prefix+'.log')),checker_sha256=digest(Path(__file__)),original_record=522,original_text_sha256=digest(Path('/repo/assets_raw/D3TXT00.TXT')),remake_parity=False,full_rgb_parity=False,return_verified=True)
p=w/(prefix+'-source-r1-receipt.json');assert not p.exists()
with p.open('x') as f:json.dump(report,f,ensure_ascii=False,indent=2);f.write('\n')
print('SOURCE_PASS',len(artifacts),digest(p),meta['producer_sha256'],meta['probe_source_sha256'],meta['probe_sha256'])
