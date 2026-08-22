"""非破壞性匯出原版地圖咒文選單、caster 與回復分支。"""

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


TARGET_FUNCTION_EAS = (
    0x1C9DF,
    0x1C9EE,
    0x1CB3C,
    0x1CC4F,
    0x14685,
    0x146EB,
    0x188A9,
    0x1885F,
    0x19834,
    0x1D86D,
    0x1D895,
    0x18CF2,
    0x1E6C9,
)
RAW_CODE_RANGES = (
    (0x14685, 0x147A0),
    (0x1C900, 0x1C9EE),
    (0x1CC4F, 0x1CD80),
)


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

# The five field-heal jump-table entries land exactly at the byte following
# sub_1CB3C.  IDA's initial auto-analysis leaves that target undefined, so add
# a function only in this disposable analysis database.  The exported sidecar
# still preserves the original address and bytes; this is a navigation aid,
# not semantic evidence.
if ida_funcs.get_func(0x1CC4F) is None:
    ida_ua.create_insn(0x1CC4F)
    ida_funcs.add_func(0x1CC4F, idaapi.BADADDR)
    ida_auto.auto_wait()

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()

# Locate the descriptor table from its first three confirmed records rather
# than assuming an IDA segment base.  The raw pattern is rec121..123:
# MP/base/flags = 2/10/0x0d, 6/80/0x0d, 12/180/0x0d.
descriptor_pattern = bytes.fromhex("02 0a 0d 06 50 0d 0c b4 0d")
descriptor_file = raw.find(descriptor_pattern)
if descriptor_file < 0:
    raise RuntimeError("找不到 DS:37C3 descriptor raw pattern")
descriptor_linear = descriptor_file + 0xEC90
dgroup_linear_base = descriptor_linear - 0x37C3

field_descriptors = []
for rec in range(161, 181):
    spell_id = rec - 0x79
    ea = dgroup_linear_base + 0x37C3 + spell_id * 3
    b = ida_bytes.get_bytes(ea, 3) or b""
    field_descriptors.append({
        "rec": rec,
        "spell_id": spell_id,
        "ida_linear": hex(ea),
        "file_offset": hex(ea - 0xEC90),
        "bytes_hex": b.hex(),
        "mp": b[0] if len(b) == 3 else None,
        "base": b[1] if len(b) == 3 else None,
        "flags": b[2] if len(b) == 3 else None,
    })

field_handlers = []
for rec in range(161, 181):
    spell_id = rec - 0x79
    table_index = spell_id - 0x28
    ptr_ea = dgroup_linear_base + 0x38CC + table_index * 2
    raw_word = ida_bytes.get_word(ptr_ea)
    target_linear = 0x10000 + raw_word if raw_word not in (0, 0xFFFF) else None
    fn = ida_funcs.get_func(target_linear) if target_linear is not None else None
    field_handlers.append({
        "rec": rec,
        "spell_id": spell_id,
        "table_index": table_index,
        "pointer_ida_linear": hex(ptr_ea),
        "pointer_word_hex": hex(raw_word),
        "target_ida_linear": hex(target_linear) if target_linear is not None else None,
        "target_original_function": ida_funcs.get_func_name(fn.start_ea) if fn else None,
    })
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

raw_range_rows = []
for start_ea, end_ea in RAW_CODE_RANGES:
    ea = start_ea
    while ea < end_ea:
        size = ida_ua.create_insn(ea)
        if size <= 0:
            raw_range_rows.append({
                "ida_linear": hex(ea),
                "file_offset": hex(ea - 0xEC90),
                "bytes_hex": (ida_bytes.get_bytes(ea, 1) or b"").hex(),
                "disassembly": "<undecoded byte>",
            })
            ea += 1
            continue
        raw_range_rows.append(instruction(ea, "raw_range"))
        ea += size

result = {
    "schema": "dq3.ida_field_heal_evidence.v1",
    "input": {
        "path": str(input_path),
        "size": len(raw),
        "sha256": hashlib.sha256(raw).hexdigest(),
    },
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
    "descriptor_locator": {
        "raw_pattern_hex": descriptor_pattern.hex(),
        "file_offset": hex(descriptor_file),
        "ida_linear": hex(descriptor_linear),
        "dgroup_linear_base": hex(dgroup_linear_base),
    },
    "field_descriptors": field_descriptors,
    "heal_descriptors": [row for row in field_descriptors if 161 <= row["rec"] <= 165],
    "field_handlers": field_handlers,
    "heal_field_handlers": [row for row in field_handlers if 161 <= row["rec"] <= 165],
    "functions": functions,
    "instructions": rows,
    "raw_code_ranges": [
        {"start_ida_linear": hex(start), "end_ida_linear": hex(end)}
        for start, end in RAW_CODE_RANGES
    ],
    "raw_range_instructions": raw_range_rows,
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
