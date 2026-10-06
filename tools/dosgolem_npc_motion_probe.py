"""相同正常422冷啟動，唯讀記錄NPC mover輸入、分支及返回。"""
from pathlib import Path
s=Path('/work/issue4-field-room-door-probe-r1.py').read_text().rstrip()
marker="\nns['main']()"
assert s.endswith(marker)
s=s[:-len(marker)].replace('issue4-field-room-door-normal-r1','issue4-npc-move-normal-r1')
scope={'__file__':__file__,'__name__':'npc_mover_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
marker='            if rawFlag&0x4000!=0'
assert h.count(marker)==1
observer=r'''
            if pilotPackets>=174 && pilotPackets<=422 && (pc==0x11fea || pc==0x12043 || pc==0x12065 || pc==0x12074 || pc==0x1207f || pc==0x12098 || pc==0x120aa || pc==0x120a6) {
                si:=m.CPU.R[cpu.SI]
                fmt.Printf("DQ3_NPC_MOVE_NATIVE step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x BX=%04x CX=%04x DX=%04x SI=%04x DI=%04x count_b32=%04x selector=%02x player_x=%d player_y=%d tile_x=%d tile_y=%d seed=%04x slot_hex=%02x%02x%02x%02x%02x%02x%02x%02x\n",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.CX],m.CPU.R[cpu.DX],si,m.CPU.R[cpu.DI],m.Read16(cpu.Addr(ds,0x0b32)),m.Read8(cpu.Addr(ds,0x2579)),m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),m.Read16(cpu.Addr(ds,0x4f25)),m.Read16(cpu.Addr(ds,0x4f27)),m.Read16(cpu.Addr(ds,0x0b5a)),m.Read8(cpu.Addr(ds,si)),m.Read8(cpu.Addr(ds,si+1)),m.Read8(cpu.Addr(ds,si+2)),m.Read8(cpu.Addr(ds,si+3)),m.Read8(cpu.Addr(ds,si+4)),m.Read8(cpu.Addr(ds,si+5)),m.Read8(cpu.Addr(ds,si+6)),m.Read8(cpu.Addr(ds,si+7)))
            }
'''
assert 'm.Write' not in observer and 'Restore' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=h.replace(marker,observer+marker)
ns['__file__']=__file__
ns['main']()
