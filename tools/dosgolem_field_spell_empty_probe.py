"""正常298續行空咒文候選；入口docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0,'/repo/tools')
s=Path('/repo/tools/dosgolem_field_status_reorder_probe.py').read_text()
assert s.count("\nns['main']()\n")==1
s=s.replace("\nns['main']()\n",'\n')
assert s.count('issue4-field-reorder-normal-r2')==1
s=s.replace('issue4-field-reorder-normal-r2','issue4-spell-empty-normal-r1')
scope={'__file__':__file__,'__name__':'spell_empty_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for old,new in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==298','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==304'),('pilotPackets>=298 || pilotUps>80','pilotPackets>=304 || pilotUps>80')]:
 assert h.count(old)==1;h=h.replace(old,new)
marker='                if next!=0 {'
assert h.count(marker)==1
h=h.replace(marker,r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=298 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==298 {next=0x39;kind="spell_empty_open_command"}
                    if choice && count==6 && cursor==1 && pilotPackets==299 {next=0x4d;kind="spell_empty_command_right"}
                    if choice && count==6 && cursor==4 && pilotPackets==300 {next=0x39;kind="spell_empty_select"}
                    if wait && pilotPackets==301 {next=0x1c;kind="spell_empty_close_message"}
                    if field && pilotPackets==302 {next=0x4b;kind="spell_empty_after_left"}
                    if field && pilotPackets==303 {next=0x4d;kind="spell_empty_after_right"}
                }
'''+marker)
marker='            if rawFlag&0x4000!=0'
assert h.count(marker)==1
observer=r'''
            if pilotPackets>=301 && (pc==0x1c9c1 || pc==0x1885f || pc==0x18869 || pc==0x1c9ee || pc==0x1ca00 || pc==0x1cb39 || pc==0x1c9e7 || pc==0x15023 || pc==0x21414 || pc==0x2111b) {
                fmt.Printf("DQ3_FIELD_SPELL_EMPTY_OBSERVE step=%d packet=%d ida_linear=%05x SI=%04x DI=%04x AX=%04x BX=%04x party_count=%d raw0716=%d raw0718=%d spell30=%d spell31=%d window3e6e=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.Read8(cpu.Addr(ds,0x5077)),m.Read16(cpu.Addr(ds,0x0716)),m.Read16(cpu.Addr(ds,0x0718)),m.Read8(cpu.Addr(ds,0x4f17+0x30)),m.Read8(cpu.Addr(ds,0x4f17+0x31)))
                for i:=uint16(0);i<24;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3e6e+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=h.replace(marker,observer+marker)
ns['__file__']=__file__
ns['main']()
