"""Reviewed WCC F5/T2 calls; entry: docs/25-match-progress.md.

TIS OMF 1.1 FIXUPP, location 1 self-relative and location 3 far pointer.
This is a fixed-placement comparator, not a replacement for the vendor linker.
Caller and target addresses are MZ-relative segment:offset, never IDA linear.
Far-call relaxation must be disabled; original source/layout remains separate.
"""

import copy
import struct

from omf_matching_probe import UnsupportedOMF, resolve_ds_offsets, select_function


def address(value):
    if not isinstance(value, dict) or set(value) != {"segment", "offset"}:
        raise ValueError("Expected explicit MZ-relative segment and offset")
    if any(type(value[k]) is not int or not 0 <= value[k] <= 65535 for k in value):
        raise ValueError("Address component is not an unsigned 16-bit integer")
    return value["segment"], value["offset"]


def resolve_candidate_fixups(obj, case):
    """Consume the manifest's explicitly typed code placement contract."""
    mode = case.get("encoded_addend_mode", "unsigned16")
    if mode not in ("unsigned16", "signed16"):
        raise ValueError("Unknown encoded addend contract")
    ds_offsets = {symbol: int(value, 16) for symbol, value in case["external_DS_offsets"].items()}
    placements = case.get("call_placements")
    if placements is None:
        code, fixes = resolve_ds_offsets(obj, case["public_symbol"], ds_offsets, signed_addends=mode == "signed16")
        return code, fixes, []
    if not isinstance(placements, dict) or set(placements) != {"caller", "near", "far"}:
        raise ValueError("Incomplete typed call placement contract")
    return resolve_function_fixups(obj, case["public_symbol"], ds_offsets, placements["caller"],
                                   near_symbols=placements["near"], far_symbols=placements["far"],
                                   signed_addends=mode == "signed16")


def validate_original_caller(case, inventory):
    placements = case.get("call_placements")
    if placements is None:
        return
    segment, offset = address(placements["caller"])
    load_base = int(inventory["address_space"]["load_base"], 16)
    start = int(case["ida_linear_start"], 16)
    original_segment = next(s for s in inventory["segments"]
                            if int(s["ida_linear_start"], 16) <= start < int(s["ida_linear_end_exclusive"], 16))
    # Inventory segment extents alone do not recover arbitrary original CS
    # frames. This integration currently admits the reviewed initial CODE
    # segment; other callers need an independently exported frame ledger.
    if int(original_segment["ida_linear_start"], 16) != load_base or original_segment["type"] != 2:
        raise UnsupportedOMF("Original caller is outside the reviewed initial CODE frame")
    original_frame = 0
    if segment != original_frame or load_base + segment * 16 + offset != start:
        raise ValueError("Caller address differs from original IDA code frame")
    if start + case["size"] > int(original_segment["ida_linear_end_exclusive"], 16):
        raise ValueError("Source unit crosses the original IDA segment")


def validate_indirect_dispatch(case, inventory, code, fixes):
    """Reviewed zero-gated near DS table call; no inferred callback targets.

    This narrow contract proves the original raw read/branch/call pattern and
    actual data fixups. Table extent, valid indices and runtime DS stay unknown.
    """
    start = int(case["ida_linear_start"], 16)
    function = next(f for f in inventory["functions"] if int(f["ida_linear_start"], 16) == start)
    rows = [inventory["instructions"][ea] for ea in function["instructions"]]
    indirect = [row for row in rows if row["mnemonic"] == "call" and not any(
        ref["iscode"] and ref["type"] in (16, 17) for ref in row["refs_from"])]
    contract = case.get("indirect_dispatch")
    if contract is None:
        if indirect:
            raise UnsupportedOMF("Original indirect call needs a reviewed dispatch contract")
        return None
    if (not isinstance(contract, dict) or set(contract) != {"kind", "index_symbol", "table_symbol"}
            or contract["kind"] != "zero-gated-near-ds-table-bx-v1"):
        raise UnsupportedOMF("Unknown indirect dispatch contract")
    index_symbol, table_symbol = contract["index_symbol"], contract["table_symbol"]
    offsets = case["external_DS_offsets"]
    if (not isinstance(index_symbol, str) or not isinstance(table_symbol, str)
            or index_symbol == table_symbol or set(offsets) != {index_symbol, table_symbol}):
        raise ValueError("Indirect dispatch requires distinct reviewed DS symbols")
    index, table = int(offsets[index_symbol], 16), int(offsets[table_symbol], 16)
    if not 0 <= index <= 65534 or not 0 <= table <= 65534:
        raise ValueError("Indirect dispatch DS word crosses its 16-bit frame")
    expected = (b"\x8b\x1e" + struct.pack("<H", index) + b"\xd1\xe3\x83\xbf"
                + struct.pack("<H", table) + b"\x00\x74\x04\xff\x97"
                + struct.pack("<H", table) + b"\xc3")
    if (case["size"] != 18 or function["chunks"] != [[hex(start), hex(start + 18)]]
            or [int(row["ida_linear"], 16) - start for row in rows] != [0, 4, 6, 11, 13, 17]
            or b"".join(bytes.fromhex(row["file_bytes"]) for row in rows) != expected):
        raise ValueError("Original indirect dispatch differs from reviewed pattern")
    if (len(indirect) != 1 or int(indirect[0]["ida_linear"], 16) != start + 13
            or not any(ref["iscode"] and ref["type"] == 19 and int(ref["to"], 16) == start + 17
                       for ref in rows[3]["refs_from"])):
        raise ValueError("Original zero-gate or indirect call xref differs")
    if code != expected:
        raise ValueError("Compiled indirect dispatch differs from reviewed pattern")
    expected_fixes = [(2, index_symbol, index), (8, table_symbol, table), (15, table_symbol, table)]
    actual = sorted((fix["offset"], fix["symbol"], fix["resolved_operand"]) for fix in fixes)
    if actual != expected_fixes or any(fix["original_addend"] != 0 for fix in fixes):
        raise ValueError("Indirect dispatch actual DS fixups differ")
    return {"kind": contract["kind"], "gate_ida_linear": hex(start + 6),
            "call_ida_linear": hex(start + 13), "zero_branch_target": hex(start + 17),
            "pointer_bytes": 2, "index_loads": 1, "table_reads": 2,
            "inference_level": "confirmed static raw dispatch pattern",
            "callback_targets": "unknown", "table_extent": "unknown", "runtime_DS": "unknown"}


def resolve_function_fixups(obj, symbol, ds_offsets, caller_address, *,
                            near_symbols=None, far_symbols=None, signed_addends=False):
    near_symbols = {} if near_symbols is None else near_symbols
    far_symbols = {} if far_symbols is None else far_symbols
    namespaces = [set(ds_offsets), set(near_symbols), set(far_symbols)]
    if any(namespaces[i] & namespaces[j] for i in range(3) for j in range(i)):
        raise ValueError("A symbol has conflicting data/near/far placement kinds")
    caller_segment, caller_offset = address(caller_address)
    code, selected = select_function(obj, symbol)
    if caller_offset + len(code) > 65536:
        raise UnsupportedOMF("Caller module crosses its 16-bit code frame")
    for target in list(near_symbols.values()) + list(far_symbols.values()):
        address(target)
    filtered = copy.deepcopy(obj)
    filtered["fixups"] = []
    calls = []
    occupied = set()
    for fixup in obj["fixups"]:
        if fixup["segment_index"] != selected["segment_index"]:
            # No unseen data/object fixups may silently enter this source unit.
            raise UnsupportedOMF("Relocation outside the selected code module")
        if fixup["frame"]["method"] != 5 or fixup["target"]["method"] != 2:
            raise UnsupportedOMF("Call comparator requires explicit F5/T2 semantics")
        index = fixup["target"]["datum"]
        if type(index) is not int or not 0 < index < len(obj["externals"]):
            raise ValueError("Invalid external target index")
        name = obj["externals"][index]["name"]
        if name in ds_offsets:
            width = 2
            filtered["fixups"].append(fixup)
        elif name in near_symbols:
            if fixup["location_type"] != 1 or fixup["segment_relative"]:
                raise UnsupportedOMF("Near code target requires self-relative location 1")
            width = 2
            calls.append((fixup, name, "near"))
        elif name in far_symbols:
            if fixup["location_type"] != 3 or not fixup["segment_relative"]:
                raise UnsupportedOMF("Far code target requires segment-relative location 3")
            width = 4
            calls.append((fixup, name, "far"))
        else:
            raise UnsupportedOMF("Missing typed symbol placement: " + name)
        offset = fixup["offset"]
        region = set(range(offset, offset + width))
        if offset < 0 or offset + width > len(code) or occupied & region:
            raise ValueError("Relocations overlap or cross the code module")
        occupied.update(region)
    linked, resolved = resolve_ds_offsets(filtered, symbol, ds_offsets, signed_addends=signed_addends)
    linked = bytearray(linked)
    mz_segment_offsets = []
    for fixup, name, kind in calls:
        offset = fixup["offset"]
        width = 2 if kind == "near" else 4
        if offset < 1 or offset - 1 in occupied or any(code[offset:offset + width]) or fixup["displacement"]:
            raise UnsupportedOMF("Only zero-addend WCC direct calls are reviewed")
        if code[offset - 1] != (0xE8 if kind == "near" else 0x9A):
            raise UnsupportedOMF("Relocation is not the reviewed direct CALL instruction")
        segment, target = address((near_symbols if kind == "near" else far_symbols)[name])
        if kind == "near":
            if segment != caller_segment:
                raise ValueError("Near target belongs to a different code frame")
            displacement = target - (caller_offset + offset + 2)
            # Explicit 8086 rel16 arithmetic, independently checked with WLINK.
            encoded = displacement & 0xFFFF
            struct.pack_into("<H", linked, offset, encoded)
            detail = {"displacement_before_rel16_encoding": displacement,
                      "resolved_operand": encoded}
        else:
            struct.pack_into("<HH", linked, offset, target, segment)
            mz_segment_offsets.append(offset + 2)
            detail = {"resolved_offset": target, "resolved_segment": segment,
                      "mz_segment_word_offset": offset + 2}
        resolved.append({**fixup, "symbol": name, "placement_kind": kind,
                         "caller_address": dict(caller_address),
                         "target_address": {"segment": segment, "offset": target},
                         "original_addend": 0, **detail})
    return bytes(linked), sorted(resolved, key=lambda r: r["offset"]), sorted(mz_segment_offsets)
