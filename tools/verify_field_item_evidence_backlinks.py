"""核對正常物品證據與舊規格勘誤回填；入口 docs/188。"""
import argparse
import json
from pathlib import Path

INPUT_HASH = '5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
EVIDENCE = 'docs/188-opening-escort-to-castle-spec.md'
OLDER = 'docs/158-field-item-owner-selection-drop-spec.md'
LEDGER = (
    ('1372F', '單人取消', '2026-10-05 勘誤：單人取消與物品格順序',
     '28995c8dd702f83d70c51f3f09212dc556356ed563d5b0b9d3977ff4b0d60eb4'),
    ('13A62', '單人穿戴物給自己重排', '初始第一件穿戴物給予自己的交易升為限定confirmed',
     'aad971bb8cbf08956384b6964b4a893251cb3b4ff316263953f7563a7e5b2fc4'),
)


def validate(repo):
    evidence = (repo / EVIDENCE).read_text()
    older = (repo / OLDER).read_text()
    assert INPUT_HASH in evidence and INPUT_HASH in older, 'original EXE identity'
    assert 'IDA linear' in evidence and 'IDA linear' in older, 'address space'
    assert '188-opening-escort-to-castle-spec.md' in older, 'older spec backlink'
    rows = []
    for address, semantic, marker, source in LEDGER:
        assert address in evidence and address in older, address
        assert marker in older and source in evidence and source in older, semantic
        rows.append(dict(input='assets_raw/DQ3.EXE', size=115282, sha256=INPUT_HASH,
                         address_space='IDA linear', address=address, semantic=semantic,
                         evidence_spec=EVIDENCE, older_spec=OLDER, required_correction=marker))
    return rows


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('repo', type=Path)
    args = parser.parse_args()
    print(json.dumps(validate(args.repo), ensure_ascii=False, indent=2))
