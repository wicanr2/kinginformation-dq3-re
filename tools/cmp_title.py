from PIL import Image, ImageChops
import sys
# DOSBox 原版標題:遊戲區 640×350 在左上
dos = Image.open('dosbox/title_03_enter.png').convert('RGB').crop((0,0,640,350))
rem = Image.open('dq3_remake/titg_cmp.png').convert('RGB')
if rem.size!=(640,350): rem=rem.resize((640,350))
# pixel diff
diff = ImageChops.difference(dos, rem)
import statistics
W,H=640,350; dp=diff.load(); same=0; near=0; tot=W*H
for y in range(H):
    for x in range(W):
        r,g,b=dp[x,y]
        if r==0 and g==0 and b==0: same+=1
        if r<24 and g<24 and b<24: near+=1
print('exact-match pixels  : %d/%d = %.1f%%'%(same,tot,100*same/tot))
print('near-match (<24/ch) : %d/%d = %.1f%%'%(near,tot,100*near/tot))
# 並排 + 差異圖
sbs=Image.new('RGB',(640*3+20,350),(30,30,30))
sbs.paste(dos,(0,0)); sbs.paste(rem,(650,0))
# 差異放大顯示
dv=diff.point(lambda v: min(255,v*6)); sbs.paste(dv,(1300,0))
sbs.save('dq3_remake/cmp_title.png'); print('saved dq3_remake/cmp_title.png (左:DOSBox原版 中:remake 右:差異×6)')
