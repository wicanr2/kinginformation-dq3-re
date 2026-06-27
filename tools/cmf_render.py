#!/usr/bin/env python3
"""精訊 DQ3 音軌(MBG.MCX 拆出的 CMF 變體)離線 render → wav,驗證事件解碼正確。

事件格式(RE 自軌資料 + 驅動,見 docs/57):
- 軌 header word[3] = event block 起點(實測恆 0x60);events 為 MIDI-like:
  [status][data...][delta] 重複。status 高 nibble:9x=note-on(note,vel)、8x=note-off(note,vel)、
  bx=control(ctrl,val)、cx=program(prog)、ex=pitch(?)。低 nibble = MIDI channel。
  每事件尾隨 1 byte delta = 到下個事件要等的 tick 數。vel=0 的 9x 視為 note-off。
- 本 render 只為「聽到精訊自己的旋律」驗證:用簡單 2-op FM 音色合成 note 事件(非 cycle-accurate OPL2;
  authentic OPL2 timbre 留給 C 版 dq3_opl2)。播放使用者正版遊戲自身資料,輸出 gitignore,不重現受版權樂曲。

用法:python cmf_render.py work/music/track_00.bin out.wav [tick_ms]
"""
import sys, struct, math, wave

SR = 44100

def parse_events(track):
    # 事件區起點:header word[3](恆 0x60)起、第一個 status byte(>=0x80)
    base = track[6] | (track[7] << 8)
    if not (0 < base < len(track)): base = 0x60
    p = base
    while p < len(track) and track[p] < 0x80:
        p += 1
    events = []          # (status, d1, d2, delta)
    status = None
    while p < len(track):
        b = track[p]
        if b >= 0x80:    # 新 status
            status = b; p += 1
            if status in (0xff, 0xf0): break   # end-of-track / meta
        if status is None: break
        hi = status & 0xf0
        if hi in (0x80, 0x90, 0xb0, 0xa0, 0xe0):   # 2 data + delta
            if p + 3 > len(track): break
            d1, d2 = track[p], track[p+1]; dl = track[p+2]; p += 3
            events.append((status, d1, d2, dl))
        elif hi in (0xc0, 0xd0):                   # 1 data + delta
            if p + 2 > len(track): break
            d1 = track[p]; dl = track[p+1]; p += 2
            events.append((status, d1, 0, dl))
        else:
            break
    return events

def midi_freq(n):
    return 440.0 * (2.0 ** ((n - 69) / 12.0))

def fm_sample(t, freq, env):
    # 2-op FM:carrier sine 被 modulator(2x)調變,mod index 隨 env 衰減 → AdLib 暖音色近似
    mod = math.sin(2*math.pi*freq*2*t) * (2.5 * env)
    return math.sin(2*math.pi*freq*t + mod) * env

def render(track, out, tick_ms=16.0, max_sec=25.0):
    import numpy as np
    events = parse_events(track)
    rendered = []            # (start_time, freq, dur)
    cur = 0.0; spt = tick_ms / 1000.0
    active = {}
    nev = 0
    for st, d1, d2, dl in events:
        hi = st & 0xf0; ch = st & 0x0f
        if hi == 0x90 and d2 > 0:
            active[(ch, d1)] = cur
        elif (hi == 0x80) or (hi == 0x90 and d2 == 0):
            key = (ch, d1)
            if key in active:
                rendered.append((active[key], midi_freq(d1), cur - active[key])); del active[key]
        cur += dl * spt
        nev += 1
        if cur > max_sec: break
    for (ch, n), tstart in active.items():
        rendered.append((tstart, midi_freq(n), 0.2))
    dur_total = min(cur, max_sec) + 0.3
    N = int(dur_total * SR)
    buf = np.zeros(N, dtype=np.float32)
    for tstart, freq, dur in rendered:
        s0 = int(tstart * SR); ns = max(1, int(min(dur, 1.2) * SR))
        if s0 >= N: continue
        ns = min(ns, N - s0)
        tt = np.arange(ns) / SR
        env = np.exp(-3.0 * tt / max(0.05, dur))
        env *= np.clip(np.minimum(tt * SR / 80.0, (ns - np.arange(ns)) / 200.0), 0, 1)
        mod = np.sin(2*np.pi*freq*2*tt) * (2.5 * env)
        buf[s0:s0+ns] += (0.16 * np.sin(2*np.pi*freq*tt + mod) * env).astype(np.float32)
    peak = max(1e-6, float(np.max(np.abs(buf))))
    pcm = (np.clip(buf * min(1.0, 0.9/peak), -1, 1) * 32767).astype('<i2').tobytes()
    w = wave.open(out, "wb"); w.setnchannels(1); w.setsampwidth(2); w.setframerate(SR)
    w.writeframes(pcm); w.close()
    print("events(total)=%d rendered_notes=%d render_dur=%.1fs (cap %.0fs) -> %s"
          % (len(events), len(rendered), dur_total, max_sec, out))

if __name__ == "__main__":
    track = open(sys.argv[1], "rb").read()
    out = sys.argv[2] if len(sys.argv) > 2 else "work/music/render.wav"
    tick = float(sys.argv[3]) if len(sys.argv) > 3 else 16.0
    render(track, out, tick)
