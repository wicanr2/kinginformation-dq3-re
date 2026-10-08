"""Docker-only complete CMFDRV source/data reconstruction; docs/25 READY spec.

Build inputs are checked-in ASM, EQU and typed JSON. Original EXE is comparison
only; original objects and machine-code arrays never enter the build.
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
from verify_sdk_instruction_sources import validate_instruction_source


ROOT = Path(__file__).resolve().parents[1]
HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
REGIONS = ((3, 0x905), (0x12d3, 0x1337))


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def render_data(data, heads):
    if data["phase"] != "READY" or data["input_EXE_sha256"] != HASH or data["module"] != "CMFDRV.ASM":
        raise ValueError("Data source is not the reviewed CMF layout")
    covered = set()
    statements = []
    for field in data["fields"]:
        offset = field["offset"]
        if not re.fullmatch(r"[A-Za-z0-9_]+", field["name"]):
            raise ValueError("Unsafe data field name")
        kind = field["kind"]
        if kind == "ascii_z":
            if field["value"] != "FMDRV":
                raise ValueError("Unknown driver signature")
            lines, size = ['DB "FMDRV", 0'], 6
        elif kind == "near_handler_table":
            labels = [entry["instruction_label"] for entry in field["entries"]]
            for label in labels:
                if not re.fullmatch(r"L_[0-9A-F]{4}", label) or int(label[2:], 16) not in heads:
                    raise ValueError("Handler label does not refer to an original instruction")
            lines, size = ["DW " + ", ".join(labels)], 2 * len(labels)
        elif kind in ("u8_array", "u16_array"):
            values = field["values"]
            bound = 255 if kind == "u8_array" else 65535
            if not values or any(type(value) is not int or not 0 <= value <= bound for value in values):
                raise ValueError("Typed data value is out of range")
            directive = "DB" if kind == "u8_array" else "DW"
            lines = [directive + " " + ",".join("0" + format(n, "X") + "h" for n in values[i:i + 16])
                     for i in range(0, len(values), 16)]
            size = len(values) * (1 if kind == "u8_array" else 2)
        else:
            raise ValueError("Unsupported typed data kind")
        positions = set(range(offset, offset + size))
        if covered & positions or not any(lo <= offset and offset + size <= hi for lo, hi in REGIONS):
            raise ValueError("Data field overlaps or crosses an approved non-code region")
        covered |= positions
        statements.append((offset, size, ["; data " + field["name"] + " offset " + hex(offset), *lines]))
    expected = {n for lo, hi in REGIONS for n in range(lo, hi)}
    if covered != expected:
        raise ValueError("Typed data has missing or extra bytes")
    result = {}
    for lo, hi in REGIONS:
        fields = sorted((field for field in statements if lo <= field[0] < hi), key=lambda field: field[0])
        cursor = lo
        lines = []
        for offset, size, content in fields:
            if offset != cursor:
                raise ValueError("Data source has a layout gap")
            lines.extend(content)
            cursor += size
        if cursor != hi:
            raise ValueError("Data region extent differs")
        result[hi] = "\n".join(lines)
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--layout", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    original = ROOT / "assets_raw/DQ3.EXE"
    raw = original.read_bytes()
    if len(raw) != 115282 or sha(original) != HASH:
        raise ValueError("Original identity differs")
    layout = json.loads(args.layout.read_text())
    if (layout["input"]["sha256"] != HASH or layout["tool"]["version"] != "9.4"
            or layout["tool"]["script_sha256"] != sha(ROOT / "tools/ida_matching_module_refs.py")):
        raise ValueError("Original layout identity differs")
    module = next(m for m in layout["modules"] if m["SDK_metadata_name"] == "CMFDRV.ASM")
    heads = {int(row["ida_linear"], 16) - 0x23920 for row in module["instructions"]}
    code_path, constant_path = ROOT / "re/match/cmfdrv_code.asm", ROOT / "re/match/cmfdrv_constants.asm"
    data_path = ROOT / "re/match/cmfdrv_data.json"
    code = code_path.read_text(encoding="ascii")
    validate_instruction_source(code)
    validate_instruction_source(constant_path.read_text())
    rendered = render_data(json.loads(data_path.read_text()), heads)
    for end, text in rendered.items():
        marker = "ORG 0" + format(end, "X") + "h"
        if code.count(marker) != 1:
            raise ValueError("Original code-source data marker differs")
        code = code.replace(marker, text)
    code = code.replace("; Verified instruction source only. Original SDK data is not reconstructed.",
                        "; Full CMF byte-layout source composed from reviewed ASM/EQU and typed data.")
    payload_path = Path("/opt/watcom/source-manifest.json")
    payload = json.loads(payload_path.read_text())
    if payload["revision"] != "2.0-20261001-r2":
        raise ValueError("Assembler revision differs")
    for record in payload["files"]:
        if sha(Path("/opt/watcom") / record["path"]) != record["sha256"]:
            raise ValueError("Assembler/linker payload differs")
    args.output.mkdir(exist_ok=False)
    build = Path("/tmp/cmf-source-module")
    build.mkdir(exist_ok=False)
    generated = build / "cmfdrv_module.asm"
    generated.write_text(code, encoding="ascii")
    shutil.copyfile(generated, args.output / generated.name)
    constants = build / constant_path.name
    shutil.copyfile(constant_path, constants)
    objects = []
    for path in (generated, constants):
        obj = path.with_suffix(".obj")
        command = ["wasm", "-0", "-zcm=masm", "-zld", "-fo=" + obj.name, path.name]
        result = subprocess.run(command, cwd=build, capture_output=True, timeout=30)
        (args.output / (path.stem + ".log")).write_bytes(result.stdout + result.stderr)
        if result.returncode:
            raise ValueError("Source module assembly failed")
        shutil.copyfile(obj, args.output / obj.name)
        objects.append(obj)
    obj = read_object(objects[0])
    segment = obj["segments"][1]
    if segment["length"] != 5296 or not all(segment["written"]):
        raise ValueError("Full source module was not completely materialized")
    exe = build / "cmfdrv.exe"
    command = ["wlink", "format", "dos", "option", "nodefaultlibs", "option", "start=module_entry",
               "name", exe.name, "file", objects[0].name, "file", objects[1].name]
    result = subprocess.run(command, cwd=build, capture_output=True, timeout=30)
    (args.output / "link.log").write_bytes(result.stdout + result.stderr)
    if result.returncode:
        raise ValueError("Source module link failed")
    linked = exe.read_bytes()
    header = struct.unpack_from("<H", linked, 8)[0] * 16
    ip, cs = struct.unpack_from("<HH", linked, 20)
    start = header + ip + cs * 16
    if linked[start:start + 5296] != raw[0x14c90:0x16140]:
        raise ValueError("Full source/data module differs from original")
    shutil.copyfile(exe, args.output / exe.name)
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__)), "code_sha256": sha(code_path),
               "constants_sha256": sha(constant_path), "data_sha256": sha(data_path), "generated_ASM_sha256": sha(generated),
               "input_EXE_sha256": HASH, "IDA_layout_sha256": sha(args.layout),
               "object_sha256": [sha(path) for path in objects], "generated_MZ_sha256": sha(exe),
               "module_source_bytes_exact": 5296, "instruction_bytes": 2890, "data_bytes": 2406,
               "built_from_original_objects": False, "whole_module_byte_layout_complete": True,
               "original_data_semantics_complete": False, "runtime_hardware_parity": False, "whole_goal_complete": False}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"module_source_bytes_exact": 5296, "data_bytes": 2406, "whole_goal_complete": False}))


if __name__ == "__main__":
    main()
