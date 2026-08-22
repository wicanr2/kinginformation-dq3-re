"""非破壞性匯出沙曼歐莎 monster89 事件與 D3MNS raw record。

IDA batch：
  idat -A '-Stools/ida_dump_monster89.py OUT_JSON D3MNS.DAT' DQ3.EXE

保留原始名稱、位址、bytes 與 xref；不 rename、不寫 comment、不修改輸入。
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


MONSTER_ID = 89
MONSTER_STRIDE = 0x29
REQUESTED = (0x14312, 0x1A973, 0x1BDDF)


def disasm(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def xref_record(x):
    fn = ida_funcs.get_func(x.frm)
    return {
        "from_ida_linear": hex(x.frm),
        "xref_type": int(x.type),
        "function_original": ida_funcs.get_func_name(fn.start_ea) if fn else "<no-func>",
        "function_start_ida_linear": hex(fn.start_ea) if fn else None,
    }


def function_record(ea):
    fn = ida_funcs.get_func(ea)
    if fn is None:
        return {"requested_ida_linear": hex(ea), "error": "no_function", "level": "unknown"}
    return {
        "requested_ida_linear": hex(ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [xref_record(x) for x in idautils.XrefsTo(fn.start_ea)],
        "instructions": [
            {
                "ida_linear": hex(i),
                "bytes_hex": (ida_bytes.get_bytes(i, ida_bytes.get_item_size(i)) or b"").hex(),
                "disassembly": disasm(i),
            }
            for i in idautils.FuncItems(fn.start_ea)
        ],
        "level": "raw_ida_export; semantics_require_review",
    }


def main():
    ida_auto.auto_wait()
    if len(idc.ARGV) != 3:
        raise RuntimeError("需要 OUT_JSON 與 D3MNS.DAT 路徑")
    output = Path(idc.ARGV[1])
    monster_path = Path(idc.ARGV[2])
    exe_path = Path(ida_nalt.get_input_file_path())
    exe = exe_path.read_bytes()
    monsters = monster_path.read_bytes()
    start = MONSTER_ID * MONSTER_STRIDE
    record = monsters[start:start + MONSTER_STRIDE]
    if len(record) != MONSTER_STRIDE:
        raise RuntimeError("monster89 record truncated")
    result = {
        "input": {
            "exe": {"path": str(exe_path), "size": len(exe), "sha256": hashlib.sha256(exe).hexdigest()},
            "monsters": {"path": str(monster_path), "size": len(monsters), "sha256": hashlib.sha256(monsters).hexdigest()},
        },
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "address_contract": {"ida_linear": "logical + 0x10000", "file": "logical + 0x1370"},
        "monster_record": {
            "id": MONSTER_ID,
            "stride": MONSTER_STRIDE,
            "file_offset": hex(start),
            "raw_hex": record.hex(),
            "selected_raw_fields": {
                "hp_base_u16": int.from_bytes(record[0:2], "little"),
                "hp_rand_u16": int.from_bytes(record[2:4], "little"),
                "mp_u16": int.from_bytes(record[4:6], "little"),
                "attack_u8": record[8],
                "defense_u16": int.from_bytes(record[9:11], "little"),
                "agility_u16": int.from_bytes(record[11:13], "little"),
                "action_gate_u8": record[13],
                "action_mask_hex": record[14:20].hex(),
                "flee_threshold_u8": record[23],
                "flee_rate_u8": record[24],
                "exp_u16": int.from_bytes(record[33:35], "little"),
                "gold_u16": int.from_bytes(record[35:37], "little"),
                "drop_rate_u8": record[37],
                "drop_item_u8": record[38],
            },
            "level": "confirmed raw bytes; field semantics depend on cited consumers",
        },
        "functions": {hex(ea): function_record(ea) for ea in REQUESTED},
        "event_raw": {
            "file_start": "0x5682",
            "file_end": "0x5733",
            "raw_hex": exe[0x5682:0x5733].hex(),
            "level": "confirmed raw bytes; branch semantics require function review",
        },
    }
    output.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    if output.stat().st_size == 0:
        raise RuntimeError("empty sidecar")
    ida_kernwin.qexit(0)


main()
