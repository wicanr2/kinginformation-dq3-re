"""正常 F5→F6 後左移再右移的有界 DRAFT 探針；證據入口 docs/188。"""
import argparse
from dosgolem_save_load_roundtrip_probe import build as build_roundtrip


def build(prefix):
    ns = build_roundtrip(prefix)
    hook = ns['GATE_HOOK']
    stop = 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==200 && field'
    assert hook.count(stop) == 1
    hook = hook.replace(stop, 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==202 && field')
    marker = '                if next!=0 {'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
                if pilotStage==6 && pilotRecruitPhase==13 && field {
                    if pilotPackets==200 {next=0x4b;kind="after_load_left"}
                    if pilotPackets==201 {next=0x4d;kind="after_load_right"}
                }
''' + marker)
    marker = '                pilotPending=false'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
                if pilotPackets>=193 {
                    fmt.Printf("DQ3_AFTER_LOAD_CLOCK step=%d packet=%d clock=%d raw526c=%d\\n",m.Steps,pilotPackets,m.Read16(cpu.Addr(ds,0x251d)),m.Read8(cpu.Addr(ds,0x526c)))
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
