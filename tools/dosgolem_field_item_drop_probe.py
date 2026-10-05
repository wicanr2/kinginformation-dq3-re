"""正常233後丟兩件，再拒絕穿戴物品，停於261的DRAFT；入口docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0, '/repo/tools')

source = Path('/repo/tools/dosgolem_field_item_use_probe.py').read_text()
assert source.count("ns['main']()") == 1
source = source.replace("ns['main']()", '')
source = source.replace("build('issue4-item-use-normal-r2')", "build('issue4-item-drop-normal-r3')")
space = {'__file__': __file__, '__name__': 'item_drop_builder'}
exec(compile(source, __file__, 'exec'), space)
ns = space['ns']
hook = ns['GATE_HOOK']
for old, new in (
    ('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==233', 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==261'),
    ('pilotPackets>=233 || pilotUps>80', 'pilotPackets>=261 || pilotUps>80')):
    assert hook.count(old) == 1
    hook = hook.replace(old, new)
marker = '                if next!=0 {'
assert hook.count(marker) == 1
hook = hook.replace(marker, r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=233 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==233 {next=0x39;kind="drop_open_command"}
                    if choice && count==6 {
                        switch pilotPackets {
                        case 234:next=0x50;kind="drop_command_down"
                        case 235:next=0x4d;kind="drop_command_right"
                        case 236:next=0x39;kind="drop_open_list"
                        }
                    }
                    if choice && count==7 && pilotPackets==237 && cursor==1 {next=0x39;kind="drop_select_first_row"}

                    if wait && (pilotPackets==241 || pilotPackets==250 || pilotPackets==260) {next=0x1c;kind="drop_close_message"}
                    if field && (pilotPackets==242 || pilotPackets==251) {next=0x39;kind="drop_reopen_command"}
                    if choice && count==6 {
                        switch pilotPackets {
                        case 243,252:next=0x50;kind="drop_reopen_down"
                        case 244,253:next=0x4d;kind="drop_reopen_right"
                        case 245,254:next=0x39;kind="drop_reopen_list"
                        }
                    }
                    if choice && count==6 && pilotPackets==246 && cursor==1 {next=0x39;kind="drop_select_after_hole"}
                    if choice && count==5 && pilotPackets==255 && cursor==1 {next=0x48;kind="drop_wrap_worn"}
                    if choice && count==5 && pilotPackets==256 && cursor==5 {next=0x39;kind="drop_select_worn"}
                    if choice && count==3 {
                        switch pilotPackets {
                        case 238:next=0x50;kind="drop_action_down_give"
                        case 239:next=0x50;kind="drop_action_down_drop"
                        case 240:if cursor==3 {next=0x39;kind="drop_choose_drop"}
                        case 247,257:next=0x50;kind="drop_next_action_give"
                        case 248,258:next=0x50;kind="drop_next_action_drop"
                        case 249,259:if cursor==3 {next=0x39;kind="drop_next_choose_drop"}

                        }
                    }
                }
''' + marker)
marker = '            if rawFlag&0x4000!=0'
assert hook.count(marker) == 1
observer = r'''
            if pilotPackets>=241 && (pc==0x18197 || pc==0x13af8 || pc==0x13abc || pc==0x13919 || pc==0x13ad5 || pc==0x13aeb || pc==0x13aee || pc==0x13af3 || pc==0x13af5 || pc==0x13b01 || pc==0x13b0f || pc==0x13b14 || pc==0x13b46 || pc==0x15002 || pc==0x15023 || pc==0x21414 || pc==0x214b9 || pc==0x2111b || pc==0x21133) {
                fmt.Printf("DQ3_ITEM_DROP_OBSERVE step=%d packet=%d ida_linear=%05x DI=%04x SI=%04x BP=%04x DX=%04x BX=%04x AX=%04x owner=%d selected=%d raw2591=%04x raw259b=%d raw259e=%d window3e6e=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.DI],m.CPU.R[cpu.SI],m.CPU.R[cpu.BP],m.CPU.R[cpu.DX],m.CPU.R[cpu.BX],m.CPU.R[cpu.AX],m.Read16(cpu.Addr(ds,0x062d)),m.Read16(cpu.Addr(ds,0x062f)),m.Read16(cpu.Addr(ds,0x2591)),m.Read8(cpu.Addr(ds,0x259b)),m.Read8(cpu.Addr(ds,0x259e)))
                for i:=uint16(0);i<30;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3e6e+i)))}
                fmt.Printf(" slots=")
                actor:=m.Read16(cpu.Addr(ds,0x4f15))
                for i:=uint16(0);i<16;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,actor+0x3a+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK'] = hook.replace(marker, observer+marker)
ns['__file__'] = __file__
ns['main']()
