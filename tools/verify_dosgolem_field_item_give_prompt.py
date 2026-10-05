"""限定正常第225步原版來源；入口docs/188。"""
from pathlib import Path
import hashlib, json, sys

sys.path.insert(0, '/repo/tools')
from verify_dosgolem_mother_return import fields, digest

r = Path('/repo')
w = Path('/work')
root = w / 'dosgolem-opening'
prefix, old = 'issue4-give-prompt-observer-r1', 'issue4-field-item-reorder-r2'
parent_path = root / (old + '-source-r3-receipt.json')
assert digest(parent_path) == 'aad971bb8cbf08956384b6964b4a893251cb3b4ff316263953f7563a7e5b2fc4'
accepted = json.loads(parent_path.read_text())
assert digest(root / (old + '.log')) == accepted['log_sha256']
lines = (root / (prefix + '.log')).read_text().splitlines()
oldlines = (root / (old + '.log')).read_text().splitlines()
tags = ['DQ3_QUIESCENT_QUEUED ', 'DQ3_QUIESCENT_INPUT ', 'DQ3_QUIESCENT_CAPTURE ',
        'DQ3_KEY_DELIVERED ', 'DQ3_COMMAND_RASTER ', 'DQ3_ITEM_RASTER ',
        'DQ3_ITEM_REORDER_WRITER ', 'DQ3_EMPTY_RECRUIT_TEXT ', 'DQ3_SAVE_OBSERVED ']
counts = {}
for tag in tags:
    a = [line.replace(prefix, old) for line in lines if line.startswith(tag)]
    b = [line for line in oldlines if line.startswith(tag)]
    assert a == b and a, tag
    counts[tag.strip()] = len(a)
artifacts = []
for artifact in accepted['artifacts']:
    original = root / artifact['path']
    repeated = root / artifact['path'].replace(old, prefix)
    assert original.stat().st_size == repeated.stat().st_size == artifact['size']
    assert digest(original) == digest(repeated) == artifact['sha256'], repeated.name
    artifacts.append(dict(path=repeated.name, size=artifact['size'], sha256=artifact['sha256']))
meta = json.loads((root / (prefix + '-meta.json')).read_text())
assert digest(root / (prefix + '-probe-source.go')) == meta['probe_source_sha256']
assert digest(root / (prefix + '-probe')) == meta['probe_sha256']
assert meta['seed'] == '1357' and meta['seed_configured_before_execution']
assert not meta['state_restore'] and not meta['gameplay_state_injection']
assert meta['producer_sha256'] == digest(r / 'tools/dosgolem_field_item_give_prompt_probe.py')
for key in ['upstream_revision', 'original_size', 'original_sha256', 'upstream_files_sha256',
            'patched_files_sha256', 'upstream_bios_sha256', 'patched_bios_sha256',
            'upstream_vga_sha256', 'patched_vga_sha256', 'normal_prefix_inputs']:
    assert meta[key] == accepted['meta'][key], key
observations = [fields(line) for line in lines if line.startswith('DQ3_GIVE_PROMPT_OBSERVE ')]
glyphs = [row for row in observations if row['ida_linear'] == '214b9']
assert [(int(row['BX'], 16), int(row['BP'], 16)*8, int(row['DX'], 16)) for row in glyphs] == [
    (code, 168+i*24, 254) for i, code in enumerate([59, 60, 504, 559, 505, 58])]
exe = r / 'assets_raw/DQ3.EXE'
assert digest(exe) == meta['original_sha256'] and exe.stat().st_size == 115282
raw_window = exe.read_bytes()[0x19fae:0x19fae+30].hex()
assert all(row['window3e6e'] == raw_window for row in observations)
assert [row['ida_linear'] for row in observations] == [
    '15002', '1500f', '21414'] + ['214b9']*6 + ['139b6', '139c2', '2111b', '21133']
report = dict(scope='normal cold230 unchanged; read-only225 raw window/glyph/wait closure',
              parent_source_sha256=digest(parent_path), meta=meta, unchanged_event_counts=counts,
              artifacts=artifacts, observations=observations,
              source_go_sha256=digest(root/(prefix+'-probe-source.go')),
              log_sha256=digest(root/(prefix+'.log')),
              ida_sha256=digest(w/'issue4-give-prompt-r3-ida.json'),
              gameplay_state_injection=False, emulator_restore=False, full_campaign_parity=False)
import argparse
parser = argparse.ArgumentParser(description='核對第225步只讀原版觀測，保持230包與498份產物。')
parser.add_argument('--output', type=Path, required=True)
output = parser.parse_args().output
assert not output.exists()
output.write_text(json.dumps(report, ensure_ascii=False, indent=2)+'\n')
print('SOURCE_PASS', len(artifacts), 'artifacts', counts, 'SHA', digest(output), flush=True)
