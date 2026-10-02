"""核對原版正常城堡接近收據；入口與限定範圍見 docs/188。

只在 Docker 執行。核對原先38次輸入與產物，新增方向鍵及唯讀觀測。
來源通過不代表重製畫面、謁見、音訊或完整主線通過。
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate(path):
    data = json.loads(path.read_text())
    assert data['scenario'] == 'king_approach'
    assert data['game_state_injection'] is False
    assert data['upstream_revision_observed'] == '2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    original = Path(data['original_path'])
    assert original.is_file() and original.stat().st_size == data['original_size'] == 115282
    assert digest(original) == data['original_sha256'] == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
    prior_path = path.with_name('issue4-mother-finish-receipt.json')
    prior = json.loads(prior_path.read_text())
    assert digest(prior_path) == '9358ce6e7a8e5c4547555daeccf74a2339f003ab106ebe4d57b32111602c0f3f'
    assert data['player_input'][:38] == prior['player_input']
    assert data['player_input'][38:] == [
        {'queued_step': 1960000000 + index*20000000, 'scan': '0x48'} for index in range(9)]
    assert len(data['actual_irq1_events']) == 94
    assert data['actual_irq1_events'][:76] == prior['actual_irq1_events']
    for index, event in enumerate(data['actual_irq1_events'][76:]):
        fields = dict(re.findall(r'(\w+)=(\S+)', event))
        assert int(fields['count']) == 77 + index
        assert fields['port60'] == ('48' if index % 2 == 0 else 'c8')
        assert int(fields['step']) == 1960000001 + (index // 2)*20000000 + (index % 2)*500000
    assert data['test_rng_seed_control'] == prior['test_rng_seed_control']
    assert data['creation_results'] == prior['creation_results']
    # 共用移動入口在新方向輸入後仍可記錄事件；先前事件必須逐項保留。
    assert data['mother_entry_events'][:len(prior['mother_entry_events'])] == prior['mother_entry_events']
    log_path = path.with_name('issue4-king-approach.log')
    lines = log_path.read_text().splitlines()
    assert [line for line in lines if line.startswith('DQ3_KEY_DELIVERED ')] == data['actual_irq1_events']
    assert '沒實作的服務（0 種）：' in lines
    assert not any('找不到的檔（' in line for line in lines)
    events = [line for line in lines if line.startswith('DQ3_KING_APPROACH ')]
    assert events == data['king_approach_events'] and len(events) == 9
    ready = [line for line in lines if line.startswith('DQ3_KING_READY ')]
    assert ready == data['king_ready_events'] and len(ready) == 1
    assert all(word in ready[0] for word in ['ordinal=9 ', 'ida_linear=2111b ',
               'player_x=15 player_y=30 ', 'return_cs=0110 return_ip=7e00 ', 'raw0b24=0006 '])
    assert [line for line in lines if line.startswith('DQ3_KING_CAMERA ')] == data['king_camera_events']
    layer_validated = False
    if 'layer_tile_events' in data:
        layer_events = [line for line in lines if line.startswith('DQ3_LAYER_TILE ')]
        assert layer_events == data['layer_tile_events'] and len(layer_events) >= 8
        town = original.with_name('CTY25.DAT')
        raw_town = town.read_bytes()
        assert len(raw_town) == 3756
        assert digest(town) == '11d5c60377c6a98bbb9cfc9532652c5e23f4e5c939e769397e8fa5b2f22f9e2b'
        u16 = lambda offset: int.from_bytes(raw_town[offset:offset+2], 'little')
        section = u16(0)
        layout = section + u16(section+14)
        width = u16(layout)
        assert raw_town[section+21:section+23] == bytes([27, 70])
        pending, completed = {}, set()
        for event in layer_events:
            fields = dict(re.findall(r'(\w+)=(\S+)', event))
            x, y = int(fields['x']), int(fields['y'])
            assert int(fields['ordinal']) == 9 and (x,y) in {(24,23),(25,23),(24,24),(25,24)}
            assert fields['layer'] == '00' and fields['raw0b56'] == '1b' and fields['raw0b57'] == '46'
            value = int(fields['raw_bx'],16)
            if fields['ida_linear'] == '11e07':
                assert value == u16(layout+4+(y*width+x)*2)
                pending[x,y] = value
            else:
                assert fields['ida_linear'] == '11e4b' and (x,y) in pending
                assert value == (pending.pop((x,y)) & 0xc000) | 27
                completed.add((x,y))
        assert not pending and completed == {(24,23),(25,23),(24,24),(25,24)}
        layer_validated = True
    observations = []
    for index, event in enumerate(events):
        fields = dict(re.findall(r'(\w+)=(\S+)', event))
        assert int(fields['step']) == 1971000000 + index*20000000
        assert fields['dgroup'] == '15ed'
        for name in ['player_x', 'player_y', 'raw4f25', 'raw4f27']:
            fields[name] = int(fields[name])
        observations.append(fields)
    artifacts = {entry['path']: entry for entry in data['artifacts']}
    assert len(artifacts) == len(data['artifacts']) == 198
    for name, entry in artifacts.items():
        assert Path(name).name == name
        artifact = path.parent / name
        assert artifact.is_file() and artifact.stat().st_size == entry['size'] > 0
        assert digest(artifact) == entry['sha256']
        assert artifact.stat().st_uid == os.getuid() and artifact.stat().st_gid == os.getgid()
    assert artifacts['issue4-king-approach-generation.py']['sha256'] == data['generation_script_sha256']
    identical = 0
    for entry in prior['artifacts']:
        if not entry['path'].endswith(('.png', '.bin')):
            continue
        name = entry['path'].replace('issue4-mother-finish-', 'issue4-king-approach-', 1)
        assert artifacts[name]['size'] == entry['size']
        assert artifacts[name]['sha256'] == entry['sha256']
        identical += 1
    for index in range(1, 10):
        for suffix in ['png', 'bin']:
            assert f'issue4-king-approach-castle-north-{index:02d}.{suffix}' in artifacts
    for suffix in ['png', 'bin']:
        assert f'issue4-king-approach-king-ready-09.{suffix}' in artifacts
    return {'source_validated': True, 'original_receipt_sha256': digest(path),
            'prior_receipt_sha256': digest(prior_path), 'inputs': 47, 'irq1_events': 94,
            'artifacts': len(artifacts), 'prior_identical_images_and_indices': identical,
            'observations': observations, 'game_state_injection': False,
            'idle_status_wait_events': ready, 'all_steps_are_complete_frames': False,
            'layer_tile_validated': layer_validated,
            'remake_parity': False, 'king_audience_parity': False, 'audio_parity': False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--receipt', type=Path, default=Path('/work/dosgolem-opening/issue4-king-approach-receipt.json'))
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    result = validate(args.receipt)
    assert args.output.parent.stat().st_uid == os.getuid()
    args.output.write_text(json.dumps(result, ensure_ascii=False, indent=2)+'\n')
    print(json.dumps(result, ensure_ascii=False))
