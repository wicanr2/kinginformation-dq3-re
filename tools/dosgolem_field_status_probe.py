"""正常261後狀況選單、上下導航、Esc與下一步 DRAFT；入口 docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0, '/repo/tools')

source = Path('/repo/tools/dosgolem_field_item_drop_probe.py').read_text()
assert source.count("\nns['main']()\n") == 1
source = source.replace("\nns['main']()\n", "\n")
assert source.count('issue4-item-drop-normal-r3') == 1
source = source.replace('issue4-item-drop-normal-r3', 'issue4-field-status-normal-r3')
space = {'__file__': __file__, '__name__': 'status_builder'}
exec(compile(source, __file__, 'exec'), space)
ns = space['ns']
hook = ns['GATE_HOOK']
for old, new in (
    ('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==261', 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==273'),
    ('pilotPackets>=261 || pilotUps>80', 'pilotPackets>=273 || pilotUps>80')):
    assert hook.count(old) == 1
    hook = hook.replace(old, new)
marker = '                if next!=0 {'
assert hook.count(marker) == 1
hook = hook.replace(marker, r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=261 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==261 {next=0x39;kind="status_open_command"}
                    if choice && count==6 && pilotPackets==262 {next=0x50;kind="status_command_down"}
                    if choice && count==6 && cursor==2 && pilotPackets==263 {next=0x39;kind="status_select"}
                    if choice && count==3 {
                        switch pilotPackets {
                        case 264,265,266:next=0x50;kind="status_down"
                        case 267,268,269:next=0x48;kind="status_up"
                        case 270:next=0x01;kind="status_cancel"
                        }
                    }
                    if field && pilotPackets==271 {next=0x4b;kind="status_after_left"}
                    if field && pilotPackets==272 {next=0x4d;kind="status_after_right"}
                }
''' + marker)
marker = '                pilotPending=false'
assert hook.count(marker) == 1
observer = r'''
                if pilotPackets>=262 {
                    fmt.Printf("DQ3_STATUS_RASTER step=%d packet=%d SI=%04x raw258f=%d raw0727=%d raw071c=%d raw0710=%d raw0712=%d raw0716=%d raw0718=%d window3eb4=",m.Steps,pilotPackets,m.CPU.R[cpu.SI],m.Read8(cpu.Addr(ds,0x258f)),m.Read8(cpu.Addr(ds,0x0727)),m.Read16(cpu.Addr(ds,0x071c)),m.Read16(cpu.Addr(ds,0x0710)),m.Read16(cpu.Addr(ds,0x0712)),m.Read16(cpu.Addr(ds,0x0716)),m.Read16(cpu.Addr(ds,0x0718)))
                    for i:=uint16(0);i<42;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3eb4+i)))}
                    fmt.Printf("\n")
                }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK'] = hook.replace(marker, observer+marker)
ns['__file__'] = __file__
ns['main']()
