"""正常入隊短曲完成、540選No、541等待及返回；只讀玩家層觀察。

入口 docs/188；不修改播放旗標、時鐘、CPU、動畫或原始資料。
"""
from pathlib import Path

code = Path('/repo/tools/dosgolem_recruitment_party_probe.py').read_text()
code = code.replace('issue4-recruit-party-r2', 'issue4-recruit-return-r2')
marker = "env['rebuilt']['main']()"
assert code.rstrip().endswith(marker)
code = code[:code.rindex(marker)]
ns = {'__file__': __file__, '__name__': 'join_return_builder'}
exec(compile(code, __file__, 'exec'), ns)
hook = ns['hook']
old = '''                if pilotPackets==199 && audioWait && pilotRecord==538 {
                    fmt.Printf("DQ3_RECRUIT_DONE step=%d packets=%d irqs=%d scope=record538_audio_wait_entry complete_join=false audio_completion=false\\n",m.Steps,pilotPackets,m.KeyIRQs)
                    break
                }
'''
assert hook.count(old) == 1
hook = hook.replace(old, '''                if pilotPackets==199 && audioWait && pilotRecord==538 {
                    fmt.Printf("DQ3_JOIN_AUDIO_ENTRY step=%d packet=%d irqs=%d record=%d\\n",m.Steps,pilotPackets,m.KeyIRQs,pilotRecord)
                }
''')
marker = '            phase:="none"'
assert hook.count(marker) == 1
observer = r'''
            if pilotStage==6 && pilotPackets>=199 && (pc==0x10459 || pc==0x1045e || pc==0x10469 || pc==0x10398 || pc==0x103ae || pc==0x103b6) {
                fmt.Printf("DQ3_JOIN_RETURN step=%d packet=%d ida_linear=%05x actual_DS=%04x irqs=%d record=%d raw4f1f=%04x raw4f15=%04x raw4f17=%04x raw4f19=%04x raw4f1b=%04x raw5077=%d choice_count=%d choice_cursor=%d player_x=%d player_y=%d seed=%04x flags=",
                    m.Steps,pilotPackets,pc,m.CPU.Seg[cpu.DS],m.KeyIRQs,pilotRecord,m.Read16(cpu.Addr(ds,0x4f1f)),m.Read16(cpu.Addr(ds,0x4f15)),m.Read16(cpu.Addr(ds,0x4f17)),m.Read16(cpu.Addr(ds,0x4f19)),m.Read16(cpu.Addr(ds,0x4f1b)),m.Read8(cpu.Addr(ds,0x5077)),m.Read16(cpu.Addr(ds,0x071e)),m.Read16(cpu.Addr(ds,0x0722)),m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),m.Read16(cpu.Addr(ds,0x0b5a)))
                for i:=uint16(0);i<64;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x4f70+i)))}
                fmt.Printf("\n")
            }
'''
assert 'm.Write' not in observer and 'SetNextKey' not in observer
hook = hook.replace(marker, observer + marker)
# The inherited registration guard only covered inputs before the music wait.
# Observe the natural flag after packet199; never clear it or change guest time.
old = '            if rawFlag&0x4000!=0 {panic("原生計時旗標4000出現；拒絕來源")} '
old = old.rstrip()
assert hook.count(old) == 1
hook = hook.replace(old, r'''
            if rawFlag&0x4000!=0 {
                if !(pilotStage==6 && pilotPackets>=199) {panic("原生計時旗標4000出現於既有前綴；拒絕來源")}
                if !pilotObserveSeen[0x4000] {
                    pilotObserveSeen[0x4000]=true
                    fmt.Printf("DQ3_JOIN_TIMER_FLAG step=%d packet=%d ida_linear=%05x raw0013=%04x ticks=%d irqs=%d\n",m.Steps,pilotPackets,pc,rawFlag,m.Ticks,m.KeyIRQs)
                }
            }
            if pilotStage==6 && pilotPackets>=199 && m.Steps%50000000==0 {
                fmt.Printf("DQ3_JOIN_PROGRESS step=%d packet=%d ida_linear=%05x raw0013=%04x ticks=%d irqs=%d\n",m.Steps,pilotPackets,pc,rawFlag,m.Ticks,m.KeyIRQs)
            }
''')
wrapper = ns['env']['base_code']
assert wrapper.count("'9000000001'") == 1
wrapper = wrapper.replace("'9000000001'", "'2500000001'")
rebuilt = {'__file__': __file__, '__name__': 'bounded_join_return_wrapper'}
exec(compile(wrapper, __file__, 'exec'), rebuilt)
rebuilt['GATE_HOOK'] = hook
rebuilt['main']()
