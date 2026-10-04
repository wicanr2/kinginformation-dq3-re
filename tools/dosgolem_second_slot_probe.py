"""正常 F5/F6 第二槽存讀檔與左右行走 DRAFT；入口 docs/188。"""
import argparse
from dosgolem_after_load_move_probe import build as build_first_slot


def build(prefix):
    ns = build_first_slot(prefix)
    hook = ns['GATE_HOOK']
    changes = {
        'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==202 && field':
            'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==204 && field',
        'if pilotPackets==195 && choice && pilotRecord==250 && m.Read16(cpu.Addr(ds,0x071e))==10 {next=0x1c;kind="save_first_slot"}':
            '''if pilotPackets==195 && choice && pilotRecord==250 && m.Read16(cpu.Addr(ds,0x071e))==10 {next=0x50;kind="save_second_slot_cursor"}
                    if pilotPackets==196 && choice && pilotRecord==250 && m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="save_second_slot"}''',
        'if pilotPackets==196 && wait && pilotRecord==252 {next=0x1c;kind="save_close"}':
            'if pilotPackets==197 && wait && pilotRecord==252 {next=0x1c;kind="save_close"}',
        'if pilotPackets==197 && field {next=0x4b;kind="after_save_move"}':
            'if pilotPackets==198 && field {next=0x4b;kind="after_save_move"}',
        'if pilotPackets==198 && field {next=0x40;kind="load_f6"}':
            'if pilotPackets==199 && field {next=0x40;kind="load_f6"}',
        'if pilotPackets==199 && choice && m.Read16(cpu.Addr(ds,0x071e))==10 {next=0x1c;kind="load_first_slot"}':
            '''if pilotPackets==200 && choice && m.Read16(cpu.Addr(ds,0x071e))==10 {next=0x50;kind="load_second_slot_cursor"}
                    if pilotPackets==201 && choice && m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="load_second_slot"}''',
        'if pilotPackets==200 {next=0x4b;kind="after_load_left"}':
            'if pilotPackets==202 {next=0x4b;kind="after_load_left"}',
        'if pilotPackets==201 {next=0x4d;kind="after_load_right"}':
            'if pilotPackets==203 {next=0x4d;kind="after_load_right"}',
    }
    for old, new in changes.items():
        assert hook.count(old) == 1, old
        hook = hook.replace(old, new)
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
