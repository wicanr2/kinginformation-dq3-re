"""Docker-only: schema0.31.0/content0.1.103 -> reviewed no-target Talk.

Arguments: writable pack, read-only original assets. READY docs/188; JSON docs/84.
"""
from pathlib import Path
import copy,hashlib,json,os,struct,sys
root,assets=map(Path,sys.argv[1:]);paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==os.getuid()==1000 for p in paths)
raws={p:p.read_text() for p in paths};objects={p:json.loads(s) for p,s in raws.items()}
assert all(o['schema_version']=='0.31.0' for o in objects.values()) and objects[root/'manifest.json']['content_version']=='0.1.103'
exe=(assets/'DQ3.EXE').read_bytes();txt=(assets/'D3TXT00.TXT').read_bytes()
assert len(exe)==115282 and hashlib.sha256(exe).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(txt).hexdigest()=='38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
def code(a,b):return exe[a-0xec90:b-0xec90]
assert code(0x14e7f,0x14e8c).hex()=='bf0401e89e01c6067b060090c3'
record=struct.unpack_from('<H',code(0x14e7f,0x14e82),1)[0]
a,b=struct.unpack_from('<HH',txt,record*2);codes=list(struct.unpack('<'+'H'*((b-a)//2),txt[a:b]));assert codes.pop()==65535
assert codes==[398,546,547,548,401,546,194,494,410,147,56]
evidence=dict(level='D3',source_kind='exe',source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; D3TXT00.TXT SHA256 38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b; dosgolem2f44a68 normal409 source 753da910efcb614851eb707a8e35bd22dfc1e978a9347ba078a067208ff197eb',address_space='linear',address='IDA9.4 14E0E..14E8C;15002..15037;2111B..21148;DGROUP3E6E/file19FAE',consumer='command first entry -> no NPC/counter -> raw01F7=0 -> DI0104/15023 -> retained message -> fresh21133 -> close1F604 -> field1997C',doc='docs/188-opening-escort-to-castle-spec.md',note='confirmed限定正常405..409健康單人城鎮徒步無對象對話、原命令背景、新鍵返回及下一步；raw01F7=1與未取樣分支unknown。字模hold沿用hardware-spec approximation。')
interface=objects[root/'data/interface.json'];assert 'field_talk' not in interface
base=copy.deepcopy(interface['field_status_menu']['reorder']);presentation=base['presentation'];presentation['raw_window']['id']='dq3:raw_window.field_talk_empty';presentation['window']['id']='dq3:window.field_talk_empty'
for o in (presentation,presentation['window'],presentation['shadow'],base['text_flow'],base['wait_indicator']):o['evidence']=copy.deepcopy(evidence)
raw=exe[0x19fae:0x19fae+24];assert raw.hex()=='0b011300ee002c0060009401000000000000000000000c09'
words=struct.unpack('<12H',raw);w=presentation['raw_window'];assert (w['flags'],w['x'],w['y'],w['width'],w['height'])==(words[0]>>8,*words[1:5])
text_id='dq3:text.field.talk.empty'
entry=dict(empty_text_id=text_id,presentation=presentation,text_flow=base['text_flow'],wait_indicator=base['wait_indicator'],scope='healthy_single_member_town_no_target',return_mode='fresh_key_to_field',evidence=evidence);interface['field_talk']=entry
glyphs=json.loads(Path('/repo/docs/data/glyph_unicode_map.json').read_text())
definition=dict(id=text_id,value=''.join(glyphs[str(c)] for c in codes),glyph_codes=codes,layout=dict(kind='dialogue',columns=presentation['window']['columns'],lines_per_page=presentation['window']['lines_per_page']),source=dict(kind='legacy_record',file='D3TXT00.TXT',record=record),evidence=evidence)
definitions=objects[root/'data/texts.json']['definitions'];assert not any(d['id']==text_id for d in definitions);definitions.append(definition)
for path,obj in objects.items():
 s=raws[path]
 if path.name=='interface.json':
  at=s.index('"field_examine": ');s=s[:at]+('"field_talk": '+json.dumps(entry,ensure_ascii=False,indent=2).replace('\n','\n  ')+',\n  ')+s[at:]
 if path.name=='texts.json':
  at=s.index('"definitions": ')+len('"definitions": ');_,length=json.JSONDecoder().raw_decode(s[at:]);end=at+length-1;assert s[end]==']'
  addition='\n'.join('    '+line for line in json.dumps(definition,ensure_ascii=False,indent=2).splitlines());s=s[:end].rstrip()+',\n'+addition+'\n  '+s[end:]
 obj['schema_version']='0.32.0';s=s.replace('"schema_version": "0.31.0"','"schema_version": "0.32.0"')
 if path.name=='manifest.json':obj['content_version']='0.1.104';s=s.replace('"content_version": "0.1.103"','"content_version": "0.1.104"')
 assert json.loads(s)==obj;path.write_text(s)
print('Rebuilt nine JSON: schema0.32.0/content0.1.104; no-target Talk.')
