"""正常199包至538返回、音樂等待入口；不執行或偽造播放完成。"""
from pathlib import Path
code=Path('/repo/tools/dosgolem_recruitment_text_probe.py').read_text()
code=code.replace('issue4-recruit-text-r1', 'issue4-recruit-text-r2')
code=code.replace('if pilotPackets==198 && inline && pilotRecord==536 {',
                  'if pilotPackets==199 && audioWait && pilotRecord==538 {')
code=code.replace('scope=record536_first_inline_wait', 'scope=record538_audio_wait_entry')
marker="env['rebuilt']['GATE_HOOK'] = hook"
assert code.count(marker)==1
extra='''
marker='            phase:="none"'
assert hook.count(marker)==1
hook=hook.replace(marker,'            audioWait:=pc==0x208e2 && pilotStage==6 && pilotPackets==199 && pilotRecord==538\\n'+marker)
marker='            if field {phase="ready"}'
assert hook.count(marker)==1
hook=hook.replace(marker,'            if audioWait {phase="audio_wait"}\\n'+marker)
'''
code=code.replace(marker,extra+marker)
exec(compile(code,__file__,'exec'),{'__file__':__file__,'__name__':'__main__'})
