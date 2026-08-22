"""非破壞性匯出 DQ3 怪物 action selector、remap 與間接 handlers。

IDA Pro 9.4 batch:
  idat -A '-Stools/ida_dump_monster_action_handlers.py OUT_JSON D3MNS.DAT' DQ3.EXE
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

ROOTS = (0x199DC, 0x19AD6, 0x1A973)
DGROUP_FILE_BASE = 0x16140
REMAP_OFF = 0x3930
HANDLER_OFF = 0x394E
HANDLER_CALLSITE = 0x19AD1


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


def ensure_function(ea):
    fn = ida_funcs.get_func(ea)
    source = "ida_database"
    if fn is None:
        ida_ua.create_insn(ea)
        if ida_funcs.add_func(ea):
            fn = ida_funcs.get_func(ea)
            source = "analysis_added_from_confirmed_near_pointer"
    return fn, source


def direct_callees(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return set()
    out = set()
    for item in idautils.FuncItems(fn.start_ea):
        for x in idautils.XrefsFrom(item):
            if x.type in (16, 17):
                target = ida_funcs.get_func(x.to)
                if target is not None:
                    out.add(target.start_ea)
    return out


def function_record(ea):
    fn, source = ensure_function(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "level": "unknown", "error": "no_function"}
    instructions = [insn(x) for x in idautils.FuncItems(fn.start_ea)]
    blocks = []
    for block in ida_gdl.FlowChart(fn):
        blocks.append({
            "start_ida_linear": hex(block.start_ea),
            "end_ida_linear": hex(block.end_ea),
            "successors": [hex(x.start_ea) for x in block.succs()],
            "predecessors": [hex(x.start_ea) for x in block.preds()],
        })
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "function_boundary_source": source,
        "instructions": instructions,
        "flow_blocks": blocks,
        "xrefs_to_function_start": [
            {"from_ida_linear": hex(x.frm), "xref_type": int(x.type), "instruction": insn(x.frm)}
            for x in idautils.XrefsTo(fn.start_ea)
        ],
        "level": "confirmed_raw_ida_export; semantics_require_review",
    }


ida_auto.auto_wait()
if len(idc.ARGV) != 3:
    raise RuntimeError("需要 OUT_JSON 與 D3MNS.DAT")

exe_path = Path(ida_nalt.get_input_file_path())
exe = exe_path.read_bytes()
mns_path = Path(idc.ARGV[2])
mns = mns_path.read_bytes()
if len(mns) % 0x29 != 0:
    raise RuntimeError("D3MNS.DAT 不是 0x29-byte records")

remap = exe[DGROUP_FILE_BASE + REMAP_OFF:DGROUP_FILE_BASE + REMAP_OFF + 30]
cs_selector = idc.get_sreg(HANDLER_CALLSITE, "cs")
seg = ida_segment.get_segm_by_sel(cs_selector)
if seg is None:
    raise RuntimeError("找不到 handler callsite CS segment")
cs_base = ida_segment.get_segm_base(seg)

handler_entries = []
handler_starts = set()
for index in range(40):
    pos = DGROUP_FILE_BASE + HANDLER_OFF + index * 2
    word = int.from_bytes(exe[pos:pos + 2], "little")
    target = cs_base + word if word else None
    handler_entries.append({
        "action_raw": 0x14 + index,
        "dgroup_offset": hex(HANDLER_OFF + index * 2),
        "file_offset": hex(pos),
        "raw_word_hex": exe[pos:pos + 2].hex(),
        "near_offset": hex(word),
        "cs_selector": hex(cs_selector),
        "cs_base_ida_linear": hex(cs_base),
        "ida_linear_target": hex(target) if target else None,
    })
    if target:
        handler_starts.add(target)

starts = set(ROOTS) | handler_starts
for ea in sorted(starts):
    ensure_function(ea)
for _ in range(5):
    found = set()
    for ea in starts:
        found.update(direct_callees(ea))
    if found.issubset(starts):
        break
    starts.update(found)

used_bits = {}
for monster in range(len(mns) // 0x29):
    record = mns[monster * 0x29:(monster + 1) * 0x29]
    for bit in range(48):
        if record[0x0E + bit // 8] & (0x80 >> (bit % 8)):
            used_bits.setdefault(bit, []).append(monster)


def action_for_bit(bit):
    if bit < 16:
        return bit
    if bit < 18:
        return bit + 2
    return remap[bit - 18]


result = {
    "schema": "dq3.ida_monster_action_handlers.v1",
    "input": {
        "path": str(exe_path), "size": len(exe), "sha256": hashlib.sha256(exe).hexdigest(),
        "d3mns_path": str(mns_path), "d3mns_size": len(mns),
        "d3mns_sha256": hashlib.sha256(mns).hexdigest(),
    },
    "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
             "address_space": "IDA linear; logical=linear-0x10000; file=logical+0x1370"},
    "annotation_contract": {"mode": "non_destructive_export", "original_names_preserved": True,
                            "semantic_levels": "raw bytes/xrefs/CFG confirmed; meanings require review"},
    "remap": {"dgroup_offset": hex(REMAP_OFF),
              "file_offset": hex(DGROUP_FILE_BASE + REMAP_OFF), "raw_hex": remap.hex(),
              "level": "confirmed_raw_bytes_and_sub_199DC_consumer"},
    "handler_entries": handler_entries,
    "used_mask_bits": [
        {"mask_bit": bit, "action_raw": action_for_bit(bit), "monster_raw_ids": monsters,
         "level": "confirmed_D3MNS_bytes_plus_selector_remap"}
        for bit, monsters in sorted(used_bits.items())
    ],
    "functions": [function_record(ea) for ea in sorted(starts)],
}

# 角色／active enemy 狀態多由 SI/DI 間接取址，IDA direct xref 不會列出欄位讀端。
# 以 operand 文字只作候選索引，仍同列保留原始位址、bytes、函式與反組譯；語意需人工閉環。
operand_needles = ("+38h]", "+4dh]", "+12h]", "+233fh]")
status_operand_hits = []
status_function_starts = set()
for fn_ea in idautils.Functions():
    for ea in idautils.FuncItems(fn_ea):
        line = disasm(ea).lower()
        if any(needle in line for needle in operand_needles):
            status_operand_hits.append(insn(ea))
            status_function_starts.add(fn_ea)
result["status_operand_hits"] = {
    "level": "confirmed_operand_index; semantics_require_writer_consumer_review",
    "note": "indirect field references are candidates, not direct xrefs or semantic proof",
    "entries": status_operand_hits,
}
result["status_consumer_functions"] = [
    function_record(ea) for ea in sorted(status_function_starts)
]

out = Path(idc.ARGV[1])
out.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
if out.stat().st_size == 0:
    raise RuntimeError("sidecar 輸出為空")
idc.qexit(0)
