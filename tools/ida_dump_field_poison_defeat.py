"""非破壞性匯出 DQ3 原野傷害致死後的控制流證據。

用法（IDA Pro 9.4 batch）：
  idat -A '-Stools/ida_dump_field_poison_defeat.py /tmp/out.json' DQ3.EXE

此工具只保留原始函式名、位址、bytes、xref 與反編譯文字；語意需由
writer → caller → consumer 的人工資料流審查判定，不修改輸入或 IDA database。
"""

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
    0x193E3,  # formal movement caller
    0x19530,  # per-step status/hazard dispatcher
    0x1964E,  # per-member poison/damage/death writer
    0x1EBD8,  # called immediately after field HP reaches zero
    0x1C5F6,  # first post-death aggregate/side-effect consumer
    0x1BCF2,  # branch selected when BP is non-zero
    0x15002,  # branch selected when BP is zero: pre-message helper
    0x15010,  # branch selected when BP is zero: post-message helper
    0x1ED39,  # branch selected when BP is zero
    0x1C03F,  # branch selected when BP is zero
    0x1C7D9,  # branch selected when BP is zero
    0x1BD97,  # no-new-death continuation
    0x1EF03,  # party condition summary candidate
    0x1527E,  # saved-location writer candidate
    0x159E4,  # alternate fixed saved-location writer candidate
    0x10030,  # new-game saved-location initializer
    0x13008,  # defeat-location reload consumer
    0x115E2,  # record-point post-write persistence consumer
)


def clean_line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def xref_record(xref):
    fn = ida_funcs.get_func(xref.frm)
    return {
        "from_ida_linear": hex(xref.frm),
        "xref_type": int(xref.type),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "disassembly": clean_line(xref.frm),
    }


def decompile_record(fn):
    try:
        text = str(ida_hexrays.decompile(fn.start_ea))
        return {"text": text, "level": "decompiler_annotation; verify_against_instructions"}
    except Exception as exc:
        return {"error": type(exc).__name__, "level": "unknown"}


def function_record(requested_ea):
    fn = ida_funcs.get_func(requested_ea)
    if fn is None:
        return {
            "requested_ida_linear": hex(requested_ea),
            "error": "no_function",
            "level": "unknown",
        }
    instructions = []
    for ea in idautils.FuncItems(fn.start_ea):
        instructions.append({
            "ida_linear": hex(ea),
            "bytes_hex": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
            "disassembly": clean_line(ea),
        })
    return {
        "requested_ida_linear": hex(requested_ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [xref_record(x) for x in idautils.XrefsTo(fn.start_ea)],
        "instructions": instructions,
        "decompiler": decompile_record(fn),
        "level": "raw_ida_export; semantics_require_data_flow_review",
    }


def operand_hits(needles):
    """列出保存位置欄位的直接讀寫候選；命中本身不賦予語意。"""
    hits = []
    for seg_ea in idautils.Segments():
        for ea in idautils.Heads(seg_ea, idc.get_segm_end(seg_ea)):
            line = clean_line(ea)
            if not any(needle in line.lower() for needle in needles):
                continue
            fn = ida_funcs.get_func(ea)
            hits.append({
                "ida_linear": hex(ea),
                "bytes_hex": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
                "disassembly": line,
                "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
                "function_start_ida_linear": hex(fn.start_ea) if fn else None,
                "level": "candidate_only; requires data_flow_review",
            })
    return hits


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
result = {
    "schema": "dq3.ida_field_poison_defeat_evidence.v1",
    "input": {
        "path": str(input_path),
        "size": len(raw),
        "sha256": hashlib.sha256(raw).hexdigest(),
    },
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": (
            "IDA linear; seg0 project logical = linear - 0x10000; "
            "seg0 file = logical + 0x1370"
        ),
    },
    "annotation_contract": {
        "mode": "non_destructive_export",
        "original_names_preserved": True,
        "semantic_level": "raw_ida_export; semantics_require_data_flow_review",
    },
    "functions": [function_record(ea) for ea in REQUESTED],
    "saved_location_operand_hits": operand_hits(
        (
            "4f48h", "4f4ah", "4f4ch", "4f4eh", "4f50h", "4f52h",
            "256ah", "256ch",
        )
    ),
}

output_path = Path(idc.ARGV[1])
output_path.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if output_path.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
