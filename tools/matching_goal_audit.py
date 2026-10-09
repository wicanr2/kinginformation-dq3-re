"""Docker-only evidence audit for the full DQ3 matching Goal; docs/25.

An evidence audit can pass while all whole-Goal gates remain unproven.
Automatic IDA functions, original-byte scaffolds and compiler module padding
are never silently promoted to complete source recovery.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import tempfile

from omf_matching_probe import UnsupportedOMF, read_object, resolve_ds_offsets
from run_matching_c_batch import msc_listing_function
from run_watcom16_abi import compiler_flags_for
from omf_call_fixups import resolve_candidate_fixups, validate_original_caller


ROOT = Path(__file__).resolve().parents[1]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--inventory", type=Path, required=True)
    parser.add_argument("--c-receipt", type=Path, required=True)
    parser.add_argument("--repeat-c-receipt", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--watcom-receipt", type=Path)
    parser.add_argument("--repeat-watcom-receipt", type=Path)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Run as host UID/GID")
    contract_path = ROOT / "tools/matching_goal_contract.json"
    contract = json.loads(contract_path.read_text())
    original_path = ROOT / contract["input"]["path"]
    raw = original_path.read_bytes()
    if sha(original_path) != contract["input"]["sha256"] or len(raw) != contract["input"]["size"]:
        raise ValueError("Original input differs")
    inventory = json.loads(args.inventory.read_text())
    if (inventory["input"]["sha256"] != sha(original_path) or inventory["tool"]["version"] != "9.4"
            or inventory["tool"]["script_sha256"] != sha(ROOT / "tools/ida_matching_inventory.py")):
        raise ValueError("Full IDA inventory identity/freshness differs")
    for ledger in inventory["semantic_ledgers"]:
        if sha(Path(ledger["path"])) != ledger["sha256"]:
            raise ValueError("Consumed semantic ledger differs")
    unique_bytes = set()
    for address, row in inventory["instructions"].items():
        offset = int(row["file_offset"], 16)
        data = bytes.fromhex(row["file_bytes"])
        if raw[offset:offset + len(data)] != data:
            raise ValueError("Inventory original instruction differs: " + address)
        if int(row["ida_linear"], 16) != int(row["logical"], 16) + int(inventory["address_space"]["load_base"], 16):
            raise ValueError("Inventory address bases differ")
        unique_bytes.update(range(offset, offset + len(data)))
    if len(unique_bytes) != inventory["verified_code_file_bytes"]:
        raise ValueError("Inventory unique code-byte count differs")
    if inventory["unbacked_code_heads"] or inventory["overlapping_file_instruction_bytes"]:
        raise ValueError("Unbacked or overlapping IDA code needs review")
    legacy = json.loads((ROOT / "docs/data/exe_funcs.json").read_text())
    if sha(ROOT / "docs/data/exe_funcs.json") != inventory["legacy_comparison"]["source_sha256"]:
        raise ValueError("Legacy inventory differs")
    if len(legacy["funcs"]) != inventory["legacy_comparison"]["legacy_entries"]:
        raise ValueError("Legacy entry count differs")
    # Missing function ownership is explicitly retained. Identify automatic
    # functions which flow past their declared end, rather than accepting them
    # as complete source functions just because IDA assigned a name.
    nonterminal = []
    function_ranges = {}
    for function in inventory["functions"]:
        function_ranges[function["ida_linear_start"]] = function
        if function["instructions"]:
            last = inventory["instructions"][function["instructions"][-1]]
            if last["mnemonic"] not in ("ret", "retn", "retf", "iret", "iretd", "jmp"):
                nonterminal.append({"original_name": function["original_name"],
                                    "ida_linear_start": function["ida_linear_start"],
                                    "last_instruction": last["ida_linear"], "mnemonic": last["mnemonic"],
                                    "inference_level": "unknown", "warning": "Automatic boundary may split live control flow"})
    receipt = json.loads(args.c_receipt.read_text())
    repeat = json.loads(args.repeat_c_receipt.read_text())
    manifest_path = ROOT / "tools/matching_c_manifest.json"
    manifest = json.loads(manifest_path.read_text())
    if receipt["manifest_sha256"] != sha(manifest_path):
        raise ValueError("C manifest freshness differs")
    for result in (receipt, repeat):
        if result["producer_sha256"] != sha(ROOT / "tools/run_matching_c_batch.py"):
            raise ValueError("C producer freshness differs")
        if result["omf_parser_sha256"] != sha(ROOT / "tools/omf_matching_probe.py"):
            raise ValueError("OMF parser freshness differs")
    exact_bytes = 0
    exact_units = []
    parser_negatives = []
    for case in receipt["cases"]:
        source = ROOT / case["source"]
        if sha(source) != case["source_sha256"]:
            raise ValueError("C source freshness differs")
        original = raw[int(case["file_start"], 16):int(case["file_start"], 16) + case["size"]]
        if original.hex() != case["original_hex"]:
            raise ValueError("C source-unit original range differs")
        function = function_ranges[case["ida_linear_start"]]
        if b"".join(bytes.fromhex(inventory["instructions"][ea]["file_bytes"]) for ea in function["instructions"]) != original:
            raise ValueError("C source-unit range differs from full inventory")
        prior = next(item for item in repeat["cases"] if item["id"] == case["id"])
        for key in ("object_sha256", "candidate_sha256", "comparison", "module_compare", "fixups"):
            if case[key] != prior[key]:
                raise ValueError("Independent C rebuild differs: " + key)
        folder = args.c_receipt.parent
        obj_path = folder / (case["stem"].upper() + ".OBJ")
        if sha(obj_path) != case["object_sha256"]:
            raise ValueError("Compiled object differs")
        obj = read_object(obj_path)
        module, _ = resolve_ds_offsets(obj, case["public_symbol"], {name: int(value, 16) for name, value in case["external_DS_offsets"].items()})
        public = next(p for p in obj["publics"] if p["name"] == case["public_symbol"])
        raw_code = bytes(obj["segments"][public["segment_index"]]["bytes"])
        listing = folder / (case["stem"].upper() + ".COD")
        extent, provenance = msc_listing_function(raw_code, listing, case["public_symbol"])
        if sha(listing) != case["compiler_function_extent"]["sha256"] or extent != case["size"]:
            raise ValueError("Compiler PROC extent differs")
        code = (folder / (case["stem"] + "-linked.bin")).read_bytes()
        if module[:extent] != code or code != original or not case["exact"]:
            raise ValueError("C function is not exact")
        if module[extent:] and not provenance["post_ENDP_bytes"]:
            raise ValueError("Trailing module bytes are unaccounted")
        exact_bytes += len(code)
        exact_units.append({"original_name": case["original_name"], "ida_linear_start": case["ida_linear_start"],
                            "source": case["source"], "bytes": len(code), "inference_level": "confirmed compiler-PROC byte match",
                            "module_padding_placement": case["module_padding_placement"]})
        # Real listing is the positive control; mutations must fail independently
        # of comparison against original bytes.
        text = listing.read_text(encoding="ascii")
        mutations = {"missing-ENDP": text.replace(case["public_symbol"] + "\tENDP", "")}
        if provenance["post_ENDP_bytes"]:
            outside = provenance["post_ENDP_bytes"][0]
            needle = format(outside["offset"], "06x") + "\t" + outside["bytes"][:2]
            mutations["changed-post-ENDP-byte"] = text.replace(needle, format(outside["offset"], "06x") + "\t" + format(int(outside["bytes"][:2], 16) ^ 1, "02x"))
        else:
            import re
            mutations["changed-PROC-byte"] = re.sub(r"(\*\*\*\s+000000\s+)[0-9a-fA-F]{2}",
                                                    lambda match: match.group(1) + format(raw_code[0] ^ 1, "02x"), text, count=1)
        with tempfile.TemporaryDirectory(prefix="dq3-listing-negative-") as temporary:
            for name, content in mutations.items():
                if content == text:
                    raise ValueError("Negative mutation did not affect the known listing")
                path = Path(temporary) / (name + ".COD")
                path.write_text(content, encoding="ascii")
                try:
                    msc_listing_function(raw_code, path, case["public_symbol"])
                except (ValueError, UnsupportedOMF):
                    parser_negatives.append({"source_unit": case["id"], "case": name, "rejected": True})
                else:
                    raise ValueError("Malformed compiler listing accepted")
    if exact_bytes != receipt["C_exact_bytes"] or len(exact_units) != receipt["C_exact_functions"]:
        raise ValueError("C source-unit coverage totals differ")
    watcom_candidates = []
    if args.watcom_receipt is not None:
        if args.repeat_watcom_receipt is None:
            raise ValueError("Watcom source units require an independent rebuild")
        native = json.loads(args.watcom_receipt.read_text())
        repeated = json.loads(args.repeat_watcom_receipt.read_text())
        native_manifest = ROOT / "tools/watcom_matching_manifest.json"
        declared_units = json.loads(native_manifest.read_text())["cases"]
        for record in (native, repeated):
            if record["candidate_manifest_sha256"] != sha(native_manifest) or record["IDA_inventory_sha256"] != sha(args.inventory):
                raise ValueError("Watcom candidate manifest/inventory freshness differs")
            if record["producer_sha256"] != sha(ROOT / "tools/run_watcom16_abi.py") or record["OMF_parser_sha256"] != sha(ROOT / "tools/omf_matching_probe.py"):
                raise ValueError("Watcom source producer/parser freshness differs")
            if record["call_resolver_sha256"] != sha(ROOT / "tools/omf_call_fixups.py"):
                raise ValueError("Watcom call resolver freshness differs")
            if sorted(result["source_unit"]["id"] for result in record["results"]) != sorted(unit["id"] for unit in declared_units):
                raise ValueError("Watcom source-unit membership differs")
        for case in native["results"]:
            unit = case["source_unit"]
            if unit != next(declared for declared in declared_units if declared["id"] == unit["id"]):
                raise ValueError("Watcom receipt source unit differs from manifest")
            expected_command = ["wcc"] + compiler_flags_for(unit) + ["-fo=" + case["case"] + ".obj", case["case"] + ".c"]
            if case["compiler_command"] != expected_command:
                raise ValueError("Watcom receipt compiler profile differs")
            if sha(ROOT / unit["source"]) != case["source_sha256"]:
                raise ValueError("Watcom source freshness differs")
            prior = next(result for result in repeated["results"] if result["case"] == case["case"])
            for key in ("source_sha256", "object_sha256", "code_hex", "applied_fixups", "original_compare", "compiler_command", "MZ_segment_word_offsets"):
                if case.get(key) != prior.get(key):
                    raise ValueError("Watcom independent source rebuild differs: " + key)
            code = (args.watcom_receipt.parent / (case["case"] + "-code.bin")).read_bytes()
            start = int(unit["file_start"], 16)
            original = raw[start:start + unit["size"]]
            exact = code == original
            if code.hex() != case["code_hex"] or exact != case["original_compare"]["byte_exact"]:
                raise ValueError("Watcom original comparison differs")
            obj_path = args.watcom_receipt.parent / (case["case"] + ".obj")
            if sha(obj_path) != case["object_sha256"]:
                raise ValueError("Watcom source object differs")
            obj = read_object(obj_path)
            resolved, fixes, mz_offsets = resolve_candidate_fixups(obj, unit)
            if resolved != code or fixes != case["applied_fixups"]:
                raise ValueError("Watcom actual relocation differs")
            if mz_offsets != case["MZ_segment_word_offsets"]:
                raise ValueError("Watcom MZ segment relocation receipt differs")
            if unit.get("call_placements"):
                validate_original_caller(unit, inventory)
                header = int.from_bytes(raw[8:10], "little") * 16
                count = int.from_bytes(raw[6:8], "little")
                table = int.from_bytes(raw[24:26], "little")
                original_relocations = []
                for n in range(count):
                    off = int.from_bytes(raw[table + 4 * n:table + 4 * n + 2], "little")
                    seg = int.from_bytes(raw[table + 4 * n + 2:table + 4 * n + 4], "little")
                    file_offset = header + seg * 16 + off
                    if start <= file_offset < start + unit["size"]:
                        original_relocations.append(file_offset - start)
                if sorted(original_relocations) != mz_offsets:
                    raise ValueError("Original MZ segment relocations differ from C module")
                if exact:
                    for fix in fixes:
                        if fix.get("placement_kind") not in ("near", "far"):
                            continue
                        call = inventory["instructions"][hex(int(unit["ida_linear_start"], 16) + fix["offset"] - 1)]
                        target = fix["target_address"]
                        linear = int(inventory["address_space"]["load_base"], 16) + target["segment"] * 16 + target["offset"]
                        expected_type = 17 if fix["placement_kind"] == "near" else 16
                        if call["mnemonic"] != "call" or not any(r["iscode"] and r["type"] == expected_type and int(r["to"], 16) == linear for r in call["refs_from"]):
                            raise ValueError("C call binding differs from original typed IDA xref")
            function = function_ranges[unit["ida_linear_start"]]
            if b"".join(bytes.fromhex(inventory["instructions"][address]["file_bytes"]) for address in function["instructions"]) != original:
                raise ValueError("Watcom original source-unit boundary differs")
            watcom_candidates.append({"original_name": unit["id"], "source": unit["source"], "compiled_bytes": len(code), "original_bytes": len(original), "exact": exact})
            if exact:
                if any(existing["ida_linear_start"] == unit["ida_linear_start"] for existing in exact_units):
                    raise ValueError("C coverage double-counted across compilers")
                exact_bytes += len(code)
                exact_units.append({"original_name": unit["id"], "ida_linear_start": unit["ida_linear_start"], "source": unit["source"],
                                    "bytes": len(code), "inference_level": "confirmed complete Watcom C module byte match; product/type meaning scoped",
                                    "module_padding_placement": "not present"})
    # Manual gates stay unproven. A passing local evidence audit cannot close
    # the whole Goal or replace a missing source-unit/layout/build contract.
    gates = [{"id": gate["id"], "completion_proven": False, "verify_kind": gate["verify"]["kind"],
              "reason": gate["verify"]["note"]} for gate in contract["completion_gates"]]
    if any(gate["verify_kind"] != "manual" for gate in gates):
        raise ValueError("New completion verifier needs explicit implementation/review")
    index = (ROOT / "docs/25-match-progress.md").read_text()
    if "ida_matching_probe.py" not in index:
        raise ValueError("Known indexed positive control missing")
    source_entries = [Path(unit["source"]).name for unit in manifest["cases"]]
    if args.watcom_receipt:
        source_entries.extend(Path(unit["source"]).name for unit in declared_units)
    for name in ("ida_matching_inventory.py", "run_matching_c_batch.py", "matching_goal_contract.json", "matching_c_manifest.json", "matching_goal_audit.py", "watcom_matching_manifest.json", "ida_matching_c_ledger.json", "verify_watcom_signed_fixup.py", "omf_call_fixups.py", "verify_watcom_call_fixups.py", *source_entries):
        if name not in index:
            raise ValueError("New file lacks a documentation entry: " + name)
    result = {"schema_version": 1, "input": contract["input"], "producer_sha256": sha(Path(__file__)),
              "contract_sha256": sha(contract_path), "IDA_inventory_sha256": sha(args.inventory),
              "C_receipt_sha256": sha(args.c_receipt), "repeat_C_receipt_sha256": sha(args.repeat_c_receipt),
              "Watcom_receipt_sha256": sha(args.watcom_receipt) if args.watcom_receipt else None,
              "repeat_Watcom_receipt_sha256": sha(args.repeat_watcom_receipt) if args.repeat_watcom_receipt else None,
              "automatic_IDA_functions": len(inventory["functions"]), "verified_instruction_heads": len(inventory["instructions"]),
              "verified_code_bytes": len(unique_bytes), "code_outside_functions": len(inventory["code_outside_functions"]),
              "automatic_nonterminal_function_ends": nonterminal,
              "legacy_comparison": inventory["legacy_comparison"], "exact_C_source_units": exact_units,
              "exact_C_bytes": exact_bytes, "independent_C_rebuild_equal": True, "listing_parser_negative_cases": parser_negatives,
              "watcom_source_candidates": watcom_candidates,
              "completion_gates": gates, "whole_goal_complete": False,
              "scope": "Authoritative current evidence only; whole source recovery, classifications and final layout remain incomplete"}
    with args.output.open("x", encoding="utf-8") as stream:
        stream.write(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"C_exact_source_units": len(exact_units), "C_exact_bytes": exact_bytes,
                      "automatic_IDA_functions": len(inventory["functions"]), "nonterminal_IDA_ends": len(nonterminal),
                      "outside_function_heads": len(inventory["code_outside_functions"]), "whole_goal_complete": False}))


if __name__ == "__main__":
    main()
