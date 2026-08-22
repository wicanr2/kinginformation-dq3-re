"""非破壞性匯出 field item owner、rec421 action 與個人丟棄候選資料流。"""

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


REQUESTED = (0x13CCF, 0x1455D, 0x14CF9, 0x19834)


def line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def xref_record(xref):
    fn = ida_funcs.get_func(xref.frm)
    return {
        "from_ida_linear": hex(xref.frm),
        "to_ida_linear": hex(xref.to),
        "xref_type": int(xref.type),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "bytes_hex": (ida_bytes.get_bytes(xref.frm, ida_bytes.get_item_size(xref.frm)) or b"").hex(),
        "disassembly": line(xref.frm),
    }


def function_record(requested_ea):
    fn = ida_funcs.get_func(requested_ea)
    if fn is None:
        return {"requested_ida_linear": hex(requested_ea), "error": "no_function", "level": "unknown"}
    instructions = []
    calls = []
    for ea in idautils.FuncItems(fn.start_ea):
        instructions.append({
            "ida_linear": hex(ea),
            "bytes_hex": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
            "disassembly": line(ea),
        })
        for xref in idautils.XrefsFrom(ea):
            target = ida_funcs.get_func(xref.to)
            if target is not None and target.start_ea == xref.to:
                calls.append(xref_record(xref))
    try:
        decompiler = {"text": str(ida_hexrays.decompile(fn.start_ea)),
                      "level": "decompiler_annotation; verify_against_instructions"}
    except Exception as exc:
        decompiler = {"error": type(exc).__name__, "level": "unknown"}
    return {
        "requested_ida_linear": hex(requested_ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [xref_record(x) for x in idautils.XrefsTo(fn.start_ea)],
        "calls": calls,
        "instructions": instructions,
        "decompiler": decompiler,
        "level": "raw_ida_export; semantics_require_data_flow_review",
    }


def operand_hits(needles):
    hits = []
    for seg in idautils.Segments():
        for ea in idautils.Heads(seg, idc.get_segm_end(seg)):
            text = line(ea)
            if not any(needle in text.lower() for needle in needles):
                continue
            fn = ida_funcs.get_func(ea)
            hits.append({
                "ida_linear": hex(ea),
                "bytes_hex": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
                "disassembly": text,
                "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
                "function_start_ida_linear": hex(fn.start_ea) if fn else None,
                "level": "candidate_only; requires data_flow_review",
            })
    return hits


def window(start, end):
    return [{
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
        "is_code": bool(ida_bytes.is_code(ida_bytes.get_flags(ea))),
        "disassembly": line(ea),
    } for ea in idautils.Heads(start, end)]


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")
input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
result = {
    "schema": "dq3.ida_field_item_owner_actions_evidence.v1",
    "input": {"path": str(input_path), "size": len(raw),
              "sha256": hashlib.sha256(raw).hexdigest()},
    "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
             "address_space": "IDA linear; logical=linear-0x10000; file=logical+0x1370"},
    "annotation_contract": {"mode": "non_destructive_export", "original_names_preserved": True,
                            "semantic_level": "raw_ida_export; semantics_require_data_flow_review"},
    "functions": [function_record(ea) for ea in REQUESTED],
    "owner_and_action_operand_hits": operand_hits((
        "259ch", "2591h", "62dh", "62fh", "722h", "1a1h", "1a5h",
    )),
    "owner_selection_window": window(0x136C0, 0x138C0),
    "item_selection_window": window(0x138C0, 0x13CD0),
    "item_action_window": window(0x14540, 0x14680),
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
