"""非破壞性匯出 DQ3 戰後掉落與全隊八格物品 writer 證據。"""

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
    0x1C509,  # docs/13 已定位的戰後掉落率／物品選擇區
    0x1684E,  # 戰後掉落 caller 直接呼叫的薄 wrapper／fall-through 入口
    0x16856,  # docs/99 已證實的全隊八格空位 consumer
    0x16898,  # 另一個明確呼叫八格 consumer 的原始入口
)


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
        decompiler = {
            "text": str(ida_hexrays.decompile(fn.start_ea)),
            "level": "decompiler_annotation; verify_against_instructions",
        }
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


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
result = {
    "schema": "dq3.ida_battle_drop_inventory_evidence.v1",
    "input": {
        "path": str(input_path),
        "size": len(raw),
        "sha256": hashlib.sha256(raw).hexdigest(),
    },
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear; logical=linear-0x10000; file=logical+0x1370",
    },
    "annotation_contract": {
        "mode": "non_destructive_export",
        "original_names_preserved": True,
        "semantic_level": "raw_ida_export; semantics_require_data_flow_review",
    },
    "functions": [function_record(ea) for ea in REQUESTED],
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
