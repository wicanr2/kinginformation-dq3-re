#!/usr/bin/env python3
"""掃全 CTY 找 boss 級 sprite NPC(地圖固定一格、走過去按空白觸發的 boss)。

第一性原理:boss 不是會走動的 NPC,是地圖上固定一格放著 boss sprite 的物件;
玩家走過去 examine(空白/確認)就開戰。所以「boss 正式觸發點」== 某 CTY/section/座標
上 b2(sprite entry)是 boss sprite 的 NPC 格。

parser 對齊 dq3_scene.c::dq3_scene_load_npcs(與 find_npc_b4.py 同,已驗證):
  檔首無 count 前綴的 section 偏移表 word[i]=section i base;
  section so 處 u16 np(0xffff 退 +2),base=so+np,cnt=cty[base],其後 cnt×7B。
  NPC record 7B = x, y, b2(sprite entry), ctrl, b4, flags, b7;sub=(ctrl>>3)&7。

DQ3MAN.BLS 角色 sprite entry=(b2-4)*4(docs/27);但 boss 用怪物大 sprite,
b2 落在一般人物 entry 之外。用法:
  tools/find_boss_sprites.py [min_b2]      # 預設掃 b2>=60 的稀有 sprite
  tools/find_boss_sprites.py --b2 N        # 只列 b2==N
"""
import struct, sys, glob, os
from collections import Counter

min_b2 = 60
only_b2 = None
args = sys.argv[1:]
if '--b2' in args:
    only_b2 = int(args[args.index('--b2') + 1])
elif args and args[0].isdigit():
    min_b2 = int(args[0])

ASSET = 'assets_raw'
rows = []           # (cty_num, sec, x, y, b2, sub, b4)
b2_counter = Counter()

for fn in sorted(glob.glob(os.path.join(ASSET, 'CTY*.DAT'))):
    d = open(fn, 'rb').read()
    if len(d) < 4:
        continue
    u16 = lambda o: struct.unpack_from('<H', d, o)[0] if o + 1 < len(d) else 0xffff
    base_name = os.path.basename(fn)
    cty_num = int(''.join(c for c in base_name if c.isdigit()))
    # section 偏移表(無 count 前綴),掃到非法 word 為止
    offs = []
    for i in range(64):
        v = u16(2 * i)
        if v == 0xffff:
            offs.append(v); continue
        if v + 0x16 > len(d):
            break
        offs.append(v)
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
        if cnt > 40:                       # 防錯位讀到垃圾
            continue
        for i in range(cnt):
            r = base + 1 + i * 7
            if r + 7 > len(d):
                break
            x, y, b2, ctrl, b4, fl, b7 = d[r:r+7]
            # 過濾明顯錯位的座標(x,y 應落在地圖範圍內)
            if x > 63 or y > 63:
                continue
            sub = (ctrl >> 3) & 7
            b2_counter[b2] += 1
            rows.append((cty_num, si, x, y, b2, sub, b4))

# sprite entry 分布(看哪些是稀有大 sprite)
print('=== b2(sprite entry)分布 ===')
common = sorted(b2_counter.items())
print(' '.join('%d:%d' % (b, c) for b, c in common))
print()

if only_b2 is not None:
    sel = [r for r in rows if r[4] == only_b2]
    print('=== b2 == %d 的 NPC ===' % only_b2)
else:
    sel = [r for r in rows if r[4] >= min_b2]
    print('=== boss 候選(b2 >= %d 的稀有 sprite NPC)===' % min_b2)
sel.sort(key=lambda r: (-r[4], r[0], r[1]))
for cty, sec, x, y, b2, sub, b4 in sel:
    tag = ' [sub=2 examine觸發]' if sub == 2 else ''
    print('  CTY%-2d sec%d (%2d,%2d) b2=%3d sub=%d b4=%3d%s' % (cty, sec, x, y, b2, sub, b4, tag))
