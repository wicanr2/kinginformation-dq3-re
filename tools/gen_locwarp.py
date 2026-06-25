#!/usr/bin/env python3
"""抽 DQ3.EXE 的 8 個 location scripted warp(call 0xd1f9 + 固定 5-byte struct)。
docs/44 §5b。每處:struct{b0,cur_map(源CTY),dest,X,Y} + gate(flag / runner[0x722])。
gate 釐清:warp 是 runner-driven scripted 事件,觸發在 runner(同主 blocker);此處只解 data。
輸出:docs/44 表 + dq3_remake/src/dq3_locwarp.{c,h} 資料(未 wiring,備將來 runner 解出)。
用法(docker uv venv + capstone):tools/gen_locwarp.py
"""
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

EXE = "assets_raw/DQ3.EXE"
DGROUP_BASE = 0x16140
CALLS = [0x56ea, 0x680b, 0x6d27, 0x7035, 0x71bf, 0x7285, 0x737e, 0x75a7]
STRUCT_AT = {0x56ea:0x4ee4, 0x680b:0x4eb5, 0x6d27:0x4ed5, 0x7035:0x4ed0,
             0x71bf:0x4eda, 0x7285:0x4ec0, 0x737e:0x4ec5, 0x75a7:0x4edf}

def main():
    d = open(EXE, "rb").read()
    md = Cs(CS_ARCH_X86, CS_MODE_16)
    def s5(ds): b = DGROUP_BASE + ds; return list(d[b:b+5])
    def gate(c):
        seg = d[c-0x60:c]; g = None; last_bx = None
        for x in md.disasm(seg, c-0x60):
            if x.mnemonic == "cmp" and "0x722" in x.op_str:
                g = ("runner", x.op_str.split(",")[-1].strip())
            if x.mnemonic == "mov" and x.op_str.startswith("bx, "):
                try: last_bx = int(x.op_str.split(",")[1], 0)
                except: pass
            if x.mnemonic == "call" and last_bx is not None and "0x8279" in x.op_str:
                g = ("flag", last_bx)
        return g
    rows = []
    for c in CALLS:
        st = s5(STRUCT_AT[c]); g = gate(c)
        rows.append((c, st[1], st[2], st[3], st[4], g))
    # 表
    print("| call@ | 源CTY | dest | 落點 | gate |")
    print("|---|---|---|---|---|")
    for c, src, dst, x, y, g in rows:
        gd = (f"flag=0x{g[1]:02x}" if g and g[0]=="flag"
              else f"runner[0x722]={g[1]}" if g else "未明")
        print(f"| 0x{c:04x} | {src} | {dst} | ({x},{y}) | {gd} |")
    # C 資料
    print("\n--- dq3_locwarp.c data ---")
    print("const dq3_locwarp dq3_locwarps[8] = {")
    for c, src, dst, x, y, g in rows:
        gt = 1 if g and g[0]=="flag" else 2 if g else 0   # 0未明 1flag 2runner
        gv = g[1] if g and g[0]=="flag" else (int(g[1],0) if g else 0)
        print(f"  {{ {src:3d}, {dst:2d}, {x:3d}, {y:2d}, {gt}, 0x{gv:02x} }},  /* call@0x{c:04x} */")
    print("};")

if __name__ == "__main__":
    main()
