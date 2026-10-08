"""IDA9.4 original-module references; docs/25, non-destructive sidecar only.

The fresh DB keeps its original names, boundaries and operands. Original SDK
module names are metadata aliases, not replacements for IDA identities.
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


HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"


def export():
    ida_auto.auto_wait()
    source = Path(ida_nalt.get_input_file_path())
    raw = source.read_bytes()
    if len(raw) != 115282 or hashlib.sha256(raw).hexdigest() != HASH:
        raise ValueError("Original input identity differs")
    base = ida_loader.get_fileregion_ea(0x1370)
    if base != 0x10000:
        raise ValueError("Original IDA address base differs")
    proof_path = Path(idc.ARGV[2])
    proof = json.loads(proof_path.read_text())
    if proof["input_EXE_sha256"] != HASH or len(proof["results"]) != 2 or not all(r["whole_segment_equal"] for r in proof["results"]):
        raise ValueError("Whole SDK module proof differs")
    library_path = Path("/repo/assets_raw/SBCM.LIB")
    if hashlib.sha256(library_path.read_bytes()).hexdigest() != proof["input_library_sha256"]:
        raise ValueError("Original SDK library identity differs")
    relocations = [0x1370 + seg * 16 + off for off, seg in
                   (struct.unpack_from("<HH", raw, 0x22 + 4 * n) for n in range(struct.unpack_from("<H", raw, 6)[0]))]

    def refs(ea):
        return {"refs_to": [{"from": hex(x.frm), "type": x.type, "iscode": bool(x.iscode)} for x in idautils.XrefsTo(ea)],
                "refs_from": [{"to": hex(x.to), "type": x.type, "iscode": bool(x.iscode)} for x in idautils.XrefsFrom(ea)]}

    def instruction(ea):
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, ea):
            raise ValueError("Cannot decode original instruction")
        offset = ida_loader.get_fileregion_offset(ea)
        if offset < 0 or offset + insn.size > len(raw):
            raise ValueError("No original file bytes")
        return {"ida_linear": hex(ea), "logical": hex(ea - base), "file_offset": hex(offset),
                "original_name": idc.get_name(ea), "file_bytes": raw[offset:offset + insn.size].hex(),
                "loaded_bytes": ida_bytes.get_bytes(ea, insn.size).hex(), "mnemonic": idc.print_insn_mnem(ea),
                "disassembly": ida_lines.tag_remove(idc.generate_disasm_line(ea, 0)),
                "operands": [{"n": op.n, "type": op.type, "dtype": op.dtype, "reg": op.reg,
                              "addr": hex(op.addr), "value": hex(op.value), "offb": op.offb}
                             for op in insn.ops if op.type], **refs(ea)}

    modules = []
    for module in proof["results"]:
        start = int(module["original_EXE_IDA_linear_start"], 16)
        end = start + module["size"]
        offset = int(module["original_EXE_file_start"], 16)
        if hashlib.sha256(raw[offset:offset + module["size"]]).hexdigest() != module["linked_segment_sha256"]:
            raise ValueError("Vendor module hash differs from original range")
        heads = []
        functions = []
        for ea in idautils.Heads(start, end):
            if ida_bytes.is_code(ida_bytes.get_full_flags(ea)):
                row = instruction(ea)
                if ea + len(bytes.fromhex(row["file_bytes"])) > end:
                    raise ValueError("Instruction crosses original module proof")
                heads.append(row)
        for ea in idautils.Functions(start, end):
            f = ida_funcs.get_func(ea)
            functions.append({"original_name": ida_funcs.get_func_name(ea), "ida_linear_start": hex(ea),
                              "ida_linear_end_exclusive": hex(f.end_ea), "raw_flags": f.flags,
                              "chunks": [[hex(lo), hex(hi)] for lo, hi in idautils.Chunks(ea)], **refs(ea)})
        modules.append({"SDK_metadata_name": module["original_module_name"], "SDK_object_sha256": module["original_module_sha256"],
                        "ida_linear_start": hex(start), "ida_linear_end_exclusive": hex(end), "file_start": hex(offset),
                        "whole_original_segment_sha256": module["linked_segment_sha256"], "original_IDA_segment_name": ida_segment.get_segm_name(ida_segment.getseg(start)),
                        "original_MZ_relocations": [hex(x) for x in relocations if offset <= x < offset + module["size"]],
                        "instructions": heads, "functions": functions,
                        "inference_level": "confirmed complete original bytes/SDK metadata mapping; function meaning remains scoped"})
    entry = base + struct.unpack_from("<H", raw, 22)[0] * 16 + struct.unpack_from("<H", raw, 20)[0]
    startup = [instruction(ea) for ea in idautils.Heads(entry, entry + 0x54) if ida_bytes.is_code(ida_bytes.get_full_flags(ea))]
    strings = []
    for offset in (0x1141b, 0x11494, 0x14309):
        ea = offset - 0x1370 + base
        consumers = refs(ea)
        windows = []
        for ref in consumers["refs_to"]:
            caller = int(ref["from"], 16)
            lo = caller
            for _ in range(4):
                previous = ida_bytes.prev_head(lo, 0)
                if previous == idc.BADADDR:
                    break
                lo = previous
            windows.append([instruction(x) for x in idautils.Heads(lo, caller + 24) if ida_bytes.is_code(ida_bytes.get_full_flags(x))])
        strings.append({"file_offset": hex(offset), "ida_linear": hex(ea), **consumers, "consumer_windows": windows,
                        "inference_level": "unknown compiler attribution; raw string consumer query only"})
    return {"schema_version": 1, "input": {"path": str(source), "size": len(raw), "sha256": HASH},
            "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(), "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest()},
            "address_space": {"kind": "IDA linear", "load_base": hex(base), "file_formula": "file = linear - 0x10000 + 0x1370"},
            "SDK_library_sha256": proof["input_library_sha256"], "SDK_linker_proof_sha256": hashlib.sha256(proof_path.read_bytes()).hexdigest(),
            "modules": modules, "MZ_entry": hex(entry), "startup_instructions": startup, "string_consumers": strings,
            "scope": "Non-destructive module/consumer evidence; primary compiler and complete source reconstruction remain unknown"}


output = Path(idc.ARGV[1])
if output.exists():
    idc.qexit(2)
try:
    result = export()
except Exception as error:
    output.write_text(json.dumps({"error": repr(error)}, indent=2) + "\n", encoding="utf-8")
    idc.qexit(1)
else:
    output.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    idc.qexit(0)
