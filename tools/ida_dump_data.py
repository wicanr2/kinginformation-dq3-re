"""IDA batch script: dump bytes at a segment-relative address.

Usage:
  idat -A "-Stools/ida_dump_data.py 0x15df0 ds 0x3d03 32 0x3d19 32" DQ3.EXE
"""

import ida_auto
import ida_bytes
import ida_segment
import idautils
import idc


ida_auto.auto_wait()
context_ea = int(idc.ARGV[1], 0)
register = idc.ARGV[2]
try:
    selector = int(register, 0)
except ValueError:
    selector = idc.get_sreg(context_ea, register)
segment = ida_segment.get_segm_by_sel(selector)
if segment is None:
    print(
        f"IDA_DATA no segment for {register} selector={selector:#x} "
        f"at {context_ea:#x}"
    )
    for segment_ea in idautils.Segments():
        current = ida_segment.getseg(segment_ea)
        print(
            f"IDA_SEGMENT name={current.name} start={current.start_ea:#x} "
            f"end={current.end_ea:#x} selector={current.sel:#x}"
        )
    idc.qexit(1)

if (len(idc.ARGV) - 3) % 2:
    print("IDA_DATA offsets and sizes must be pairs")
    idc.qexit(1)

base = ida_segment.get_segm_base(segment)
for arg in range(3, len(idc.ARGV), 2):
    offset = int(idc.ARGV[arg], 0)
    size = int(idc.ARGV[arg + 1], 0)
    linear_ea = base + offset
    data = ida_bytes.get_bytes(linear_ea, size)
    if data is None:
        print(
            f"IDA_DATA unreadable {register}:{offset:#x} "
            f"linear={linear_ea:#x} size={size}"
        )
        continue

    print(
        f"IDA_DATA context={context_ea:#x} register={register} "
        f"selector={selector:#x} segment={segment.name} base={base:#x} "
        f"offset={offset:#x} linear={linear_ea:#x} size={size}"
    )
    for pos in range(0, len(data), 16):
        chunk = data[pos : pos + 16]
        print(f"IDA_BYTES {linear_ea + pos:#x} {chunk.hex(' ')}")

idc.qexit(0)
