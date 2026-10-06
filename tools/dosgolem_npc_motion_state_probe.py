"""相同固定458正常輸入，補足落步條件與完整NPC表的唯讀初始狀態。"""
from pathlib import Path
s=Path('/work/issue4-npc-move-continue-probe-r1.py').read_text().rstrip()
marker="\nns['main']()"
assert s.endswith(marker)
s=s[:-len(marker)].replace('issue4-npc-move-continue-normal-r1','issue4-npc-move-state-normal-r1')
scope={'__file__':__file__,'__name__':'npc_mover_state_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
marker='            if rawFlag&0x4000!=0'
assert h.count(marker)==1
observer=r'''
            if pc==0x131ec || pc==0x13282 || pc==0x13285 || pc==0x1329c {
                fmt.Printf("DQ3_NPC_TABLE_NATIVE step=%d packet=%d ida_linear=%05x DS=%04x CX=%04x BP=%04x raw_b32=%04x\n",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.CX],m.CPU.R[cpu.BP],m.Read16(cpu.Addr(ds,0x0b32)))
            }
            if pilotPackets>=422 && pilotPackets<=458 && (pc==0x12025 || pc==0x120b0 || pc==0x12161 || pc==0x1217e || pc==0x1218d || pc==0x120a6) {
                slot:=m.CPU.R[cpu.SI]
                if pc==0x12025 {slot=0x0b66+uint16((m.CPU.R[cpu.BX]>>8)&0x1f)*8}
                count:=m.Read16(cpu.Addr(ds,0x0b32))
                if count>32 {panic("NPC source table exceeds 5-bit slot scope")}
                all:=make([]byte,int(count)*8)
                for i:=range all {all[i]=m.Read8(cpu.Addr(ds,0x0b66+uint16(i)))}
                sx:=int(m.Read8(cpu.Addr(ds,slot)));sy:=int(m.Read8(cpu.Addr(ds,slot+1)));dir:=uint16(m.Read8(cpu.Addr(ds,slot+3))&3)
                tx:=sx+int(int16(m.Read16(cpu.Addr(ds,0x0b35+dir*4))))
                ty:=sy+int(int16(m.Read16(cpu.Addr(ds,0x0b37+dir*4))))
                width:=int(m.Read16(cpu.Addr(ds,0x0b28)));height:=int(m.Read16(cpu.Addr(ds,0x0b2a)))
                mapseg:=m.Read16(cpu.Addr(ds,0x2536));base:=m.Read16(cpu.Addr(ds,0x0b26))
                cell:=m.Read16(cpu.Addr(mapseg,base+uint16((sy*width+sx)*2)))
                target:=uint16(0);attr:=uint16(0);valid:=tx>=0 && ty>=0 && tx<width && ty<height
                if valid {target=m.Read16(cpu.Addr(mapseg,base+uint16((ty*width+tx)*2)));attr=m.Read16(cpu.Addr(ds,0x308e+(target&0xff)*2))}
                fmt.Printf("DQ3_NPC_MOVE_STATE step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x BX=%04x CX=%04x DX=%04x SI=%04x DI=%04x seed=%04x player_x=%d player_y=%d width=%d height=%d count=%d slot=%d cell_word=%04x target_valid=%t target_x=%d target_y=%d target_word=%04x target_attr=%04x npc_slots=%x\n",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.CX],m.CPU.R[cpu.DX],slot,m.CPU.R[cpu.DI],m.Read16(cpu.Addr(ds,0x0b5a)),m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),width,height,count,(slot-0x0b66)/8,cell,valid,tx,ty,target,attr,all)
            }
'''
assert 'm.Write' not in observer and 'Restore' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=h.replace(marker,observer+marker)
ns['__file__']=__file__
ns['main']()
