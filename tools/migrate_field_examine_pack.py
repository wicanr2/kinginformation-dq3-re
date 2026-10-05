"""Docker-only: schema0.30.0/content0.1.102 -> reviewed empty-examine entry.

Arguments: writable pack, read-only original assets. READY docs/188; JSON docs/84.
"""
from pathlib import Path
import copy,hashlib,json,os,struct,sys
root,assets=map(Path,sys.argv[1:]);paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==os.getuid()==1000 for p in paths)
raws={p:p.read_text() for p in paths};objects={p:json.loads(s) for p,s in raws.items()}
assert all(o['schema_version']=='0.30.0' for o in objects.values()) and objects[root/'manifest.json']['content_version']=='0.1.102'
exe=(assets/'DQ3.EXE').read_bytes();txt=(assets/'D3TXT00.TXT').read_bytes()
assert len(exe)==115282 and hashlib.sha256(exe).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(txt).hexdigest()=='38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
def code(a,b):return exe[a-0xec90:b-0xec90]
# File bytes retain the original relocation words; IDA loaded bytes differ.
assert code(0x18c93,0x18cb0).hex()=='bf08019a64021b11bf09019a64021b119adb0004118d366e3ee85569c3'
records=[struct.unpack_from('<H',code(a,a+3),1)[0] for a in [0x18c93,0x18c9b]]
assert records==[264,265]
all_codes=[]
for record in records:
 a,b=struct.unpack_from('<HH',txt,record*2);codes=list(struct.unpack('<'+'H'*((b-a)//2),txt[a:b]));assert codes.pop()==65535;all_codes.append(codes)
assert all_codes==[[65531,411,541,542,149,543,551,544,545,123],[508,399,435,436,494,410,147,431,432,56]]
codes=all_codes[0]
evidence=dict(level='D3',source_kind='exe',source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; D3TXT00.TXT SHA256 38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b; dosgolem2f44a68 normal404 source 3af2954a524f521670aad308377efd2aca4c28633a971c111f7af6fc709c44eb',address_space='linear',address='IDA9.4 18966..18986;18C75..18CB0;2111B..21148;15002..15037;DGROUP3E6E/file19FAE',consumer='command sixth entry -> 18966/18986 -> 18C75 -> DI0108 then0109/21414 -> retained message -> fresh21133 -> close1F604 -> field1997C',doc='docs/188-opening-escort-to-castle-spec.md',note='confirmed限定正常395..404健康單人空結果調查、原命令背景、兩record、新按鍵返回與下一步；其他事件分支未驗收。字模hold為沿用hardware-spec approximation。')
interface=objects[root/'data/interface.json'];assert 'field_examine' not in interface
base=copy.deepcopy(interface['field_status_menu']['reorder']);presentation=base['presentation'];presentation['raw_window']['id']='dq3:raw_window.field_examine_empty';presentation['window']['id']='dq3:window.field_examine_empty'
for o in (presentation,presentation['window'],presentation['shadow'],base['text_flow'],base['wait_indicator']):o['evidence']=copy.deepcopy(evidence)
raw=exe[0x19fae:0x19fae+24];assert raw.hex()=='0b011300ee002c0060009401000000000000000000000c09'
words=struct.unpack('<12H',raw);w=presentation['raw_window'];assert (w['flags'],w['x'],w['y'],w['width'],w['height'])==(words[0]>>8,*words[1:5])
ids=['dq3:text.field.examine.intro','dq3:text.field.examine.empty']
entry=dict(intro_text_id=ids[0],result_text_id=ids[1],record_join='new_line',presentation=presentation,text_flow=base['text_flow'],wait_indicator=base['wait_indicator'],actor_variable_code=codes[0],scope='healthy_single_member_no_result',return_mode='fresh_key_to_field',evidence=evidence)
interface['field_examine']=entry
glyphs=json.loads(Path('/repo/docs/data/glyph_unicode_map.json').read_text())
new=[]
for text_id,record,codes in zip(ids,records,all_codes):
 new.append(dict(id=text_id,value=''.join('{actor}' if c==65531 else glyphs[str(c)] for c in codes),glyph_codes=codes,layout=dict(kind='dialogue',columns=presentation['window']['columns'],lines_per_page=presentation['window']['lines_per_page']),source=dict(kind='legacy_record',file='D3TXT00.TXT',record=record),evidence=evidence))
definitions=objects[root/'data/texts.json']['definitions'];assert not any(d['id'] in ids for d in definitions);definitions.extend(new)
for path,obj in objects.items():
 s=raws[path]
 if path.name=='interface.json':
  at=s.index('"field_status_menu": ')
  s=s[:at]+('"field_examine": '+json.dumps(entry,ensure_ascii=False,indent=2).replace('\n','\n  ')+',\n  ')+s[at:]
 if path.name=='texts.json':
  at=s.index('"definitions": ')+len('"definitions": ');_,length=json.JSONDecoder().raw_decode(s[at:]);end=at+length-1;assert s[end]==']'
  addition=',\n'.join('\n'.join('    '+line for line in json.dumps(d,ensure_ascii=False,indent=2).splitlines()) for d in new);s=s[:end].rstrip()+',\n'+addition+'\n  '+s[end:]
 obj['schema_version']='0.31.0';s=s.replace('"schema_version": "0.30.0"','"schema_version": "0.31.0"')
 if path.name=='manifest.json':obj['content_version']='0.1.103';s=s.replace('"content_version": "0.1.102"','"content_version": "0.1.103"')
 assert json.loads(s)==obj;path.write_text(s)
print('Rebuilt nine JSON: schema0.31.0/content0.1.103; empty-examine single-member response.')
