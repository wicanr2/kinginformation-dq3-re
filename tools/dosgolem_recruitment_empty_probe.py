"""正常姓名取消後下樓空Join／Leave；有界DRAFT及工具入口docs/188。"""
import argparse
from pathlib import Path


def build(prefix, action):
    assert action in ('join','leave')
    code=Path('/repo/tools/dosgolem_recruitment_empty_view_probe.py').read_text()
    code=code[:code.rindex("if __name__=='__main__':")]
    code=code.replace('empty_view_', 'empty_'+action+'_')
    code=code.replace('DQ3_EMPTY_VIEW_', 'DQ3_EMPTY_RECRUIT_')
    old='if m.Read16(cpu.Addr(ds,0x0722))==3 {next=0x1c;'
    assert code.count(old)==1
    code=code.replace(old,'if m.Read16(cpu.Addr(ds,0x0722))=='+('1' if action=='join' else '2')+' {next=0x1c;')
    old='if pilotStage>=5 && (pc==0x10624 || pc==0x10627 || pc==0x1062f || pc==0x10696 || pc==0x10699 || pc==0x1069e) {'
    assert code.count(old)==1
    points=[0x103d7,0x103e1,0x103ec,0x103f4,0x1046a,0x1046d,0x10472,0x104c4,0x104c8,0x104cd,0x105bc,0x105bf,0x105c4,0x103ae,0x103c2,0x103d6]
    code=code.replace(old,'if pilotStage>=5 && ('+' || '.join('pc=='+hex(pc) for pc in points)+') {')
    old='raw5060=%d roster_flags='
    assert code.count(old)==1
    code=code.replace(old,'raw5060=%d raw5077=%d roster_flags=')
    old='m.Read16(cpu.Addr(ds,0x5060)))'
    assert code.count(old)==1
    code=code.replace(old,'m.Read16(cpu.Addr(ds,0x5060)),m.Read8(cpu.Addr(ds,0x5077)))')
    old='fmt.Printf("\\n")\n            }\n            if pilotStage>=5 && pc==0x21414'
    assert code.count(old)==1
    code=code.replace(old,'''fmt.Printf(" party_pointers=")
                for i:=uint16(0);i<8;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x4f15+i)))}
                fmt.Printf("\\n")
            }
            if pilotStage>=5 && pc==0x21414''')
    ns={'__file__':__file__,'__name__':'empty_recruit_builder'}
    exec(compile(code,__file__,'exec'),ns)
    return ns['build'](prefix)


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prefix',required=True)
    parser.add_argument('--action',choices=('join','leave'),required=True)
    args=parser.parse_args()
    if not args.prefix or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.prefix):
        parser.error('prefix只能使用英數、連字號與底線')
    build(args.prefix,args.action)['main']()


if __name__=='__main__':
    main()
