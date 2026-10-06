"""原版422相同冷啟動，附加唯讀NPC圖層分支觀測；不改輸入或狀態。"""
from pathlib import Path
s = Path('/work/issue4-field-room-door-probe-r1.py').read_text().rstrip()
marker = "\nns['main']()"
assert s.endswith(marker)
s = s[:-len(marker)].replace('issue4-field-room-door-normal-r1', 'issue4-room-npc-layer-normal-r1')
scope = {'__file__': __file__, '__name__': 'room_npc_layer_builder'}
exec(compile(s, __file__, 'exec'), scope)
ns = scope['ns']
h = ns['GATE_HOOK']
marker = '            if rawFlag&0x4000!=0'
assert h.count(marker) == 1
observer = r'''
            if pilotPackets>=418 && pilotPackets<=422 && (pc==0x11955 || ((pc==0x11e25 || pc==0x11e19 || pc==0x11e33) && m.CPU.R[cpu.AX]&0x2000!=0)) {
                fmt.Printf("DQ3_NPC_LAYER_NATIVE step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x BX=%04x SI=%04x selector=%02x player_x=%d player_y=%d tile_x=%d tile_y=%d\n",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.SI],m.Read8(cpu.Addr(ds,0x2579)),m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),m.Read16(cpu.Addr(ds,0x4f25)),m.Read16(cpu.Addr(ds,0x4f27)))
            }
'''
assert 'm.Write' not in observer and 'Restore' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK'] = h.replace(marker, observer + marker)
ns['__file__'] = __file__
ns['main']()
