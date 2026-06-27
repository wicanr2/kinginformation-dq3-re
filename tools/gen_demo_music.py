#!/usr/bin/env python3
"""為 dq3_remake demo 影片生成**原創** chiptune 配樂(避免版權;非 DQ3 旋律)。
heroic RPG 風:方波 lead + 三角波 bass + 簡單噪音打點。輸出 work/video/demo_bgm.wav。
docker uv(numpy)執行。"""
import numpy as np, struct, wave, os, sys
SR = 44100
def note(f, dur, kind="square", vol=0.25):
    n = int(SR*dur); t = np.arange(n)/SR
    if f <= 0: return np.zeros(n)
    if kind == "square": w = np.sign(np.sin(2*np.pi*f*t))
    elif kind == "tri":  w = 2*np.abs(2*(t*f - np.floor(t*f+0.5)))-1
    else: w = np.sin(2*np.pi*f*t)
    env = np.ones(n); a = int(0.01*SR); d = int(0.04*SR)
    env[:a] = np.linspace(0,1,a); env[-d:] *= np.linspace(1,0.0,d)
    return w*env*vol
# 音名→頻率
A4=440.0
def hz(semi): return A4*2**(semi/12)
N = {'C':-9,'D':-7,'E':-5,'F':-4,'G':-2,'A':0,'B':2,'r':None}
def f(name, octv):
    if name=='r': return 0
    return hz(N[name]+(octv-4)*12)
# 原創旋律(heroic,8 小節 loop;每拍 0.34s)。(音名,八度,拍數)
beat = 0.34
lead = [('E',5,1),('G',5,1),('A',5,2), ('G',5,1),('E',5,1),('D',5,2),
        ('C',5,1),('E',5,1),('G',5,2), ('A',5,1),('B',5,1),('A',5,2),
        ('G',5,1),('E',5,1),('F',5,2), ('E',5,1),('D',5,1),('C',5,2),
        ('D',5,1),('E',5,1),('G',5,1),('A',5,1), ('G',5,2),('E',5,2)]
# bass:每小節根音(2 拍一音)
bass_roots = [('A',2),('A',2),('F',2),('F',2),('C',3),('C',3),('G',2),('G',2),
              ('A',2),('A',2),('F',2),('F',2),('C',3),('C',3),('G',2),('A',2)]
def build_lead():
    buf=[]
    for nm,oc,b in lead: buf.append(note(f(nm,oc), beat*b, "square", 0.22))
    return np.concatenate(buf)
def build_bass(total):
    buf=[]
    while sum(len(x) for x in buf) < total:
        for nm,oc in bass_roots:
            buf.append(note(f(nm,oc), beat*2, "tri", 0.30))
    b=np.concatenate(buf); return b[:total]
def build_hat(total):
    out=np.zeros(total); step=int(beat*SR)
    rng=np.random.default_rng(7)
    for i in range(0,total,step):
        n=min(int(0.03*SR), total-i)
        out[i:i+n]+=rng.uniform(-1,1,n)*np.linspace(0.12,0,n)
    return out
lead_w = build_lead()
total = len(lead_w)
mix = lead_w + build_bass(total) + build_hat(total)
# loop 接成 ~75s(影片夠長),頭尾淡入淡出
loops = int(np.ceil(78/(total/SR)))
full = np.tile(mix, loops)
full = full/np.max(np.abs(full))*0.85
fi=int(2*SR); fo=int(3*SR)
full[:fi]*=np.linspace(0,1,fi); full[-fo:]*=np.linspace(1,0,fo)
pcm=(full*32767).astype('<i2')
out=os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),"work","video","demo_bgm.wav")
os.makedirs(os.path.dirname(out),exist_ok=True)
w=wave.open(out,'wb'); w.setnchannels(1); w.setsampwidth(2); w.setframerate(SR); w.writeframes(pcm.tobytes()); w.close()
print("BGM ->", out, "%.1fs"%(len(full)/SR))
