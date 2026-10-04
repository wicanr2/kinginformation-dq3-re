"""正常空Join／單人Leave選Yes、重選同動作、No返回；有界入口docs/188。"""
import argparse
from pathlib import Path
from dosgolem_recruitment_empty_probe import build as build_empty


def build(prefix, action):
    ns=build_empty(prefix,action)
    hook=ns['GATE_HOOK']
    old='''                    if choice && pilotRecruitPhase==10 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x4d;kind="empty_ACTION_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="empty_ACTION_decline";pilotRecruitPhase=13}
                    }'''.replace('ACTION',action)
    assert hook.count(old)==1
    target=1 if action=='join' else 2
    new='''                    if choice && pilotRecruitPhase==10 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        if m.Read16(cpu.Addr(ds,0x0722))!=1 {panic("空清單Yes初始游標不符")}
                        next=0x1c;kind="empty_ACTION_continue_yes";pilotRecruitPhase=11
                    }
                    if choice && pilotRecruitPhase==11 && pilotRecord==528 && m.Read16(cpu.Addr(ds,0x071e))==3 {
                        next=0x50;kind="empty_ACTION_again_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==TARGET {next=0x1c;kind="empty_ACTION_again_open";pilotRecruitPhase=12}
                    }
                    if choice && pilotRecruitPhase==12 && pilotRecord==540 && m.Read16(cpu.Addr(ds,0x071e))==2 {
                        next=0x4d;kind="empty_ACTION_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;kind="empty_ACTION_decline";pilotRecruitPhase=13}
                    }'''.replace('ACTION',action).replace('TARGET',str(target))
    hook=hook.replace(old,new)
    old='pc==0x103d7 ||';assert hook.count(old)==1
    hook=hook.replace(old,'pc==0x10378 || pc==0x10384 || pc==0x103b9 || pc==0x103c0 || '+old)
    ns['GATE_HOOK']=hook;ns['__file__']=__file__
    return ns


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prefix',required=True);parser.add_argument('--action',choices=('join','leave'),required=True)
    args=parser.parse_args()
    if not args.prefix or any(c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in args.prefix):parser.error('prefix只能使用英數、連字號與底線')
    build(args.prefix,args.action)['main']()
