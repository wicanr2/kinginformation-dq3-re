"""Docker-only complete identified SDK source/data reconstruction; docs/25 specs.

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
PROFILES = {
    "CMFDRV.ASM": {"stem": "cmfdrv", "linear": 0x23920, "file": 0x14c90, "size": 5296,
                   "regions": ((3, 0x905), (0x12d3, 0x1337)), "code_bytes": 2890, "data_bytes": 2406},
    "CTVMEM.ASM": {"stem": "ctvmem", "linear": 0x22f60, "file": 0x142d0, "size": 2493,
                   "regions": ((3, 0xe3), (0x749, 0x754)), "code_bytes": 2258, "data_bytes": 235},
}


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def render_data(data, heads):
    if data["phase"] != "READY" or data["input_EXE_sha256"] != HASH or data["module"] not in PROFILES:
        raise ValueError("Data source is not a reviewed SDK layout")
    profile = PROFILES[data["module"]]
    regions = profile["regions"]
    covered = set()
    statements = []
    for field in data["fields"]:
        offset = field["offset"]
        if not re.fullmatch(r"[A-Za-z0-9_]+", field["name"]):
            raise ValueError("Unsafe data field name")
        kind = field["kind"]
        if kind == "ascii_z":
            expected = {"driver_signature": "FMDRV"} if data["module"] == "CMFDRV.ASM" else {
                "driver_signature": "CT-VOICE", "device_description": "Creative Sound Blaster Card"}
            if field["name"] not in expected or field["value"] != expected[field["name"]]:
                raise ValueError("Unknown driver signature")
            lines, size = ['DB "' + field["value"] + '", 0'], len(field["value"]) + 1
        elif kind == "ascii_fixed":
            if (data["module"] != "CTVMEM.ASM" or field["name"] != "copyright_header_text"
                    or field["offset"] != 0x39 or not re.fullmatch(r"[ -~]{76}", field["value"])
                    or any(c in field["value"] for c in ('"', "\\"))):
                raise ValueError("Unsafe or unknown fixed header string")
            lines, size = ['DB "' + field["value"] + '"'], len(field["value"])
        elif kind == "near_handler_table":
            entries = []
            for index, entry in enumerate(field["entries"]):
                if data["module"] == "CTVMEM.ASM" and field["offset"] == 0x85 and index == 6:
                    if entry.get("initial_u16") != 0x6b06 or "instruction_label" in entry:
                        raise ValueError("Unexpected mutable handler seed")
                    entries.append("06B06h")
                elif "initial_u16" in entry:
                    raise ValueError("Literal seed outside the reviewed mutable slot")
                else:
                    label = entry["instruction_label"]
                    if not re.fullmatch(r"L_[0-9A-F]{4}", label) or int(label[2:], 16) not in heads:
                        raise ValueError("Handler label does not refer to an original instruction")
                    entries.append(label)
            lines, size = ["DW " + ", ".join(entries)], 2 * len(entries)
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
        if covered & positions or not any(lo <= offset and offset + size <= hi for lo, hi in regions):
            raise ValueError("Data field overlaps or crosses an approved non-code region")
        covered |= positions
        statements.append((offset, size, ["; data " + field["name"] + " offset " + hex(offset), *lines]))
    expected = {n for lo, hi in regions for n in range(lo, hi)}
    if covered != expected:
        raise ValueError("Typed data has missing or extra bytes")
    result = {}
    for lo, hi in regions:
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
    parser.add_argument("--module", choices=tuple(PROFILES), default="CMFDRV.ASM")
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
    profile = PROFILES[args.module]
    module = next(m for m in layout["modules"] if m["SDK_metadata_name"] == args.module)
    heads = {int(row["ida_linear"], 16) - profile["linear"] for row in module["instructions"]}
    code_path = ROOT / ("re/match/" + profile["stem"] + "_code.asm")
    constant_path = ROOT / ("re/match/" + profile["stem"] + "_constants.asm")
    data_path = ROOT / ("re/match/" + profile["stem"] + "_data.json")
    code = code_path.read_text(encoding="ascii")
    validate_instruction_source(code)
    validate_instruction_source(constant_path.read_text())
    data = json.loads(data_path.read_text())
    if data["module"] != args.module:
        raise ValueError("Selected module/data source differs")
    rendered = render_data(data, heads)
    for end, text in rendered.items():
        marker = "ORG 0" + format(end, "X") + "h"
        if code.count(marker) != 1:
            raise ValueError("Original code-source data marker differs")
        code = code.replace(marker, text)
    code = code.replace("; Verified instruction source only. Original SDK data is not reconstructed.",
                        "; Full SDK byte-layout source composed from reviewed ASM/EQU and typed data.")
    payload_path = Path("/opt/watcom/source-manifest.json")
    payload = json.loads(payload_path.read_text())
    if payload["revision"] != "2.0-20261001-r2":
        raise ValueError("Assembler revision differs")
    for record in payload["files"]:
        if sha(Path("/opt/watcom") / record["path"]) != record["sha256"]:
            raise ValueError("Assembler/linker payload differs")
    args.output.mkdir(exist_ok=False)
    build = Path("/tmp/sdk-source-module")
    build.mkdir(exist_ok=False)
    generated = build / (profile["stem"] + "_module.asm")
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
    if segment["length"] != profile["size"] or not all(segment["written"]):
        raise ValueError("Full source module was not completely materialized")
    exe = build / (profile["stem"] + ".exe")
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
    if linked[start:start + profile["size"]] != raw[profile["file"]:profile["file"] + profile["size"]]:
        raise ValueError("Full source/data module differs from original")
    shutil.copyfile(exe, args.output / exe.name)
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__)), "code_sha256": sha(code_path),
               "constants_sha256": sha(constant_path), "data_sha256": sha(data_path), "generated_ASM_sha256": sha(generated),
               "input_EXE_sha256": HASH, "IDA_layout_sha256": sha(args.layout),
               "object_sha256": [sha(path) for path in objects], "generated_MZ_sha256": sha(exe),
               "SDK_metadata_name": args.module, "module_source_bytes_exact": profile["size"],
               "instruction_bytes": profile["code_bytes"], "data_bytes": profile["data_bytes"],
               "built_from_original_objects": False, "whole_module_byte_layout_complete": True,
               "original_data_semantics_complete": False, "runtime_hardware_parity": False, "whole_goal_complete": False}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"module_source_bytes_exact": profile["size"], "data_bytes": profile["data_bytes"], "whole_goal_complete": False}))


if __name__ == "__main__":
    main()
