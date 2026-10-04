"""原版正常 F5 存檔、移動、F6 讀檔 DRAFT；入口 docs/188。"""
import argparse
from dosgolem_save_load_continue_probe import build as build_save


def build(prefix):
    ns = build_save(prefix, 'save')
    hook = ns['GATE_HOOK']
    old = 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==196 && field'
    assert hook.count(old) == 1
    hook = hook.replace(old, 'pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==200 && field')
    marker = '                if next!=0 {'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
                if pilotStage==6 && pilotRecruitPhase==13 {
                    if pilotPackets==196 && wait && pilotRecord==252 {next=0x1c;kind="save_close"}
                    if pilotPackets==197 && field {next=0x4b;kind="after_save_move"}
                    if pilotPackets==198 && field {next=0x40;kind="load_f6"}
                    if pilotPackets==199 && choice && m.Read16(cpu.Addr(ds,0x071e))==10 {next=0x1c;kind="load_first_slot"}
                }
''' + marker)
    # 只觀察 save/load caller 與完整原生持久區，不修改任何遊戲欄位。
    marker = '            if rawFlag&0x4000!=0'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
            if pilotPackets>=194 && (pc==0x11484 || pc==0x114c8 || pc==0x114d3 || pc==0x114d9 || pc==0x1157d || pc==0x1158b || pc==0x11591 || pc==0x1165f) {
                fmt.Printf("DQ3_SAVE_LOAD_NATIVE step=%d packet=%d ida_linear=%05x DS=%04x AX=%04x CX=%04x DX=%04x raw0722=%d raw0726=%d raw01f0=%d\\n",m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.CPU.R[cpu.AX],m.CPU.R[cpu.CX],m.CPU.R[cpu.DX],m.Read16(cpu.Addr(ds,0x0722)),m.Read8(cpu.Addr(ds,0x0726)),m.Read8(cpu.Addr(ds,0x01f0)))
            }
''' + marker)
    marker = '                pilotPending=false'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
                if pilotPackets>=193 {
                    data:=make([]byte,0x57a5-0x4f29)
                    for i:=range data {data[i]=m.Read8(cpu.Addr(ds,uint16(0x4f29+i)))}
                    if err:=os.WriteFile(fmt.Sprintf("/work/dosgolem-opening/__PREFIX__-persistent-%03d.bin",pilotPackets),data,0o644);err!=nil {die(err)}
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
