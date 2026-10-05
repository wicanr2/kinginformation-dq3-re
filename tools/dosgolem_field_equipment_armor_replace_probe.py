"""正常366後同部位兩個物理格換穿、詳細狀況與返回；DRAFT docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0,'/repo/tools')
s=Path('/repo/tools/dosgolem_field_equipment_armor_probe.py').read_text()
assert s.count("\nns['main']()\n")==1
s=s.replace("\nns['main']()\n",'\n')
assert s.count('issue4-equip-armor-normal-r1')==1
s=s.replace('issue4-equip-armor-normal-r1','issue4-equip-armor-replace-normal-r1')
scope={'__file__':__file__,'__name__':'equipment_armor_replace_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for a,b in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==366','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==381'),('pilotPackets>=366 || pilotUps>80','pilotPackets>=381 || pilotUps>80')]:
    assert h.count(a)==1;h=h.replace(a,b)
marker='                if next!=0 {';assert h.count(marker)==1
inputs=r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=366 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==366 {next=0x39;kind="armor_replace_open_command"}
                    if choice && count==6 && cursor==1 && pilotPackets==367 {next=0x50;kind="armor_replace_command_down_one"}
                    if choice && count==6 && cursor==2 && pilotPackets==368 {next=0x50;kind="armor_replace_command_down_two"}
                    if choice && count==6 && cursor==3 && pilotPackets==369 {next=0x39;kind="armor_replace_select_equipment"}
                    if choice && count==3 && cursor==1 && pilotPackets==370 {next=0x01;kind="armor_replace_skip_weapon"}
                    if choice && count==4 && cursor==1 && pilotPackets==371 {next=0x39;kind="armor_replace_confirm_physical_four"}
                    if choice && count==1 && (pilotPackets==372 || pilotPackets==373) {next=0x01;kind="armor_replace_skip_empty_slot"}
                    if field && pilotPackets==374 {next=0x39;kind="armor_replace_status_open_command"}
                    if choice && count==6 && cursor==1 && pilotPackets==375 {next=0x50;kind="armor_replace_status_command_down"}
                    if choice && count==6 && cursor==2 && pilotPackets==376 {next=0x39;kind="armor_replace_status_open_menu"}
                    if choice && count==3 && cursor==1 && pilotPackets==377 {next=0x39;kind="armor_replace_status_detail"}
                    if wait && pilotPackets==378 {next=0x1c;kind="armor_replace_status_close"}
                    if field && pilotPackets==379 {next=0x4b;kind="armor_replace_after_left"}
                    if field && pilotPackets==380 {next=0x4d;kind="armor_replace_after_right"}
                }
'''
assert 'm.Write' not in inputs and 'Restore' not in inputs
ns['GATE_HOOK']=h.replace(marker,inputs+marker)
ns['__file__']=__file__
ns['main']()
