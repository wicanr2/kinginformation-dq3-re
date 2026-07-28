"""IDA batch script: list analyzed instructions containing an immediate value.

Usage:
  idat -A '-Stools/ida_find_imm.py 0x79' DQ3.EXE

This intentionally trusts only IDA-analyzed code. Pair the result with
tools/re_find_imm.py, which scans every file byte using Capstone and therefore
also finds candidates hidden behind unresolved DOS real-mode jump tables.
"""

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_lines
import ida_ua
import idautils
import idc


ida_auto.auto_wait()
want = int(idc.ARGV[1], 0)
lo, hi = ida_ida.inf_get_min_ea(), ida_ida.inf_get_max_ea()

for ea in idautils.Heads(lo, hi):
    if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
        continue
    insn = ida_ua.insn_t()
    if ida_ua.decode_insn(insn, ea) <= 0:
        continue
    matched = False
    for op in insn.ops:
        if op.type == ida_ua.o_void:
            break
        if op.type == ida_ua.o_imm and op.value == want:
            matched = True
            break
    if not matched:
        continue
    fn = ida_funcs.get_func(ea)
    fname = ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
    line = ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""
    print(f"IDA_IMM ea={ea:#x} func={fname} start={fn.start_ea:#x}" if fn else
          f"IDA_IMM ea={ea:#x} func=<no-func>",
          line)

ida_kernwin.qexit(0)
