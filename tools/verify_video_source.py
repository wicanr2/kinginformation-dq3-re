"""Docker-only low-level video source region; entry: docs/25 READY spec.

Only repository semantic ASM and one typed alignment byte enter the build.
Original executable and fresh IDA layout are verification inputs.
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


HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def render_source(text, data):
    validate_instruction_source(text)
    if data["schema_version"] != 1 or data["phase"] != "READY" or data["input_sha256"] != HASH:
        raise ValueError("Unreviewed video source layout")
    if data["ida_linear_start"] != "0x209ce" or data["file_start"] != "0x11d3e" or data["region_bytes"] != 402:
        raise ValueError("Original region geometry differs")
    alignment = data["alignment"]
    if alignment["kind"] != "u8" or alignment["offset"] != 401 or type(alignment["value"]) is not int or alignment["value"] != 0:
        raise ValueError("Alignment may only preserve the reviewed original u8 at region+401")
    if len(re.findall(r"^VIDEO_CODE ENDS$", text, re.M)) != 1:
        raise ValueError("Code source requires one explicit region boundary")
    return text.replace("VIDEO_CODE ENDS", "original_alignment_20B5F:\n    DB 0\nVIDEO_CODE ENDS")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--layout", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    root = Path(__file__).resolve().parents[1]
    code_path = root / "re/match/video_bios_code.asm"
    data_path = root / "re/match/video_bios_data.json"
    data = json.loads(data_path.read_text())
    generated = render_source(code_path.read_text(), data)
    raw_path = root / "assets_raw/DQ3.EXE"
    raw = raw_path.read_bytes()
    if len(raw) != 115282 or sha(raw_path) != HASH:
        raise ValueError("Original comparison input differs")
    layout = json.loads(args.layout.read_text())
    if layout["input"]["sha256"] != HASH or layout["tool"]["version"] != "9.4" or layout["tool"]["script_sha256"] != sha(root / "tools/ida_matching_video.py"):
        raise ValueError("Fresh IDA layout identity differs")
    if layout["region"]["start"] != "0x209ce" or layout["region"]["end_exclusive"] != "0x20b60" or layout["MZ_relocations"]:
        raise ValueError("Region/frame relocation contract differs")
    covered = set()
    for row in layout["instructions"]:
        offset = int(row["ida_linear"], 16) - 0x209CE
        instruction = bytes.fromhex(row["file_bytes"])
        if raw[0x11D3E + offset:0x11D3E + offset + len(instruction)] != instruction:
            raise ValueError("IDA instruction is not the original input")
        covered.update(range(offset, offset + len(instruction)))
    if covered != set(range(401)) or len(layout["instructions"]) != 209:
        raise ValueError("Reviewed instruction extent differs")
    gaps = layout["uncovered_bytes"]
    if len(gaps) != 1 or gaps[0]["ida_linear"] != "0x20b5f" or not gaps[0]["is_data"] or not gaps[0]["is_align"] or gaps[0]["raw_u8"] != 0:
        raise ValueError("Typed alignment lacks reviewed non-code identity")
    payload_path = Path("/opt/watcom/source-manifest.json")
    payload = json.loads(payload_path.read_text())
    if payload["revision"] != "2.0-20261001-r2" or not payload["complete_archive_verified"]:
        raise ValueError("Unreviewed assembler payload")
    for item in payload["files"]:
        if sha(Path("/opt/watcom") / item["path"]) != item["sha256"]:
            raise ValueError("Locked tool payload differs")
    args.output.mkdir(exist_ok=False)
    work = Path("/tmp/video-source-build")
    work.mkdir(exist_ok=False)
    source = work / "video.asm"
    source.write_text(generated, encoding="ascii")
    for command in [
        ["wasm", "-0", "-zcm=masm", "-zld", "-fo=video.obj", "video.asm"],
        ["wlink", "format", "dos", "option", "nodefaultlibs", "option", "nofarcalls",
         "option", "start=region_entry", "name", "video.exe", "file", "video.obj"],
    ]:
        result = subprocess.run(command, cwd=work, capture_output=True, timeout=30)
        if result.returncode:
            raise ValueError(result.stdout.decode(errors="replace") + result.stderr.decode(errors="replace"))
    obj = read_object(work / "video.obj")
    segments = [s for s in obj["segments"][1:] if s["class"] == "CODE"]
    if len(segments) != 1 or segments[0]["length"] != 402 or not all(segments[0]["written"]):
        raise ValueError("Source does not materialize the full region")
    if obj["fixups"]:
        raise ValueError("Unexpected source relocation")
    linked = (work / "video.exe").read_bytes()
    if linked[:2] != b"MZ":
        raise ValueError("Vendor output is not MZ")
    header = struct.unpack_from("<H", linked, 8)[0] * 16
    ip, cs = struct.unpack_from("<HH", linked, 20)
    start = header + cs * 16 + ip
    if linked[start:start + 402] != raw[0x11D3E:0x11ED0]:
        raise ValueError("Whole source region differs from original")
    for path in work.iterdir():
        shutil.copyfile(path, args.output / path.name)
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__)),
               "code_source_sha256": sha(code_path), "data_source_sha256": sha(data_path),
               "generated_ASM_sha256": sha(source), "object_sha256": sha(work / "video.obj"),
               "MZ_sha256": sha(work / "video.exe"), "IDA_layout_sha256": sha(args.layout),
               "input_EXE_sha256": HASH, "region_source_bytes_exact": 402,
               "instruction_bytes": 401, "alignment_bytes": 1,
               "original_object_identity": "unknown", "low_level_ownership": "confirmed display hardware services",
               "runtime_hardware_parity": False, "whole_goal_complete": False,
               "original_code_used_in_build": False}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"region_source_bytes_exact": 402, "instruction_bytes": 401,
                      "alignment_bytes": 1, "whole_goal_complete": False}))


if __name__ == "__main__":
    main()
