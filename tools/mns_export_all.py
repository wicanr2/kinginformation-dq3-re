#!/usr/bin/env python3
"""把 DQ3MNS.SHP 全部怪物 sprite 各自存成 docs/monsters/spr_NNN.png(原尺寸 ×3 NEAREST)。
重用 tools/mns_sprite.py 的解碼。"""
import sys, os
sys.path.insert(0, 'tools')
import mns_sprite as M
from PIL import Image

m = open(M.SHP, 'rb').read()
offs = M.read_offsets(m)
pal = M.load_pal(M.PAL)
nspr = len(offs) - 1
os.makedirs('docs/monsters', exist_ok=True)
ok = 0
for idx in range(nspr):
    img = M.render_one(m, offs, pal, idx, plane_order="hi2lo")
    if img:
        img = img.resize((img.width * 3, img.height * 3), Image.NEAREST)
        img.save(f"docs/monsters/spr_{idx:03d}.png")
        ok += 1
print(f"saved {ok}/{nspr} individual sprites to docs/monsters/spr_NNN.png")
