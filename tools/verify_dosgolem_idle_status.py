"""核對正常新遊戲進城後的原版閒置窗來源；限定範圍與入口見 docs/188。

只在 Docker 執行。來源通過不代表 remake 已實作生命週期或完整 RGB 通過。
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def fields(line):
    return dict(re.findall(r'(\w+)=(\S+)', line))


def validate(path):
    data = json.loads(path.read_text())
    assert data['scenario'] == 'king_idle' and data['game_state_injection'] is False
    assert data['upstream_revision_observed'] == '2f44a68ebfc54b28fb15dd4a34510b0b04a5415d'
    exe = Path(data['original_path'])
    assert exe.stat().st_size == data['original_size'] == 115282
    assert digest(exe) == data['original_sha256'] == '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
    prior_path = path.with_name('issue4-king-approach-receipt.json')
    assert digest(prior_path) == '16f0c0fc60f5f4aeb96e452eb0d0a6c25024c23a3c2662c90262a5722f50e918'
    prior = json.loads(prior_path.read_text())
    assert data['player_input'][:47] == prior['player_input']
    assert data['player_input'][47:] == [
        {'queued_step':2160000000,'scan':'0x48'},
        {'queued_step':2200000000,'scan':'0x48'}]
    assert data['test_rng_seed_control'] == prior['test_rng_seed_control']
    assert data['creation_results'] == prior['creation_results']
    assert data['actual_irq1_events'][:94] == prior['actual_irq1_events']
    assert len(data['actual_irq1_events']) == 98
    for index, event in enumerate(data['actual_irq1_events'][94:]):
        f = fields(event)
        assert int(f['count']) == 95+index
        assert f['port60'] == ('48' if index%2 == 0 else 'c8')
    for index, queued in enumerate([2160000000,2200000000]):
        make,release = [int(fields(x)['step']) for x in data['actual_irq1_events'][94+index*2:96+index*2]]
        assert queued < make < release < queued+1100000 and release-make >= 500000
    log = path.with_name('issue4-king-idle.log').read_text().splitlines()
    assert '沒實作的服務（0 種）：' in log
    assert not any('找不到的檔（' in line for line in log)
    for label, key in [('DQ3_KEY_DELIVERED ','actual_irq1_events'),
                       ('DQ3_KING_APPROACH ','king_approach_events'),
                       ('DQ3_LAYER_TILE ','layer_tile_events'),
                       ('DQ3_IDLE_STATUS ','idle_status_events'),
                       ('DQ3_IDLE_WINDOW ','idle_window_events')]:
        assert [x for x in log if x.startswith(label)] == data[key]
    assert data['king_approach_events'] == prior['king_approach_events']
    assert data['layer_tile_events'] == prior['layer_tile_events']
    events = [fields(x) for x in data['idle_status_events']]
    assert events
    openings = [x for x in events if x['ida_linear'] == '19952']
    assert len(openings) == 3
    assert all(int(x['delta']) == 300 and x['raw4f1f'] == '00ff' for x in openings)
    assert all(int(x['pit_divisor']) == 12428 for x in events)
    assert all((int(x['player_x']),int(x['player_y']),int(x['party_count'])) == (15,30,1) for x in events)
    assert all(int(a['step']) < int(b['step']) for a,b in zip(events,events[1:]))
    creation = fields(data['creation_results'][0])
    actor = bytes.fromhex(events[0]['actor'])
    assert len(actor) == 58 and all(bytes.fromhex(x['actor']) == actor for x in events)
    u16_actor = lambda n: int.from_bytes(actor[n:n+2],'little')
    assert actor[1] == int(creation['class_raw']) and actor[2] == int(creation['gender_raw'])
    assert actor[0x15] == int(creation['level'])
    for offset,key in [(0x16,'current_hp'),(0x18,'current_mp'),(0x2a,'max_hp'),(0x2c,'max_mp')]:
        assert u16_actor(offset) == int(creation[key])
    assert [u16_actor(n) for n in [3,5,7,0x38]] == [1,0xffff,0,0]
    boundaries = [x for x in events if x['ida_linear'] == '19940']
    assert len(boundaries) == 9
    for index,opening in enumerate(openings):
        triple = boundaries[index*3:index*3+3]
        assert [int(x['delta']) for x in triple] == [298,299,300]
        assert all(x['start000d'] == opening['start000d'] for x in triple)
        assert all(int(x['counter0000'])-int(x['start000d']) == int(x['delta']) for x in triple)
        assert triple[-1]['ticks'] == opening['ticks'] and triple[-1]['counter0000'] == opening['counter0000']
        assert int(triple[-1]['step']) < int(opening['step'])
        for a,b in zip(triple,triple[1:]):
            assert int(b['ticks'])-int(a['ticks']) == int(b['counter0000'])-int(a['counter0000']) == 1
    headers = [bytes.fromhex(x['window']) for x in events if x['ida_linear'] == '17de5']
    assert len(headers) == 3
    for header in headers:
        assert len(header) == 20
        u16 = lambda n: int.from_bytes(header[n:n+2],'little')
        assert [u16(n) for n in [2,4,6,8,10,12,14,16]] == [19,238,14,80,401,1,402,403]
    windows = [fields(x) for x in data['idle_window_events']]
    assert [x['phase'] for x in windows] == ['waiting','restore','waiting','restore','waiting']
    assert [x['ida_linear'] for x in windows] == ['2111b','17e11','2111b','17e11','2111b']
    assert all(int(x['ordinal']) == i+1 and int(x['party_count']) == 1 for i,x in enumerate(windows))
    assert all((int(x['player_x']),int(x['player_y'])) == (15,30) for x in windows)
    assert all(int(a['step']) < int(b['step']) for a,b in zip(windows,windows[1:]))
    for index,opening in enumerate(openings):
        waiting = windows[index*2]
        assert int(opening['step']) < int(waiting['step'])
        if index < 2:
            closing = windows[index*2+1]
            queued = data['player_input'][47+index]['queued_step']
            assert int(waiting['step']) < queued < int(closing['step']) < queued+1100000
    artifacts = {x['path']:x for x in data['artifacts']}
    assert len(artifacts) == len(data['artifacts'])
    assert len(artifacts) == 216
    for name,item in artifacts.items():
        assert Path(name).name == name
        artifact = path.parent/name
        assert artifact.is_file() and artifact.stat().st_size == item['size'] > 0
        assert digest(artifact) == item['sha256']
        assert artifact.stat().st_uid == os.getuid() and artifact.stat().st_gid == os.getgid()
    assert artifacts['issue4-king-idle-generation.py']['sha256'] == data['generation_script_sha256']
    unchanged = 0
    for item in prior['artifacts']:
        if not item['path'].endswith(('.png','.bin')):
            continue
        name = item['path'].replace('issue4-king-approach-','issue4-king-idle-',1)
        assert artifacts[name]['size'] == item['size'] and artifacts[name]['sha256'] == item['sha256']
        unchanged += 1
    assert unchanged == 196
    for window in windows:
        for suffix in ['png','bin']:
            name = 'issue4-king-idle-idle-'+window['phase']+'-'+window['ordinal'].zfill(2)+'.'+suffix
            assert name in artifacts
    for name in ['idle-key-1','idle-after-key-1','idle-key-2','idle-after-key-2']:
        for suffix in ['png','bin']:
            assert 'issue4-king-idle-'+name+'.'+suffix in artifacts
    return {'source_validated':True,'receipt_sha256':digest(path),'inputs':49,'irq1_events':98,
            'artifacts':len(artifacts),'prior_identical_images_and_indices':unchanged,
            'opening_events':openings,'window_events':windows,'game_state_injection':False,
            'timer_boundary_events':boundaries,'timer_scope':'three observed 298/299/300 boundaries; one counter increment per PIT tick',
            'actor_unchanged':True,'dismissal_movement':False,
            'remake_lifecycle_parity':False,'full_rgb_parity':False,
            'original_save_load_oracle':False,'audio_parity':False}


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('--receipt',type=Path,default=Path('/work/dosgolem-opening/issue4-king-idle-receipt.json'))
    parser.add_argument('--output',type=Path,required=True)
    args = parser.parse_args()
    result = validate(args.receipt)
    assert args.output.parent.stat().st_uid == os.getuid()
    args.output.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n')
    print(json.dumps(result,ensure_ascii=False))
