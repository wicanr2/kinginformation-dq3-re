from pathlib import Path
import ast,json,sys
sys.path.insert(0,'/repo/tools')
from verify_dosgolem_mother_return import digest,fields,png
from verify_dosgolem_first_move import RUNTIME_HASHES
root=Path('/work/dosgolem-opening');pre='issue4-talk-empty-first-normal-r2';pp='issue4-search-return-normal-r1'
parent_path=root/(pp+'-source-r1-receipt.json')
assert digest(parent_path)=='3af2954a524f521670aad308377efd2aca4c28633a971c111f7af6fc709c44eb'
parent=json.loads(parent_path.read_text());meta=json.loads((root/(pre+'-meta.json')).read_text())
for k in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','build_flags','docker_image'):assert meta[k]==parent['meta'][k],k
assert all(meta[k]==v for k,v in RUNTIME_HASHES.items())
assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
assert digest(Path('/work/issue4-talk-empty-first-probe-r2.py'))==meta['producer_sha256']=='569dbc500bcecf7d7ad679e6f60e3f260fdf510f32cb59dd2f1af4f72a6b2105'
for suffix,key,value in [('-probe-source.go','probe_source_sha256','07495af1f84fc54f2a56143445399fe5bbb3c366e6ad160a5cf86bee9cffccb3'),('-probe','probe_sha256','bb406d7e927528b08d2f45a37803d511bb40f656b5341cfb855e1898c9574501')]:assert digest(root/(pre+suffix))==meta[key]==value
exe=Path('/repo/assets_raw/DQ3.EXE');assert exe.stat().st_size==meta['original_size'] and digest(exe)==meta['original_sha256']
tree=ast.parse(Path('/repo/tools/verify_dosgolem_field_equipment_wear.py').read_text());tags=next(ast.literal_eval(n.value) for n in tree.body if isinstance(n,ast.Assign) and any(isinstance(t,ast.Name) and t.id=='tags' for t in n.targets))
tags.update(wear_observations='DQ3_FIELD_EQUIP_WEAR_OBSERVE',native_events='DQ3_ARMOR_NATIVE',search_native_events='DQ3_SEARCH_NATIVE')
groups={k:[] for k in tags};seed=[];done=[];supported=False
with (root/(pre+'.log')).open() as f:
 for line in f:
  assert not line.startswith('DQ3_SAVE_BOUND ') and '找不到的檔（' not in line
  supported|=line.strip()=='沒實作的服務（0 種）：'
  for k,tag in tags.items():
   if line.startswith(tag+' '):groups[k].append(fields(line));break
  if line.startswith('DQ3_CREATION_SEED '):seed.append(fields(line))
  if line.startswith('DQ3_SAVE_OBSERVED '):done.append(fields(line))
assert supported and len(seed)==1 and seed[0]['fixed']=='1357'
for k,rows in groups.items():assert rows[:len(parent[k])]==parent[k],k
q,c,states,irq=(groups[k] for k in ('queued','consumed','states','actual_irq1_events'))
assert len(q)==len(c)==len(states)==406 and len(irq)==888
scans=[v for _,v in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in q]
assert len(scans)==444 and [int(s['port60'],16) for s in irq]==[v for scan in scans for v in (scan,scan|128)]
assert [int(s['count']) for s in irq]==list(range(1,889))
assert [s['scan'] for s in q[404:]]==['39','39'] and [s['phase'] for s in states[404:]]==['choice','waiting']
assert states[404]['choice_count']=='6' and states[404]['choice_cursor']=='1' and states[405]['last_record']=='260' and states[405]['ida_linear']=='21133'
artifacts=[]
for a in parent['artifacts']:
 rel=a['path'].replace(pp,pre,1);path=(root if '/' not in rel else root.parent)/rel
 assert path.stat().st_size==a['size'] and digest(path)==a['sha256'],rel
 artifacts.append(dict(path=rel,size=a['size'],sha256=a['sha256']))
base=(root/(pp+'-persistent-404.bin')).read_bytes()
for n,(queued,taken,state) in enumerate(zip(q,c,states),1):
 assert queued['packet']==taken['packet']==state['packet']==str(n)
 assert queued['scan']==taken['scan']==state['scan'] and queued['kind']==state['kind']
 assert int(queued['step'])==int(state['queued_step'])<int(taken['step'])<int(state['step'])
 assert int(taken['irqs'])==76+n*2-1 and int(irq[76+n*2-1]['step'])<=int(state['step'])
 if n>1:assert all(queued[k]==states[n-2][k] for k in ('step','phase','ida_linear'))
 if n<=404:continue
 path=root/(pre+f'-persistent-{n:03d}.bin');assert path.read_bytes()==base
 assert bytes.fromhex(state['actor'])==base[342:470] and state['flags']==states[403]['flags'] and state['player_x']=='3' and state['player_y']=='18'
 label=pre+f'-packet-{n:03d}-'+state['phase'];im,idx=root/(label+'.png'),root/(label+'.bin');width,height,indices,_=png(im)
 assert (width,height)==(640,350) and indices==idx.read_bytes()
 for path in (im,idx,path):artifacts.append(dict(path=path.name,size=path.stat().st_size,sha256=digest(path)))
assert len(artifacts)==1028
clocks=[s for s in groups['clocks'] if int(s['packet'])>=405];assert [s['packet'] for s in clocks]==['405','406'] and all(s['clock']=='0' and s['step']==states[int(s['packet'])-1]['step'] for s in clocks)
ops_path=root/(pre+'-fileops.json');ops=json.loads(ops_path.read_text());old=json.loads((root/(pp+'-fileops.json')).read_text());assert ops[:len(old)]==old
assert not any(o['Fn'] in (0x3c,0x40,0x41) for o in ops if o['Step']>=int(states[403]['step']))
assert len(done)==1 and done[0]['packet']=='406' and done[0]['step']==states[-1]['step']
report=dict(scope='normal404 -> first no-target Talk406; source only, spec and remake unverified',meta=meta,**groups,artifacts=artifacts,parent_source_sha256=digest(parent_path),prefix404_unchanged=True,normal_inputs=444,seed_control_once=True,state_injection=False,emulator_snapshot_restore=False,persistent_transaction_verified=True,original_saved_files_unchanged=True,native_clock_limit=parent['native_clock_limit'],fileops_sha256=digest(ops_path),log_sha256=digest(root/(pre+'.log')),checker_sha256=digest(Path(__file__)),remake_parity=False,full_rgb_parity=False)
out=root/(pre+'-source-r1-receipt.json');assert not out.exists();out.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print('SOURCE_ONLY_PASS',len(artifacts),digest(out))
