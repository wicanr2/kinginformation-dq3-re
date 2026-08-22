"""非破壞性匯出 DQ3 正式／alternate 戰鬥 action queue 證據。

IDA Pro 9.4 batch：
  idat -A '-Stools/ida_dump_battle_action_queue.py OUT_JSON D3MNS.DAT' DQ3.EXE

輸出保留原始 sub_* 名稱、IDA linear 位址、指令 bytes、xref 與原始
D3MNS records；不 rename、不寫 comment、不修改 database 或輸入檔。
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
    0x18E88,  # third direct caller of the queue builder
    0x18FEE,  # containing function for alternate loc_190D7 runner
    0x1A973,  # enemy actor consumer
    0x1B3F3,  # player actor consumer
    0x1BDDF,  # formal battle entry
    0x1C08B,  # formal round runner
    0x1C34F,  # action queue builder/sorter
)
KEY_ADDRESSES = (
    0x190D7,  # alternate runner label
    0x1C08B,
    0x1C34F,
    0x1A973,
    0x1B3F3,
)
MONSTER_STRIDE = 0x29


def raw_sha(data):
    return hashlib.sha256(data).hexdigest()


def byte_occurrences(data, needle):
    out = []
    start = 0
    while True:
        found = data.find(needle, start)
        if found < 0:
            return out
        out.append(hex(found))
        start = found + 1


def disasm(ea):
    return ida_lines.generate_disasm_line(
        ea, ida_lines.GENDSM_REMOVE_TAGS
    ) or ""


def instruction(ea):
    fn = ida_funcs.get_func(ea)
    size = ida_bytes.get_item_size(ea)
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": disasm(ea),
        "function_original": (
            ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
        ),
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
    }


def xref_record(x):
    fn = ida_funcs.get_func(x.frm)
    return {
        "from_ida_linear": hex(x.frm),
        "xref_type": int(x.type),
        "function_original": (
            ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>"
        ),
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        "instruction": instruction(x.frm),
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {
            "requested_ida_linear": hex(ea),
            "level": "unknown",
            "error": "no_function",
        }
    items = list(idautils.FuncItems(fn.start_ea))
    calls = []
    for item_ea in items:
        for x in idautils.XrefsFrom(item_ea):
            if x.type in (16, 17):  # fl_CF / fl_CN
                target_fn = ida_funcs.get_func(x.to)
                calls.append({
                    "from": instruction(item_ea),
                    "to_ida_linear": hex(x.to),
                    "to_function_original": (
                        ida_funcs.get_func_name(target_fn.start_ea)
                        if target_fn else "<no-func>"
                    ),
                    "xref_type": int(x.type),
                })
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to_function_start": [
            xref_record(x) for x in idautils.XrefsTo(fn.start_ea)
        ],
        "calls_from_function": calls,
        "instructions": [instruction(i) for i in items],
        "level": "confirmed_raw_ida_export; semantics_require_review",
    }


def range_record(start, end):
    return {
        "start_ida_linear": hex(start),
        "end_ida_linear": hex(end),
        "raw_bytes_hex": (ida_bytes.get_bytes(start, end - start) or b"").hex(),
        "instructions": [
            instruction(ea) for ea in idautils.Heads(start, end)
            if ida_bytes.is_code(ida_bytes.get_flags(ea))
        ],
        "level": "confirmed_raw_ida_export; function boundary and semantics require review",
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 3:
    raise RuntimeError("需要 OUT_JSON 與 D3MNS.DAT 兩個參數")

output_path = Path(idc.ARGV[1])
exe_path = Path(ida_nalt.get_input_file_path())
monster_path = Path(idc.ARGV[2])
exe = exe_path.read_bytes()
monsters = monster_path.read_bytes()
if len(monsters) % MONSTER_STRIDE != 0:
    raise RuntimeError(
        "D3MNS.DAT 大小不是 0x29 record stride 的整數倍：{}".format(
            len(monsters)
        )
    )
monster_count = len(monsters) // MONSTER_STRIDE

result = {
    "schema": "dq3.ida_battle_action_queue.v1",
    "input": {
        "exe_path": str(exe_path),
        "exe_size": len(exe),
        "exe_sha256": raw_sha(exe),
        "monster_path": str(monster_path),
        "monster_size": len(monsters),
        "monster_sha256": raw_sha(monsters),
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
            "raw instructions, xrefs and D3MNS bytes are confirmed; "
            "absence/repeat semantics require reviewed control-flow closure"
        ),
    },
    "functions": [function_record(ea) for ea in FUNCTIONS],
    "alternate_runner_range": range_record(0x19074, 0x19130),
    "key_addresses": [instruction(ea) for ea in KEY_ADDRESSES],
    "exact_xrefs": {
        hex(ea): [xref_record(x) for x in idautils.XrefsTo(ea)]
        for ea in KEY_ADDRESSES
    },
    "raw_logical_word_occurrences": {
        "sub_1A973_logical_0xA973_le": byte_occurrences(
            exe, bytes((0x73, 0xA9))
        ),
        "sub_1B3F3_logical_0xB3F3_le": byte_occurrences(
            exe, bytes((0xF3, 0xB3))
        ),
        "level": (
            "confirmed raw byte search; occurrences are navigation evidence, "
            "not automatically typed as function pointers"
        ),
    },
    "monster_records": [
        {
            "monster_id": monster_id,
            "file_offset": hex(monster_id * MONSTER_STRIDE),
            "raw_hex": monsters[
                monster_id * MONSTER_STRIDE:(monster_id + 1) * MONSTER_STRIDE
            ].hex(),
            "level": "confirmed_raw_record; no repeat field semantic asserted",
        }
        for monster_id in range(monster_count)
    ],
}

output_path.write_text(
    json.dumps(result, ensure_ascii=False, indent=2) + "\n",
    encoding="utf-8",
)
if output_path.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
