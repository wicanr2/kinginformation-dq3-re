"""Docker-only scoped review of fresh IDA boundary evidence; docs/25.

Export factual boundary resolutions without rewriting the original database or
guessing C source-unit extents. Keep callee-based conclusions at strong level.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
DOS_EXIT_REFERENCE = "https://fd.lod.bz/rbil/interrup/dos_kernel/214c.html"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--inventory", type=Path, required=True)
    parser.add_argument("--boundary-evidence", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    inventory = json.loads(args.inventory.read_text(encoding="utf-8"))
    evidence = json.loads(args.boundary_evidence.read_text(encoding="utf-8"))
    original = ROOT / "assets_raw/DQ3.EXE"
    raw = original.read_bytes()
    expected = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
    if sha(original) != expected or len(raw) != 115282 or evidence["input"]["sha256"] != expected:
        raise ValueError("Original input differs")
    if evidence["tool"]["version"] != "9.4" or evidence["tool"]["script_sha256"] != sha(ROOT / "tools/ida_matching_boundaries.py"):
        raise ValueError("Boundary exporter freshness differs")
    if evidence["inventory_source"]["sha256"] != sha(args.inventory):
        raise ValueError("Boundary inventory source differs")
    rows = inventory["instructions"]
    ordered = sorted(int(address, 16) for address in rows)
    before = {hex(address): rows[hex(ordered[index-1])] if index else None for index, address in enumerate(ordered)}
    functions = {f["ida_linear_start"]: f for f in inventory["functions"]}
    reviewed = []
    counts = {"cross_boundary_flow": 0, "DOS_terminal_end": 0, "suppressed_call_return": 0}
    for boundary in evidence["nonterminal_boundaries"]:
        row = boundary["last_instruction"]
        offset = int(row["file_offset"], 16)
        code = bytes.fromhex(row["file_bytes"])
        if raw[offset:offset + len(code)] != code:
            raise ValueError("Boundary original bytes differ")
        next_address = int(row["ida_linear"], 16) + len(code)
        flows = [x for x in boundary["fresh_last_xrefs"]["refs_from"] if x["type"] == 21]
        record = {"original_name": boundary["original_name"], "ida_linear_start": boundary["ida_linear_start"],
                  "original_IDA_end_exclusive": boundary["ida_linear_end_exclusive"],
                  "original_flags": boundary["original_flags"], "decoded_flags": boundary["decoded_flags"],
                  "last_instruction": row, "physical_next": boundary["physical_next"],
                  "source_unit_extent_approved": False}
        if flows:
            if len(flows) != 1 or int(flows[0]["to"], 16) != next_address or boundary["physical_next"] is None:
                raise ValueError("Unexpected nonterminal boundary flow")
            record.update(kind="cross_boundary_flow", inference_level="confirmed scoped static flow",
                          resolution="The instruction has an explicit IDA fl_F edge into the physical next instruction beyond the automatic function boundary",
                          continuation_original_owner=boundary["physical_next_original_owner"], flow_xref=flows[0])
        elif row["file_bytes"] == "cd21":
            previous = before[row["ida_linear"]]
            if previous is None or previous["file_bytes"] != "b44c":
                raise ValueError("INT21 terminal requires the actual AH=4C writer")
            record.update(kind="DOS_terminal_end", inference_level="confirmed local service selection",
                          resolution="Immediate MOV AH,4C followed by INT21 selects the standard DOS termination service; local end is valid, whole-function source span still needs review",
                          AH_writer=previous, platform_reference=DOS_EXIT_REFERENCE)
        elif row["mnemonic"] == "call":
            callees = [x for x in row["refs_from"] if x["type"] in (16, 17)]
            if len(callees) != 1 or callees[0]["to"] not in functions:
                raise ValueError("Suppressed call needs one original direct callee")
            callee = functions[callees[0]["to"]]
            if not callee["original_flags"] & evidence["IDA_function_flag_constants"]["FUNC_NORET"]:
                raise ValueError("IDA suppressed return without a NORET callee")
            record.update(kind="suppressed_call_return", inference_level="strong scoped callee-based interpretation",
                          resolution="IDA suppresses the post-call flow because its direct callee is marked NORET; inspect the callee loop/exit chain rather than extending blindly",
                          callee_original_name=callee["original_name"], callee_IDA_linear_start=callee["ida_linear_start"],
                          callee_original_flags=callee["original_flags"],
                          warning="Callee effects and indirect/nonlocal transfers are not inferred from the flag alone")
        else:
            raise ValueError("Unreviewed boundary class")
        counts[record["kind"]] += 1
        reviewed.append(record)
    if counts != {"cross_boundary_flow": 17, "DOS_terminal_end": 2, "suppressed_call_return": 3}:
        raise ValueError("Reviewed boundary set differs")
    legacy = []
    for entry in evidence["legacy_interior_entries"]:
        if entry["is_code_head"] or not entry["offset_inside_instruction"] or entry["fresh_entry_xrefs"]["refs_to"]:
            raise ValueError("Legacy entry needs separate incoming-edge review")
        row = entry["containing_instruction"]
        offset = int(row["file_offset"], 16)
        code = bytes.fromhex(row["file_bytes"])
        if raw[offset:offset + len(code)] != code:
            raise ValueError("Legacy containing bytes differ")
        legacy.append({**entry, "inference_level": "confirmed current input/IDA boundary",
                       "resolution": "Inside an already verified instruction, with no direct incoming xref; excluded from source-function entries until independent entry evidence exists"})
    context = evidence["first_C_table_context"]
    entry = evidence["first_C_table_entry"]
    dgroup = int(context["IDA_DGROUP_linear"], 16)
    table_base = int(context["table_base"]["ida_linear"], 16)
    word_index = context["selected_word_index"]
    if word_index != 47 or table_base != dgroup + 0x3bb4 or int(entry["ida_linear"], 16) != table_base + word_index * 2:
        raise ValueError("First C static table mapping differs")
    if int.from_bytes(bytes.fromhex(entry["file_bytes"]), "little") != 0x5d49:
        raise ValueError("First C original table target differs")
    window = context["address_taken_and_consumer"]
    expected_window = {"0x14feb": "8a5d04", "0x14fee": "32ff", "0x14ff0": "d1e3",
                       "0x14ff2": "8d36b43b", "0x14ff6": "03f3", "0x14ff8": "8b04",
                       "0x14ffa": "3d0000", "0x14ffd": "7402", "0x14fff": "ff14", "0x15001": "c3"}
    if {row["ida_linear"]: row["file_bytes"] for row in window} != expected_window:
        raise ValueError("First C address-taken/consumer window differs")
    result = {"schema_version": 1, "input": evidence["input"], "address_space": evidence["address_space"],
              "IDA_evidence_sha256": sha(args.boundary_evidence), "inventory_sha256": sha(args.inventory),
              "producer_sha256": sha(Path(__file__)), "reviewed_boundaries": reviewed, "counts": counts,
              "legacy_entry_resolutions": legacy, "first_C_indirect_entry": {**context, "table_entry": entry,
                  "selected_record_field_value": 47, "original_field_operand": "[di+4]",
                  "target_logical": "0x5d49", "target_IDA_linear": "0x15d49",
                  "zero_word_gate": "AX loaded from selected word; CMP AX,0 / JZ skips CALL [SI]",
                  "runtime_DS_and_DI_origin_verified": False, "field_product_meaning": "unknown"},
              "scope": "Boundary facts and conditional static indirect-entry closure only; no C unit merge or runtime/player-path claim",
              "whole_goal_complete": False}
    with args.output.open("x", encoding="utf-8") as stream:
        stream.write(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"counts": counts, "legacy_interior_entries": len(legacy), "first_C_conditional_table_index": word_index,
                      "source_unit_extents_approved": 0, "whole_goal_complete": False}))


if __name__ == "__main__":
    main()
