#!/usr/bin/env python3
"""DQ3.EXE call census:線性反組譯主程式碼段,統計 call/lcall 目標。
遠呼叫(lcall seg:off)目標段揭露 runtime library;近呼叫熱點揭露主程式函式。
在 docker 內跑(capstone)。"""
import struct, collections
from capstone import Cs, CS_ARCH_X86, CS_MODE_16

d=open("assets_raw/DQ3.EXE","rb").read()
e_cparhdr=struct.unpack_from('<H',d,8)[0]
HDR=e_cparhdr*16            # 0x1370,載入影像起點
md=Cs(CS_ARCH_X86, CS_MODE_16)

# 主程式碼段 = seg 0 → 影像前 64KB(file HDR..HDR+0x10000)
start, end = HDR, min(HDR+0x10000, len(d))
code=d[start:end]
far_seg=collections.Counter()    # lcall 目標 segment
far_tgt=collections.Counter()    # lcall seg:off
near_tgt=collections.Counter()   # near call 目標(影像內 offset)
int_calls=collections.Counter()
n_ins=0
for ins in md.disasm(code, 0):   # 位址用相對 seg0 offset
    n_ins+=1
    m=ins.mnemonic; ops=ins.op_str
    if m=='lcall' and ',' in ops:
        try:
            seg,off=ops.split(',')
            seg=int(seg.strip(),16); off=int(off.strip(),16)
            far_seg[seg]+=1; far_tgt[(seg,off)]+=1
        except: pass
    elif m=='call' and ops.startswith('0x'):
        try: near_tgt[int(ops,16)]+=1
        except: pass
    elif m=='int':
        int_calls[ops]+=1

print(f"# 反組譯 seg0 主碼 file 0x{start:x}..0x{end:x}, {n_ins} 指令")
print(f"\n## lcall 目標 segment(runtime library 面),前 25:")
for seg,c in far_seg.most_common(25):
    fileoff=HDR+seg*16
    print(f"  seg 0x{seg:04x} (file 0x{fileoff:x}): {c} 次")
print(f"\n## 最常被遠呼叫的 runtime 函式(seg:off),前 25:")
for (seg,off),c in far_tgt.most_common(25):
    print(f"  {seg:04x}:{off:04x} (file 0x{HDR+seg*16+off:x}): {c} 次")
print(f"\n## 近呼叫熱點(主程式函式),前 20:")
for off,c in near_tgt.most_common(20):
    print(f"  seg0:{off:04x} (file 0x{HDR+off:x}): {c} 次")
print(f"\n## int 呼叫:")
for i,c in int_calls.most_common():
    print(f"  int {i}: {c} 次")
print(f"\n# 遠呼叫總數={sum(far_seg.values())}, 近呼叫總數={sum(near_tgt.values())}, 不同 runtime 段={len(far_seg)}")
