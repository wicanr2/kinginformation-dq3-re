"""非破壞性匯出 DQ3 戰鬥道具 selector／owner／command consumer 證據。

IDA Pro 9.4 batch：
  idat -A '-Stools/ida_dump_battle_item.py OUT_JSON ITEM.DAT' DQ3.EXE

輸出保留原始函式名、IDA linear 位址、bytes、CFG、xref 與 operand；
不 rename、不寫 comment、不修改原始 EXE 或 IDA database。
"""

import hashlib
import json
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_gdl
import ida_kernwin
import ida_lines
import ida_nalt
import ida_segment
import ida_ua
import idautils
import idc


ROOTS = (0x1B836, 0x1B3F3, 0x1C1D8)
DGROUP_FILE_BASE = 0x16140
ITEM_STRIDE = 7
OPERAND_NEEDLES = (
    "[si+3ah]", "[si+4bh]", "[si+4ch]", "[di+3ah]", "[di+4bh]",
    "ds:631h", "ds:722h", "ds:726h", "ds:24adh", "ds:259ch",
)


def disasm(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def insn(ea):
    fn = ida_funcs.get_func(ea)
    size = ida_bytes.get_item_size(ea)
    return {
        "ida_linear": hex(ea),
        "bytes_hex": (ida_bytes.get_bytes(ea, size) or b"").hex(),
        "disassembly": disasm(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
    }


def xref_record(x):
    return {
        "from_ida_linear": hex(x.frm),
        "to_ida_linear": hex(x.to),
        "xref_type": int(x.type),
        "instruction": insn(x.frm),
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    boundary_source = "ida_database"
    if fn is None:
        # 間接 near pointer 不會形成 IDA direct xref。只在一次性工作 DB
        # 建立可重生的函式邊界，原始 EXE／既有名稱與位址完全不變。
        ida_ua.create_insn(ea)
        if ida_funcs.add_func(ea):
            fn = ida_funcs.get_func(ea)
            boundary_source = "analysis_added_from_confirmed_near_pointer"
    if fn is None:
        return {"requested_ida_linear": hex(ea), "level": "unknown", "error": "no_function"}
    calls = []
    instructions = []
    for item_ea in idautils.FuncItems(fn.start_ea):
        instructions.append(insn(item_ea))
        for x in idautils.XrefsFrom(item_ea):
            if x.type not in (16, 17):
                continue
            target = ida_funcs.get_func(x.to)
            calls.append({
                "callsite": insn(item_ea),
                "target_ida_linear": hex(x.to),
                "target_function_original": (
                    ida_funcs.get_func_name(target.start_ea) if target else "<no-func>"
                ),
            })
    blocks = []
    for block in ida_gdl.FlowChart(fn):
        blocks.append({
            "start_ida_linear": hex(block.start_ea),
            "end_ida_linear": hex(block.end_ea),
            "successors": [hex(s.start_ea) for s in block.succs()],
            "predecessors": [hex(p.start_ea) for p in block.preds()],
        })
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to_function_start": [xref_record(x) for x in idautils.XrefsTo(fn.start_ea)],
        "calls_from_function": calls,
        "flow_blocks": blocks,
        "instructions": instructions,
        "function_boundary_source": boundary_source,
        "level": "confirmed_raw_ida_export; semantics_require_review",
    }


def direct_callee_starts(root):
    fn = ida_funcs.get_func(root)
    out = set()
    if fn is None:
        return out
    for ea in idautils.FuncItems(fn.start_ea):
        for x in idautils.XrefsFrom(ea):
            if x.type not in (16, 17):
                continue
            target = ida_funcs.get_func(x.to)
            if target is not None:
                out.add(target.start_ea)
    return out


def operand_hits():
    hits = []
    for seg in idautils.Segments():
        for ea in idautils.Heads(seg, idc.get_segm_end(seg)):
            text = disasm(ea).lower()
            if any(needle in text for needle in OPERAND_NEEDLES):
                hits.append(insn(ea))
    return hits


ida_auto.auto_wait()
if len(idc.ARGV) != 3:
    raise RuntimeError("需要 OUT_JSON 與 ITEM.DAT 兩個參數")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
item_path = Path(idc.ARGV[2])
item_raw = item_path.read_bytes()
if len(item_raw) < 128 * ITEM_STRIDE:
    raise RuntimeError("ITEM.DAT 小於 128*7 bytes")

def dgroup_slice(off, size):
    start = DGROUP_FILE_BASE + off
    return {
        "dgroup_offset": hex(off),
        "file_offset": hex(start),
        "size": size,
        "raw_hex": raw[start:start + size].hex(),
        "level": "confirmed_raw_bytes; array boundary/semantics require reviewed consumers",
    }


def pointer_words(off, count, callsite):
    start = DGROUP_FILE_BASE + off
    cs_selector = idc.get_sreg(callsite, "cs")
    code_segment = ida_segment.get_segm_by_sel(cs_selector)
    if code_segment is None:
        raise RuntimeError(
            f"找不到 callsite {callsite:#x} CS selector {cs_selector:#x} 的 segment"
        )
    code_segment_base = ida_segment.get_segm_base(code_segment)
    out = []
    for index in range(count):
        pos = start + index * 2
        word = int.from_bytes(raw[pos:pos + 2], "little")
        out.append({
            "index": index,
            "dgroup_offset": hex(off + index * 2),
            "file_offset": hex(pos),
            "raw_word_hex": raw[pos:pos + 2].hex(),
            "logical_target": hex(word),
            # 這些是 near function pointer：word 是當前 CS 的 offset，不是
            # EXE load base-relative logical address。必須以實際 callsite 所在
            # IDA segment base 換算，否則會錯落到 0x13xxx 的非程式區。
            "callsite_ida_linear": hex(callsite),
            "cs_selector": hex(cs_selector),
            "code_segment_base_ida_linear": hex(code_segment_base),
            "near_offset": hex(word),
            "ida_linear_target": hex(code_segment_base + word) if word else None,
        })
    return out


equipment_handlers = pointer_words(0x3652, 10, 0x1B9E1)
general_handlers = pointer_words(0x366A, 0x80 - 0x41, 0x1B9FA)
function_starts = set(ROOTS)
for root in ROOTS:
    function_starts.update(direct_callee_starts(root))
for entry in equipment_handlers + general_handlers:
    if entry["ida_linear_target"] is not None:
        function_starts.add(int(entry["ida_linear_target"], 16))
# 先在一次性 database 建立間接 handler 邊界，再做 caller/callee 閉包；否則
# 它們的消耗／狀態 consumer 只會留在未解的相對 call operand。
for ea in sorted(function_starts):
    if ida_funcs.get_func(ea) is None:
        ida_ua.create_insn(ea)
        ida_funcs.add_func(ea)
for _ in range(4):
    discovered = set()
    for ea in function_starts:
        discovered.update(direct_callee_starts(ea))
    if discovered.issubset(function_starts):
        break
    function_starts.update(discovered)

result = {
    "schema": "dq3.ida_battle_item.v1",
    "input": {
        "path": str(input_path),
        "size": len(raw),
        "sha256": hashlib.sha256(raw).hexdigest(),
        "item_path": str(item_path),
        "item_size": len(item_raw),
        "item_sha256": hashlib.sha256(item_raw).hexdigest(),
    },
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear; logical=linear-0x10000; file=logical+0x1370",
    },
    "annotation_contract": {
        "mode": "non_destructive_export",
        "original_names_preserved": True,
        "semantic_levels": (
            "raw bytes/xrefs/CFG are confirmed; battle-item meanings require reviewed data flow"
        ),
    },
    "root_functions": [hex(ea) for ea in ROOTS],
    "functions": [function_record(ea) for ea in sorted(function_starts)],
    "operand_hits": operand_hits(),
    "raw_tables": {
        "enemy_target_item_ids_candidate": dgroup_slice(0x3637, 11),
        "party_target_item_ids_candidate": dgroup_slice(0x3642, 6),
        "equipment_handler_item_ids": dgroup_slice(0x3648, 10),
        "equipment_handler_pointers": equipment_handlers,
        "general_item_handler_pointers_0x41_to_0x7f": general_handlers,
    },
    "item_records": [
        {
            "item_raw": item_id,
            "file_offset": hex(item_id * ITEM_STRIDE),
            "raw_hex": item_raw[item_id * ITEM_STRIDE:(item_id + 1) * ITEM_STRIDE].hex(),
            "level": "confirmed_raw_item_record; field semantics require consumer closure",
        }
        for item_id in range(128)
    ],
}

out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("IDA sidecar 輸出為空")
idc.qexit(0)
