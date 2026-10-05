from pathlib import Path
import hashlib, json, sys
sys.path.insert(0, '/repo/tools')
from verify_dosgolem_mother_return import fields, png
w = Path('/work/dosgolem-opening')
prefix = 'issue4-equip-entry-normal-r3'
parent_prefix = 'issue4-spell-empty-normal-r1'
digest = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
parent_file = w / (parent_prefix + '-source-r1-receipt.json')
assert digest(parent_file) == 'c096702865ce694706ba2aac2f9333408b895bf45d856bae4a14334d1275cfcf'
parent = json.loads(parent_file.read_text())
meta = json.loads((w / (prefix + '-meta.json')).read_text())
for key in ('original_size','original_sha256','upstream_revision','seed','normal_prefix_inputs','generator_sha256','build_flags'):
    assert meta[key] == parent['meta'][key], key
assert meta['seed_configured_before_execution'] and not meta['state_restore'] and not meta['gameplay_state_injection']
assert digest(Path('/repo/tools/dosgolem_field_equipment_probe.py')) == meta['producer_sha256']
for suffix,key in [('-probe-source.go','probe_source_sha256'),('-probe','probe_sha256')]:
    assert digest(w / (prefix + suffix)) == meta[key]
lines = (w / (prefix + '.log')).read_text().splitlines()
assert '沒實作的服務（0 種）：' in lines
assert len([s for s in lines if s.startswith('DQ3_CREATION_SEED ')]) == 1
groups = {}
for key,tag in [('queued','DQ3_QUIESCENT_QUEUED'),('consumed','DQ3_QUIESCENT_INPUT'),('states','DQ3_QUIESCENT_CAPTURE'),('actual_irq1_events','DQ3_KEY_DELIVERED'),('clocks','DQ3_COMMAND_CLOCK'),('rasters','DQ3_COMMAND_RASTER'),('item_rasters','DQ3_ITEM_RASTER'),('writer_events','DQ3_ITEM_REORDER_WRITER'),('text_events','DQ3_EMPTY_RECRUIT_TEXT'),('observations','DQ3_ITEM_DROP_OBSERVE'),('status_rasters','DQ3_STATUS_RASTER'),('detail_observations','DQ3_FIELD_DETAIL_OBSERVE'),('summary_observations','DQ3_FIELD_SUMMARY_OBSERVE'),('reorder_observations','DQ3_FIELD_REORDER_OBSERVE'),('spell_empty_observations','DQ3_FIELD_SPELL_EMPTY_OBSERVE')]:
    rows = [fields(s) for s in lines if s.startswith(tag + ' ')]
    assert rows[:len(parent[key])] == parent[key], key
    groups[key] = rows
assert len(groups['queued']) == len(groups['consumed']) == len(groups['states']) == 314
assert len(groups['actual_irq1_events']) == 704
assert [s['scan'] for s in groups['queued'][304:]] == ['39','50','50','39','01','01','01','01','4b','4d']
assert [(s['phase'],s['choice_count'],s['choice_cursor']) for s in groups['states'][304:]] == [('choice','6','1'),('choice','6','2'),('choice','6','3'),('choice','3','1'),('choice','4','1'),('choice','1','1'),('choice','1','1'),('ready','6','3'),('ready','6','3'),('ready','6','3')]
artifacts = []
for a in parent['artifacts']:
    p = w / a['path'].replace(parent_prefix,prefix,1)
    assert p.stat().st_size == a['size'] and digest(p) == a['sha256']
    artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
base = (w / (parent_prefix + '-persistent-304.bin')).read_bytes()
for n in range(305,315):
    state = groups['states'][n-1]
    p = w / (prefix + f'-persistent-{n:03d}.bin')
    expected=bytearray(base)
    if n==313:expected[10]=2
    assert p.read_bytes() == expected
    image = w / (prefix + f'-packet-{n:03d}-' + state['phase'] + '.png')
    indexed = image.with_suffix('.bin')
    width,height,pixels,_ = png(image)
    assert (width,height) == (640,350) and pixels == indexed.read_bytes()
    for p in (p,image,indexed):
        artifacts.append(dict(path=p.name,size=p.stat().st_size,sha256=digest(p)))
assert len(artifacts) == 750
done=[fields(s) for s in lines if s.startswith('DQ3_SAVE_OBSERVED ')]
assert len(done)==1 and done[0]['packet']=='314' and done[0]['step']==groups['states'][-1]['step']
scans=[v for _,v in meta['normal_prefix_inputs']]+[int(s['scan'],16) for s in groups['queued']]
assert len(scans)==352 and [int(s['port60'],16) for s in groups['actual_irq1_events']]==[v for scan in scans for v in (scan,scan|128)]
assert [int(s['count']) for s in groups['actual_irq1_events']]==list(range(1,705))
for n,(queued,consumed,state) in enumerate(zip(groups['queued'],groups['consumed'],groups['states']),1):
    assert queued['packet']==consumed['packet']==state['packet']==str(n)
    assert queued['scan']==consumed['scan']==state['scan'] and queued['kind']==state['kind']
    assert int(queued['step'])==int(state['queued_step'])<int(consumed['step'])<int(state['step'])
    assert int(consumed['irqs'])==76+n*2-1 and int(groups['actual_irq1_events'][76+n*2-1]['step'])<=int(state['step'])
    if n>1:assert all(queued[k]==groups['states'][n-2][k] for k in ('step','phase','ida_linear'))
    if n>=305:
        assert bytes.fromhex(state['actor'])==base[342:470]
        assert all(state[k]==groups['states'][303][k] for k in ('flags','gold_lo','gold_hi','player_y','raw0b24','last_record'))
        assert int(state['player_x'])==(2 if n==313 else 3)
        assert state['ida_linear']==('1f7b7' if n<=311 else '1997c')
opsfile=w/(prefix+'-fileops.json')
ops=json.loads(opsfile.read_text())
assert not any(s['Fn'] in (0x3c,0x40,0x41) or s.get('Name','').lower().startswith('dragon') for s in ops if s['Step']>=int(groups['states'][303]['step']))

obs = [fields(s) for s in lines if s.startswith('DQ3_FIELD_EQUIP_ENTRY_OBSERVE ')]
assert obs[0]['ida_linear'] == '17e12' and obs[0]['packet'] == '308'
assert not list((w.parent / (prefix + '-scratch')).iterdir())
report = dict(scope='normal new-game314 single-member four equipment slots skipped via Esc; native return and next moves verified; remake RED',meta=meta,**groups,artifacts=artifacts,equip_observations=obs,parent_source_sha256=digest(parent_file),prefix304_artifacts_unchanged=True,persistent_unchanged_except_left_step=True,normal_inputs=352,return_verified=True,fileops_sha256=digest(opsfile),remake_parity=False,full_rgb_parity=False,log_sha256=digest(w/(prefix+'.log')),checker_sha256=digest(Path(__file__)))
p = w / (prefix + '-source-r1-receipt.json')
assert not p.exists()
p.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print('SOURCE_PASS',len(artifacts),digest(p))
