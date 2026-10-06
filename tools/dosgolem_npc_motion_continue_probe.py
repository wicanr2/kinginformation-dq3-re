"""正常422之後固定13右、5下、13左、5上；首個非field結果即停。"""
from pathlib import Path
s=Path('/work/issue4-npc-move-probe-r1.py').read_text().rstrip()
marker="\nns['main']()"
assert s.endswith(marker)
s=s[:-len(marker)].replace('issue4-npc-move-normal-r1','issue4-npc-move-continue-normal-r1')
s=s.replace('pilotPackets<=422','pilotPackets<=458')
s=s.replace('pc==0x120aa || pc==0x120a6','pc==0x120aa || pc==0x120b0 || pc==0x121ad || pc==0x121b2 || pc==0x120a6')
scope={'__file__':__file__,'__name__':'npc_mover_continue_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
old='pilotPackets==422 || (pilotPackets>=415'
assert h.count(old)==1
h=h.replace(old,'pilotPackets==458 || (pilotPackets>=415')
old='pilotPackets>=422 || pilotUps>80'
assert h.count(old)==1
h=h.replace(old,'pilotPackets>=458 || pilotUps>80')
marker='                if next!=0 {'
assert h.count(marker)==1
inputs=r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=422 {
                    next=0;kind="none"
                    if field && pilotPackets<458 {
                        k:=pilotPackets-422
                        if k<13 {next=0x4d;kind="npc_move_right"} else if k<18 {next=0x50;kind="npc_move_down"} else if k<31 {next=0x4b;kind="npc_move_left"} else {next=0x48;kind="npc_move_up"}
                    }
                }
'''
assert 'm.Write' not in inputs and 'Restore' not in inputs
ns['GATE_HOOK']=h.replace(marker,inputs+marker)
ns['__file__']=__file__
ns['main']()
