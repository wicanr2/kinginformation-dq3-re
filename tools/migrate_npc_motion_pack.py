"""Docker-only: schema0.32.0/content0.1.105 -> 0.33.0/0.1.106.

Arguments: writable pack directory, read-only original assets. READY docs/188.
The finite movement parameters are extracted from the pinned original bytes.
"""
from pathlib import Path
import hashlib,json,os,struct,sys
root,assets=map(Path,sys.argv[1:])
paths=[root/'manifest.json',*sorted((root/'data').glob('*.json'))]
assert len(paths)==9 and all(p.is_file() and p.stat().st_uid==os.getuid()==1000 for p in paths)
raws={p:p.read_text() for p in paths};objects={p:json.loads(s) for p,s in raws.items()}
assert all(o['schema_version']=='0.32.0' for o in objects.values())
manifest,characters=root/'manifest.json',root/'data/characters.json'
assert objects[manifest]['content_version']=='0.1.105'
exe=(assets/'DQ3.EXE').read_bytes()
assert len(exe)==115282 and hashlib.sha256(exe).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
def u16(pc,operand):return struct.unpack_from('<H',exe,pc-0xec90+operand)[0]
def u8(pc,operand):return exe[pc-0xec90+operand]
steps=[dict(x=x,y=y) for x,y in struct.iter_unpack('<hh',exe[0x16140+0xb35:0x16140+0xb45])]
evidence=dict(level='D3',source_kind='exe',source='DQ3.EXE sha256:5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c; dosgolem2f44a68 readonly458 b2fdf7818b4d8ca07d97d5796db6ac241490a6899f8352ab2396cabb3b96c9c6',address_space='linear',address='IDA9.4 linear0x11F4E..0x12024;0x12025..0x121EE; DGROUP0x0B35..0x0B44 base linear0x24DD0; file=linear-0xEC90',consumer='automatic NPC visible-cell scan, random qualification, turn and step transaction',doc='docs/188-opening-escort-to-castle-spec.md#npc-motion-ready',note='1672 native entry/return cases, 19 outdoor turns, 110 step rolls and 5 coordinate writes; full global RNG and animation phase parity unknown')
layer_evidence=dict(evidence,level='D3',source=evidence['source']+'; controlled four-layer component fe39794adbe6c4c3a64fcaf320e7bec75984ff913afdc1906f8fc423d66358b1',address='0x131E6..0x13282;0x12079..0x12098;0x11FB0..0x11FEA',consumer='signed packed cell word compared with half byte-count; quotient AL delta by typed layer',note='byte count<=255; occupied layer0/1 -> decrement; signed layer2/3 -> increment. Four native component cases seed1e2c, explicit layer injection and CPU reentry; normal upper-layer player parity unknown')
motion=dict(viewport_anchor=dict(x=u8(0x11f61,2),y=u16(0x11f51,1)),viewport_columns=u16(0x11f7e,1),viewport_rows=u16(0x11f74,1),evaluation=dict(bound=u16(0x1203d,1),accepted=u8(0x12043,2)),direction_bound=u16(0x1205f,1),direction_mask=u8(0x12068,1),turn=dict(bound=u16(0x1206e,1),accepted=u8(0x12074,2)),step=dict(bound=u16(0x120aa,1),accepted=u8(0x120b0,2)),move_mask=u8(0x1204e,3),frozen_mask=u8(0x12048,3),quotient_mask=255,turn_delta_by_layer=[-1,-1,1,1],directions=steps,minimum_axis_distance=u8(0x120e7,2),blocked_attribute_mask=255,evidence=evidence,turn_layer_evidence=layer_evidence)
assert motion['viewport_anchor']==dict(x=9,y=7) and motion['viewport_columns']==20 and motion['viewport_rows']==15
assert u8(0x12117,2)==motion['minimum_axis_distance'] and exe[0x1218d-0xec90:0x1218d-0xec90+2]==bytes.fromhex('3c00')
assert 'npc_motion' not in objects[characters]
objects[characters]['npc_motion']=motion
objects[characters]['schema_version']='0.33.0'
for p,s in raws.items():
    updated=s.replace('"schema_version": "0.32.0"','"schema_version": "0.33.0"')
    if p==manifest:updated=updated.replace('"content_version": "0.1.105"','"content_version": "0.1.106"')
    if p==characters:
        entry='\n'.join(json.dumps({'npc_motion':motion},ensure_ascii=False,indent=2).splitlines()[1:-1])
        updated=updated.rstrip()[:-1].rstrip()+',\n'+entry+'\n}\n'
        assert json.loads(updated)==objects[p]
    p.write_text(updated)
    assert p.stat().st_uid==os.getuid()
print('Rebuilt npc_motion from original bytes; schema0.33.0/content0.1.106; storage1/save2 unchanged.')
