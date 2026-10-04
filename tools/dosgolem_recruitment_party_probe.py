"""正常199包的唯讀角色副本與插名觀察。入口 docs/188、Issue #4。"""
from pathlib import Path

code = Path('/repo/tools/dosgolem_recruitment_text_probe.py').read_text()
code = code.replace('issue4-recruit-text-r1', 'issue4-recruit-party-r2')
assert code.rstrip().endswith("env['rebuilt']['main']()")
code = code[:code.rindex("env['rebuilt']['main']()")]
ns = {'__file__':__file__, '__name__':'join_party_observer'}
exec(compile(code, __file__, 'exec'), ns)
env=ns['env']
hook=ns['hook']
old='if pilotPackets==198 && inline && pilotRecord==536 {'
assert hook.count(old)==1
hook=hook.replace(old,'if pilotPackets==199 && audioWait && pilotRecord==538 {')
hook=hook.replace('scope=record536_first_inline_wait','scope=record538_audio_wait_entry')
marker='            phase:="none"'
assert hook.count(marker)==1
observe=r'''
            if pilotStage==6 && pilotPackets>=198 && pilotPackets<=199 {
                if pc==0x10415 || pc==0x104b5 || pc==0x104b9 || pc==0x104be || pc==0x104c3 || pc==0x10422 || pc==0x1043f || pc==0x10454 ||
                    ((pc==0x21651 || pc==0x21697 || pc==0x2169a || pc==0x215ee) && pilotRecord>=536 && pilotRecord<=538) {
                    rd:=func(off uint16,n uint16) string {
                        b:=make([]byte,n)
                        for i:=uint16(0);i<n;i++ {b[i]=m.Read8(cpu.Addr(ds,off+i))}
                        return fmt.Sprintf("%x",b)
                    }
                    fmt.Printf("DQ3_JOIN_OBSERVE step=%d packet=%d ida_linear=%05x DS=%04x actual_DS=%04x actual_ES=%04x AX=%04x BX=%04x CX=%04x DX=%04x SI=%04x DI=%04x record=%d raw5077=%d raw259c=%04x raw4f62=%04x slot1=%s scratch=%s",
                        m.Steps,pilotPackets,pc,ds,m.CPU.Seg[cpu.DS],m.CPU.Seg[cpu.ES],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.CX],m.CPU.R[cpu.DX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],pilotRecord,
                        m.Read8(cpu.Addr(ds,0x5077)),m.Read16(cpu.Addr(ds,0x259c)),m.Read16(cpu.Addr(ds,0x4f62)),rd(0x535e,97),rd(0x520b,97))
                    for i:=uint16(0);i<4;i++ {
                        ptr:=m.Read16(cpu.Addr(ds,0x4f15+i*2))
                        fmt.Printf(" ptr%d=%04x party%d=%s",i,ptr,i,rd(ptr,97))
                    }
                    fmt.Printf(" si_words=%s\n",rd(m.CPU.R[cpu.SI],20))
                }
            }
            audioWait:=pc==0x208e2 && pilotStage==6 && pilotPackets==199 && pilotRecord==538
'''
hook=hook.replace(marker,observe+marker)
marker='            if field {phase="ready"}'
assert hook.count(marker)==1
hook=hook.replace(marker,'            if audioWait {phase="audio_wait"}\n'+marker)
env['rebuilt']['GATE_HOOK']=hook
env['rebuilt']['main']()
