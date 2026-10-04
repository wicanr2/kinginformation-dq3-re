"""正常 K 改名的相同姓名、不同姓名、取消來源；入口與有限 READY 見 docs/188。"""
import argparse
from pathlib import Path

SCANS={
    2:[0x48,0x4b,0x1c,0x1c,0x4d,0x50,0x1c,0x48,0x4b,0x1c,0x48,0x1c],
    4:[0x48,0x4b,0x1c,0x50,0x50,0x50,0x1c],
    5:[0x48,0x4b,0x1c,0x1c,0x4d,0x50,0x4d,0x1c,0x48,0x4b,0x4b,0x1c,0x48,0x1c],
}

def build_contract(run,prefix):
    assert run in SCANS and prefix and '/' not in prefix
    p=Path('/repo/tools/dosgolem_recruitment_view_detail_probe.py')
    s=p.read_text().replace('issue4-recruit-view-detail-r1',prefix)
    end="env['env']['rebuilt']['main']()"
    assert s.rstrip().endswith(end)
    e={'__name__':'rename_builder','__file__':__file__}
    exec(compile(s[:s.rindex(end)],str(p),'exec'),e)
    h=e['hook']
    old='pilotStage==6 && pilotRecruitPhase==13 && pilotViewEntered && field'
    assert h.count(old)==1
    h=h.replace(old,'pilotStage==6 && pilotRecruitPhase==16 && pilotViewEntered && field')
    old='if viewWait && pilotRecruitPhase==11 {next=0x01;kind="recruit_view_detail_close";pilotRecruitPhase=12}'
    assert h.count(old)==1
    h=h.replace(old,'if viewWait && pilotRecruitPhase==11 {next=0x25;kind="recruit_view_rename";pilotRecruitPhase=14;pilotBirthNames=0}')
    if run!=2:
        old='next=0x25;kind="recruit_view_rename";pilotRecruitPhase=14;pilotBirthNames=0'
        scans='pilotBirthScans=[]uint8{'+','.join(f'0x{v:02x}' for v in SCANS[run])+'}'
        assert h.count(old)==1
        h=h.replace(old,old+';'+scans)
    marker='            if rawFlag&0x4000!=0'
    assert h.count(marker)==1
    observer=r'''
            if pilotViewEntered && (pc==0x1069f || pc==0x106ab || pc==0x106be || pc==0x106c2 || pc==0x106cb || pc==0x106d8 || pc==0x106da || pc==0x1068e) {
                if m.CPU.Seg[cpu.DS]!=ds {panic("改名DS不符")}
                fmt.Printf("DQ3_RENAME_OBSERVE step=%d packet=%d ida_linear=%05x DS=%04x BX=%04x SI=%04x DI=%04x party_count=%d cursor=%d cancel=%d pointers=",m.Steps,pilotPackets,pc,ds,m.CPU.R[cpu.BX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.Read8(cpu.Addr(ds,0x5077)),m.Read16(cpu.Addr(ds,0x0722)),m.Read8(cpu.Addr(ds,0x0726)))
                for i:=uint16(0);i<5;i++ {fmt.Printf("%04x",m.Read16(cpu.Addr(ds,0x4f15+2*i)))}
                fmt.Printf(" actor520b=")
                for i:=uint16(0);i<97;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x520b+i)))}
                fmt.Printf(" slot1=")
                for i:=uint16(0);i<97;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x535e+i)))}
                fmt.Printf(" hero=")
                ptr:=m.Read16(cpu.Addr(ds,0x4f15))
                for i:=uint16(0);i<97;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,ptr+i)))}
                fmt.Printf("\n")
            }
'''
    assert 'm.Write' not in observer
    h=h.replace(marker,observer+marker)
    marker='                if next!=0 {'
    assert h.count(marker)==1
    h=h.replace(marker,r'''
                if pilotStage==6 && pilotViewEntered {
                    if pilotRecruitPhase==14 && (naming || (choice && m.Read16(cpu.Addr(ds,0x071e))==5)) {
                        if pilotBirthNames>=len(pilotBirthScans) {panic("改名正常輸入用盡")}
                        next=pilotBirthScans[pilotBirthNames];pilotBirthNames++;kind="recruit_rename_name"
                        if pilotBirthNames==len(pilotBirthScans) {pilotRecruitPhase=15}
                    }
                    if choice && pilotRecruitPhase==15 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x4d;kind="recruit_rename_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="recruit_rename_decline";pilotRecruitPhase=16}
                    }
                }
'''+marker)
    if run==5:
        h=h.replace(marker,'                if pilotStage==6 && pilotRecruitPhase==15 && naming {panic("姓名鍵序用盡，尚未返回caller；保留DRAFT來源")}'+'\n'+marker)
    wrapper=e['env']['env']['rebuilt']
    wrapper['GATE_HOOK']=h
    return wrapper

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--route',type=int,choices=SCANS,required=True)
    p.add_argument('--prefix',required=True,help='新來源前綴；拒絕覆寫既有證據')
    args=p.parse_args()
    build_contract(args.route,args.prefix)['main']()

if __name__=='__main__':main()
