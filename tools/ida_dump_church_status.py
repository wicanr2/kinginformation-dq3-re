"""非破壞性匯出 DQ3 教會解毒／解詛咒／復活的 IDA 證據。

用法（IDA 9.4 batch）：
  idat -A '-Stools/ida_dump_church_status.py /tmp/church-status.json' DQ3.EXE

保留原始函式名、IDA linear 位址、bytes、caller/xref、輸入 SHA-256 與推論等級；
不 rename、不寫 comment，也不修改原始 EXE。
"""

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


REQUESTED = (
    0x17068,  # church dispatcher, file 0x83d8
    0x1712B,  # poison service, file 0x849b
    0x171CC,  # curse service, file 0x853c
    0x1728F,  # revive service, file 0x85ff
    0x199DC,  # enemy spell/special-action selector
    0x19AD6,  # selected action precondition/target preparation
    0x19207,  # one enemy-action source feeding DH/BP
    0x1AB83,  # alternate enemy-action source feeding DH/BP
    0x17E6B,  # item status/equipment bit transaction
    0x17ED9,  # equipment selection transaction
    0x182E4,  # party status aggregation candidate
    0x1964E,  # field movement/status consumer candidate
    0x19530,  # formal movement caller of field status consumer
    0x1EF95,  # field poison damage accumulator helper
    0x1EF03,  # field party status summary candidate
    0x1A701,  # player status/death writer helper
    0x1A973,  # enemy action dispatcher
    0x1C8C6,  # battle party status setup
)
RAW_RANGES = {
    "church_dispatch": (0x83D8, 0xC3),
    "poison_service": (0x849B, 0x101),
    "curse_service": (0x853C, 0xC3),
    "revive_service": (0x85FF, 0xAE),
}
IDA_WINDOWS = {
    "enemy_action_id_remap_table": (0x28700, 0x28730),
    "enemy_action_handler_table": (0x2871E, 0x2877E),
    "poison_status_writer_candidate": (0x1A2E0, 0x1A360),
    "status_reader_13949_candidate": (0x13920, 0x13980),
    "status_reader_13a27_candidate": (0x13A00, 0x13A60),
    "status_reader_13ac7_candidate": (0x13AA0, 0x13B00),
    "battle_status_copy_candidate": (0x1B400, 0x1B550),
    "field_status_consumer_candidate": (0x1EEE0, 0x1EF50),
}


def clean_line(ea):
    return ida_lines.generate_disasm_line(
        ea, ida_lines.GENDSM_REMOVE_TAGS
    ) or ""


def xref_record(xref):
    fn = ida_funcs.get_func(xref.frm)
    return {
        "from_ida_linear": hex(xref.frm),
        "xref_type": int(xref.type),
        "function_original": (
            ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
        ),
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
    }


def exact_xrefs(ea):
    return [xref_record(x) for x in idautils.XrefsTo(ea)]


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
            "bytes_hex": (ida_bytes.get_bytes(
                ea, ida_bytes.get_item_size(ea)
            ) or b"").hex(),
            "disassembly": clean_line(ea),
        })
    return {
        "requested_ida_linear": hex(requested_ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [xref_record(x) for x in idautils.XrefsTo(fn.start_ea)],
        "instructions": instructions,
        "level": "raw_ida_export; semantics_require_review",
    }


def instruction_hits():
    """輸出可能讀寫角色 +0x38 或物品 bit0x4000 的原始指令，供人工審查。

    這只是候選索引，不把字串命中自動升格為語意證據。
    """
    hits = []
    for seg_ea in idautils.Segments():
        seg_end = idc.get_segm_end(seg_ea)
        for ea in idautils.Heads(seg_ea, seg_end):
            line = clean_line(ea)
            lower = line.lower()
            if (
                "38h]" not in lower
                and "4000h" not in lower
                and "394eh" not in lower
            ):
                continue
            fn = ida_funcs.get_func(ea)
            hits.append({
                "ida_linear": hex(ea),
                "bytes_hex": (ida_bytes.get_bytes(
                    ea, ida_bytes.get_item_size(ea)
                ) or b"").hex(),
                "disassembly": line,
                "function_original": (
                    ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
                ),
                "function_start_ida_linear": hex(fn.start_ea) if fn else None,
                "level": "candidate_only; requires data-flow review",
            })
    return hits


def window_record(start, end):
    out = []
    for ea in idautils.Heads(start, end):
        if not ida_bytes.is_code(ida_bytes.get_flags(ea)):
            continue
        fn = ida_funcs.get_func(ea)
        out.append({
            "ida_linear": hex(ea),
            "bytes_hex": (ida_bytes.get_bytes(
                ea, ida_bytes.get_item_size(ea)
            ) or b"").hex(),
            "disassembly": clean_line(ea),
            "function_original": (
                ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
            ),
        })
    return out


def status_bit40_sequences():
    """找出載入角色 +0x38 後短距離測試 bit0x40 的候選資料流。"""
    out = []
    for seg_ea in idautils.Segments():
        seg_end = idc.get_segm_end(seg_ea)
        heads = list(idautils.Heads(seg_ea, seg_end))
        for pos, ea in enumerate(heads):
            line = clean_line(ea).lower()
            if "38h]" not in line:
                continue
            window = []
            has_bit40 = False
            for next_ea in heads[pos:pos + 12]:
                next_line = clean_line(next_ea)
                window.append({
                    "ida_linear": hex(next_ea),
                    "bytes_hex": (ida_bytes.get_bytes(
                        next_ea, ida_bytes.get_item_size(next_ea)
                    ) or b"").hex(),
                    "disassembly": next_line,
                })
                if "40h" in next_line.lower() or "0bfh" in next_line.lower():
                    has_bit40 = True
            if has_bit40:
                fn = ida_funcs.get_func(ea)
                out.append({
                    "start_ida_linear": hex(ea),
                    "function_original": (
                        ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
                    ),
                    "instructions": window,
                    "level": "candidate_only; requires control/data-flow review",
                })
    return out


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
result = {
    "schema": "dq3.ida_church_status_evidence.v1",
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
        "semantic_level": "raw_ida_export; semantics_require_review",
    },
    "functions": [function_record(ea) for ea in REQUESTED],
    "exact_xrefs": {
        "poison_gas_entry_0x1a307": exact_xrefs(0x1A307),
        "poison_writer_0x1a335": exact_xrefs(0x1A335),
    },
    "instruction_hits": instruction_hits(),
    "status_bit40_sequences": status_bit40_sequences(),
    "ida_windows": {
        name: {
            "start_ida_linear": hex(start),
            "end_ida_linear": hex(end),
            "instructions": window_record(start, end),
            "level": "candidate_only; semantics_require_data-flow_review",
        }
        for name, (start, end) in IDA_WINDOWS.items()
    },
    "ida_data_ranges": {
        name: {
            "start_ida_linear": hex(start),
            "end_ida_linear": hex(end),
            "raw_hex": (ida_bytes.get_bytes(start, end - start) or b"").hex(),
            "level": "confirmed_raw_bytes; semantics_require_consumer_review",
        }
        for name, (start, end) in {
            "enemy_action_id_remap_table": (0x28700, 0x28730),
            "enemy_action_handler_table": (0x2871E, 0x2877E),
        }.items()
    },
    "raw_ranges": {
        name: {
            "file_offset": hex(offset),
            "size": size,
            "raw_hex": raw[offset:offset + size].hex(),
            "level": "confirmed_raw_bytes",
        }
        for name, (offset, size) in RAW_RANGES.items()
    },
}

output_path = Path(idc.ARGV[1])
output_path.write_text(
    json.dumps(result, ensure_ascii=False, indent=2) + "\n",
    encoding="utf-8",
)
if output_path.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
