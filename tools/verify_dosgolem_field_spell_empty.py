"""正常304空咒文原版來源稽核；入口docs/188。"""
from pathlib import Path
import json,sys
sys.path.insert(0,'/repo/tools')
from verify_dosgolem_mother_return import digest,fields,png
from verify_dosgolem_first_move import RUNTIME_HASHES
w=Path('/work/dosgolem-opening'); prefix='issue4-spell-empty-normal-r1'; parent_prefix='issue4-field-reorder-normal-r2'
parent_file=w/(parent_prefix+'-source-r1-receipt.json')
assert digest(parent_file)=='e06a5e419d255303140c1516af0aa09021bbb967e94b0db1e0db89d1aca90370'
parent=json.loads(parent_file.read_text());meta=json.loads((w/(prefix+'-meta.json')).read_text())
for k in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','docker_image','build_flags'):assert meta[k]==parent['meta'][k],k
assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection'] and '-load-state' not in meta['args']
assert digest(Path('/repo/tools/dosgolem_field_spell_empty_probe.py'))==meta['producer_sha256']
for suffix,key in [('-probe-source.go','probe_source_sha256'),('-probe','probe_sha256')]:assert digest(w/(prefix+suffix))==meta[key]
lines=(w/(prefix+'.log')).read_text().splitlines()
assert '沒實作的服務（0 種）：' in lines and not any('DQ3_SAVE_BOUND ' in s or '找不到的檔（' in s for s in lines)
seeds=[fields(s) for s in lines if s.startswith('DQ3_CREATION_SEED ')]
assert len(seeds)==1 and seeds[0]['fixed']=='1357'
tags=dict(queued='DQ3_QUIESCENT_QUEUED',consumed='DQ3_QUIESCENT_INPUT',states='DQ3_QUIESCENT_CAPTURE',actual_irq1_events='DQ3_KEY_DELIVERED',clocks='DQ3_COMMAND_CLOCK',rasters='DQ3_COMMAND_RASTER',item_rasters='DQ3_ITEM_RASTER',writer_events='DQ3_ITEM_REORDER_WRITER',text_events='DQ3_EMPTY_RECRUIT_TEXT',observations='DQ3_ITEM_DROP_OBSERVE',status_rasters='DQ3_STATUS_RASTER',detail_observations='DQ3_FIELD_DETAIL_OBSERVE',summary_observations='DQ3_FIELD_SUMMARY_OBSERVE',reorder_observations='DQ3_FIELD_REORDER_OBSERVE')
groups={k:[fields(s) for s in lines if s.startswith(t+' ')] for k,t in tags.items()}
for k,rows in groups.items():assert rows[:len(parent[k])]==parent[k],k
q,c,states,irq=(groups[k] for k in ['queued','consumed','states','actual_irq1_events'])
assert len(q)==len(c)==len(states)==304 and len(irq)==684
scans=[v for _,v in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in q]
assert len(scans)==342 and [int(s['port60'],16) for s in irq]==[v for scan in scans for v in (scan,scan|128)]
assert [int(s['count']) for s in irq]==list(range(1,685))
assert [int(s['scan'],16) for s in q[298:]]==[0x39,0x4d,0x39,0x1c,0x4b,0x4d]
artifacts=[]
for a in parent['artifacts']:
 old=w/a['path'];new=w/a['path'].replace(parent_prefix,prefix,1)
 assert old.stat().st_size==new.stat().st_size==a['size'] and digest(old)==digest(new)==a['sha256']
 artifacts.append(dict(path=new.name,size=a['size'],sha256=a['sha256']))
baseline=(w/(parent_prefix+'-persistent-298.bin')).read_bytes()
for n,(queued,consumed,state) in enumerate(zip(q,c,states),1):
 assert queued['packet']==consumed['packet']==state['packet']==str(n)
 assert queued['scan']==consumed['scan']==state['scan'] and queued['kind']==state['kind']
 assert int(queued['step'])==int(state['queued_step'])<int(consumed['step'])<int(state['step'])
 assert int(consumed['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
 if n>1:assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
 if n<=298:continue
 expected=bytearray(baseline)
 if n==303:expected[10]=2
 assert (w/(prefix+f'-persistent-{n:03d}.bin')).read_bytes()==expected
 assert bytes.fromhex(state['actor'])==baseline[342:470]
 assert all(state[k]==states[297][k] for k in ('flags','gold_lo','gold_hi','player_y','raw0b24'))
 assert state['last_record']==('262' if n>=301 else states[297]['last_record'])
 assert int(state['player_x'])==(2 if n==303 else 3)
 label=prefix+f'-packet-{n:03d}-'+state['phase'];im,idx=w/(label+'.png'),w/(label+'.bin')
 width,height,pixels,_=png(im);assert (width,height)==(640,350) and pixels==idx.read_bytes()
 for p in (im,idx,w/(prefix+f'-persistent-{n:03d}.bin')):artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
assert len(artifacts)==720
assert [(s['phase'],int(s['choice_count']),int(s['choice_cursor'])) for s in states[298:]]==[('choice',6,1),('choice',6,4),('waiting',6,1),('ready',6,4),('ready',6,4),('ready',6,4)]
assert states[300]['ida_linear']=='21133' and all(s['ida_linear']=='1997c' for s in states[301:])
obs=[fields(s) for s in lines if s.startswith('DQ3_FIELD_SPELL_EMPTY_OBSERVE ')]
assert [int(s['ida_linear'],16) for s in obs]==[0x1c9c1,0x1885f,0x18869,0x1c9ee,0x1ca00,0x1cb39,0x1c9e7,0x15023,0x21414,0x2111b]
raw=Path('/repo/assets_raw/DQ3.EXE').read_bytes()
assert len(raw)==meta['original_size'] and digest(Path('/repo/assets_raw/DQ3.EXE'))==meta['original_sha256']
assert all(s['party_count']=='1' and s['spell30']=='0' and s['spell31']=='153' and s['window3e6e']==raw[0x19fae:0x19fae+24].hex() and s['packet']=='301' for s in obs)
assert obs[-3]['DI']==obs[-2]['DI']=='0106' and obs[-2]['SI']=='3e6e'
assert bytes.fromhex(states[300]['actor'])[0x30:0x32]==b'\x00\x00'
assert obs[4]['SI']=='507f' and obs[5]['AX']=='0000' and obs[6]['AX']=='0001'
assert int(c[300]['step'])<int(obs[0]['step'])<int(obs[-1]['step'])<int(states[300]['step'])
opsfile=w/(prefix+'-fileops.json');ops=json.loads(opsfile.read_text())
assert not list((w.parent/(prefix+'-scratch')).iterdir())
assert not any(s['Fn'] in (0x3c,0x40,0x41) or s.get('Name','').lower().startswith('dragon') for s in ops if s['Step']>=int(states[297]['step']))
done=[fields(s) for s in lines if s.startswith('DQ3_SAVE_OBSERVED ')]
assert len(done)==1 and done[0]['packet']=='304' and done[0]['step']==states[-1]['step']
report=dict(scope='正常新遊戲至304單人未學咒文、TXT00/262新按鍵返回及左右行走；remake尚未驗收',meta=meta,**groups,spell_empty_observations=obs,artifacts=artifacts,parent_source_sha256=digest(parent_file),prefix298_unchanged=True,persistent_unchanged_except_left_step=True,normal_inputs=342,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,fileops_sha256=digest(opsfile),log_sha256=digest(w/(prefix+'.log')),checker_sha256=digest(Path(__file__)),original_record=262,original_text_sha256=digest(Path('/repo/assets_raw/D3TXT00.TXT')),remake_parity=False,full_rgb_parity=False,return_verified=True)
p=w/(prefix+'-source-r1-receipt.json');assert not p.exists()
with p.open('x') as f:json.dump(report,f,ensure_ascii=False,indent=2);f.write('\n')
print('SOURCE_PASS',len(artifacts),digest(p),meta['producer_sha256'],meta['probe_source_sha256'],meta['probe_sha256'])
