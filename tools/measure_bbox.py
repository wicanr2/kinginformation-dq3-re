from PIL import Image
import sys
for f in ['dosbox/title_03_enter.png','dosbox/ow_04_finish.png']:
    try:
        im=Image.open(f).convert('RGB'); W,H=im.size; px=im.load()
        x0,y0,x1,y1=W,H,0,0
        for y in range(H):
            for x in range(W):
                if px[x,y]!=(0,0,0):
                    x0=min(x0,x);y0=min(y0,y);x1=max(x1,x);y1=max(y1,y)
        print(f, 'size',W,H,'content bbox',(x0,y0,x1+1,y1+1),'=>',(x1+1-x0),'x',(y1+1-y0))
    except Exception as e:
        print(f,'ERR',e)
