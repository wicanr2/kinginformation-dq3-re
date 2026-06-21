#!/usr/bin/env python3
"""字型抽取/atlas 產生器。
字模格式:16x16, row-major, MSB-first, 2 bytes/列, 32 bytes/字(以 D3TXT00.FON 驗證)。

- D3TXT00.FON / SETTXT.FON:無 header,整檔即連續字模 → 完整、已驗證正確。
- CHINA.FON:變寬(proportional)母字庫,逐字位置由每段索引/偏移表決定(見
  docs/02-font-format.md)。固定步長無法正確切割,故此處僅輸出「以 D3TXT00 範本
  在 CHINA 完全比中的真值字模子集」當作正確性佐證;全字抽取待偏移表解開。

atlas PNG 為衍生字表;raw 遊戲檔不入版控,自備 assets_raw/ 後執行本工具重新產生。
"""
import struct, os
from PIL import Image
RAW="assets_raw"; OUT="assets_out/font"
os.makedirs(OUT, exist_ok=True)

def rows16(blob, base):
    return [(blob[base+y*2]<<8)|blob[base+y*2+1] for y in range(16)]

def atlas(glyphs, cols, scale, path):
    rows_n=(len(glyphs)+cols-1)//cols; pad=1
    img=Image.new("L",(cols*(16+pad)+pad, rows_n*(16+pad)+pad),40); px=img.load()
    for i,g in enumerate(glyphs):
        gx=(i%cols)*(16+pad)+pad; gy=(i//cols)*(16+pad)+pad
        for y in range(16):
            for x in range(16):
                if (g[y]>>(15-x))&1: px[gx+x,gy+y]=255
    img.resize((img.width*scale,img.height*scale),Image.NEAREST).save(path)
    return len(glyphs)

# ---- ASCII / 遊戲內字型(無 header,完整正確) ----
for name in ("D3TXT00.FON","SETTXT.FON"):
    blob=open(f"{RAW}/{name}","rb").read(); n=len(blob)//32
    cnt=atlas([rows16(blob,i*32) for i in range(n)],16,3,f"{OUT}/{name}.atlas.png")
    print(f"{name}: {cnt} glyphs -> {name}.atlas.png")

# ---- CHINA.FON:輸出真值比中子集(正確),並存錨點 ----
d3=open(f"{RAW}/D3TXT00.FON","rb").read()
china=open(f"{RAW}/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',china,i*4)[0] for i in range(22)]+[len(china)]
tmpls={}
for i in range(len(d3)//32):
    g=d3[i*32:(i+1)*32]
    if sum(bin(b).count('1') for b in g)>=24: tmpls[g]=i
anchors=[]
for off in range(len(china)-32):
    seg=china[off:off+32]
    if seg in tmpls:
        s=next(k for k in range(22) if offs[k]<=off<offs[k+1])
        anchors.append((off, s, off-offs[s], tmpls[seg]))
with open(f"{OUT}/china_d3txt00_anchors.tsv","w") as f:
    f.write("china_abs_off\tsection\trel_off\td3txt00_index\n")
    for a in anchors: f.write("\t".join(map(str,a))+"\n")
verified=[rows16(china,off) for off,_,_,_ in anchors]
atlas(verified,48,2,f"{OUT}/CHINA.verified_subset.atlas.png")
print(f"CHINA.FON: {len(anchors)} 個真值比中字模 -> CHINA.verified_subset.atlas.png + anchors.tsv")
