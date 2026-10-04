"""唯讀核對可比十槽資料的 F5/F6 完整畫面；入口 docs/188。

保留全部畫面差異，不修改動畫相位，不宣稱完整存檔互通。
"""
import argparse
import json
from pathlib import Path
import struct
import sys

from verify_dosgolem_mother_return import png
from verify_dq3_recruitment_sprite_raster import ASSETS, frame_pixel, plane_pixel, require, sha

SOURCE_HASH = '363f8f7002fb97254fe6bd8fcac8d4725aad07401dccfb39e667be0576d5dad3'
RUNTIME_HASH = '58877b0475c4618b650dc2e5ddc0099958912df146f7808d8c3659bcb6863ae6'
PACK_HASH = 'sha256:9d6325addef6db4d6f4c049f44fe517e61d2a0550724c67d5ae7ec0f649a3917'
PREFIX = 'issue4-field-pose-normal-r1'


def verify(args):
    source_raw = args.source_receipt.read_bytes()
    require(sha(source_raw) == SOURCE_HASH, 'original source identity differs')
    source = json.loads(source_raw)
    require(source['seed'] == '1357' and source['normal_inputs'] == 240
            and source['state_injection'] is False and source['emulator_snapshot_restore'] is False,
            'original input conditions differ')
    names = set()
    for artifact in source['artifacts']:
        name = artifact['path']
        path = Path(name)
        # Accepted source checker stores native Scratch beside the output root.
        base = args.original.parent if path.parts[0] == PREFIX + '-scratch' else args.original
        require(not path.is_absolute() and '..' not in path.parts and name not in names
                and (base / path).resolve().is_relative_to(base.resolve()),
                'invalid source artifact path')
        names.add(name)
        raw = (base / path).read_bytes()
        require(len(raw) == artifact['size'] and sha(raw) == artifact['sha256'],
                'original artifact differs: ' + name)
    require(sha((args.original / (PREFIX + '.log')).read_bytes()) == source['log_sha256'],
            'original observer log differs')
    runtime_raw = (args.runtime / 'receipt.json').read_bytes()
    require(sha(runtime_raw) == RUNTIME_HASH, 'runtime receipt identity differs')
    runtime = json.loads(runtime_raw)
    require(runtime['pack_schema'] == '0.18.0' and runtime['pack_hash'] == PACK_HASH
            and runtime['normal_final_packet'] == 202 and runtime['rng_unchanged']
            and runtime['initial_slot_metadata_same_state'] is True
            and runtime['initial_storage_same_state'] is False,
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
    results = []
    for number in (195, 199):
        samples = [s for s in runtime['samples'] if s['packet'] == number]
        require(len(samples) == 1, 'runtime sample absent or duplicated')
        sample, state = samples[0], source['states'][number - 1]
        original_path = args.original / f'{PREFIX}-packet-{number:03d}-choice.png'
        require(original_path.name in names and original_path.with_suffix('.bin').name in names,
                'source canvas missing from manifest')
        ow, oh, indices, original = png(original_path)
        require(indices == original_path.with_suffix('.bin').read_bytes(), 'PNG/bin differ')
        runtime_path = args.runtime / f'packet-{number:03d}.png'
        rw, rh, _, remake = png(runtime_path)
        require((ow, oh, rw, rh) == (640, 350, 640, 350), 'canvas shape differs')
        differences = [i for i, (a, b) in enumerate(zip(remake, original)) if a != b]
        require(len(differences) == sample['full_rgb_difference'], 'difference metadata differs')
        if number == 195:
            require(not differences, 'F5 full canvas differs')
        else:
            palette = {}
            for index, color in zip(indices, original):
                require(index not in palette or palette[index] == color, 'inconsistent palette')
                palette[index] = color
            npcs = [n for n in sample['npc_visuals'] if n['record'] == 15]
            require(len(npcs) == 1, 'NPC15 missing or duplicated')
            npc, raw = npcs[0], cty[0x8f:0x96]
            wx, wy = raw[:2]
            require((npc['x'], npc['y']) == (wx, wy)
                    and (0, 2, 1, 3)[npc['facing']] == raw[3] & 3, 'NPC state differs')
            boundary = source['boundaries'][number - 193]
            require(boundary['packet'] == str(number), 'original boundary differs')
            phase = int(boundary['raw0004'])
            require(phase in (0, 1) and npc['walk'] in (0, 1), 'invalid phase')
            base_frame = ((raw[2] - 4) * 4 + (raw[3] & 3)) * 2
            sx, sy = (wx - int(state['player_x']) + 9) * 32, (wy - int(state['player_y']) + 7) * 24
            require(0 <= wx < width and 0 <= wy < height and 0 <= sx <= 608 and 0 <= sy <= 326,
                    'NPC outside map or canvas')
            tile = cty[layout + 4 + 2 * (wy * width + wx)]
            require(tile < 170, 'tile out of bounds')
            expected = set()
            for y in range(24):
                for x in range(32):
                    background = plane_pixel(blk, 6 + tile * 384, x, y)
                    a = frame_pixel(raw_assets['DQ3MAN.BLS'], base_frame + npc['walk'], x, y, background)
                    b = frame_pixel(raw_assets['DQ3MAN.BLS'], base_frame + phase, x, y, background)
                    require(a in palette and b in palette, 'palette entry absent')
                    pos = (sy + y) * 640 + sx + x
                    require((remake[pos], original[pos]) == (palette[a], palette[b]),
                            'complete NPC/background raster differs')
                    if a != b:
                        expected.add(pos)
            require(set(differences) == expected, 'difference outside NPC15 raster')
            require(len(differences) == 123, 'accepted diagnostic count differs')
        results.append({'packet': number, 'full_rgb_difference': len(differences),
                        'unexplained_difference': 0,
                        'original_png_sha256': sha(original_path.read_bytes()),
                        'runtime_png_sha256': sha(runtime_path.read_bytes())})
    return {'scope': 'read-only whole-canvas slot metadata comparison',
            'source_receipt_sha256': sha(source_raw), 'runtime_receipt_sha256': sha(runtime_raw),
            'tool_sha256': sha(Path(__file__).read_bytes()), 'python_version': sys.version,
            'samples': results, 'f5_full_rgb_parity': True, 'f6_full_rgb_parity': False,
            'f6_raster_explanation': 'confirmed for this capture only',
            'animation_timing': 'unknown', 'initial_storage_same_state': False,
            'production_changed': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ('assets', 'original', 'source-receipt', 'runtime', 'output'):
        parser.add_argument('--' + name, type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), 'refuse to overwrite existing receipt')
    result = verify(args)
    with args.output.open('x') as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2)
        stream.write('\n')
    print('F5完整RGB零差異；F6完整差123均由NPC15原始影格解釋，V3未通過')
