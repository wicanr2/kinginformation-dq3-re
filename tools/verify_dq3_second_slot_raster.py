"""唯讀核對正常第二槽十一張完整畫面；入口 docs/188。"""
import argparse
import json
from pathlib import Path
import struct
import sys

from verify_dosgolem_mother_return import png
from verify_dq3_recruitment_sprite_raster import ASSETS, frame_pixel, plane_pixel, require, sha

PREFIX = 'issue4-second-slot-r1'
SOURCE_HASH = '1a5a7c22c4599ce1bf4f6070d178a6d24ee49a8bbc4f374e84e39459a7a3f9d2'
RUNTIME_HASH = '4578ab4aa889af13446c95e2a980fb9bdf28cceb33415e5bd12c54b71296dd1d'
PACK_HASH = 'sha256:9d6325addef6db4d6f4c049f44fe517e61d2a0550724c67d5ae7ec0f649a3917'
COUNTS = {194: 0, 195: 0, 196: 0, 197: 0, 198: 411, 199: 356, 200: 123, 201: 123, 202: 0, 203: 0, 204: 229}


def verify(args):
    source_raw = args.source_receipt.read_bytes()
    require(sha(source_raw) == SOURCE_HASH, 'original source identity differs')
    source = json.loads(source_raw)
    require(source['normal_inputs'] == 242 and source['seed'] == '1357'
            and source['state_injection'] is False and source['emulator_snapshot_restore'] is False,
            'original input conditions differ')
    require(sha((args.original / (PREFIX + '.log')).read_bytes()) == source['log_sha256'], 'original log differs')
    names = set()
    for artifact in source['artifacts']:
        name = artifact['path']
        relative = Path(name)
        require(not relative.is_absolute() and '..' not in relative.parts and name not in names, 'invalid artifact path')
        base = args.original.parent if relative.parts[0] == PREFIX + '-scratch' else args.original
        require((base / name).resolve().is_relative_to(base.resolve()), 'artifact outside input')
        names.add(name)
        data = (base / name).read_bytes()
        require(len(data) == artifact['size'] and sha(data) == artifact['sha256'], 'original artifact differs')
    require(len(names) == 422, 'artifact count differs')
    runtime_raw = (args.runtime / 'receipt.json').read_bytes()
    require(sha(runtime_raw) == RUNTIME_HASH, 'runtime receipt identity differs')
    runtime = json.loads(runtime_raw)
    require(runtime['source_sha256'] == SOURCE_HASH and runtime['pack_schema'] == '0.18.0'
            and runtime['pack_hash'] == PACK_HASH and runtime['normal_final_packet'] == 204
            and runtime['rng_unchanged'] and runtime['selected_slot'] == 2
            and runtime['other_slots_unchanged_during_route'] and runtime['normal_save_load_and_next_step'],
            'runtime conditions differ')
    raw_assets = {}
    for name, (size, digest) in ASSETS.items():
        data = (args.assets / name).read_bytes()
        require(len(data) == size and sha(data) == digest, 'asset differs: ' + name)
        raw_assets[name] = data
    cty, blk = raw_assets['CTY00.DAT'], raw_assets['DQ31.BLK']
    section = struct.unpack_from('<H', cty)[0]
    layout = section + struct.unpack_from('<H', cty, section + 14)[0]
    width, height = struct.unpack_from('<HH', cty, layout)
    require((section, width, height) == (12, 42, 43)
            and struct.unpack_from('<HHH', blk) == (4, 24, 170), 'map shape differs')
    samples = []
    for number, count in COUNTS.items():
        matches = [s for s in runtime['samples'] if s['packet'] == number]
        require(len(matches) == 1, 'runtime sample missing or duplicated')
        sample, state = matches[0], source['states'][number - 1]
        original_path = args.original / f'{PREFIX}-packet-{number:03d}-{state["phase"]}.png'
        require(original_path.name in names and original_path.with_suffix('.bin').name in names, 'canvas outside manifest')
        ow, oh, indices, original = png(original_path)
        require(indices == original_path.with_suffix('.bin').read_bytes(), 'PNG/bin differ')
        runtime_path = args.runtime / f'packet-{number:03d}.png'
        rw, rh, _, remake = png(runtime_path)
        require((ow, oh, rw, rh) == (640, 350, 640, 350), 'canvas shape differs')
        differences = {i for i, (a, b) in enumerate(zip(remake, original)) if a != b}
        require(len(differences) == sample['full_rgb_difference'] == count, 'complete canvas difference differs')
        actors = []
        if count:
            palette = {}
            for index, color in zip(indices, original):
                require(index not in palette or palette[index] == color, 'inconsistent palette')
                palette[index] = color
            px, py = int(state['player_x']), int(state['player_y'])
            if number not in (200, 201):
                direction = (0, 2, 1, 3)[sample['hero_facing']]
                actors.append(('hero', 'DQ3MST.BLS', direction * 2 + sample['hero_walk'], direction * 2 + 1, px, py, 288, 168))
            for record, offset in ((14, 0x88), (15, 0x8f)):
                if number in (200, 201) and record == 14:
                    continue
                entries = [npc for npc in sample['npc_visuals'] if npc['record'] == record]
                require(len(entries) == 1, 'NPC absent or duplicated')
                npc, raw = entries[0], cty[offset:offset + 7]
                wx, wy = raw[:2]
                require((npc['x'], npc['y']) == (wx, wy) and (0, 2, 1, 3)[npc['facing']] == raw[3] & 3, 'NPC state differs')
                frame = ((raw[2] - 4) * 4 + (raw[3] & 3)) * 2
                actors.append((f'npc{record}', 'DQ3MAN.BLS', frame + npc['walk'], frame + 1, wx, wy, (wx - px + 9) * 32, (wy - py + 7) * 24))
            expected = set()
            for name, asset, old_frame, new_frame, wx, wy, sx, sy in actors:
                require(0 <= wx < width and 0 <= wy < height and 0 <= sx <= 608 and 0 <= sy <= 326, 'actor outside map or canvas')
                tile = cty[layout + 4 + 2 * (wy * width + wx)]
                require(tile < 170, 'tile outside archive')
                for y in range(24):
                    for x in range(32):
                        background = plane_pixel(blk, 6 + tile * 384, x, y)
                        a = frame_pixel(raw_assets[asset], old_frame, x, y, background)
                        b = frame_pixel(raw_assets[asset], new_frame, x, y, background)
                        require(a in palette and b in palette, 'palette entry missing')
                        pos = (sy + y) * 640 + sx + x
                        require((remake[pos], original[pos]) == (palette[a], palette[b]), 'complete actor/background raster differs')
                        if a != b:
                            expected.add(pos)
                # 完整圖塊包含透明像素；不只核對差異位置。
            require(differences == expected, 'difference outside complete known actors')
        samples.append({'packet': number, 'full_rgb_difference': len(differences), 'unexplained_difference': 0,
                        'actors': actors, 'original_png_sha256': sha(original_path.read_bytes()), 'runtime_png_sha256': sha(runtime_path.read_bytes())})
    return {'scope': 'read-only eleven full second-slot canvases; no image replacement',
            'source_sha256': SOURCE_HASH, 'runtime_sha256': RUNTIME_HASH, 'tool_sha256': sha(Path(__file__).read_bytes()),
            'python_version': sys.version, 'assets': ASSETS, 'samples': samples,
            'raster_explanation': 'confirmed for these captures only', 'animation_timing': 'unknown',
            'original_animation_counter': 'not observed in this source; raw bitmap matching only',
            'production_changed': False, 'full_rgb_parity': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    for name in ('assets', 'original', 'source-receipt', 'runtime', 'output'):
        parser.add_argument('--' + name, type=Path, required=True)
    args = parser.parse_args()
    require(not args.output.exists(), 'refuse to overwrite receipt')
    result = verify(args)
    with args.output.open('x') as stream:
        json.dump(result, stream, ensure_ascii=False, indent=2)
        stream.write('\n')
    print('十一張完整畫面核對，五張差異由原始人物圖塊解釋；動畫時鐘仍unknown')
