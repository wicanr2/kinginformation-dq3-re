"""Docker-only：明示局部CPU重入、固定seed與四個cell layer；非正常玩家收據。

入口docs/188。沿已核對458生成Go與相同patched runtime建置；在原424
自然轉向入口暖機後，固定1e2c各執行一次原始12025，含原layer0正對照。
"""
from pathlib import Path
import hashlib,json,os,subprocess,tempfile

root=Path('/work/dosgolem-opening');pre='issue4-npc-turn-layers-component-r1'
assert root.is_dir() and root.stat().st_uid==os.getuid()==1000 and not list(root.glob(pre+'*'))
parent=root/'issue4-npc-move-state-normal-r1-probe-source.go'
assert hashlib.sha256(parent.read_bytes()).hexdigest()=='f4ec495afa57d001d181c165bc9141643dc246fd781eca8a6e646bd494186321'
original=(Path('/repo/assets_raw/DQ3.EXE')).read_bytes()
assert len(original)==115282 and hashlib.sha256(original).hexdigest()=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
meta=json.loads((root/'issue4-npc-move-state-normal-r1-meta.json').read_text())
s=parent.read_text().replace('issue4-npc-move-state-normal-r1',pre)
marker='\tvar pilotPackets, pilotStage, pilotReturn, pilotRegistry, pilotNames, pilotUps int'
declarations=r'''
    var componentCase=-1
    var componentRegs [8]uint16
    var componentSegments [4]uint16
    var componentIP,componentFlags uint16
    var componentStart uint64
    var componentMapSegment,componentCell,componentHeroCell uint32
    var componentCellWord,componentHeroWord uint16
    var componentNPC [112]byte
'''
assert s.count(marker)==1;s=s.replace(marker,declarations+'\n'+marker)
marker='            if rawFlag&0x4000!=0'
# The frozen source is gofmt-formatted, so use its tab-indented gate marker.
marker='\t\t\tif rawFlag&0x4000 != 0'
assert s.count(marker)==1
hook=r'''
            if componentCase<0 && m.Steps==2282686367 && pc==0x12025 {
                if m.Read16(cpu.Addr(ds,0xb5a))!=0x1e2c || m.CPU.R[cpu.BX]!=0x2650 || m.Read16(cpu.Addr(ds,0xb32))!=14 {panic("native warm entry differs")}
                componentRegs=m.CPU.R;componentSegments=m.CPU.Seg;componentIP=m.CPU.IP;componentFlags=m.CPU.Flags;componentStart=m.Steps
                mapseg:=m.Read16(cpu.Addr(ds,0x2536));base:=m.Read16(cpu.Addr(ds,0xb26));width:=m.Read16(cpu.Addr(ds,0xb28))
                componentMapSegment=uint32(mapseg)
                componentCell=uint32(base+(29*width+5)*2)
                componentHeroCell=uint32(base+(23*width+5)*2)
                componentCellWord=m.Read16(cpu.Addr(mapseg,uint16(componentCell)));componentHeroWord=m.Read16(cpu.Addr(mapseg,uint16(componentHeroCell)))
                for i:=range componentNPC {componentNPC[i]=m.Read8(cpu.Addr(ds,0xb66+uint16(i)))}
                componentCase=0
                fmt.Printf("DQ3_TURN_COMPONENT_WARM step=%d packet=%d seed=1e2c slot=6 count=14 original_word=%04x npc_slots=%x\n",m.Steps,pilotPackets,componentCellWord,componentNPC)
            }
            if componentCase>=0 && pc==0x12025 {
                word:=(componentCellWord&0x3fff)|uint16(componentCase)<<14
                m.CPU.R[cpu.BX]=word
                m.Write16(cpu.Addr(uint16(componentMapSegment),uint16(componentCell)),word)
                m.Write16(cpu.Addr(uint16(componentMapSegment),uint16(componentHeroCell)),(componentHeroWord&0x3fff)|uint16(componentCase)<<14)
                for i,b:=range componentNPC {m.Write8(cpu.Addr(ds,0xb66+uint16(i)),b)}
                m.Write8(cpu.Addr(ds,0xb96+6),(componentNPC[6*8+6]&0x3f)|uint8(componentCase)<<6)
                m.Write8(cpu.Addr(ds,0x2579),uint8(componentCase)<<6)
                m.Write16(cpu.Addr(ds,0xb5a),0x1e2c)
                fmt.Printf("DQ3_TURN_COMPONENT_ENTRY case=%d layer=%d step=%d seed=%04x word=%04x ctrl=%02x x=%d y=%d player_x=%d player_y=%d state_injection=true cpu_reentry=%t\n",componentCase,componentCase,m.Steps,m.Read16(cpu.Addr(ds,0xb5a)),word,m.Read8(cpu.Addr(ds,0xb99)),m.Read8(cpu.Addr(ds,0xb96)),m.Read8(cpu.Addr(ds,0xb97)),m.Read16(cpu.Addr(ds,0x4f33)),m.Read16(cpu.Addr(ds,0x4f35)),componentCase!=0)
            }
            if componentCase>=0 && (pc==0x12043 || pc==0x12065 || pc==0x12074 || pc==0x1207f || pc==0x12098 || pc==0x120b0 || pc==0x120a6) {
                fmt.Printf("DQ3_TURN_COMPONENT_NATIVE case=%d layer=%d step=%d ida_linear=%05x AX=%04x DX=%04x DI=%04x CX=%04x seed=%04x ctrl=%02x x=%d y=%d\n",componentCase,componentCase,m.Steps,pc,m.CPU.R[cpu.AX],m.CPU.R[cpu.DX],m.CPU.R[cpu.DI],m.CPU.R[cpu.CX],m.Read16(cpu.Addr(ds,0xb5a)),m.Read8(cpu.Addr(ds,0xb99)),m.Read8(cpu.Addr(ds,0xb96)),m.Read8(cpu.Addr(ds,0xb97)))
                if pc==0x120a6 {
                    if componentCase==3 {fmt.Printf("DQ3_TURN_COMPONENT_DONE step=%d cases=4 no_normal_path_claim=true\n",m.Steps);break}
                    componentCase++
                    m.CPU.R=componentRegs;m.CPU.Seg=componentSegments;m.CPU.IP=componentIP;m.CPU.Flags=componentFlags
                    continue
                }
            }
            if componentCase>=0 && m.Steps-componentStart>50000 {panic("native component exceeded fixed instruction bound")}
'''
s=s.replace(marker,hook+'\n'+marker)
real_run=subprocess.run
class Built(Exception):pass
binary=root/(pre+'-probe');frozen=root/(pre+'-probe-source.go');runtime={}
def instrument(args,**kwargs):
    if args[:2]==['go','build']:
        source=Path(kwargs['cwd']);main=source/'cmd/probe/main.go';main.write_text(s)
        real_run(['gofmt','-w',str(main)],check=True);frozen.write_bytes(main.read_bytes())
        for name,rel in [('files','internal/dos/files.go'),('bios','internal/dos/bios.go'),('vga','internal/machine/vga.go')]:
            for prefix,path in [('upstream',Path('/dosgolem')/rel),('patched',source/rel)]:
                key=prefix+'_'+name+'_sha256';runtime[key]=hashlib.sha256(path.read_bytes()).hexdigest();assert runtime[key]==meta[key]
        args=list(args);args[args.index('-o')+1]=str(binary);result=real_run(args,**kwargs);assert result.returncode==0;raise Built()
    return real_run(args,**kwargs)
subprocess.run=instrument;os.environ['DQ3_NEWGAME_PROBE_SCENARIO']='mother_finish'
generator=Path('/repo/tools/dosgolem_newgame_probe.py')
namespace={'__name__':'__main__','__file__':str(generator)}
try:
    with tempfile.TemporaryDirectory(prefix='dq3-npc-turn-build-') as directory:
        code=generator.read_text();marker="out = Path('/work/dosgolem-opening')";assert code.count(marker)==1
        code=code.replace(marker,'out = Path('+repr(directory)+')');exec(compile(code,str(generator),'exec'),namespace)
except Built:pass
finally:subprocess.run=real_run
args=[v.replace('issue4-npc-move-state-normal-r1',pre) for v in meta['args']];args[0]=str(binary)
metadata=dict(runtime,original_size=len(original),original_sha256=meta['original_sha256'],upstream_revision=meta['upstream_revision'],docker_image=meta['docker_image'],scenario='controlled original automatic-turn component; not normal player receipt',warm_source_sha256='b2fdf7818b4d8ca07d97d5796db6ac241490a6899f8352ab2396cabb3b96c9c6',warm_seed='1357',component_seed='1e2c',component_seed_fixed_before_each_call=True,layers=[0,1,2,3],state_restore=True,emulator_snapshot_restore=False,cpu_register_reentry=True,gameplay_state_injection=True,normal_player_path=False,producer_sha256=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),probe_source_sha256=hashlib.sha256(frozen.read_bytes()).hexdigest(),probe_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),frozen_parent_go_sha256=hashlib.sha256(parent.read_bytes()).hexdigest(),generator_sha256=hashlib.sha256(generator.read_bytes()).hexdigest(),build_flags=meta['build_flags'],args=args)
(root/(pre+'-meta.json')).write_text(json.dumps(metadata,ensure_ascii=False,indent=2)+'\n')
with (root/(pre+'.log')).open('x') as stream:result=real_run(args,stdout=stream,stderr=subprocess.STDOUT,timeout=900)
assert result.returncode==0
assert any(line.startswith('DQ3_TURN_COMPONENT_DONE ') for line in (root/(pre+'.log')).open())
print('CONTROLLED_COMPONENT_SOURCE_DONE; requires independent checker; normal458 receipts unchanged.')
