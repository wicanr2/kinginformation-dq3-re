"""核對正常指令窗十二張完整RGB與未完成道具入口；入口 docs/188。"""
import argparse
import json
from pathlib import Path

from verify_dosgolem_mother_return import digest, png

SOURCE_HASH = '5b2a78c152e77f1ddbfd850af0cc129837bf206cd9a15731529927d21b12978a'
RUNTIME_HASH = '4293924f2b76a228dc80d91680ebcfcc989eefd765c589669e9fd07831276740'
PACK_HASH = 'sha256:66224bc04ff5f7d412640c986c35e0aa5f4eb7f49d4b1344b2df2a47d778a773'
PREFIX = 'issue4-command-navigation-r1'


def verify(args):
    assert digest(args.source_receipt) == SOURCE_HASH
    source = json.loads(args.source_receipt.read_text())
    assert source['seed'] == '1357' and source['seed_control_once'] and source['normal_inputs'] == 244
    assert not source['state_injection'] and not source['emulator_snapshot_restore']
    assert source['prefix194_unchanged'] and len(source['states']) == 206
    assert digest(args.original / (PREFIX+'.log')) == source['log_sha256']
    names = set()
    for artifact in source['artifacts']:
        name = artifact['path']
        relative = Path(name)
        assert not relative.is_absolute() and '..' not in relative.parts and name not in names
        names.add(name)
        path = args.original / relative
        assert path.resolve().is_relative_to(args.original.resolve())
        assert path.stat().st_size == artifact['size'] and digest(path) == artifact['sha256']
    assert len(names) == 426
    path = args.runtime / 'receipt.json'
    assert digest(path) == RUNTIME_HASH
    runtime = json.loads(path.read_text())
    assert runtime['source_sha256'] == SOURCE_HASH and runtime['pack_schema'] == '0.19.0' and runtime['pack_hash'] == PACK_HASH
    assert runtime['production_changed'] and runtime['snapshot_unchanged_before_save'] and runtime['rng_unchanged']
    assert runtime['normal_save_load_and_next_step'] and runtime['item_equipped_row_parity'] is False
    samples = runtime['samples']
    assert [s['packet'] for s in samples] == list(range(194,207))
    result = []
    for sample in samples:
        n = sample['packet']
        state = source['states'][n-1]
        name = f'{PREFIX}-packet-{n:03d}-{state["phase"]}.png'
        assert name in names and name.replace('.png','.bin') in names
        ow,oh,indices,original = png(args.original/name)
        assert indices == (args.original/name.replace('.png','.bin')).read_bytes()
        remake_path = args.runtime/f'packet-{n:03d}.png'
        rw,rh,_,remake = png(remake_path)
        assert (ow,oh,rw,rh) == (640,350,640,350)
        assert len(original) == len(remake) == 224000
        count = sum(a != b for a,b in zip(original,remake))
        expected = 44816 if n == 206 else 0
        assert count == expected == sample['full_rgb_difference']
        if n not in (202,206):
            assert sample['native_cursor'] == int(state['choice_cursor'])
        result.append({'packet':n,'full_rgb_difference':count,'remake_png_sha256':digest(remake_path),
                       'scope':'DRAFT Item entry' if n==206 else 'complete command canvas'})
    return {'source_sha256':SOURCE_HASH,'runtime_sha256':RUNTIME_HASH,'pack_hash':PACK_HASH,
            'twelve_complete_rgb_zero':True,'item_full_rgb_parity':False,'samples':result,
            'crop':False,'mask':False,'animation_override':False,'checker_sha256':digest(Path(__file__))}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--original',type=Path,required=True)
    parser.add_argument('--source-receipt',type=Path,required=True)
    parser.add_argument('--runtime',type=Path,required=True)
    parser.add_argument('--output',type=Path,required=True)
    args = parser.parse_args()
    assert not args.output.exists()
    result = verify(args)
    with args.output.open('x') as stream:
        json.dump(result,stream,ensure_ascii=False,indent=2)
        stream.write('\n')
    print('正常指令窗十二張完整RGB0；道具入口仍RED',digest(args.output))
