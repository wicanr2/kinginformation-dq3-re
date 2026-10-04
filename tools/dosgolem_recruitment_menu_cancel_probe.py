"""正常空Join選Yes後，在招募主選單按Esc；有限來源入口docs/188。"""
import argparse
from dosgolem_recruitment_empty_yes_probe import build as build_yes


def build(prefix):
    ns=build_yes(prefix,'join')
    hook=ns['GATE_HOOK']
    start=hook.index('                    if choice && pilotRecruitPhase==11 && pilotRecord==528')
    end=hook.index('\n                }',start)
    assert 'pilotRecruitPhase==12' in hook[start:end]
    hook=hook[:start]+'''                    if choice && pilotRecruitPhase==11 && pilotRecord==528 && m.Read16(cpu.Addr(ds,0x071e))==3 {
                        next=0x01;kind="recruit_menu_cancel";pilotRecruitPhase=13
                    }'''+hook[end:]
    old='pc==0x10378 ||';assert hook.count(old)==1
    hook=hook.replace(old,'pc==0x10387 || pc==0x1038c || '+old)
    old='fmt.Printf(" party_pointers=")';assert hook.count(old)==1
    hook=hook.replace(old,'fmt.Printf(" raw0726=%d party_pointers=",m.Read8(cpu.Addr(ds,0x0726)))')
    ns['GATE_HOOK']=hook;ns['__file__']=__file__
    return ns


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prefix',required=True)
    args=parser.parse_args()
    if not args.prefix or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.prefix):parser.error('prefix只能使用英數、連字號與底線')
    build(args.prefix)['main']()
