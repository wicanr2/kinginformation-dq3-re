"""IDA9.4 complete code/function inventory for the full matching Goal, docs/25.

Run from a new database with pristine input copied into a private Docker output.
This exports original identities; it does not repair boundaries or rename code.
"""

import hashlib
import json
from pathlib import Path
import struct

import ida_auto
import ida_bytes
import ida_funcs
import ida_kernwin
import ida_lines
import ida_loader
import ida_nalt
import ida_segment
import ida_ua
import idautils
import idc


EXPECTED_HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"


def export():
    ida_auto.auto_wait()
    source = Path(ida_nalt.get_input_file_path())
    raw = source.read_bytes()
    digest = hashlib.sha256(raw).hexdigest()
    if len(raw) != 115282 or digest != EXPECTED_HASH:
        raise ValueError("Original input identity differs")
    header = struct.unpack_from("<H", raw, 8)[0] * 16
    base = ida_loader.get_fileregion_ea(header)
    if base < 0 or base % 16:
        raise ValueError("MZ load base is unavailable")
    count = struct.unpack_from("<H", raw, 6)[0]
    table = struct.unpack_from("<H", raw, 24)[0]
    relocations = [header + seg * 16 + off for off, seg in
                   (struct.unpack_from("<HH", raw, table + n * 4) for n in range(count))]
    ledgers = []
    annotations = {}
    for name in ("ida_npc_animation_ledger.json", "ida_rng_abi_ledger.json", "ida_matching_c_ledger.json"):
        path = Path(__file__).with_name(name)
        content = path.read_bytes()
        ledger = json.loads(content)
        if ledger["input_sha256"] != digest or ledger["input_size"] != len(raw):
            raise ValueError("Semantic ledger input differs")
        ledgers.append({"path": str(path), "sha256": hashlib.sha256(content).hexdigest()})
        for item in ledger["annotations"]:
            ea = int(item["ida_linear"], 16)
            if ea in annotations:
                raise ValueError("Duplicate semantic annotation address")
            annotations[ea] = {**item, "ledger_path": str(path)}
    instructions = {}
    file_coverage = set()
    file_owners = {}
    unbacked = []
    segments = []
    for start in idautils.Segments():
        segment = ida_segment.getseg(start)
        end = segment.end_ea
        segments.append({"original_name": ida_segment.get_segm_name(segment),
                         "ida_linear_start": hex(start), "ida_linear_end_exclusive": hex(end),
                         "bitness": segment.bitness, "type": segment.type})
        for ea in idautils.Heads(start, end):
            if not ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
                continue
            instruction = ida_ua.insn_t()
            if not ida_ua.decode_insn(instruction, ea):
                raise ValueError("IDA code head failed to decode: " + hex(ea))
            offset = ida_loader.get_fileregion_offset(ea)
            if offset < 0 or offset + instruction.size > len(raw):
                unbacked.append({"ida_linear": hex(ea), "size": instruction.size,
                                 "inference_level": "unknown", "reason": "No complete original file range"})
                continue
            if ida_loader.get_fileregion_offset(ea + instruction.size - 1) != offset + instruction.size - 1:
                raise ValueError("Instruction spans incompatible file mappings")
            original = raw[offset:offset + instruction.size]
            loaded = bytearray(original)
            fixes = []
            for location in relocations:
                if offset <= location < offset + instruction.size:
                    if location + 2 > offset + instruction.size:
                        raise ValueError("MZ relocation crosses IDA instruction boundary")
                    word = struct.unpack_from("<H", raw, location)[0]
                    value = (word + base // 16) & 0xffff
                    struct.pack_into("<H", loaded, location - offset, value)
                    fixes.append({"file_offset": hex(location), "original_word": word, "IDA_loaded_word": value})
            if ida_bytes.get_bytes(ea, instruction.size) != loaded:
                raise ValueError("IDA loaded bytes differ from original plus actual MZ relocation")
            for address in range(offset, offset + instruction.size):
                file_coverage.add(address)
                file_owners.setdefault(address, []).append(hex(ea))
            row = {"ida_linear": hex(ea), "logical": hex(ea - base), "file_offset": hex(offset),
                   "file_bytes": original.hex(), "loaded_bytes": loaded.hex(), "mz_relocations": fixes,
                   "original_name": idc.get_name(ea), "mnemonic": idc.print_insn_mnem(ea),
                   "disassembly": ida_lines.tag_remove(ida_lines.generate_disasm_line(ea, 0) or ""),
                   "refs_from": [{"to": hex(x.to), "type": x.type, "iscode": bool(x.iscode)} for x in idautils.XrefsFrom(ea)],
                   "refs_to": [{"from": hex(x.frm), "type": x.type, "iscode": bool(x.iscode)} for x in idautils.XrefsTo(ea)],
                   "inference_level": "unknown", "warning": "Verified instruction identity; source language and product semantics unknown"}
            if ea in annotations:
                annotation = annotations[ea]
                if int(annotation["file_offset"], 16) != offset or bytes.fromhex(annotation["bytes"]) != original:
                    raise ValueError("Semantic annotation original bytes differ")
                row["annotation"] = annotation
                row["inference_level"] = annotation["inference_level"]
            instructions[hex(ea)] = row
    functions = []
    membership = {}
    for ea in idautils.Functions():
        function = ida_funcs.get_func(ea)
        items = [hex(address) for address in idautils.FuncItems(ea) if hex(address) in instructions]
        for address in items:
            membership.setdefault(address, []).append(hex(ea))
        functions.append({"original_name": ida_funcs.get_func_name(ea), "ida_linear_start": hex(ea),
                          "ida_linear_end_exclusive": hex(function.end_ea),
                          "original_flags": function.flags,
                          "chunks": [[hex(lo), hex(hi)] for lo, hi in idautils.Chunks(ea)],
                          "instructions": items,
                          "entry_refs_to": [{"from": hex(x.frm), "type": x.type, "iscode": bool(x.iscode)} for x in idautils.XrefsTo(ea)],
                          "inference_level": "unknown",
                          "warning": "Automatic IDA boundary; not yet an approved C/ASM source unit"})
    legacy = Path(__file__).resolve().parents[1] / "docs/data/exe_funcs.json"
    legacy_data = json.loads(legacy.read_text())
    legacy_entries = {base + item["entry"] for item in legacy_data["funcs"]}
    ida_entries = {int(item["ida_linear_start"], 16) for item in functions}
    return {"schema_version": 1, "input": {"path": str(source), "size": len(raw), "sha256": digest},
            "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                     "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest()},
            "address_space": {"kind": "IDA linear", "load_base": hex(base), "header_bytes": header,
                              "file_formula": "file = linear - load_base + header_bytes"},
            "semantic_ledgers": ledgers, "segments": segments, "functions": functions,
            "instructions": instructions, "unbacked_code_heads": unbacked,
            "code_outside_functions": [address for address in instructions if address not in membership],
            "shared_function_instructions": {address: owners for address, owners in membership.items() if len(owners) > 1},
            "overlapping_file_instruction_bytes": {hex(address): owners for address, owners in file_owners.items() if len(owners) > 1},
            "verified_code_file_bytes": len(file_coverage),
            "legacy_comparison": {"source": "docs/data/exe_funcs.json", "source_sha256": hashlib.sha256(legacy.read_bytes()).hexdigest(),
                                  "legacy_entries": len(legacy_entries), "IDA_entries": len(ida_entries),
                                  "shared_entries": len(legacy_entries & ida_entries),
                                  "legacy_only_entries": [hex(ea) for ea in sorted(legacy_entries - ida_entries)],
                                  "IDA_only_entries": [hex(ea) for ea in sorted(ida_entries - legacy_entries)]},
            "scope": "Complete automatic IDA code inventory; unclassified data and boundaries remain open, not full decompilation",
            "whole_EXE_source_complete": False}


output = Path(idc.ARGV[1])
if output.exists():
    idc.qexit(2)
try:
    result = export()
except Exception as error:
    output.write_text(json.dumps({"error": repr(error)}, indent=2) + "\n", encoding="utf-8")
    idc.qexit(1)
else:
    output.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)
