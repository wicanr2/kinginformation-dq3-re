import sys; sys.path.insert(0,'tools')
from tile_lib import load_pal
from PIL import Image
im=Image.open('work/packbg_full.png').convert('RGB')
# 右上草原(px 320-639, row 0-87)放大 2x
crop=im.crop((320,0,640,88)).resize((640,176),Image.NEAREST)
crop.save('work/packbg_grass.png'); print('saved grassland', crop.size)
# 右下海洋(px 320-639, row 88-245)
crop2=im.crop((320,88,640,246)).resize((640,316),Image.NEAREST)
crop2.save('work/packbg_ocean.png'); print('saved ocean', crop2.size)
