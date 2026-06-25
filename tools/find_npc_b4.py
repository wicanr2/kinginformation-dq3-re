#!/usr/bin/env python3
"""掃某 CTY 檔每個 section 的 NPC 清單,印出 (section, x, y, sub=(ctrl>>3)&7, b4)。
格式對齊 dq3_scene.c::dq3_scene_load_npcs:
  檔首 u16 n + n×u16 section 偏移;section so 處 u16 np(0xffff 退 +2),base=so+np,
  cnt=cty[base],其後 cnt×7B 記錄 = x,y,b2,ctrl,b4,flags,b7。
用法: tools/dockrun.sh tools/find_npc_b4.py CTY22.DAT [目標b4]
"""
import struct, sys
cty_name = sys.argv[1]
target = int(sys.argv[2]) if len(sys.argv) > 2 else None
d = open('assets_raw/%s' % cty_name, 'rb').read()
u16 = lambda o: struct.unpack_from('<H', d, o)[0]
# section 偏移表無 count 前綴:word[i]=section i 的 base(對齊 dq3_town.c::dq3_town_load)。
# 掃到第一個指向檔尾外/明顯非法的 word 為止(或上限 64)。
offs = []
for i in range(64):
    v = u16(2 * i)
    if v == 0xffff:
        offs.append(v); continue
    if v + 0x16 > len(d):
        break
    offs.append(v)
print('%s: %d sections' % (cty_name, len(offs)))
for si, so in enumerate(offs):
    if so == 0xffff or so + 4 > len(d):
        continue
    np = u16(so)
    if np == 0xffff:
        np = u16(so + 2)
    if np == 0xffff or so + np >= len(d):
        continue
    base = so + np
    cnt = d[base]
    for i in range(cnt):
        r = base + 1 + i * 7
        if r + 7 > len(d):
            break
        x, y, b2, ctrl, b4, fl, b7 = d[r:r+7]
        sub = (ctrl >> 3) & 7
        if target is None or b4 == target:
            print('  sect%d (%d,%d) sub=%d b4=%d(0x%02x)' % (si, x, y, sub, b4, b4))
