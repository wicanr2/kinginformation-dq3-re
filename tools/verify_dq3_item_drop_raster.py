"""Docker-only: normal packets241/250/260 full-canvas/raw-actor proof; entry docs/188.

No crop, mask, override or replacement image. Nonzero RGB does not prove V3.
"""
import argparse
import json
from pathlib import Path
import struct
from verify_dosgolem_mother_return import png
from verify_dq3_recruitment_sprite_raster import ASSETS, frame_pixel, plane_pixel, require, sha

SOURCE_HASH='5ca9fce6caff5bd1c1f27b489fba1ef9f687a808374d7fd544e94728d45f1971'
PACK_HASH='sha256:44ce09cbf64b97a4f40b797a10bcadf739347c3106298149c6b5409d5fc407c6'
PREFIX='issue4-item-drop-normal-r3'

def verify(args):
    source_raw=args.source_receipt.read_bytes()
    require(sha(source_raw)==SOURCE_HASH,'original source identity differs')
    source=json.loads(source_raw)
    require(source['seed']=='1357' and source['seed_control_once'] and not source['state_injection'] and not source['emulator_snapshot_restore'],'source conditions differ')
    names=set()
    for a in source['artifacts']:
        name=a['path']
        require(Path(name).name==name and name not in names,'unsafe or duplicate source artifact')
        names.add(name);b=(args.original/name).read_bytes()
        require(len(b)==a['size'] and sha(b)==a['sha256'],'source artifact differs')
    require(len(names)==591 and len(source['states'])==261,'source shape differs')
    runtime_raw=(args.runtime/'drop-receipt.json').read_bytes()
    require(sha(runtime_raw)==args.runtime_receipt_sha256,'runtime receipt identity differs')
    runtime=json.loads(runtime_raw)
    require(runtime['source_sha256']==SOURCE_HASH and runtime['pack_schema']=='0.24.0' and runtime['pack_content_version']=='0.1.96' and runtime['pack_hash']==PACK_HASH,'runtime pack/source differs')
    require(runtime['only_reviewed_slot_transaction'] and runtime['fresh_enter_returns_to_field'] and runtime['rng_unchanged'] and not runtime['game_state_injection'] and runtime['save_version']==2,'runtime flow incomplete')
    require(not runtime['animation_timing_parity'],'unsupported animation parity claim')
    require([s['packet'] for s in runtime['samples']]==list(range(231,262)),'sample shape differs')
    for s in runtime['samples']:
        b=(args.runtime/f'drop-packet-{s["packet"]:03d}.png').read_bytes()
        require(len(b)==s['png_size'] and sha(b)==s['png_sha256'],'runtime PNG differs')
    flow=json.loads((args.runtime/'item-ordered/receipt.json').read_text())
    require(flow['normal_reopen_save_load_and_next_step'] and flow['pack_hash']==PACK_HASH and flow['rng_unchanged'],'post-drop normal save/load incomplete')
    completed=json.loads((args.runtime.parent/'game-receipt.json').read_text())
    require(any(r['test']=='TestFieldItemDropDosgolemNormalInputComparison' and r['exit']==0 and r['skip']==0 for r in completed['results']),'normal drop test did not finish')
    results=[]
    for packet in (241,250,260):
        sample=next(s for s in runtime['samples'] if s['packet']==packet)
        state=source['states'][packet-1]
        path=args.original/f'{PREFIX}-packet-{packet:03d}-{state["phase"]}.png'
        require(path.name in names and path.with_suffix('.bin').name in names,'source canvas absent')
        ow,oh,indices,original=png(path)
        require(indices==path.with_suffix('.bin').read_bytes(),'source PNG/bin differ')
        rw,rh,_,remake=png(args.runtime/f'drop-packet-{packet:03d}.png')
        require((ow,oh,rw,rh)==(640,350,640,350),'canvas shape differs')
        difference={i for i,(a,b) in enumerate(zip(remake,original)) if a!=b}
        require(len(difference)==sample['full_rgb_difference'],'difference metadata differs')
        assets={}
        for name,(size,digest) in ASSETS.items():
            b=(args.assets/name).read_bytes();require(len(b)==size and sha(b)==digest,'original asset differs: '+name);assets[name]=b
        cty=assets['CTY00.DAT'];section=struct.unpack_from('<H',cty)[0]
        layout=section+struct.unpack_from('<H',cty,section+14)[0]
        width,height=struct.unpack_from('<HH',cty,layout)
        require((section,width,height)==(12,42,43),'map shape differs')
        visible=[n for n in sample['npc_visuals'] if n['record']==14]
        require(len(visible)==1,'NPC14 absent or duplicated')
        npc=visible[0];raw_npc=cty[0x88:0x8f];wx,wy=raw_npc[:2]
        require((npc['x'],npc['y'])==(wx,wy) and (0,2,1,3)[npc['facing']]==raw_npc[3]&3,'NPC14 state differs from CTY')
        px,py=int(state['player_x']),int(state['player_y'])
        sx,sy=(wx-px+9)*32,(wy-py+7)*24
        require((sx,sy)==(256,120) and npc['walk'] in (0,1),'NPC14 visual bounds differ')
        base=((raw_npc[2]-4)*4+(raw_npc[3]&3))*2
        tile=cty[layout+4+2*(wy*width+wx)]
        palette={}
        for index,color in zip(indices,original):
            require(index not in palette or palette[index]==color,'inconsistent palette');palette[index]=color
        matches=[]
        for phase in (0,1):
            pairs=[]
            for y in range(24):
                for x in range(32):
                    bg=plane_pixel(assets['DQ31.BLK'],6+tile*384,x,y)
                    old=frame_pixel(assets['DQ3MAN.BLS'],base+npc['walk'],x,y,bg)
                    new=frame_pixel(assets['DQ3MAN.BLS'],base+phase,x,y,bg)
                    require(old in palette and new in palette,'required palette entry absent')
                    pairs.append(((sy+y)*640+sx+x,palette[old],palette[new]))
            if all((remake[pos],original[pos])==(old,new) for pos,old,new in pairs):
                matches.append((base+phase,{pos for pos,old,new in pairs if old!=new}))
        require(len(matches)==1,'complete NPC14/background raster not uniquely explained')
        original_frame,explained=matches[0]
        actors=[dict(actor='npc14',asset='DQ3MAN.BLS',runtime_frame=base+npc['walk'],original_frame=original_frame,difference=len(explained),background_tile=tile)]
        # No movement follows230. Facing is recorded by the normal parent flow;
        # both captured hero phases must be uniquely identified by all768 pixels.
        facing=flow['samples'][-1]['hero_facing']
        require(flow['samples'][-1]['packet']==230 and facing==3 and (px,py)==(3,18),'hero initial state differs')
        hero_base=(0,2,1,3)[facing]*2
        hero_tile=cty[layout+4+2*(py*width+px)]
        hero_matches=[]
        for runtime_phase in (0,1):
            for original_phase in (0,1):
                pairs=[]
                for y in range(24):
                    for x in range(32):
                        bg=plane_pixel(assets['DQ31.BLK'],6+hero_tile*384,x,y)
                        old=frame_pixel(assets['DQ3MST.BLS'],hero_base+runtime_phase,x,y,bg)
                        new=frame_pixel(assets['DQ3MST.BLS'],hero_base+original_phase,x,y,bg)
                        require(old in palette and new in palette,'hero palette entry absent')
                        pairs.append(((168+y)*640+288+x,palette[old],palette[new]))
                if all((remake[pos],original[pos])==(old,new) for pos,old,new in pairs):
                    hero_matches.append((hero_base+runtime_phase,hero_base+original_phase,{pos for pos,old,new in pairs if old!=new}))
        require(len(hero_matches)==1,'complete hero/background raster not uniquely explained')
        hero_runtime,hero_original,hero_delta=hero_matches[0]
        require(not explained&hero_delta,'overlapping actor explanations')
        explained=explained|hero_delta
        actors.append(dict(actor='hero',asset='DQ3MST.BLS',runtime_frame=hero_runtime,original_frame=hero_original,difference=len(hero_delta),background_tile=hero_tile,frame_method='unique complete raw raster'))
        require(difference==explained,'difference outside complete original actor rasters')
        require(len(difference)=={241:122,250:106,260:122}[packet],'READY limited difference changed')
        results.append({'packet':packet,'scope':'production normal drop results: full canvas and complete raw actor rasters','source_receipt_sha256':SOURCE_HASH,'runtime_receipt_sha256':sha(runtime_raw),'checker_sha256':sha(Path(__file__).read_bytes()),'pack_hash':PACK_HASH,'runtime_png_sha256':sample['png_sha256'],'original_png_sha256':sha(path.read_bytes()),'full_rgb_difference':len(difference),'unexplained_difference':0,'actors':actors,'raw_npc_offset_file':'0x88','raw_npc_bytes':raw_npc.hex(),'full_rgb_parity':False,'animation_timing':'unknown','crop':False,'mask':False,'animation_override':False,'replacement_image':False,'production_implementation':True})
    return dict(samples=results,source_sha256=SOURCE_HASH,pack_hash=PACK_HASH,unexplained_difference=0,full_rgb_parity=False,animation_timing='unknown')


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    for name in ('assets','original','source-receipt','runtime','output'):p.add_argument('--'+name,type=Path,required=True)
    p.add_argument('--runtime-receipt-sha256',required=True);args=p.parse_args()
    require(not args.output.exists(),'refuse to overwrite receipt')
    result=verify(args)
    with args.output.open('x') as stream:json.dump(result,stream,ensure_ascii=False,indent=2);stream.write('\n')
    print('正式241／250／260全畫布逐像素核對，差異由完整原始人物圖塊解釋；完整V3與動畫時序仍未知')
