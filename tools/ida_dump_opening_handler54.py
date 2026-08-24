"""IDA 9.4 batch exporter for the DQ3 opening handler around file 0x147b.

Usage:
  idat -A -c -o/work/opening.i64 \
    "-S/tools/ida_dump_opening_handler54.py /work/opening-handler54-ida.json 0x147b" \
    /input/DQ3.EXE

The output deliberately retains file offsets, IDA linear addresses, raw bytes,
original IDA names, xrefs, and an explicit inference level.  It does not rename
or otherwise annotate the disposable database.
"""

import hashlib
import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_lines
import ida_loader
import ida_nalt
import idautils
import idc


def clean_line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def item_record(ea):
    size = max(1, ida_bytes.get_item_size(ea))
    raw = ida_bytes.get_bytes(ea, size) or b""
    file_offset = ida_loader.get_fileregion_offset(ea)
    refs_from = []
    for dst in idautils.CodeRefsFrom(ea, False):
        refs_from.append({
            "ida_linear": hex(dst),
            "original_name": ida_funcs.get_func_name(dst) or idc.get_name(dst) or "",
        })
    return {
        "ida_linear": hex(ea),
        "file_offset": hex(file_offset) if file_offset >= 0 else None,
        "bytes": raw.hex(),
        "disassembly": clean_line(ea),
        "code_refs_from": refs_from,
    }


def main():
    ida_auto.auto_wait()
    if len(idc.ARGV) < 3:
        raise RuntimeError("output path and target file offset are required")
    output = idc.ARGV[1]
    target_file = int(idc.ARGV[2], 0)
    input_path = ida_nalt.get_input_file_path()
    with open(input_path, "rb") as fh:
        blob = fh.read()

    target_ea = ida_loader.get_fileregion_ea(target_file)
    if target_ea == idc.BADADDR:
        for ea in idautils.Heads(ida_ida.inf_get_min_ea(), ida_ida.inf_get_max_ea()):
            off = ida_loader.get_fileregion_offset(ea)
            size = max(1, ida_bytes.get_item_size(ea))
            if off >= 0 and off <= target_file < off + size:
                target_ea = ea
                break
    if target_ea == idc.BADADDR:
        raise RuntimeError(f"file offset {target_file:#x} is not mapped")

    fn = ida_funcs.get_func(target_ea)
    if fn is None:
        raise RuntimeError(f"no function contains IDA address {target_ea:#x}")

    callers = []
    for ref in idautils.CodeRefsTo(fn.start_ea, False):
        caller = ida_funcs.get_func(ref)
        callers.append({
            "callsite_ida_linear": hex(ref),
            "callsite_file_offset": hex(ida_loader.get_fileregion_offset(ref)),
            "original_function": ida_funcs.get_func_name(caller.start_ea) if caller else "",
            "disassembly": clean_line(ref),
        })

    instructions = [item_record(ea) for ea in idautils.FuncItems(fn.start_ea)]
    result = {
        "evidence_contract": {
            "semantic_annotation": "opening handler containing historical file offset 0x147b",
            "inference_level": "unknown",
            "note": "Navigation label only; interpretation requires caller/writer/consumer review.",
        },
        "input": {
            "path": input_path,
            "size": len(blob),
            "sha256": hashlib.sha256(blob).hexdigest(),
        },
        "tool": {
            "name": "IDA Pro",
            "version": ida_kernwin.get_kernel_version(),
            "address_space": "IDA linear address plus original MZ file offset",
        },
        "target": {
            "requested_file_offset": hex(target_file),
            "ida_linear": hex(target_ea),
            "original_function": ida_funcs.get_func_name(fn.start_ea),
            "function_start_ida_linear": hex(fn.start_ea),
            "function_start_file_offset": hex(ida_loader.get_fileregion_offset(fn.start_ea)),
            "function_end_ida_linear": hex(fn.end_ea),
        },
        "callers": callers,
        "instructions": instructions,
    }
    os.makedirs(os.path.dirname(output), exist_ok=True)
    with open(output, "w", encoding="utf-8") as fh:
        json.dump(result, fh, ensure_ascii=False, indent=2)
        fh.write("\n")
    idc.qexit(0)


try:
    main()
except Exception as exc:
    output = idc.ARGV[1] if len(idc.ARGV) > 1 else "/tmp/opening-handler54-ida-error.json"
    with open(output, "w", encoding="utf-8") as fh:
        json.dump({"error": repr(exc)}, fh, ensure_ascii=False, indent=2)
        fh.write("\n")
    idc.qexit(1)
