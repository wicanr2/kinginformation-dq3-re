"""Issue4 DRAFT：正常414後右三次、下五次，首次非field結果即停。"""
from pathlib import Path
s = Path('/work/issue4-field-left-transition-probe-r1.py').read_text().rstrip()
marker = "\nns['main']()"
assert s.endswith(marker)
s = s[:-len(marker)].replace('issue4-field-left-transition-normal-r2', 'issue4-field-room-door-normal-r1')
scope = {'__file__': __file__, '__name__': 'room_door_builder'}
exec(compile(s, __file__, 'exec'), scope)
ns = scope['ns']
h = ns['GATE_HOOK']
old = 'pilotStage==6 && pilotRecruitPhase==13 && (pilotPackets==414 || (pilotPackets>=411 && (m.Read16(cpu.Addr(ds,0x4f2d))!=1 || !field)))'
assert h.count(old) == 1
h = h.replace(old, 'pilotStage==6 && pilotRecruitPhase==13 && (pilotPackets==422 || (pilotPackets>=415 && (m.Read16(cpu.Addr(ds,0x4f2d))!=1 || !field)))')
old = 'pilotPackets>=414 || pilotUps>80'
assert h.count(old) == 1
h = h.replace(old, 'pilotPackets>=422 || pilotUps>80')
marker = '                if next!=0 {'
assert h.count(marker) == 1
inputs = r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=414 {
                    next=0;kind="none"
                    if field && pilotPackets>=414 && pilotPackets<417 {next=0x4d;kind="field_room_door_right"}
                    if field && pilotPackets>=417 && pilotPackets<422 {next=0x50;kind="field_room_door_down"}
                }
'''
assert 'm.Write' not in inputs and 'Restore' not in inputs
ns['GATE_HOOK'] = h.replace(marker, inputs + marker)
ns['__file__'] = __file__
ns['main']()
