"""Disposable normal packet230 -> wooden-stick use; evidence entry docs/188."""
import sys
sys.path.insert(0,'/repo/tools')
from dosgolem_field_item_reorder_probe import build

ns=build('issue4-item-use-normal-r2')
hook=ns['GATE_HOOK']
old='pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==230'
assert hook.count(old)==1
hook=hook.replace(old,'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==233')
old='pilotPackets>=230 || pilotUps>80'
assert hook.count(old)==1
hook=hook.replace(old,'pilotPackets>=233 || pilotUps>80')
marker='                if next!=0 {'
assert hook.count(marker)==1
hook=hook.replace(marker,r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=230 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    if choice && count==7 && pilotPackets==230 {next=0x39;kind="use_select_first_row"}
                    if choice && count==3 && pilotPackets==231 && m.Read16(cpu.Addr(ds,0x0722))==1 {next=0x39;kind="use_choose_use"}
                    if wait && pilotPackets==232 {next=0x1c;kind="use_finish_no_effect"}
                }
'''+marker)
marker='            if rawFlag&0x4000!=0'
assert hook.count(marker)==1
observer=r'''
            if pilotPackets>=231 && (pc==0x15002 || pc==0x15023 || pc==0x21414 || pc==0x214b9 || pc==0x2111b || pc==0x21133 || pc==0x13942 || pc==0x13b46 || pc==0x13c6d) {
                fmt.Printf("DQ3_ITEM_USE_OBSERVE step=%d packet=%d ida_linear=%05x DI=%04x SI=%04x BP=%04x DX=%04x BX=%04x AX=%04x owner=%d selected=%d raw2591=%04x raw259b=%d raw259e=%d window3e6e=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.DI],m.CPU.R[cpu.SI],m.CPU.R[cpu.BP],m.CPU.R[cpu.DX],m.CPU.R[cpu.BX],m.CPU.R[cpu.AX],m.Read16(cpu.Addr(ds,0x062d)),m.Read16(cpu.Addr(ds,0x062f)),m.Read16(cpu.Addr(ds,0x2591)),m.Read8(cpu.Addr(ds,0x259b)),m.Read8(cpu.Addr(ds,0x259e)))
                for i:=uint16(0);i<30;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3e6e+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=hook.replace(marker,observer+marker)
ns['__file__']=__file__
ns['main']()
