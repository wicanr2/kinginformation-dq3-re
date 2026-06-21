#!/usr/bin/env python3
"""對齊正確率量測:用 D3TXT00.FON 字模當地面真值,檢查 run-based 抽取的 byte_off
是否落在真值錨點上。回報每段命中率與整體分數。純診斷,不產圖。"""
import struct
import font_align as A

d3 = open("assets_raw/D3TXT00.FON", "rb").read()
data, offs = A.load()

anchors = set()
tmpls = set()
for i in range(len(d3) // 32):
    g = d3[i * 32:(i + 1) * 32]
    if sum(bin(b).count('1') for b in g) < 24:
        continue
    tmpls.add(g)
for off in range(0, len(data) - 32):
    if data[off:off + 32] in tmpls:
        anchors.add(off)
print(f"地面真值錨點(D3TXT00 命中 CHINA): {len(anchors)}")

tot_g = 0
tot_hit = 0
covered = set()
for si in range(22):
    gl = A.extract_section(data, offs, si)
    hit = sum(1 for off, _ in gl if off in anchors)
    for off, _ in gl:
        if off in anchors:
            covered.add(off)
    tot_g += len(gl)
    tot_hit += hit
    # 該段錨點數
    a_in = sum(1 for a in anchors if offs[si] <= a < offs[si + 1])
    print(f"sec{si:2}: 抽出 {len(gl):4} 字, 命中錨點 {hit:3}/{a_in:3}")
print(f"\n合計抽出 {tot_g} 字, 落在錨點上 {tot_hit}")
print(f"錨點覆蓋率: {len(covered)}/{len(anchors)} = {len(covered)/len(anchors)*100:.1f}%")
