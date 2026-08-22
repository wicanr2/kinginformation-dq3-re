"""非破壞性匯出 handler40 建城商人與物品預存資料流。

IDA Pro 9.4 batch：
  idat -A '-Stools/ida_dump_merchant_founder_inventory.py OUT' DQ3.EXE
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


REQUESTED = (0x15AA2, 0x16CE2, 0x19834)
WINDOWS = ((0x15AA2, 0x180), (0x16CE2, 0x40))


def line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "level": "unknown", "error": "no_function"}
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [
            {
                "from_ida_linear": hex(x.frm),
                "xref_type": int(x.type),
                "caller_original": ida_funcs.get_func_name(ida_funcs.get_func(x.frm).start_ea)
                if ida_funcs.get_func(x.frm) else "<no-func>",
            }
            for x in idautils.XrefsTo(fn.start_ea)
        ],
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
    return {
        "start_ida_linear": hex(start),
        "end_ida_linear": hex(start + size),
        "raw_bytes_hex": (ida_bytes.get_bytes(start, size) or b"").hex(),
        "instructions": [
            {
                "ida_linear": hex(i),
                "bytes_hex": (ida_bytes.get_bytes(i, ida_bytes.get_item_size(i)) or b"").hex(),
                "disassembly": line(i),
                "is_code": bool(ida_bytes.is_code(ida_bytes.get_flags(i))),
            }
            for i in idautils.Heads(start, start + size)
        ],
        "level": "raw_ida_export; boundary_and_semantics_require_review",
    }


def main():
    ida_auto.auto_wait()
    if len(idc.ARGV) != 2:
        raise RuntimeError("需要 OUT")
    output = Path(idc.ARGV[1])
    exe_path = Path(ida_nalt.get_input_file_path())
    exe = exe_path.read_bytes()
    payload = {
        "input": {
            "path": str(exe_path),
            "size": len(exe),
            "sha256": hashlib.sha256(exe).hexdigest(),
        },
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "address_contract": {"ida_linear": "logical + 0x10000", "file": "logical + 0x1370"},
        "functions": {hex(ea): function_record(ea) for ea in REQUESTED},
        "windows": {hex(ea): window(ea, size) for ea, size in WINDOWS},
        "fixed_data": {
            "shared_storage_dgroup": "0x526f",
            "shared_storage_capacity": "0x78 bytes",
            "founder_identity_dgroup": "0x5058",
            "level": "addresses require instruction-level consumer review",
        },
    }
    output.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    if output.stat().st_size == 0:
        raise RuntimeError("empty sidecar")
    idc.qexit(0)


main()
