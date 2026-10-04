"""正常僧侶詳細頁第二次等待按 A 返回；原始證據及用法見 docs/188。"""
import argparse
from dosgolem_character_spells_probe import build_contract


def build(prefix):
    ns = build_contract(prefix)
    hook = ns['GATE_HOOK']
    old = 'next=0x01;kind="recruit_view_detail_close";pilotRecruitPhase=12'
    assert hook.count(old) == 1
    hook = hook.replace(old, 'next=0x1e;kind="recruit_view_detail_close";pilotRecruitPhase=12')
    marker = '            if rawFlag&0x4000!=0'
    assert hook.count(marker) == 1
    observer = r'''
            if pilotViewEntered && pilotRecruitPhase==12 && (pc==0x10686 || pc==0x10689 || pc==0x1068e) {
                fmt.Printf("DQ3_VIEW_CLOSE_KEY step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x\n",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.AX])
            }
'''
    assert 'm.Write' not in observer and 'SetNextKey' not in observer
    ns['GATE_HOOK'] = hook.replace(marker, observer + marker)
    ns['__file__'] = __file__
    return ns


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prefix', required=True)
    args = parser.parse_args()
    if not args.prefix or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.prefix):
        parser.error('prefix只能使用英數、連字號與底線')
    build(args.prefix)['main']()


if __name__ == '__main__':
    main()
