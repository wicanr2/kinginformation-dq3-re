"""正常 Space 命令窗 DRAFT 原版觀察；入口 docs/188。"""
import argparse
from dosgolem_save_load_probe import build as build_entry


def build(prefix):
    ns = build_entry(prefix)
    hook = ns['GATE_HOOK']
    old = 'next=0x3f;kind="save_f5"'
    assert hook.count(old) == 1
    hook = hook.replace(old, 'next=0x39;kind="command_space_open"')
    marker = '                pilotPending=false'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''
                if pilotPackets>=193 {
                    data:=make([]byte,0x57a5-0x4f29)
                    for i:=range data {data[i]=m.Read8(cpu.Addr(ds,uint16(0x4f29+i)))}
                    if err:=os.WriteFile(fmt.Sprintf("/work/dosgolem-opening/__PREFIX__-persistent-%03d.bin",pilotPackets),data,0o644);err!=nil {die(err)}
                    fmt.Printf("DQ3_COMMAND_CLOCK step=%d packet=%d clock=%d raw526c=%d raw0726=%d\\n",m.Steps,pilotPackets,m.Read16(cpu.Addr(ds,0x251d)),m.Read8(cpu.Addr(ds,0x526c)),m.Read8(cpu.Addr(ds,0x0726)))
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
