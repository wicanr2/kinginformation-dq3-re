"""正常304後單人裝備入口與取消的DRAFT探針；入口docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0, '/repo/tools')
source = Path('/repo/tools/dosgolem_field_spell_empty_probe.py').read_text()
assert source.count("\nns['main']()\n") == 1
source = source.replace("\nns['main']()\n", '\n')
assert source.count('issue4-spell-empty-normal-r1') == 1
source = source.replace('issue4-spell-empty-normal-r1', 'issue4-equip-entry-normal-r3')
scope = {'__file__': __file__, '__name__': 'equip_entry_builder'}
exec(compile(source, __file__, 'exec'), scope)
ns = scope['ns']
hook = ns['GATE_HOOK']
for old, new in (
    ('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==304', 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==314'),
    ('pilotPackets>=304 || pilotUps>80', 'pilotPackets>=314 || pilotUps>80'),
):
    assert hook.count(old) == 1
    hook = hook.replace(old, new)
marker = '                if next!=0 {'
assert hook.count(marker) == 1
hook = hook.replace(marker, r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=304 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==304 {next=0x39;kind="equip_entry_open_command"}
                    if choice && count==6 && cursor==1 && pilotPackets==305 {next=0x50;kind="equip_entry_command_down_one"}
                    if choice && count==6 && cursor==2 && pilotPackets==306 {next=0x50;kind="equip_entry_command_down_two"}
                    if choice && count==6 && cursor==3 && pilotPackets==307 {next=0x39;kind="equip_entry_select"}
                    if choice && pilotPackets>=308 && pilotPackets<=311 {next=0x01;kind="equip_entry_skip_slot"}
                    if field && pilotPackets==312 {next=0x4b;kind="equip_entry_after_left"}
                    if field && pilotPackets==313 {next=0x4d;kind="equip_entry_after_right"}
                }
''' + marker)
marker = '            if rawFlag&0x4000!=0'
assert hook.count(marker) == 1
observer = r'''
            if pilotPackets>=308 && (pc==0x17e12 || pc==0x17e26 || pc==0x17ed9 || pc==0x18006 || pc==0x18060 || pc==0x18197 || pc==0x1885f || pc==0x18869 || pc==0x1f4e3 || pc==0x1f779 || pc==0x1f908 || pc==0x15023 || pc==0x21414 || pc==0x2111b) {
                fmt.Printf("DQ3_FIELD_EQUIP_ENTRY_OBSERVE step=%d packet=%d ida_linear=%05x SI=%04x DI=%04x AX=%04x BX=%04x party_count=%d choice_count=%d choice_cursor=%d actor_pointer=%04x\n",m.Steps,pilotPackets,pc,m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.Read8(cpu.Addr(ds,0x5077)),m.Read16(cpu.Addr(ds,0x071e)),m.Read16(cpu.Addr(ds,0x0722)),m.Read16(cpu.Addr(ds,0x4f15)))
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK'] = hook.replace(marker, observer + marker)
ns['__file__'] = __file__
ns['main']()
