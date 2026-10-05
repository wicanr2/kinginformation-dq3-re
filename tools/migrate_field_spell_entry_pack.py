"""Docker-only: schema0.28.0/content0.1.100 -> reviewed empty-spell entry.

Arguments: writable pack, read-only original assets. READY docs/188; JSON docs/84.
"""
from pathlib import Path
import copy,hashlib,json,os,struct,sys
root,assets=map(Path,sys.argv[1:]);paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==os.getuid()==1000 for p in paths)
raws={p:p.read_text() for p in paths};objects={p:json.loads(s) for p,s in raws.items()}
assert all(o['schema_version']=='0.28.0' for o in objects.values()) and objects[root/'manifest.json']['content_version']=='0.1.100'
exe=(assets/'DQ3.EXE').read_bytes();txt=(assets/'D3TXT00.TXT').read_bytes()
assert len(exe)==115282 and hashlib.sha256(exe).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(txt).hexdigest()=='38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
def code(a,b):return exe[a-0xec90:b-0xec90]
assert code(0x18869,0x18870).hex()=='c70622070100c3'
assert code(0x1c9e7,0x1c9ee).hex()=='bf0601e83686c3'
assert code(0x1cb39,0x1cb3c).hex()=='b001c3'
record=struct.unpack_from('<H',code(0x1c9e7,0x1c9ea),1)[0]
a,b=struct.unpack_from('<HH',txt,record*2);codes=list(struct.unpack('<'+'H'*((b-a)//2),txt[a:b]));assert codes.pop()==65535 and codes==[65531,228,415,210,412,429,430,56]
evidence=dict(level='D3',source_kind='exe',source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; D3TXT00.TXT SHA256 38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b; dosgolem2f44a68 normal304 source c096702865ce694706ba2aac2f9333408b895bf45d856bae4a14334d1275cfcf',address_space='linear',address='IDA9.4 1C9C1..1C9EE;1885F..18870;1C9EE..1CB3C;15002..15037;DGROUP3E6E/file19FAE',consumer='command fourth entry -> automatic single actor -> zero learned counts -> DI0106/15023 -> retained message -> fresh21133 -> field1997C',doc='docs/188-opening-escort-to-castle-spec.md',note='confirmed限定正常299..304健康單人未學咒文入口、原命令背景、新按鍵返回與下一步；其他施法分支未驗收。字模hold為沿用hardware-spec approximation。')
interface=objects[root/'data/interface.json'];assert 'field_spell_entry' not in interface
base=copy.deepcopy(interface['field_status_menu']['reorder']);presentation=base['presentation'];presentation['raw_window']['id']='dq3:raw_window.field_spell_empty';presentation['window']['id']='dq3:window.field_spell_empty'
for o in (presentation,presentation['window'],presentation['shadow'],base['text_flow'],base['wait_indicator']):o['evidence']=copy.deepcopy(evidence)
raw=exe[0x19fae:0x19fae+24];assert raw.hex()=='0b011300ee002c0060009401000000000000000000000c09'
words=struct.unpack('<12H',raw);w=presentation['raw_window'];assert (w['flags'],w['x'],w['y'],w['width'],w['height'])==(words[0]>>8,*words[1:5])
text_id='dq3:text.field.spell.empty'
entry=dict(empty_text_id=text_id,presentation=presentation,text_flow=base['text_flow'],wait_indicator=base['wait_indicator'],actor_variable_code=codes[0],scope='single_member_no_spells',return_mode='fresh_key_to_field',evidence=evidence)
interface['field_spell_entry']=entry
glyphs=json.loads(Path('/repo/docs/data/glyph_unicode_map.json').read_text())
new=dict(id=text_id,value='{actor}'+''.join(glyphs[str(c)] for c in codes[1:]),glyph_codes=codes,layout=dict(kind='dialogue',columns=presentation['window']['columns'],lines_per_page=presentation['window']['lines_per_page']),source=dict(kind='legacy_record',file='D3TXT00.TXT',record=record),evidence=evidence)
definitions=objects[root/'data/texts.json']['definitions'];assert not any(d['id']==text_id for d in definitions);definitions.append(new)
for path,obj in objects.items():
 s=raws[path]
 if path.name=='interface.json':
  at=s.index('"field_status_menu": ')
  s=s[:at]+('"field_spell_entry": '+json.dumps(entry,ensure_ascii=False,indent=2).replace('\n','\n  ')+',\n  ')+s[at:]
 if path.name=='texts.json':
  at=s.index('"definitions": ')+len('"definitions": ');_,length=json.JSONDecoder().raw_decode(s[at:]);end=at+length-1;assert s[end]==']'
  addition='\n'.join('    '+line for line in json.dumps(new,ensure_ascii=False,indent=2).splitlines());s=s[:end].rstrip()+',\n'+addition+'\n  '+s[end:]
 obj['schema_version']='0.29.0';s=s.replace('"schema_version": "0.28.0"','"schema_version": "0.29.0"')
 if path.name=='manifest.json':obj['content_version']='0.1.101';s=s.replace('"content_version": "0.1.100"','"content_version": "0.1.101"')
 assert json.loads(s)==obj;path.write_text(s)
print('Rebuilt nine JSON: schema0.29.0/content0.1.101; empty-spell single-member response.')
