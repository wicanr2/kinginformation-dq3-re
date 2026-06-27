#!/usr/bin/env python3
"""把精訊 DQ3 的音樂封包 MBG.MCX 拆成各音軌。

格式(RE 自 DQ3.EXE 的 _sbfm_play_music 載入路徑,見 docs/57):
- MBG.MCX 開頭 = 連續遞增的 dword 偏移表(little-endian),每筆 = 一條音軌在檔內的起點。
  第一筆同時等於「表佔的 bytes」(表後緊接第一軌)。表到「下一筆不再遞增」為止。
- 每軌 = Creative CMF 變體結構(header 內含樂器/事件/tempo 指標),由 EXE 連入的 CMFDRV(OPL2 FM)解譯。
- 載入路徑:遊戲對某 open file 做 int21 AH=0x42 lseek 到該軌 offset、AH=0x3f read 進配置段,
  再 _sbfm_instrument(樂器)+ _sbfm_play_music(事件)。

本工具只「拆」遊戲自身資料檔(使用者正版),輸出 gitignore 的 work/music/,不重現受版權旋律。
用法:python extract_cmf.py [MBG.MCX] [out_dir]
"""
import sys, os, struct

def split(path, outdir):
    d = open(path, "rb").read()
    offs = []
    i = 0
    while i + 4 <= len(d):
        v = struct.unpack_from("<I", d, i)[0]
        if offs and v <= offs[-1]: break
        if v >= len(d) or v < (len(offs)+1)*4: break
        offs.append(v); i += 4
    print("MBG.MCX size 0x%05x (%d)  音軌數 %d  偏移表 %d bytes" % (len(d), len(d), len(offs), offs[0]))
    os.makedirs(outdir, exist_ok=True)
    for k, o in enumerate(offs):
        end = offs[k+1] if k+1 < len(offs) else len(d)
        body = d[o:end]
        fn = os.path.join(outdir, "track_%02d.bin" % k)
        open(fn, "wb").write(body)
        # header 前 16 byte 供對照(struct 內 word 指標)
        hh = " ".join("%02x" % x for x in body[:16])
        print("  track %02d  @0x%05x  len 0x%04x (%6d)  hdr: %s" % (k, o, len(body), len(body), hh))

if __name__ == "__main__":
    src = sys.argv[1] if len(sys.argv) > 1 else "assets_raw/MBG.MCX"
    out = sys.argv[2] if len(sys.argv) > 2 else "work/music"
    split(src, out)
