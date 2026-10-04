"""重建 docs/188 有限 READY 的 F5/F6 pack；於 Docker 執行，原始輸入唯讀。"""
from pathlib import Path
import hashlib,json,struct,sys

root,assets,mapping=map(Path,sys.argv[1:])
paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==1000 for p in paths)
raw=[p.read_text() for p in paths];objects=[json.loads(s) for s in raw]
assert all(d['schema_version']=='0.17.0' for d in objects) and objects[0]['content_version']=='0.1.89'
exe=(assets/'DQ3.EXE').read_bytes();text=(assets/'D3TXT00.TXT').read_bytes()
assert len(exe)==115282 and hashlib.sha256(exe).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(text).hexdigest()=='38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
assert exe[0x27af:0x27b2].hex()=='bd1200' # IDA linear1143F
end=int.from_bytes(text[:2],'little');ptrs=struct.unpack('<'+'H'*(end//2),text[:end]);assert len(ptrs)-1==759
glyphs=json.loads(mapping.read_text())
ev={'level':'D3','source_kind':'exe','source':'DQ3.EXE SHA256 '+hashlib.sha256(exe).hexdigest()+'; D3TXT00.TXT SHA256 '+hashlib.sha256(text).hexdigest()+'; dosgolem source3350c30f/08a3fead/0a88466e/30a4f547',
 'address_space':'linear','address':'IDA9.4 linear113A9..11890;1D8E9..1D94C;DGROUP3FBA; file1A0FA',
 'consumer':'normal F5 ->251/253 ->Yes slots250 or No ->252 read-key; slot cancel ->252; native save2172 ->left ->F6 restore',
 'doc':'docs/188-opening-escort-to-castle-spec.md','note':'有限READY；正常單人第1槽、No、YesEsc與選槽Esc；JSON存檔與DOS格式不同。既有十槽初始內容差異保留；不宣稱全campaign/V3。'}
names={'prompt':250,'experience':251,'farewell':252,'question':253,'header':465,'row':466,'footer':467,'empty':468}
texts=objects[paths.index(root/'data/texts.json')];interface=objects[paths.index(root/'data/interface.json')]
assert 'field_save_load' not in interface
ids={role:'dq3:text.field_save_load.'+role for role in names}
new=[]
for role,rec in names.items():
 data=text[ptrs[rec]:ptrs[rec+1]];codes=list(struct.unpack('<'+'H'*(len(data)//2),data));assert codes[-1]==65535;codes=codes[:-1]
 value=''.join(glyphs[str(c)] if c<1476 else '\n' if c==65534 else '{名字}' if c in [65531,65525] else '{數值}' if c==65530 else '{控制}' for c in codes)
 layout={'kind':'dialogue','columns':20,'lines_per_page':4} if rec<465 else {'kind':'menu_record','columns':21 if rec<468 else len(codes),'lines_per_page':2 if role=='header' else 1}
 d={'id':ids[role],'value':value,'glyph_codes':codes,'layout':layout,'source':{'kind':'legacy_record','file':'D3TXT00.TXT','record':rec},'evidence':ev}
 assert not any(t['id']==d['id'] for t in texts['definitions']);texts['definitions'].append(d);new.append(d)
w=struct.unpack_from('<16H',exe,0x1a0fa)
assert w==(786,19,30,42,208,465,10,466,467,0,10,3,21,62,0,787)
def anchor(x):return {'x':x,'y':w[13],'step_x':16,'step_y':16}
load_clocks=[{'layer':0,'clock':0,'evidence':dict(ev,address='IDA9.4 linear1160B..1162B;DGROUP251D/526C/4F31',consumer='upper-world native F6 ->clock0 ->palette bank1',note='正常200包2172bytes讀回後色盤bank1，D3有限上層。')},
 {'layer':1,'clock':120,'evidence':dict(ev,level='D2',address='IDA9.4 linear1160B..1161D;DGROUP251D/526C/4F31',consumer='lower-world static F6 ->clock120 ->night',note='原始writer及world層號契約；缺下層正常F6來源，不能宣稱正常parity。')}]
assert exe[0x2995:0x299b].hex()=='c7061d250000' # IDA linear11625 clock0
s={'load_clock_rules':load_clocks,'shadow':dict(interface['recruitment_entry']['shadow'],evidence=ev),'id':'dq3:field_save_load','slot_count':w[10],'extra_rows':w[11],'primary_class_raw':0,'experience_max_level':44,
 'experience_text_id':ids['experience'],'question_text_id':ids['question'],'prompt_text_id':ids['prompt'],'farewell_text_id':ids['farewell'],
 'header_text_id':ids['header'],'row_text_id':ids['row'],'footer_text_id':ids['footer'],'empty_text_id':ids['empty'],
 'name_control_codes':[65531,65525],'experience_control_code':65530,
 'raw_window':{'id':'field_save_load','flags':w[0]>>8,'x':w[1],'y':w[2],'width':w[3],'height':w[4],'address':'linear:0x28d8a'},
 'number':{'x':(w[12]-4)*8,'y':w[13],'digits':5},'name':anchor((w[12]+8)*8),'name_capacity':4,
 'level':{'x':(w[12]+16)*8,'y':w[13],'digits':5},'gender':anchor((w[12]+8+24)*8),
 'cursor':dict(anchor(w[12]*8),step_x=0),'hit_rect':{'x':w[12]*8,'y':w[13],'width':(w[1]+w[3])*8-w[12]*8-16,'height':16},
 'gender_text_ids':interface['recruitment_selection']['gender_text_ids'],
 'sound':{'cue_raw':18,'wait_for_completion':True,'evidence':dict(ev,address='IDA9.4 linear1143A..1147E;20770..207CC; DGROUP253E',consumer='cue18 VOC -> write PLAYER200/DRAGON2172 ->208E2 completion wait',note='VOC已知sample duration契約；ceil(source_duration_nanos*60/1e9)，hardware-spec approximation；不深挖DAC/PIT/DMA。')},'evidence':ev}
interface['field_save_load']=s
for d in objects:d['schema_version']='0.18.0'
objects[0]['content_version']='0.1.90'
outputs=[]
for p,d,old in zip(paths,objects,raw):
 out=old.replace('"schema_version": "0.17.0"','"schema_version": "0.18.0"')
 if p.name=='manifest.json':out=out.replace('"content_version": "0.1.89"','"content_version": "0.1.90"')
 if p.name=='interface.json':
  end=out.rindex('\n}');fragment=json.dumps(s,ensure_ascii=False,indent=2).splitlines();out=out[:end]+',\n  "field_save_load": '+fragment[0]+'\n'+'\n'.join('  '+l for l in fragment[1:])+out[end:]
 if p.name=='texts.json':
  end=out.rindex('\n  ]');out=out[:end]+',\n'+',\n'.join('\n'.join('    '+l for l in json.dumps(t,ensure_ascii=False,indent=2).splitlines()) for t in new)+out[end:]
 assert json.loads(out)==d,p;outputs.append(out)
for p,out in zip(paths,outputs):p.write_text(out);assert p.stat().st_uid==1000
print('九份JSON重建：schema0.18.0／content0.1.90；原版F5/F6有限契約')
