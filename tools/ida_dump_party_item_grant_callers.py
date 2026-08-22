"""非破壞性匯出 DQ3 全隊八格物品 writer 的所有直接 caller。"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_nalt
import idautils
import idc


TARGETS = (0x1684E, 0x16856, 0x16898)


def disasm(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def instruction(ea):
    size = ida_bytes.get_item_size(ea)
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": disasm(ea),
    }


def caller(xref):
    fn = ida_funcs.get_func(xref.frm)
    start = fn.start_ea if fn else None
    window = []
    ea = xref.frm
    for _ in range(10):
        prev = ida_bytes.prev_head(ea, max(0, ea - 32))
        if prev == idaapi.BADADDR or prev == ea:
            break
        ea = prev
    for _ in range(24):
        if ea == idaapi.BADADDR:
            break
        window.append(instruction(ea))
        ea = ida_bytes.next_head(ea, xref.frm + 48)
        if ea == idaapi.BADADDR or ea >= xref.frm + 48:
            break
    return {
        "from_ida_linear": hex(xref.frm),
        "to_ida_linear": hex(xref.to),
        "xref_type": int(xref.type),
        "function_original": ida_funcs.get_func_name(start) if start is not None else "<no-func>",
        "function_start_ida_linear": hex(start) if start is not None else None,
        "call_instruction": instruction(xref.frm),
        "nearby_raw_window": window,
        "level": "raw_ida_export; semantics_require_data_flow_review",
    }


import idaapi

ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
targets = []
for target in TARGETS:
    fn = ida_funcs.get_func(target)
    targets.append({
        "requested_ida_linear": hex(target),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "xrefs_to": [caller(x) for x in idautils.XrefsTo(target)],
    })

result = {
    "schema": "dq3.ida_party_item_grant_callers_evidence.v1",
    "input": {"path": str(input_path), "size": len(raw), "sha256": hashlib.sha256(raw).hexdigest()},
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear; logical=linear-0x10000; file=logical+0x1370",
    },
    "annotation_contract": {
        "mode": "non_destructive_export",
        "original_names_preserved": True,
        "semantic_level": "raw callers only; each transaction requires review",
    },
    "targets": targets,
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
