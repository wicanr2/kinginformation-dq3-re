#!/usr/bin/env python3
"""FIRST.SCR 開場/標題畫面 render。
格式:640×350、4bpp、row-interleaved planar(每列 = plane0..3 各 80 byte = 320 B/row)。

色盤:開場插圖**不用** DQ3.PAL(遊戲內 tile 色盤)。原版在顯示開場前另設一組獨立的
16 色 VGA DAC,其膚色在 index 14(DQ3.PAL[14] 是灰色,直接套會把臉畫成灰的)。
此 16 色由實機截圖(dq3_org_pic/727339426…,姓名輸入畫面同圖)反推:女主角區(右側無
視窗遮擋)逐 index 取所有像素的中位色;主色(膚/紅/藍/金/白)樣本量大、結果穩定,
idx 8/11 在本圖未使用。膚色 idx14 釘到 DQ 標準肉色 fcd89c。
輸出 docs/title/first_opening.png。 用法: tools/dockrun.sh tools/title_render.py
"""
import os, sys
sys.path.insert(0,'tools')
from PIL import Image

# 開場專用 16 色(實機反推;見上方 docstring)。
OPENING_PAL = [
    (0x00,0x00,0x00),  #  0 背景黑
    (0xcc,0xd4,0xd4),  #  1 亮灰/白邊
    (0xa8,0xa8,0xa8),  #  2 灰細節
    (0xfc,0x48,0x00),  #  3 橙紅(披風)
    (0x08,0x8c,0xfc),  #  4 亮藍(髮)
    (0x00,0x54,0xdc),  #  5 藍(衣)
    (0x00,0x40,0xb4),  #  6 深藍(衣陰影)
    (0xe0,0x94,0x10),  #  7 金/橙(飾邊)
    (0x00,0x00,0x00),  #  8 (本圖未用)
    (0xfc,0xe8,0x20),  #  9 黃金(冠/飾)
    (0xcc,0x78,0x10),  # 10 棕(靴/飾)
    (0x00,0x00,0x00),  # 11 (本圖未用)
    (0xdc,0x00,0x00),  # 12 紅
    (0xfc,0xe0,0xb4),  # 13 亮膚(高光)
    (0xfc,0xd8,0x9c),  # 14 膚色(主)← 修正:原誤套 DQ3.PAL[14] 灰
    (0xfc,0xfc,0xfc),  # 15 白
]

d=open('assets_raw/FIRST.SCR','rb').read()
W,H=640,350; BPR=W//8                 # 80 byte/plane/row
im=Image.new('P',(W,H)); fl=[]
for c in OPENING_PAL: fl+=list(c)
while len(fl)<768: fl+=[0,0,0]
im.putpalette(fl); px=im.load()
for y in range(H):
    for xb in range(BPR):
        b=[d[y*4*BPR + p*BPR + xb] for p in range(4)]   # row-interleaved planes
        for bit in range(8):
            v=0
            for p in range(4): v|=((b[p]>>(7-bit))&1)<<p
            px[xb*8+bit,y]=v
os.makedirs('docs/title',exist_ok=True)
im.convert('RGB').save('docs/title/first_opening.png')
print('saved docs/title/first_opening.png', im.size)
