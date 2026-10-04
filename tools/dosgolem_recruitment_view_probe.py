"""正常冷啟動至酒場觀看名單的第一個等待點；不注入遊戲狀態，入口 docs/188。"""
from pathlib import Path

code = Path('/repo/tools/dosgolem_recruitment_selection_probe.py').read_text()
code = code.replace('issue4-recruit-selection-cancel-r1', 'issue4-recruit-view-r1')
assert code.count('var pilotRecruitMotion, pilotRecruitPhase int') == 1
code = code.replace('var pilotRecruitMotion, pilotRecruitPhase int',
                    'var pilotRecruitMotion, pilotRecruitPhase int\n    var pilotViewEntered bool')
assert code.rstrip().endswith("rebuilt['main']()")
code = code[:code.rindex("rebuilt['main']()")]
env = {'__file__': __file__, '__name__': 'view_builder'}
exec(compile(code, __file__, 'exec'), env)
hook = env['hook']
old = 'pilotStage==6 && pilotRecruitPhase==3 && field'
assert hook.count(old) == 1
hook = hook.replace(old, 'pilotStage==6 && pilotRecruitPhase==10 && pilotViewEntered && (field || wait || inline || choice)')
old = 'if choice && pilotRecruitPhase==0 && m.Read16(cpu.Addr(ds,0x071e))==3 {next=0x1c;kind="recruit_join";pilotRecruitPhase=1}'
assert hook.count(old) == 1
hook = hook.replace(old, '''if choice && pilotRecruitPhase==0 && m.Read16(cpu.Addr(ds,0x071e))==3 {
                        next=0x50;kind="recruit_view_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==3 {next=0x1c;kind="recruit_view";pilotRecruitPhase=10}
                    }''')
marker = '            if rawFlag&0x4000!=0'
assert hook.count(marker) == 1
observer = r'''
            if pilotPackets>=196 && pc==0x10624 {
                if m.CPU.Seg[cpu.DS]!=ds {panic("觀看入口DS不符")}
                pilotViewEntered=true
                fmt.Printf("DQ3_VIEW_ENTRY step=%d packet=%d ida_linear=10624 DS=%04x AX=%04x BX=%04x SI=%04x DI=%04x choice_count=%d choice_cursor=%d\n",m.Steps,pilotPackets,ds,m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.Read16(cpu.Addr(ds,0x071e)),m.Read16(cpu.Addr(ds,0x0722)))
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
hook = hook.replace(marker, observer + marker)
env['rebuilt']['GATE_HOOK'] = hook
env['rebuilt']['main']()
