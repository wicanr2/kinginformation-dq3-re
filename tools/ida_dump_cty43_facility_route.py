"""非破壞性匯出 CTY43／44 入口、NPC、facility 與 transition 證據。

IDA batch：
  idat -A '-Stools/ida_dump_cty43_facility_route.py OUT CTY43 CTY44' DQ3.EXE

原始 EXE／CTY 僅讀；不 rename、不寫 comment、不修改 database。輸出保留原始位址、
bytes、函式名、輸入 hash 與推論等級，語意仍須人工審查。
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


REQUESTED = (
    0x125FE,  # overworld CTY lookup, file 0x396e
    0x13008,  # load current CTY, file 0x4378
    0x13689,  # section transition consumer, file 0x49f9
    0x1702F,  # facility dispatcher, file 0x839f
)


def u16(raw, off):
    if off < 0 or off + 2 > len(raw):
        raise RuntimeError("u16 out of range: %#x" % off)
    return int.from_bytes(raw[off:off + 2], "little")


def clean_line(ea):
    return ida_lines.generate_disasm_line(ea, ida_lines.GENDSM_REMOVE_TAGS) or ""


def function_record(requested_ea):
    fn = ida_funcs.get_func(requested_ea)
    if fn is None:
        return {
            "requested_ida_linear": hex(requested_ea),
            "error": "no_function",
            "level": "unknown",
        }
    return {
        "requested_ida_linear": hex(requested_ea),
        "function_original": ida_funcs.get_func_name(fn.start_ea),
        "start_ida_linear": hex(fn.start_ea),
        "end_ida_linear": hex(fn.end_ea),
        "xrefs_to": [
            {
                "from_ida_linear": hex(x.frm),
                "xref_type": int(x.type),
                "caller_original": (
                    ida_funcs.get_func_name(ida_funcs.get_func(x.frm).start_ea)
                    if ida_funcs.get_func(x.frm) else "<no-func>"
                ),
            }
            for x in idautils.XrefsTo(fn.start_ea)
        ],
        "instructions": [
            {
                "ida_linear": hex(ea),
                "bytes_hex": (
                    ida_bytes.get_bytes(ea, ida_bytes.get_item_size(ea)) or b""
                ).hex(),
                "disassembly": clean_line(ea),
            }
            for ea in idautils.FuncItems(fn.start_ea)
        ],
        "level": "raw_ida_export; semantics_require_review",
    }


def parse_npc_list(raw, section_base, relative):
    if relative == 0xFFFF:
        return {"relative": hex(relative), "count": 0, "records": []}
    start = section_base + relative
    if start >= len(raw):
        raise RuntimeError("NPC list out of range: %#x" % start)
    count = raw[start]
    end = start + 1 + count * 7
    if end > len(raw):
        raise RuntimeError("NPC list truncated: %#x..%#x" % (start, end))
    records = []
    for index in range(count):
        off = start + 1 + index * 7
        rec = raw[off:off + 7]
        records.append({
            "index": index,
            "file_offset": hex(off),
            "raw_hex": rec.hex(),
            "x": rec[0],
            "y": rec[1],
            "sprite_raw": rec[2],
            "ctrl_raw": rec[3],
            "interaction_subtype": (rec[3] >> 3) & 7,
            "byte4_raw": rec[4],
            "visibility_flag_raw": rec[5],
            "byte6_raw": rec[6],
            "facility_candidate": ((rec[3] >> 3) & 7) >= 3,
            "level": "confirmed raw record; semantic fields cite original consumers",
        })
    return {
        "relative": hex(relative),
        "file_offset": hex(start),
        "count": count,
        "records": records,
    }


def parse_facilities(raw, section_base, relative):
    if relative == 0xFFFF:
        return {"relative": hex(relative), "count": 0, "blocks": []}
    table = section_base + relative
    if table + 2 > len(raw):
        raise RuntimeError("facility table out of range: %#x" % table)
    first_block_rel = u16(raw, table)
    if first_block_rel < relative or (first_block_rel - relative) % 2:
        raise RuntimeError("invalid facility table boundary")
    count = (first_block_rel - relative) // 2
    blocks = []
    for index in range(count):
        block_rel = u16(raw, table + index * 2)
        off = section_base + block_rel
        if off >= len(raw):
            raise RuntimeError("facility block out of range: %#x" % off)
        typ = raw[off]
        if typ == 0:
            size = 3
        elif typ in (1, 2):
            size = 2 + raw[off + 1]
        else:
            size = 1
        blocks.append({
            "index": index,
            "relative": hex(block_rel),
            "file_offset": hex(off),
            "raw_hex": raw[off:off + size].hex(),
            "type_raw": typ,
            "level": "confirmed raw block; type dispatch semantics from IDA 0x1702f",
        })
    return {
        "relative": hex(relative),
        "file_offset": hex(table),
        "count": count,
        "blocks": blocks,
    }


def parse_transitions(raw, section_base, relative, layout_relative):
    if relative == 0xFFFF:
        return {"relative": hex(relative), "count": 0, "records": []}
    start = section_base + relative
    end = section_base + layout_relative
    if start > end or end > len(raw) or (end - start) % 4:
        raise RuntimeError("invalid transition range %#x..%#x" % (start, end))
    return {
        "relative": hex(relative),
        "file_offset": hex(start),
        "count": (end - start) // 4,
        "records": [
            {
                "index": index,
                "file_offset": hex(start + index * 4),
                "raw_hex": raw[start + index * 4:start + (index + 1) * 4].hex(),
                "dest_cty_raw": raw[start + index * 4],
                "dest_section_raw": raw[start + index * 4 + 1],
                "dest_x_raw": raw[start + index * 4 + 2],
                "dest_y_raw": raw[start + index * 4 + 3],
                "level": "confirmed raw transition record",
            }
            for index in range((end - start) // 4)
        ],
    }


def parse_cty(path):
    raw = path.read_bytes()
    section_count = u16(raw, 0) // 2
    sections = []
    for section in range(section_count):
        section_base = u16(raw, section * 2)
        if section_base == 0xFFFF:
            continue
        day_npc_rel = u16(raw, section_base)
        night_npc_rel = u16(raw, section_base + 2)
        facility_rel = u16(raw, section_base + 6)
        transition_rel = u16(raw, section_base + 0x0C)
        layout_rel = u16(raw, section_base + 0x0E)
        layout = section_base + layout_rel
        sections.append({
            "section": section,
            "section_base_file": hex(section_base),
            "header_raw_hex": raw[section_base:section_base + 0x19].hex(),
            "spawn": {"x": raw[section_base + 0x13], "y": raw[section_base + 0x14]},
            "dialogue_bank_raw": raw[section_base + 0x17],
            "map_id_raw": raw[section_base + 0x18],
            "layout": {
                "relative": hex(layout_rel),
                "file_offset": hex(layout),
                "width": u16(raw, layout),
                "height": u16(raw, layout + 2),
            },
            "day_npcs": parse_npc_list(raw, section_base, day_npc_rel),
            "night_npcs": parse_npc_list(raw, section_base, night_npc_rel),
            "facilities": parse_facilities(raw, section_base, facility_rel),
            "transitions": parse_transitions(raw, section_base, transition_rel, layout_rel),
            "level": "confirmed raw CTY structure; reachability requires BLK collision graph",
        })
    return {
        "path": str(path),
        "size": len(raw),
        "sha256": hashlib.sha256(raw).hexdigest(),
        "section_count": section_count,
        "sections": sections,
    }


def main():
    ida_auto.auto_wait()
    if len(idc.ARGV) != 4:
        raise RuntimeError("需要 OUT、CTY43.DAT、CTY44.DAT")
    output = Path(idc.ARGV[1])
    exe_path = Path(ida_nalt.get_input_file_path())
    exe = exe_path.read_bytes()
    result = {
        "schema": "dq3.ida_cty43_facility_route_evidence.v1",
        "input": {
            "exe": {
                "path": str(exe_path),
                "size": len(exe),
                "sha256": hashlib.sha256(exe).hexdigest(),
            },
            "cty43": parse_cty(Path(idc.ARGV[2])),
            "cty44": parse_cty(Path(idc.ARGV[3])),
        },
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version()},
        "address_contract": {
            "ida_linear": "logical + 0x10000",
            "exe_file": "logical + 0x1370",
            "cty_file": "absolute byte offset in named CTY file",
        },
        "functions": {hex(ea): function_record(ea) for ea in REQUESTED},
        "claims": {
            "world_location_collision": {
                "value": "CTY43 and CTY44 share raw world coordinate; original lookup order must be reviewed",
                "level": "strong until function 0x125fe and raw table are reviewed together",
            },
            "facility_reachability": {
                "value": "raw NPC/facility presence does not prove a walkable route",
                "level": "confirmed methodological constraint",
            },
        },
    }
    output.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    if output.stat().st_size == 0:
        raise RuntimeError("empty sidecar")
    idc.qexit(0)


main()
