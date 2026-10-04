"""正常取消選Yes、再進清單、取消選No；只讀取人物取圖，不改相位或時間。"""
from pathlib import Path
code=Path('/repo/tools/dosgolem_recruitment_selection_probe.py').read_text()
code=code.replace('issue4-recruit-selection-cancel-r1','issue4-recruit-yes-r1')
assert code.rstrip().endswith("rebuilt['main']()")
code=code[:code.rindex("rebuilt['main']()")]
env={'__file__':__file__,'__name__':'yes_builder'}
exec(compile(code,__file__,'exec'),env)
hook=env['hook']
old='pilotStage==6 && pilotRecruitPhase==3 && field'
assert hook.count(old)==1
hook=hook.replace(old,'pilotStage==6 && pilotRecruitPhase==6 && field')
old='''                    if choice && pilotRecruitPhase==2 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x4d;kind="recruit_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="recruit_decline";pilotRecruitPhase=3}
                    }'''
assert hook.count(old)==1
hook=hook.replace(old,'''                    if choice && pilotRecruitPhase==2 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        if m.Read16(cpu.Addr(ds,0x0722))!=1 {panic("Yes初始游標不符")}
                        next=0x1c;kind="recruit_continue_yes";pilotRecruitPhase=3
                    }
                    if choice && pilotRecruitPhase==3 && pilotRecord==528 && m.Read16(cpu.Addr(ds,0x071e))==3 {next=0x1c;kind="recruit_join_again";pilotRecruitPhase=4}
                    if choice && pilotRecruitPhase==4 && pilotRecord==530 && m.Read16(cpu.Addr(ds,0x071e))==1 {next=0x01;kind="recruit_selection_cancel_again";pilotRecruitPhase=5}
                    if choice && pilotRecruitPhase==5 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x4d;kind="recruit_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="recruit_decline";pilotRecruitPhase=6}
                    }''')
marker='            if rawFlag&0x4000!=0'
assert hook.count(marker)==1
observer=r'''
            if pilotPackets>=194 && (pc==0x11ed0 || pc==0x11ee8 || pc==0x1e2ff || pc==0x1e30b) {
                if m.CPU.Seg[cpu.DS]!=ds {panic("取圖DS不符")}
                fmt.Printf("DQ3_RECRUIT_SPRITE step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x BX=%04x DX=%04x SI=%04x DI=%04x BP=%04x counter0002=%d phase0004=%d ticks=%d player_x=%d player_y=%d\n",m.Steps,pilotPackets,pc,ds,m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.DX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.BP],m.Read16(cpu.Addr(ds,2)),m.Read8(cpu.Addr(ds,4)),m.Ticks,m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)))
            }
'''
hook=hook.replace(marker,observer+marker)
assert 'm.Write' not in observer and 'SetNextKey' not in observer
env['rebuilt']['GATE_HOOK']=hook
env['rebuilt']['main']()
