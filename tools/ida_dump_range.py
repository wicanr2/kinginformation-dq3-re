"""IDA batch script: dump analyzed instructions in an address range.

Usage:
  idat -A '-Stools/ida_dump_range.py 0x173c0 0x17430' DQ3.EXE

This is useful when IDA has identified code but has not recovered a function
boundary, as commonly happens around 16-bit DOS jump tables.
"""

import ida_auto
import ida_bytes
import ida_lines
import idautils
import idc


ida_auto.auto_wait()
start = int(idc.ARGV[1], 0)
end = int(idc.ARGV[2], 0)

for ea in idautils.Heads(start, end):
    flags = ida_bytes.get_full_flags(ea)
    if not ida_bytes.is_code(flags):
        continue
    line = ida_lines.generate_disasm_line(
        ea, ida_lines.GENDSM_REMOVE_TAGS
    ) or ""
    print(f"IDA_RANGE {ea:#x} {line}")

idc.qexit(0)
