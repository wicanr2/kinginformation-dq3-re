"""非破壞性匯出瑪依拉特殊商店入口、狀態與買賣 consumer。"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_nalt
import ida_segment
import idautils
import idc


def disasm(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def instruction(ea):
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
        "disassembly": disasm(ea),
    }


def function_dump(requested_ea):
    f = ida_funcs.get_func(requested_ea)
    if f is None:
        return {"requested_ea": hex(requested_ea), "error": "no_function"}
    callers = []
    for xref in idautils.CodeRefsTo(f.start_ea, False):
        owner = ida_funcs.get_func(xref)
        callers.append({
            "call_ea": hex(xref),
            "caller_original": idc.get_func_name(owner.start_ea) if owner else "<no-func>",
            "caller_start": hex(owner.start_ea) if owner else None,
        })
    return {
        "requested_ea": hex(requested_ea),
        "function_original": idc.get_func_name(f.start_ea),
        "start": hex(f.start_ea),
        "end": hex(f.end_ea),
        "callers": callers,
        "instructions": [instruction(ea) for ea in idautils.Heads(f.start_ea, f.end_ea)
                         if ida_bytes.is_code(ida_bytes.get_full_flags(ea))],
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

operand_hits = []
for seg_ea in idautils.Segments():
    seg = ida_segment.getseg(seg_ea)
    for ea in idautils.Heads(seg.start_ea, seg.end_ea):
        if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
            continue
        line = disasm(ea)
        if "0B34h" in line or "0B62h" in line:
            row = instruction(ea)
            owner = ida_funcs.get_func(ea)
            row["function_original"] = idc.get_func_name(owner.start_ea) if owner else "<no-func>"
            row["function_start"] = hex(owner.start_ea) if owner else None
            operand_hits.append(row)

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
result = {
    "schema": "dq3.ida_maira_special_shop_evidence.v1",
    "input": {"path": str(input_path), "size": len(raw), "sha256": hashlib.sha256(raw).hexdigest()},
    "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(), "address_space": "linear"},
    "annotation_contract": {"mode": "non_destructive_export", "original_names_preserved": True},
    "functions": [function_dump(ea) for ea in (0x16315, 0x17034, 0x174D8)],
    "handler_table_entry": {
        "handler_raw": 72,
        "ida_linear": "0x28a14",
        "bytes_hex": (ida_bytes.get_bytes(0x28A14, 2) or b"").hex(),
        "near_offset": hex(ida_bytes.get_word(0x28A14)),
        "target_linear": hex(0x10000 + ida_bytes.get_word(0x28A14)),
        "level": "raw_data",
    },
    "operand_hits": operand_hits,
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
