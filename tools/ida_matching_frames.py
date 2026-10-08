"""Export original MZ frames without changing names or database annotations.

Entry: docs/25-match-progress.md. Run IDA 9.4 from the read-only original EXE:
idat -A -oWORK.i64 '-Stools/ida_matching_frames.py OUT.json' DQ3.EXE
Segment extents, selector bases and CALL operands are distinct address spaces.
The result proves static mapping, never a runtime CS/DS value or source unit.
"""

import hashlib
import json
from pathlib import Path
import struct

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_loader
import ida_nalt
import ida_segment
import idautils
import idc


def export():
    ida_auto.auto_wait()
    input_path = ida_nalt.get_input_file_path()
    if not input_path:
        raise ValueError("Database input identity unavailable; use a fresh original-EXE loader")
    source = Path(input_path)
    raw = source.read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if len(raw) != 115282 or digest != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c":
        raise ValueError("Original executable identity differs")
    if ida_kernwin.get_kernel_version() != "9.4":
        raise ValueError("Unreviewed IDA version")
    header = struct.unpack_from("<H", raw, 8)[0] * 16
    load = ida_loader.get_fileregion_ea(header)
    if load != 0x10000:
        raise ValueError("Original address-space basis differs")
    count = struct.unpack_from("<H", raw, 6)[0]
    table = struct.unpack_from("<H", raw, 24)[0]
    relocs = {header + segment * 16 + offset
              for offset, segment in (struct.unpack_from("<HH", raw, table + 4 * n) for n in range(count))}
    segments = []
    calls = []
    for start in idautils.Segments():
        segment = ida_segment.getseg(start)
        base = ida_segment.get_segm_base(segment)
        segments.append({
            "original_name": ida_segment.get_segm_name(segment),
            "ida_linear_start": hex(segment.start_ea),
            "ida_linear_end_exclusive": hex(segment.end_ea),
            "selector": hex(segment.sel), "IDA_segment_base": hex(base),
            "type": segment.type, "bitness": segment.bitness,
            "MZ_relative_frame": hex((base - load) // 16) if base >= load and (base - load) % 16 == 0 else None,
            "inference_level": "confirmed", "scope": "IDA metadata; runtime CS/DS values unknown",
        })
        if segment.type != ida_segment.SEG_CODE:
            continue
        for ea in idautils.Heads(segment.start_ea, segment.end_ea):
            if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)) or idc.print_insn_mnem(ea) != "call":
                continue
            offset = ida_loader.get_fileregion_offset(ea)
            if offset < 0 or raw[offset] != 0x9A:
                continue
            code = raw[offset:offset + 5]
            ip, cs = struct.unpack_from("<HH", code, 1)
            relocated = offset + 3 in relocs
            expected = bytearray(code)
            if relocated:
                struct.pack_into("<H", expected, 3, (cs + load // 16) & 0xFFFF)
            if ida_bytes.get_bytes(ea, 5) != expected:
                raise ValueError("Loaded CALL differs from original plus MZ relocation")
            target = (load if relocated else 0) + cs * 16 + ip
            target_segment = ida_segment.getseg(target)
            refs = [{"to": hex(x.to), "type": x.type, "iscode": bool(x.iscode)}
                    for x in idautils.XrefsFrom(ea)]
            if not any(x["iscode"] and x["type"] == 16 and int(x["to"], 16) == target for x in refs):
                raise ValueError("Original far CALL target differs from typed IDA xref")
            owner = ida_funcs.get_func(ea)
            calls.append({
                "ida_linear": hex(ea), "logical": hex(ea - load),
                "file_offset": hex(offset), "raw_bytes": code.hex(),
                "original_owner": ida_funcs.get_func_name(owner.start_ea) if owner else None,
                "MZ_segment_word_file": hex(offset + 3), "MZ_relocated": relocated,
                "raw_segment": hex(cs), "raw_offset": hex(ip), "target_linear": hex(target),
                "target_original_name": idc.get_name(target),
                "target_IDA_segment_base": hex(ida_segment.get_segm_base(target_segment)) if target_segment else None,
                "frame_matches_segment_base": bool(target_segment and
                    ida_segment.get_segm_base(target_segment) == (load if relocated else 0) + cs * 16),
                "typed_xrefs": refs, "inference_level": "confirmed",
                "scope": "Original operands and IDA mapping; full runtime ABI unknown",
            })
    return {
        "schema_version": 1, "input": {"path": str(source), "size": len(raw), "sha256": digest},
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest()},
        "address_space": {"kind": "IDA linear", "load_base": hex(load), "header_bytes": header,
                          "file_formula": "file = linear - load_base + header_bytes"},
        "segments": segments, "far_calls": calls, "whole_goal_complete": False,
        "limitations": ["IDA segment extents are not original module boundaries",
                        "Indirect calls and runtime CS/DS changes are not covered",
                        "Static frame mapping does not approve a source unit"],
    }


output = Path(idc.ARGV[1])
if output.exists():
    idc.qexit(2)
try:
    result = export()
except Exception as error:
    output.write_text(json.dumps({"error": repr(error)}, indent=2) + "\n")
    idc.qexit(1)
else:
    output.write_text(json.dumps(result, indent=2) + "\n")
    idc.qexit(0)
