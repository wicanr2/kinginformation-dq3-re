#!/usr/bin/env python3
"""在 cty_loc 標記旁標地點中文名(已確認)+ CTY#。下層 local_y=Y-300。
   名稱對照見 docs/32/33;未確認者僅標 CTY#。"""
import struct
from PIL import Image, ImageDraw, ImageFont
d=open("assets_raw/DQ3.EXE","rb").read();B=0x16140+0x748
def F(s):
    try:return ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",s)
    except:return ImageFont.load_default()
# 中文需 CJK 字型
def CJK(s):
    for p in ["/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc","/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc","/usr/share/fonts/truetype/arphic/uming.ttc"]:
        try: return ImageFont.truetype(p,s)
        except: pass
    return F(s)
CON={0:"阿里阿罕"}
UND={86:"利姆達爾",93:"彩虹水滴祠堂",84:"Haukness",81:"Kol",82:"Kol",
     77:"Brecconaly",87:"Brecconaly",78:"洞窟",88:"洞窟",
     79:"佐瑪城",80:"佐瑪城",90:"佐瑪城",91:"佐瑪城",85:"Cantlin",92:"雲雨之杖",89:"利姆達爾北"}
def go(mapimg,tw,th,filt,labels,out):
    im=Image.open(mapimg).convert("RGB");W,H=im.size;dr=ImageDraw.Draw(im)
    fn=F(30); fc=CJK(30)
    for i in range(100):
        x=d[B+i*4]; y=struct.unpack_from("<H",d,B+i*4+2)[0]; m=d[B+i*4+3]
        yy=filt(x,y,m)
        if yy is None: continue
        px,py=x/tw*W,yy/th*H
        dr.ellipse([px-9,py-9,px+9,py+9],outline=(255,0,0),width=3)
        nm=labels.get(i)
        if nm:
            f=fc if any(ord(ch)>0x2e80 for ch in nm) else fn
            dr.text((px+11,py-16),nm,fill=(255,255,0),font=f,stroke_width=4,stroke_fill=(0,0,0))
        else:
            dr.text((px+11,py-14),str(i),fill=(0,255,255),font=fn,stroke_width=3,stroke_fill=(0,0,0))
    im.save(out);print(out)
go("docs/maps/world_con_full.png",244,205,(lambda x,y,m: y if(0<x<244 and 0<y<205) else None),CON,"docs/maps/world_con_named.png")
go("docs/maps/world_und_full.png",244,167,(lambda x,y,m: y-300 if(0<x<244 and 300<y<467) else None),UND,"docs/maps/world_und_named.png")
