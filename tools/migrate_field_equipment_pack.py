"""Docker-only: reviewed single-owner equipment, READY docs/188; JSON docs/84.

Arguments: writable pack, read-only original assets. Requires schema0.29.0.
"""
from pathlib import Path
import copy, hashlib, json, os, struct, sys

root, assets = map(Path, sys.argv[1:])
paths = [root/'manifest.json', *sorted((root/'data').glob('*.json'))]
assert len(paths) == 9 and all(p.is_file() and p.stat().st_uid == os.getuid() == 1000 for p in paths)
raws = {p: p.read_text() for p in paths}
objects = {p: json.loads(s) for p, s in raws.items()}
assert all(o['schema_version'] == '0.29.0' for o in objects.values())
assert objects[root/'manifest.json']['content_version'] == '0.1.101'
exe, txt, item = [(assets/n).read_bytes() for n in ['DQ3.EXE','D3TXT00.TXT','ITEM.DAT']]
assert len(exe) == 115282 and hashlib.sha256(exe).hexdigest() == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
assert hashlib.sha256(txt).hexdigest() == '38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b'
assert len(item) == 896 and hashlib.sha256(item).hexdigest() == '7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d'
def words(offset, count): return list(struct.unpack_from('<'+'H'*count, exe, offset))
lw, pw = words(0x1a21c, 15), words(0x1a23a, 12)
assert lw == [793,43,30,20,160,425,3,430,431,0,0,3,45,46,33011]
assert pw == [282,43,110,20,64,413,0,0,0,0,0,0]
evidence = dict(level='D3', source_kind='exe', source='DQ3.EXE SHA256 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; ITEM.DAT SHA256 7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d; dosgolem2f44a68 normal339 source 721d93f6de2123f03ad82ce991185f7329fa918898c1b1cfbf25d2870f3cf414', address_space='linear', address='IDA9.4 17E12..18301; DGROUP40DC/file1A21C;DGROUP40FA/file1A23A', consumer='single actor -> four category lists -> physical word writer -> cached stats -> field1997C', doc='docs/188-opening-escort-to-castle-spec.md', note='confirmed限定健康單人、無詛咒且候選符合資格；其他分支未驗收。')
def window(w, name, address):
    return dict(id='dq3:raw_window.field_equipment.'+name, flags=w[0]>>8, x=w[1], y=w[2], width=w[3], height=w[4], address=address)
records = [425,426,427,428,430,431,271,413]
ids = {r: 'dq3:text.field.equipment.'+str(r) for r in records}
eligibility = []
for i in range(len(item)//7):
    meta = struct.unpack_from('<H',item,i*7+4)[0]
    eligibility.append(dict(classes=[c for c in range(8) if item[i*7+6] & (1<<c)], required_gender=1 if meta&0x100 else -1))
entry = dict(scope='healthy_single_owner_eligible_uncursed', parts=[dict(part=p,header_text_id=ids[425+i]) for i,p in enumerate([0,1,3,2])], raw_window=window(lw,'list','DGROUP40DC/file1A21C'), preview_window=window(pw,'preview','DGROUP40FA/file1A23A'), row_text_id=ids[430],footer_text_id=ids[431],none_text_id=ids[271],preview_text_id=ids[413],frame_rows=3,row_step=16,name=dict(x=(lw[1]+4)*8,y=lw[13]),cursor=dict(x=lw[12]*8,y=lw[13]),worn=dict(x=lw[1]*8,y=lw[13]),cursor_glyph=11,worn_glyph=10,attack=dict(x=(lw[1]+8)*8,y=16,digits=5),defense=dict(x=(lw[1]+8)*8,y=32,digits=5),eligibility=eligibility,evidence=evidence)
interface=objects[root/'data/interface.json'];assert 'field_equipment' not in interface;interface['field_equipment']=entry
glyphs=json.loads(Path('/repo/docs/data/glyph_unicode_map.json').read_text())
new=[]
definitions=objects[root/'data/texts.json']['definitions']
for record in records:
    a,b=struct.unpack_from('<HH',txt,record*2)
    codes=list(struct.unpack('<'+'H'*((b-a)//2),txt[a:b]));assert codes.pop()==65535
    assert all(c<1476 or c==65534 for c in codes)
    # Keep record controls out of the visible value. Glyph stream is authoritative.
    value=''.join('\n' if c==65534 else glyphs[str(c)] for c in codes)
    obj=dict(id=ids[record],value=value,glyph_codes=codes,layout=dict(kind='dialogue',columns=10,lines_per_page=4 if record==413 else 1),source=dict(kind='legacy_record',file='D3TXT00.TXT',record=record),evidence=copy.deepcopy(evidence))
    assert not any(d['id']==obj['id'] for d in definitions)
    definitions.append(obj);new.append(obj)
for path,obj in objects.items():
    s=raws[path]
    if path.name=='interface.json':
        at=s.index('"field_spell_entry": ')
        s=s[:at]+('"field_equipment": '+json.dumps(entry,ensure_ascii=False,indent=2).replace('\n','\n  ')+',\n  ')+s[at:]
    if path.name=='texts.json':
        at=s.index('"definitions": ')+len('"definitions": ');_,length=json.JSONDecoder().raw_decode(s[at:]);end=at+length-1
        addition=',\n'.join('\n'.join('    '+line for line in json.dumps(o,ensure_ascii=False,indent=2).splitlines()) for o in new)
        s=s[:end].rstrip()+',\n'+addition+'\n  '+s[end:]
    obj['schema_version']='0.30.0';s=s.replace('"schema_version": "0.29.0"','"schema_version": "0.30.0"')
    if path.name=='manifest.json':
        obj['content_version']='0.1.102';s=s.replace('"content_version": "0.1.101"','"content_version": "0.1.102"')
    assert json.loads(s)==obj;path.write_text(s)
print('Rebuilt nine JSON: schema0.30.0/content0.1.102; physical equipment lists.')
