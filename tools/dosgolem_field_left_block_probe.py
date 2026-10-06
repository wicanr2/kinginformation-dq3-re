"""Issue4 DRAFT：正常410後最多四次左移，首次非城鎮或非field結果即停。"""
from pathlib import Path

s = Path('/work/issue4-command-enter-first-probe-r1.py').read_text().rstrip()
marker = ";ns['main']()"
assert s.endswith(marker)
s = s[:-len(marker)].replace('issue4-command-enter-first-normal-r1', 'issue4-field-left-transition-normal-r2')
scope = {'__file__': __file__, '__name__': 'field_left_transition_builder'}
exec(compile(s, __file__, 'exec'), scope)
ns = scope['ns']
h = ns['GATE_HOOK']
old = 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==410'
assert h.count(old) == 1
h = h.replace(old, 'pilotStage==6 && pilotRecruitPhase==13 && (pilotPackets==414 || (pilotPackets>=411 && (m.Read16(cpu.Addr(ds,0x4f2d))!=1 || !field)))')
old = 'pilotPackets>=410 || pilotUps>80'
assert h.count(old) == 1
h = h.replace(old, 'pilotPackets>=414 || pilotUps>80')
marker = '                if next!=0 {'
assert h.count(marker) == 1
inputs = r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=410 {
                    next=0;kind="none"
                    if field && pilotPackets>=410 && pilotPackets<414 {next=0x4b;kind="field_left_transition"}
                }
'''
assert 'm.Write' not in inputs and 'Restore' not in inputs
ns['GATE_HOOK'] = h.replace(marker, inputs + marker)
ns['__file__'] = __file__
ns['main']()
