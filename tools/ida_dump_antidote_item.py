"""非破壞性匯出 DQ3 item 0x42 驅毒草的 dispatcher／handler 證據。

IDA 9.4 batch：
  idat -A '-Stools/ida_dump_antidote_item.py /tmp/antidote.json' DQ3.EXE

只讀原始 database／bytes；保留原始位址、函式名、xref、輸入 hash 與推論等級。
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


DISPATCHER = 0x13CCF
TABLE = 0x2843A  # DGROUP 0x366a; file 0x197aa
ITEMS = range(0x41, 0x46)


def line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def xrefs(ea):
    out = []
    for x in idautils.XrefsTo(ea):
        fn = ida_funcs.get_func(x.frm)
        out.append({
            "from_ida_linear": hex(x.frm),
            "xref_type": int(x.type),
            "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
            "function_start_ida_linear": hex(fn.start_ea) if fn else None,
        })
    return out


def function(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "level": "unknown", "error": "no_function"}
    insns = []
    for head in idautils.FuncItems(fn.start_ea):
        insns.append({
            "ida_linear": hex(head),
            "bytes_hex": (ida_bytes.get_bytes(head, ida_bytes.get_item_size(head)) or b"").hex(),
            "disassembly": line(head),
        })
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": xrefs(fn.start_ea),
        "instructions": insns,
        "level": "raw_ida_export; semantics_require_review",
    }


def window(start, end):
    out = []
    for head in idautils.Heads(start, end):
        out.append({
            "ida_linear": hex(head),
            "bytes_hex": (ida_bytes.get_bytes(head, ida_bytes.get_item_size(head)) or b"").hex(),
            "is_code": bool(ida_bytes.is_code(ida_bytes.get_flags(head))),
            "disassembly": line(head),
        })
    return out


ida_auto.auto_wait()
if len(idc.ARGV) != 2:
    raise RuntimeError("需要唯一輸出 JSON 路徑")

input_path = Path(ida_nalt.get_input_file_path())
raw = input_path.read_bytes()
entries = []
handlers = {}
for item in ITEMS:
    entry_ea = TABLE + (item - 0x41) * 2
    logical = ida_bytes.get_word(entry_ea)
    handler_ea = 0x10000 + logical
    entries.append({
        "item_raw": hex(item),
        "table_entry_ida_linear": hex(entry_ea),
        "table_entry_dgroup": hex(0x366A + (item - 0x41) * 2),
        "table_entry_file": hex(0x197AA + (item - 0x41) * 2),
        "raw_word_le": hex(logical),
        "handler_logical": hex(logical),
        "handler_ida_linear": hex(handler_ea),
        "level": "confirmed raw table mapping; handler semantics require review",
    })
    handlers[hex(item)] = function(handler_ea)

result = {
    "schema": "dq3.ida_antidote_item_evidence.v1",
    "input": {"path": str(input_path), "size": len(raw), "sha256": hashlib.sha256(raw).hexdigest()},
    "tool": {
        "name": "IDA Pro",
        "version": ida_kernwin.get_kernel_version(),
        "address_space": "IDA linear; logical = linear - 0x10000; file = logical + 0x1370",
    },
    "dispatcher": function(DISPATCHER),
    "table_entries": entries,
    "handlers": handlers,
    "handler_window": window(0x13D90, 0x13E60),
    "helpers": {
        "sub_14CF9": function(0x14CF9),
        "sub_1469F": function(0x1469F),
        "sub_146EB": function(0x146EB),
        "sub_14685": function(0x14685),
        "sub_14307": function(0x14307),
        "sub_19834": function(0x19834),
    },
    "inference": {
        "level": "unknown pending manual writer-consumer review",
        "note": "No rename/comment is applied; table mapping is raw evidence, names are navigation only.",
    },
}

output = Path(idc.ARGV[1])
output.write_text(json.dumps(result, ensure_ascii=False, indent=2), encoding="utf-8")
if output.stat().st_size == 0:
    raise RuntimeError("sidecar output is empty")
idc.qexit(0)
