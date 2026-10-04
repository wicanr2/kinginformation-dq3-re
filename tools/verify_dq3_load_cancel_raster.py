"""唯讀核對正常 F6 取消的四張完整畫面；入口 docs/188。"""
import argparse
import json
from pathlib import Path
import struct
import sys

from verify_dosgolem_mother_return import png
from verify_dq3_recruitment_sprite_raster import ASSETS, frame_pixel, plane_pixel, require, sha

PREFIX = 'issue4-load-cancel-r1'
SOURCE_HASH = '426c7623d239af6f83fa9715de861ca932a9fa6bbc2a223bb7eac2f49aa484a7'
RUNTIME_HASH = '4ce86b0b73fae6a1439b315b98a4bda4f06ec71e5b3d3d79e00a12dd5b7c9bd9'
PACK_HASH = 'sha256:9d6325addef6db4d6f4c049f44fe517e61d2a0550724c67d5ae7ec0f649a3917'


def verify(args):
    source_raw = args.source_receipt.read_bytes()
    require(sha(source_raw) == SOURCE_HASH, 'original source identity differs')
    source = json.loads(source_raw)
    require(source['normal_inputs'] == 235 and source['seed'] == '1357'
            and source['state_injection'] is False and source['emulator_snapshot_restore'] is False,
            'original input conditions differ')
    require(sha((args.original / (PREFIX + '.log')).read_bytes()) == source['log_sha256'], 'original log differs')
    names = set()
    for artifact in source['artifacts']:
        name = artifact['path']
        require(Path(name).name == name and name not in names, 'invalid artifact path')
        names.add(name)
        data = (args.original / name).read_bytes()
        require(len(data) == artifact['size'] and sha(data) == artifact['sha256'], 'original artifact differs')
    require(len(names) == 399, 'artifact count differs')
    runtime_raw = (args.runtime / 'receipt.json').read_bytes()
    require(sha(runtime_raw) == RUNTIME_HASH, 'runtime receipt identity differs')
    runtime = json.loads(runtime_raw)
    require(runtime['source_sha256'] == SOURCE_HASH and runtime['pack_schema'] == '0.18.0'
            and runtime['pack_hash'] == PACK_HASH and runtime['normal_final_packet'] == 197
            and runtime['rng_unchanged'] and runtime['cancel_storage_unchanged']
            and runtime['normal_save_load_and_next_step'], 'runtime conditions differ')
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
    for number in range(194, 198):
        matches = [s for s in runtime['samples'] if s['packet'] == number]
        require(len(matches) == 1, 'runtime sample missing or duplicated')
        sample = matches[0]
        state = source['states'][number - 1]
        original_path = args.original / f'{PREFIX}-packet-{number:03d}-{state["phase"]}.png'
        require(original_path.name in names and original_path.with_suffix('.bin').name in names, 'canvas outside manifest')
        ow, oh, indices, original = png(original_path)
        require(indices == original_path.with_suffix('.bin').read_bytes(), 'PNG/bin differ')
        runtime_path = args.runtime / f'packet-{number:03d}.png'
        rw, rh, _, remake = png(runtime_path)
        require((ow, oh, rw, rh) == (640, 350, 640, 350), 'canvas shape differs')
        differences = {i for i, (a, b) in enumerate(zip(remake, original)) if a != b}
        require(len(differences) == sample['full_rgb_difference'], 'difference report differs')
        if number < 197:
            require(not differences, 'complete canvas differs')
        else:
            require(sample['hero_facing'] == 3 and sample['hero_walk'] == 0
                    and state['player_x'] == '2' and state['player_y'] == '18', 'hero state differs')
            palette = {}
            for index, color in zip(indices, original):
                require(index not in palette or palette[index] == color, 'inconsistent palette')
                palette[index] = color
            tile = cty[layout + 4 + 2 * (18 * width + 2)]
            require(tile < 170, 'tile outside archive')
            expected = set()
            # 固定已知右向6/7兩個完整原始圖塊，只核對像素，不選相位或修改圖片。
            for y in range(24):
                for x in range(32):
                    background = plane_pixel(blk, 6 + tile * 384, x, y)
                    a = frame_pixel(raw_assets['DQ3MST.BLS'], 6, x, y, background)
                    b = frame_pixel(raw_assets['DQ3MST.BLS'], 7, x, y, background)
                    require(a in palette and b in palette, 'palette entry missing')
                    pos = (168 + y) * 640 + 288 + x
                    require((remake[pos], original[pos]) == (palette[a], palette[b]), 'complete hero/background raster differs')
                    if a != b:
                        expected.add(pos)
            require(differences == expected and len(differences) == 122, 'difference outside complete hero raster')
        samples.append({'packet': number, 'full_rgb_difference': len(differences), 'unexplained_difference': 0,
                        'original_png_sha256': sha(original_path.read_bytes()), 'runtime_png_sha256': sha(runtime_path.read_bytes())})
    return {'scope': 'read-only four full canvases; raw frame6/7 diagnosis only',
            'source_sha256': SOURCE_HASH, 'runtime_sha256': RUNTIME_HASH, 'tool_sha256': sha(Path(__file__).read_bytes()),
            'python_version': sys.version, 'assets': ASSETS, 'samples': samples,
            'hero_raw_frames': [6, 7], 'hero_file_ranges': [[2886, 3366], [3366, 3846]],
            'raster_explanation': 'confirmed for these captures only', 'animation_timing': 'unknown',
            'original_animation_counter': 'not observed in this source; no counter claim',
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
    print('四張完整畫面核對；右移122像素為完整原始英雄6/7圖塊差異，時鐘仍unknown')
