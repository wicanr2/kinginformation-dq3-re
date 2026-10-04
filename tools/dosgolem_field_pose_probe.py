"""正常 F5/F6 後的人物影格唯讀觀測；DRAFT 證據入口 docs/188。"""
import argparse

from dosgolem_after_load_move_probe import build as build_after_load


def build(prefix):
    ns = build_after_load(prefix)
    hook = ns['GATE_HOOK']
    marker = '            if rawFlag&0x4000!=0'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
            // Read-only observation at the original sprite-table consumer.
            // BX is the original table byte offset; no phase/time/state writes.
            if pilotPackets>=193 && pc==0x1e307 {
                fmt.Printf("DQ3_FIELD_POSE_BLIT step=%d packet=%d ida_linear=%05x DS=%04x BX=%04x SI=%04x DI=%04x raw0004=%d raw26f0=%d raw26ad=%d raw26b0=%d\\n",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.BX],m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.Read8(cpu.Addr(ds,0x0004)),m.Read16(cpu.Addr(ds,0x26f0)),m.Read8(cpu.Addr(ds,0x26ad)),m.Read16(cpu.Addr(ds,0x26b0)))
            }
''' + marker)
    marker = '                pilotPending=false'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
                if pilotPackets>=193 {
                    fmt.Printf("DQ3_FIELD_POSE_BOUNDARY step=%d packet=%d raw0004=%d raw26ad=%d raw26b0=%d raw26d0=%d raw26d1=%d\\n",m.Steps,pilotPackets,m.Read8(cpu.Addr(ds,0x0004)),m.Read8(cpu.Addr(ds,0x26ad)),m.Read16(cpu.Addr(ds,0x26b0)),m.Read8(cpu.Addr(ds,0x26d0)),m.Read8(cpu.Addr(ds,0x26d1)))
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
