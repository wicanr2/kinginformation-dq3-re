"""Docker-only source rebuild of identified SDK instruction portions; docs/25.

Original EXE supplies comparison bytes only. No original objects are build
inputs; no code arrays, db, dw, dd or incbin are allowed in these sources.
ORG holes deliberately remain unknown and are excluded from source coverage.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import struct
import subprocess

from omf_matching_probe import read_object


ROOT = Path(__file__).resolve().parents[1]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def validate_instruction_source(text):
    """Reject data emission and source imports even behind a label or TIMES."""
    forbidden = r"\b(db|dw|dd|dq|dt|dup|incbin|include|macro|endm|\.incbin)\b"
    for line in text.splitlines():
        statement = line.split(";", 1)[0]
        if re.search(forbidden, statement, re.I):
            raise ValueError("Source contains raw byte/data import")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--layout", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    manifest_path = ROOT / "tools/sdk_instruction_source_manifest.json"
    manifest = json.loads(manifest_path.read_text())
    input_path = ROOT / manifest["input_path"]
    raw = input_path.read_bytes()
    if len(raw) != manifest["input_size"] or sha(input_path) != manifest["input_sha256"]:
        raise ValueError("Original EXE differs")
    layout = json.loads(args.layout.read_text())
    if (layout["input"]["sha256"] != manifest["input_sha256"] or layout["tool"]["version"] != "9.4"
            or layout["tool"]["script_sha256"] != sha(ROOT / "tools/ida_matching_module_refs.py")):
        raise ValueError("Fresh IDA layout identity differs")
    payload_path = Path("/opt/watcom/source-manifest.json")
    payload = json.loads(payload_path.read_text())
    if payload["revision"] != "2.0-20261001-r2" or not payload["complete_archive_verified"]:
        raise ValueError("Wasm payload revision differs")
    for item in payload["files"]:
        if sha(Path("/opt/watcom") / item["path"]) != item["sha256"]:
            raise ValueError("Assembler/linker payload differs")
    args.output.mkdir(exist_ok=False)
    compile_root = Path("/tmp/sdk-instruction-source")
    compile_root.mkdir(exist_ok=False)
    results = []
    for case in manifest["cases"]:
        module = next(m for m in layout["modules"] if m["SDK_metadata_name"] == case["SDK_metadata_name"])
        base = int(case["module_IDA_linear_start"], 16)
        heads = {int(r["ida_linear"], 16): r for r in module["instructions"]}
        approved = set()
        for lo, hi in case["instruction_ranges_IDA_linear"]:
            cursor = int(lo, 16)
            while cursor < int(hi, 16):
                row = heads[cursor]
                size = len(bytes.fromhex(row["file_bytes"]))
                if cursor + size > int(hi, 16):
                    raise ValueError("Instruction crosses source-range extent")
                approved.update(range(cursor - base, cursor - base + size))
                cursor += size
        if len(approved) != case["verified_instruction_bytes"]:
            raise ValueError("Declared instruction count differs")
        sources = []
        objects = []
        for field in ("source", "constants_source"):
            source = ROOT / case[field]
            text = source.read_text(encoding="ascii")
            validate_instruction_source(text)
            shutil.copyfile(source, compile_root / source.name)
            obj_name = source.with_suffix(".obj").name
            command = ["wasm", "-0", "-zcm=masm", "-zld", "-fo=" + obj_name, source.name]
            result = subprocess.run(command, cwd=compile_root, capture_output=True, timeout=30)
            (args.output / (source.stem + ".log")).write_bytes(result.stdout + result.stderr)
            if result.returncode:
                raise ValueError("Semantic source did not assemble")
            compiled = compile_root / obj_name
            shutil.copyfile(compiled, args.output / obj_name)
            sources.append({"source": case[field], "source_sha256": sha(source), "object_sha256": sha(compiled), "command": command})
            objects.append(compiled)
        obj = read_object(objects[0])
        public = next(p for p in obj["publics"] if p["name"] == "module_entry")
        segment = obj["segments"][public["segment_index"]]
        actual_written = {n for n, written in enumerate(segment["written"]) if written}
        if actual_written != approved:
            raise ValueError("Assembler emitted bytes outside the verified instruction ranges")
        executable = compile_root / (objects[0].stem + ".exe")
        command = ["wlink", "format", "dos", "option", "nodefaultlibs", "option", "start=module_entry",
                   "name", executable.name, "file", objects[0].name, "file", objects[1].name]
        result = subprocess.run(command, cwd=compile_root, capture_output=True, timeout=30)
        (args.output / (objects[0].stem + "-link.log")).write_bytes(result.stdout + result.stderr)
        if result.returncode:
            raise ValueError("Semantic source link failed")
        linked = executable.read_bytes()
        if linked[:2] != b"MZ":
            raise ValueError("Output is not DOS MZ")
        header = struct.unpack_from("<H", linked, 8)[0] * 16
        ip, cs = struct.unpack_from("<HH", linked, 20)
        start = header + cs * 16 + ip
        if start + case["module_size"] > len(linked):
            raise ValueError("Linked module extent is incomplete")
        original_start = base - 0x10000 + 0x1370
        for offset in approved:
            if linked[start + offset] != raw[original_start + offset]:
                raise ValueError("Instruction source differs at " + hex(base + offset))
        shutil.copyfile(executable, args.output / executable.name)
        results.append({"SDK_metadata_name": case["SDK_metadata_name"], "sources": sources,
                        "generated_MZ_sha256": sha(executable), "semantic_instruction_bytes_exact": len(approved),
                        "unknown_data_bytes_excluded": case["unknown_data_bytes_excluded"],
                        "whole_module_source_complete": False, "built_from_original_objects": False,
                        "actual_fixups": obj["fixups"]})
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__)), "manifest_sha256": sha(manifest_path),
               "input_EXE_sha256": manifest["input_sha256"], "IDA_layout_sha256": sha(args.layout),
               "compiler_payload_manifest_sha256": sha(payload_path), "results": results,
               "instruction_source_bytes_exact": sum(r["semantic_instruction_bytes_exact"] for r in results),
               "whole_modules_complete": False, "whole_goal_complete": False,
               "scope": "Source-rebuilt instruction portions only; ORG gaps and original data excluded"}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"instruction_source_bytes_exact": receipt["instruction_source_bytes_exact"], "whole_modules_complete": False}))


if __name__ == "__main__":
    main()
