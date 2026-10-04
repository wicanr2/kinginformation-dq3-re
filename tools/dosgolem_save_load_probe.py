"""原版正常招募返回後的 F5／F6 DRAFT 探針；入口 docs/188。"""
import argparse
from pathlib import Path

from dosgolem_recruitment_menu_cancel_probe import build as build_menu_cancel


def build(prefix):
    ns = build_menu_cancel(prefix)
    scratch = Path('/work') / (prefix + '-scratch')
    assert not scratch.exists(), '保留既有存檔證據，必須使用新 prefix'
    assert scratch.parent.stat().st_uid == 1000
    scratch.mkdir()
    hook = ns['GATE_HOOK']
    # 使用既有 DOS.Scratch 契約，不改 CPU、遊戲記憶體或虛擬時間。
    hook = '''
        if d.Scratch=="" {
            d.Scratch="__SCRATCH__"
            fmt.Printf("DQ3_SAVE_SCRATCH step=%d path=%s\\n",m.Steps,d.Scratch)
        }
        if m.Steps>=2800000000 {
            data,err:=json.MarshalIndent(d.FileOps,"","  ");if err!=nil {die(err)}
            if err:=os.WriteFile("/work/dosgolem-opening/__PREFIX__-fileops.json",data,0o644);err!=nil {die(err)}
            fmt.Printf("DQ3_SAVE_BOUND step=%d packet=%d CSIP=%04x:%04x\\n",m.Steps,pilotPackets,m.CPU.Seg[cpu.CS],m.CPU.IP)
            break
        }
''' .replace('__SCRATCH__', str(scratch)) + hook
    old = 'if pilotStage==6 && pilotRecruitPhase==13 && field {'
    assert hook.count(old) == 1
    hook = hook.replace(old, '''if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==194 {
                    data,err:=json.MarshalIndent(d.FileOps,"","  ");if err!=nil {die(err)}
                    if err:=os.WriteFile("/work/dosgolem-opening/__PREFIX__-fileops.json",data,0o644);err!=nil {die(err)}
                    fmt.Printf("DQ3_SAVE_OBSERVED step=%d packet=%d phase=%s record=%d scratch=%s\\n",m.Steps,pilotPackets,phase,pilotRecord,d.Scratch)''')
    marker = '                if next!=0 {'
    assert hook.count(marker) == 1
    hook = hook.replace(marker, '''                if pilotStage==6 && pilotRecruitPhase==13 && pilotPackets==193 && field {
                    next=0x3f;kind="save_f5"
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
