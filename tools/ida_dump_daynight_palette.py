"""非破壞性匯出 DQ3 日夜 clock／palette bank 的 IDA 證據。

用法（IDA 9.4 batch）：
  idat -A '-Stools/ida_dump_daynight_palette.py /tmp/daynight.json' DQ3.EXE

輸出同時保留原始函式名／IDA linear address、disassembly、caller、原始檔案
offset bytes、輸入 SHA-256 與推論等級；不對 IDA database rename 或寫 comment。
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


FUNCTIONS = (0x1ECDC, 0x1EE23, 0x1EE76, 0x1EE9B)
FILE_RANGES = {
    "palette_index_table": (0x18705, 12),
    "palette_load": (0x1004C, 0x120),
    "clock_writer": (0x10193, 0x53),
    "palette_upload_selector": (0x101E6, 0x25),
    "scene_palette_selector": (0x1020B, 0x30),
}


def clean_line(ea):
    return ida_lines.generate_disasm_line(
        ea, ida_lines.GENDSM_REMOVE_TAGS
    ) or ""


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {
            "requested_ida_linear": hex(ea),
            "error": "no_function",
            "level": "unknown",
        }
    callers = []
    for xref in idautils.CodeRefsTo(fn.start_ea, False):
        caller = ida_funcs.get_func(xref)
        callers.append({
            "xref_ida_linear": hex(xref),
            "function_original": (
                ida_funcs.get_func_name(caller.start_ea)
                if caller else "<no-func>"
            ),
            "function_start_ida_linear": (
                hex(caller.start_ea) if caller else None
            ),
        })
    instructions = []
    for insn_ea in idautils.FuncItems(fn.start_ea):
        instructions.append({
            "ida_linear": hex(insn_ea),
            "bytes_hex": (ida_bytes.get_bytes(
                insn_ea, ida_bytes.get_item_size(insn_ea)
            ) or b"").hex(),
            "disassembly": clean_line(insn_ea),
        })
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "callers": callers,
        "instructions": instructions,
        "level": "raw_ida_export; semantics_require_review",
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
file_ranges = {}
for name, (offset, size) in FILE_RANGES.items():
    file_ranges[name] = {
        "file_offset": hex(offset),
        "size": size,
        "raw_hex": raw[offset:offset + size].hex(),
        "level": "confirmed_raw_bytes",
    }

result = {
    "schema": "dq3.ida_daynight_palette_evidence.v1",
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
        "semantic_level": "raw_ida_export; semantics_require_review",
    },
    "functions": [function_record(ea) for ea in FUNCTIONS],
    "file_ranges": file_ranges,
}

output_path = Path(idc.ARGV[1])
output_path.write_text(
    json.dumps(result, ensure_ascii=False, indent=2) + "\n",
    encoding="utf-8",
)
if output_path.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
