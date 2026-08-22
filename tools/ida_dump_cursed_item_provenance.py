"""非破壞性匯出 DQ3 詛咒裝備來源與教會 consumer 證據。

用法（IDA Pro 9.4 batch）：
  idat -A '-Stools/ida_dump_cursed_item_provenance.py OUT ITEM.DAT' DQ3.EXE

輸出保留原始 sub_* 名稱、IDA linear 位址、指令 bytes、xref、原始 ITEM
7-byte records 與推論等級；不 rename、不寫 comment、不修改輸入檔。
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


FUNCTIONS = (
    0x171CC,  # church curse service
    0x17E6B,  # accessory/status transaction
    0x17ED9,  # equipment selector and cursed-word writer
    0x18197,  # post-church equipment recalculation
    0x1EBD8,  # accessory/status recalculation called by sub_17E6B
)
KEY_ADDRESSES = (
    0x17205,  # church first cursed-word test
    0x1725A,  # church paid-removal loop test
    0x17260,  # church writes 0x00ff
    0x17EF5,  # item word strips bits 14/15 before ITEM index
    0x17F04,  # ITEM index * 7
    0x17F09,  # ITEM +4 category byte used by the listing filter
    0x18040,  # ITEM +6 profession mask used by the equipability gate
    0x18044,  # ITEM +4/+5 little-endian metadata word
    0x180A0,  # metadata word & 0x0e00
    0x180A6,  # prepare bit0x4000
    0x180A9,  # writer ORs bit0x4000 into equipped item word
)


def disasm(ea):
    return ida_lines.generate_disasm_line(
        ea, ida_lines.GENDSM_REMOVE_TAGS
    ) or ""


def instruction(ea):
    fn = ida_funcs.get_func(ea)
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(
            ea, ida_bytes.get_item_size(ea)
        ) or b"").hex(),
        "disassembly": disasm(ea),
        "function_original": (
            ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
        ),
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
    }


def xref(x):
    fn = ida_funcs.get_func(x.frm)
    return {
        "from_ida_linear": hex(x.frm),
        "xref_type": int(x.type),
        "function_original": (
            ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
        ),
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
    }


def function(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {
            "requested_ida_linear": hex(ea),
            "level": "unknown",
            "error": "no_function",
        }
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [xref(x) for x in idautils.XrefsTo(fn.start_ea)],
        "instructions": [instruction(i) for i in idautils.FuncItems(fn.start_ea)],
        "level": "raw_ida_export; semantics_require_review",
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 3:
    raise RuntimeError("需要 OUT_JSON 與 ITEM.DAT 兩個參數")

exe_path = Path(ida_nalt.get_input_file_path())
item_path = Path(idc.ARGV[2])
exe = exe_path.read_bytes()
item = item_path.read_bytes()
if len(item) != 128 * 7:
    raise RuntimeError(f"ITEM.DAT 大小錯誤：{len(item)} != 896")

records = []
for item_id in range(128):
    raw = item[item_id * 7:(item_id + 1) * 7]
    metadata_word = raw[4] | raw[5] << 8
    records.append({
        "item_id": hex(item_id),
        "file_offset": hex(item_id * 7),
        "raw_hex": raw.hex(),
        "byte4_category": hex(raw[4]),
        "byte5_flags": hex(raw[5]),
        "metadata_word_le": hex(metadata_word),
        "curse_mask_0x0e00": hex(metadata_word & 0x0E00),
        "writes_item_word_bit_0x4000_when_equipped": bool(
            metadata_word & 0x0E00
        ),
        "level": "confirmed_raw_record; semantics_from_sub_17ED9",
    })

result = {
    "schema": "dq3.ida_cursed_item_provenance.v1",
    "input": {
        "exe_path": str(exe_path),
        "exe_size": len(exe),
        "exe_sha256": hashlib.sha256(exe).hexdigest(),
        "item_path": str(item_path),
        "item_size": len(item),
        "item_sha256": hashlib.sha256(item).hexdigest(),
    },
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": (
            "IDA linear; seg0 logical = linear - 0x10000; "
            "seg0 file = logical + 0x1370"
        ),
    },
    "annotation_contract": {
        "mode": "non_destructive_export",
        "original_names_preserved": True,
        "semantic_levels": (
            "raw instructions and ITEM records are confirmed; "
            "human semantic interpretation requires review"
        ),
    },
    "functions": [function(ea) for ea in FUNCTIONS],
    "key_instructions": [instruction(ea) for ea in KEY_ADDRESSES],
    "exact_xrefs": {
        hex(ea): [xref(x) for x in idautils.XrefsTo(ea)]
        for ea in KEY_ADDRESSES
    },
    "item_records": records,
    "cursed_item_candidates": [
        rec for rec in records
        if rec["writes_item_word_bit_0x4000_when_equipped"]
    ],
}

output_path = Path(idc.ARGV[1])
output_path.write_text(
    json.dumps(result, ensure_ascii=False, indent=2) + "\n",
    encoding="utf-8",
)
if output_path.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
