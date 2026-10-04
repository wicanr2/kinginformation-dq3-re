"""schema0.16.0 pack重建空Join／單人Leave；READY與Docker用法見docs/188。"""
from pathlib import Path
import hashlib,json,struct,sys

root,assets,mapping=map(Path,sys.argv[1:])
paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==1000 for p in paths)
raw=[p.read_text() for p in paths]; objects=[json.loads(s) for s in raw]
assert all(d['schema_version']=='0.16.0' for d in objects) and objects[0]['content_version']=='0.1.88'
exe=(assets/'DQ3.EXE').read_bytes(); text=(assets/'D3TXT00.TXT').read_bytes()
assert len(exe)==115282 and hashlib.sha256(exe).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(text).hexdigest()=='38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
assert exe[0x17da:0x17dd].hex()=='bf3c01' and exe[0x1838:0x183b].hex()=='80fb01' and exe[0x192c:0x192f].hex()=='bf1e02'
end=int.from_bytes(text[:2],'little');ptrs=struct.unpack('<'+'H'*(end//2),text[:end])
assert len(ptrs)-1==759 and ptrs[-1]==len(text)
record=text[ptrs[542]:ptrs[543]]
codes=list(struct.unpack('<'+'H'*(len(record)//2),record))
assert codes==[59,60,764,502,534,123,123,564,505,592,58,65534,12,12,65531,432,411,538,147,541,542,65534,12,12,401,546,194,592,57,65535]
codes=codes[:-1];glyphs=json.loads(mapping.read_text())
value=''.join(glyphs[str(c)] if c<1476 else '\n' if c==65534 else '{名字}' for c in codes)
id='dq3:text.recruitment_selection.empty_leave'
ev={'level':'D3','source_kind':'exe','source':'DQ3.EXE SHA256 '+hashlib.sha256(exe).hexdigest()+'; D3TXT00.TXT SHA256 '+hashlib.sha256(text).hexdigest()+'; dosgolem193 Joinsource47d1a115／Leavesourcedf05d0d4',
    'address_space':'linear','address':'IDA9.4 linear103D7..10472;104C4..105C4;103AE..103D6;DGROUP5060;DGROUP5077',
    'consumer':'empty Join530 ->316 inline wait ->540; singleton Leave542 primary actor name ->540; No ->541 read-key ->field',
    'doc':'docs/188-opening-escort-to-castle-spec.md','note':'正常姓名取消、空名冊加入及單人分離的有限D3；不外推滿隊、非空分離、原版存讀檔或完整V3。'}
interface=objects[paths.index(root/'data/interface.json')];selection=interface['recruitment_selection']
assert 'empty_join_text_id' not in selection and 'empty_leave_text_id' not in selection
selection['empty_join_text_id']=selection['empty_view_text_id'];selection['empty_leave_text_id']=id
selection['evidence']['note']+=' 空Join193包530／316及單人Leave193包542已有限D3；非空分離與滿隊仍未驗收。'
definition={'id':id,'value':value,'glyph_codes':codes,'layout':{'kind':'dialogue','columns':20,'lines_per_page':4},'source':{'kind':'legacy_record','file':'D3TXT00.TXT','record':542},'evidence':ev}
texts=objects[paths.index(root/'data/texts.json')];assert not any(t['id']==id for t in texts['definitions']);texts['definitions'].append(definition)
for d in objects:d['schema_version']='0.17.0'
objects[0]['content_version']='0.1.89'
outputs=[]
for p,d,old in zip(paths,objects,raw):
    assert old.count('"schema_version": "0.16.0"')==1
    out=old.replace('"schema_version": "0.16.0"','"schema_version": "0.17.0"')
    if p.name=='manifest.json':out=out.replace('"content_version": "0.1.88"','"content_version": "0.1.89"')
    if p.name=='interface.json':
        anchor='    "empty_view_text_id": "'+selection['empty_view_text_id']+'",';assert out.count(anchor)==1
        out=out.replace(anchor,anchor+'\n    "empty_join_text_id": "'+selection['empty_join_text_id']+'",\n    "empty_leave_text_id": "'+id+'",')
        before=json.loads(old)['recruitment_selection']['evidence']['note']
        anchor='"note": '+json.dumps(before,ensure_ascii=False);assert out.count(anchor)==1
        out=out.replace(anchor,'"note": '+json.dumps(selection['evidence']['note'],ensure_ascii=False))
    if p.name=='texts.json':
        end=out.rindex('\n  ]');fragment=json.dumps(definition,ensure_ascii=False,indent=2).splitlines()
        out=out[:end]+',\n'+'\n'.join('    '+line for line in fragment)+out[end:]
    assert json.loads(out)==d,p
    outputs.append(out)
for p,out in zip(paths,outputs):p.write_text(out);assert p.stat().st_uid==1000
print('九份JSON重建：schema0.17.0／content0.1.89；空Join316與單人Leave542姓名插值')
