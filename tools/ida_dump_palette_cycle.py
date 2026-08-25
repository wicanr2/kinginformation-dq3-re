"""非破壞性匯出 VGA palette uploader 的 caller 與候選色槽 writer。

本腳本只在一次性 IDA database 中讀取原始函式、指令、bytes 與 xref；不 rename、
不寫 comment。輸出仍需人工審查，不能因 immediate 2/4/5/10 命中就宣稱色槽交換。
"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_lines
import ida_nalt
import ida_ua
import idautils
import idc


PALETTE_UPLOAD = 0x20A3A
CANDIDATE_IMMEDIATES = {2, 4, 5, 10, 0x10, 0x30, 0x1012}


def disasm(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def instruction(ea):
    size = ida_bytes.get_item_size(ea)
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": disasm(ea),
        "code_refs_from": [hex(x) for x in idautils.CodeRefsFrom(ea, False)],
        "data_refs_from": [hex(x) for x in idautils.DataRefsFrom(ea)],
    }


def function(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "level": "unknown", "error": "no_function"}
    rows = [instruction(head) for head in idautils.FuncItems(fn.start_ea)]
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [hex(x.frm) for x in idautils.XrefsTo(fn.start_ea)],
        "instructions": rows,
        "level": "raw_ida_export; semantics_require_review",
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()

caller_starts = set()
call_sites = []
for xref in idautils.XrefsTo(PALETTE_UPLOAD):
    fn = ida_funcs.get_func(xref.frm)
    if fn:
        caller_starts.add(fn.start_ea)
    call_sites.append({
        "from_ida_linear": hex(xref.frm),
        "xref_type": int(xref.type),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "instruction": instruction(xref.frm),
    })

immediate_hits = []
for ea in idautils.Heads(ida_ida.inf_get_min_ea(), ida_ida.inf_get_max_ea()):
    if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
        continue
    insn = ida_ua.insn_t()
    if ida_ua.decode_insn(insn, ea) <= 0:
        continue
    values = []
    for op in insn.ops:
        if op.type == ida_ua.o_void:
            break
        if op.type == ida_ua.o_imm and op.value in CANDIDATE_IMMEDIATES:
            values.append(op.value)
    if not values:
        continue
    fn = ida_funcs.get_func(ea)
    immediate_hits.append({
        "values": [hex(v) for v in values],
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "instruction": instruction(ea),
        "level": "navigation_only",
    })

result = {
    "schema": "dq3.ida_palette_cycle_evidence.v1",
    "input": {
        "path": str(input_path),
        "size": len(raw),
        "sha256": hashlib.sha256(raw).hexdigest(),
    },
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear; project logical = linear - 0x10000 for seg0 code",
    },
    "annotation_contract": {
        "mode": "non_destructive_export",
        "original_names_preserved": True,
        "semantic_level": "raw instructions/xrefs only; conclusions require reviewed data flow",
    },
    "palette_upload": function(PALETTE_UPLOAD),
    "direct_call_sites": call_sites,
    "direct_caller_functions": [function(ea) for ea in sorted(caller_starts)],
    "candidate_immediate_hits": immediate_hits,
}

out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
