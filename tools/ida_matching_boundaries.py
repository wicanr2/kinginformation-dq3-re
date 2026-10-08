"""IDA9.4 non-destructive boundary/indirect-entry evidence; entry in docs/25.

Usage: idat -A '-S... OUT.json FULL_INVENTORY.json' NEW_DB.i64
Never repair names or boundaries in the database. Preserve raw flag values and
the actual IDA constants so an offline review does not guess flag semantics.
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
import ida_ua
import idautils
import idc


HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"


def export():
    ida_auto.auto_wait()
    source = Path(ida_nalt.get_input_file_path())
    raw = source.read_bytes()
    if hashlib.sha256(raw).hexdigest() != HASH or len(raw) != 115282:
        raise ValueError("Original input identity differs")
    inventory_path = Path(idc.ARGV[2])
    inventory = json.loads(inventory_path.read_text(encoding="utf-8"))
    if inventory["input"]["sha256"] != HASH:
        raise ValueError("Inventory input differs")
    header = struct.unpack_from("<H", raw, 8)[0] * 16
    base = ida_loader.get_fileregion_ea(header)
    if base != int(inventory["address_space"]["load_base"], 16):
        raise ValueError("Inventory/IDA address bases differ")

    def references(ea):
        return {"refs_to": [{"from": hex(x.frm), "type": x.type, "iscode": bool(x.iscode)} for x in idautils.XrefsTo(ea)],
                "refs_from": [{"to": hex(x.to), "type": x.type, "iscode": bool(x.iscode)} for x in idautils.XrefsFrom(ea)]}

    def mapped_bytes(ea, size):
        offset = ida_loader.get_fileregion_offset(ea)
        if offset < 0 or offset + size > len(raw):
            raise ValueError("No original file range")
        return {"ida_linear": hex(ea), "logical": hex(ea - base), "file_offset": hex(offset),
                "file_bytes": raw[offset:offset + size].hex(), "loaded_bytes": ida_bytes.get_bytes(ea, size).hex()}

    constants = {name: getattr(ida_funcs, name) for name in
                 ("FUNC_NORET", "FUNC_LIB", "FUNC_THUNK", "FUNC_TAIL", "FUNC_FRAME", "FUNC_BOTTOMBP")
                 if hasattr(ida_funcs, name)}
    boundaries = []
    functions = []
    by_start = {f["ida_linear_start"]: f for f in inventory["functions"]}
    for function in inventory["functions"]:
        if not function["instructions"]:
            continue
        last = inventory["instructions"][function["instructions"][-1]]
        if last["mnemonic"] in ("ret", "retn", "retf", "iret", "iretd", "jmp"):
            continue
        start = int(function["ida_linear_start"], 16)
        native = ida_funcs.get_func(start)
        if native is None or native.start_ea != start or native.end_ea != int(function["ida_linear_end_exclusive"], 16):
            raise ValueError("Fresh IDA boundary differs from inventory")
        ea = int(last["ida_linear"], 16)
        next_ea = ea + len(bytes.fromhex(last["file_bytes"]))
        next_function = ida_funcs.get_func(next_ea)
        boundaries.append({"original_name": function["original_name"], "ida_linear_start": hex(start),
                           "ida_linear_end_exclusive": hex(native.end_ea), "original_flags": native.flags,
                           "decoded_flags": [name for name, value in constants.items() if native.flags & value],
                           "last_instruction": last, "fresh_last_xrefs": references(ea),
                           "physical_next": inventory["instructions"].get(hex(next_ea)),
                           "physical_next_original_owner": ida_funcs.get_func_name(next_function.start_ea) if next_function else None})
    legacy_entries = inventory["legacy_comparison"]["legacy_only_entries"]
    legacy = []
    for address in legacy_entries:
        ea = int(address, 16)
        head = ida_bytes.get_item_head(ea)
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, head):
            raise ValueError("Legacy containing instruction failed to decode")
        original = inventory["instructions"].get(hex(head))
        if original is None:
            raise ValueError("Legacy entry is outside the verified code inventory")
        legacy.append({"legacy_ida_linear_entry": address, "is_code_head": head == ea,
                       "containing_instruction": original, "offset_inside_instruction": ea - head,
                       "fresh_entry_xrefs": references(ea)})
    for target in (0x15d49, 0x193e3, 0x1e713, 0x1e7f3, 0x236f5, 0x24a8d):
        native = ida_funcs.get_func(target)
        if native is None or native.start_ea != target:
            raise ValueError("Required function target lacks a boundary")
        record = by_start[hex(target)]
        rows = [inventory["instructions"][ea] for ea in record["instructions"]]
        for row in rows:
            current = mapped_bytes(int(row["ida_linear"], 16), len(bytes.fromhex(row["file_bytes"])))
            if current["file_bytes"] != row["file_bytes"] or current["loaded_bytes"] != row["loaded_bytes"]:
                raise ValueError("Function input/loaded bytes differ")
        functions.append({"original_name": record["original_name"], "ida_linear_start": hex(target),
                          "ida_linear_end_exclusive": hex(native.end_ea), "original_flags": native.flags,
                          "decoded_flags": [name for name, value in constants.items() if native.flags & value],
                          "chunks": record["chunks"], "instructions": rows, "entry_xrefs": references(target)})
    table_entry = 0x289e2
    item_head = ida_bytes.get_item_head(table_entry)
    item_end = ida_bytes.get_item_end(table_entry)
    table = {**mapped_bytes(table_entry, 2), "original_name": idc.get_name(table_entry),
             "original_item_head": hex(item_head), "original_item_end_exclusive": hex(item_end),
             "item_head_xrefs": references(item_head), "entry_xrefs": references(table_entry)}
    carrier = inventory["instructions"]["0x192ab"]
    set_ds = inventory["instructions"]["0x192ae"]
    table_lea = inventory["instructions"]["0x14ff2"]
    if carrier["file_bytes"] != "b8dd14" or set_ds["file_bytes"] != "8ed8" or table_lea["file_bytes"] != "8d36b43b":
        raise ValueError("Reviewed DS carrier/table address-taken bytes differ")
    dgroup = int.from_bytes(bytes.fromhex(carrier["loaded_bytes"])[1:3], "little") * 16
    table_offset = int.from_bytes(bytes.fromhex(table_lea["file_bytes"])[2:4], "little")
    table_base = dgroup + table_offset
    if table_entry < table_base or (table_entry - table_base) % 2:
        raise ValueError("Table entry is not an aligned word in the reviewed table")
    call_window = [inventory["instructions"][address] for address in
                   ("0x14feb", "0x14fee", "0x14ff0", "0x14ff2", "0x14ff6", "0x14ff8", "0x14ffa", "0x14ffd", "0x14fff", "0x15001")]
    table_context = {"startup_DS_carrier": carrier, "startup_set_DS": set_ds,
                     "IDA_DGROUP_linear": hex(dgroup), "table_DS_offset": hex(table_offset),
                     "table_base": {**mapped_bytes(table_base, 2), "entry_xrefs": references(table_base)},
                     "selected_word_index": (table_entry - table_base) // 2,
                     "address_taken_and_consumer": call_window,
                     "scope": "Conditional static mapping with DS=the startup DGROUP; caller DI origin and field +4 meaning remain unknown",
                     "inference_level": "strong static table/caller mapping; no normal-player-path claim"}
    indirect_calls = []
    for address, row in inventory["instructions"].items():
        if row["mnemonic"] != "call":
            continue
        ea = int(address, 16)
        insn = ida_ua.insn_t()
        if not ida_ua.decode_insn(insn, ea):
            raise ValueError("Call decode failed")
        operand = insn.ops[0]
        if operand.type not in (ida_ua.o_mem, ida_ua.o_phrase, ida_ua.o_displ, ida_ua.o_reg):
            continue
        indirect_calls.append({"instruction": row, "operand_type": operand.type,
                               "operand_addr": hex(operand.addr), "operand_reg_or_phrase": operand.reg,
                               "original_operand": idc.print_operand(ea, 0),
                               "inference_level": "unknown", "warning": "Indirect call target/base context needs review"})
    return {"schema_version": 1, "input": {"path": str(source), "size": len(raw), "sha256": HASH},
            "tool": {"name": "IDA Pro", "version": ida_kernwin.get_kernel_version(),
                     "script_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest()},
            "inventory_source": {"path": str(inventory_path), "sha256": hashlib.sha256(inventory_path.read_bytes()).hexdigest()},
            "address_space": inventory["address_space"], "IDA_function_flag_constants": constants,
            "nonterminal_boundaries": boundaries, "legacy_interior_entries": legacy,
            "functions": functions, "first_C_table_entry": table, "first_C_table_context": table_context,
            "indirect_calls": indirect_calls,
            "scope": "Fresh IDA evidence only; no boundary repair, semantic renaming or automatic source-unit approval"}


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
