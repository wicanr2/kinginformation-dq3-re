"""乾淨schema0.15.0 pack重建空名冊View契約；READY與用法見docs/188。"""
from pathlib import Path
import hashlib,json,struct,sys

root,assets,mapping=map(Path,sys.argv[1:])
paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==1000 for p in paths)
raw=[p.read_text() for p in paths]; objects=[json.loads(s) for s in raw]
assert all(d['schema_version']=='0.15.0' for d in objects) and objects[0]['content_version']=='0.1.87'
exe=(assets/'DQ3.EXE').read_bytes(); text=(assets/'D3TXT00.TXT').read_bytes()
assert len(exe)==115282 and hashlib.sha256(exe).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(text).hexdigest()=='38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
assert exe[0x199f:0x19a1].hex()=='eb65' and exe[0x1a06:0x1a09].hex()=='bf3c01'
end=int.from_bytes(text[:2],'little'); ptrs=struct.unpack('<'+'H'*(end//2),text[:end])
assert len(ptrs)-1==759 and ptrs[-1]==len(text)
record=text[ptrs[316]:ptrs[317]]
assert record.hex()=='3b003c00f8015e02c5011502da00bb0095005f0214026002fcff3b003c00c001120261026202630259025a029500c200feff0c000c005f0214024702cd013700feff0c000c003d02c5018e010002cd013900ffff'
codes=list(struct.unpack('<'+'H'*(len(record)//2),record))[:-1]
glyphs=json.loads(mapping.read_text())
value=''.join(glyphs[str(c)] if c<1476 else '\n' for c in codes)
assert set(c for c in codes if c>=1476)=={65532,65534}
id='dq3:text.recruitment_selection.empty_view'
ev={'level':'D3','source_kind':'exe','source':'DQ3.EXE SHA256 '+hashlib.sha256(exe).hexdigest()+'; D3TXT00.TXT SHA256 '+hashlib.sha256(text).hexdigest()+'; dosgolem195 sourceff7a8abd',
    'address_space':'linear','address':'IDA9.4 linear10624..1069E;10974..109A9;103A8..103D6;DGROUP5060',
    'consumer':'empty roster View -> record316 inline wait -> caller540 No ->541 read-key ->field',
    'doc':'docs/188-opening-escort-to-castle-spec.md','note':'正常姓名取消、名冊空的有限D3；不外推其他空清單、滿隊、原版存讀檔或完整V3。'}
interface=objects[paths.index(root/'data/interface.json')]
selection=interface['recruitment_selection'];assert 'empty_view_text_id' not in selection
selection['empty_view_text_id']=id
definition={'id':id,'value':value,'glyph_codes':codes,'layout':{'kind':'dialogue','columns':20,'lines_per_page':4},'source':{'kind':'legacy_record','file':'D3TXT00.TXT','record':316},'evidence':ev}
texts=objects[paths.index(root/'data/texts.json')];assert not any(t['id']==id for t in texts['definitions'])
texts['definitions'].append(definition)
for d in objects:d['schema_version']='0.16.0'
objects[0]['content_version']='0.1.88'
outputs=[]
for p,d,old in zip(paths,objects,raw):
    assert old.count('"schema_version": "0.15.0"')==1
    out=old.replace('"schema_version": "0.15.0"','"schema_version": "0.16.0"')
    if p.name=='manifest.json':out=out.replace('"content_version": "0.1.87"','"content_version": "0.1.88"')
    if p.name=='interface.json':
        anchor='    "continue_text_id": "'+selection['continue_text_id']+'",'
        assert out.count(anchor)==1
        out=out.replace(anchor,anchor+'\n    "empty_view_text_id": "'+id+'",')
    if p.name=='texts.json':
        end=out.rindex('\n  ]');fragment=json.dumps(definition,ensure_ascii=False,indent=2).splitlines()
        out=out[:end]+',\n'+'\n'.join('    '+line for line in fragment)+out[end:]
    assert json.loads(out)==d,p
    outputs.append(out)
for p,out in zip(paths,outputs):p.write_text(out);assert p.stat().st_uid==1000
print('九份JSON重建：schema0.16.0／content0.1.88；空名冊316完整glyph及內文等待')
