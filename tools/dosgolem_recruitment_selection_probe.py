"""正常冷啟動至選人清單、Esc、繼續詢問No、告別及場景返回。入口 docs/188、Issue #4。"""
from pathlib import Path

code = Path('/repo/tools/dosgolem_registry_birth_probe.py').read_text()
code = code.replace('issue4-registry-birth-normal-r2', 'issue4-recruit-selection-cancel-r1')
assert code.rstrip().endswith("env2['ns']['main']()")
code = code[:code.rindex("env2['ns']['main']()")]
extra = '''
    var pilotRecruitMotion, pilotRecruitPhase int
    var pilotRecruitScans=[]uint8{0x4d,0x50,0x4d,0x4d,0x48,0x48,0x48,0x48,0x4d,0x4d,0x4d,0x50,0x50,0x50,0x50,0x4b,0x4b,0x4b,0x4b,0x4b,0x4b,0x48,0x1c}
'''
marker = '    var pilotBirthPhase, pilotBirthNames int'
assert code.count(marker) == 1
code = code.replace(marker, marker + extra)
env = {'__file__': __file__, '__name__': 'recruit_builder'}
exec(compile(code, 'isolated-recruit-builder', 'exec'), env)
hook = env['hook']
marker = '                if pilotFinishing && field {'
assert hook.count(marker) == 1
hook = hook.replace(marker, '''
                if pilotPackets>=172 {
                    fmt.Printf("DQ3_RECRUIT_STATE step=%d packet=%d phase=%s roster_flags=",m.Steps,pilotPackets,phase)
                    for i:=uint16(0);i<12;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x4f64+i)))}
                    fmt.Printf(" slot1=")
                    for i:=uint16(0);i<97;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x535e+i)))}
                    fmt.Printf(" raw4f1f=%04x raw4f15=%04x raw4f17=%04x raw4f19=%04x raw4f1b=%04x\\n",m.Read16(cpu.Addr(ds,0x4f1f)),m.Read16(cpu.Addr(ds,0x4f15)),m.Read16(cpu.Addr(ds,0x4f17)),m.Read16(cpu.Addr(ds,0x4f19)),m.Read16(cpu.Addr(ds,0x4f1b)))
                }
                if pilotPackets==172 && field {pilotFinishing=false;pilotStage=5}
                if pilotStage==5 && choice && pilotRecord==528 && m.Read16(cpu.Addr(ds,0x071e))==3 {pilotStage=6}
                if pilotStage==6 && pilotRecruitPhase==3 && field {
                    fmt.Printf("DQ3_RECRUIT_DONE step=%d packets=%d irqs=%d player_x=%d player_y=%d raw0b24=%04x record=%d choice_count=%d choice_cursor=%d\\n",m.Steps,pilotPackets,m.KeyIRQs,x,y,m.Read16(cpu.Addr(ds,0x0b24)),pilotRecord,m.Read16(cpu.Addr(ds,0x071e)),m.Read16(cpu.Addr(ds,0x0722)))
                    break
                }
                if pilotStage==99 && pilotFinishing && field {''')
marker = '                if next!=0 {'
assert hook.count(marker) == 1
hook = hook.replace(marker, '''
                if pilotStage==5 {
                    if wait || inline {next=0x1c;kind="recruit_wait"}
                    if field {
                        if pilotRecruitMotion>=len(pilotRecruitScans) {panic("招募路線用盡仍回field；保留DRAFT觀測")}
                        next=pilotRecruitScans[pilotRecruitMotion];pilotRecruitMotion++;kind="recruit_approach"
                    }
                }

                if pilotStage==6 {
                    if wait || inline {next=0x1c;kind="recruit_wait"}
                    if choice && pilotRecruitPhase==0 && m.Read16(cpu.Addr(ds,0x071e))==3 {next=0x1c;kind="recruit_join";pilotRecruitPhase=1}
                    if choice && pilotRecruitPhase==1 && m.Read16(cpu.Addr(ds,0x071e))==1 {next=0x01;kind="recruit_selection_cancel";pilotRecruitPhase=2}
                    if choice && pilotRecruitPhase==2 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x4d;kind="recruit_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="recruit_decline";pilotRecruitPhase=3}
                    }
                }
''' + marker)
ns = env['env2']['ns']
ns['GATE_HOOK'] = hook
# The wrapper requires a completion marker but this new finite stop is a menu.
assert ns['main'].__globals__ is ns
source = Path('/repo/tools/dosgolem_mother_return.py').read_text()
# Modify only the already built wrapper function's expected marker via code text.
# Recompile its source from the composed namespace, retaining its declarations.
base_code = env['env2']['code'] if 'code' in env['env2'] else None
assert base_code is not None
base_code = base_code.replace("'DQ3_QUIESCENT_DONE '", "'DQ3_RECRUIT_DONE '")
base_code = base_code.replace("'DQ3_BIRTH_DONE '", "'DQ3_RECRUIT_DONE '")
# env2 code has already had main removed; rebuild it without running.
rebuilt = {'__file__': __file__, '__name__': 'recruit_wrapper'}
exec(compile(base_code, 'isolated-recruit-wrapper', 'exec'), rebuilt)
rebuilt['GATE_HOOK'] = hook
rebuilt['main']()
