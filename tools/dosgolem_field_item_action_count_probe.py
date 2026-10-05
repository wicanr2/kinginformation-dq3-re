"""正常381後道具動作列數、繞回、返回的DRAFT來源；docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0, '/repo/tools')
s = Path('/repo/tools/dosgolem_field_equipment_armor_replace_probe.py').read_text()
assert s.count("\nns['main']()\n") == 1
s = s.replace("\nns['main']()\n", '\n')
assert s.count('issue4-equip-armor-replace-normal-r1') == 1
s = s.replace('issue4-equip-armor-replace-normal-r1', 'issue4-item-action-count-normal-r1')
scope = {'__file__': __file__, '__name__': 'item_action_count_builder'}
exec(compile(s, __file__, 'exec'), scope)
ns = scope['ns']; h = ns['GATE_HOOK']
for a, b in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==381', 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==394'), ('pilotPackets>=381 || pilotUps>80', 'pilotPackets>=394 || pilotUps>80')]:
    assert h.count(a) == 1
    h = h.replace(a, b)
marker = '                if next!=0 {'
assert h.count(marker) == 1
inputs = r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=381 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==381 {next=0x39;kind="item_count_open_command"}
                    if choice && count==6 && pilotPackets>=382 && pilotPackets<=385 && cursor==uint16(pilotPackets-381) {next=0x50;kind="item_count_command_down"}
                    if choice && count==6 && cursor==5 && pilotPackets==386 {next=0x39;kind="item_count_open_list"}
                    if choice && count==5 && cursor==1 && pilotPackets==387 {next=0x39;kind="item_count_select_first"}
                    if choice && pilotPackets>=388 && pilotPackets<=390 {next=0x50;kind="item_count_action_down"}
                    if choice && pilotPackets==391 {next=0x01;kind="item_count_action_cancel"}
                    if field && pilotPackets==392 {next=0x4b;kind="item_count_after_left"}
                    if field && pilotPackets==393 {next=0x4d;kind="item_count_after_right"}
                }
'''
assert 'm.Write' not in inputs and 'Restore' not in inputs
ns['GATE_HOOK'] = h.replace(marker, inputs + marker)
ns['__file__'] = __file__
ns['main']()
