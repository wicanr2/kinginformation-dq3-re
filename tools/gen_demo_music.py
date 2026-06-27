#!/usr/bin/env python3
"""dq3_remake demo 影片的**原創** BGM —— FM 合成(OPL2 / AdLib / 聲霸卡風暖音色),非 DQ3 旋律(避版權)。
2-operator FM(sine carrier + sine modulator,ADSR 控振幅與調變指數)→ 像當年 FM 音源的圓潤音色。
編制:FM brass 主旋律 + FM organ 和聲 pad + FM bass + 柔噪 hat。輸出 work/video/demo_bgm.wav。
docker uv(numpy)執行。"""
import numpy as np, wave, os
SR = 44100
A4 = 440.0
def hz(semi): return A4 * 2**(semi/12)
NAME = {'C':-9,'Cs':-8,'D':-7,'Ds':-6,'E':-5,'F':-4,'Fs':-3,'G':-2,'Gs':-1,'A':0,'As':1,'B':2}
def f(nm, octv): return 0.0 if nm=='r' else hz(NAME[nm] + (octv-4)*12)

def adsr(n, a, d, s, r):
    env = np.zeros(n); a=int(a*SR); d=int(d*SR); r=int(r*SR)
    a=max(a,1); d=max(d,1); r=max(r,1)
    i=0
    seg=min(a,n); env[:seg]=np.linspace(0,1,seg); i=seg
    if i<n: seg=min(d,n-i); env[i:i+seg]=np.linspace(1,s,seg); i+=seg
    if i<n-r: env[i:n-r]=s; i=n-r
    if i<n: env[i:]=np.linspace(env[i-1] if i>0 else s,0,n-i)
    return env

def fm(freq, dur, ratio, idx_peak, vol, a=0.02,d=0.12,s=0.6,rel=0.18):
    """2-op FM voice:carrier=sin(2πft + I(t)·sin(2π·ratio·f·t))。"""
    n=int(dur*SR); t=np.arange(n)/SR
    if freq<=0: return np.zeros(n)
    aenv=adsr(n,a,d,s,rel)
    ienv=adsr(n,a*0.6,d*1.2,s*0.5,rel)*idx_peak   # 調變指數隨時間衰減(brass/bell 起音亮、尾巴柔)
    mod=ienv*np.sin(2*np.pi*ratio*freq*t)
    return np.sin(2*np.pi*freq*t+mod)*aenv*vol

# ── 原創 heroic 主題(C 大調區,溫暖上揚)。每拍 beat 秒 ──
beat=0.40
# 旋律(音名,八度,拍數);兩段 16 小節,A 段莊嚴、B 段抒情
mel=[('G',4,1),('C',5,1.5),('E',5,0.5),('D',5,1),('C',5,1),
     ('D',5,2),('E',5,1),('G',5,1),
     ('A',5,1.5),('G',5,0.5),('E',5,1),('D',5,1),
     ('C',5,2),('r',0,1),('G',4,1),
     ('A',4,1),('C',5,1.5),('E',5,0.5),('G',5,1),('E',5,1),
     ('F',5,2),('E',5,1),('D',5,1),
     ('E',5,1.5),('D',5,0.5),('C',5,1),('B',4,1),
     ('C',5,3),('r',0,1)]
# 和弦進行(每小節一個三和弦根+三+五;C Am F G 風)
chords=[('C',4),('A',3),('F',3),('G',3), ('C',4),('E',4),('F',3),('G',3),
        ('F',3),('C',4),('D',4),('G',3), ('C',4),('F',3),('G',3),('C',4)]
def triad_freqs(root,octv):
    base=NAME[root]+(octv-4)*12
    return [hz(base),hz(base+4 if root in('C','F','G','E','D','A') else base+3),hz(base+7)]

def build_melody():
    buf=[]
    for nm,oc,b in mel:
        buf.append(fm(f(nm,oc), beat*b, 2.0, 2.2, 0.30, a=0.015,d=0.10,s=0.55,rel=0.12))  # FM brass(ratio2)
    return np.concatenate(buf)

def build_chords(total):
    out=np.zeros(total); pos=0; ci=0
    while pos<total:
        root,octv=chords[ci%len(chords)]; ci+=1
        dur=beat*2; n=int(dur*SR)
        seg=np.zeros(min(n,total-pos))
        for fr in triad_freqs(root,octv):
            v=fm(fr,dur,1.0,1.0,0.10,a=0.04,d=0.2,s=0.7,rel=0.25)  # FM organ(ratio1,柔 pad)
            seg+=v[:len(seg)]
        out[pos:pos+len(seg)]+=seg; pos+=len(seg)
    return out

def build_bass(total):
    out=np.zeros(total); pos=0; ci=0
    while pos<total:
        root,octv=chords[ci%len(chords)]; ci+=1
        for _ in range(2):  # 每小節兩拍 bass
            dur=beat; n=int(dur*SR)
            v=fm(hz(NAME[root]+(octv-2-4)*12),dur,1.0,1.4,0.34,a=0.005,d=0.08,s=0.5,rel=0.1)
            seg=v[:min(n,total-pos)]; out[pos:pos+len(seg)]+=seg; pos+=len(seg)
            if pos>=total: break
    return out

def build_hat(total):
    out=np.zeros(total); step=int(beat*SR); rng=np.random.default_rng(3)
    for i in range(0,total,step):
        n=min(int(0.025*SR),total-i)
        out[i:i+n]+=rng.uniform(-1,1,n)*np.linspace(0.06,0,n)
    return out

mel_w=build_melody(); total=len(mel_w)
mix=mel_w+build_chords(total)+build_bass(total)+build_hat(total)
# loop 到 ~80s
loops=int(np.ceil(80/(total/SR)))
full=np.tile(mix,loops)
# 輕量「房間感」:短延遲疊加
d=int(0.06*SR); echo=np.zeros_like(full); echo[d:]=full[:-d]*0.18; full=full+echo
full=full/np.max(np.abs(full))*0.82
fi=int(2*SR); fo=int(3*SR)
full[:fi]*=np.linspace(0,1,fi); full[-fo:]*=np.linspace(1,0,fo)
pcm=(full*32767).astype('<i2')
out=os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))),"work","video","demo_bgm.wav")
os.makedirs(os.path.dirname(out),exist_ok=True)
w=wave.open(out,'wb'); w.setnchannels(1); w.setsampwidth(2); w.setframerate(SR); w.writeframes(pcm.tobytes()); w.close()
print("BGM(FM 合成)->",out,"%.1fs"%(len(full)/SR))
