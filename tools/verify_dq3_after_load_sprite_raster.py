"""只讀核對 F6 後兩張完整畫面的 BLS 影格差異；入口 docs/188。

不修改影格，不產生替代圖，不遮罩差異，不證明動畫時鐘 parity。
"""
import argparse
import json
from pathlib import Path
import struct
import sys

from verify_dosgolem_mother_return import png
from verify_dq3_recruitment_sprite_raster import ASSETS, frame_pixel, plane_pixel, require, sha

SOURCE_HASH = '363f8f7002fb97254fe6bd8fcac8d4725aad07401dccfb39e667be0576d5dad3'
RUNTIME_HASH = 'f289c066b424a7886beb4d523a8adf815ea79df7d74bb1a52053f7c92b461d71'
PACK_HASH = 'sha256:9d6325addef6db4d6f4c049f44fe517e61d2a0550724c67d5ae7ec0f649a3917'
PREFIX = 'issue4-field-pose-normal-r1'


def verify(args):
    source_raw = args.source_receipt.read_bytes()
    require(sha(source_raw) == SOURCE_HASH, 'original source identity differs')
    source = json.loads(source_raw)
    require(source['normal_inputs'] == 240 and source['seed'] == '1357'
            and source['state_injection'] is False and source['emulator_snapshot_restore'] is False,
            'original input conditions differ')
    require(sha((args.original / (PREFIX + '.log')).read_bytes()) == source['log_sha256'],
            'original consumer log differs')
    runtime_raw = (args.runtime / 'receipt.json').read_bytes()
    require(sha(runtime_raw) == RUNTIME_HASH, 'runtime receipt identity differs')
    runtime = json.loads(runtime_raw)
    require(runtime['pack_schema'] == '0.18.0' and runtime['pack_hash'] == PACK_HASH
            and runtime['normal_final_packet'] == 202 and runtime['rng_unchanged'],
            'runtime conditions differ')
    raw_assets = {}
    for name, (size, digest) in ASSETS.items():
        raw = (args.assets / name).read_bytes()
        require(len(raw) == size and sha(raw) == digest, 'asset differs: ' + name)
        raw_assets[name] = raw
    cty, blk = raw_assets['CTY00.DAT'], raw_assets['DQ31.BLK']
    section = struct.unpack_from('<H', cty)[0]
    layout = section + struct.unpack_from('<H', cty, section + 14)[0]
    width, height = struct.unpack_from('<HH', cty, layout)
    require((section, width, height) == (12, 42, 43), 'map shape differs')
    require(struct.unpack_from('<HHH', blk) == (4, 24, 170), 'tile shape differs')
    artifacts = {a['path']: a for a in source['artifacts']}
    require(len(artifacts) == len(source['artifacts']), 'duplicate source artifact')
    results = []
    for number in (201, 202):
        samples = [s for s in runtime['samples'] if s['packet'] == number]
        require(len(samples) == 1, 'runtime sample absent or duplicated')
        sample = samples[0]
        state = source['states'][number - 1]
        px, py = int(state['player_x']), int(state['player_y'])
        original_path = args.original / f'{PREFIX}-packet-{number:03d}-ready.png'
        for path in (original_path, original_path.with_suffix('.bin')):
            artifact = artifacts[path.name]
            raw = path.read_bytes()
            require(len(raw) == artifact['size'] and sha(raw) == artifact['sha256'],
                    'source canvas differs')
        ow, oh, indices, original = png(original_path)
        require(indices == original_path.with_suffix('.bin').read_bytes(), 'PNG/bin differ')
        runtime_path = args.runtime / f'packet-{number:03d}.png'
        rw, rh, _, remake = png(runtime_path)
        require((ow, oh, rw, rh) == (640, 350, 640, 350), 'canvas shape differs')
        palette = {}
        for index, color in zip(indices, original):
            require(index not in palette or palette[index] == color, 'inconsistent palette')
            palette[index] = color
        blits = [b for b in source['blits'] if b['packet'] == str(number)]
        require(blits and all(b['SI'] == '0000' and b['DI'] == '0000' for b in blits),
                'single-hero consumer differs')
        boundary = source['boundaries'][number - 193]
        last = blits[-1]
        phase = int(last['raw0004'])
        require(boundary['packet'] == str(number) and int(boundary['raw0004']) == phase,
                'last blit/boundary phase differs')
        raw_direction = (0, 2, 1, 3)[sample['hero_facing']]
        hero_original_frame = int(last['BX'], 16) // 2
        require(hero_original_frame == raw_direction * 2 + phase, 'hero direction consumer differs')
        actors = [('hero', 'DQ3MST.BLS', raw_direction * 2 + sample['hero_walk'],
                   hero_original_frame, px, py, 288, 168, None)]
        for record, offset in ((14, 0x88), (15, 0x8f)):
            visible = [n for n in sample['npc_visuals'] if n['record'] == record]
            require(len(visible) == 1, 'diagnostic NPC absent or duplicated')
            npc, raw = visible[0], cty[offset:offset + 7]
            wx, wy = raw[:2]
            require((npc['x'], npc['y']) == (wx, wy), 'NPC position differs')
            require((0, 2, 1, 3)[npc['facing']] == raw[3] & 3, 'NPC direction differs')
            frame = ((raw[2] - 4) * 4 + (raw[3] & 3)) * 2
            actors.append((f'npc{record}', 'DQ3MAN.BLS', frame + npc['walk'], frame + phase,
                           wx, wy, (wx - px + 9) * 32, (wy - py + 7) * 24, offset))
        expected, actor_reports = {}, []
        for name, asset, old_frame, new_frame, wx, wy, sx, sy, offset in actors:
            require(0 <= wx < width and 0 <= wy < height, 'actor outside map')
            require(0 <= sx <= 608 and 0 <= sy <= 326, 'actor outside canvas')
            tile = cty[layout + 4 + 2 * (wy * width + wx)]
            require(tile < 170, 'background tile out of bounds')
            for y in range(24):
                for x in range(32):
                    background = plane_pixel(blk, 6 + tile * 384, x, y)
                    old = frame_pixel(raw_assets[asset], old_frame, x, y, background)
                    new = frame_pixel(raw_assets[asset], new_frame, x, y, background)
                    require(old in palette and new in palette, 'required palette entry absent')
                    pos = (sy + y) * 640 + sx + x
                    require(pos not in expected, 'overlapping diagnostic actors')
                    expected[pos] = (palette[old], palette[new], name)
                    require((remake[pos], original[pos]) == (palette[old], palette[new]),
                            'complete actor/background raster differs')
            actor_reports.append({'actor': name, 'asset': asset, 'runtime_raw_frame': old_frame,
                                  'original_raw_frame': new_frame, 'background_tile': tile,
                                  'cty_record_file': hex(offset) if offset is not None else None,
                                  'difference': 0})
        differences = [i for i, (old, new) in enumerate(zip(remake, original)) if old != new]
        counts = {a['actor']: a for a in actor_reports}
        for pos in differences:
            require(pos in expected, 'difference outside known actor raster')
            old, new, name = expected[pos]
            require((remake[pos], original[pos]) == (old, new), 'unexplained difference color')
            counts[name]['difference'] += 1
        require(len(differences) == sample['full_rgb_difference'], 'runtime difference metadata differs')
        results.append({'packet': number, 'full_rgb_difference': len(differences), 'actors': actor_reports,
                        'unexplained_difference': 0, 'original_consumer': last,
                        'original_png_sha256': sha(original_path.read_bytes()),
                        'runtime_png_sha256': sha(runtime_path.read_bytes())})
    return {'scope': 'read-only whole-canvas raw sprite diagnosis; no replacement image',
            'source_receipt_sha256': sha(source_raw), 'runtime_receipt_sha256': sha(runtime_raw),
            'tool_sha256': sha(Path(__file__).read_bytes()), 'python_version': sys.version,
            'assets': ASSETS, 'samples': results, 'raster_explanation': 'confirmed for these two captures',
            'animation_timing': 'unknown', 'production_changed': False, 'remake_parity': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--assets', type=Path, required=True)
    parser.add_argument('--original', type=Path, required=True)
    parser.add_argument('--source-receipt', type=Path, required=True)
    parser.add_argument('--runtime', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), 'refuse to overwrite existing receipt')
    result = verify(args)
    with args.output.open('x') as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2)
        stream.write('\n')
    print('完整畫布兩張人物影格差異已解釋；parity仍未通過')
