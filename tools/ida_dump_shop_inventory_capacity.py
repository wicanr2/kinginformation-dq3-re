"""非破壞性匯出 DQ3 商店購買與個人道具容量候選控制流。"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_hexrays
import ida_kernwin
import ida_lines
import ida_nalt
import idautils
import idc


REQUESTED = (
    0x174A4,  # docs/40 type1 handler，file 0x8814
    0x1776F,  # docs/40 type2 handler，file 0x8adf
    0x17747,  # selected-character eight-slot writer
    0x17C31,  # item-shop purchase menu consumer
)


def line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "error": "no_function", "level": "unknown"}
    instructions = []
    for insn in idautils.FuncItems(fn.start_ea):
        instructions.append({
            "ida_linear": hex(insn),
            "bytes_hex": (ida_bytes.get_bytes(insn, ida_bytes.get_item_size(insn)) or b"").hex(),
            "disassembly": line(insn),
        })
    calls = []
    for insn in idautils.FuncItems(fn.start_ea):
        for xref in idautils.XrefsFrom(insn):
            target = ida_funcs.get_func(xref.to)
            if target is None or target.start_ea != xref.to:
                continue
            calls.append({
                "from_ida_linear": hex(insn),
                "to_ida_linear": hex(xref.to),
                "target_original": ida_funcs.get_func_name(xref.to),
                "disassembly": line(insn),
            })
    try:
        decompiler = {"text": str(ida_hexrays.decompile(fn.start_ea)),
                      "level": "decompiler_annotation; verify_against_instructions"}
    except Exception as exc:
        decompiler = {"error": type(exc).__name__, "level": "unknown"}
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "calls": calls,
        "instructions": instructions,
        "decompiler": decompiler,
        "level": "raw_ida_export; semantics_require_data_flow_review",
    }


def window_record(center, radius=0x220):
    instructions = []
    ea = ida_bytes.next_head(center - radius, center + radius)
    while ea != idc.BADADDR and ea < center + radius:
        instructions.append({
            "ida_linear": hex(ea),
            "bytes_hex": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
            "disassembly": line(ea),
        })
        ea = ida_bytes.next_head(ea, center + radius)
    return {"center_ida_linear": hex(center), "instructions": instructions,
            "level": "raw_window; function_boundary_unknown"}


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")
input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
result = {
    "schema": "dq3.ida_shop_inventory_capacity_evidence.v1",
    "input": {"path": str(input_path), "size": len(raw),
              "sha256": hashlib.sha256(raw).hexdigest()},
    "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
             "address_space": "IDA linear; logical=linear-0x10000; file=logical+0x1370"},
    "annotation_contract": {"mode": "non_destructive_export", "original_names_preserved": True,
                            "semantic_level": "raw_ida_export; semantics_require_data_flow_review"},
    "functions": [function_record(ea) for ea in REQUESTED],
    "windows": [window_record(ea) for ea in REQUESTED],
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
