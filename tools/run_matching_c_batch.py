"""Docker-only exact C batch for the full matching Goal; entry in docs/25.

Compare only freshly compiled, fully selected and actually relocated OMF code.
No instruction-byte masks, binary arrays or original-byte candidate generation.
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

from omf_matching_probe import UnsupportedOMF, read_object, resolve_ds_offsets
from run_matching_probe import COMPILER_HASHES, ORIGINAL_HASH, compare


ROOT = Path(__file__).resolve().parents[1]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def msc_listing_function(code, listing_path, symbol):
    """Use the compiler's PROC/ENDP range, never a guessed RET/NOP cutoff.

    Only a single near PROC at public offset zero is supported. Validate all
    listing bytes against materialized OMF bytes, including post-ENDP padding.
    The padding remains recorded and is not called source function code.
    """
    lines = listing_path.read_text(encoding="ascii").splitlines()
    opened = closed = False
    function_end = 0
    covered = set()
    outside = []
    for line in lines:
        if re.match(r"^" + re.escape(symbol) + r"\s+PROC\s+NEAR\s*$", line):
            if opened or closed:
                raise ValueError("Duplicate compiler PROC")
            opened = True
            continue
        if re.match(r"^" + re.escape(symbol) + r"\s+ENDP\s*$", line):
            if not opened or closed or not function_end:
                raise ValueError("Invalid compiler ENDP")
            closed = True
            continue
        match = re.match(r"^\s*\*\*\*\s+([0-9a-fA-F]{6})\s+((?:[0-9a-fA-F]{2}[ \t]+)+)", line)
        if not match:
            continue
        if not opened:
            raise UnsupportedOMF("Listing contains bytes before the selected PROC")
        offset = int(match.group(1), 16)
        encoded = bytes.fromhex(match.group(2))
        if offset + len(encoded) > len(code) or code[offset:offset + len(encoded)] != encoded:
            raise ValueError("Compiler listing bytes differ from raw OMF")
        if any(address in covered for address in range(offset, offset + len(encoded))):
            raise ValueError("Compiler listing has overlapping ranges")
        covered.update(range(offset, offset + len(encoded)))
        if closed:
            outside.append({"offset": offset, "bytes": encoded.hex(),
                            "scope": "compiler-emitted bytes after ENDP; not part of the selected C function"})
        else:
            if offset != function_end:
                raise UnsupportedOMF("Compiler PROC contains an unverified listing gap")
            function_end += len(encoded)
    if not closed or covered != set(range(len(code))):
        raise UnsupportedOMF("Compiler listing does not cover the complete OMF code segment")
    return function_end, {"path": str(listing_path), "sha256": sha(listing_path), "public_symbol": symbol,
                          "compiler_PROC_offset": 0, "compiler_ENDP_offset": function_end,
                          "materialized_module_size": len(code), "post_ENDP_bytes": outside}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--ida-evidence", type=Path, required=True)
    parser.add_argument("--manifest", type=Path, default=ROOT / "tools/matching_c_manifest.json")
    parser.add_argument("--msc-root", type=Path, default=Path("/msc"))
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Run as the host UID/GID")
    manifest = json.loads(args.manifest.read_text())
    source_input = ROOT / manifest["input_path"]
    raw = source_input.read_bytes()
    if (manifest["input_sha256"] != ORIGINAL_HASH or manifest["input_size"] != 115282
            or sha(source_input) != ORIGINAL_HASH or len(raw) != 115282):
        raise ValueError("Original input identity differs")
    evidence = json.loads(args.ida_evidence.read_text())
    if evidence["input"]["sha256"] != ORIGINAL_HASH or evidence["tool"]["version"] != "9.4":
        raise ValueError("IDA evidence identity/version differs")
    if not 1 <= len(manifest["cases"]) <= 32:
        raise ValueError("Batch size exceeds the explicit 1..32 bound")
    if len({case["id"] for case in manifest["cases"]}) != len(manifest["cases"]):
        raise ValueError("Duplicate C candidate ID")
    ranges = sorted((int(case["file_start"], 16), int(case["file_start"], 16) + case["size"])
                    for case in manifest["cases"])
    if any(hi > next_lo for (_, hi), (next_lo, _) in zip(ranges, ranges[1:])):
        raise ValueError("Overlapping C function replacements")
    header = int.from_bytes(raw[8:10], "little") * 16
    compiler = {}
    for name, expected in COMPILER_HASHES.items():
        component = args.msc_root / "BIN" / name
        if sha(component) != expected:
            raise ValueError("Compiler component differs: " + name)
        compiler[name] = {"sha256": expected, "size": component.stat().st_size}
    args.output.mkdir(exist_ok=False)
    commands = ["@echo off", "set PATH=C:\\BIN", "set INCLUDE=C:\\INCLUDE\\INCLUDE",
                "set LIB=C:\\LIB", "set TMP=D:\\", "d:"]
    cases = []
    for number, case in enumerate(manifest["cases"]):
        start = int(case["ida_linear_start"], 16)
        function = next(f for f in evidence["functions"] if int(f["ida_linear_start"], 16) == start)
        logical, file_start, size = int(case["logical_start"], 16), int(case["file_start"], 16), case["size"]
        if start != logical + int(evidence["address_space"]["load_base"], 16) or file_start != header + logical:
            raise ValueError("Candidate address bases differ")
        if function["chunks"] != [[hex(start), hex(start + size)]]:
            raise ValueError("Candidate is not one complete contiguous IDA function")
        original = raw[file_start:file_start + size]
        instructions = function["instructions"]
        if b"".join(bytes.fromhex(row["file_bytes"]) for row in instructions) != original:
            raise ValueError("Full IDA function bytes differ")
        cursor = file_start
        for row in instructions:
            if int(row["file_offset"], 16) != cursor:
                raise ValueError("IDA instruction order/coverage differs")
            cursor += len(bytes.fromhex(row["file_bytes"]))
        source = ROOT / case["source"]
        content = source.read_text()
        if re.search(r"\b(__asm|_asm)\b|\b(asm|db|incbin)\s*\(", content):
            raise ValueError("C-only manifest contains inline assembly or binary injection")
        if case["compiler_flags"] not in ("/c /AS /Ox", "/c /AS /Os", "/c /AS /Os /Fc", "/c /AS /Os /Gs /Fc"):
            raise ValueError("Unreviewed compiler flags in the C batch")
        stem = "c" + str(number).zfill(3)
        shutil.copyfile(source, args.output / (stem + ".c"))
        commands += ["C:\\BIN\\CL.EXE " + case["compiler_flags"] + " " + stem + ".c > " + stem + ".log",
                     "if errorlevel 1 goto failed"]
        cases.append({**case, "stem": stem, "source_sha256": sha(source),
                      "original_hex": original.hex(), "original_size": size})
    commands += ["echo SUCCESS > DONE.TXT", "goto end", ":failed", "echo FAIL > DONE.TXT", ":end"]
    (args.output / "go.bat").write_text("\r\n".join(commands) + "\r\n", encoding="ascii")
    config = ("[sdl]\noutput=surface\n[cpu]\ncycles=fixed 100000\n[mixer]\nnosound=true\n"
              "[autoexec]\nmount c " + str(args.msc_root) + "\nmount d " + str(args.output)
              + "\nd:\ncall go.bat\nexit\n")
    (args.output / "dosbox.conf").write_text(config, encoding="ascii")
    started = time.monotonic()
    with (args.output / "dosbox.log").open("wb") as stream:
        process = subprocess.run(["dosbox", "-conf", str(args.output / "dosbox.conf"), "-exit"],
                                 stdout=stream, stderr=subprocess.STDOUT, timeout=120, check=False)
    seconds = time.monotonic() - started
    marker = args.output / "DONE.TXT"
    if process.returncode or not marker.is_file() or marker.read_text().strip() != "SUCCESS":
        raise ValueError("Compiler batch failed; preserve and inspect logs")
    results = []
    for case in cases:
        obj = read_object(args.output / (case["stem"].upper() + ".OBJ"))
        placements = {name: int(offset, 16) for name, offset in case["external_DS_offsets"].items()}
        result = {**case, "object_sha256": obj["sha256"]}
        try:
            module, fixups = resolve_ds_offsets(obj, case["public_symbol"], placements)
        except UnsupportedOMF as error:
            result.update(status="REFUSED", reason=str(error), exact=False)
        else:
            original = bytes.fromhex(case["original_hex"])
            result["module_compare"] = compare(original, module)
            listing = args.output / (case["stem"].upper() + ".COD")
            extent = len(module)
            if listing.is_file():
                # The listing describes pre-relocation bytes, so validate it
                # against the original OMF LEDATA, then apply the reviewed
                # relocations to that same independently identified function.
                selected = next(p for p in obj["publics"] if p["name"] == case["public_symbol"])
                raw_code = bytes(obj["segments"][selected["segment_index"]]["bytes"])
                extent, result["compiler_function_extent"] = msc_listing_function(raw_code, listing, case["public_symbol"])
                if any(f["offset"] + 2 > extent for f in fixups):
                    raise UnsupportedOMF("Relocation lies outside compiler PROC")
            code = module[:extent]
            result.update(status="MATCH" if original == code else "DIFF", exact=original == code,
                          comparison=compare(original, code), fixups=fixups,
                          candidate_sha256=hashlib.sha256(code).hexdigest(),
                          module_padding_placement="unresolved" if len(module) != extent else "not present")
            (args.output / (case["stem"] + "-module.bin")).write_bytes(module)
            (args.output / (case["stem"] + "-linked.bin")).write_bytes(code)
        results.append(result)
    exact_cases = [case for case in results if case["exact"]]
    scaffold = bytearray(raw)
    for case in exact_cases:
        start = int(case["file_start"], 16)
        code = (args.output / (case["stem"] + "-linked.bin")).read_bytes()
        scaffold[start:start + case["size"]] = code
    if hashlib.sha256(scaffold).hexdigest() != ORIGINAL_HASH:
        raise ValueError("Exact C replacement changed original bytes")
    (args.output / "partial-C-scaffold.exe").write_bytes(scaffold)
    receipt = {"schema_version": 1, "input": {"path": manifest["input_path"], "size": len(raw), "sha256": ORIGINAL_HASH},
               "ida_evidence": {"path": str(args.ida_evidence), "sha256": sha(args.ida_evidence), "address_space": evidence["address_space"]},
               "manifest_sha256": sha(args.manifest), "producer_sha256": sha(Path(__file__)),
               "omf_parser_sha256": sha(ROOT / "tools/omf_matching_probe.py"), "compiler_inputs": compiler,
               "compile_batch_seconds": seconds, "cases": results, "C_exact_functions": len(exact_cases),
               "C_exact_bytes": sum(case["size"] for case in exact_cases),
               "retained_original_bytes": len(raw) - sum(case["size"] for case in exact_cases),
               "partial_scaffold_sha256": ORIGINAL_HASH,
               "scope": "Compiler PROC/ENDP function-code match only; module-padding placement and retained original bytes do not count toward full Goal",
               "original_compiler": "unknown", "whole_EXE_source_complete": False}
    (args.output / "receipt.json").write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({"cases": [(case["id"], case["status"]) for case in results],
                      "C_exact_functions": len(exact_cases), "C_exact_bytes": receipt["C_exact_bytes"], "seconds": seconds}))


if __name__ == "__main__":
    main()
