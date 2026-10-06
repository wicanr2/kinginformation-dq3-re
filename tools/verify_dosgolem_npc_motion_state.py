"""同一458輸入的唯讀狀態來源；逐項與已接受來源比較。"""
from pathlib import Path
import json, hashlib, sys
sys.path.insert(0, '/repo/tools')
from verify_dosgolem_mother_return import fields, digest
root=Path('/work/dosgolem-opening')
pre='issue4-npc-move-state-normal-r1'
meta=json.loads((root/(pre+'-meta.json')).read_text())
assert meta['producer_sha256']=='0fec90384ce72a4e38c5593500300c9172c060d8d84da69116591d3e78e2853b' and meta['probe_source_sha256']=='f4ec495afa57d001d181c165bc9141643dc246fd781eca8a6e646bd494186321' and meta['probe_sha256']=='4499c9a19e2e5b233c26629e92f307c4985926625a750c7c1b9631e01f77bce2'
s=Path('/repo/tools/verify_dosgolem_npc_motion_continue.py').read_text()
s=s.replace('issue4-npc-move-continue-normal-r1',pre).replace('issue4-npc-move-continue-probe-r1.py','issue4-npc-move-state-probe-r1.py')
for old,key in [('eac915fed3355b670dd1a136c44886542cd6acfb687f75bcd8fb045be8c5f51a','producer_sha256'),('1a885912bcb2193099b8241b39eac370a11915029d0513d1b5c99269efe02cc0','probe_source_sha256'),('f51cb34bec434548754d096d7810821d10616b3517756abf1def43b56ef24c14','probe_sha256')]:
    s=s.replace(old,meta[key])
s=s.replace("print('SOURCE_ONLY_PASS', len(artifacts), digest(out), observed)","print('READONLY_SOURCE_PASS', len(artifacts), digest(out))")
ns={'__file__':__file__,'__name__':'readonly_source_check'}
exec(compile(s,__file__,'exec'),ns)
report=ns['report'];out=ns['out']
parent=root/'issue4-npc-move-continue-normal-r1-source-r1-receipt.json'
assert digest(parent)=='56ef662ad02931486a60bbfd0820a518887e699320ec1c3ff131070734ba95be'
old=json.loads(parent.read_text())
for k in ns['groups']:
    assert report[k]==old[k],k
assert report['mover_events']==old['mover_events']
assert report['talk_native_events']==old['talk_native_events']
assert report['fileops_sha256']==old['fileops_sha256']
for a,b in zip(report['artifacts'],old['artifacts']):
    assert a['path'].replace(pre,'issue4-npc-move-continue-normal-r1',1)==b['path']
    assert a['size']==b['size'] and a['sha256']==b['sha256']
assert len(report['artifacts'])==len(old['artifacts'])==1184
del old
states=[];tables=[]
with (root/(pre+'.log')).open() as f:
    for line in f:
        if line.startswith('DQ3_NPC_MOVE_STATE '):states.append(fields(line))
        if line.startswith('DQ3_NPC_TABLE_NATIVE '):tables.append(fields(line))
cases=[];entry=None;orphan=[]
for r in states:
    assert r['DS']=='15ed' and int(r['count'])==14
    slots=bytes.fromhex(r['npc_slots'])
    assert len(slots)==8*int(r['count'])
    idx=int(r['slot']);assert 0<=idx<14
    assert int(r['SI'],16)==0xb66+idx*8
    if r['ida_linear']=='12025':
        assert entry is None
        entry=r
        assert (int(r['cell_word'],16)>>8)&31==idx
        assert int(r['cell_word'],16)&0x2000
    elif r['ida_linear']=='120a6':
        if entry is None:
            orphan.append(r);continue
        assert entry['slot']==r['slot'] and entry['packet']==r['packet']
        ev=[e for e in report['mover_events'] if int(entry['step'])<=int(e['step'])<=int(r['step'])]
        state=int(entry['seed'],16)
        bounds={'12043':4,'12065':4,'12074':20,'120b0':10}
        for e in ev:
            if e['ida_linear'] in bounds:
                state=(state+0x9018)&65535;state=((state<<3)|(state>>13))&65535
                assert int(e['seed'],16)==state
                assert int(e['DX'],16)==state%bounds[e['ida_linear']]
                assert int(e['AX'],16)==state//bounds[e['ida_linear']]
        assert state==int(r['seed'],16)
        before=bytes.fromhex(entry['npc_slots']);after=bytes.fromhex(r['npc_slots'])
        assert all(before[i]==after[i] for i in range(len(before)) if i not in (idx*8,idx*8+1,idx*8+3,idx*8+6))
        turns=[e for e in ev if e['ida_linear']=='1207f']
        if turns:
            assert len(turns)==1
            e=turns[0];word=int(e['DI'],16);signed=word if word<32768 else word-65536
            delta=1 if signed<int(e['CX'],16) else -1
            assert after[idx*8+3]==(before[idx*8+3]&252)|(((int(e['AX'],16)&255)+delta)&3)
            assert before[idx*8:idx*8+2]==after[idx*8:idx*8+2]
            assert not any(e['ida_linear']=='120b0' for e in ev)
        cases.append({'entry':entry,'return':r,'events':ev})
        entry=None
assert entry is None
for duplicate in orphan:
    i=states.index(duplicate)
    previous=states[i-1]
    assert {k:v for k,v in duplicate.items() if k!='step'}=={k:v for k,v in previous.items() if k!='step'}
    assert int(duplicate['step'])-int(previous['step'])==27
assert any(e['ida_linear']=='13282' for e in tables)
assert all(e['DS']=='15ed' for e in tables)
report['checked_source_sha256']='b2fdf7818b4d8ca07d97d5796db6ac241490a6899f8352ab2396cabb3b96c9c6';report['move_state_events']=states;report['table_writer_events']=tables;report['same458_source_sha256']=digest(parent)
report['component_cases']=cases;report['duplicate_return_observations']=orphan
report['all458_events_artifacts_mover_events_unchanged']=True
out.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n')
print('READONLY_STATE_PASS',len(cases),len(tables),len(orphan),digest(out))

accepted=root/(pre+'-source-r1-receipt.json')
assert digest(accepted)==report['checked_source_sha256']
component=root/(pre+'-component-r1.json')
assert digest(component)=='6586abb198648dc27859901813f2c40cad156579b511c101db820eb7023b69b4'
compact=json.loads(component.read_text());assert compact['cases']==cases and compact['source_sha256']==digest(accepted)
print('COMPONENT_FIXTURE_PASS',len(cases),digest(component))
