"""Docker-only: rebuild the reviewed item UI/drop contract from clean 0.20.0.

Arguments: writable pack copy, read-only original assets, glyph Unicode map.
Contract and evidence: docs/84 and docs/188. No original asset is published.
"""
from pathlib import Path
import hashlib,json,os,struct,sys
root,assets,mapping=map(Path,sys.argv[1:])
paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==os.getuid()==1000 for p in paths)
old=[p.read_text() for p in paths];objects=[json.loads(s) for s in old]
assert all(d['schema_version']=='0.20.0' for d in objects) and objects[0]['content_version']=='0.1.92'
exe=(assets/'DQ3.EXE').read_bytes();text=(assets/'D3TXT00.TXT').read_bytes();items=(assets/'ITEM.DAT').read_bytes()
for raw,size,digest in [(exe,115282,'5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'),(text,None,'38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'),(items,896,'7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d')]:
 assert (size is None or len(raw)==size) and hashlib.sha256(raw).hexdigest()==digest
for linear,raw in [(0x13801,'050200'),(0x1380d,'051000'),(0x1381c,'050200b104d3e0'),(0x138a7,'8306160704'),(0x138ac,'8306180710'),(0x138cf,'bb0a00'),(0x13ad5,'a900e0'),(0x13aeb,'a802')]:
 assert exe[linear-0xec90:linear-0xec90+len(bytes.fromhex(raw))]==bytes.fromhex(raw),hex(linear)
characters=objects[paths.index(root/'data/characters.json')];interface=objects[paths.index(root/'data/interface.json')];texts=objects[paths.index(root/'data/texts.json')]
characters['item_storage']['drop_blocked_mask']=int.from_bytes(exe[0x13ad6-0xec90:0x13ad8-0xec90],'little')
for code,m in enumerate(characters['item_storage']['items']):m['drop_forbidden']=bool(items[code*7+5]&exe[0x13aec-0xec90])
ev={'level':'D3','source_kind':'exe','source':'DQ3.EXE + D3TXT00.TXT + dosgolem2f44a68正常218/230包來源28995c8d/aad971bb','address_space':'dgroup','address':'0x3fd8/0x4050','consumer':'IDAlinear1372F→137F9/13887→1F779/1F908；13919 physical item reader','doc':'docs/188-opening-escort-to-castle-spec.md','note':'完整EXE sha2565178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c；正常來源與glyph/frame試作限定confirmed。'}
def rawwindow(address,name):
 w=struct.unpack_from('<14H',exe,0x16140+address)
 return w,{'id':name,'flags':w[0]>>8,'x':w[1],'y':w[2],'width':w[3],'height':w[4],'address':hex(address)}
w,window=rawwindow(0x3fd8,'dq3:raw_window.field_items');a,action=rawwindow(0x4050,'dq3:raw_window.field_item_actions')
assert w[:9]==(0x313,43,46,16,160,418,3,419,420) and a[:6]==(0x314,11,116,12,80,421)
glyphs=json.loads(mapping.read_text());end=int.from_bytes(text[:2],'little');ptrs=struct.unpack('<'+'H'*(end//2),text[:end])
ids={}
for role,record,cols,rows in [('header',w[5],8,1),('row',w[7],8,1),('footer',w[8],8,1),('actions',a[5],6,5),('give_prompt',308,32,6)]:
 codes=list(struct.unpack('<'+'H'*((ptrs[record+1]-ptrs[record])//2),text[ptrs[record]:ptrs[record+1]]));assert codes[-1]==65535;codes=codes[:-1]
 id='dq3:text.field.items.'+role;ids[role]=id
 value=''.join('\n' if c==65534 else glyphs[str(c)] for c in codes)
 if role=='give_prompt':cols,rows=interface['dialogue']['columns'],interface['dialogue']['lines_per_page']
 definition={'id':id,'value':value,'glyph_codes':codes,'layout':{'kind':'menu_record' if role!='give_prompt' else 'dialogue','columns':cols,'lines_per_page':rows},'source':{'kind':'legacy_record','file':'D3TXT00.TXT','record':record},'evidence':dict(ev,source_kind='data_file',address_space='record',address=str(record))}
 assert not any(t['id']==id for t in texts['definitions']);texts['definitions'].append(definition)
interface['field_items']={'id':'dq3:window.field_items','raw_window':window,'action_window':action,'text_ids':ids,'frame_rows':2,'row_step':16,'name':{'x':(w[1]+4)*8,'y':w[2]+16},'cursor':{'x':w[12]*8,'y':w[13]},'worn':{'x':w[1]*8,'y':w[2]+16},'action_cursor':{'x':a[12]*8,'y':a[13]},'cursor_glyph':11,'worn_glyph':int.from_bytes(exe[0x138d0-0xec90:0x138d2-0xec90],'little'),'evidence':ev}
interface['field_items']['marker_mask']=int.from_bytes(exe[0x138c5-0xec90:0x138c7-0xec90],'little')|int.from_bytes(exe[0x138ca-0xec90:0x138cc-0xec90],'little')
outputs=[]
for path,obj,raw in zip(paths,objects,old):
 obj['schema_version']='0.21.0';out=raw.replace('"schema_version": "0.20.0"','"schema_version": "0.21.0"')
 if path.name=='manifest.json':obj['content_version']='0.1.93';out=out.replace('"content_version": "0.1.92"','"content_version": "0.1.93"')
 if path.name=='characters.json':out=json.dumps(obj,ensure_ascii=False,indent=2)+'\n'
 if path.name=='interface.json':
  end=out.rindex('\n}');fragment=json.dumps(interface['field_items'],ensure_ascii=False,indent=2).splitlines()
  out=out[:end]+',\n  "field_items": '+fragment[0]+'\n'+'\n'.join('  '+line for line in fragment[1:])+out[end:]
 if path.name=='texts.json':
  new=texts['definitions'][len(json.loads(raw)['definitions']):]
  end=out.rindex('\n  ]');out=out[:end]+',\n'+',\n'.join('\n'.join('    '+line for line in json.dumps(d,ensure_ascii=False,indent=2).splitlines()) for d in new)+out[end:]
 assert json.loads(out)==obj;outputs.append(out)
for path,out in zip(paths,outputs):path.write_text(out)
print('Rebuilt nine JSON files: schema0.21.0/content0.1.93; item UI and original drop gates.')
