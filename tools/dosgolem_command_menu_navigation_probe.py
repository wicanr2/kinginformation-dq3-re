"""正常命令窗導覽、Esc及道具入口 DRAFT；入口 docs/188。"""
import argparse
from dosgolem_command_menu_probe import build as build_open


def build(prefix):
    ns = build_open(prefix)
    hook = ns['GATE_HOOK']
    old = 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==194'
    assert hook.count(old) == 1
    hook = hook.replace(old, 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==206')
    marker = '                if next!=0 {'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=194 {
                    next=0;kind="none"
                    if choice && m.Read16(cpu.Addr(ds,0x071e))==6 {
                        switch pilotPackets {
                        case 194:next=0x50;kind="command_down_first"
                        case 195:next=0x50;kind="command_down_second"
                        case 196:next=0x50;kind="command_down_boundary"
                        case 197:next=0x48;kind="command_up_boundary"
                        case 198:next=0x4b;kind="command_left"
                        case 199:next=0x4d;kind="command_right"
                        case 200:next=0x4d;kind="command_right_boundary"
                        case 201:next=0x01;kind="command_escape"
                        case 203:next=0x50;kind="command_item_down"
                        case 204:next=0x4d;kind="command_item_right"
                        case 205:
                            if m.Read16(cpu.Addr(ds,0x0722))==5 {next=0x39;kind="command_item_space"}
                        }
                    }
                    if pilotPackets==202 && field {next=0x39;kind="command_reopen"}
                }
''' + marker)
    marker = '                pilotPending=false'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
                if pilotPackets>=194 {
                    fmt.Printf("DQ3_COMMAND_RASTER step=%d packet=%d raw258f=%d raw0727=%d raw2784=%d raw071c=%d raw5077=%d window3d6c=",m.Steps,pilotPackets,m.Read8(cpu.Addr(ds,0x258f)),m.Read8(cpu.Addr(ds,0x0727)),m.Read16(cpu.Addr(ds,0x2784)),m.Read16(cpu.Addr(ds,0x071c)),m.Read8(cpu.Addr(ds,0x5077)))
                    for i:=uint16(0);i<64;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3d6c+i)))}
                    fmt.Printf(" hud3e9c=")
                    for i:=uint16(0);i<24;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x3e9c+i)))}
                    fmt.Printf("\\n")
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
