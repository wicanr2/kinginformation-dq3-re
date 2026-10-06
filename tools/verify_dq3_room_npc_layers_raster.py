"""Docker-only: normal405..422 full-canvas and raw BLS/background diagnosis.

Read-only. No crop, mask, phase override or replacement image. Entry docs/188.
"""
import argparse,json,struct
from pathlib import Path
from verify_dosgolem_mother_return import png
from verify_dq3_recruitment_sprite_raster import ASSETS,frame_pixel,plane_pixel,require,sha
SOURCE_HASH='b7d0b1534a8b1055b57d78de8c343d12b736f355cf17584211b24dc2b84bcf3e'
PACK_HASH='sha256:a56f924086b32892dabf20d7812116096ef205518790d90e002a613498007385'
PREFIX='issue4-field-room-door-normal-r1'

def verify(args):
 source_raw=args.source_receipt.read_bytes();require(sha(source_raw)==SOURCE_HASH,'source identity differs');source=json.loads(source_raw)
 require(len(source['states'])==422 and len(source['artifacts'])==1076 and source['normal_inputs']==460 and source['seed_control_once'] and not source['state_injection'] and not source['emulator_snapshot_restore'],'source conditions differ')
 runtime_raw=(args.runtime/'room-door-receipt.json').read_bytes();require(sha(runtime_raw)==args.runtime_receipt_sha256,'runtime identity differs');runtime=json.loads(runtime_raw)
 require(runtime['source_sha256']==SOURCE_HASH and runtime['pack_hash']==PACK_HASH and runtime['pack_schema']=='0.32.0' and runtime['pack_content_version']=='0.1.105','runtime pack differs')
 require(runtime['rng_unchanged'] and runtime['normal_save_load_roundtrip'] and runtime['save_checkpoint_verified'] and runtime['move_after_load'] and runtime['save_version']==2 and not runtime['game_state_injection'] and not runtime['animation_timing_parity'],'runtime flow incomplete')
 require([s['packet'] for s in runtime['samples']]==list(range(405,423)),'runtime samples differ')
 assets={}
 for name,(size,digest) in ASSETS.items():
  b=(args.assets/name).read_bytes();require(len(b)==size and sha(b)==digest,'asset differs: '+name);assets[name]=b
 cty=assets['CTY00.DAT'];section=struct.unpack_from('<H',cty)[0];layout=section+struct.unpack_from('<H',cty,section+14)[0];width,height=struct.unpack_from('<HH',cty,layout)
 require((section,width,height)==(12,42,43),'map shape differs')
 exe=(args.assets/'DQ3.EXE').read_bytes();require(len(exe)==115282 and sha(exe)=='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c','EXE differs')
 window=json.loads(Path('/repo/dq3_remake_ebitan/internal/gamepack/packs/dq3_cht/data/interface.json').read_text())['field_talk']['presentation']['window']
 words=struct.unpack('<12H',exe[0x19fae:0x19fae+24]);require((window['x'],window['y'],window['width'],window['height'])==(words[1]*8,words[2],words[3]*8,words[4]),'native overlay bounds differ')
 artifacts={a['path']:a for a in source['artifacts']};require(len(artifacts)==1076,'duplicate artifact')
 results=[]
 for sample in runtime['samples']:
  n=sample['packet'];state=source['states'][n-1];px,py=int(state['player_x']),int(state['player_y'])
  op=args.original/f'{PREFIX}-packet-{n:03d}-{state["phase"]}.png'
  for path in (op,op.with_suffix('.bin')):
   a=artifacts[path.name];b=path.read_bytes();require(len(b)==a['size'] and sha(b)==a['sha256'],'original artifact differs')
  ow,oh,indices,original=png(op);require(indices==op.with_suffix('.bin').read_bytes(),'original PNG/bin differ')
  rp=args.runtime/f'room-door-packet-{n:03d}.png';require(sha(rp.read_bytes())==sample['png_sha256'],'runtime PNG differs');rw,rh,_,remake=png(rp)
  require((ow,oh,rw,rh)==(640,350,640,350),'canvas shape differs')
  # Use the source PNG PLTE, including entries unused in this capture.
  # png() has already checked all chunk CRCs; do not invent missing colors.
  raw=op.read_bytes();at=8;palette=None
  while at<len(raw):
   length=struct.unpack_from('>I',raw,at)[0];kind=raw[at+4:at+8]
   if kind==b'PLTE':
    value=raw[at+8:at+8+length];palette={i//3:tuple(value[i:i+3]) for i in range(0,len(value),3)}
   at+=length+12
  require(palette is not None,'source PLTE missing')
  require(all(palette[index]==color for index,color in zip(indices,original)),'PNG palette differs')
  require(sample['hero_facing']==(0 if n>=418 else (2 if n==408 or 411<=n<=414 else 3)),'hero direction differs')
  actors=[('hero','DQ3MST.BLS',(0,2,1,3)[sample['hero_facing']]*2,px,py,288,168)]
  for record,offset in ((14,0x88),(15,0x8f)):
   ns=[v for v in sample['npc_visuals'] if v['record']==record];require(len(ns)==1,'NPC missing/duplicate');npc=ns[0];raw=cty[offset:offset+7];wx,wy=raw[:2]
   require((npc['x'],npc['y'])==(wx,wy) and (0,2,1,3)[npc['facing']]==raw[3]&3,'NPC state differs')
   base=((raw[2]-4)*4+(raw[3]&3))*2;actors.append((f'npc{record}','DQ3MAN.BLS',base,wx,wy,(wx-px+9)*32,(wy-py+7)*24))
  explained=set();actor_reports=[]
  for name,asset,base,wx,wy,sx,sy in actors:
   require(0<=wx<width and 0<=wy<height and 0<=sx<=608 and 0<=sy<=326,'actor bounds differ');tile=cty[layout+4+2*(wy*width+wx)]
   if state['phase']=='waiting' and window['x']<=sx and window['y']<=sy and sx+32<=window['x']+window['width'] and sy+24<=window['y']+window['height']:
    # This entire actor is occluded. Verify all768 overlay pixels rather than
    # pretending to identify a hidden BLS frame; the full canvas stays compared.
    positions=[(sy+y)*640+sx+x for y in range(24) for x in range(32)]
    require(all(remake[pos]==original[pos] for pos in positions),'occluding native window differs')
    actor_reports.append(dict(actor=name,occluded_by_native_window=True,checked_pixels=len(positions),difference=0));continue
   tile_layer=(cty[layout+5+2*(wy*width+wx)]>>6)&3
   player_layer=(cty[layout+5+2*(py*width+px)]>>6)&3
   if name!='hero' and tile_layer!=player_layer:
    require(n==422,'unexpected hidden sample')
    roof=cty[section+0x15] if player_layer==0 else cty[section+0x16]
    positions=[]
    for y in range(24):
     for x in range(32):
      pos=(sy+y)*640+sx+x;index=plane_pixel(assets['DQ31.BLK'],6+roof*384,x,y)
      require(original[pos]==remake[pos]==palette[index],'complete hidden cell differs: '+name)
      positions.append(pos)
    actor_reports.append(dict(actor=name,hidden_by_cell_layer=True,checked_pixels=len(positions),difference=0,background_tile=roof));continue
   matches=[]
   # Identify each complete 32x24 raster from both original asset frames.
   # This does not alter either capture or assert a comparable animation clock.
   for old in (base,base+1):
    for new in (base,base+1):
     pairs=[]
     for y in range(24):
      for x in range(32):
       bg=plane_pixel(assets['DQ31.BLK'],6+tile*384,x,y);a=frame_pixel(assets[asset],old,x,y,bg);b=frame_pixel(assets[asset],new,x,y,bg)
       require(a in palette and b in palette,'palette entry absent');pairs.append(((sy+y)*640+sx+x,palette[a],palette[b]))
     if all((remake[pos],original[pos])==(a,b) for pos,a,b in pairs):matches.append((old,new,{pos for pos,a,b in pairs if a!=b}))
   require(len(matches)==1,'complete actor/background raster not unique: '+name)
   old,new,points=matches[0];require(not explained.intersection(points),'overlapping differences');explained.update(points)
   actor_reports.append(dict(actor=name,asset=asset,runtime_frame=old,original_frame=new,difference=len(points),background_tile=tile))
  actual={i for i,(a,b) in enumerate(zip(remake,original)) if a!=b}
  require(len(actual)==sample['full_rgb_difference'],'RGB metadata differs');require(explained.issubset(actual),'diagnosed pixels absent')
  if n<422:require(actual==explained,'unexplained full-canvas pixels remain')
  results.append(dict(packet=n,full_rgb_difference=len(actual),unexplained_difference=len(actual-explained),actors=actor_reports,full_rgb_parity=len(actual)==0))
 return dict(scope='normal405..422 full640x350, complete raw BLS/background differences; animation clock and moving exterior NPC positions unknown',source_sha256=SOURCE_HASH,runtime_sha256=sha(runtime_raw),pack_hash=PACK_HASH,results=results,animation_timing_parity=False,no_cropping_masking_or_image_replacement=True,checker_sha256=sha(Path(__file__).read_bytes()))

if __name__=='__main__':
 p=argparse.ArgumentParser(description=__doc__)
 for name in ('source_receipt','original','runtime','assets','output'):p.add_argument(name,type=Path)
 p.add_argument('runtime_receipt_sha256');args=p.parse_args();report=verify(args);require(not args.output.exists(),'output exists');args.output.write_text(json.dumps(report,ensure_ascii=False,indent=2)+'\n');print('RAW_RASTER_PASS',sha(args.output.read_bytes()),[(r['packet'],r['full_rgb_difference']) for r in report['results']])
