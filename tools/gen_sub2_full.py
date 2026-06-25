#!/usr/bin/env python3
"""Step 2:抽每個 sub2 handler(跳表 0x3bb4)的條件對話結構。
docs/re-log-722 Step6 模式:test_flag(0x8279,bx=flag) / set_flag(0x8264) /
cmp [0x722],region / 給物 [0x2593]=item / 顯示 di(rec=di-0xbb8)/ warp call 0xd1f9。
輸出每個 byte4 的 {flags, recs, give_items, regions, special} → 供 data-driven wiring。
用法(docker uv venv + capstone):tools/gen_sub2_full.py
"""
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"; DG = 0x16140; HDR = 0x1370
TBL = 0x3bb4            # sub2 跳表(DGROUP logical)
import struct

def main():
    d = open(EXE, "rb").read()
    md = Cs(CS_ARCH_X86, CS_MODE_16)
    def hw(off): return struct.unpack_from("<H", d, DG + off)[0]
    def handler(n): return hw(TBL + n*2)   # byte4 → handler logical
    # 掃多少 byte4?跳表到 0xff/重複前。先抓 0..63
    print("byte4 | handler | test_flag | set/clr_flag | give_item | recs(di-0xbb8) | region | warp")
    print("|---|---|---|---|---|---|---|---|")
    for n in range(0, 64):
        h = handler(n)
        if h == 0 or h == 0xffff: continue
        fl = h + HDR
        flags_t=[]; flags_s=[]; items=[]; recs=[]; regions=[]; warp=False
        last_bx=None
        for ins in md.disasm(d[fl:fl+0x140], fl):
            m, o = ins.mnemonic, ins.op_str
            if m=="mov" and o.startswith("bx, 0x"):
                try: last_bx=int(o.split(",")[1],0)
                except: last_bx=None
            elif m=="call":
                if "0x8279" in o and last_bx is not None: flags_t.append(last_bx)
                elif "0x8264" in o and last_bx is not None: flags_s.append(last_bx)
                elif "0xd1f9" in o: warp=True
            elif m=="mov" and o.startswith("di, 0x"):
                v=int(o.split(",")[1],0)
                if 0xbb8<=v<0xd00: recs.append(v-0xbb8)
            elif m=="mov" and "[0x2593]" in o.split(",")[0]:
                try: items.append(int(o.split(",")[1],0))
                except: pass
            elif m=="cmp" and "[0x722]" in o:
                rv=o.split(",")[-1].strip()
                if rv.startswith("0x") or rv.isdigit(): regions.append(int(rv,0))
            if m in ("ret","retf"): break
        def fmt(l): return ",".join("0x%02x"%x for x in l) if l else "-"
        print(f"| {n} | L0x{h:04x} | {fmt(flags_t)} | {fmt(flags_s)} | {fmt(items)} | {','.join(map(str,recs)) or '-'} | {fmt(regions)} | {'Y' if warp else '-'} |")

if __name__ == "__main__":
    main()
