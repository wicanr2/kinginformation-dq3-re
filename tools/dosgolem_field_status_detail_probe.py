"""正常273後重開狀況並確認首列的DRAFT；入口docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0, '/repo/tools')
s=Path('/repo/tools/dosgolem_field_status_probe.py').read_text()
assert s.count("\nns['main']()\n")==1
s=s.replace("\nns['main']()\n",'\n')
assert s.count('issue4-field-status-normal-r3')==1
s=s.replace('issue4-field-status-normal-r3','issue4-field-detail-normal-r5')
scope={'__file__':__file__,'__name__':'detail_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for old,new in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==273','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==280'),('pilotPackets>=273 || pilotUps>80','pilotPackets>=280 || pilotUps>80')]:
 assert h.count(old)==1
 h=h.replace(old,new)
marker='                if next!=0 {'
assert h.count(marker)==1
h=h.replace(marker,r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=273 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==273 {next=0x39;kind="detail_open_command"}
                    if choice && count==6 && pilotPackets==274 {next=0x50;kind="detail_command_down"}
                    if choice && count==6 && cursor==2 && pilotPackets==275 {next=0x39;kind="detail_open_status"}
                    if choice && count==3 && cursor==1 && pilotPackets==276 {next=0x39;kind="detail_select_first"}
                }
'''+marker)
marker='                if next!=0 {'
assert h.count(marker)==1
h=h.replace(marker,r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=277 {
                    next=0;kind="none"
                    if wait && pilotPackets==277 {next=0x1c;kind="detail_close_page"}
                    if field && pilotPackets==278 {next=0x4b;kind="detail_after_left"}
                    if field && pilotPackets==279 {next=0x4d;kind="detail_after_right"}
                }
'''+marker)
marker='            if rawFlag&0x4000!=0'
assert h.count(marker)==1
observer=r'''
            if pilotPackets>=277 && (pc==0x18313 || pc==0x18338 || pc==0x1834e || pc==0x1f4e3 || pc==0x2111b || pc==0x21103) {
                fmt.Printf("DQ3_FIELD_DETAIL_OBSERVE step=%d packet=%d ida_linear=%05x SI=%04x DI=%04x AX=%04x BX=%04x raw4f1d=%04x raw5077=%d window3da8=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.Read16(cpu.Addr(ds,0x4f1d)),m.Read8(cpu.Addr(ds,0x5077)))
                for i:=uint16(0);i<28;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3da8+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=h.replace(marker,observer+marker)
ns['__file__']=__file__
ns['main']()
