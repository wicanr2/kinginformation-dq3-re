"""正常288後選重新排序的DRAFT原版探針；入口docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0,'/repo/tools')
s=Path('/repo/tools/dosgolem_field_party_summary_probe.py').read_text()
assert s.count("\nns['main']()\n")==1
s=s.replace("\nns['main']()\n",'\n')
assert s.count('issue4-field-summary-normal-r2')==1
s=s.replace('issue4-field-summary-normal-r2','issue4-field-reorder-normal-r2')
scope={'__file__':__file__,'__name__':'reorder_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for old,new in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==288','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==298'),('pilotPackets>=288 || pilotUps>80','pilotPackets>=298 || pilotUps>80')]:
 assert h.count(old)==1
 h=h.replace(old,new)
marker='                if next!=0 {'
assert h.count(marker)==1
h=h.replace(marker,r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=288 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==288 {next=0x39;kind="reorder_open_command"}
                    if choice && count==6 && pilotPackets==289 {next=0x50;kind="reorder_command_down"}
                    if choice && count==6 && cursor==2 && pilotPackets==290 {next=0x39;kind="reorder_open_status"}
                    if choice && count==3 && cursor==1 && pilotPackets==291 {next=0x50;kind="reorder_status_down_one"}
                    if choice && count==3 && cursor==2 && pilotPackets==292 {next=0x50;kind="reorder_status_down_two"}
                    if choice && count==3 && cursor==3 && pilotPackets==293 {next=0x39;kind="reorder_select"}
                    if inline && pilotPackets==294 {next=0x1c;kind="reorder_next_page"}
                    if wait && pilotPackets==295 {next=0x1c;kind="reorder_close_message"}
                    if field && pilotPackets==296 {next=0x4b;kind="reorder_after_left"}
                    if field && pilotPackets==297 {next=0x4d;kind="reorder_after_right"}
                }
'''+marker)
marker='            if rawFlag&0x4000!=0'
assert h.count(marker)==1
observer=r'''
            if pilotPackets>=294 && (pc==0x18685 || pc==0x18694 || pc==0x15023 || pc==0x21414 || pc==0x2111b || pc==0x1f4e3) {
                fmt.Printf("DQ3_FIELD_REORDER_OBSERVE step=%d packet=%d ida_linear=%05x SI=%04x DI=%04x BP=%04x DX=%04x AX=%04x BX=%04x party_count=%d raw0716=%d raw0718=%d window3e6e=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.BP],m.CPU.R[cpu.DX],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.Read8(cpu.Addr(ds,0x5077)),m.Read16(cpu.Addr(ds,0x0716)),m.Read16(cpu.Addr(ds,0x0718)))
                for i:=uint16(0);i<24;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3e6e+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=h.replace(marker,observer+marker)
ns['__file__']=__file__
ns['main']()
