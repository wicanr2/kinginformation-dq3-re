"""正常280後選全體狀況的DRAFT原版探針；入口docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0,'/repo/tools')
s=Path('/repo/tools/dosgolem_field_status_detail_probe.py').read_text()
assert s.count("\nns['main']()\n")==1
s=s.replace("\nns['main']()\n",'\n')
assert s.count('issue4-field-detail-normal-r5')==1
s=s.replace('issue4-field-detail-normal-r5','issue4-field-summary-normal-r2')
scope={'__file__':__file__,'__name__':'summary_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for old,new in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==280','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==288'),('pilotPackets>=280 || pilotUps>80','pilotPackets>=288 || pilotUps>80')]:
 assert h.count(old)==1
 h=h.replace(old,new)
marker='                if next!=0 {'
assert h.count(marker)==1
h=h.replace(marker,r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=280 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==280 {next=0x39;kind="summary_open_command"}
                    if choice && count==6 && pilotPackets==281 {next=0x50;kind="summary_command_down"}
                    if choice && count==6 && cursor==2 && pilotPackets==282 {next=0x39;kind="summary_open_status"}
                    if choice && count==3 && cursor==1 && pilotPackets==283 {next=0x50;kind="summary_status_down"}
                    if choice && count==3 && cursor==2 && pilotPackets==284 {next=0x39;kind="summary_select"}
                    if wait && pilotPackets==285 {next=0x1c;kind="summary_close_page"}
                    if field && pilotPackets==286 {next=0x4b;kind="summary_after_left"}
                    if field && pilotPackets==287 {next=0x4d;kind="summary_after_right"}
                }
'''+marker)
marker='            if rawFlag&0x4000!=0'
assert h.count(marker)==1
observer=r'''
            if pilotPackets>=285 && (pc==0x185ef || pc==0x1f4e3 || pc==0x1f590 || pc==0x2111b) {
                fmt.Printf("DQ3_FIELD_SUMMARY_OBSERVE step=%d packet=%d ida_linear=%05x SI=%04x DI=%04x BP=%04x DX=%04x AX=%04x BX=%04x raw0716=%d raw0718=%d window_si=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.BP],m.CPU.R[cpu.DX],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.Read16(cpu.Addr(ds,0x0716)),m.Read16(cpu.Addr(ds,0x0718)))
                for i:=uint16(0);i<28;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,m.CPU.R[cpu.SI]+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=h.replace(marker,observer+marker)
ns['__file__']=__file__
ns['main']()
