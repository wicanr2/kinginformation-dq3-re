"""正常姓名取消後下樓觀看空名冊；完整來源及有限規格入口 docs/188。"""
import argparse
from pathlib import Path


def build(prefix):
    code=Path('/repo/tools/dosgolem_registry_quiescent_probe.py').read_text()
    code=code.replace('issue4-registry-quiescent-r2',prefix)
    assert code.rstrip().endswith("ns['main']()")
    code=code[:code.rindex("ns['main']()")]
    declaration='    var pilotRegistryScans=[]uint8{__REGISTRY__}'
    assert code.count(declaration)==1
    code=code.replace(declaration,declaration+'''
    var pilotRecruitMotion, pilotRecruitPhase int
    var pilotRecruitScans=[]uint8{0x4d,0x50,0x4d,0x4d,0x48,0x48,0x48,0x48,0x4d,0x4d,0x4d,0x50,0x50,0x50,0x50,0x4b,0x4b,0x4b,0x4b,0x4b,0x4b,0x48,0x1c}
''')
    ns={'__file__':__file__,'__name__':'empty_view_builder'}
    exec(compile(code,__file__,'exec'),ns)
    hook=ns['hook']
    old='                if pilotFinishing && field {'
    assert hook.count(old)==1
    hook=hook.replace(old,'''                if pilotFinishing && field && pilotStage==4 {pilotFinishing=false;pilotStage=5}
                if pilotStage==6 && pilotRecruitPhase==13 && field {''')
    marker='                if next!=0 {'
    assert hook.count(marker)==1
    hook=hook.replace(marker,r'''
                if pilotStage==5 {
                    if wait || inline {next=0x1c;kind="recruit_wait"}
                    if field {
                        if pilotRecruitMotion>=len(pilotRecruitScans) {panic("正常空名冊路線用盡")}
                        next=pilotRecruitScans[pilotRecruitMotion];pilotRecruitMotion++;kind="recruit_approach"
                    }
                    if choice && pilotRecord==528 && m.Read16(cpu.Addr(ds,0x071e))==3 {pilotStage=6}
                }
                if pilotStage==6 {
                    if wait || inline {next=0x1c;kind="empty_view_wait"}
                    if choice && pilotRecruitPhase==0 && pilotRecord==528 && m.Read16(cpu.Addr(ds,0x071e))==3 {
                        next=0x50;kind="empty_view_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==3 {next=0x1c;kind="empty_view_open";pilotRecruitPhase=10}
                    }
                    if choice && pilotRecruitPhase==10 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x4d;kind="empty_view_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="empty_view_decline";pilotRecruitPhase=13}
                    }
                }
'''+marker)
    marker='            if rawFlag&0x4000!=0'
    assert hook.count(marker)==1
    observer=r'''
            if pilotStage>=5 && (pc==0x10624 || pc==0x10627 || pc==0x1062f || pc==0x10696 || pc==0x10699 || pc==0x1069e) {
                fmt.Printf("DQ3_EMPTY_VIEW_ENTRY step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x DI=%04x raw5060=%d roster_flags=",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.AX],m.CPU.R[cpu.DI],m.Read16(cpu.Addr(ds,0x5060)))
                for i:=uint16(0);i<12;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x4f64+i)))}
                fmt.Printf("\n")
            }
            if pilotStage>=5 && pc==0x21414 {
                fmt.Printf("DQ3_EMPTY_VIEW_TEXT step=%d packet=%d DI=%04x text_segment=%04x\n",m.Steps,pilotPackets,m.CPU.R[cpu.DI],m.Read16(cpu.Addr(ds,0x252e)))
            }
'''
    assert 'm.Write' not in observer and 'SetNextKey' not in observer
    ns['ns']['GATE_HOOK']=hook.replace(marker,observer+marker)
    return ns['ns']


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prefix',required=True)
    args=parser.parse_args()
    if not args.prefix or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.prefix):
        parser.error('prefix只能使用英數、連字號與底線')
    build(args.prefix)['main']()


if __name__=='__main__':
    main()
