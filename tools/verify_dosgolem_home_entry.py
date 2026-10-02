#!/usr/bin/env python3
"""在 Docker 內驗證原版家中開場收據；不宣稱 remake 對拍通過。

入口與證據邊界見 docs/188-opening-escort-to-castle-spec.md。
"""
import argparse
import hashlib
import json
from pathlib import Path
import struct


def require(condition, message):
    if not condition:
        raise ValueError(message)


def fields(line):
    return dict(part.split('=', 1) for part in line.split()[1:] if '=' in part)


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def verify(receipt_path, assets):
    raw = receipt_path.read_bytes()
    data = json.loads(raw)
    folder = receipt_path.parent
    executable = (assets / 'DQ3.EXE').read_bytes()
    require(len(executable) == data['original_size'] == 115282, '原版 EXE 大小不符')
    require(sha(executable) == data['original_sha256'] ==
            '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c', '原版 EXE 身分不符')
    require(data['scenario'] in ('mother_home_entry', 'mother_home_contract', 'mother_home_navigation'), '情境不符')
    navigation = data['scenario'] == 'mother_home_navigation'
    require(data['upstream_revision_observed'] == '2f44a68ebfc54b28fb15dd4a34510b0b04a5415d', 'dosgolem 版本未審查')
    require(not data['game_state_injection'], '禁止遊戲狀態注入')
    seed = data['test_rng_seed_control']
    require(seed['seed'] == '0x1357' and seed['configured_before_execution'] and
            seed['only_rng_state_modified'] and not seed['other_gameplay_state_injection'], 'seed 條件不符')
    inputs, events = data['player_input'], data['actual_irq1_events']
    expected_count = 42 if navigation else 37
    require(len(inputs) == expected_count and len(events) == expected_count * 2, '正常輸入或 IRQ1 數量不符')
    modal_input = [0x1c, 0x01, 0x4b, 0x48, 0x4d, 0x50] + [0x1c] * 4 if navigation else [0x1c] * 5
    expected_tail = modal_input + [0x50] * 2 + [0x4b] * 2 + [0x50] * 3 + [0x4d] * 6
    require([int(i['scan'], 16) for i in inputs[19:]] == expected_tail, '選圖及接近輸入不符')
    for index, item in enumerate(inputs):
        make, release = map(fields, events[index * 2:index * 2 + 2])
        scan = int(item['scan'], 16)
        require(int(make['port60'], 16) == scan and int(release['port60'], 16) == scan | 0x80,
                '按下／放開 scan 不符')
        require(item['queued_step'] < int(make['step']) < int(release['step']) and
                int(release['step']) - int(make['step']) >= data['minimum_scan_interval'], 'IRQ1 順序不符')
    artifact_names = [item['path'] for item in data['artifacts']]
    require(len(artifact_names) == len(set(artifact_names)), '產物重複列入收據')
    for item in data['artifacts']:
        name = item['path']
        require(Path(name).name == name, '產物路徑不得離開收據目錄')
        artifact = folder / name
        content = artifact.read_bytes()
        require(len(content) == item['size'] and sha(content) == item['sha256'], '產物內容不符：' + name)
        require((artifact.stat().st_uid, artifact.stat().st_gid) == (1000, 1000), '產物擁有權不符：' + name)
    prefix = receipt_path.name.removesuffix('-receipt.json')
    generation = (folder / (prefix + '-generation.py')).read_bytes()
    require(sha(generation) == data['generation_script_sha256'], '實際生成腳本身分不符')
    log = (folder / (prefix + '.log')).read_text()
    require(log.count('DQ3_CREATION_SEED ') == 1, 'seed 未固定一次')
    for name in artifact_names:
        if name.endswith('.png'):
            require(name in log, '不是這次重生的 PNG：' + name)

    regions = {}
    for line in data['home_modal_data']:
        f = fields(line)
        offset, content = int(f['DGROUP'], 16), bytes.fromhex(f['raw'])
        require(offset not in regions, '原始資料區域重複')
        # 4348視窗與00DB檔名是EXE靜態資料；09F1選項與2B7A偏移表由runtime寫入。
        if offset in (0x4348, 0x00db):
            require(content == executable[0x16140 + offset:0x16140 + offset + len(content)],
                    'runtime／EXE 靜態區域不符')
        regions[offset] = content
    require(set(regions) == {0x09f1, 0x4348, 0x2b7a, 0x00db}, '選圖資料區域缺失')
    filename = regions[0x00db].split(b'\0', 1)[0].decode('ascii').upper()
    require(filename == 'DQ3LIN.BLS', '圖像來源不符')
    sprite = (assets / filename).read_bytes()
    width, height, count = struct.unpack('<3H', sprite[:6])
    require((width, height, count) == (4, 24, 96) and len(sprite) == 6 + count * 480, 'BLS 實際形狀不符')
    options = list(regions[0x09f1][1:])
    offsets = struct.unpack('<66H', regions[0x2b7a])
    require(offsets == tuple(i * 480 for i in range(66)), 'runtime Sprite 偏移表不符')
    require(len(options) == 18 and all(i < count and i < len(offsets) and offsets[i] == i * 480
                                    for i in options), '選項素材索引越界或檔案偏移不符')

    sequence = [fields(line) for line in data['npc_sequence_events']]
    moves = [f for f in sequence if f['ida_linear'] == '122cd' and 1 <= int(f['ordinal']) <= 16]
    waits = [f for f in sequence if f['ida_linear'] == '16822' and 1 <= int(f['ordinal']) <= 16]
    require(len(moves) == len(waits) == 16, '家中16動作缺失')
    raw_sequence = bytes.fromhex('020100060000070300010102ff')
    require(executable[0x19dfc:0x19dfc + len(raw_sequence)] == raw_sequence, '原始 NPC 序列 bytes 不符')
    x, y = 5, 4
    actions = []
    for repeat, direction, mode in zip(raw_sequence[0:-1:3], raw_sequence[1:-1:3], raw_sequence[2:-1:3]):
        for _ in range(repeat):
            if mode != 2:
                dx, dy = ((0, 1), (-1, 0), (0, -1), (1, 0))[direction]
                x, y = x + dx, y + dy
            actions.append((x, y, direction))
    for index, (move, wait, expected) in enumerate(zip(moves, waits, actions), 1):
        actor = bytes.fromhex(move['npc0'])
        require((actor[0], actor[1], actor[3] & 3) == expected, '母親動作不符')
        require(int(move['ordinal']) == int(wait['ordinal']) == index and
                int(wait['ticks']) - int(move['ticks']) == 8, '動作順序或來源 tick 不符')
        require(move['player_x'] == move['player_y'] == wait['player_x'] == wait['player_y'] == '5',
                '原版主角不應隨母親移動')

    modal = [fields(line) for line in data['home_modal_events']]
    rounds = [f for f in modal if f['ida_linear'] == '21f79']
    require([int(f['raw26fe']) for f in rounds] == [0, 1, 2], '三輪圖像選擇不符')
    result = [f for f in modal if f['ida_linear'] == '21e3e']
    returned = [f for f in modal if f['ida_linear'] == '21e94']
    require(len(result) == len(returned) == 1 and result[0]['AX'] == returned[0]['AX'] == '0001' and
            result[0]['choices'] == returned[0]['choices'] == '010101', '選擇交易或結果確認不符')
    ack_index = 27 if navigation else 23
    require(int(result[0]['step']) < inputs[ack_index]['queued_step'] < int(returned[0]['step']), '結果確認未經正式輸入')
    if data['scenario'] != 'mother_home_entry':
        initialization = [fields(line) for line in data['picture_init_events']]
        require([f['ida_linear'] for f in initialization] == ['16f4b', '16f56', '16f65', '16fce'], '獨立圖像初始化缺失')
        require(initialization[1]['AX'] == initialization[1]['bios046c'] == '151b', '自然BIOS時鐘來源不符')
        table = executable[0x16a00:0x16a00 + 300]
        require(len(table) == 300 and max(table) < count, '原始100組問題資料無效')
        random_state = int(initialization[1]['AX'], 16)
        def draw():
            nonlocal random_state
            random_state = (random_state + 0x9014) & 0xffff
            random_state = ((random_state << 3) | (random_state >> 13)) & 0xffff
            return random_state
        def sample(limit):
            for _ in range(65536):
                value = draw() & 0x7f
                if value < limit:
                    return value
            raise ValueError('原始圖像PRNG未能取得有效值')
        question = sample(100)
        generated = []
        for target in table[question * 3:question * 3 + 3]:
            random_first = draw() & 1
            alternate = sample(count)
            bases = [target // 3 * 3, alternate // 3 * 3]
            if random_first:
                bases.reverse()
            generated += [base + i for base in bases for i in range(3)]
        require(question == int(initialization[3]['challenge']) == regions[0x09f1][0] and
                bytes(generated).hex() == initialization[3]['options'] and generated == options,
                '自然時鐘初始化、EXE表格、runtime選項不符')
    if navigation:
        escaped = [f for f in modal if f['ida_linear'] == '21fd2' and f['AX'] == 'ffff']
        require(len(escaped) == 1 and int(escaped[0]['raw26fe']) == 0 and
                int(escaped[0]['raw0726']) & 0xff == 1, 'Escape原始返回狀態不符')
        require(rounds[1]['choices'] == '010000' and int(rounds[1]['step']) > int(escaped[0]['step']),
                '此EXE的Escape應選定當前選項並進下一輪')
        cursors = [int(f['raw0722']) for f in modal if f['ida_linear'] == '1f908' and
                   inputs[21]['queued_step'] < int(f['step']) < inputs[25]['queued_step']]
        require(cursors == [1, 6, 6, 5, 5, 6, 6, 1], '四方向環繞游標不符')
    entry = [fields(line) for line in data['mother_entry_events']]
    chain = []
    for address in ('1010b', '10121', '10130'):
        rows = [f for f in entry if f['ida_linear'] == address]
        require(len(rows) == 1, '正式入口鏈缺失：' + address)
        chain.append(rows[0])
    require(int(returned[0]['step']) < inputs[-1]['queued_step'] < int(chain[0]['step']) <
            int(chain[1]['step']) < int(chain[2]['step']), '手動接近與轉場順序不符')
    require((chain[0]['player_x'], chain[0]['player_y']) == ('9', '10') and
            (chain[1]['player_x'], chain[1]['player_y']) == ('9', '10') and
            bytes.fromhex(chain[1]['npc0'])[:2] == bytes((10, 11)) and
            (chain[2]['player_x'], chain[2]['player_y']) == ('8', '38') and
            bytes.fromhex(chain[2]['npc0'])[:2] == bytes((8, 37)), '家中末步或正常轉場落點不符')
    return {'kind': '原版家中開場自然輸入證據核對', 'receipt_sha256': sha(raw),
            'original_inputs': len(inputs), 'irq1_events': len(events), 'unique_artifacts_verified': len(artifact_names),
            'mother_actions_verified': len(actions), 'picture_rounds_verified': len(rounds),
            'picture_asset': filename, 'picture_asset_size': len(sprite), 'picture_asset_sha256': sha(sprite),
            'picture_asset_count': count, 'picture_option_raw_indices': options,
            'manual_approach_entry_verified': True,
            'remake_parity': False, 'full_rgb_parity': False, 'audio_verified': False,
            'timing_scope': '來源動作 tick 閾值；不宣稱 DOS 硬體 wall-clock'}


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--receipt', type=Path, default=Path('/work/dosgolem-opening/issue4-home-entry-receipt.json'))
    parser.add_argument('--assets', type=Path, default=Path('/repo/assets_raw'))
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    report = verify(args.receipt, args.assets)
    text = json.dumps(report, ensure_ascii=False, indent=2) + '\n'
    if args.output:
        args.output.write_text(text)
    print(text, end='')
