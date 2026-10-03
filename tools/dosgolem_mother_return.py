"""母親返回後首次南行的正常冷啟動来源；入口 docs/188。"""
from pathlib import Path
import hashlib,json,os,subprocess,tempfile

GATE_HOOK = '\n        if m.Steps>=1940000000 {\n            pc:=uint32(m.CPU.Seg[cpu.CS])*16+uint32(m.CPU.IP)+0xef00\n            ds:=uint16(0x15ed)\n            if m.KeyIRQs!=previousKeyIRQs {\n                fmt.Printf("DQ3_KEY_DELIVERED step=%d count=%d port60=%02x CSIP=%04x:%04x\\n",m.Steps,m.KeyIRQs,m.In8(0x60),m.CPU.Seg[cpu.CS],m.CPU.IP)\n                previousKeyIRQs=m.KeyIRQs\n            }\n            capture:=func(label string) {\n                fmt.Printf("DQ3_GATE_CAPTURE step=%d label=%s ida_linear=%05x irqs=%d player_x=%d player_y=%d raw0b24=%04x SI=%04x DI=%04x raw252e=%04x raw259b=%d raw259e=%d raw251a=%d raw0b34=%02x raw4f1f=%04x seed=%04x flags=",m.Steps,label,pc,m.KeyIRQs,m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),m.Read16(cpu.Addr(ds,0x0b24)),m.CPU.R[cpu.SI],m.CPU.R[cpu.DI],m.Read16(cpu.Addr(ds,0x252e)),m.Read8(cpu.Addr(ds,0x259b)),m.Read8(cpu.Addr(ds,0x259e)),m.Read8(cpu.Addr(ds,0x251a)),m.Read8(cpu.Addr(ds,0x0b34)),m.Read16(cpu.Addr(ds,0x4f1f)),m.Read16(cpu.Addr(ds,0x0b5a)))\n                for i:=uint16(0);i<64;i++ {fmt.Printf("%02x",m.Read8(cpu.Addr(ds,0x4f70+i)))}\n                fmt.Printf("\\n")\n                path:="/work/dosgolem-opening/__PREFIX__-"+label\n                if err:=writeScreen(m,path+".png",0,0);err!=nil {die(err)}\n                if err:=os.WriteFile(path+".bin",m.Indexed(),0o644);err!=nil {die(err)}\n            }\n            if !gateStarted && pc==0x1997c && m.KeyIRQs==76 && m.KeyQueueLen()==0 {\n                capture("before-down")\n                gateStarted=true;gateQueued=m.Steps\n                m.SetNextKey(m.Steps+1);m.QueueKey(0x50)\n                fmt.Printf("DQ3_GATE_QUEUED step=%d scan=50 expected_irqs=78\\n",m.Steps)\n            }\n            if gateStarted && !gateSeen[pc] {\n                switch pc {\n                case 0x1020b,0x10221,0x1022d,0x10232,0x1023c,0x10242,0x10245,0x1024b,0x216d8,0x21133:\n                    gateSeen[pc]=true;capture(fmt.Sprintf("pc-%05x",pc))\n                }\n            }\n            if gateStarted && pc==0x1991d && m.Steps>gateQueued+11000000 && m.KeyIRQs==78 && m.KeyQueueLen()==0 {\n                capture("after-down")\n                if !gateSeen[0x1020b] || !gateSeen[0x1024b] {panic("handler55 incomplete")}\n                fmt.Printf("DQ3_GATE_DONE step=%d total_inputs=39 irqs=%d\\n",m.Steps,m.KeyIRQs)\n                break\n            }\n        }\n'
def main():
    out=Path('/work/dosgolem-opening')
    prefix='issue4-mother-return-gate-r3'
    assert out.is_dir() and out.stat().st_uid==os.getuid()
    assert not list(out.glob(prefix+'-*')) and not (out/(prefix+'.log')).exists(), 'output already exists; preserve previous evidence'
    parent=out/'issue4-mother-finish-receipt.json'
    assert hashlib.sha256(parent.read_bytes()).hexdigest()=='9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    real_run=subprocess.run
    class Built(Exception): pass
    binary=out/f'{prefix}-probe'
    frozen=out/f'{prefix}-probe-source.go'
    metadata={}
    def instrument(args,**kwargs):
        if args[:2]==['go','build']:
            src=Path(kwargs['cwd']);p=src/'cmd/probe/main.go';s=p.read_text()
            marker='\tfor m.Steps < *steps && !m.CPU.Halted && !d.Exited {\n'
            assert s.count(marker)==1
            start=s.index(marker)+len(marker);end=s.index('\t\t// -ega-every',start)
            s=s[:start]+'\t\tif m.Steps<=1940000000 {\n'+s[start:end]+'\t\t}\n'+GATE_HOOK.replace('__PREFIX__',prefix)+s[end:]
            declaration='\tvar gateQueued uint64\n\tvar gateStarted bool\n\tvar gateSeen=map[uint32]bool{}\n'
            s=s.replace(marker,declaration+marker,1).replace('issue4-mother-finish',prefix)
            p.write_text(s)
            real_run(['gofmt','-w',str(p)],check=True)
            frozen.write_bytes(p.read_bytes())
            for name,path in [('files','internal/dos/files.go'),('bios','internal/dos/bios.go'),('vga','internal/machine/vga.go')]:
                metadata['upstream_'+name+'_sha256']=hashlib.sha256((Path('/dosgolem')/path).read_bytes()).hexdigest()
                metadata['patched_'+name+'_sha256']=hashlib.sha256((src/path).read_bytes()).hexdigest()
            args=list(args);args[args.index('-o')+1]=str(binary)
            result=real_run(args,**kwargs);assert result.returncode==0
            raise Built()
        return real_run(args,**kwargs)
    subprocess.run=instrument
    os.environ['DQ3_NEWGAME_PROBE_SCENARIO']='mother_finish'
    generator=Path('/repo/tools/dosgolem_newgame_probe.py')
    namespace={'__name__':'__main__','__file__':str(generator)}
    try:
        with tempfile.TemporaryDirectory(prefix='dq3-mother-return-build-') as directory:
            s=generator.read_text();marker="out = Path('/work/dosgolem-opening')";assert s.count(marker)==1
            s=s.replace(marker,'out = Path('+repr(directory)+')')
            exec(compile(s,str(generator),'exec'),namespace)
    except Built: pass
    finally: subprocess.run=real_run
    assert len(namespace['keys'])==38
    captures=namespace['captures']
    dump=';'.join(f'{step}:{out}/{prefix}-{label}.png' for step,label in captures)
    args=[str(binary),'-exe','/repo/assets_raw/DQ3.EXE','-root','/repo/assets_raw','-steps','2000000001','-trace','0','-log-calls','-dump-at',dump]
    metadata.update(original_size=115282,original_sha256='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c',
                    upstream_revision='2f44a68ebfc54b28fb15dd4a34510b0b04a5415d',docker_image='dq3-ebiten-test:20260822-r1',
                    scenario='mother_return_first_down_cold',seed='1357',seed_configured_before_execution=True,
                    parent_receipt_sha256=hashlib.sha256(parent.read_bytes()).hexdigest(),
                    probe_source_sha256=hashlib.sha256(frozen.read_bytes()).hexdigest(),probe_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                    generator_sha256=hashlib.sha256(generator.read_bytes()).hexdigest(),producer_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
                    build_flags=['-trimpath','-p','2'],args=args,normal_prefix_inputs=namespace['keys'],
                    state_restore=False,gameplay_state_injection=False)
    (out/f'{prefix}-meta.json').write_text(json.dumps(metadata,ensure_ascii=False,indent=2)+'\n')
    with (out/(prefix+'.log')).open('w') as stream:
        result=real_run(args,stdout=stream,stderr=subprocess.STDOUT,timeout=1500)
    assert result.returncode==0
    assert 'DQ3_GATE_DONE ' in (out/(prefix+'.log')).read_text()
    print('正常母親返回加一次Down的冷啟動來源已重生；下一步由獨立validator接受收據。')

if __name__=='__main__': main()
