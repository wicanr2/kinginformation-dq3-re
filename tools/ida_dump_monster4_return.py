"""非破壞性匯出 monster4 action 與返航遭遇所需靜態證據。

IDA Pro 9.4 batch：
  idat -A '-Stools/ida_dump_monster4_return.py OUT D3MNS.DAT' DQ3.EXE

只讀原始輸入並輸出 sidecar；不 rename、不修改 database 或 binary。
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


MONSTER_ID = 4
STRIDE = 0x29
REMAP_IDA = 0x28700
HANDLERS_IDA = 0x2871E
DESCRIPTORS_IDA = 0x28593
REQUESTED = (0x199DC, 0x19AD6, 0x19B4C, 0x19BC3, 0x1A1B4)


def sha(data):
    return hashlib.sha256(data).hexdigest()


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


def set_bits(mask):
    return [bit for bit in range(48) if mask[bit // 8] & (0x80 >> (bit % 8))]


def action_for(bit, remap):
    if bit < 16:
        action, source = bit, "direct_bit_0_15"
    elif bit < 18:
        action, source = bit + 2, "bit_plus_2"
    else:
        action, source = remap[bit - 18], "DGROUP_0x3930_remap"
    descriptor_ea = DESCRIPTORS_IDA + action * 3
    out = {
        "mask_bit": bit,
        "action_raw": action,
        "mapping_source": source,
        "descriptor": {
            "dgroup": hex(0x37C3 + action * 3),
            "ida_linear": hex(descriptor_ea),
            "raw_hex": (ida_bytes.get_bytes(descriptor_ea, 3) or b"").hex(),
        },
        "level": "confirmed raw mapping; semantic name requires consumer review",
    }
    if action >= 0x14:
        entry = HANDLERS_IDA + (action - 0x14) * 2
        logical = ida_bytes.get_word(entry)
        out["handler"] = {
            "entry_ida_linear": hex(entry),
            "raw_word_le": hex(logical),
            "ida_linear": hex(0x10000 + logical),
            "logical": hex(logical),
            "file": hex(logical + 0x1370),
        }
        out["handler_function"] = function_record(0x10000 + logical)
    return out


def main():
    ida_auto.auto_wait()
    if len(idc.ARGV) != 3:
        raise RuntimeError("需要 OUT 與 D3MNS.DAT")
    output = Path(idc.ARGV[1])
    monster_path = Path(idc.ARGV[2])
    exe_path = Path(ida_nalt.get_input_file_path())
    exe, monsters = exe_path.read_bytes(), monster_path.read_bytes()
    start = MONSTER_ID * STRIDE
    rec = monsters[start:start + STRIDE]
    if len(rec) != STRIDE:
        raise RuntimeError("monster4 record truncated")
    mask = rec[14:20]
    remap = list(ida_bytes.get_bytes(REMAP_IDA, 30) or b"")
    if len(remap) != 30:
        raise RuntimeError("remap table unavailable")
    bits = set_bits(mask)
    payload = {
        "input": {
            "exe": {"path": str(exe_path), "size": len(exe), "sha256": sha(exe)},
            "monsters": {"path": str(monster_path), "size": len(monsters), "sha256": sha(monsters)},
        },
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "address_contract": {"ida_linear": "logical + 0x10000", "file": "logical + 0x1370"},
        "monster_record": {
            "id": MONSTER_ID,
            "stride": STRIDE,
            "file_offset": hex(start),
            "raw_hex": rec.hex(),
            "selected_raw_fields": {
                "hp_base_u16": int.from_bytes(rec[0:2], "little"),
                "hp_rand_u16": int.from_bytes(rec[2:4], "little"),
                "mp_u16": int.from_bytes(rec[4:6], "little"),
                "attack_u8": rec[8],
                "defense_u16": int.from_bytes(rec[9:11], "little"),
                "agility_u16": int.from_bytes(rec[11:13], "little"),
                "action_gate_u8": rec[13],
                "mask_hex": mask.hex(),
                "set_bits": bits,
                "flee_threshold_u8": rec[23],
                "flee_rate_u8": rec[24],
                "exp_u16": int.from_bytes(rec[33:35], "little"),
                "gold_u16": int.from_bytes(rec[35:37], "little"),
            },
            "actions": [action_for(bit, remap) for bit in bits],
            "level": "confirmed raw bytes; semantic names require reviewed consumers",
        },
        "remap": {
            "dgroup": "0x3930",
            "ida_linear": hex(REMAP_IDA),
            "raw_30_bytes_hex": bytes(remap).hex(),
        },
        "functions": {hex(ea): function_record(ea) for ea in REQUESTED},
        "action42_handler_window": window(0x1A1B4, 0x100),
    }
    output.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    if output.stat().st_size == 0:
        raise RuntimeError("empty sidecar")
    idc.qexit(0)


main()
