"""IDA batch script: list analyzed instructions referencing an operand value.

Usage:
  idat -A '-Stools/ida_find_operand.py 0x526c' DQ3.EXE

Unlike ida_find_imm.py, this checks immediate, direct-memory, near, and far
operands.  It is intended for tracing DGROUP variables and code/data
references in 16-bit DOS executables.
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
interesting = {
    ida_ua.o_imm,
    ida_ua.o_mem,
    ida_ua.o_near,
    ida_ua.o_far,
}

for ea in idautils.Heads(lo, hi):
    if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
        continue
    insn = ida_ua.insn_t()
    if ida_ua.decode_insn(insn, ea) <= 0:
        continue
    matched = []
    for index, op in enumerate(insn.ops):
        if op.type == ida_ua.o_void:
            break
        if op.type not in interesting:
            continue
        values = {int(op.value), int(op.addr)}
        if want in values:
            matched.append(index)
    if not matched:
        continue
    fn = ida_funcs.get_func(ea)
    fname = ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
    line = ida_lines.generate_disasm_line(
        ea, ida_lines.GENDSM_REMOVE_TAGS
    ) or ""
    start = f"{fn.start_ea:#x}" if fn else "<none>"
    print(
        f"IDA_OPERAND ea={ea:#x} func={fname} start={start} "
        f"operands={','.join(str(index) for index in matched)} {line}"
    )

ida_kernwin.qexit(0)
