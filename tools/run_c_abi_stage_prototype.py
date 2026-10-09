"""Build the opt-in C ABI stage prototype; docs/25-match-progress.md.

The compiler API receives only C and an ABI profile. Original EXE data is read
after assembly/object construction, solely for comparison. Coverage stays zero.
"""

import argparse
import builtins
import copy
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import struct
import subprocess

from omf_matching_probe import read_object, select_function
from omf_call_fixups import resolve_candidate_fixups, resolve_function_fixups
from prototype_c_abi_stage import compile_source, UnsupportedC
from verify_watcom_call_fixups import public_address


ROOT = Path(__file__).resolve().parents[1]
INPUT_SHA = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
PROFILES = {"restore_void": "scoped-word-restore-v1", "restore_ax_return": "scoped-word-restore-v1",
            "register_envelope": "exact-gpr-envelope-v1"}
ASSEMBLY_WORK_ROOT = Path("/tmp/dq3-c-abi-stage-v1")


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(command, work, log):
    result = subprocess.run(command, cwd=work, capture_output=True, timeout=30)
    log.write_bytes(result.stdout + result.stderr)
    if result.returncode:
        raise ValueError("Tool failed: " + log.read_text(errors="replace"))


def build(source, profile, case, work):
    work.mkdir(exist_ok=False)
    # Enforce the pure compiler boundary while source and profile are in memory.
    original_open = builtins.open
    def no_file_access(*args, **kwargs):
        raise AssertionError("Compiler attempted filesystem access")
    try:
        builtins.open = no_file_access
        compiled = compile_source(source, profile)
    finally:
        builtins.open = original_open
    assembly = compiled["assembly"]
    if re.search(r"\b(DB|DW|DD|INCBIN|ORG|INCLUDE|MACRO)\b", assembly, re.I):
        raise ValueError("Compiler output contains unreviewed byte/data/import directives")
    (work / "input.c").write_text(source, encoding="ascii")
    (work / "lowered.asm").write_text(assembly, encoding="ascii")
    (work / "IR.json").write_text(json.dumps(compiled["IR"], indent=2) + "\n")
    # Wasm writes its absolute input path into THEADR. Build at a stable path
    # inside each fresh container, preserving complete unmodified OBJ bytes.
    assembly_work = ASSEMBLY_WORK_ROOT / work.name
    assembly_work.mkdir(parents=True, exist_ok=False)
    shutil.copyfile(work / "lowered.asm", assembly_work / "lowered.asm")
    command = ["wasm", "-q", "-0", "-zcm=masm", "-zld", "-fo=lowered.obj", "lowered.asm"]
    run(command, assembly_work, work / "assemble.log")
    shutil.copyfile(assembly_work / "lowered.obj", work / "lowered.obj")
    obj = read_object(work / "lowered.obj")
    raw_code, public = select_function(obj, compiled["public_symbol"])
    segment = obj["segments"][public["segment_index"]]
    if public["offset"] != 0 or not all(segment["written"]):
        raise ValueError("Compiler object has holes or a hidden prefix")
    linked, fixes, mz = resolve_candidate_fixups(obj, case)
    (work / "code.bin").write_bytes(linked)
    record = {"profile": profile, "IR": compiled["IR"], "source_sha256": sha(work / "input.c"),
              "assembly_sha256": sha(work / "lowered.asm"), "object_sha256": sha(work / "lowered.obj"),
              "code_sha256": sha(work / "code.bin"), "code_hex": linked.hex(),
              "raw_code_hex": raw_code.hex(), "compiled_bytes": len(linked),
              "actual_fixups": fixes, "MZ_segment_word_offsets": mz, "assembler_command": command,
              "assembler_work_directory": str(assembly_work)}
    return record, obj


def vendor_link_control(obj, record, case, work):
    """Independent synthetic WLINK check; no original EXE is consumed."""
    symbol = case["public_symbol"]
    target = next(iter(case["call_placements"]["far"]))
    data = list(case["external_DS_offsets"])
    fixture = [".8086", "FARCODE SEGMENT PARA PUBLIC USE16 'FARCODE'", "FARGROUP GROUP FARCODE",
               "PUBLIC " + target, target + " PROC FAR", "retf", target + " ENDP", "FARCODE ENDS"]
    if data:
        fixture += ["_DATA SEGMENT WORD PUBLIC USE16 'DATA'", "DGROUP GROUP _DATA"]
        for name in data:
            fixture += ["PUBLIC " + name, name + " DW 0"]
        fixture += ["_DATA ENDS"]
    fixture += ["END"]
    (work / "fixture.asm").write_text("\n".join(fixture) + "\n")
    assembly_work = ASSEMBLY_WORK_ROOT / work.name
    shutil.copyfile(work / "fixture.asm", assembly_work / "fixture.asm")
    run(["wasm", "-q", "-0", "-zcm=masm", "-zld", "-fo=fixture.obj", "fixture.asm"], assembly_work, work / "fixture.log")
    run(["wlink", "format", "dos", "option", "nodefaultlibs", "option", "nofarcalls",
         "option", "start=" + symbol, "option", "map=linked.map", "name", "linked.exe",
         "file", "lowered.obj", "file", "fixture.obj"], assembly_work, work / "link.log")
    for name in ("fixture.obj", "linked.exe", "linked.map"):
        shutil.copyfile(assembly_work / name, work / name)
    text = (work / "linked.map").read_text()
    caller, far = public_address(text, symbol), {target: public_address(text, target)}
    ds = {name: public_address(text, name)["offset"] for name in data}
    code, fixes, mz = resolve_function_fixups(obj, symbol, ds, caller, far_symbols=far)
    executable = (work / "linked.exe").read_bytes()
    header = struct.unpack_from("<H", executable, 8)[0] * 16
    offset = header + caller["segment"] * 16 + caller["offset"]
    if executable[offset:offset + len(code)] != code:
        raise ValueError("Actual vendor link differs from OMF resolver")
    count, table = struct.unpack_from("<H", executable, 6)[0], struct.unpack_from("<H", executable, 24)[0]
    locations = [seg * 16 + off - (offset - header) for off, seg in
                 (struct.unpack_from("<HH", executable, table + 4 * n) for n in range(count))]
    if sorted(x for x in locations if 0 <= x < len(code)) != mz:
        raise ValueError("Vendor MZ segment relocation differs")
    return {"vendor_code_equal": True, "caller": caller, "far_targets": far, "DS_offsets": ds,
            "MZ_segment_word_offsets": mz, "EXE_sha256": sha(work / "linked.exe"),
            "fixture_object_sha256": sha(work / "fixture.obj")}


def compare_builds(reference, repeat):
    first = json.loads((reference / "receipt.json").read_text())
    second = json.loads((repeat / "receipt.json").read_text())
    if first != second:
        raise ValueError("Independent compiler receipts differ")
    artifacts, maps = [], []
    names = ("input.c", "IR.json", "lowered.asm", "lowered.obj", "code.bin",
             "fixture.asm", "fixture.obj", "linked.exe")
    for result in first["results"]:
        for name in names:
            a, b = reference / result["case"] / name, repeat / result["case"] / name
            if a.read_bytes() != b.read_bytes():
                raise ValueError("Independent artifact differs: " + str(a))
            artifacts.append({"case": result["case"], "artifact": name, "sha256": sha(a)})
        a, b = reference / result["case"] / "linked.map", repeat / result["case"] / "linked.map"
        lines_a, lines_b = a.read_text().splitlines(), b.read_text().splitlines()
        if len(lines_a) != len(lines_b):
            raise ValueError("Independent map lengths differ")
        differences = []
        for x, y in zip(lines_a, lines_b):
            if x == y:
                continue
            if not any(x.startswith(k) and y.startswith(k) for k in ("Created on:", "Link time:")):
                raise ValueError("Independent map differs beyond diagnostic timing")
            differences.append({"first": x, "repeat": y})
        maps.append({"case": result["case"], "differences": differences,
                     "first_sha256": sha(a), "repeat_sha256": sha(b), "raw_maps_preserved": True})
    return {"complete_receipts_equal": True, "full_artifacts_equal": artifacts,
            "map_metadata_differences": maps, "formal_C_coverage_increment": 0,
            "prototype_receipt_sha256": sha(reference / "receipt.json")}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--controls", type=Path, required=True)
    parser.add_argument("--ida-inventory", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--repeat-reference", type=Path,
                        help="Compare full outputs from an earlier fresh-container build")
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    if sha(Path("/opt/watcom/binl64/wasm")) != "7e216ab56214fe36f80fa60cc757d05f4508683ef28e992969556945fe28f2cd":
        raise ValueError("Use the locked official Wasm revision")
    controls = json.loads((args.controls / "compiled/receipt.json").read_text())
    inventory = json.loads(args.ida_inventory.read_text())
    if set(r["case"] for r in controls["results"]) != set(PROFILES) or inventory["input"]["sha256"] != INPUT_SHA:
        raise ValueError("Control membership or input differs")
    args.output.mkdir(exist_ok=False)
    results, builds = [], {}
    for control in controls["results"]:
        case = control["source_unit"]
        source_path = Path(case["source"])
        if sha(source_path) != control["source_sha256"]:
            raise ValueError("C control source freshness differs")
        source, name = source_path.read_text(), control["case"]
        work = args.output / name
        record, obj = build(source, PROFILES[name], case, work)
        record["vendor_link_control"] = vendor_link_control(obj, record, case, work)
        record["case"], record["source_unit"] = name, case
        results.append(record)
        builds[name] = (source, case, record)
    # The original input is deliberately opened only after all build outputs.
    original_path = ROOT / "assets_raw/DQ3.EXE"
    raw = original_path.read_bytes()
    if len(raw) != 115282 or hashlib.sha256(raw).hexdigest() != INPUT_SHA:
        raise ValueError("Original comparison input differs")
    header, count, table = struct.unpack_from("<H", raw, 8)[0] * 16, struct.unpack_from("<H", raw, 6)[0], struct.unpack_from("<H", raw, 24)[0]
    relocations = [header + seg * 16 + off for off, seg in
                   (struct.unpack_from("<HH", raw, table + 4 * n) for n in range(count))]
    for record in results:
        unit = record["source_unit"]
        start = int(unit["file_start"], 16)
        record["original_byte_exact"] = bytes.fromhex(record["code_hex"]) == raw[start:start + unit["size"]]
        if not record["original_byte_exact"]:
            raise ValueError("Prototype positive control differs from original: " + record["case"])
        original_mz = [x - start for x in relocations if start <= x < start + unit["size"]]
        if sorted(original_mz) != record["MZ_segment_word_offsets"]:
            raise ValueError("Original MZ relocation pattern differs")
        for fix in record["actual_fixups"]:
            if fix.get("placement_kind") != "far":
                continue
            site = int(unit["ida_linear_start"], 16) + fix["offset"] - 1
            target = fix["target_address"]
            linear = 0x10000 + target["segment"] * 16 + target["offset"]
            if not any(x["iscode"] and x["type"] == 16 and int(x["to"], 16) == linear
                       for x in inventory["instructions"][hex(site)]["refs_from"]):
                raise ValueError("Original typed far-call target differs")
    # Changed C semantics must affect generated code; rejected syntax stays closed.
    tests = []
    mutations = [
        ("changed_state", "restore_void", lambda s: s.replace("unknown_DS_259c = 1;", "unknown_DS_259c = 2;")),
        ("changed_argument", "restore_void", lambda s: s.replace("(0x13a)", "(0x13b)")),
        ("allowed_AX_clobber", "register_envelope", lambda s: s.replace("modify exact [es];", "modify exact [ax es];")),
        ("callee_preserves_BX", "register_envelope", lambda s: s.replace("[ax bx cx dx si di bp es]", "[ax cx dx si di bp es]")),
    ]
    for label, name, mutate in mutations:
        source, case, positive = builds[name]
        changed = mutate(source)
        if changed == source:
            raise ValueError("Mutation failed to change its input")
        candidate, _ = build(changed, PROFILES[name], case, args.output / label)
        if candidate["code_hex"] == positive["code_hex"]:
            raise ValueError("Changed C semantics did not affect compiler output")
        tests.append({"case": label, "result": "DIFF", "code_sha256": candidate["code_sha256"]})
    source, case, positive = builds["restore_void"]
    renamed = source.replace("unknown_DS_259c", "opaque_word").replace("sub_21414", "opaque_target").replace("sub_5037", "opaque_entry")
    renamed_case = copy.deepcopy(case)
    renamed_case["public_symbol"] = "_opaque_entry"
    renamed_case["external_DS_offsets"] = {"_opaque_word": "0x259c"}
    renamed_case["call_placements"]["far"] = {"_opaque_target": next(iter(case["call_placements"]["far"].values()))}
    renamed_record, _ = build(renamed, PROFILES["restore_void"], renamed_case, args.output / "renamed_symbols")
    if renamed_record["code_hex"] != positive["code_hex"]:
        raise ValueError("Symbol renaming changed equivalent generated code")
    rejected = [
        ("inline_assembly", source.replace('"_*" far', '"_*" = "nop" far'), PROFILES["restore_void"]),
        ("missing_restore", source.replace("unknown_DS_259c = saved;", "unknown_DS_259c = 0;"), PROFILES["restore_void"]),
        ("word_overflow", source.replace("unknown_DS_259c = 1;", "unknown_DS_259c = 65536;"), PROFILES["restore_void"]),
        ("shadowed_global", source.replace("unsigned saved = unknown_DS_259c;", "unsigned unknown_DS_259c = unknown_DS_259c;"), PROFILES["restore_void"]),
        ("unknown_profile", source, "unreviewed"),
        ("object_function_conflict", source + "\nextern volatile unsigned sub_21414;\n", PROFILES["restore_void"]),
    ]
    ax_source = builds["restore_ax_return"][0]
    before_declaration = ax_source.replace("unsigned result; ", "", 1).replace(
        "return result;", "unsigned result; return result;", 1)
    if before_declaration == ax_source:
        raise ValueError("Declaration-order mutation did not change source")
    rejected += [
        ("use_before_declaration", before_declaration, PROFILES["restore_ax_return"]),
        ("unsupported_preserved_AX_result", ax_source.replace("[ax bx cx dx si di bp es]", "[bx cx dx si di bp es]", 1),
         PROFILES["restore_ax_return"]),
    ]
    for label, changed, profile in rejected:
        try:
            compile_source(changed, profile)
        except UnsupportedC as error:
            tests.append({"case": label, "result": "REFUSED", "reason": str(error)})
        else:
            raise ValueError("Unsupported compiler input accepted: " + label)
    result = {"prototype_only": True, "formal_C_coverage_increment": 0, "optimizer_reads_original_EXE": False,
              "input_sha256": INPUT_SHA, "compiler_stage_sha256": sha(ROOT / "tools/prototype_c_abi_stage.py"),
              "producer_sha256": sha(Path(__file__)), "controls_receipt_sha256": sha(args.controls / "receipt.json"),
              "assembler_sha256": sha(Path("/opt/watcom/binl64/wasm")), "linker_sha256": sha(Path("/opt/watcom/binl64/wlink")),
              "results": results, "source_mutation_and_refusal_tests": tests,
              "renaming_preserves_code": True, "whole_EXE_complete": False,
              "scope": "Two opt-in C-subset ABI profiles; no formal adoption or original compiler identity claim"}
    (args.output / "receipt.json").write_text(json.dumps(result, indent=2) + "\n")
    if args.repeat_reference:
        comparison = compare_builds(args.repeat_reference, args.output)
        (args.output / "repeat-comparison.json").write_text(json.dumps(comparison, indent=2) + "\n")
    print(json.dumps({"exact_controls": sum(r["original_byte_exact"] for r in results),
                      "tests": len(tests), "formal_C_coverage_increment": 0}))


if __name__ == "__main__":
    main()
