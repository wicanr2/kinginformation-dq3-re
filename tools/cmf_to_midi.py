#!/usr/bin/env python3
"""精訊 DQ3 音軌(MBG.MCX 拆出)→ 標準 MIDI 檔(.mid),供「MT-32 / Roland 風」render。

事件流本身是 MIDI-like(note on/off、program change、channel),轉成 SMF 後可用任何 GM/MT-32
音色庫(SoundFont / MUNT)合成成管弦 / Roland 音色。格式 RE 見 docs/57:
event base = header w1(track[2])、running status、尾隨 1-byte delta、tempo div = w3(track[6]) → tick ≈ 10.4ms。
FM program 是 OPL 樂器索引(非 GM),映射到一組好聽的 GM 樂器(對齊 dq3_cmf 8 patch 角色)。
用法:python cmf_to_midi.py track.bin out.mid
"""
import sys, struct

GM = [48, 32, 19, 49, 73, 46, 71, 11]   # strings, ac.bass, church organ, slow strings, flute, harp, clarinet, vibraphone

def parse_events(d):
    base = d[2] | (d[3] << 8)
    if not (0 < base < len(d)): base = d[6] | (d[7] << 8)
    if not (0 < base < len(d)): base = 0x60
    p = base
    while p < len(d) and d[p] < 0x80: p += 1
    evs = []; status = None
    while p < len(d):
        b = d[p]
        if b >= 0x80:
            status = b; p += 1
            if status in (0xff, 0xf0): break
        if status is None: break
        hi = status & 0xf0
        if hi in (0x80, 0x90, 0xb0, 0xa0, 0xe0):
            if p + 3 > len(d): break
            evs.append((status, d[p], d[p+1], d[p+2])); p += 3
        elif hi in (0xc0, 0xd0):
            if p + 2 > len(d): break
            evs.append((status, d[p], 0, d[p+1])); p += 2
        else: break
    return evs

def vlq(n):
    out = bytearray([n & 0x7f]); n >>= 7
    while n: out.insert(0, (n & 0x7f) | 0x80); n >>= 7
    return bytes(out)

def to_midi(track, out):
    d = open(track, "rb").read()
    evs = parse_events(d)
    PPQ = 48                                  # 48 tick/quarter + tempo 500000us → 1 tick = 10.417ms ≈ CMF tick
    trk = bytearray(b"\x00\xff\x51\x03" + bytes([0x07, 0xa1, 0x20]))   # set tempo 500000us(120bpm)
    pending = 0                               # delta 累積(MIDI delta 在事件前)
    for st, d1, d2, dl in evs:
        hi = st & 0xf0; ch = st & 0x0f
        if   hi == 0xc0: msg = bytes([0xc0 | ch, GM[d1 % len(GM)] & 0x7f])
        elif hi == 0x90: msg = bytes([0x90 | ch, d1 & 0x7f, d2 & 0x7f])
        elif hi == 0x80: msg = bytes([0x80 | ch, d1 & 0x7f, d2 & 0x7f])
        elif hi == 0xb0: msg = bytes([0xb0 | ch, d1 & 0x7f, d2 & 0x7f])
        elif hi == 0xe0: msg = bytes([0xe0 | ch, d1 & 0x7f, d2 & 0x7f])
        else: msg = None
        if msg is None:
            pending += dl; continue
        trk += vlq(pending) + msg
        pending = dl
    trk += vlq(pending) + b"\xff\x2f\x00"      # end of track
    hdr = b"MThd" + struct.pack(">IHHH", 6, 0, 1, PPQ)
    body = b"MTrk" + struct.pack(">I", len(trk)) + bytes(trk)
    open(out, "wb").write(hdr + body)
    print("track %s: %d events -> %s" % (track, len(evs), out))

if __name__ == "__main__":
    to_midi(sys.argv[1], sys.argv[2] if len(sys.argv) > 2 else "out.mid")
