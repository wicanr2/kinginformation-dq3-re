"""核對單人道具取消完整畫布的原始人物圖塊；入口 docs/188。

只讀診斷，不改圖、不遮罩、不指定動畫相位；非零 RGB 仍未通過 V3。
"""
import argparse
import json
from pathlib import Path
import struct

from verify_dosgolem_mother_return import png
from verify_dq3_recruitment_sprite_raster import ASSETS, frame_pixel, plane_pixel, require, sha

SOURCE_HASH = '28995c8dd702f83d70c51f3f09212dc556356ed563d5b0b9d3977ff4b0d60eb4'
PACK_HASH = 'sha256:0839ecc939188bb2193786c0dae571345145a92de65f4179a4bbd3903ebbc68b'
PREFIX = 'issue4-field-item-navigation-r2'


def verify(args):
    source_raw = args.source_receipt.read_bytes()
    require(sha(source_raw) == SOURCE_HASH, 'original source identity differs')
    source = json.loads(source_raw)
    require(source['seed'] == '1357' and source['seed_control_once'] and source['normal_inputs'] == 256,
            'original test conditions differ')
    require(not source['state_injection'] and not source['emulator_snapshot_restore'], 'injected source')
    require(sha((args.original / (PREFIX + '.log')).read_bytes()) == source['log_sha256'], 'source log differs')
    names = set()
    for artifact in source['artifacts']:
        name = artifact['path']
        require(Path(name).name == name and name not in names, 'unsafe or duplicate original artifact')
        names.add(name)
        raw = (args.original / name).read_bytes()
        require(len(raw) == artifact['size'] and sha(raw) == artifact['sha256'], 'original artifact differs')
    require(len(names) == 462, 'source artifact shape differs')
    runtime_raw = (args.runtime / 'receipt.json').read_bytes()
    require(sha(runtime_raw) == args.runtime_receipt_sha256, 'runtime receipt identity differs')
    runtime = json.loads(runtime_raw)
    require(runtime['source_sha256'] == SOURCE_HASH and runtime['pack_schema'] == '0.20.0'
            and runtime['pack_hash'] == PACK_HASH, 'runtime pack/source differs')
    require(runtime['cancel_without_transaction'] and runtime['rng_unchanged']
            and runtime['normal_reopen_save_load_and_next_step'], 'runtime flow incomplete')
    require(runtime['item_storage_parity'] is False and runtime['selected_item_parity'] is False
            and runtime['cancel_full_rgb_parity'] is False, 'unsupported parity claim')
    assets = {}
    for name, (size, digest) in ASSETS.items():
        raw = (args.assets / name).read_bytes()
        require(len(raw) == size and sha(raw) == digest, 'original asset differs: ' + name)
        assets[name] = raw
    cty, blk = assets['CTY00.DAT'], assets['DQ31.BLK']
    section = struct.unpack_from('<H', cty)[0]
    layout = section + struct.unpack_from('<H', cty, section + 14)[0]
    width, height = struct.unpack_from('<HH', cty, layout)
    require((section, width, height) == (12, 42, 43), 'map shape differs')
    samples = runtime['samples']
    require([s['packet'] for s in samples] == list(range(194, 219)), 'runtime sample shape differs')
    results = []
    for sample in samples:
        number = sample['packet']
        state = source['states'][number - 1]
        original_path = args.original / f'{PREFIX}-packet-{number:03d}-{state["phase"]}.png'
        require(original_path.name in names and original_path.with_suffix('.bin').name in names,
                'original canvas absent')
        ow, oh, indices, original = png(original_path)
        require(indices == original_path.with_suffix('.bin').read_bytes(), 'original PNG/bin differ')
        runtime_path = args.runtime / f'packet-{number:03d}.png'
        encoded = runtime_path.read_bytes()
        require(len(encoded) == sample['png_size'] and sha(encoded) == sample['png_sha256'], 'runtime PNG differs')
        rw, rh, _, remake = png(runtime_path)
        require((ow, oh, rw, rh) == (640, 350, 640, 350), 'canvas shape differs')
        differences = {i for i, (a, b) in enumerate(zip(remake, original)) if a != b}
        require(len(differences) == sample['full_rgb_difference'], 'runtime difference metadata differs')
        result = {'packet': number, 'full_rgb_difference': len(differences),
                  'original_png_sha256': sha(original_path.read_bytes()), 'runtime_png_sha256': sha(encoded)}
        if number in (208, 217):
            require(sample['panel'] == 0 and state['phase'] == 'ready', 'cancelled field state differs')
            px, py = int(state['player_x']), int(state['player_y'])
            require((px, py) == (2, 18), 'diagnostic checkpoint differs')
            palette = {}
            for index, color in zip(indices, original):
                require(index not in palette or palette[index] == color, 'inconsistent palette')
                palette[index] = color
            direction = (0, 2, 1, 3)[sample['hero_facing']]
            actors = [('hero', 'DQ3MST.BLS', direction * 2, sample['hero_walk'], px, py, 288, 168)]
            for record, offset in ((14, 0x88), (15, 0x8f)):
                visible = [n for n in sample['npc_visuals'] if n['record'] == record]
                require(len(visible) == 1, 'runtime NPC absent or duplicated')
                npc, raw = visible[0], cty[offset:offset + 7]
                wx, wy = raw[:2]
                require((npc['x'], npc['y']) == (wx, wy), 'NPC position differs')
                require((0, 2, 1, 3)[npc['facing']] == raw[3] & 3, 'NPC direction differs')
                base = ((raw[2] - 4) * 4 + (raw[3] & 3)) * 2
                actors.append((f'npc{record}', 'DQ3MAN.BLS', base, npc['walk'], wx, wy,
                               (wx - px + 9) * 32, (wy - py + 7) * 24))
            explained, actor_reports = set(), []
            for name, asset, base, walk, wx, wy, sx, sy in actors:
                require(walk in (0, 1) and 0 <= wx < width and 0 <= wy < height, 'actor bounds differ')
                require(0 <= sx <= 608 and 0 <= sy <= 326, 'actor outside canvas')
                tile = cty[layout + 4 + 2 * (wy * width + wx)]
                require(tile < 170, 'background tile out of bounds')
                matches = []
                for phase in (0, 1):
                    pairs = []
                    for y in range(24):
                        for x in range(32):
                            background = plane_pixel(blk, 6 + tile * 384, x, y)
                            old = frame_pixel(assets[asset], base + walk, x, y, background)
                            new = frame_pixel(assets[asset], base + phase, x, y, background)
                            require(old in palette and new in palette, 'required palette entry absent')
                            pairs.append(((sy + y) * 640 + sx + x, palette[old], palette[new]))
                    if all((remake[pos], original[pos]) == (old, new) for pos, old, new in pairs):
                        delta = {pos for pos, old, new in pairs if old != new}
                        require(not explained & delta, 'overlapping diagnostic actors')
                        explained.update(delta)
                        matches.append({'original_frame': base + phase, 'difference': len(delta)})
                require(len(matches) == 1, 'complete actor/background raster not uniquely explained')
                actor_reports.append({'actor': name, 'asset': asset, 'runtime_frame': base + walk,
                                      'background_tile': tile, **matches[0]})
            require(differences == explained, 'difference outside complete original actor rasters')
            result.update(actors=actor_reports, unexplained_difference=0)
        results.append(result)
    return {'scope': 'read-only complete canvas and cancelled-field raw sprite diagnosis',
            'source_receipt_sha256': SOURCE_HASH, 'runtime_receipt_sha256': sha(runtime_raw),
            'checker_sha256': sha(Path(__file__).read_bytes()), 'assets': ASSETS, 'samples': results,
            'cancel_full_rgb_parity': False, 'animation_timing': 'unknown',
            'crop': False, 'mask': False, 'animation_override': False, 'replacement_image': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--assets', type=Path, required=True)
    parser.add_argument('--original', type=Path, required=True)
    parser.add_argument('--source-receipt', type=Path, required=True)
    parser.add_argument('--runtime', type=Path, required=True)
    parser.add_argument('--runtime-receipt-sha256', required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), 'refuse to overwrite receipt')
    result = verify(args)
    with args.output.open('x') as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2)
        stream.write('\n')
    print('完整取消畫布及原始人物圖塊核對；動畫時序與完整V3仍未知')
