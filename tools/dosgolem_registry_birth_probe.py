"""DRAFT正常登錄出生來源；只在原生poll排入IRQ1，入口docs/188。"""
from pathlib import Path

code=Path('/repo/tools/dosgolem_registry_quiescent_probe.py').read_text()
assert code.count("prefix='issue4-registry-quiescent-r2'")==1
code=code.replace("prefix='issue4-registry-quiescent-r2'","prefix='issue4-registry-birth-normal-r2'")
code=code.replace('issue4-registry-quiescent-r2-plan.json','issue4-registry-birth-normal-r2-plan.json')
code=code.replace('registry_quiescent_cold_draft','registry_birth_cold_draft')
code=code.replace('DQ3_QUIESCENT_DONE','DQ3_BIRTH_DONE')
assert code.rstrip().endswith("ns['main']()")
code=code[:code.rindex("ns['main']()")]
env={'__file__':__file__,'__name__':'birth_builder'}
exec(compile(code,'isolated-birth-builder','exec'),env)
hook=env['hook']
start=hook.index('                if pilotStage==4 {')
end=hook.index('                if next!=0 {',start)
birth=r'''
                if pilotStage==4 {
                    if wait || inline {next=0x1c;kind="registry_wait"}
                    if choice && pilotBirthPhase==0 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x1c;pilotBirthPhase=1;kind="registry_accept"
                    } else if pilotBirthPhase==1 && (naming || (choice && m.Read16(cpu.Addr(ds,0x071e))==5)) {
                        if pilotBirthNames>=len(pilotBirthScans) {panic("姓名鍵超出DRAFT路線")}
                        next=pilotBirthScans[pilotBirthNames];pilotBirthNames++;kind="registry_birth_name"
                        if pilotBirthNames==len(pilotBirthScans) {pilotBirthPhase=2}
                    } else if choice && pilotBirthPhase==2 && m.Read16(cpu.Addr(ds,0x071e))==6 {
                        next=0x1c;pilotBirthPhase=3;kind="registry_birth_class"
                    } else if choice && pilotBirthPhase==3 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x1c;pilotBirthPhase=4;kind="registry_birth_gender"
                    } else if choice && pilotBirthPhase==4 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x1c;pilotBirthPhase=5;kind="registry_birth_confirm"
                    } else if choice && pilotBirthPhase==5 && pilotRecord==559 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x4d;kind="registry_birth_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;pilotBirthPhase=6;pilotFinishing=true;kind="registry_birth_decline"}
                    }
                }
'''
hook=hook[:start]+birth+hook[end:]
observer=r'''
            if (pc==0x10924 || pc==0x10a9f || pc==0x10816 || pc==0x1081c) && !pilotBirthObserved[pc] {
                pilotBirthObserved[pc]=true
                fmt.Printf("DQ3_BIRTH_WRITER step=%d packet=%d ida_linear=%05x dx=%04x si=%04x candidate=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.DX],m.CPU.R[cpu.SI])
                for i:=uint16(0);i<128;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x520b+i)))}
                fmt.Printf(" roster_flags=")
                for i:=uint16(0);i<12;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x4f64+i)))}
                fmt.Printf("\n")
            }
'''
marker='            if rawFlag&0x4000!=0'
assert hook.count(marker)==1
hook=hook.replace(marker,observer+'\n'+marker)
hook=hook.replace('DQ3_QUIESCENT_DONE','DQ3_BIRTH_DONE')
extra='''
    var pilotBirthPhase, pilotBirthNames int
    var pilotBirthObserved=map[uint32]bool{}
    var pilotBirthScans=[]uint8{0x48,0x4b,0x1c,0x1c,0x4d,0x50,0x1c,0x48,0x4b,0x1c,0x48,0x1c}
    _ = pilotAccepted
    _ = pilotNames
'''
# Rebuild the same reviewed wrapper with the extra observer declarations before execution.
marker='    var pilotRegistryScans=[]uint8{__REGISTRY__}'
assert code.count(marker)==1
code2=code.replace(marker,marker+extra)
assert code2!=code
env2={'__file__':__file__,'__name__':'birth_builder_final'}
exec(compile(code2,'isolated-birth-builder-final','exec'),env2)
env2['ns']['GATE_HOOK']=hook
env2['ns']['main']()
