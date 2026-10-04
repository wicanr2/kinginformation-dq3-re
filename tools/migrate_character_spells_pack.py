"""從乾淨0.14.0 pack與原版唯讀資料重建咒文頁；READY入口docs/188。"""
from pathlib import Path
import hashlib,json,struct,sys

root,assets,mapping=map(Path,sys.argv[1:])
paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==1000 for p in paths)
raw=[p.read_text() for p in paths]; objects=[json.loads(s) for s in raw]
assert all(d['schema_version']=='0.14.0' for d in objects)
assert objects[0]['content_version']=='0.1.86'
exe=(assets/'DQ3.EXE').read_bytes(); text=(assets/'D3TXT00.TXT').read_bytes()
assert len(exe)==115282 and hashlib.sha256(exe).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert len(text)==18680
window=exe[0x1a074:0x1a08c]
assert window.hex()=='120313002e002c006000d7010100d801d901000000000000'
ptrs=struct.unpack('<'+'H'*(int.from_bytes(text[:2],'little')//2),text[:int.from_bytes(text[:2],'little')])
assert len(ptrs)-1==759 and ptrs[-1]==len(text)
glyphs=json.loads(mapping.read_text())
# 已有frame契約的可讀字元；glyph_codes仍逐項保留原始record。
glyphs.update({'47':'┌','48':'─','49':'┐','50':'└','51':'┘','54':'│'})
def record(n):
    data=text[ptrs[n]:ptrs[n+1]]
    codes=list(struct.unpack('<'+'H'*(len(data)//2),data))
    assert codes[-1]==65535 and all(0<=code<1476 for code in codes[:-1])
    return codes[:-1]
ev={'level':'D3','source_kind':'exe','source':'DQ3.EXE SHA256 '+hashlib.sha256(exe).hexdigest()+'; D3TXT00.TXT SHA256 '+hashlib.sha256(text).hexdigest()+'; dosgolem209 source4dcb99c8',
    'address_space':'linear','address':'IDA9.4 linear184A1..185EF;1F590..1F604;DGROUP3F34/232D',
    'consumer':'registration ability -> spells -> accept; roster ability -> spells -> restored ability read-key ->540 No ->541 ->field',
    'doc':'docs/188-opening-escort-to-castle-spec.md','note':'正常Class3 level1初裝無異常D3；全catalog與多列static D2；不宣稱全幀V3或原版Save/Load。'}
static={**ev,'level':'D2','consumer':'18573..185C1 sorted union index +79h selects D3TXT00 record121..180; four columns, ceil rows'}
interface=objects[paths.index(root/'data/interface.json')]
assert 'character_spells' not in interface
s={'raw_window':{'id':'character_spells','flags':window[1],'x':int.from_bytes(window[2:4],'little'),'y':int.from_bytes(window[4:6],'little'),'width':int.from_bytes(window[6:8],'little'),'height':int.from_bytes(window[8:10],'little'),'address':'DGROUP:0x3F34;file:0x1A074'},
   'columns':4,'extra_rows':2,'names':{'x':168,'y':62,'step_x':80,'step_y':16},
   'header_text_id':'dq3:text.character_spells.header','row_text_id':'dq3:text.character_spells.row','footer_text_id':'dq3:text.character_spells.footer',
   'catalog':[{'record_raw':n,'text_id':f'dq3:text.character_spells.name.{n}','evidence':static} for n in range(121,181)],'evidence':ev}
interface['character_spells']=s
texts=objects[paths.index(root/'data/texts.json')]
added=[]
for n,id,level in [(471,s['header_text_id'],ev),(472,s['row_text_id'],ev),(473,s['footer_text_id'],ev)]+[(row['record_raw'],row['text_id'],static) for row in s['catalog']]:
    codes=record(n)
    value=''.join(glyphs[str(c)] for c in codes)
    added.append({'id':id,'value':value,'glyph_codes':codes,'layout':{'kind':'menu_record','columns':22,'lines_per_page':1},'source':{'kind':'legacy_record','file':'D3TXT00.TXT','record':n},'evidence':level})
assert record(161)==[376,359,347]
assert all(len(record(n))<=5 for n in range(121,181))
assert not {t['id'] for t in texts['definitions']} & {t['id'] for t in added}
texts['definitions'].extend(added)
for d in objects:d['schema_version']='0.15.0'
objects[0]['content_version']='0.1.87'
outputs=[]
for p,d,old in zip(paths,objects,raw):
    assert old.count('"schema_version": "0.14.0"')==1
    out=old.replace('"schema_version": "0.14.0"','"schema_version": "0.15.0"')
    if p.name=='manifest.json':out=out.replace('"content_version": "0.1.86"','"content_version": "0.1.87"')
    if p.name=='interface.json':
        end=out.rindex('\n}')
        fragment=json.dumps(s,ensure_ascii=False,indent=2).splitlines()
        out=out[:end]+',\n  "character_spells": '+fragment[0]+'\n'+'\n'.join('  '+line for line in fragment[1:])+out[end:]
    if p.name=='texts.json':
        end=out.rindex('\n  ]')
        fragments=[json.dumps(t,ensure_ascii=False,indent=2).splitlines() for t in added]
        out=out[:end]+',\n'+',\n'.join('\n'.join('    '+line for line in f) for f in fragments)+out[end:]
    assert json.loads(out)==d,p
    outputs.append(out)
for p,out in zip(paths,outputs):p.write_text(out);assert p.stat().st_uid==1000
print('九份JSON重建：schema0.15.0 / content0.1.87；完整60咒文及3框record')
