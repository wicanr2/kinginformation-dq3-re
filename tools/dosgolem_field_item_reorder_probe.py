"""正常將穿戴物給自己，觀察八格順序的 DRAFT；入口 docs/188。"""
import argparse
from dosgolem_field_item_navigation_probe import build as build_parent


def build(prefix):
    ns = build_parent(prefix)
    hook = ns['GATE_HOOK']
    old = 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==218'
    assert hook.count(old) == 1
    hook = hook.replace(old, 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==230')
    # This route has exactly 230 planned packets. The inherited 220-packet
    # input guard would reject packet221 before reaching the observed writer.
    old = 'pilotPackets>=220 || pilotUps>80'
    assert hook.count(old) == 1
    hook = hook.replace(old, 'pilotPackets>=230 || pilotUps>80')
    marker = '                if next!=0 {'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=218 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    if field {
                        switch pilotPackets {
                        case 218:next=0x39;kind="reorder_open_command"
                        case 226:next=0x39;kind="reorder_reopen_command"
                        }
                    }
                    if choice && count==6 {
                        switch pilotPackets {
                        case 219,227:next=0x50;kind="reorder_command_down"
                        case 220,228:next=0x4d;kind="reorder_command_right"
                        case 221,229:next=0x39;kind="reorder_open_list"
                        }
                    }
                    if choice && count==7 && pilotPackets==222 {next=0x39;kind="reorder_select_equipped"}
                    if choice && count==3 {
                        switch pilotPackets {
                        case 223:next=0x50;kind="reorder_select_give"
                        case 224:
                            if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x39;kind="reorder_give_self"}
                        }
                    }
                    if (wait || inline) && pilotPackets==225 {next=0x1c;kind="reorder_finish_wait"}
                }
''' + marker)
    marker = '            if rawFlag&0x4000!=0'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, r'''
            if pilotPackets>=222 && (pc==0x13a62 || pc==0x13a87 || pc==0x13a9f || pc==0x18197) {
                fmt.Printf("DQ3_ITEM_REORDER_WRITER step=%d packet=%d ida_linear=%05x owner=%d selected=%d AX=%04x BX=%04x CX=%04x SI=%04x raw2591=%04x slots=",m.Steps,pilotPackets,pc,m.Read16(cpu.Addr(ds,0x062d)),m.Read16(cpu.Addr(ds,0x062f)),m.CPU.R[cpu.AX],m.CPU.R[cpu.BX],m.CPU.R[cpu.CX],m.CPU.R[cpu.SI],m.Read16(cpu.Addr(ds,0x2591)))
                actor:=m.Read16(cpu.Addr(ds,0x4f15))
                for i:=uint16(0);i<16;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,actor+0x3a+i)))}
                fmt.Printf("\n")
            }
''' + marker)
    ns['GATE_HOOK'] = hook
    ns['__file__'] = __file__
    return ns


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prefix', required=True)
    args = parser.parse_args()
    if not args.prefix or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.prefix):
        parser.error('prefix只能使用英數、連字號與底線')
    build(args.prefix)['main']()
