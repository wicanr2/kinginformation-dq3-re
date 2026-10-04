"""正常 F5 確認後的拒絕或存檔分支 DRAFT；入口 docs/188。"""
import argparse
from dosgolem_save_load_probe import build as build_entry


def build(prefix, action):
    assert action in ('decline', 'save')
    ns = build_entry(prefix)
    hook = ns['GATE_HOOK']
    old = 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==194'
    assert hook.count(old) == 1
    if action == 'decline':
        hook = hook.replace(old, 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==197 && field')
        extra = '''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=194 {
                    next=0;kind="none"
                    if pilotPackets==194 && choice && pilotRecord==253 && m.Read16(cpu.Addr(ds,0x071e))==2 {next=0x4d;kind="save_decline_cursor"}
                    if pilotPackets==195 && choice && pilotRecord==253 && m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="save_decline"}
                    if pilotPackets==196 && wait && pilotRecord==252 {next=0x1c;kind="save_decline_close"}
                }
'''
    else:
        hook = hook.replace(old, 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==196 && field')
        extra = '''
                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets>=194 {
                    next=0;kind="none"
                    if pilotPackets==194 && choice && pilotRecord==253 && m.Read16(cpu.Addr(ds,0x071e))==2 {next=0x1c;kind="save_accept"}
                    if pilotPackets==195 && choice && pilotRecord==250 && m.Read16(cpu.Addr(ds,0x071e))==10 {next=0x1c;kind="save_first_slot"}
                }
'''
        marker = '        if m.Steps>=2800000000 {'
        assert hook.count(marker) == 1
        hook = hook.replace(marker, '''        if m.Steps>=2800000000 || (pilotPackets==196 && m.Steps>pilotQueued+50000000) {''')
    marker = '                if next!=0 {'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, extra + marker)
    ns['GATE_HOOK'] = hook
    ns['__file__'] = __file__
    return ns


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prefix', required=True)
    parser.add_argument('--action', choices=('decline', 'save'), required=True)
    args = parser.parse_args()
    if not args.prefix or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.prefix):
        parser.error('prefix只能使用英數、連字號與底線')
    build(args.prefix, args.action)['main']()
