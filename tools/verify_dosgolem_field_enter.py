from pathlib import Path
import ast,json,sys
sys.path.insert(0,'/repo/tools')
from verify_dosgolem_mother_return import digest,fields,png
from verify_dosgolem_first_move import RUNTIME_HASHES
root=Path('/work/dosgolem-opening');pre='issue4-command-enter-first-normal-r1';pp='issue4-talk-empty-return-normal-r1'
parent_path=root/(pp+'-source-r1-receipt.json')
assert digest(parent_path)=='753da910efcb614851eb707a8e35bd22dfc1e978a9347ba078a067208ff197eb'
parent=json.loads(parent_path.read_text());meta=json.loads((root/(pre+'-meta.json')).read_text())
for k in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','build_flags','docker_image'):assert meta[k]==parent['meta'][k],k
assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
assert digest(Path('/repo/tools/dosgolem_field_enter_probe.py'))==meta['producer_sha256']
for suffix,key in [('-probe-source.go','probe_source_sha256'),('-probe','probe_sha256')]:assert digest(root/(pre+suffix))==meta[key]
assert meta['producer_sha256']=='e7916f125f27c16d39efe5083501feb8bc59fb24ee95ff806c85fbf325ab9c6b'
assert meta['probe_source_sha256']=='3ec57d1023ba9e1ff2f83accd2d7b7c3e70960e52a37e1473a2db0b9bca79ea3'
assert meta['probe_sha256']=='a02f301c0ecdb12b7046bfb6b7999ed8e2cfa27ac0d0f72842b8909f7024d227'
exe=Path('/repo/assets_raw/DQ3.EXE');assert exe.stat().st_size==meta['original_size'] and digest(exe)==meta['original_sha256']
tree=ast.parse(Path('/repo/tools/verify_dosgolem_field_equipment_wear.py').read_text());tags=next(ast.literal_eval(n.value) for n in tree.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='tags' for t in n.targets))
tags.update(wear_observations='DQ3_FIELD_EQUIP_WEAR_OBSERVE',native_events='DQ3_ARMOR_NATIVE',search_native_events='DQ3_SEARCH_NATIVE')
groups={k:[] for k in tags};talk=[];seed=[];done=[];supported=False
with (root/(pre+'.log')).open() as f:
 for line in f:
  assert not line.startswith('DQ3_SAVE_BOUND ') and '找不到的檔（' not in line
  supported|=line.strip()=='沒實作的服務（0 種）：'
  for k,tag in tags.items():
   if line.startswith(tag+' '):groups[k].append(fields(line));break
  if line.startswith('DQ3_TALK_NATIVE '):talk.append(fields(line))
  if line.startswith('DQ3_CREATION_SEED '):seed.append(fields(line))
  if line.startswith('DQ3_SAVE_OBSERVED '):done.append(fields(line))
assert supported and len(seed)==1 and seed[0]['fixed']=='1357'
for k,rows in groups.items():assert rows[:len(parent[k])]==parent[k],k
q,c,states,irq=(groups[k] for k in ('queued','consumed','states','actual_irq1_events'))
assert len(q)==len(c)==len(states)==410 and len(irq)==896
scans=[v for _,v in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in q]
assert len(scans)==448 and [int(s['port60'],16) for s in irq]==[v for scan in scans for v in (scan,scan|128)]
assert [int(s['count']) for s in irq]==list(range(1,897))
assert [s['scan'] for s in q[409:]]==['1c'] and states[409]['phase'] in ('choice','ready','waiting')
assert states[404]['choice_count']=='6' and states[404]['choice_cursor']=='1' and states[405]['last_record']=='260' and states[405]['ida_linear']=='21133'
artifacts=[]
for a in parent['artifacts']:
 rel=a['path'].replace(pp,pre,1);path=(root if '/' not in rel else root.parent)/rel
 assert path.stat().st_size==a['size'] and digest(path)==a['sha256'],rel
 artifacts.append(dict(path=rel,size=a['size'],sha256=a['sha256']))
base=(root/(pp+'-persistent-409.bin')).read_bytes()
for n,(queued,taken,state) in enumerate(zip(q,c,states),1):
 assert queued['packet']==taken['packet']==state['packet']==str(n)
 assert queued['scan']==taken['scan']==state['scan'] and queued['kind']==state['kind']
 assert int(queued['step'])==int(state['queued_step'])<int(taken['step'])<int(state['step'])
 assert int(taken['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
 if n>1:assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
 if n<=409:continue
 path=root/(pre+f'-persistent-{n:03d}.bin');expected=bytearray(base)
 assert path.read_bytes()==expected,n
 assert bytes.fromhex(state['actor'])==base[342:470] and state['flags']==states[403]['flags'] and state['player_x']=='3' and state['player_y']=='18'
 label=pre+f'-packet-{n:03d}-'+state['phase'];im,idx=root/(label+'.png'),root/(label+'.bin');width,height,indices,_=png(im)
 assert (width,height)==(640,350) and indices==idx.read_bytes()
 for path in (im,idx,path):artifacts.append(dict(path=path.name,size=path.stat().st_size,sha256=digest(path)))
assert len(artifacts)==1040
clocks=[s for s in groups['clocks'] if int(s['packet'])>=410];assert [s['packet'] for s in clocks]==['410'] and all(s['clock']=='0' and s['step']==states[int(s['packet'])-1]['step'] for s in clocks)
ops_path=root/(pre+'-fileops.json');ops=json.loads(ops_path.read_text());old=json.loads((root/(pp+'-fileops.json')).read_text());assert ops[:len(old)]==old
assert not any(o['Fn'] in (0x3c,0x40,0x41) for o in ops if o['Step']>=int(states[403]['step']))
assert len(done)==1 and done[0]['packet']=='410' and done[0]['step']==states[-1]['step']
assert talk[:len(parent['talk_native_events'])]==parent['talk_native_events']
report=dict(talk_native_events=talk,scope='normal409 -> first Enter410; source only; no behavior preselected, return/spec/remake unknown',meta=meta,**groups,artifacts=artifacts,parent_source_sha256=digest(parent_path),prefix409_unchanged=True,normal_inputs=448,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,persistent_transaction_verified=True,original_saved_files_unchanged=True,native_clock_limit=parent['native_clock_limit'],fileops_sha256=digest(ops_path),log_sha256=digest(root/(pre+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False)
out=root/(pre+'-public-check-r1.json');assert not out.exists();out.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print('SOURCE_ONLY_PASS',len(artifacts),digest(out))
