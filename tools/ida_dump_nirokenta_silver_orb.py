"""非破壞性匯出 CTY64 handler49 銀寶珠交易與共用物品 writer。"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_nalt
import idaapi
import idautils
import idc


TARGET_FUNCTION_EAS = (0x15D70, 0x1684E, 0x16856, 0x16898)


def instruction(ea, function_name):
    size = ida_bytes.get_item_size(ea)
    return {
        "function_original": function_name,
        "ida_linear": hex(ea),
        "file_offset": hex(ea - 0xEC90),
        "bytes_hex": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": ida_lines.generate_disasm_line(
            ea, ida_lines.GENDSM_REMOVE_TAGS
        ) or "",
        "code_refs_from": [hex(ref) for ref in idautils.CodeRefsFrom(ea, False)],
        "data_refs_from": [hex(ref) for ref in idautils.DataRefsFrom(ea)],
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
cty_path = Path("/input/CTY64.DAT")
cty_raw = cty_path.read_bytes()

functions = []
rows = []
for requested_ea in TARGET_FUNCTION_EAS:
    fn = ida_funcs.get_func(requested_ea)
    if fn is None:
        functions.append({"requested_ida_linear": hex(requested_ea), "error": "no function"})
        continue
    name = ida_funcs.get_func_name(fn.start_ea)
    functions.append({
        "requested_ida_linear": hex(requested_ea),
        "original_name": name,
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [hex(x.frm) for x in idautils.XrefsTo(fn.start_ea)],
    })
    ea = ida_bytes.next_head(fn.start_ea - 1, fn.end_ea)
    while ea != idaapi.BADADDR and ea < fn.end_ea:
        rows.append(instruction(ea, name))
        ea = ida_bytes.next_head(ea, fn.end_ea)

result = {
    "schema": "dq3.ida_nirokenta_silver_orb_evidence.v1",
    "inputs": [
        {"path": str(input_path), "size": len(raw), "sha256": hashlib.sha256(raw).hexdigest()},
        {"path": str(cty_path), "size": len(cty_raw), "sha256": hashlib.sha256(cty_raw).hexdigest()},
    ],
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear; logical=linear-0x10000; file=linear-0xec90",
    },
    "annotation_contract": {
        "mode": "non_destructive_export",
        "original_names_preserved": True,
        "semantic_level": "raw instructions/xrefs only; conclusions require reviewed data flow",
    },
    "requested_functions_ida_linear": [hex(ea) for ea in TARGET_FUNCTION_EAS],
    "functions": functions,
    "instructions": rows,
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
