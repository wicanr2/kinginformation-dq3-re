"""非破壞性匯出 DQ3 rec166..171 野外支援咒文 handlers 與共用 consumer。"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_nalt
import ida_ua
import idaapi
import idautils
import idc

HANDLERS = (0x1CCA8, 0x1CCB2, 0x1CCBC, 0x1CD04, 0x1CD0E, 0x1CD18)
CONSUMERS = (0x14685, 0x1469F, 0x14307, 0x1CEAE)


def ensure_function(ea):
    fn = ida_funcs.get_func(ea)
    source = "ida_database"
    if fn is None:
        ida_ua.create_insn(ea)
        if ida_funcs.add_func(ea, idaapi.BADADDR):
            fn = ida_funcs.get_func(ea)
            source = "analysis_added_from_confirmed_near_pointer"
    return fn, source


def insn(ea, name):
    size = ida_bytes.get_item_size(ea)
    return {
        "function_original": name,
        "ida_linear": hex(ea),
        "file_offset": hex(ea - 0xEC90),
        "bytes_hex": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or "",
        "code_refs_from": [hex(x) for x in idautils.CodeRefsFrom(ea, False)],
        "data_refs_from": [hex(x) for x in idautils.DataRefsFrom(ea)],
    }


def function_record(ea):
    fn, source = ensure_function(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "level": "unknown", "error": "no_function"}
    name = ida_funcs.get_func_name(fn.start_ea)
    return {
        "requested_ida_linear": hex(ea),
        "function_original": name,
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "function_boundary_source": source,
        "xrefs_to_function_start": [hex(x.frm) for x in idautils.XrefsTo(fn.start_ea)],
        "instructions": [insn(x, name) for x in idautils.FuncItems(fn.start_ea)],
        "level": "confirmed_raw_ida_export; semantics_require_review",
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
records = []
for rec, handler in zip(range(166, 172), HANDLERS):
    spell_id = rec - 0x79
    descriptor_file = 0x16140 + 0x37C3 + spell_id * 3
    records.append({
        "record_raw": rec,
        "spell_id": spell_id,
        "descriptor_file_offset": hex(descriptor_file),
        "descriptor_raw_hex": raw[descriptor_file:descriptor_file + 3].hex(),
        "handler_ida_linear": hex(handler),
    })

result = {
    "schema": "dq3.ida_field_support.v1",
    "input": {"path": str(input_path), "size": len(raw), "sha256": hashlib.sha256(raw).hexdigest()},
    "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
             "address_space": "IDA linear; logical=linear-0x10000; file=linear-0xec90"},
    "annotation_contract": {"mode": "non_destructive_export", "original_names_preserved": True,
                            "semantic_levels": "raw bytes/xrefs confirmed; meanings require writer-consumer review"},
    "records": records,
    "functions": [function_record(ea) for ea in HANDLERS + CONSUMERS],
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("sidecar 輸出為空")
idc.qexit(0)
