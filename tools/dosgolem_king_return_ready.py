"""冷啟動逐鍵回程。正常100次謁見及所有既有畫面保持；入口與限制見docs/188。"""
from pathlib import Path
import os
import subprocess
import tempfile

prefix = 'issue4-return-cold-ready-r4'
# 共用產生器會先逐檔按 SHA-256 歸檔既有前綴，再清除舊產物並重生。
# 不在外層拒絕第二次執行；本批已接受的 frozen generation／Go 不改寫。
source = Path('/repo/tools/dosgolem_newgame_probe.py').read_text()
old = "'king_text':'issue4-king-text',"
assert source.count(old) == 1
source = source.replace(old, "'king_text':'"+prefix+"',", 1)
assert source.count("'-trace','16'") == 1
source = source.replace("'-trace','16'", "'-trace','0'", 1)
assert source.count('    stop = 3320000000\n') == 1
source = source.replace('    stop = 3320000000\n', '    stop = 9000000000\n', 1)
marker = '    probe.write_text(probetext)\n'
assert source.count(marker) == 1
addition = r'''
    ready_marker = '\tfor m.Steps < *steps && !m.CPU.Halted && !d.Exited {\n'
    declarations = '\tvar dq3ReturnQueued uint64\n\tvar dq3ReturnOrdinal, dq3ReturnPackets int\n\tvar dq3ReturnPending, dq3ReturnIdlePending bool\n'
    hook = r"""
        if m.KeyIRQs != previousKeyIRQs {
            fmt.Printf("DQ3_KEY_DELIVERED step=%d count=%d port60=%02x CSIP=%04x:%04x\n",m.Steps,m.KeyIRQs,m.In8(0x60),m.CPU.Seg[cpu.CS],m.CPU.IP)
            previousKeyIRQs=m.KeyIRQs
        }
        rdyPC:=uint32(m.CPU.Seg[cpu.CS])*16+uint32(m.CPU.IP)+0xef00
        rdyDS:=uint16(0x15ed)
        if dq3ReturnPending && rdyPC==0x11991 {
            fmt.Printf("DQ3_RETURN_CAMERA step=%d ida_linear=%05x ordinal=%d player_x=%d player_y=%d raw0b24=%04x origin_x=%d origin_y=%d width=%d height=%d\n",m.Steps,rdyPC,dq3ReturnOrdinal,m.Read16(cpu.Addr(rdyDS,0x4f33)),m.Read16(cpu.Addr(rdyDS,0x4f35)),m.Read16(cpu.Addr(rdyDS,0x0b24)),int16(m.Read16(cpu.Addr(rdyDS,0x4f25))),int16(m.Read16(cpu.Addr(rdyDS,0x4f27))),m.Read16(cpu.Addr(rdyDS,0x4f21)),m.Read16(cpu.Addr(rdyDS,0x4f23)))
        }
        if dq3ReturnPending && m.Steps>dq3ReturnQueued+11000000 && rdyPC==0x1991d && m.KeyIRQs==uint64(200+dq3ReturnPackets*2) && m.KeyQueueLen()==0 {
            actor:=m.Read16(cpu.Addr(rdyDS,0x4f15))
            fmt.Printf("DQ3_RETURN_READY step=%d ida_linear=%05x ordinal=%d queued_step=%d ticks=%d pit_divisor=%d actual_ds=%04x dgroup=%04x player_x=%d player_y=%d raw0b24=%04x raw4f1f=%04x raw4f25=%d raw4f27=%d gold_lo=%04x gold_hi=%04x raw0b34=%02x seed=%04x actor=",m.Steps,rdyPC,dq3ReturnOrdinal,dq3ReturnQueued,m.Ticks,m.PITDivisor(),m.CPU.Seg[cpu.DS],rdyDS,m.Read16(cpu.Addr(rdyDS,0x4f33)),m.Read16(cpu.Addr(rdyDS,0x4f35)),m.Read16(cpu.Addr(rdyDS,0x0b24)),m.Read16(cpu.Addr(rdyDS,0x4f1f)),int16(m.Read16(cpu.Addr(rdyDS,0x4f25))),int16(m.Read16(cpu.Addr(rdyDS,0x4f27))),m.Read16(cpu.Addr(rdyDS,0x4f37)),m.Read16(cpu.Addr(rdyDS,0x4f39)),m.Read8(cpu.Addr(rdyDS,0x0b34)),m.Read16(cpu.Addr(rdyDS,0x0b5a)))
            for i:=uint16(0);i<128;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(rdyDS,actor+i)))}
            fmt.Printf(" flags=")
            for i:=uint16(0);i<64;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(rdyDS,0x4f70+i)))}
            fmt.Printf("\n")
            path:=fmt.Sprintf("/work/dosgolem-opening/__PREFIX__-return-ready-%02d",dq3ReturnOrdinal)
            if err:=writeScreen(m,path+".png",0,0);err!=nil {die(err)}
            if err:=os.WriteFile(path+".bin",m.Indexed(),0o644);err!=nil {die(err)}
            dq3ReturnPending=false
            if dq3ReturnOrdinal==85 {
                fmt.Printf("DQ3_RETURN_DONE step=%d inputs=%d packets=%d irqs=%d\n",m.Steps,dq3ReturnOrdinal,dq3ReturnPackets,m.KeyIRQs)
                break
            }
        }
        if m.Steps>=3324000000 && rdyPC==0x21133 && !dq3ReturnIdlePending && m.KeyQueueLen()==0 && m.KeyIRQs==uint64(200+dq3ReturnPackets*2) {
            ss,sp:=m.CPU.Seg[cpu.SS],m.CPU.R[cpu.SP]
            if m.Read16(cpu.Addr(ss,sp+6))==0x7e00 && m.Read16(cpu.Addr(ss,sp+8))==0x0110 {
                dq3ReturnIdlePending=true
                dq3ReturnPackets++
                fmt.Printf("DQ3_RETURN_IDLE step=%d ida_linear=%05x ordinal=%d player_x=%d player_y=%d raw0b24=%04x return_ip=7e00 return_cs=0110 key_flag=%02x\n",m.Steps,rdyPC,dq3ReturnOrdinal,m.Read16(cpu.Addr(rdyDS,0x4f33)),m.Read16(cpu.Addr(rdyDS,0x4f35)),m.Read16(cpu.Addr(rdyDS,0x0b24)),m.Read8(cpu.Addr(rdyDS,0x2856)))
                path:=fmt.Sprintf("/work/dosgolem-opening/__PREFIX__-idle-%02d",dq3ReturnPackets)
                if err:=writeScreen(m,path+".png",0,0);err!=nil {die(err)}
                if err:=os.WriteFile(path+".bin",m.Indexed(),0o644);err!=nil {die(err)}
                m.SetNextKey(m.Steps+1)
                m.QueueKey(0x1c)
                fmt.Printf("DQ3_RETURN_QUEUED step=%d kind=idle_dismiss ordinal=%d scan=1c packet=%d\n",m.Steps,dq3ReturnOrdinal,dq3ReturnPackets)
            }
        }
        if rdyPC==0x1991d && dq3ReturnIdlePending && m.KeyQueueLen()==0 && m.KeyIRQs==uint64(200+dq3ReturnPackets*2) {dq3ReturnIdlePending=false}
        if !dq3ReturnPending && !dq3ReturnIdlePending && rdyPC==0x1997c && m.Steps>=3324000000 && m.KeyQueueLen()==0 && m.KeyIRQs==uint64(200+dq3ReturnPackets*2) {
            dq3ReturnOrdinal++
            dq3ReturnQueued=m.Steps
            dq3ReturnPending=true
            dq3ReturnPackets++
            scan:=uint8(0x50)
            if dq3ReturnOrdinal>=55 && dq3ReturnOrdinal<=63 {scan=0x4b}
            if dq3ReturnOrdinal==64 || (dq3ReturnOrdinal>=72 && dq3ReturnOrdinal<=82) {scan=0x48}
            if dq3ReturnOrdinal>=65 && dq3ReturnOrdinal<=71 {scan=0x4b}
            if dq3ReturnOrdinal>=83 {scan=0x4d}
            m.SetNextKey(m.Steps+1)
            m.QueueKey(scan)
            fmt.Printf("DQ3_RETURN_QUEUED step=%d kind=motion ida_linear=1997c ordinal=%d scan=%02x packet=%d\n",m.Steps,dq3ReturnOrdinal,scan,dq3ReturnPackets)
        }
""".replace('__PREFIX__',prefix)
    assert probetext.count(ready_marker)==1
    loop_start=probetext.index(ready_marker)
    body_start=probetext.index('\t\t// -ega-every',loop_start)
    old_observers=probetext[loop_start+len(ready_marker):body_start]
    replacement=declarations+ready_marker+'\t\tif m.Steps<=3320000000 {\n'+old_observers+'\t\t}\n\t\tif m.Steps>=3324000000 {\n'+hook+'\t\t}\n'
    probetext=probetext[:loop_start]+replacement+probetext[body_start:]
'''
# These snippets are Python source; their Go raw strings need one literal escape.
addition = addition.replace('\\\\t', '\\t').replace('\\\\n', '\\n')
source = source.replace(marker, addition+marker, 1)
old = '        deadline = min(step for step,_ in captures if step > queued)\n'
assert source.count(old) == 1
source = source.replace(old, '        deadline = min(step for step,_ in captures if step > queued) if queued <= 3320000000 else queued+11000000\n', 1)
marker = "    expected_steps = [event for step, _ in keys for event in (step+1, step+500001)]\n"
assert source.count(marker) == 1
addition = """    ready_queue=[dict(re.findall(r'(\\w+)=(\\S+)',line)) for line in lines if line.startswith('DQ3_RETURN_QUEUED ')]
    assert len([x for x in ready_queue if x['kind']=='motion'])==85 and any(line.startswith('DQ3_RETURN_DONE ') for line in lines)
    keys += [(int(event['step']),int(event['scan'],16)) for event in ready_queue]
    meta['player_input']=[{'queued_step':step,'scan':hex(scan)} for step,scan in keys]
"""
source = source.replace(marker, addition+marker, 1)
marker = "    meta['artifacts'] = []\n"
assert source.count(marker) == 1
addition = """    meta['scenario']='king_return_cold_ready'
    meta['scope']='冷啟動100次謁見後，85次正常步行；原版自然等待窗以Enter關閉，逐鍵等runner返回；不重設seed或注入狀態'
    for label,key in [('DQ3_RETURN_READY ','return_events'),('DQ3_RETURN_CAMERA ','camera_events'),('DQ3_RETURN_QUEUED ','queued_events'),('DQ3_RETURN_IDLE ','return_idle_events')]:
        meta[key]=[line for line in lines if line.startswith(label)]
    meta['input_contract']='motion queued at native1997C after the idle check returns; initialized21133 idle windows closed with normal Enter; all extra inputs recorded'
"""
source = source.replace(marker, addition+marker, 1)
compile(source, 'cold-ready-generation.py', 'exec')
with tempfile.TemporaryDirectory(prefix='dq3-cold-ready-generation-') as directory:
    script = Path(directory)/'generation.py'
    script.write_text(source)
    result = subprocess.run(['python3', str(script)], env=dict(os.environ, DQ3_NEWGAME_PROBE_SCENARIO='king_text'), timeout=2100)
    raise SystemExit(result.returncode)
