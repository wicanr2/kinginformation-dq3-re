"""正常314後穿戴與卸下的DRAFT原版來源；入口docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0,'/repo/tools')
s=Path('/repo/tools/dosgolem_field_equipment_probe.py').read_text()
assert s.count("\nns['main']()\n")==1
s=s.replace("\nns['main']()\n",'\n')
assert s.count('issue4-equip-entry-normal-r3')==1
s=s.replace('issue4-equip-entry-normal-r3','issue4-equip-wear-normal-r1')
scope={'__file__':__file__,'__name__':'equip_wear_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for old,new in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==314','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==339'),('pilotPackets>=314 || pilotUps>80','pilotPackets>=339 || pilotUps>80')]:
    assert h.count(old)==1;h=h.replace(old,new)
marker='                if next!=0 {'
assert h.count(marker)==1
h=h.replace(marker,r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=314 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && (pilotPackets==314 || pilotPackets==327) {next=0x39;kind="equip_wear_open_command"}
                    if choice && count==6 && cursor==1 && (pilotPackets==315 || pilotPackets==328) {next=0x50;kind="equip_wear_command_down_one"}
                    if choice && count==6 && cursor==2 && (pilotPackets==316 || pilotPackets==329) {next=0x50;kind="equip_wear_command_down_two"}
                    if choice && count==6 && cursor==3 && (pilotPackets==317 || pilotPackets==330) {next=0x39;kind="equip_wear_select"}
                    if choice && count==3 && cursor==1 && pilotPackets==318 {next=0x50;kind="equip_wear_second_weapon"}
                    if choice && count==3 && cursor==2 && pilotPackets==319 {next=0x39;kind="equip_wear_confirm_weapon"}
                    if choice && count==4 && cursor==1 && pilotPackets==320 {next=0x50;kind="equip_wear_armor_down_one"}
                    if choice && count==4 && cursor==2 && pilotPackets==321 {next=0x50;kind="equip_wear_armor_down_two"}
                    if choice && count==4 && cursor==3 && pilotPackets==322 {next=0x39;kind="equip_wear_confirm_same_armor"}
                    if choice && count==1 && (pilotPackets==323 || pilotPackets==324 || pilotPackets==335 || pilotPackets==336) {next=0x01;kind="equip_wear_skip_empty_slot"}
                    if field && (pilotPackets==325 || pilotPackets==337) {next=0x4b;kind="equip_wear_after_left"}
                    if field && (pilotPackets==326 || pilotPackets==338) {next=0x4d;kind="equip_wear_after_right"}
                    if choice && count==3 && cursor==1 && pilotPackets==331 {next=0x48;kind="equip_wear_weapon_none_cursor"}
                    if choice && count==3 && cursor==3 && pilotPackets==332 {next=0x39;kind="equip_wear_weapon_none"}
                    if choice && count==4 && cursor==1 && pilotPackets==333 {next=0x48;kind="equip_wear_armor_none_cursor"}
                    if choice && count==4 && cursor==4 && pilotPackets==334 {next=0x39;kind="equip_wear_armor_none"}
                }
'''+marker)
marker='            if rawFlag&0x4000!=0'
assert h.count(marker)==1
observer=r'''
            if pilotPackets>=318 && (pc==0x17ed9 || pc==0x17f6a || pc==0x17feb || pc==0x1807b || pc==0x18098 || pc==0x180ae || pc==0x180cc || pc==0x180d9 || pc==0x180dd || pc==0x180f3 || pc==0x18197 || pc==0x181b1 || pc==0x18215 || pc==0x1821d) {
                fmt.Printf("DQ3_FIELD_EQUIP_WEAR_OBSERVE step=%d packet=%d ida_linear=%05x SI=%04x DI=%04x AX=%04x BX=%04x CX=%04x DX=%04x BP=%04x raw0638=%d raw0678=%04x raw0682=%04x raw259c=%d raw0726=%d window40dc=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.CX],m.CPU.R[cpu.DX],m.CPU.R[cpu.BP],m.Read16(cpu.Addr(ds,0x0638)),m.Read16(cpu.Addr(ds,0x0678)),m.Read16(cpu.Addr(ds,0x0682)),m.Read16(cpu.Addr(ds,0x259c)),m.Read8(cpu.Addr(ds,0x0726)))
                for i:=uint16(0);i<30;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x40dc+i)))}
                fmt.Printf(" window40fa=")
                for i:=uint16(0);i<24;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x40fa+i)))}
                fmt.Printf(" candidates=")
                for i:=uint16(0);i<24;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x063a+i)))}
                fmt.Printf(" actor=")
                actor:=m.Read16(cpu.Addr(ds,0x4f15))
                for i:=uint16(0);i<128;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,actor+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=h.replace(marker,observer+marker)
ns['__file__']=__file__
ns['main']()
