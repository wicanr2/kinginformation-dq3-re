"""正常409後Enter第一個結果；Issue4 source-only DRAFT，不改production。"""
from pathlib import Path
s=Path('/repo/tools/dosgolem_field_talk_return_probe.py').read_text()
assert s.count(";ns['main']()")==1
s=s.replace(";ns['main']()",'').replace('issue4-talk-empty-return-normal-r1','issue4-command-enter-first-normal-r1')
scope={'__file__':__file__,'__name__':'command_enter_first_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for a,b in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==409','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==410'),('pilotPackets>=409 || pilotUps>80','pilotPackets>=410 || pilotUps>80')]:
 assert h.count(a)==1;h=h.replace(a,b)
marker='                if next!=0 {';assert h.count(marker)==1
inputs=r"""
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=409 {
                    next=0;kind="none"
                    if field && pilotPackets==409 {next=0x1c;kind="field_command_enter_first"}
                }
"""
assert 'm.Write' not in inputs and 'Restore' not in inputs
ns['GATE_HOOK']=h.replace(marker,inputs+marker);ns['__file__']=__file__;ns['main']()
