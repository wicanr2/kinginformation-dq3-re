"""Non-destructive IDA9.4 video-region evidence; docs/25-match-progress.md.

The region is an IDA segment extent, not a recovered original object boundary.
Original names, instructions, operands and typed references are preserved.
"""

import hashlib
import json
from pathlib import Path
import struct

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_loader
import ida_nalt
import ida_segment
import ida_ua
import idautils
import idc


def export():
    ida_auto.auto_wait()
    source = Path(ida_nalt.get_input_file_path())
    raw = source.read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if len(raw) != 115282 or digest != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c":
        raise ValueError("Original input identity differs")
    if ida_kernwin.get_kernel_version() != "9.4" or ida_loader.get_fileregion_ea(0x1370) != 0x10000:
        raise ValueError("Tool/address basis differs")
    start, end = 0x209CE, 0x20B60
    segment = ida_segment.getseg(start)
    if segment.start_ea != start or segment.end_ea != end or ida_segment.get_segm_base(segment) != 0x209C0:
        raise ValueError("Original video-region mapping differs")

    def refs(ea):
        return {"refs_from": [{"to": hex(x.to), "type": x.type, "iscode": bool(x.iscode)} for x in idautils.XrefsFrom(ea)],
                "refs_to": [{"from": hex(x.frm), "type": x.type, "iscode": bool(x.iscode)} for x in idautils.XrefsTo(ea)]}

    instructions = []
    covered = set()
    functions = []
    for ea in idautils.Heads(start, end):
        if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
            continue
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, ea) or ea + insn.size > end:
            raise ValueError("Instruction crosses the reviewed region")
        offset = ida_loader.get_fileregion_offset(ea)
        code = raw[offset:offset + insn.size]
        if ida_bytes.get_bytes(ea, insn.size) != code:
            raise ValueError("Loaded instruction differs from original bytes")
        instructions.append({"ida_linear": hex(ea), "logical": hex(ea - 0x10000), "file_offset": hex(offset),
                             "original_name": idc.get_name(ea), "file_bytes": code.hex(),
                             "mnemonic": idc.print_insn_mnem(ea),
                             "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0)),
                             "operands": [idc.print_operand(ea, n) for n in range(6) if insn.ops[n].type],
                             "inference_level": "unknown", "warning": "Original identity verified; semantic annotation separate",
                             **refs(ea)})
        covered.update(range(ea, ea + insn.size))
    for ea in idautils.Functions(start, end):
        function = ida_funcs.get_func(ea)
        functions.append({"original_name": ida_funcs.get_func_name(ea), "start": hex(ea),
                          "end_exclusive": hex(function.end_ea),
                          "chunks": [[hex(lo), hex(hi)] for lo, hi in idautils.Chunks(ea)], **refs(ea)})
    gaps = []
    for ea in range(start, end):
        if ea in covered:
            continue
        flags = ida_bytes.get_full_flags(ea)
        offset = ida_loader.get_fileregion_offset(ea)
        gaps.append({"ida_linear": hex(ea), "file_offset": hex(offset), "raw_u8": raw[offset],
                     "raw_flags": flags, "is_data": ida_bytes.is_data(flags),
                     "is_align": ida_bytes.is_align(flags), "is_unknown": ida_bytes.is_unknown(flags),
                     "original_name": idc.get_name(ea), "inference_level": "unknown", **refs(ea)})
    count = struct.unpack_from("<H", raw, 6)[0]
    table = struct.unpack_from("<H", raw, 24)[0]
    relocations = [0x1370 + seg * 16 + off for off, seg in
                   (struct.unpack_from("<HH", raw, table + 4 * n) for n in range(count))]
    file_start = start - 0x10000 + 0x1370
    return {"schema_version": 1, "input": {"path": str(source), "size": len(raw), "sha256": digest},
            "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                     "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest()},
            "address_space": {"kind": "IDA linear", "load_base": "0x10000", "header_bytes": 0x1370,
                              "file_formula": "file = linear - 0x10000 + 0x1370"},
            "region": {"start": hex(start), "end_exclusive": hex(end), "file_start": hex(file_start),
                       "original_segment_name": ida_segment.get_segm_name(segment),
                       "IDA_segment_base": "0x209c0", "MZ_frame": "0x109c", "entry_offset": "0xe",
                       "raw_sha256": hashlib.sha256(raw[file_start:file_start + end - start]).hexdigest()},
            "instructions": instructions, "functions": functions, "uncovered_bytes": gaps,
            "MZ_relocations": [hex(x) for x in relocations if file_start <= x < file_start + end - start],
            "scope": "Static region/entry evidence; hardware meaning and original object ownership reviewed separately"}


output = Path(idc.ARGV[1])
if output.exists():
    idc.qexit(2)
try:
    result = export()
except Exception as error:
    output.write_text(json.dumps({"error": repr(error)}, indent=2) + "\n")
    idc.qexit(1)
else:
    output.write_text(json.dumps(result, indent=2) + "\n")
    idc.qexit(0)
