"""Docker-only C/ASM synthetic call controls; entry: docs/25-match-progress.md.

Compile readable C, then compare fixed-placement relocation with real WLINK.
The fixture has no original DQ3 input and contributes zero source coverage.
"""

import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import re
import struct
import subprocess

from omf_call_fixups import resolve_function_fixups
from omf_matching_probe import UnsupportedOMF, read_object, select_function
from run_watcom16_abi import compiler_flags_for


C_SOURCE = '''extern volatile unsigned flag;
void near_target(void);
#pragma aux near_target "_*" modify exact [];
void far_target(void);
#pragma aux far_target "_*" far modify exact [];
void call_probe(void);
#pragma aux call_probe "_*";
void call_probe(void) {
    flag = 1;
    near_target();
    flag = 2;
    far_target();
    flag = 3;
}
'''

ASM_SOURCE = '''.8086
_TEXT SEGMENT BYTE PUBLIC 'CODE'
PUBLIC _near_target, _small_code_
ORG 0180h
_small_code_ LABEL NEAR
_near_target PROC NEAR
 ret
_near_target ENDP
_TEXT ENDS
FAR_TEXT SEGMENT PARA PUBLIC 'CODE'
PUBLIC _far_target
ORG 002Dh
_far_target PROC FAR
 retf
_far_target ENDP
FAR_TEXT ENDS
_DATA SEGMENT WORD PUBLIC 'DATA'
PUBLIC _flag
ORG 1234h
_flag DW 0
_DATA ENDS
DGROUP GROUP _DATA
FARGROUP GROUP FAR_TEXT
END
'''


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def public_address(text, symbol):
    match = re.search(r"^\s*([0-9a-fA-F]+):([0-9a-fA-F]+)\s+" + re.escape(symbol) + r"\s*$", text, re.M)
    if not match:
        raise ValueError("Missing public in the vendor map: " + symbol)
    return {"segment": int(match[1], 16), "offset": int(match[2], 16)}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    payload = Path("/opt/watcom/source-manifest.json")
    locked = json.loads(payload.read_text())
    if not locked["complete_archive_verified"] or locked["release"] != "2026-10-01-Build":
        raise ValueError("Unreviewed compiler payload")
    for item in locked["files"]:
        if sha(Path("/opt/watcom") / item["path"]) != item["sha256"]:
            raise ValueError("Compiler payload changed")
    if sha(Path("/opt/watcom/binl64/wasm")) != "7e216ab56214fe36f80fa60cc757d05f4508683ef28e992969556945fe28f2cd":
        raise ValueError("Unreviewed assembler payload")
    args.output.mkdir(exist_ok=False)
    work = Path("/tmp/watcom-call-fixture")
    work.mkdir(exist_ok=False)
    for name, text in [("call.c", C_SOURCE), ("fixture.asm", ASM_SOURCE)]:
        (work / name).write_text(text, encoding="ascii")
        (args.output / name).write_text(text, encoding="ascii")
    compiler = ["wcc"] + compiler_flags_for({}) + ["-fo=call.obj", "call.c"]
    for command in [compiler, ["wasm", "-q", "-0", "-zld", "-fo=fixture.obj", "fixture.asm"]]:
        result = subprocess.run(command, cwd=work, capture_output=True, text=True, timeout=20)
        if result.returncode:
            raise ValueError(result.stdout + result.stderr)
    obj = read_object(work / "call.obj")
    original_code, _ = select_function(obj, "_call_probe")
    variants = []
    for label, files in [("forward", ["call.obj", "fixture.obj"]), ("backward", ["fixture.obj", "call.obj"])]:
        command = ["wlink", "format", "dos", "option", "nodefaultlibs", "option", "nofarcalls",
                   "option", "map=" + label + ".map", "option", "start=_call_probe", "name", label + ".exe"]
        for name in files:
            command += ["file", name]
        result = subprocess.run(command, cwd=work, capture_output=True, text=True, timeout=20)
        (args.output / (label + ".log")).write_text(result.stdout + result.stderr)
        if result.returncode:
            raise ValueError("WLINK fixture failed")
        map_text = (work / (label + ".map")).read_text()
        caller = public_address(map_text, "_call_probe")
        near = {"_near_target": public_address(map_text, "_near_target")}
        far = {"_far_target": public_address(map_text, "_far_target")}
        data = {"_flag": public_address(map_text, "_flag")["offset"]}
        linked, fixes, relocations = resolve_function_fixups(obj, "_call_probe", data, caller,
                                                           near_symbols=near, far_symbols=far)
        raw = (work / (label + ".exe")).read_bytes()
        if raw[:2] != b"MZ":
            raise ValueError("Vendor output is not MZ")
        header = struct.unpack_from("<H", raw, 8)[0] * 16
        ip, cs = struct.unpack_from("<HH", raw, 20)
        if (cs, ip) != (caller["segment"], caller["offset"]):
            raise ValueError("MZ entry and public map disagree")
        entry = header + cs * 16 + ip
        if raw[entry:entry + len(linked)] != linked:
            raise ValueError("Actual WLINK code differs from reviewed FIXUPP arithmetic")
        count = struct.unpack_from("<H", raw, 6)[0]
        table = struct.unpack_from("<H", raw, 24)[0]
        actual_relocations = []
        for n in range(count):
            offset, segment = struct.unpack_from("<HH", raw, table + 4 * n)
            actual_relocations.append(segment * 16 + offset - (entry - header))
        if sorted(actual_relocations) != relocations:
            raise ValueError("Actual MZ segment relocations differ")
        variants.append({"case": label, "caller": caller, "near": near, "far": far,
                         "ds_offsets": data, "code_hex": linked.hex(), "fixups": fixes,
                         "MZ_segment_word_offsets": relocations, "vendor_equal": True,
                         "MZ_sha256": sha(work / (label + ".exe"))})
    v = variants[0]
    negatives = []

    def reject(label, changed=None, **kwargs):
        options = {"near_symbols": copy.deepcopy(v["near"]), "far_symbols": copy.deepcopy(v["far"])}
        options.update(kwargs)
        try:
            resolve_function_fixups(obj if changed is None else changed, "_call_probe",
                                    v["ds_offsets"], v["caller"], **options)
        except (ValueError, UnsupportedOMF):
            negatives.append({"case": label, "rejected": True})
        else:
            raise ValueError("Invalid call placement accepted: " + label)

    reject("missing-near", near_symbols={})
    reject("missing-far", far_symbols={})
    reject("conflicting-kind", near_symbols={"_near_target": v["near"]["_near_target"],
                                            "_flag": v["caller"]})
    reject("cross-frame-near", near_symbols={"_near_target": {"segment": 1, "offset": 0x19B}})
    reject("far-segment-overflow", far_symbols={"_far_target": {"segment": 65536, "offset": 0x2D}})
    reject("far-offset-negative", far_symbols={"_far_target": {"segment": 0x1A, "offset": -1}})
    for label, target_offset, key, value in [("unsupported-frame", 7, "frame", {"method": 0, "datum": 1}),
                                            ("wrong-location", 16, "location_type", 1),
                                            ("near-mode", 7, "segment_relative", True),
                                            ("explicit-addend", 16, "displacement", 1),
                                            ("cross-module", 7, "segment_index", 2),
                                            ("outside-function", 16, "offset", len(original_code) - 2)]:
        changed = copy.deepcopy(obj)
        next(f for f in changed["fixups"] if f["offset"] == target_offset)[key] = value
        reject(label, changed)
    changed = copy.deepcopy(obj)
    changed["fixups"].append(copy.deepcopy(changed["fixups"][1]))
    reject("overlap", changed)
    for label, offset in [("nonzero-implicit-addend", 7), ("non-call-opcode", 15)]:
        changed = copy.deepcopy(obj)
        changed["segments"][1]["bytes"][offset] = 1
        reject(label, changed)
    import shutil
    for path in work.iterdir():
        if path.is_file():
            shutil.copyfile(path, args.output / path.name)
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__)),
               "call_resolver_sha256": sha(Path(__file__).with_name("omf_call_fixups.py")),
               "OMF_parser_sha256": sha(Path(__file__).with_name("omf_matching_probe.py")),
               "compiler_payload_manifest_sha256": sha(payload), "compiler_command": compiler,
               "C_source_sha256": sha(work / "call.c"), "ASM_source_sha256": sha(work / "fixture.asm"),
               "C_object_sha256": sha(work / "call.obj"), "ASM_object_sha256": sha(work / "fixture.obj"),
               "linker_sha256": sha(Path("/opt/watcom/binl64/wlink")),
               "assembler_sha256": sha(Path("/opt/watcom/binl64/wasm")),
               "variants": variants, "negative_cases": negatives, "whole_goal_complete": False,
               "source_coverage_increment": 0,
               "scope": "Synthetic WCC zero-addend F5/T2 calls with explicit group and NOFARCALLS; not original layout"}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"vendor_equal_variants": len(variants), "negative_cases": len(negatives),
                      "source_coverage_increment": 0}))


if __name__ == "__main__":
    main()
