"""正常登錄僧侶，觀看能力與咒文頁後返回；用法及證據見docs/188。"""
from pathlib import Path
import argparse


def build_contract(prefix, first_ability=False):
    code=Path('/repo/tools/dosgolem_recruitment_view_detail_probe.py').read_text()
    code=code.replace('issue4-recruit-view-detail-r1',prefix)
    suffix="env['env']['rebuilt']['main']()"
    assert code.rstrip().endswith(suffix)
    code=code[:code.rindex(suffix)]
    ns={'__file__':__file__,'__name__':'character_spells_builder'}
    exec(compile(code,__file__,'exec'),ns)
    hook=ns['hook']
    old='''                    } else if choice && pilotBirthPhase==2 && m.Read16(cpu.Addr(ds,0x071e))==6 {
                        next=0x1c;pilotBirthPhase=3;kind="registry_birth_class"'''
    assert hook.count(old)==1
    hook=hook.replace(old,'''                    } else if choice && pilotBirthPhase==2 && m.Read16(cpu.Addr(ds,0x071e))==6 {
                        next=0x50;kind="registry_birth_class_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==3 {next=0x1c;pilotBirthPhase=3;kind="registry_birth_class"}''')
    old='if pilotPackets==172 && field {pilotFinishing=false;pilotStage=5}'
    assert hook.count(old)==1
    hook=hook.replace(old,'if pilotStage==4 && pilotFinishing && field {pilotFinishing=false;pilotStage=5}')
    if first_ability:
        old='pilotStage==6 && pilotRecruitPhase==13 && pilotViewEntered && field'
        assert hook.count(old)==1
        hook=hook.replace(old,'pilotStage==6 && pilotRecruitPhase==11 && pilotViewEntered && wait')
    marker='            if rawFlag&0x4000!=0'
    assert hook.count(marker)==1
    observer=r'''
            if pilotViewEntered && (pc==0x10671 || pc==0x1834e || pc==0x184a1 || pc==0x184c0 || pc==0x18530) {
                if m.CPU.Seg[cpu.DS]!=ds {panic("咒文觀察DS不符")}
                fmt.Printf("DQ3_VIEW_SPELLS step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x BX=%04x SI=%04x DI=%04x choice_cursor=%d actor520b=",m.Steps,pilotPackets,pc,ds,m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.Read16(cpu.Addr(ds,0x0722)))
                for i:=uint16(0);i<128;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x520b+i)))}
                fmt.Printf("\n")
            }
'''
    assert 'm.Write' not in observer and 'SetNextKey' not in observer
    hook=hook.replace(marker,observer+marker)
    if not first_ability:
        observer=r'''
            if pilotViewEntered && (pc==0x184c0 || pc==0x18567 || pc==0x18573 || pc==0x185c1 || pc==0x185d7 || pc==0x185dc || pc==0x10681 || pc==0x213c4) {
                if m.CPU.Seg[cpu.DS]!=ds {panic("咒文窗口觀察DS不符")}
                fmt.Printf("DQ3_VIEW_SPELL_WINDOW step=%d packet=%d ida_linear=%05x AX=%04x BX=%04x CX=%04x DX=%04x SI=%04x DI=%04x BP=%04x text_x=%d text_y=%d window3f34=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.CX],m.CPU.R[cpu.DX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.BP],m.Read16(cpu.Addr(ds,0x0716)),m.Read16(cpu.Addr(ds,0x0718)))
                for i:=uint16(0);i<24;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3f34+i)))}
                fmt.Printf(" union232d=")
                for i:=uint16(0);i<60;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x232d+i)))}
                fmt.Printf("\n")
            }
'''
        assert 'm.Write' not in observer and 'SetNextKey' not in observer
        hook=hook.replace(marker,observer+marker)
    rebuilt=ns['env']['env']['rebuilt']
    rebuilt['GATE_HOOK']=hook
    return rebuilt


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prefix',required=True,help='全新輸出前綴；既有產物拒絕覆寫')
    parser.add_argument('--first-ability',action='store_true',help='停在觀看第一能力等待；預設續行209包至正常返回')
    args=parser.parse_args()
    if not args.prefix or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.prefix):
        parser.error('prefix只能使用英數、連字號與底線')
    build_contract(args.prefix,args.first_ability)['main']()


if __name__=='__main__':
    main()
