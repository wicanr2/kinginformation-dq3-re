"""正常 F5 確認 Esc 與選槽 Esc 的 DRAFT 來源；入口 docs/188。"""
import argparse
from dosgolem_save_load_continue_probe import build as build_decline


def build(prefix):
    ns = build_decline(prefix, 'decline')
    hook = ns['GATE_HOOK']
    old = 'next=0x4d;kind="save_decline_cursor"'
    assert hook.count(old) == 1
    hook = hook.replace(old, 'next=0x01;kind="save_confirm_escape"')
    old = 'if pilotPackets==195 && choice && pilotRecord==253 && m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="save_decline"}'
    assert hook.count(old) == 1
    hook = hook.replace(old, 'if pilotPackets==195 && choice && pilotRecord==250 && m.Read16(cpu.Addr(ds,0x071e))==10 {next=0x01;kind="save_slot_escape"}')
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
