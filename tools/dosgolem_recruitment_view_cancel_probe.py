"""正常觀看名單、Esc取消、No告別及場景返回；入口 docs/188。"""
from pathlib import Path

code = Path('/repo/tools/dosgolem_recruitment_view_probe.py').read_text()
code = code.replace('issue4-recruit-view-r1', 'issue4-recruit-view-cancel-r1')
assert code.rstrip().endswith("env['rebuilt']['main']()")
code = code[:code.rindex("env['rebuilt']['main']()")]
env = {'__file__': __file__, '__name__': 'view_cancel_builder'}
exec(compile(code, __file__, 'exec'), env)
hook = env['hook']
old = 'pilotStage==6 && pilotRecruitPhase==10 && pilotViewEntered && (field || wait || inline || choice)'
assert hook.count(old) == 1
hook = hook.replace(old, 'pilotStage==6 && pilotRecruitPhase==3 && pilotViewEntered && field')
marker = '                if next!=0 {'
assert hook.count(marker) == 1
hook = hook.replace(marker, '''
                if pilotStage==6 && pilotViewEntered && choice && pilotRecruitPhase==10 && m.Read16(cpu.Addr(ds,0x071e))==1 {
                    next=0x01;kind="recruit_view_cancel";pilotRecruitPhase=2
                }
''' + marker)
env['env']['rebuilt']['GATE_HOOK'] = hook
env['env']['rebuilt']['main']()
