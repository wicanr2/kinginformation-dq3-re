"""Docker-only producer-scoped OMF signed-addend control; indexed in docs/25.

Link the actual C object against synthetic data with the vendor WLINK. The
fixture contains no original DQ3 bytes. This establishes only the WCC producer
contract, never a universal OMF sign rule or a complete EXE layout.
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

from omf_matching_probe import UnsupportedOMF, read_object, resolve_ds_offsets


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def record(kind, payload):
    raw = bytes([kind]) + struct.pack("<H", len(payload) + 1) + payload
    return raw + bytes([-sum(raw) & 255])


def name(value):
    raw = value.encode("ascii")
    return bytes([len(raw)]) + raw


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--candidate-obj", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Run as host UID/GID")
    args.output.mkdir(exist_ok=False)
    symbol, function = "_unknown_DS_265d", "_sub_32a3"
    obj = read_object(args.candidate_obj)
    placed, fixups = resolve_ds_offsets(obj, function, {symbol: 0x265d}, signed_addends=True)
    if len(fixups) != 1 or fixups[0]["original_addend"] != 0xfffe or fixups[0]["arithmetic_addend"] != -2:
        raise ValueError("Fixture requires the actual WCC symbol-minus-two operand")
    negatives = []
    for label, offsets, signed, changed in (
        ("unsigned-overflow", {symbol: 0x265d}, False, obj),
        ("missing-placement", {}, True, obj),
        ("signed-underflow", {symbol: 1}, True, obj),
        ("negative-placement", {symbol: -1}, True, obj),
        ("invalid-placement-with-in-range-result", {symbol: 65537}, True, obj),
    ):
        try:
            resolve_ds_offsets(changed, function, offsets, signed_addends=signed)
        except (ValueError, UnsupportedOMF):
            negatives.append({"case": label, "rejected": True})
        else:
            raise ValueError("Invalid relocation accepted: " + label)
    changed = copy.deepcopy(obj)
    changed["fixups"][0]["frame"]["method"] = 0
    try:
        resolve_ds_offsets(changed, function, {symbol: 0x265d}, signed_addends=True)
    except (ValueError, UnsupportedOMF):
        negatives.append({"case": "unsupported-frame", "rejected": True})
    else:
        raise ValueError("Unsupported frame accepted")
    # TIS OMF records: THEADR, LNAMES, SEGDEF, GRPDEF, PUBDEF, LEDATA, MODEND.
    # The linker is allowed to insert a frame bias; read the actual public's
    # offset from its map rather than assuming the synthetic PUBDEF is DS.
    payload = b"".join([
        record(0x80, name("synthetic-data")),
        record(0x96, name("") + name("_DATA") + name("DATA") + name("DGROUP")),
        record(0x98, bytes([0x48]) + struct.pack("<H", 0x269d) + bytes([2, 3, 1])),
        record(0x9a, bytes([4, 0xff, 1])),
        record(0x90, bytes([1, 1]) + name(symbol) + struct.pack("<H", 0x265d) + bytes([0]) +
               name("_small_code_") + struct.pack("<H", 0) + bytes([0])),
        record(0xa0, bytes([1]) + struct.pack("<H", 0x265d) + bytes(64)),
        record(0x8a, bytes([0])),
    ])
    data = args.output / "data.obj"
    data.write_bytes(payload)
    exe, map_path = args.output / "linked.exe", args.output / "linked.map"
    command = ["wlink", "format", "dos", "option", "nodefaultlibs", "option", "map=" + str(map_path),
               "option", "start=" + function, "name", str(exe), "file", str(args.candidate_obj), "file", str(data)]
    result = subprocess.run(command, capture_output=True, text=True, timeout=30)
    log = result.stdout + result.stderr
    (args.output / "link.log").write_text(log)
    if result.returncode:
        raise ValueError("Vendor WLINK failed: " + log)
    map_text = map_path.read_text()
    match = re.search(r"^\s*([0-9a-fA-F]+):([0-9a-fA-F]+)\s+" + re.escape(symbol) + r"\s*$", map_text, re.M)
    if not match:
        raise ValueError("Vendor map lacks the actual frame-relative public")
    actual_offset = int(match[2], 16)
    linked, actual_fixes = resolve_ds_offsets(obj, function, {symbol: actual_offset}, signed_addends=True)
    raw = exe.read_bytes()
    if raw[:2] != b"MZ":
        raise ValueError("Vendor output is not DOS MZ")
    header = int.from_bytes(raw[8:10], "little") * 16
    ip, cs = struct.unpack_from("<HH", raw, 20)
    entry = header + cs * 16 + ip
    if raw[entry:entry + len(linked)] != linked:
        raise ValueError("Vendor linked code differs from producer-scoped relocation")
    receipt = {
        "schema_version": 1, "producer_sha256": sha(Path(__file__)),
        "OMF_parser_sha256": sha(Path(__file__).with_name("omf_matching_probe.py")),
        "candidate_object_sha256": sha(args.candidate_obj), "synthetic_data_sha256": sha(data),
        "linker": {"path": "/opt/watcom/binl64/wlink", "sha256": sha(Path("/opt/watcom/binl64/wlink")),
                   "version_banner": log.splitlines()[:4]},
        "vendor_map_sha256": sha(map_path), "vendor_MZ_sha256": sha(exe),
        "original_pubdef_offset": "0x265d", "actual_vendor_frame_offset": hex(actual_offset),
        "original_encoded_addend": "0xfffe", "arithmetic_addend": -2,
        "vendor_operand": hex(actual_fixes[0]["resolved_operand"]),
        "vendor_code_hex": linked.hex(), "explicit_DQ3_placement_code_hex": placed.hex(),
        "negative_cases": negatives, "vendor_link_equal": True,
        "scope": "Synthetic WCC/WLINK F5/T2 control only; no original EXE layout or universal OMF rule",
    }
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"vendor_link_equal": True, "negative_cases_rejected": len(negatives),
                      "actual_vendor_frame_offset": hex(actual_offset), "vendor_operand": receipt["vendor_operand"]}))


if __name__ == "__main__":
    main()
