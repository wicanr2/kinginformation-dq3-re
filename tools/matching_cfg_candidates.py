"""Docker-only multi-entry CFG candidates from original IDA evidence, docs/25.

Follow actual fl_F/fl_J edges, not automatic function ownership. Calls are
dependencies, not local CFG edges. Never invent suppressed return edges or
approve a source unit merely because its reachable set has no missing nodes.
"""

import argparse
from collections import Counter
import hashlib
import json
import os
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
EXPECTED_HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def byte_ranges(addresses):
    result = []
    for address in sorted(addresses):
        if result and result[-1][1] == address:
            result[-1][1] += 1
        else:
            result.append([address, address + 1])
    return [[hex(lo), hex(hi)] for lo, hi in result]


def build_candidates(inventory, review):
    rows = inventory["instructions"]
    functions = {f["ida_linear_start"]: f for f in inventory["functions"]}
    owners = {}
    for function in inventory["functions"]:
        for address in function["instructions"]:
            owners.setdefault(address, []).append(function["ida_linear_start"])
    ordered = sorted(int(address, 16) for address in rows)
    previous = {hex(address): rows[hex(ordered[index-1])] if index else None for index, address in enumerate(ordered)}
    candidates = []
    for boundary in review["reviewed_boundaries"]:
        if boundary["kind"] != "cross_boundary_flow":
            continue
        start = boundary["ida_linear_start"]
        queue = [start]
        visited = set()
        local_edges = []
        calls = []
        exits = []
        unresolved = []
        while queue:
            address = queue.pop()
            if address in visited:
                continue
            if len(visited) >= 4096:
                raise ValueError("CFG candidate exceeds explicit 4096-head research bound")
            if address not in rows:
                raise ValueError("CFG successor has no verified IDA instruction: " + address)
            visited.add(address)
            row = rows[address]
            mnemonic = row["mnemonic"]
            successors = [ref for ref in row["refs_from"] if ref["iscode"] and ref["type"] in (18, 19, 21)]
            if mnemonic in ("ret", "retn", "retf", "iret", "iretd"):
                if successors:
                    raise ValueError("Return instruction has unexpected local successors")
                exits.append({"ida_linear": address, "kind": "original_return", "mnemonic": mnemonic})
                continue
            if row["file_bytes"] == "cd21" and previous[address] and previous[address]["file_bytes"] == "b44c":
                if successors:
                    raise ValueError("DOS terminate has an unexpected local successor")
                exits.append({"ida_linear": address, "kind": "DOS_AH4C", "AH_writer": previous[address]["ida_linear"]})
                continue
            if mnemonic == "call":
                targets = [ref for ref in row["refs_from"] if ref["iscode"] and ref["type"] in (16, 17)]
                code = bytes.fromhex(row["file_bytes"])
                # A typed xref on FF /2 may be an IDA-resolved table target.
                # It is not promoted to an original direct E8/9A call.
                direct = code[0] in (0xe8, 0x9a)
                calls.append({"ida_linear": address, "file_bytes": row["file_bytes"],
                              "kind": "encoded_direct_call" if direct else "indirect_call",
                              "IDA_target_candidates": targets,
                              "post_call_flow_recorded": any(ref["type"] == 21 for ref in successors),
                              "inference_level": "confirmed call encoding; effects and return behavior not inferred"})
                if not successors:
                    unresolved.append({"ida_linear": address, "kind": "suppressed_or_unrecovered_call_return",
                                       "warning": "No synthetic continuation was added"})
            elif not successors:
                unresolved.append({"ida_linear": address, "kind": "unrecovered_local_successor", "mnemonic": mnemonic})
            for reference in successors:
                destination = reference["to"]
                if destination not in rows:
                    raise ValueError("Typed edge points into an unverified/mid-instruction address")
                local_edges.append({"from": address, "to": destination, "xref_type": reference["type"]})
                if destination not in visited:
                    queue.append(destination)
        inbound = []
        for address in sorted(visited, key=lambda value: int(value, 16)):
            refs = [ref for ref in rows[address]["refs_to"] if ref["from"] not in visited]
            if refs:
                inbound.append({"ida_linear_entry": address, "original_name": rows[address]["original_name"],
                                "original_IDA_owners": owners.get(address, []), "incoming_xrefs": refs,
                                "scope": "External code/data references retained; data references need consumer review"})
        file_addresses = set()
        for address in visited:
            row = rows[address]
            offset = int(row["file_offset"], 16)
            file_addresses.update(range(offset, offset + len(bytes.fromhex(row["file_bytes"]))))
        candidates.append({"original_entry_name": functions[start]["original_name"], "ida_linear_entry": start,
                           "original_IDA_end_exclusive": functions[start]["ida_linear_end_exclusive"],
                           "reachable_instruction_heads": sorted(visited, key=lambda value: int(value, 16)),
                           "reachable_file_byte_ranges": byte_ranges(file_addresses), "reachable_file_bytes": len(file_addresses),
                           "local_edges": local_edges, "calls": calls, "exits": exits, "unresolved": unresolved,
                           "external_entries": inbound, "original_IDA_owners_crossed": sorted({owner for address in visited for owner in owners.get(address, [])}),
                           "source_unit_extent_approved": False,
                           "inference_level": "confirmed scoped graph of verified typed edges; source-unit grouping unapproved"})
    memberships = {}
    for candidate in candidates:
        for address in candidate["reachable_instruction_heads"]:
            memberships.setdefault(address, []).append(candidate["ida_linear_entry"])
    shared = {address: entries for address, entries in memberships.items() if len(entries) > 1}
    return candidates, shared


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--inventory", type=Path, required=True)
    parser.add_argument("--boundary-review", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    original = ROOT / "assets_raw/DQ3.EXE"
    raw = original.read_bytes()
    inventory = json.loads(args.inventory.read_text(encoding="utf-8"))
    review = json.loads(args.boundary_review.read_text(encoding="utf-8"))
    if sha(original) != EXPECTED_HASH or len(raw) != 115282:
        raise ValueError("Original input differs")
    if review["inventory_sha256"] != sha(args.inventory) or inventory["input"]["sha256"] != EXPECTED_HASH:
        raise ValueError("Reviewed inventory differs")
    for row in inventory["instructions"].values():
        offset = int(row["file_offset"], 16)
        code = bytes.fromhex(row["file_bytes"])
        if raw[offset:offset + len(code)] != code:
            raise ValueError("CFG original bytes differ")
    candidates, shared = build_candidates(inventory, review)
    if len(candidates) != 17:
        raise ValueError("Cross-boundary candidate count differs")
    union_heads = {address for candidate in candidates for address in candidate["reachable_instruction_heads"]}
    exits = Counter(exit["kind"] for candidate in candidates for exit in candidate["exits"])
    result = {"schema_version": 1, "input": inventory["input"], "address_space": inventory["address_space"],
              "IDA_tool": inventory["tool"], "inventory_sha256": sha(args.inventory),
              "boundary_review_sha256": sha(args.boundary_review), "producer_sha256": sha(Path(__file__)),
              "candidates": candidates, "shared_instruction_entries": shared,
              "union_reachable_heads": len(union_heads), "exit_kind_counts": dict(exits),
              "scope": "Candidates from 17 reviewed cross-boundary edges only; calls/indirect effects and all external entries retained",
              "approved_source_units": 0, "whole_goal_complete": False}
    with args.output.open("x", encoding="utf-8") as stream:
        stream.write(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"candidates": len(candidates), "union_heads": len(union_heads), "shared_heads": len(shared),
                      "with_unresolved": sum(bool(candidate["unresolved"]) for candidate in candidates),
                      "approved_source_units": 0, "whole_goal_complete": False}))


if __name__ == "__main__":
    main()
