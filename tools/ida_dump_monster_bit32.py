"""非破壞性匯出 DQ3 monster action-mask bit32 的 remap 與 handler。

IDA 9.4 batch：
  idat -A '-Stools/ida_dump_monster_bit32.py /tmp/monster-bit32.json' DQ3.EXE

只輸出 raw bytes／原始名稱／xref／位址，不 rename、不寫 comment、不修改輸入。
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


BIT = 32
REMAP_IDA = 0x28700       # DGROUP 0x3930
HANDLERS_IDA = 0x2871E    # DGROUP 0x394e
REQUESTED = (0x199DC, 0x19AD6, 0x1A973)


def line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def xref(x):
    fn = ida_funcs.get_func(x.frm)
    return {
        "from_ida_linear": hex(x.frm),
        "xref_type": int(x.type),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
    }


def function(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "level": "unknown", "error": "no_function"}
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [xref(x) for x in idautils.XrefsTo(fn.start_ea)],
        "instructions": [
            {
                "ida_linear": hex(i),
                "bytes_hex": (ida_bytes.get_bytes(i, ida_bytes.get_item_size(i)) or b"").hex(),
                "disassembly": line(i),
            }
            for i in idautils.FuncItems(fn.start_ea)
        ],
        "level": "raw_ida_export; semantics_require_review",
    }


def window(start, size):
    end = start + size
    return {
        "start_ida_linear": hex(start),
        "end_ida_linear": hex(end),
        "raw_bytes_hex": (ida_bytes.get_bytes(start, size) or b"").hex(),
        "instructions": [
            {
                "ida_linear": hex(i),
                "bytes_hex": (ida_bytes.get_bytes(i, ida_bytes.get_item_size(i)) or b"").hex(),
                "disassembly": line(i),
                "is_code": bool(ida_bytes.is_code(ida_bytes.get_flags(i))),
            }
            for i in idautils.Heads(start, end)
        ],
        "level": "raw_ida_export; boundaries_and_semantics_require_review",
    }


def main():
    ida_auto.auto_wait()
    out = Path(idc.ARGV[1]) if len(idc.ARGV) > 1 else Path("/tmp/monster-bit32.json")
    input_path = Path(ida_nalt.get_input_file_path())
    raw = input_path.read_bytes()
    remap = list(ida_bytes.get_bytes(REMAP_IDA, 30) or b"")
    action = remap[BIT - 18]
    handler_entry = HANDLERS_IDA + (action - 0x14) * 2
    handler_logical = ida_bytes.get_word(handler_entry)
    handler_ida = 0x10000 + handler_logical
    payload = {
        "input": {
            "path": str(input_path),
            "size": len(raw),
            "sha256": hashlib.sha256(raw).hexdigest(),
        },
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "address_contract": {
            "ida_linear": "logical + 0x10000",
            "file": "logical + 0x1370 for main load image",
        },
        "selected_bit": BIT,
        "remap": {
            "dgroup": "0x3930",
            "ida_linear": hex(REMAP_IDA),
            "raw_30_bytes_hex": bytes(remap).hex(),
            "index": BIT - 18,
            "action_raw": hex(action),
            "level": "confirmed raw table mapping; handler semantics require review",
        },
        "handler_pointer": {
            "dgroup_table": "0x394e",
            "entry_ida_linear": hex(handler_entry),
            "entry_raw_word_le": hex(handler_logical),
            "handler_ida_linear": hex(handler_ida),
            "handler_logical": hex(handler_logical),
            "handler_file": hex(handler_logical + 0x1370),
            "level": "confirmed raw pointer mapping; handler semantics require review",
        },
        "functions": {hex(ea): function(ea) for ea in REQUESTED},
        "handler_function": function(handler_ida),
        "handler_window": window(handler_ida, 0x100),
    }
    out.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    if out.stat().st_size <= 0:
        raise RuntimeError("empty sidecar")
    ida_kernwin.qexit(0)


main()
