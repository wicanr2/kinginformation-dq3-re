"""完整畫布與狀況首選單核對；入口 docs/188。所有剩餘差異保留。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_mother_return import digest, png

SOURCE_HASH='a30e50edd8ab3a3f8f8f95532b84aff4773d5a19879385ab31158e4d57b9b9aa'
PACK_HASH='sha256:694b740cd5baa579ee9382a5e87e9cae3cdae9f3b89895600009cfab54ad48e4'
PREFIX='issue4-field-status-normal-r3'

def verify(original, source_path, runtime, runtime_hash):
    assert digest(source_path)==SOURCE_HASH
    source=json.loads(source_path.read_text())
    assert source['seed_control_once'] and not source['state_injection'] and not source['emulator_snapshot_restore']
    assert len(source['artifacts'])==627 and len(source['states'])==273
    for a in source['artifacts']:
        assert Path(a['path']).name==a['path']
        p=original/a['path'];assert p.stat().st_size==a['size'] and digest(p)==a['sha256']
    receipt=runtime/'status-receipt.json'
    assert digest(receipt)==runtime_hash
    report=json.loads(receipt.read_text())
    assert report['source_sha256']==SOURCE_HASH and report['pack_hash']==PACK_HASH
    assert report['pack_schema']=='0.25.0' and report['pack_content_version']=='0.1.97' and report['save_version']==2
    assert report['persistent_status_unchanged'] and report['rng_unchanged'] and report['cancel_returns_to_field'] and not report['game_state_injection']
    assert report['animation_timing_parity'] is False
    assert [s['packet'] for s in report['samples']]==list(range(262,274))
    expected=[229,229,142,142,142,142,142,142,142,122,229,0]
    results=[]
    for sample, count in zip(report['samples'],expected):
        n=sample['packet'];state=source['states'][n-1]
        a=original/f'{PREFIX}-packet-{n:03d}-{state["phase"]}.png'
        b=runtime/f'status-packet-{n:03d}.png'
        assert b.stat().st_size==sample['png_size'] and digest(b)==sample['png_sha256']
        ow,oh,indices,old=png(a);rw,rh,_,new=png(b)
        assert (ow,oh,rw,rh)==(640,350,640,350) and indices==a.with_suffix('.bin').read_bytes()
        differences=[i for i,(x,y) in enumerate(zip(old,new)) if x!=y]
        assert len(differences)==count==sample['full_rgb_difference']
        menu_difference=None
        if 264<=n<=270:
            # Classify every full-canvas difference; do not alter either image.
            # Native opaque window is X120/Y62/W160/H80. Shadow extends beyond
            # this rectangle and retains underlying actor phase differences.
            menu_difference=sum(120<=i%640<280 and 62<=i//640<142 for i in differences)
            assert menu_difference==0
        bounds=None
        if differences:
            bounds=[min(i%640 for i in differences),min(i//640 for i in differences),max(i%640 for i in differences)+1,max(i//640 for i in differences)+1]
        results.append(dict(packet=n,full_rgb_difference=count,menu_frame_text_cursor_rgb_difference=menu_difference,remaining_difference=count,remaining_phase_explanation='unknown; no raw actor phase acceptance added by this checker',difference_bounds=bounds,original_png_sha256=digest(a),runtime_png_sha256=digest(b),full_rgb_parity=count==0))
    return dict(scope='正常狀況首選單、游標、取消及下一步；完整畫布差異全部保留',source_sha256=SOURCE_HASH,runtime_sha256=runtime_hash,pack_hash=PACK_HASH,checker_sha256=digest(Path(__file__)),samples=results,menu_frame_text_cursor_rgb_parity=True,all_full_canvases_identical=False,animation_timing='unknown',crop=False,mask=False,animation_override=False,replacement_image=False)

if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    for name in ('original','source','runtime','output'):p.add_argument('--'+name,type=Path,required=True)
    p.add_argument('--runtime-sha256',required=True);a=p.parse_args()
    assert not a.output.exists()
    report=verify(a.original,a.source,a.runtime,a.runtime_sha256)
    with a.output.open('x') as f:json.dump(report,f,ensure_ascii=False,indent=2);f.write('\n')
    print('RASTER_PASS',digest(a.output),'first menu RGB0; full canvases retain remaining differences')
