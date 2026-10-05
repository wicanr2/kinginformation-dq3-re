"""正常401原版調查訊息後Enter返回、左右行走與唯讀consumer觀察；docs/188 DRAFT。"""
from pathlib import Path
import sys
sys.path.insert(0, '/repo/tools')
s = Path('/work/issue4-search-first-probe-r1.py').read_text()
assert s.count("\nns['main']()\n") == 1
s = s.replace("\nns['main']()\n", '\n')
assert s.count('issue4-search-first-normal-r1') == 1
s = s.replace('issue4-search-first-normal-r1', 'issue4-search-return-normal-r1')
scope = {'__file__': __file__, '__name__': 'search_return_builder'}
exec(compile(s, __file__, 'exec'), scope)
ns = scope['ns']; h = ns['GATE_HOOK']
for a, b in [('pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==401', 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==404'), ('pilotPackets>=401 || pilotUps>80', 'pilotPackets>=404 || pilotUps>80')]:
    assert h.count(a) == 1
    h = h.replace(a, b)
marker = '                if next!=0 {'
assert h.count(marker) == 1
inputs = r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=401 {
                    next=0;kind="none"
                    if wait && pilotRecord==265 && pilotPackets==401 {next=0x1c;kind="search_result_close"}
                    if field && pilotPackets==402 {next=0x4b;kind="search_after_left"}
                    if field && pilotPackets==403 {next=0x4d;kind="search_after_right"}
                }
'''
assert 'm.Write' not in inputs and 'Restore' not in inputs
h = h.replace(marker, inputs + marker)
marker = '            if rawFlag&0x4000!=0'
assert h.count(marker) == 1
observer = r'''
            if pilotPackets>=401 && (pc==0x18966 || pc==0x18977 || pc==0x18986 || pc==0x18990 || pc==0x1899f || pc==0x189ab || pc==0x189bc || pc==0x189c4 || pc==0x189cf || pc==0x18c75 || pc==0x18c83 || pc==0x18c93 || pc==0x18c96 || pc==0x18c9b || pc==0x18ca3 || pc==0x18ca8 || pc==0x18cac || pc==0x1f604 || pc==0x15002 || pc==0x21414 || pc==0x2111b) {
                fmt.Printf("DQ3_SEARCH_NATIVE step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x BX=%04x SI=%04x DI=%04x BP=%04x DX=%04x raw01f7=%d raw0b5d=%d raw0b5e=%d raw4f2d=%d raw4f3b=%d raw0b26=%04x raw0b28=%d raw2536=%04x x=%d y=%d window3e6e=",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.CPU.R[cpu.BP],m.CPU.R[cpu.DX],m.Read16(cpu.Addr(ds,0x01f7)),m.Read8(cpu.Addr(ds,0x0b5d)),m.Read16(cpu.Addr(ds,0x0b5e)),m.Read16(cpu.Addr(ds,0x4f2d)),m.Read8(cpu.Addr(ds,0x4f3b)),m.Read16(cpu.Addr(ds,0x0b26)),m.Read16(cpu.Addr(ds,0x0b28)),m.Read16(cpu.Addr(ds,0x2536)),m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)))
                for i:=uint16(0);i<24;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3e6e+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK'] = h.replace(marker, observer + marker)
ns['__file__'] = __file__
ns['main']()
