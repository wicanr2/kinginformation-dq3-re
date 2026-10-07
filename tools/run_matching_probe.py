"""Run the bounded matching experiment from docs/25 entirely inside Docker.

Compile original C candidates with the mounted MSC 5.1 candidate toolchain,
resolve actual OMF external fixups when supported, and assemble one IDA-reviewed
RNG routine. ASM coverage, C coverage and retained original bytes stay separate.
"""

import argparse
import copy
import hashlib
import json
import os
from pathlib import Path
import shutil
import struct
import subprocess
import time

from omf_matching_probe import UnsupportedOMF, read_object, resolve_ds_offsets, select_function


ROOT = Path(__file__).resolve().parents[1]
ORIGINAL_HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
COMPILER_HASHES = {
    "CL.EXE": "cad40cef732beea3455225fbd62286c664166db7c3f87e2402e22b2e3f107771",
    "C1.EXE": "8d33c6d87aa74b042b7d7cb01a23241c4cc3b1b8b446b8f1548bd167e3c7bf62",
    "C2.EXE": "a513c98bc2ce11bd00901b560252b5727ab9b6bacd9319b9eaba08f9e314b99b",
    "C3.EXE": "c1269628534d5deb48171ba9e42f761dcb1c28a0087c73780db0ee2e4352e6fa",
}


def digest(data):
    return hashlib.sha256(data).hexdigest()


def compare(original, candidate):
    """A source candidate is exact only when the complete byte sequences agree."""
    return {"original_size": len(original), "candidate_size": len(candidate),
            "equal_bytes": sum(a == b for a, b in zip(original, candidate)),
            "exact": original == candidate,
            "original_hex": original.hex(), "candidate_hex": candidate.hex()}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--ida-evidence", type=Path, required=True)
    parser.add_argument("--msc-root", type=Path, default=Path("/msc"))
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Run with the host UID/GID, not root")
    args.output.mkdir(exist_ok=False)
    raw = (ROOT / "assets_raw/DQ3.EXE").read_bytes()
    if len(raw) != 115282 or digest(raw) != ORIGINAL_HASH:
        raise ValueError("Original executable identity differs")
    header = struct.unpack_from("<H", raw, 8)[0] * 16
    evidence_raw = args.ida_evidence.read_bytes()
    evidence = json.loads(evidence_raw)
    if evidence["input"]["sha256"] != ORIGINAL_HASH or evidence["tool"]["version"] != "9.4":
        raise ValueError("IDA evidence identity or version differs")
    rng = [f for f in evidence["functions"] if f["ida_linear_start"] == "0x1e6b9"]
    if len(rng) != 1 or len(rng[0]["instructions"]) != 7:
        raise ValueError("Reviewed RNG function range is unavailable")
    original_rng = b"".join(bytes.fromhex(r["file_bytes"]) for r in rng[0]["instructions"])
    if original_rng != raw[header + 0xE6B9:header + 0xE6C9]:
        raise ValueError("IDA RNG instruction bytes differ from original")
    compiler_inputs = {}
    for name, expected in COMPILER_HASHES.items():
        p = args.msc_root / "BIN" / name
        actual = digest(p.read_bytes())
        if actual != expected:
            raise ValueError("Mounted compiler component differs: " + name)
        compiler_inputs[name] = {"sha256": actual, "size": p.stat().st_size}

    batch = args.output / "batch"
    batch.mkdir()
    cases = []
    manifest = json.loads((ROOT / "tools/re_match_manifest.json").read_text())
    for n, entry in enumerate(manifest):
        stem = "base" + str(n)
        source = ROOT / "re/match" / ("sub_" + entry["off"] + ".c")
        content = source.read_bytes()
        (batch / (stem + ".c")).write_bytes(content)
        cases.append({"stem": stem, "source": str(source), "source_sha256": digest(content),
                      "logical": int(entry["off"], 16), "original_size": entry["size"],
                      "symbol": "_sub_" + entry["off"], "flags": "/c /AS /Ox"})
    variants = {
        "rotone": "return g_rng = _rotl(_rotl(_rotl(g_rng + 0x9018, 1), 1), 1);",
        "shifts": "g_rng += 0x9018; g_rng = (g_rng << 3) | (g_rng >> 13); return g_rng;",
    }
    for stem, body in variants.items():
        content = ("#include <stdlib.h>\n#pragma intrinsic(_rotl)\nunsigned g_rng;\n"
                   "unsigned sub_e6b9(void) { " + body + " }\n").encode()
        (batch / (stem + ".c")).write_bytes(content)
        cases.append({"stem": stem, "source": "generated research candidate",
                      "source_sha256": digest(content), "logical": 0xE6B9, "original_size": 16,
                      "symbol": "_sub_e6b9", "flags": "/c /AS /Ox"})
    commands = ["@echo off", "set PATH=C:\\BIN", "set INCLUDE=C:\\INCLUDE\\INCLUDE",
                "set LIB=C:\\LIB", "set TMP=D:\\", "d:"]
    for case in cases:
        commands.extend(["C:\\BIN\\CL.EXE " + case["flags"] + " " + case["stem"] + ".c > "
                         + case["stem"] + ".log", "if errorlevel 1 goto failed"])
    commands.extend(["echo SUCCESS > DONE.TXT", "goto end", ":failed", "echo FAIL > DONE.TXT", ":end"])
    (batch / "go.bat").write_text("\r\n".join(commands) + "\r\n", encoding="ascii")
    config = ("[sdl]\noutput=surface\n[dosbox]\nmemsize=16\n[cpu]\ncycles=fixed 100000\n"
              "[mixer]\nnosound=true\n[autoexec]\nmount c " + str(args.msc_root)
              + "\nmount d " + str(batch) + "\nd:\ncall go.bat\nexit\n")
    (batch / "dosbox.conf").write_text(config, encoding="ascii")
    started = time.monotonic()
    with (args.output / "dosbox.log").open("wb") as log:
        process = subprocess.run(["dosbox", "-conf", str(batch / "dosbox.conf"), "-exit"],
                                 stdout=log, stderr=subprocess.STDOUT, timeout=150, check=False)
    elapsed = time.monotonic() - started
    done = batch / "DONE.TXT"
    if process.returncode or not done.exists() or done.read_text().strip() != "SUCCESS":
        raise ValueError("DOSBox compilation failed; inspect preserved logs")

    compiled = []
    parser_checks = []
    for case in cases:
        obj_path = batch / (case["stem"].upper() + ".OBJ")
        obj = read_object(obj_path)
        original = raw[header + case["logical"]:header + case["logical"] + case["original_size"]]
        result = {**case, "obj_sha256": obj["sha256"], "compile_status": "success"}
        try:
            code, public = select_function(obj, case["symbol"])
        except UnsupportedOMF as error:
            result.update(selection_status="refused", comparison_status="refused", reason=str(error), exact=False)
        else:
            result.update(selection_status="selected", public=public, raw_compare=compare(original, code))
            try:
                linked, fixups = resolve_ds_offsets(obj, case["symbol"], {"_g_rng": 0x0B5A})
            except UnsupportedOMF as error:
                result.update(relocation_status="refused", comparison_status="refused", reason=str(error), exact=False)
            else:
                result.update(relocation_status="resolved", fixups=fixups,
                              comparison_status="match" if original == linked else "diff",
                              resolved_compare=compare(original, linked), exact=original == linked)
        compiled.append(result)
        if case["stem"] == "base0":
            linked, fixups = resolve_ds_offsets(obj, case["symbol"], {"_g_rng": 0x0B5A})
            if len(fixups) != 2 or any(f["symbol"] != "_g_rng" for f in fixups):
                raise ValueError("Fresh compiler RNG fixups differ from reviewed object")
            try:
                resolve_ds_offsets(obj, case["symbol"], {})
            except UnsupportedOMF:
                parser_checks.append({"case": "missing-external-placement", "rejected": True})
            else:
                raise ValueError("Missing placement was incorrectly accepted")
            altered = copy.deepcopy(obj)
            altered["fixups"][0]["location_type"] = 3
            try:
                resolve_ds_offsets(altered, case["symbol"], {"_g_rng": 0x0B5A})
            except UnsupportedOMF:
                parser_checks.append({"case": "unsupported-far-fixup", "rejected": True})
            else:
                raise ValueError("Unsupported far fixup was incorrectly accepted")
            try:
                select_function(obj, "_not_present")
            except ValueError:
                parser_checks.append({"case": "unknown-PUBDEF", "rejected": True})
            else:
                raise ValueError("Unknown public was incorrectly accepted")
            verified = next(r for r in obj["records"] if r["checksum_status"] == "verified")
            corrupt = bytearray(obj_path.read_bytes())
            corrupt[verified["file_offset"] + 3] ^= 1
            corrupt_path = args.output / "corrupt-checksum.obj"
            corrupt_path.write_bytes(corrupt)
            try:
                read_object(corrupt_path)
            except ValueError:
                parser_checks.append({"case": "corrupt-OMF-checksum", "rejected": True})
            else:
                raise ValueError("Corrupt object checksum was incorrectly accepted")

    asm_source = ROOT / "re/match/sub_e6b9.asm"
    asm_binary = args.output / "sub_e6b9.bin"
    started = time.monotonic()
    subprocess.run(["nasm", "-f", "bin", str(asm_source), "-o", str(asm_binary)], check=True, timeout=20)
    asm_seconds = time.monotonic() - started
    asm = asm_binary.read_bytes()
    if asm != original_rng:
        raise ValueError("Semantic assembly is not byte-identical")
    negative_source = args.output / "rng-negative.asm"
    negative_source.write_text(asm_source.read_text().replace("add ax, 0x9018", "add ax, 0x9019"))
    negative_binary = args.output / "rng-negative.bin"
    subprocess.run(["nasm", "-f", "bin", str(negative_source), "-o", str(negative_binary)],
                   check=True, timeout=20)
    if negative_binary.read_bytes() == original_rng:
        raise ValueError("Changed assembly source was incorrectly accepted")

    scaffold = bytearray(raw)
    scaffold[header + 0xE6B9:header + 0xE6C9] = asm
    scaffold_path = args.output / "RE-DQ3-ASM-SCAFFOLD.EXE"
    scaffold_path.write_bytes(scaffold)
    if digest(scaffold) != ORIGINAL_HASH:
        raise ValueError("Scaffold bytes changed unexpectedly")
    receipt = {
        "schema_version": 1, "issue": "https://github.com/wicanr2/kinginformation-dq3-re/issues/5",
        "input": {"path": "assets_raw/DQ3.EXE", "size": len(raw), "sha256": ORIGINAL_HASH},
        "ida_evidence": {"path": str(args.ida_evidence), "sha256": digest(evidence_raw),
                         "address_space": evidence["address_space"]},
        "compiler_inputs": compiler_inputs,
        "compiler_stdout_sample": (batch / "BASE0.LOG").read_text(errors="replace").splitlines()[:8],
        "tools": {"nasm": subprocess.check_output(["nasm", "-v"], text=True).strip(),
                  "package_manifest_sha256": digest(Path("/opt/matching/package-manifest.txt").read_bytes()),
                  "probe_sha256": digest(Path(__file__).read_bytes()),
                  "omf_parser_sha256": digest((ROOT / "tools/omf_matching_probe.py").read_bytes())},
        "compile_batch": {"cases": len(cases), "wall_seconds": elapsed, "dosbox_returncode": process.returncode,
                          "cycles": 100000, "note": "One boot; no claim of individually measured case timings"},
        "c_candidates": compiled, "c_exact_candidates": sum(c["exact"] for c in compiled),
        "c_completed_comparisons": sum(c["comparison_status"] in ("match", "diff") for c in compiled),
        "c_refused_comparisons": sum(c["comparison_status"] == "refused" for c in compiled),
        "semantic_asm": {"original_name": "sub_1E6B9", "ida_linear": "0x1e6b9",
                         "logical": "0xe6b9", "file": hex(header + 0xE6B9),
                         "source_path": "re/match/sub_e6b9.asm", "source_sha256": digest(asm_source.read_bytes()),
                         "binary_sha256": digest(asm), "bytes": len(asm), "exact": True,
                         "assembly_wall_seconds": asm_seconds},
        "negative_cases": parser_checks + [{"case": "changed-ASM-source", "rejected": True}],
        "scaffold": {"path": str(scaffold_path), "sha256": digest(scaffold), "semantic_asm_bytes": len(asm),
                     "c_recompiled_bytes": 0, "retained_original_bytes": len(raw) - len(asm),
                     "note": "Original-byte scaffold; not a complete source rebuild"},
        "inference_levels": {"ASM_instruction_reproduction": "confirmed", "original_compiler": "hypothesis",
                             "whole_binary_source_recovery": "unknown", "parity_speedup": "unknown"},
    }
    receipt_path = args.output / "receipt.json"
    receipt_path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    shutil.copyfile("/opt/matching/package-manifest.txt", args.output / "package-manifest.txt")
    print(json.dumps({"receipt": str(receipt_path), "c_candidates": len(cases),
                      "c_exact": receipt["c_exact_candidates"], "semantic_asm_exact_bytes": len(asm),
                      "compile_batch_wall_seconds": elapsed, "assembly_wall_seconds": asm_seconds,
                      "scaffold_sha256": digest(scaffold), "uid": receipt_path.stat().st_uid}))


if __name__ == "__main__":
    main()
