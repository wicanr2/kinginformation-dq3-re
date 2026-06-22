#!/usr/bin/env python3
"""從 oracle 截圖裁出主角區域放大,量尺寸/配色(供找 sprite 資產來源,不入版控)。"""
from PIL import Image
import sys

src = sys.argv[1] if len(sys.argv) > 1 else 'dosbox/ow_04_finish.png'
im = Image.open(src).convert('RGB')
W, H = im.size
print('image', W, H)

# 房間在左上;主角約在房間中央。先裁房間中心一塊放大觀察。
# frame04 房間約 x[180:430] y[85:290];主角約中心 x[285:325] y[148:195]
box = (270, 140, 340, 205)
crop = im.crop(box).resize(((box[2]-box[0])*8, (box[3]-box[1])*8), Image.NEAREST)
crop.save('work/hero_crop.png')
print('saved work/hero_crop.png from box', box)

# 也裁城鎮戶外幀(兩角色)
try:
    im2 = Image.open('dosbox/ow_20_walk.png').convert('RGB')
    box2 = (270, 150, 360, 230)
    c2 = im2.crop(box2).resize(((box2[2]-box2[0])*6, (box2[3]-box2[1])*6), Image.NEAREST)
    c2.save('work/hero_crop_town.png')
    print('saved work/hero_crop_town.png from box', box2)
except Exception as e:
    print('town crop skip', e)
