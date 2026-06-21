#!/usr/bin/env python3
"""貪婪非重疊字模抽取:處理 CHINA.FON 變位佈局(連續 32-byte 字模 + 零星插入位元組)。
逐 byte 掃描:合格 16x16 字模 → 輸出, 前進 32;否則前進 1 重新對齊。
跳過段首索引表(不合格)與字模間插入位元組。"""
import struct, os, sys
from PIL import Image
RAW="assets_raw"; OUT="assets_out/font"
os.makedirs(OUT, exist_ok=True)
data=open(f"{RAW}/CHINA.FON","rb").read()
offs=[struct.unpack_from('<I',data,i*4)[0] for i in range(22)]+[len(data)]

def rows16(blob,b): return [(blob[b+y*2]<<8)|blob[b+y*2+1] for y in range(16)]

def is_glyph(blob,b):
    if b+32>len(blob): return False
    r=rows16(blob,b)
    # 強對齊約束:此字型為 16x14,row14/row15 為行距留白(D3TXT00 實測 98%/99% 全空)
    if r[15]!=0: return False
    if bin(r[14]).count('1')>2: return False
    body=r[:14]
    dens=sum(bin(x).count('1') for x in body)/(16*14)
    if not (0.08<=dens<=0.55): return False
    cc=sum(1 for c in range(16) if len({(rr>>(15-c))&1 for rr in body})==1)
    if cc>6: return False                              # 排除梳齒表格
    if len(set(body))<7: return False
    return True

def extract_section(si):
    blob=data[offs[si]:offs[si+1]]; L=len(blob)
    glyphs=[]; b=0
    while b+32<=L:
        if is_glyph(blob,b):
            glyphs.append((b,rows16(blob,b))); b+=32
        else:
            b+=1
    return glyphs

def atlas(glyphs,cols,scale,path):
    rows_n=(len(glyphs)+cols-1)//cols; pad=1
    img=Image.new("L",(cols*(16+pad)+pad,rows_n*(16+pad)+pad),40); px=img.load()
    for i,(_,g) in enumerate(glyphs):
        gx=(i%cols)*(16+pad)+pad; gy=(i//cols)*(16+pad)+pad
        for y in range(16):
            for x in range(16):
                if (g[y]>>(15-x))&1: px[gx+x,gy+y]=255
    img.resize((img.width*scale,img.height*scale),Image.NEAREST).save(path)

if __name__=="__main__":
    secs=[int(x) for x in sys.argv[1:]] or list(range(22))
    allg=[]
    for si in secs:
        gl=extract_section(si); allg+=gl
        if len(secs)<=3:
            atlas(gl,24,4,f"{OUT}/greedy_sec{si:02d}.png")
        print(f"sec{si:2}: {len(gl)} 字")
    print("合計:",len(allg))
    if len(secs)>3:
        atlas(allg,48,3,f"{OUT}/CHINA.greedy.atlas.png")
        print("-> CHINA.greedy.atlas.png")
