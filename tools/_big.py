import struct
from PIL import Image, ImageDraw, ImageFont
d=open("assets_raw/DQ3.EXE","rb").read();B=0x16140+0x748
def F(s):
    try:return ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",s)
    except:return ImageFont.load_default()
def go(mapimg,tw,th,filt,out):
    im=Image.open(mapimg).convert("RGB");W,H=im.size;dr=ImageDraw.Draw(im);f=F(34)
    for i in range(100):
        x=d[B+i*4]; y=struct.unpack_from("<H",d,B+i*4+2)[0]; m=d[B+i*4+3]
        yy=filt(x,y,m)
        if yy is None: continue
        px=x/tw*W; py=yy/th*H
        dr.ellipse([px-9,py-9,px+9,py+9],outline=(255,0,0),width=3)
        dr.text((px+10,py-18),str(i),fill=(255,255,0),font=f,stroke_width=4,stroke_fill=(0,0,0))
    im.save(out)
go("docs/maps/world_con_full.png",244,205,(lambda x,y,m: y if (0<x<244 and 0<y<205) else None),"/out/ovbig.png")
go("docs/maps/world_und_full.png",244,167,(lambda x,y,m: y-300 if (0<x<244 and 300<y<467) else None),"/out/unbig.png")
print("ok")
