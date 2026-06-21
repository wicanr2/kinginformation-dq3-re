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

def deobf_rows(blob,b):
    """去混淆:左右半對調(左欄←低byte、右欄←高byte)+ 左半下移 1 列。"""
    hi=[blob[b+y*2] for y in range(16)]; lo=[blob[b+y*2+1] for y in range(16)]
    return [((lo[y-1] if y>=1 else 0)<<8)|hi[y] for y in range(16)]

def validity(rows):
    """字模『像不像正字』分數:筆畫連通性 - 孤立雜點懲罰。打亂的字連通低、孤立點多。"""
    px=[[(rows[y]>>(15-x))&1 for x in range(16)] for y in range(16)]
    ink=sum(sum(r) for r in px)
    if ink==0: return -9
    nb=0; iso=0
    for y in range(16):
        for x in range(16):
            if not px[y][x]: continue
            c=0
            for dy,dx in ((1,0),(-1,0),(0,1),(0,-1)):
                ny,nx=y+dy,x+dx
                if 0<=ny<16 and 0<=nx<16 and px[ny][nx]: c+=1
            nb+=c
            if c==0: iso+=1
    return nb/ink - 1.5*(iso/ink)

connectivity=validity  # 相容舊呼叫

def extract_section(si):
    """回傳 [(byte_off, rows, obf_flag) ...]。對每字自動判斷是否被精訊左右對調混淆,
    若去混淆後筆畫連通性明顯更高則套用去混淆。"""
    blob=data[offs[si]:offs[si+1]]; L=len(blob)
    glyphs=[]; b=0
    while b+32<=L:
        if is_glyph(blob,b):
            o=rows16(blob,b); d=deobf_rows(blob,b)
            if connectivity(d) > connectivity(o)+0.05:
                glyphs.append((b,d,True))
            else:
                glyphs.append((b,o,False))
            b+=32
        else:
            b+=1
    return glyphs

def atlas(glyphs,cols,scale,path):
    rows_n=(len(glyphs)+cols-1)//cols; pad=1
    img=Image.new("L",(cols*(16+pad)+pad,rows_n*(16+pad)+pad),40); px=img.load()
    for i,item in enumerate(glyphs):
        g=item[1]
        gx=(i%cols)*(16+pad)+pad; gy=(i//cols)*(16+pad)+pad
        for y in range(16):
            for x in range(16):
                if (g[y]>>(15-x))&1: px[gx+x,gy+y]=255
    img.resize((img.width*scale,img.height*scale),Image.NEAREST).save(path)

if __name__=="__main__":
    secs=[int(x) for x in sys.argv[1:]] or list(range(22))
    allg=[]; nobf=0
    for si in secs:
        gl=extract_section(si); allg+=gl
        nobf+=sum(1 for x in gl if x[2])
        if len(secs)<=3:
            atlas(gl,24,4,f"{OUT}/greedy_sec{si:02d}.png")
        print(f"sec{si:2}: {len(gl)} 字 (混淆 {sum(1 for x in gl if x[2])})")
    print(f"合計: {len(allg)} 字, 其中被混淆 {nobf}")
    if len(secs)>3:
        atlas(allg,48,3,f"{OUT}/CHINA.greedy.atlas.png")
        print("-> CHINA.greedy.atlas.png")
