"""正常394後調查首個結果DRAFT，未預設消息或返回；docs/188。"""
from pathlib import Path
import sys
sys.path.insert(0, '/repo/tools')
s = Path('/repo/tools/dosgolem_field_item_action_count_probe.py').read_text()
assert s.count("\nns['main']()\n") == 1
s = s.replace("\nns['main']()\n", '\n')
assert s.count('issue4-item-action-count-normal-r1') == 1
s = s.replace('issue4-item-action-count-normal-r1', 'issue4-search-first-normal-r1')
scope = {'__file__': __file__, '__name__': 'search_first_builder'}
exec(compile(s, __file__, 'exec'), scope)
ns = scope['ns']; h = ns['GATE_HOOK']
for a, b in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==394', 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==401'), ('pilotPackets>=394 || pilotUps>80', 'pilotPackets>=401 || pilotUps>80')]:
    assert h.count(a) == 1
    h = h.replace(a, b)
marker = '                if next!=0 {'
assert h.count(marker) == 1
inputs = r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=394 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    cursor:=m.Read16(cpu.Addr(ds,0x0722))
                    if field && pilotPackets==394 {next=0x39;kind="search_open_command"}
                    if choice && count==6 && pilotPackets>=395 && pilotPackets<=399 && cursor==uint16(pilotPackets-394) {next=0x50;kind="search_command_down"}
                    if choice && count==6 && cursor==6 && pilotPackets==400 {next=0x39;kind="search_select"}
                }
'''
assert 'm.Write' not in inputs and 'Restore' not in inputs
ns['GATE_HOOK'] = h.replace(marker, inputs + marker)
ns['__file__'] = __file__
ns['main']()
