"""正常出生後的首次招募選單來源稽核；範圍與容器用法見 docs/188。"""
import argparse, json
from pathlib import Path
from verify_dosgolem_registry_birth import validate as validate_birth
from verify_dosgolem_mother_return import fields, png, digest

PREFIX='issue4-recruit-normal-r4'
RECEIPT_HASH='a85cad67ab6611291b82824973ec39de0e2192f4ad9808b0f362bd18a6e4fda9'

def source_receipt(root):
    prefix = 'issue4-recruit-normal-r4'
    parent = validate_birth(root/'issue4-registry-birth-normal-r2-source-r1-receipt.json')
    meta = json.loads((root/(prefix+'-meta.json')).read_text())
    old_meta = json.loads((root/'issue4-registry-birth-normal-r2-meta.json').read_text())
    for k in ('original_size','original_sha256','upstream_revision','seed','seed_configured_before_execution','parent_receipt_sha256','generator_sha256','build_flags','normal_prefix_inputs','state_restore','gameplay_state_injection','upstream_files_sha256','patched_files_sha256','upstream_bios_sha256','patched_bios_sha256','upstream_vga_sha256','patched_vga_sha256'):
        assert meta[k] == old_meta[k], k
    assert meta['producer_sha256']==digest(Path(__file__).with_name('dosgolem_recruitment_entry_probe.py'))
    assert meta['state_restore'] is False and meta['gameplay_state_injection'] is False
    for suffix, key in [('-probe-source.go','probe_source_sha256'),('-probe','probe_sha256')]:
        assert digest(root/(prefix+suffix)) == meta[key]
    lines = (root/(prefix+'.log')).read_text().splitlines()
    groups = {}
    for name, tag in [('queued','DQ3_QUIESCENT_QUEUED'),('consumed','DQ3_QUIESCENT_INPUT'),('states','DQ3_QUIESCENT_CAPTURE'),('actual_irq1_events','DQ3_KEY_DELIVERED')]:
        groups[name] = [fields(l) for l in lines if l.startswith(tag+' ')]
        assert groups[name][:len(parent[name])] == parent[name], name
    queues, inputs, states, irq = (groups[k] for k in ('queued','consumed','states','actual_irq1_events'))
    assert len(queues) == len(inputs) == len(states) == 196
    route=[0x4d,0x50,0x4d,0x4d]+[0x48]*4+[0x4d]*3+[0x50]*4+[0x4b]*6+[0x48,0x1c,0x1c]
    assert [int(q['scan'],16) for q in queues[172:]]==route
    positions=[(3,5),(3,6),(4,6),(5,6),(5,5),(5,4),(5,3),(5,2),(6,2),(7,2),
        (8,14),(8,15),(8,16),(8,17),(8,18),(7,18),(6,18),(5,18),(4,18),(3,18),
        (2,18),(2,18),(2,18),(2,18)]
    assert [(int(s['player_x']),int(s['player_y'])) for s in states[172:]]==positions
    assert [s['raw0b24'] for s in states[172:]]==['111d']*10+['000c']*14
    assert [s['phase'] for s in states[172:]]==['ready']*22+['inline_wait','choice']
    assert [s['last_record'] for s in states[172:]]==['560']*22+['527','528']
    assert all(q['stage']==c['stage']==s['stage']=='5' for q,c,s in zip(queues[172:],inputs[172:],states[172:]))
    assert len(irq) == (38+len(queues))*2
    assert [int(x['count']) for x in irq] == list(range(1,len(irq)+1))
    scans = [s for _,s in meta['normal_prefix_inputs']]+[int(q['scan'],16) for q in queues]
    assert [int(x['port60'],16) for x in irq] == [v for s in scans for v in (s,s|128)]
    artifacts = []
    for n, (q,c,s) in enumerate(zip(queues,inputs,states),1):
        assert int(q['packet']) == int(c['packet']) == int(s['packet']) == n
        assert q['scan'] == c['scan'] == s['scan'] and q['kind'] == s['kind']
        assert int(q['step']) == int(s['queued_step']) < int(c['step']) < int(s['step'])
        assert int(c['irqs']) == 76+n*2-1
        assert int(irq[76+n*2-1]['step']) <= int(s['step'])
        assert int(s['raw0013'],16)&0x4000 == 0
        assert s['phase'] in ('ready','inline_wait','waiting','choice','name')
        assert s['ida_linear']=={'ready':'1997c','inline_wait':'216d8','waiting':'21133','choice':'1f7b7','name':'11096'}[s['phase']]
        if n>1:
            assert q['step'] == states[n-2]['step'] and q['phase'] == states[n-2]['phase']
        name = prefix+f'-packet-{n:03d}-'+s['phase']
        width,height,indices,_ = png(root/(name+'.png'))
        assert (width,height) == (640,350) and indices == (root/(name+'.bin')).read_bytes()
        for ext in ('.png','.bin'):
            p = root/(name+ext)
            artifacts.append({'path':p.name,'size':p.stat().st_size,'sha256':digest(p)})
            if n<=172:
                prior = root/('issue4-registry-birth-normal-r2'+f'-packet-{n:03d}-'+s['phase']+ext)
                assert p.read_bytes() == prior.read_bytes()
    count = 0
    for p in root.glob(prefix+'-*'):
        if p.suffix not in ('.png','.bin') or '-packet-' in p.name or '-event-' in p.name:
            continue
        prior = root/p.name.replace(prefix,'issue4-registry-birth-normal-r2',1)
        assert prior.is_file() and p.read_bytes() == prior.read_bytes()
        count += 1
    assert count == 174, count
    done = [fields(l) for l in lines if l.startswith('DQ3_RECRUIT_DONE ')]
    assert len(done) == 1 and done[0]['step'] == states[-1]['step']
    assert states[-1]['phase'] == 'choice' and states[-1]['choice_count'] == '3'
    assert states[-1]['choice_cursor'] == '1' and states[-1]['last_record'] == '528'
    obs = [fields(l) for l in lines if l.startswith('DQ3_RECRUIT_STATE ')]
    assert len(obs) == len(states)-171
    assert all(o['roster_flags'] == '020100000000000000000000' for o in obs)
    assert len({o['slot1'] for o in obs}) == 1 and len(bytes.fromhex(obs[0]['slot1'])) == 97
    writers = [fields(l) for l in lines if l.startswith('DQ3_BIRTH_WRITER ')]
    copy = next(w for w in writers if w['ida_linear']=='10a9f')
    assert obs[0]['slot1'] == copy['candidate'][:194]
    for s in states[172:]:
        assert s['gold_lo']=='0032' and s['gold_hi']=='0000' and s['flags']==states[171]['flags']
    report = {'original_sha256':meta['original_sha256'],'upstream_revision':meta['upstream_revision'],'seed':meta['seed'],'seed_control_once':True,'state_injection':False,'scope':'正常冷啟動、登錄一名戰士男性、正常下樓至首次招募選單；原版來源審查',
        'normal_inputs':38+len(states),'parent_birth_sha256':digest(root/'issue4-registry-birth-normal-r2-source-r1-receipt.json'),
        'meta':meta,'log_sha256':digest(root/(prefix+'.log')),'artifacts':artifacts,
        **groups,'recruit_observations':obs,'done':done[0],
        'parent_174_png_bin_unchanged':True,'birth_172_packets_344_png_bin_unchanged':True,
        'slot1_97_bytes_equal_birth_writer':True,'recruitment_transaction':'not attempted',
        'remake_parity':False,'audio_parity':False,'original_save_load_parity':False}
    return report

def validate(path):
    assert path.name==PREFIX+'-source-r2-receipt.json'
    assert digest(path)==RECEIPT_HASH
    data=json.loads(path.read_text())
    assert data==source_receipt(path.parent)
    return data

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('receipt',type=Path)
    parser.add_argument('--emit',action='store_true')
    args=parser.parse_args()
    if args.emit:
        assert not args.receipt.exists()
        text=json.dumps(source_receipt(args.receipt.parent),ensure_ascii=False,indent=2)+'\n'
        import hashlib
        assert hashlib.sha256(text.encode()).hexdigest()==RECEIPT_HASH
        args.receipt.write_text(text)
    d=validate(args.receipt)
    print('原版首次招募選單來源 PASS',d['normal_inputs'],len(d['actual_irq1_events']),len(d['artifacts']))

if __name__=='__main__':main()
