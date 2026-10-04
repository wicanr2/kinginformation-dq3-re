"""正常觀看名單、詳細狀況、Esc關閉、No告別與場景返回；入口 docs/188。"""
from pathlib import Path

code=Path('/repo/tools/dosgolem_recruitment_view_probe.py').read_text()
code=code.replace('issue4-recruit-view-r1','issue4-recruit-view-detail-r1')
assert code.rstrip().endswith("env['rebuilt']['main']()")
code=code[:code.rindex("env['rebuilt']['main']()")]
env={'__file__':__file__,'__name__':'view_detail_builder'}
exec(compile(code,__file__,'exec'),env)
hook=env['hook']
old='pilotStage==6 && pilotRecruitPhase==10 && pilotViewEntered && (field || wait || inline || choice)'
assert hook.count(old)==1
hook=hook.replace(old,'pilotStage==6 && pilotRecruitPhase==13 && pilotViewEntered && field')
old='            naming:='
# 原始格式沒有命名標籤，插入點保持單一原生等待判斷。
marker='            naming := '
assert marker not in hook
marker='            naming:='
assert marker in hook
hook=hook.replace(marker,'            viewWait:=pc==0x21103 && pilotViewEntered && pilotRecruitPhase==11 && m.CPU.Seg[cpu.DS]==ds && flag==0\n'+marker)
marker='            if naming {phase="name"}'
assert hook.count(marker)==1
hook=hook.replace(marker,marker+'\n            if viewWait {phase="view_detail_wait"}')
marker='pc==0x2113a || pc==0x21155 || (pc==0x210ca && flag!=0)'
assert hook.count(marker)==1
hook=hook.replace(marker,'pc==0x2113a || pc==0x21155 || (pc==0x2110b && flag!=0) || (pc==0x210ca && flag!=0)')
marker='                if next!=0 {'
assert hook.count(marker)==1
hook=hook.replace(marker,r'''
                if pilotStage==6 && pilotViewEntered {
                    if choice && pilotRecruitPhase==10 && m.Read16(cpu.Addr(ds,0x071e))==1 {next=0x1c;kind="recruit_view_pick";pilotRecruitPhase=11}
                    if viewWait && pilotRecruitPhase==11 {next=0x01;kind="recruit_view_detail_close";pilotRecruitPhase=12}
                    if choice && pilotRecruitPhase==12 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x4d;kind="recruit_view_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="recruit_view_decline";pilotRecruitPhase=13}
                    }
                }
'''+marker)
marker='            if rawFlag&0x4000!=0'
assert hook.count(marker)==1
observer=r'''
            if pilotViewEntered && (pc==0x10668 || pc==0x10671 || pc==0x1834e || pc==0x1068e) {
                if m.CPU.Seg[cpu.DS]!=ds {panic("詳細狀況DS不符")}
                fmt.Printf("DQ3_VIEW_DETAIL step=%d packet=%d ida_linear=%05x DS=%04x BX=%04x SI=%04x DI=%04x choice_cursor=%d raw4f1d=%04x actor520b=",m.Steps,pilotPackets,pc,ds,m.CPU.R[cpu.BX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.Read16(cpu.Addr(ds,0x0722)),m.Read16(cpu.Addr(ds,0x4f1d)))
                for i:=uint16(0);i<97;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x520b+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
hook=hook.replace(marker,observer+marker)
env['env']['rebuilt']['GATE_HOOK']=hook
env['env']['rebuilt']['main']()
