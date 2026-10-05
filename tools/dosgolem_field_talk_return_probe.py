"""正常406無對象對話後新Enter返回與左右移動；Issue4 DRAFT。"""
from pathlib import Path
s=Path('/work/issue4-talk-empty-first-probe-r2.py').read_text()
assert s.count("\nns['main']()\n")==1
s=s.replace("\nns['main']()\n",'\n').replace('issue4-talk-empty-first-normal-r2','issue4-talk-empty-return-normal-r1')
scope={'__file__':__file__,'__name__':'talk_return_builder'}
exec(compile(s,__file__,'exec'),scope)
ns=scope['ns'];h=ns['GATE_HOOK']
for a,b in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==406','pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==409'),('pilotPackets>=406 || pilotUps>80','pilotPackets>=409 || pilotUps>80')]:
 assert h.count(a)==1;h=h.replace(a,b)
marker='                if next!=0 {'
assert h.count(marker)==1
inputs=r"""
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=406 {
                    next=0;kind="none"
                    if wait && pilotRecord==260 && pilotPackets==406 {next=0x1c;kind="talk_empty_close"}
                    if field && pilotPackets==407 {next=0x4b;kind="talk_after_left"}
                    if field && pilotPackets==408 {next=0x4d;kind="talk_after_right"}
                }
"""
assert 'm.Write' not in inputs and 'Restore' not in inputs
h=h.replace(marker,inputs+marker)
marker='            if rawFlag&0x4000!=0'
assert h.count(marker)==1
observer=r"""
            if pilotPackets>=406 && (pc==0x14e0e || pc==0x14e47 || pc==0x14e4d || pc==0x14e5d || pc==0x14e6b || pc==0x14e7f || pc==0x14e82 || pc==0x14e85 || pc==0x14e8b || pc==0x15023 || pc==0x15002 || pc==0x21414 || pc==0x2111b || pc==0x1f604) {
                fmt.Printf("DQ3_TALK_NATIVE step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x BX=%04x SI=%04x DI=%04x BP=%04x DX=%04x raw01f7=%d raw4f2d=%d raw4f3b=%d raw0b26=%04x raw0b28=%d raw2536=%04x x=%d y=%d window3e6e=",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.BP],m.CPU.R[cpu.DX],m.Read16(cpu.Addr(ds,0x01f7)),m.Read16(cpu.Addr(ds,0x4f2d)),m.Read8(cpu.Addr(ds,0x4f3b)),m.Read16(cpu.Addr(ds,0x0b26)),m.Read16(cpu.Addr(ds,0x0b28)),m.Read16(cpu.Addr(ds,0x2536)),m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)))
                for i:=uint16(0);i<24;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3e6e+i)))}
                fmt.Printf("\n")
            }
"""
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK']=h.replace(marker,observer+marker);ns['__file__']=__file__;ns['main']()
