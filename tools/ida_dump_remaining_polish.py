"""非破壞性匯出四項收尾未知的 IDA 9.4 原始證據。

只輸出原始函式、交叉參照與 operand 命中；語意與推論等級由版控文件另行審查。
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


REQUESTED = (
    0x16856,  # battle drop party-slot writer; also writes DS:259C name selector
    0x21651,  # file 0x129C1: text control 0xFFFB variable-name consumer
    0x159E4,  # Lancel handler37
    0x1608A,  # Lancel return handler62
    0x16685,  # courage cave handler85
    0x1668C,  # courage cave handler86
    0x125FE,  # overlapping world entrance flag reader
    0x16EDF,  # file 0x824F: flag set variant
    0x16EF4,  # file 0x8264: flag set
    0x16F09,  # file 0x8279: flag get
    0x1834E,  # spell dispatch/table setup containing raw 158
    0x1C6E5,  # battle spell branch containing raw 158
    0x1A973,  # formal actor consumer and spell/physical dispatch
    0x1C425,  # victory/drop display caller
    0x199DC,  # monster raw-action selector/dispatcher
)

OPERAND_WANTS = {
    0x13,    # Lancel completed flag raw id
    0x9C,    # spell record 156
    0x9E,    # spell record 158
    0xD9F,   # D3MNS base 0xD78 + record offset 0x27
    0xFFFB,  # variable-name text control
    *range(0xD78, 0xDA1),  # every direct D3MNS record displacement
}


def disasm(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def instruction(ea):
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b"").hex(),
        "disassembly": disasm(ea),
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "level": "unknown", "error": "no_function"}
    return {
        "requested_ida_linear": hex(ea),
        "original_name": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [
            {
                "from_ida_linear": hex(x.frm),
                "xref_type": int(x.type),
                "instruction": instruction(x.frm),
                "caller_original_name": ida_funcs.get_func_name(
                    ida_funcs.get_func(x.frm).start_ea
                ) if ida_funcs.get_func(x.frm) else "<no-func>",
            }
            for x in idautils.XrefsTo(fn.start_ea)
        ],
        "instructions": [instruction(item) for item in idautils.FuncItems(fn.start_ea)],
        "level": "raw_ida_export; semantics_require_data_flow_review",
    }


def operand_hits():
    hits = []
    for ea in idautils.Heads(ida_ida.inf_get_min_ea(), ida_ida.inf_get_max_ea()):
        if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
            continue
        insn = ida_ua.insn_t()
        if ida_ua.decode_insn(insn, ea) <= 0:
            continue
        matched = []
        for index, operand in enumerate(insn.ops):
            if operand.type == ida_ua.o_void:
                break
            values = {int(operand.value), int(operand.addr)}
            for want in sorted(OPERAND_WANTS.intersection(values)):
                matched.append({"operand_index": index, "value": hex(want)})
        if not matched:
            continue
        fn = ida_funcs.get_func(ea)
        hits.append({
            **instruction(ea),
            "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
            "function_start_ida_linear": hex(fn.start_ea) if fn else None,
            "matches": matched,
        })
    return hits


def setter_call_records(target):
    records = []
    for xref in idautils.XrefsTo(target):
        history = []
        ea = xref.frm
        for _ in range(10):
            ea = idc.prev_head(ea)
            if ea == idc.BADADDR:
                break
            history.append(instruction(ea))
        records.append({
            "target_ida_linear": hex(target),
            "call": instruction(xref.frm),
            "preceding_instructions_reverse_order": history,
            "level": "raw_ida_export; inspect BX provenance before assigning flag semantics",
        })
    return records


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
result = {
    "schema": "dq3.ida_remaining_polish_evidence.v1",
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
    "operand_hits": operand_hits(),
    "flag_setter_calls": setter_call_records(0x16EDF) + setter_call_records(0x16EF4),
}
out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
