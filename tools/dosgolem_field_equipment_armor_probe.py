"""正常339後穿戴另一物理格甲胄、狀況及第二槽存讀檔；DRAFT docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0,'/repo/tools')
s=Path('/repo/tools/dosgolem_field_equipment_wear_probe.py').read_text()
assert s.count("\nns['main']()\n")==1
s=s.replace("\nns['main']()\n",'\n')
assert s.count('issue4-equip-wear-normal-r1')==1
s=s.replace('issue4-equip-wear-normal-r1','issue4-equip-armor-normal-r1')
scope={'__file__':__file__,'__name__':'equipment_armor_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for a,b in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==339','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==366'),('pilotPackets>=339 || pilotUps>80','pilotPackets>=366 || pilotUps>80')]:
 assert h.count(a)==1;h=h.replace(a,b)
marker='                if next!=0 {';assert h.count(marker)==1
inputs=r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=339 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==339 {next=0x39;kind="armor_open_command"}
                    if choice && count==6 && cursor==1 && pilotPackets==340 {next=0x50;kind="armor_command_down_one"}
                    if choice && count==6 && cursor==2 && pilotPackets==341 {next=0x50;kind="armor_command_down_two"}
                    if choice && count==6 && cursor==3 && pilotPackets==342 {next=0x39;kind="armor_select_equipment"}
                    if choice && count==3 && cursor==1 && pilotPackets==343 {next=0x01;kind="armor_skip_weapon"}
                    if choice && count==4 && cursor==1 && pilotPackets==344 {next=0x50;kind="armor_second_physical_copy"}
                    if choice && count==4 && cursor==2 && pilotPackets==345 {next=0x39;kind="armor_confirm_physical_five"}
                    if choice && count==1 && (pilotPackets==346 || pilotPackets==347) {next=0x01;kind="armor_skip_empty_slot"}
                    if field && pilotPackets==348 {next=0x39;kind="armor_status_open_command"}
                    if choice && count==6 && cursor==1 && pilotPackets==349 {next=0x50;kind="armor_status_command_down"}
                    if choice && count==6 && cursor==2 && pilotPackets==350 {next=0x39;kind="armor_status_open_menu"}
                    if choice && count==3 && cursor==1 && pilotPackets==351 {next=0x39;kind="armor_status_detail"}
                    if wait && pilotPackets==352 {next=0x1c;kind="armor_status_close"}
                    if field && pilotPackets==353 {next=0x4b;kind="armor_status_after_left"}
                    if field && pilotPackets==354 {next=0x4d;kind="armor_status_after_right"}
                    if field && pilotPackets==355 {next=0x3f;kind="armor_save_f5"}
                    if choice && count==2 && cursor==1 && pilotPackets==356 {next=0x1c;kind="armor_save_accept"}
                    if choice && count==10 && cursor==1 && pilotPackets==357 {next=0x50;kind="armor_save_second_cursor"}
                    if choice && count==10 && cursor==2 && pilotPackets==358 {next=0x1c;kind="armor_save_second_slot"}
                    if wait && pilotRecord==252 && pilotPackets==359 {next=0x1c;kind="armor_save_close"}
                    if field && pilotPackets==360 {next=0x4b;kind="armor_after_save_left"}
                    if field && pilotPackets==361 {next=0x40;kind="armor_load_f6"}
                    if choice && count==10 && cursor==1 && pilotPackets==362 {next=0x50;kind="armor_load_second_cursor"}
                    if choice && count==10 && cursor==2 && pilotPackets==363 {next=0x1c;kind="armor_load_second_slot"}
                    if field && pilotPackets==364 {next=0x4b;kind="armor_after_load_left"}
                    if field && pilotPackets==365 {next=0x4d;kind="armor_after_load_right"}
                }
'''
assert 'm.Write' not in inputs
h=h.replace(marker,inputs+marker)
marker='            if rawFlag&0x4000!=0';assert h.count(marker)==1
observer=r'''
            if pilotPackets>=340 && (pc==0x1807b || pc==0x18098 || pc==0x180ae || pc==0x18197 || pc==0x1821d || pc==0x18313 || pc==0x18338 || pc==0x1834e || pc==0x11484 || pc==0x114c8 || pc==0x114d3 || pc==0x114d9 || pc==0x1157d || pc==0x1158b || pc==0x11591 || pc==0x1165f) {
                fmt.Printf("DQ3_ARMOR_NATIVE step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x CX=%04x DX=%04x raw0722=%d raw0726=%d raw01f0=%d clock=%d actor=",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.AX],m.CPU.R[cpu.CX],m.CPU.R[cpu.DX],m.Read16(cpu.Addr(ds,0x0722)),m.Read8(cpu.Addr(ds,0x0726)),m.Read8(cpu.Addr(ds,0x01f0)),m.Read16(cpu.Addr(ds,0x001f)))
                actor:=m.Read16(cpu.Addr(ds,0x4f15))
                for i:=uint16(0);i<128;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,actor+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=h.replace(marker,observer+marker)
ns['__file__']=__file__
ns['main']()
