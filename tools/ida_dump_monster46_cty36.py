"""非破壞性匯出 monster46 action 與 CTY36 遭遇證據。

IDA 9.4 batch：
  idat -A '-Stools/ida_dump_monster46_cty36.py OUT D3MNS.DAT CTY36.DAT' DQ3.EXE

只輸出原始 bytes、原始函式名、xref、位址與資料雜湊；不 rename、不寫回輸入。
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


MONSTER_ID = 46
MONSTER_STRIDE = 0x29
REMAP_IDA = 0x28700       # DGROUP 0x3930
HANDLERS_IDA = 0x2871E    # DGROUP 0x394e
REQUESTED = (0x199DC, 0x19AD6, 0x19B4C, 0x1A51A, 0x1A973, 0x1BD97, 0x1BDDF)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def line(ea):
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


def set_bits(mask):
    # sub_19AD6 逐 byte 由 0x80 向右測 bit；raw action index 因此是 MSB-first。
    return [bit for bit in range(48) if mask[bit // 8] & (0x80 >> (bit % 8))]


def action_record(bit, remap):
    if bit < 16:
        action = bit
        source = "direct_bit_0_15"
    elif bit < 18:
        action = bit + 2
        source = "bit_plus_2"
    else:
        action = remap[bit - 18]
        source = "DGROUP_0x3930_remap"
    result = {
        "mask_bit": bit,
        "action_raw": hex(action),
        "mapping_source": source,
        "level": "confirmed raw mapping; handler semantics require review",
    }
    if action >= 0x14:
        entry = HANDLERS_IDA + (action - 0x14) * 2
        logical = ida_bytes.get_word(entry)
        result["handler_pointer"] = {
            "entry_ida_linear": hex(entry),
            "raw_word_le": hex(logical),
            "handler_ida_linear": hex(0x10000 + logical),
            "handler_logical": hex(logical),
            "handler_file": hex(logical + 0x1370),
        }
    else:
        result["handler_pointer"] = None
    return result


def cty_sections(raw):
    if len(raw) < 2:
        return []
    first = int.from_bytes(raw[0:2], "little")
    if first < 2 or first % 2:
        return []
    count = first // 2
    offsets = [int.from_bytes(raw[i * 2:i * 2 + 2], "little") for i in range(count)]
    out = []
    for i, start in enumerate(offsets):
        end = offsets[i + 1] if i + 1 < len(offsets) else len(raw)
        header = raw[start:start + 0x15]
        out.append({
            "section": i,
            "file_start": hex(start),
            "file_end": hex(end),
            "size": max(0, end - start),
            "header_0x15_hex": header.hex(),
            "encounter_gate_plus_0x11": header[0x11] if len(header) > 0x11 else None,
            "spawn_plus_0x13_0x14": list(header[0x13:0x15]) if len(header) >= 0x15 else None,
        })
    return out


def main():
    ida_auto.auto_wait()
    if len(idc.ARGV) != 4:
        raise RuntimeError("需要 OUT、D3MNS.DAT、CTY36.DAT")
    output = Path(idc.ARGV[1])
    monster_path = Path(idc.ARGV[2])
    cty_path = Path(idc.ARGV[3])
    exe_path = Path(ida_nalt.get_input_file_path())
    exe = exe_path.read_bytes()
    monsters = monster_path.read_bytes()
    cty = cty_path.read_bytes()
    start = MONSTER_ID * MONSTER_STRIDE
    record = monsters[start:start + MONSTER_STRIDE]
    if len(record) != MONSTER_STRIDE:
        raise RuntimeError("monster46 record truncated")
    mask = record[14:20]
    remap = list(ida_bytes.get_bytes(REMAP_IDA, 30) or b"")
    if len(remap) != 30:
        raise RuntimeError("monster action remap table unavailable")
    payload = {
        "input": {
            "exe": {"path": str(exe_path), "size": len(exe), "sha256": sha(exe)},
            "monsters": {"path": str(monster_path), "size": len(monsters), "sha256": sha(monsters)},
            "cty36": {"path": str(cty_path), "size": len(cty), "sha256": sha(cty)},
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
                "action_mask_hex": mask.hex(),
                "action_mask_set_bits": set_bits(mask),
                "flee_threshold_u8": record[23],
                "flee_rate_u8": record[24],
                "exp_u16": int.from_bytes(record[33:35], "little"),
                "gold_u16": int.from_bytes(record[35:37], "little"),
            },
            "actions": [action_record(bit, remap) for bit in set_bits(mask)],
            "level": "confirmed raw bytes; semantic names require handler review",
        },
        "cty36": {
            "raw_prefix_64_hex": cty[:64].hex(),
            "sections": cty_sections(cty),
            "level": "confirmed raw bytes; encounter consumer is in EXE",
        },
        "remap": {
            "dgroup": "0x3930",
            "ida_linear": hex(REMAP_IDA),
            "raw_30_bytes_hex": bytes(remap).hex(),
        },
        "functions": {hex(ea): function_record(ea) for ea in REQUESTED},
        "selected_handler_windows": {
            hex(0x10000 + ida_bytes.get_word(HANDLERS_IDA + (a["action_raw_int"] - 0x14) * 2)):
                window(0x10000 + ida_bytes.get_word(HANDLERS_IDA + (a["action_raw_int"] - 0x14) * 2), 0x100)
            for a in [
                {**action_record(bit, remap), "action_raw_int": (
                    bit if bit < 16 else bit + 2 if bit < 18 else remap[bit - 18]
                )}
                for bit in set_bits(mask)
            ]
            if a["action_raw_int"] >= 0x14
        },
    }
    output.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    if output.stat().st_size == 0:
        raise RuntimeError("empty sidecar")
    idc.qexit(0)


main()
