"""Export non-destructive matching-decompilation evidence from IDA 9.4.

Entry: docs/25-match-progress.md, GitHub Issue #5.
Usage: idat -A '-Stools/ida_matching_probe.py OUT.json IDA_LINEAR ...' DB.i64
Original names, file bytes, loaded bytes and typed xrefs are preserved.
The export does not identify a compiler or promote inferred gameplay semantics.
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
        raise ValueError("Unexpected original executable identity")
    header = struct.unpack_from("<H", raw, 8)[0] * 16
    base = ida_loader.get_fileregion_ea(header)
    if base < 0 or base % 16:
        raise ValueError("MZ load base is unavailable or unaligned")
    load_segment = base // 16
    reloc_count = struct.unpack_from("<H", raw, 6)[0]
    reloc_table = struct.unpack_from("<H", raw, 24)[0]
    relocations = []
    for n in range(reloc_count):
        off, seg = struct.unpack_from("<HH", raw, reloc_table + 4 * n)
        relocations.append(header + seg * 16 + off)

    ledger_path = Path(__file__).with_name("ida_npc_animation_ledger.json")
    ledger_raw = ledger_path.read_bytes()
    ledger = json.loads(ledger_raw)
    if ledger["input_sha256"] != digest or ledger["input_size"] != len(raw):
        raise ValueError("Reviewed semantic ledger input differs")
    annotations = {int(a["ida_linear"], 16): a for a in ledger["annotations"]}
    semantic_ledgers = [{"path": str(ledger_path), "sha256": hashlib.sha256(ledger_raw).hexdigest()}]
    abi_ledger_path = Path(__file__).with_name("ida_rng_abi_ledger.json")
    if abi_ledger_path.exists():
        abi_raw = abi_ledger_path.read_bytes()
        abi_ledger = json.loads(abi_raw)
        if abi_ledger["input_sha256"] != digest or abi_ledger["input_size"] != len(raw):
            raise ValueError("Reviewed RNG ABI ledger input differs")
        for annotation in abi_ledger["annotations"]:
            ea = int(annotation["ida_linear"], 16)
            if ea in annotations:
                raise ValueError("Duplicate reviewed semantic address")
            annotations[ea] = {**annotation, "review_reference": abi_ledger["review_reference"],
                               "dynamic_evidence": abi_ledger["dynamic_evidence"]}
        semantic_ledgers.append({"path": str(abi_ledger_path), "sha256": hashlib.sha256(abi_raw).hexdigest()})

    def row(ea):
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, ea):
            raise ValueError("Instruction decoding failed at " + hex(ea))
        off = ida_loader.get_fileregion_offset(ea)
        if off < 0 or off + insn.size > len(raw):
            raise ValueError("Instruction has no original file bytes")
        file_bytes = raw[off:off + insn.size]
        expected = bytearray(file_bytes)
        fixups = []
        for location in relocations:
            if off <= location < off + insn.size:
                if location + 2 > off + insn.size:
                    raise ValueError("MZ relocation crosses instruction")
                word = struct.unpack_from("<H", raw, location)[0]
                loaded = (word + load_segment) & 0xFFFF
                struct.pack_into("<H", expected, location - off, loaded)
                fixups.append({"file_offset": hex(location), "original_word": word,
                               "loaded_word": loaded})
        if ida_bytes.get_bytes(ea, insn.size) != expected:
            raise ValueError("IDA bytes differ from original plus MZ relocation")
        result = {
            "ida_linear": hex(ea), "logical": hex(ea - base),
            "file_offset": hex(off), "file_bytes": file_bytes.hex(),
            "loaded_bytes": expected.hex(), "mz_relocations": fixups,
            "original_name": idc.get_name(ea),
            "mnemonic": idc.print_insn_mnem(ea),
            "disassembly": ida_lines.tag_remove(ida_lines.generate_disasm_line(ea, 0) or ""),
            "refs_from": [{"to": hex(x.to), "type": x.type, "iscode": bool(x.iscode)}
                          for x in idautils.XrefsFrom(ea)],
            "refs_to": [{"from": hex(x.frm), "type": x.type, "iscode": bool(x.iscode)}
                        for x in idautils.XrefsTo(ea)],
            "inference_level": "unknown",
            "warning": "Instruction identity is verified; semantic meaning remains unknown",
        }
        if ea in annotations:
            a = annotations[ea]
            if int(a["file_offset"], 16) != off or bytes.fromhex(a["bytes"]) != file_bytes:
                raise ValueError("Reviewed semantic ledger original bytes differ")
            result["annotation"] = a
            result["inference_level"] = a["inference_level"]
            result["warning"] = "" if a["inference_level"] == "confirmed" else "Unconfirmed semantic annotation"
        return result

    targets = [int(arg, 0) for arg in idc.ARGV[2:]]
    if not targets:
        raise ValueError("Explicit IDA linear targets are required")

    def caller_window(reference):
        owner = ida_funcs.get_func(reference.frm)
        addresses = [reference.frm]
        previous = reference.frm
        following = reference.frm
        for _ in range(8):
            previous = idc.prev_head(previous, owner.start_ea if owner else 0)
            if previous == idc.BADADDR:
                break
            if ida_bytes.is_code(ida_bytes.get_full_flags(previous)):
                addresses.append(previous)
        for _ in range(8):
            following = idc.next_head(following, owner.end_ea if owner else reference.frm + 64)
            if following == idc.BADADDR:
                break
            if ida_bytes.is_code(ida_bytes.get_full_flags(following)):
                addresses.append(following)
        return {
            "xref_from": hex(reference.frm), "xref_type": reference.type,
            "original_owner": ida_funcs.get_func_name(owner.start_ea) if owner else None,
            "owner_start": hex(owner.start_ea) if owner else None,
            "owner_end_exclusive": hex(owner.end_ea) if owner else None,
            "instructions": [row(ea) for ea in sorted(set(addresses))],
            "inference_level": "unknown",
            "warning": "Caller window is navigation evidence; register provenance needs review",
        }
    functions = []
    unresolved_targets = []
    seen = set()
    for target in targets:
        fn = ida_funcs.get_func(target)
        if fn is None:
            unresolved_targets.append({
                "requested_ida_linear": hex(target), "boundary_status": "missing",
                "instruction": row(target) if ida_bytes.is_code(ida_bytes.get_full_flags(target)) else None,
                "inference_level": "unknown",
                "warning": "No IDA function boundary; excluded from function matching coverage",
            })
            continue
        if fn.start_ea in seen:
            continue
        seen.add(fn.start_ea)
        items = [ea for ea in idautils.FuncItems(fn.start_ea)
                 if ida_bytes.is_code(ida_bytes.get_full_flags(ea))]
        direct_callers = [reference for reference in idautils.XrefsTo(fn.start_ea) if reference.iscode]
        if len(direct_callers) > 128:
            raise ValueError("Caller export exceeds the explicit 128-window diagnostic budget")
        functions.append({
            "requested_ida_linear": hex(target), "original_name": ida_funcs.get_func_name(fn.start_ea),
            "ida_linear_start": hex(fn.start_ea), "ida_linear_end_exclusive": hex(fn.end_ea),
            "chunks": [[hex(a), hex(b)] for a, b in idautils.Chunks(fn.start_ea)],
            "instructions": [row(ea) for ea in items], "inference_level": "unknown",
            "direct_caller_windows": [caller_window(reference) for reference in direct_callers],
        })
    if not functions:
        raise ValueError("None of the requested targets has an IDA function boundary")

    inventory = []
    for ea in idautils.Functions():
        fn = ida_funcs.get_func(ea)
        inventory.append({"original_name": ida_funcs.get_func_name(ea),
                          "ida_linear_start": hex(ea), "ida_linear_end_exclusive": hex(fn.end_ea),
                          "first_bytes": (ida_bytes.get_bytes(ea, min(12, fn.end_ea - ea)) or b"").hex(),
                          "inference_level": "unknown"})

    decompiler = {"available": False, "results": []}
    try:
        import ida_hexrays
        decompiler["available"] = bool(ida_hexrays.init_hexrays_plugin())
        if decompiler["available"]:
            for fn in functions:
                ea = int(fn["ida_linear_start"], 16)
                try:
                    failure = ida_hexrays.hexrays_failure_t()
                    pseudocode = ida_hexrays.decompile(ea, failure)
                    decompiler["results"].append({"ida_linear": hex(ea), "success": pseudocode is not None,
                                                   "pseudocode": str(pseudocode) if pseudocode else None,
                                                   "error": failure.desc() if pseudocode is None else None})
                except Exception as exc:
                    decompiler["results"].append({"ida_linear": hex(ea), "success": False,
                                                   "error": str(exc)})
    except Exception as exc:
        decompiler["error"] = str(exc)

    cs = struct.unpack_from("<H", raw, 22)[0]
    ip = struct.unpack_from("<H", raw, 20)[0]
    return {
        "schema_version": 1,
        "input": {"path": str(source), "size": len(raw), "sha256": digest},
        "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                 "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest()},
        "address_space": {"kind": "IDA linear", "load_base": hex(base),
                          "header_bytes": header, "file_formula": "file = linear - load_base + header_bytes"},
        "entry": {"cs": hex(cs), "ip": hex(ip), "logical": hex(cs * 16 + ip),
                  "file": hex(header + cs * 16 + ip), "ida_linear": hex(base + cs * 16 + ip)},
        "packer_marker_offsets": {p.decode(): raw.find(p) for p in (b"PKLITE", b"LZEXE", b"UPX!")},
        "packer_inference_level": "unknown",
        "semantic_ledger": {"path": str(ledger_path), "sha256": hashlib.sha256(ledger_raw).hexdigest()},
        "semantic_ledgers": semantic_ledgers,
        "functions": functions, "unresolved_targets": unresolved_targets,
        "inventory": inventory, "decompiler": decompiler,
        "limitations": ["IDA automatic boundaries are navigation evidence, not source-level declarations",
                        "Direct xrefs do not include all indirect reads or writes",
                        "Missing packer markers do not prove an unpacked executable",
                        "No compiler family or exact version is established by this export"],
    }


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
