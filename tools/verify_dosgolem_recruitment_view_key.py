"""嚴格核對詳細頁第二次等待的正常 A 鍵來源；入口 docs/188。"""
import argparse
import json
from pathlib import Path
from verify_dosgolem_character_spells import validate
from verify_dosgolem_mother_return import digest, fields


def verify(root, prefix, producer):
    report = validate(root, prefix, producer, full=True, close_scan='1e')
    prior_path = root / 'issue4-view-spells-class3-normal-r2-source-r2-receipt.json'
    assert digest(prior_path) == '4dcb99c80e5cddc17940b09772c37169d5348cc40249996827029812365b8c08'
    prior = json.loads(prior_path.read_text())
    for key in ('queued', 'consumed', 'states'):
        assert report[key][:205] == prior[key][:205]
    assert report['actual_irq1_events'][:486] == prior['actual_irq1_events'][:486]
    for n, state in enumerate(report['states'][:205], 1):
        suffix = f'-packet-{n:03d}-{state["phase"]}'
        for ext in ('.png', '.bin'):
            assert (root / (prefix + suffix + ext)).read_bytes() == (root / ('issue4-view-spells-class3-normal-r2' + suffix + ext)).read_bytes()
    events = [fields(s) for s in (root / (prefix + '.log')).read_text().splitlines() if s.startswith('DQ3_VIEW_CLOSE_KEY ')]
    assert [s['ida_linear'] for s in events] == ['10686', '10689', '1068e']
    assert all(s['packet'] == '206' and s['DS'] == '15ed' and s['AX'] == '1e00' for s in events)
    assert report['consumed'][205]['ida_linear'] == '2110b'
    report.update(scope='正常第三職業男性209包；第二次等待A鍵關閉，540→No→541→場景', close_key_events=events, prefix205_unchanged=True, close_scan='1e', checker_sha256=digest(Path(__file__)), base_checker_sha256=digest(Path('/repo/tools/verify_dosgolem_character_spells.py')))
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('root', type=Path)
    parser.add_argument('prefix')
    parser.add_argument('producer', type=Path)
    parser.add_argument('receipt', type=Path)
    args = parser.parse_args()
    assert not args.receipt.exists(), 'refuse to overwrite receipt'
    report = verify(args.root, args.prefix, args.producer)
    with args.receipt.open('x') as stream:
        json.dump(report, stream, ensure_ascii=False, indent=2)
        stream.write('\n')
    print('正常A鍵來源接受', len(report['states']), digest(args.receipt))


if __name__ == '__main__':
    main()
