"""Docker-only 16-bit C register-ABI controls; docs/25 and tools/build/README.

Only instruction-free ABI pragmas are used. A candidate compiler is not evidence
of the original toolchain. Keep raw OMF and unsupported relocations visible.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import time

from omf_matching_probe import UnsupportedOMF, read_object, resolve_ds_offsets, select_function
from omf_call_fixups import resolve_candidate_fixups, validate_original_caller


ROOT = Path(__file__).resolve().parents[1]
COMPILER_PROFILES = {
    "cdecl-size-reorder": ["-bt=dos", "-ms", "-0", "-os", "-oi", "-s", "-ofr", "-ecc", "-zld"],
    "watcall-speed-no-reorder": ["-bt=dos", "-ms", "-0", "-ot", "-oi", "-s", "-of", "-ecw", "-zld"],
    "cdecl-size-calls": ["-bt=dos", "-ms", "-0", "-os", "-oi", "-s", "-ofr", "-ecc", "-zld", "-oc"],
}


def compiler_flags_for(case):
    profile = case.get("compiler_profile", "cdecl-size-reorder")
    if profile not in COMPILER_PROFILES:
        raise ValueError("Unknown reviewed compiler profile: " + str(profile))
    return list(COMPILER_PROFILES[profile])


CONTROLS = {
    "bxecho": ("unsigned bxecho(unsigned value);\n#pragma aux bxecho \"_*\" parm [bx] value [ax] modify exact [ax];\nunsigned bxecho(unsigned value) { return value; }\n", "_bxecho", {}),
    "siecho": ("unsigned siecho(unsigned value);\n#pragma aux siecho \"_*\" parm [si] value [ax] modify exact [ax];\nunsigned siecho(unsigned value) { return value; }\n", "_siecho", {}),
    "axstore": ("extern volatile unsigned unknown_ds_0032;\nvoid axstore(unsigned value);\n#pragma aux axstore \"_*\" parm [ax] modify exact [];\nvoid axstore(unsigned value) { unknown_ds_0032 = value; }\n", "_axstore", {"_unknown_ds_0032": 0x32}),
    "rngcore": ("#include <stdlib.h>\nextern volatile unsigned unknown_ds_0b5a;\nunsigned rngcore(void);\n#pragma aux rngcore \"_*\" value [ax] modify exact [ax];\nunsigned rngcore(void) { return unknown_ds_0b5a = _rotl(_rotl(_rotl(unknown_ds_0b5a + 0x9018u, 1), 1), 1); }\n", "_rngcore", {"_unknown_ds_0b5a": 0xb5a}),
    "rngbound": ("#include <stdlib.h>\nextern volatile unsigned unknown_ds_0b5a;\nunsigned long rngbound(unsigned bound, unsigned incoming_ax);\n#pragma aux rngbound \"_*\" parm [bx] [ax] value [dx ax] modify exact [ax dx];\nunsigned long rngbound(unsigned bound, unsigned incoming_ax) { unsigned value; if (!bound) return (unsigned long)incoming_ax; value = _rotl(_rotl(_rotl(unknown_ds_0b5a + 0x9018u, 1), 1), 1); unknown_ds_0b5a = value; return ((unsigned long)(value % bound) << 16) | (unsigned long)(value / bound); }\n", "_rngbound", {"_unknown_ds_0b5a": 0xb5a}),
    "rngshift": ("extern volatile unsigned unknown_ds_0b5a;\nunsigned rngshift(void);\n#pragma aux rngshift \"_*\" value [ax] modify exact [ax];\nunsigned rngshift(void) { unsigned value = unknown_ds_0b5a + 0x9018u; value = (value << 1) | (value >> 15); value = (value << 1) | (value >> 15); value = (value << 1) | (value >> 15); unknown_ds_0b5a = value; return value; }\n", "_rngshift", {"_unknown_ds_0b5a": 0xb5a}),
}


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--candidate-manifest", type=Path)
    parser.add_argument("--ida-inventory", type=Path)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    manifest = json.loads(Path("/opt/watcom/source-manifest.json").read_text())
    if not manifest["complete_archive_verified"] or manifest["release"] != "2026-10-01-Build":
        raise ValueError("Compiler payload identity differs")
    for record in manifest["files"]:
        path = Path("/opt/watcom") / record["path"]
        if sha(path) != record["sha256"]:
            raise ValueError("Compiler/header payload differs")
    raw = (ROOT / "assets_raw/DQ3.EXE").read_bytes()
    if hashlib.sha256(raw).hexdigest() != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c":
        raise ValueError("Original input differs")
    header = int.from_bytes(raw[8:10], "little") * 16
    controls = CONTROLS
    original_ranges = {"axstore": (0x14ae6, 4), "rngcore": (0xe6b9, 16), "rngbound": (0xe6c9, 30), "rngshift": (0xe6b9, 16)}
    candidate_records = {}
    if args.candidate_manifest is not None:
        if args.ida_inventory is None:
            raise ValueError("C source candidates require the original IDA inventory")
        selected = json.loads(args.candidate_manifest.read_text())
        inventory = json.loads(args.ida_inventory.read_text())
        if selected["input_sha256"] != hashlib.sha256(raw).hexdigest() or inventory["input"]["sha256"] != selected["input_sha256"]:
            raise ValueError("Candidate/inventory input differs")
        controls = {}
        original_ranges = {}
        for case in selected["cases"]:
            compiler_flags_for(case)
            validate_original_caller(case, inventory)
            name = case["id"].lower()
            if not re.fullmatch(r"[a-z0-9_]{1,32}", name):
                raise ValueError("Unsafe candidate filename")
            original_start = int(case["ida_linear_start"], 16)
            function = next(f for f in inventory["functions"] if int(f["ida_linear_start"], 16) == original_start)
            if function["chunks"] != [[hex(original_start), hex(original_start + case["size"])]]:
                raise ValueError("Candidate source-unit boundary is not complete and contiguous")
            original_code = b"".join(bytes.fromhex(inventory["instructions"][address]["file_bytes"]) for address in function["instructions"])
            file_start = int(case["file_start"], 16)
            if original_code != raw[file_start:file_start + case["size"]] or file_start != header + int(case["logical_start"], 16):
                raise ValueError("Candidate original range/address bases differ")
            content = (ROOT / case["source"]).read_text(encoding="ascii")
            if re.search(r"\b(__asm|_asm)\b|#pragma\s+aux[^\n]*=", content):
                raise ValueError("Primary C source contains assembly instructions")
            placements = {symbol: int(value, 16) for symbol, value in case["external_DS_offsets"].items()}
            controls[name] = (content, case["public_symbol"], placements)
            original_ranges[name] = (int(case["logical_start"], 16), case["size"])
            candidate_records[name] = case
    args.output.mkdir(exist_ok=False)
    compile_root = Path("/tmp/watcom16-compile")
    compile_root.mkdir(exist_ok=False)
    results = []
    started = time.monotonic()
    for name, (content, symbol, placements) in controls.items():
        source = args.output / (name + ".c")
        source.write_text(content, encoding="ascii")
        shutil.copyfile(source, compile_root / source.name)
        obj_path = args.output / (name + ".obj")
        command = ["wcc"] + compiler_flags_for(candidate_records.get(name, {})) + ["-fo=" + obj_path.name, source.name]
        with (args.output / (name + ".log")).open("wb") as stream:
            process = subprocess.run(command, cwd=compile_root, stdout=stream, stderr=subprocess.STDOUT, timeout=30, check=False)
        if (compile_root / obj_path.name).is_file():
            shutil.copyfile(compile_root / obj_path.name, obj_path)
        result = {"case": name, "source_sha256": sha(source), "compiler_command": command,
                  "compiler_returncode": process.returncode, "source_has_assembly_instructions": False,
                  "scope": "Authored compiler control, not original-language evidence"}
        if name in candidate_records:
            result["source_unit"] = candidate_records[name]
            result["scope"] = "Original-ID-addressed readable C candidate; source type/module meaning remains scoped"
        if process.returncode or not obj_path.is_file():
            result.update(status="COMPILE_FAILED", log=(args.output / (name + ".log")).read_text(errors="replace"))
            results.append(result)
            continue
        subprocess.run(["wdis", "-l=" + str(args.output / (name + ".lst")), str(obj_path)], check=True, timeout=20,
                       stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT)
        result["object_sha256"] = sha(obj_path)
        try:
            obj = read_object(obj_path)
            code, public = select_function(obj, symbol)
            result.update(raw_code_hex=code.hex(), public=public, fixups=obj["fixups"], externals=obj["externals"],
                          segments=[{"name": segment["name"], "class": segment["class"], "length": segment["length"]} for segment in obj["segments"][1:]])
            mode = candidate_records.get(name, {}).get("encoded_addend_mode", "unsigned16")
            if mode not in ("unsigned16", "signed16"):
                raise ValueError("Unknown encoded addend contract")
            if name in candidate_records:
                linked, fixes, mz_segment_offsets = resolve_candidate_fixups(obj, candidate_records[name])
            else:
                linked, fixes = resolve_ds_offsets(obj, symbol, placements, signed_addends=mode == "signed16")
                mz_segment_offsets = []
        except (UnsupportedOMF, ValueError) as error:
            result.update(status="REFUSED", reason=str(error))
        else:
            (args.output / (name + "-code.bin")).write_bytes(linked)
            result.update(status="RESOLVED", code_hex=linked.hex(), code_size=len(linked), applied_fixups=fixes,
                          MZ_segment_word_offsets=mz_segment_offsets,
                          code_sha256=hashlib.sha256(linked).hexdigest())
            if name in original_ranges:
                offset, size = original_ranges[name]
                original = raw[header + offset:header + offset + size]
                result["original_compare"] = {"logical": hex(offset), "size": size,
                                              "original_hex": original.hex(), "byte_exact": linked == original}
        results.append(result)
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__)),
               "call_resolver_sha256": sha(ROOT / "tools/omf_call_fixups.py"),
               "OMF_parser_sha256": sha(ROOT / "tools/omf_matching_probe.py"),
               "compiler_payload_manifest_sha256": sha(Path("/opt/watcom/source-manifest.json")),
               "compiler_release": manifest["release"], "results": results, "wall_seconds": time.monotonic() - started,
               "candidate_manifest_sha256": sha(args.candidate_manifest) if args.candidate_manifest is not None else None,
               "IDA_inventory_sha256": sha(args.ida_inventory) if args.ida_inventory is not None else None,
               "original_compiler": "unknown", "whole_goal_complete": False}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"results": [(result["case"], result["status"], result.get("code_size")) for result in results],
                      "seconds": receipt["wall_seconds"]}))


if __name__ == "__main__":
    main()
