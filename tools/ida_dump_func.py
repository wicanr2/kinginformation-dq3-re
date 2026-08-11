"""IDA batch script: dump one analyzed function, callers, and pseudocode.

Usage:
  idat -A "-Stools/ida_dump_func.py 0x15daf 0x1689c" DQ3.EXE
"""

import ida_auto
import ida_funcs
import ida_hexrays
import ida_lines
import idautils
import idc


ida_auto.auto_wait()
for raw_ea in idc.ARGV[1:]:
    ea = int(raw_ea, 0)
    fn = ida_funcs.get_func(ea)
    if fn is None:
        print(f"IDA_FUNC no function contains {ea:#x}")
        continue

    print(
        f"IDA_FUNC start={fn.start_ea:#x} end={fn.end_ea:#x} "
        f"name={ida_funcs.get_func_name(fn.start_ea)}"
    )
    for xref in idautils.CodeRefsTo(fn.start_ea, False):
        caller = ida_funcs.get_func(xref)
        caller_name = (
            ida_funcs.get_func_name(caller.start_ea) if caller else "<no-func>"
        )
        print(f"IDA_CALLER ea={xref:#x} func={caller_name}")

    for insn_ea in idautils.FuncItems(fn.start_ea):
        line = ida_lines.generate_disasm_line(
            insn_ea, ida_lines.GENDSM_REMOVE_TAGS
        ) or ""
        print(f"IDA_DISASM {insn_ea:#x} {line}")

    try:
        cfunc = ida_hexrays.decompile(fn.start_ea)
    except Exception as exc:
        print(f"IDA_PSEUDOCODE unavailable: {exc}")
    else:
        print("IDA_PSEUDOCODE_BEGIN")
        print(str(cfunc))
        print("IDA_PSEUDOCODE_END")

idc.qexit(0)
