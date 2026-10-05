"""只讀觀察正常第225步消息窗口及字模consumer；證據入口docs/188。"""
import sys
from pathlib import Path

sys.path.insert(0, '/repo/tools')
from dosgolem_field_item_reorder_probe import build

ns = build('issue4-give-prompt-observer-r1')
hook = ns['GATE_HOOK']
marker = '            if rawFlag&0x4000!=0'
assert hook.count(marker) == 1
observer = r'''
            if pilotPackets==225 && (pc==0x15002 || pc==0x1500f || pc==0x21414 || pc==0x214b9 || pc==0x139b6 || pc==0x139c2 || pc==0x2111b || pc==0x21133) {
                once:=pc!=0x214b9
                if !once || !pilotObserveSeen[pc] {
                    if once {pilotObserveSeen[pc]=true}
                    fmt.Printf("DQ3_GIVE_PROMPT_OBSERVE step=%d packet=%d ida_linear=%05x DI=%04x SI=%04x BP=%04x DX=%04x BX=%04x raw0716=%d raw0718=%d raw259b=%d raw259e=%d window3e6e=",m.Steps,pilotPackets,pc,m.CPU.R[cpu.DI],m.CPU.R[cpu.SI],m.CPU.R[cpu.BP],m.CPU.R[cpu.DX],m.CPU.R[cpu.BX],m.Read16(cpu.Addr(ds,0x0716)),m.Read16(cpu.Addr(ds,0x0718)),m.Read8(cpu.Addr(ds,0x259b)),m.Read8(cpu.Addr(ds,0x259e)))
                    for i:=uint16(0);i<30;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3e6e+i)))}
                    fmt.Printf("\n")
                }
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
ns['GATE_HOOK'] = hook.replace(marker, observer + marker)
ns['__file__'] = __file__
ns['main']()
