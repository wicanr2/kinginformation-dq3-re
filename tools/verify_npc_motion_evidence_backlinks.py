"""Docker-only；核對NPC自動移動原始位址及舊規格勘誤。入口docs/188。"""
import json,sys
from pathlib import Path
repo=Path(sys.argv[1]);evidence=(repo/'docs/188-opening-escort-to-castle-spec.md').read_text();older=(repo/'docs/35-script-format.md').read_text()
original='5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c'
source='b2fdf7818b4d8ca07d97d5796db6ac241490a6899f8352ab2396cabb3b96c9c6'
assert original in evidence and original in older and source in evidence and source in older
assert '2026-10-06 NPC自動移動勘誤' in older and '188-opening-escort-to-castle-spec.md#npc-motion-ready' in older
rows=[]
for address,semantic in [('12025','visible cell mover'),('1207F','signed cell word / quotient turn'),('120B0','RND10 step gate'),('1218D','whole AL attribute gate')]:
    assert address in evidence and address in older
    rows.append(dict(input='assets_raw/DQ3.EXE',size=115282,sha256=original,address_space='IDA9.4 linear',address=address,semantic=semantic,evidence_spec='docs/188-opening-escort-to-castle-spec.md#npc-motion-ready',older_spec='docs/35-script-format.md',required_correction='2026-10-06 NPC自動移動勘誤'))
print(json.dumps(rows,ensure_ascii=False,indent=2))
