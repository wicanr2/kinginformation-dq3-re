"""從前一乾淨九份 JSON 重建單人隊伍改名契約；證據入口 docs/188。"""
import json
from pathlib import Path
import sys

root=Path(sys.argv[1])
paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() for p in paths)
raw=[p.read_text() for p in paths]
objects=[json.loads(s) for s in raw]
assert all(d['schema_version']=='0.13.0' for d in objects)
assert objects[0]['content_version']=='0.1.85'
for d in objects:d['schema_version']='0.14.0'
objects[0]['content_version']='0.1.86'
interface=objects[paths.index(root/'data/interface.json')]
selection=interface['recruitment_selection']
assert 'view_rename' not in selection
selection['view_rename']={
    'input_action':'rename','target_scope':'singleton_party_leader','require_nonempty_name':True,
    'evidence':{
        'level':'D3','source_kind':'exe',
        'source':'DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c + dosgolem rename entry fe3df116, success631414a8, cancelc139c899',
        'address_space':'linear','address':'IDA9.4 linear10686..106DB;1885F..1886F;10D17..10DA3;DGROUP5077/722/726/4F15',
        'consumer':'second read-key K -> singleton party leader -> shared nonempty name input -> 18-byte name copy or cancel ->540 No ->541 ->field',
        'doc':'docs/188-opening-escort-to-castle-spec.md',
        'note':'有限單人隊伍READY；觀看名冊與520B副本不改。多角色隊伍選人未READY，缺資料拒絕；不宣稱原版存讀檔或完整V3。'
    }
}
for p,d,s in zip(paths,objects,raw):
    assert p.stat().st_uid==1000,p
    old='"schema_version": "0.13.0"'
    assert s.count(old)==1,p
    s=s.replace(old,'"schema_version": "0.14.0"')
    if p.name=='manifest.json':
        old='"content_version": "0.1.85"'
        assert s.count(old)==1
        s=s.replace(old,'"content_version": "0.1.86"')
    if p.name=='interface.json':
        start=s.index('"recruitment_selection": {')
        point=s.index('    "evidence": {',start)
        fragment=json.dumps(selection['view_rename'],ensure_ascii=False,indent=2).splitlines()
        block='    "view_rename": '+fragment[0]+'\n'+'\n'.join('    '+line for line in fragment[1:])+',\n'
        s=s[:point]+block+s[point:]
    assert json.loads(s)==d,p
    p.write_text(s)
    assert p.stat().st_uid==1000,p
print('九份JSON重建：schema0.14.0 / content0.1.86')
