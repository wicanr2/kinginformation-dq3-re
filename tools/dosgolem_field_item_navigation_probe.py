"""正常道具清單、穿戴列選取、取消與返回 DRAFT；入口 docs/188。"""
import argparse
from dosgolem_command_menu_navigation_probe import build as build_parent


def build(prefix):
    ns = build_parent(prefix)
    hook = ns['GATE_HOOK']
    old = 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==206'
    assert hook.count(old) == 1
    hook = hook.replace(old, 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==218')
    marker = '                if next!=0 {'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, r'''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=206 {
                    next=0;kind="none"
                    count:=m.Read16(cpu.Addr(ds,0x071e))
                    if choice && count==7 {
                        switch pilotPackets {
                        case 206:next=0x39;kind="item_equipped_select"
                        case 212:next=0x50;kind="item_down_backpack"
                        case 213:next=0x48;kind="item_up_equipped"
                        case 214:next=0x48;kind="item_up_wrap"
                        case 215:next=0x50;kind="item_down_wrap"
                        case 216:next=0x01;kind="item_list_cancel"
                        }
                    }
                    if choice && count==3 && pilotPackets==207 {next=0x01;kind="item_action_cancel"}
                    if choice && count==6 {
                        switch pilotPackets {
                        case 209:next=0x50;kind="item_reopen_down"
                        case 210:next=0x4d;kind="item_reopen_right"
                        case 211:next=0x39;kind="item_reopen_select"
                        }
                    }
                    if field && pilotPackets==208 {next=0x39;kind="item_reopen_command"}
                    if field && pilotPackets==217 {next=0x4d;kind="item_after_right"}
                }
''' + marker)
    marker = '                pilotPending=false'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, r'''
                if pilotPackets>=206 {
                    fmt.Printf("DQ3_ITEM_RASTER step=%d packet=%d owner062d=%d selected062f=%d raw2591=%04x raw071c=%d raw0710=%04x raw0712=%04x raw0716=%d raw0718=%d raw258f=%d raw0727=%d raw2784=%d window3fd8=",m.Steps,pilotPackets,m.Read16(cpu.Addr(ds,0x062d)),m.Read16(cpu.Addr(ds,0x062f)),m.Read16(cpu.Addr(ds,0x2591)),m.Read16(cpu.Addr(ds,0x071c)),m.Read16(cpu.Addr(ds,0x0710)),m.Read16(cpu.Addr(ds,0x0712)),m.Read16(cpu.Addr(ds,0x0716)),m.Read16(cpu.Addr(ds,0x0718)),m.Read8(cpu.Addr(ds,0x258f)),m.Read8(cpu.Addr(ds,0x0727)),m.Read16(cpu.Addr(ds,0x2784)))
                    for i:=uint16(0);i<96;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3fd8+i)))}
                    fmt.Printf(" action4050=")
                    for i:=uint16(0);i<64;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x4050+i)))}
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
