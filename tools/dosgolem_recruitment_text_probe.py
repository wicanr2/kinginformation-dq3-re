"""正常入隊至 record536 第一個內文等待；不進音樂完成探查。"""
from pathlib import Path

code = Path('/repo/tools/dosgolem_recruitment_selection_probe.py').read_text()
code = code.replace('issue4-recruit-selection-cancel-r1', 'issue4-recruit-text-r1')
assert code.rstrip().endswith("rebuilt['main']()")
code = code[:code.rindex("rebuilt['main']()")]
env = {'__file__': __file__, '__name__': 'join_text_builder'}
exec(compile(code, __file__, 'exec'), env)
hook = env['hook']
old = 'next=0x01;kind="recruit_selection_cancel";pilotRecruitPhase=2'
assert hook.count(old) == 1
hook = hook.replace(old, 'next=0x1c;kind="recruit_pick";pilotRecruitPhase=2')
marker = '                if pilotPackets==172 && field {pilotFinishing=false;pilotStage=5}'
assert hook.count(marker) == 1
stop = '''
                if pilotPackets==198 && inline && pilotRecord==536 {
                    fmt.Printf("DQ3_RECRUIT_DONE step=%d packets=%d irqs=%d scope=record536_first_inline_wait complete_join=false audio_completion=false\\n",m.Steps,pilotPackets,m.KeyIRQs)
                    break
                }
'''
hook = hook.replace(marker, stop + marker)
env['rebuilt']['GATE_HOOK'] = hook
env['rebuilt']['main']()
