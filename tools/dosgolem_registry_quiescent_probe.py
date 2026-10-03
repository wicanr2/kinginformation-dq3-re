"""原生空讀鍵與正式事件完成點的 DRAFT 冷啟動；入口 docs/188。"""
from pathlib import Path
import hashlib,json

root=Path('/repo');out=Path('/work/dosgolem-opening')
prefix='issue4-registry-quiescent-r2'
assert out.is_dir() and out.stat().st_uid==1000
assert not list(out.glob(prefix+'-*')) and not (out/(prefix+'.log')).exists()
back=[]
for ordinal in range(1,86):
 scan=0x50
 if 55<=ordinal<=63:scan=0x4b
 if ordinal==64 or 72<=ordinal<=82:scan=0x48
 if 65<=ordinal<=71:scan=0x4b
 if ordinal>=83:scan=0x4d
 back.append(scan)
registry=[0x4b]*3+[0x50]*4+[0x4b]*2+[0x48,0x4b,0x48,0x1c]
declarations='''
    var pilotPackets, pilotStage, pilotReturn, pilotRegistry, pilotNames, pilotUps int
    var pilotPending, pilotConsumed, pilotEmpty, pilotHello, pilotAccepted, pilotCancelled, pilotFinishing bool
    var pilotScan uint8
    var pilotQueued uint64
    var pilotKind string
    var pilotRecord uint16
    var pilotObserveSeen=map[uint32]bool{}
    var pilotBack=[]uint8{__BACK__}
    var pilotRegistryScans=[]uint8{__REGISTRY__}
'''.replace('__BACK__',','.join(hex(v) for v in back)).replace('__REGISTRY__',','.join(hex(v) for v in registry))
hook=r'''
        if m.Steps>=1940000000 {
            pc:=uint32(m.CPU.Seg[cpu.CS])*16+uint32(m.CPU.IP)+0xef00
            ds:=uint16(0x15ed)
            rawFlag:=m.Read16(cpu.Addr(ds,0x13))
            if pilotPackets<=3 && !pilotObserveSeen[pc] && (pc==0x11b0b || pc==0x19530 || pc==0x196d2 || pc==0x1020b || pc==0x10232 || pc==0x10245) {
                pilotObserveSeen[pc]=true
                fmt.Printf("DQ3_QUIESCENT_EVENT step=%d packet=%d ida_linear=%05x player_x=%d player_y=%d raw4f46=%04x raw258c=%04x raw4f1f=%04x\n",m.Steps,pilotPackets,pc,m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),m.Read16(cpu.Addr(ds,0x4f46)),m.Read16(cpu.Addr(ds,0x258c)),m.Read16(cpu.Addr(ds,0x4f1f)))
                path:=fmt.Sprintf("/work/dosgolem-opening/__PREFIX__-event-%05x",pc)
                if err:=writeScreen(m,path+".png",0,0);err!=nil {die(err)}
                if err:=os.WriteFile(path+".bin",m.Indexed(),0o644);err!=nil {die(err)}
            }

            if rawFlag&0x4000!=0 {panic("原生計時旗標4000出現；拒絕來源")}
            if m.KeyIRQs!=previousKeyIRQs {
                fmt.Printf("DQ3_KEY_DELIVERED step=%d count=%d port60=%02x CSIP=%04x:%04x\n",m.Steps,m.KeyIRQs,m.In8(0x60),m.CPU.Seg[cpu.CS],m.CPU.IP)
                previousKeyIRQs=m.KeyIRQs
            }
            if pc==0x21414 {
                pilotRecord=m.CPU.R[cpu.DI]
                if pilotRecord==0x0c06 && pilotStage==0 {pilotStage=1}
                if pilotRecord==550 {pilotHello=true}
                fmt.Printf("DQ3_QUIESCENT_RECORD step=%d stage=%d packet=%d record=%d player_x=%d player_y=%d raw0b24=%04x\n",m.Steps,pilotStage,pilotPackets,pilotRecord,m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),m.Read16(cpu.Addr(ds,0x0b24)))
            }
            flag:=m.Read8(cpu.Addr(ds,0x2856))
            delivered:=m.KeyIRQs==uint64(76+pilotPackets*2) && m.KeyQueueLen()==0
            if pc==0x1941c && uint8(m.CPU.R[cpu.AX]>>8)==0 && flag==0 && delivered && (!pilotPending || pilotConsumed) {
                if !pilotEmpty {fmt.Printf("DQ3_QUIESCENT_EMPTY step=%d packet=%d ida_linear=1941c ax=%04x irqs=%d raw4f46=%04x raw258c=%04x\n",m.Steps,pilotPackets,m.CPU.R[cpu.AX],m.KeyIRQs,m.Read16(cpu.Addr(ds,0x4f46)),m.Read16(cpu.Addr(ds,0x258c)))}
                pilotEmpty=true
            }
            field:=pc==0x1997c && pilotEmpty
            wait:=pc==0x21133 && m.CPU.Seg[cpu.DS]==ds && flag==0
            inline:=pc==0x216d8 && m.CPU.Seg[cpu.DS]==ds && flag==0
            choice:=pc==0x1f7b7 && m.CPU.Seg[cpu.DS]==ds && flag==0
            naming:=pc==0x11096 && m.CPU.Seg[cpu.DS]==ds && flag==0
            phase:="none"
            if field {phase="ready"}
            if wait {phase="waiting"}
            if inline {phase="inline_wait"}
            if choice {phase="choice"}
            if naming {phase="name"}
            if pilotPending && !pilotConsumed && m.KeyIRQs>=uint64(76+pilotPackets*2-1) {
                consumer:=pc==0x1941c && uint8(m.CPU.R[cpu.AX]>>8)==pilotScan
                consumer=consumer || (m.CPU.Seg[cpu.DS]==ds && (pc==0x2113a || pc==0x21155 || (pc==0x210ca && flag!=0)))
                if consumer {
                    pilotConsumed=true
                    fmt.Printf("DQ3_QUIESCENT_INPUT step=%d packet=%d stage=%d scan=%02x ida_linear=%05x key_flag=%02x ax=%04x irqs=%d\n",m.Steps,pilotPackets,pilotStage,pilotScan,pc,flag,m.CPU.R[cpu.AX],m.KeyIRQs)
                }
            }
            if pilotPending && pilotConsumed && delivered && phase!="none" && flag==0 {
                x,y:=m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35))
                if pilotKind=="king_up" && pilotUps<=3 && (x!=21 || int(y)!=16-pilotUps) {
                    fmt.Printf("DQ3_QUIESCENT_REJECT step=%d packet=%d reason=first_up_position got=%d,%d want=21,%d\n",m.Steps,pilotPackets,x,y,16-pilotUps)
                    panic("首三個正常Up與既有原生觀測不符；拒絕來源")
                }
                fmt.Printf("DQ3_QUIESCENT_CAPTURE step=%d packet=%d stage=%d kind=%s phase=%s ida_linear=%05x queued_step=%d scan=%02x ticks=%d pit_divisor=%d player_x=%d player_y=%d raw0b24=%04x raw0013=%04x raw0007=%04x last_record=%d gold_lo=%04x gold_hi=%04x return_ordinal=%d registry_ordinal=%d choice_count=%d choice_cursor=%d name_mode=%d name_cursor=%d name_length=%d actor=",m.Steps,pilotPackets,pilotStage,pilotKind,phase,pc,pilotQueued,pilotScan,m.Ticks,m.PITDivisor(),x,y,m.Read16(cpu.Addr(ds,0x0b24)),rawFlag,m.Read16(cpu.Addr(ds,7)),pilotRecord,m.Read16(cpu.Addr(ds,0x4f37)),m.Read16(cpu.Addr(ds,0x4f39)),pilotReturn,pilotRegistry,m.Read16(cpu.Addr(ds,0x071e)),m.Read16(cpu.Addr(ds,0x0722)),m.Read16(cpu.Addr(ds,0x26fc)),m.Read16(cpu.Addr(ds,0x26fe)),m.Read16(cpu.Addr(ds,0x270a)))
                actor:=m.Read16(cpu.Addr(ds,0x4f15))
                for i:=uint16(0);i<128;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,actor+i)))}
                fmt.Printf(" flags=")
                for i:=uint16(0);i<64;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x4f70+i)))}
                fmt.Printf("\n")
                path:=fmt.Sprintf("/work/dosgolem-opening/__PREFIX__-packet-%03d-%s",pilotPackets,phase)
                if err:=writeScreen(m,path+".png",0,0);err!=nil {die(err)}
                if err:=os.WriteFile(path+".bin",m.Indexed(),0o644);err!=nil {die(err)}
                pilotPending=false
                if pilotFinishing && field {
                    fmt.Printf("DQ3_QUIESCENT_DONE step=%d packets=%d irqs=%d cancelled=%t player_x=%d player_y=%d raw0b24=%04x\n",m.Steps,pilotPackets,m.KeyIRQs,pilotCancelled,x,y,m.Read16(cpu.Addr(ds,0x0b24)))
                    break
                }
            }
            if !pilotPending && delivered && phase!="none" && flag==0 {
                next:=uint8(0);kind:="none"
                x,y:=m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35))
                scene:=m.Read16(cpu.Addr(ds,0x0b24))
                kingDone:=m.Read8(cpu.Addr(ds,0x4f70+23/8))&1==0 && m.Read8(cpu.Addr(ds,0x4f70+24/8))&128!=0 && m.Read16(cpu.Addr(ds,0x4f37))==50 && m.Read16(cpu.Addr(ds,0x4f39))==0
                if pilotStage==1 && field && kingDone {
                    pilotStage=2
                    fmt.Printf("DQ3_QUIESCENT_KING_DONE step=%d packet=%d player_x=%d player_y=%d raw0b24=%04x\n",m.Steps,pilotPackets,x,y,scene)
                }
                if pilotStage==0 {
                    if wait || inline {next=0x1c;kind="preking_wait"}
                    if field {
                        if scene==0x08c7 && x==9 && y==7 {next=0x1c;kind="king_talk";pilotStage=1} else {next=0x48;kind="king_up";pilotUps++}
                    }
                } else if pilotStage==1 {
                    if wait || inline {next=0x1c;kind="king_wait"}
                    if field && !kingDone {panic("謁見未完成且已回field；拒絕來源")}
                    if choice || naming {panic("謁見進入非預期選單；拒絕來源")}
                } else if pilotStage==2 {
                    if wait || inline {next=0x1c;kind="idle_dismiss"}
                    if field {
                        if pilotReturn<len(pilotBack) {next=pilotBack[pilotReturn];pilotReturn++;kind="return_motion"} else {
                            if scene!=0x111d || x!=8 || y!=2 {panic("回程未到正常登錄所入口；拒絕來源")}
                            pilotStage=3
                        }
                    }
                }
                if pilotStage==3 {
                    if field && pilotRegistry<len(pilotRegistryScans) {next=pilotRegistryScans[pilotRegistry];pilotRegistry++;kind="registry_approach"}
                    if pilotRegistry==len(pilotRegistryScans) && pilotHello {pilotStage=4}
                }
                if pilotStage==4 {
                    if wait || inline {next=0x1c;kind="registry_wait"}
                    if choice && m.Read16(cpu.Addr(ds,0x071e))==2 && !pilotAccepted {next=0x1c;pilotAccepted=true;kind="registry_accept"}
                    if naming && pilotAccepted && !pilotCancelled && pilotNames<3 {next=[]uint8{0x48,0x4b,0x1c}[pilotNames];pilotNames++;kind="registry_name"}
                    if choice && m.Read16(cpu.Addr(ds,0x071e))==5 && pilotNames==3 && !pilotCancelled {
                        next=0x50;kind="registry_cancel_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==4 {next=0x1c;pilotCancelled=true;kind="registry_cancel"}
                    }
                    if choice && m.Read16(cpu.Addr(ds,0x071e))==2 && pilotCancelled {
                        next=0x4d;kind="registry_decline_cursor"
                        if m.Read16(cpu.Addr(ds,0x0722))==2 {next=0x1c;pilotFinishing=true;kind="registry_decline"}
                    }
                }
                if next!=0 {
                    if pilotPackets>=220 || pilotUps>80 {panic("正常來源有界上限；拒絕未完成來源")}
                    pilotPackets++;pilotScan=next;pilotQueued=m.Steps;pilotKind=kind
                    pilotPending=true;pilotConsumed=false;pilotEmpty=false
                    m.SetNextKey(m.Steps+1);m.QueueKey(next)
                    fmt.Printf("DQ3_QUIESCENT_QUEUED step=%d packet=%d stage=%d kind=%s scan=%02x ida_linear=%05x phase=%s\n",m.Steps,pilotPackets,pilotStage,kind,next,pc,phase)
                }
            }
        }
'''
base=root/'tools/dosgolem_mother_return.py';code=base.read_text()
for old,new in [("prefix='issue4-mother-return-gate-r3'",f"prefix={prefix!r}"),
                ("'-steps','2000000001'","'-steps','9000000001'"),("timeout=1500","timeout=4200"),
                ("scenario='mother_return_first_down_cold'","scenario='registry_quiescent_cold_draft'"),
                ("'DQ3_GATE_DONE '","'DQ3_QUIESCENT_DONE '"),
                ("正常母親返回加一次Down的冷啟動來源已重生；下一步由獨立validator接受收據。","原生空讀鍵DRAFT探針已停止；另需獨立來源稽核。")]:
 assert code.count(old)==1,old
 code=code.replace(old,new)
lines=code.splitlines();hits=[i for i,s in enumerate(lines) if s.strip().startswith('declaration=')];assert len(hits)==1
lines[hits[0]]='            declaration='+repr(declarations);code='\n'.join(lines)+'\n'
ns={'__name__':'quiescent_builder','__file__':__file__};exec(compile(code,'isolated-quiescent-builder','exec'),ns);ns['GATE_HOOK']=hook
plan={'scope':'DRAFT 正常空讀鍵；完成與否尚未接受','normal_prefix_inputs':38,
 'initial_up_expected':[[21,15],[21,14],[21,13]],'return_scans':back,'registry_scans':registry,
 'king_completion':'native field + clear17h/set18h + gold50; no fixed confirm count',
 'base_producer_sha256':hashlib.sha256(base.read_bytes()).hexdigest(),
 'producer_sha256':hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
 'state_injection':False,'clock_modified':False,'read_key_quiescence':'break IRQ delivered;1941C AH0;1997C'}
(Path('/work')/'issue4-registry-quiescent-r2-plan.json').write_text(json.dumps(plan,ensure_ascii=False,indent=2)+'\n')
ns['main']()
