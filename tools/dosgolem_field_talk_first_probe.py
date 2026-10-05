"""正常404後首項對話的第一個無對象結果，Issue4 DRAFT；不改正式remake。"""
from pathlib import Path
import sys
sys.path.insert(0,'/repo/tools')
s=Path('/repo/tools/dosgolem_field_search_return_probe.py').read_text()
assert s.count("\nns['main']()\n")==1
s=s.replace("\nns['main']()\n",'\n')
assert s.count('issue4-search-return-normal-r1')==1
s=s.replace('issue4-search-return-normal-r1','issue4-talk-empty-first-normal-r2')
scope={'__file__':__file__,'__name__':'talk_empty_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for a,b in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==404','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==406'),('pilotPackets>=404 || pilotUps>80','pilotPackets>=406 || pilotUps>80')]:
 assert h.count(a)==1;h=h.replace(a,b)
marker='                if next!=0 {'
assert h.count(marker)==1
inputs=r"""
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=404 {
                    next=0;kind="none"
                    if field && pilotPackets==404 {next=0x39;kind="talk_empty_command"}
                    if choice && pilotPackets==405 {next=0x39;kind="talk_empty_select"}
                }
"""
assert 'm.Write' not in inputs and 'Restore' not in inputs
ns['GATE_HOOK']=h.replace(marker,inputs+marker)
ns['__file__']=__file__
ns['main']()
