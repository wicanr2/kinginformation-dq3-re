"""Docker-only vendor re-link controls for original SDK modules; docs/25.

Consumes original OMF objects only as identity evidence. This is expressly
not a source rebuild and contributes zero bytes to C/ASM source coverage.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import struct
import subprocess


ROOT = Path(__file__).resolve().parents[1]
EXE_HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
LIB_HASH = "01b242cb99193d006e23b73babe122887e59713410f88798283a9120df1d0683"
CONTROLS = (("CTVMEM.ASM", "CTVM_VOICE_DRV", 0x22f60), ("CMFDRV.ASM", "SBFM_CMF_DRV", 0x23920))


def sha(raw):
    return hashlib.sha256(raw).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--module-inventory", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    raw = (ROOT / "assets_raw/DQ3.EXE").read_bytes()
    library = (ROOT / "assets_raw/SBCM.LIB").read_bytes()
    if len(raw) != 115282 or sha(raw) != EXE_HASH or len(library) != 35840 or sha(library) != LIB_HASH:
        raise ValueError("Original inputs differ")
    inventory = json.loads(args.module_inventory.read_text())
    if inventory["input"]["sha256"] != LIB_HASH or inventory["original_EXE_sha256"] != EXE_HASH:
        raise ValueError("SDK module inventory identity differs")
    if inventory["producer_sha256"] != sha((ROOT / "tools/probe_sbcm_modules.py").read_bytes()):
        raise ValueError("SDK inventory producer freshness differs")
    args.output.mkdir(exist_ok=False)
    results = []
    linker_path = Path("/opt/watcom/binl64/wlink")
    for name, public, linear in CONTROLS:
        module = next(row for row in inventory["modules"] if row["original_module_name"] == name)
        path = args.module_inventory.parent / module["artifact_name"]
        obj = path.read_bytes()
        lib_offset = int(module["library_file_start"], 16)
        if sha(obj) != module["sha256"] or obj != library[lib_offset:lib_offset + module["size"]]:
            raise ValueError("Original OMF object differs from pinned SDK library")
        segment = module["segments"]
        if len(segment) != 1 or segment[0]["class"] != "CODE":
            raise ValueError("Control is not one original CODE segment")
        stem = path.stem
        exe_path, map_path = args.output / (stem + ".exe"), args.output / (stem + ".map")
        command = [str(linker_path), "format", "dos", "option", "nodefaultlibs", "option", "start=" + public,
                   "option", "map=" + str(map_path), "name", str(exe_path), "file", str(path)]
        process = subprocess.run(command, capture_output=True, text=True, timeout=20)
        (args.output / (stem + ".log")).write_text(process.stdout + process.stderr)
        if process.returncode or not exe_path.is_file() or not map_path.is_file():
            raise ValueError("Vendor linker did not produce a complete control")
        linked = exe_path.read_bytes()
        if linked[:2] != b"MZ":
            raise ValueError("Vendor output is not DOS MZ")
        header = struct.unpack_from("<H", linked, 8)[0] * 16
        ip, cs = struct.unpack_from("<HH", linked, 20)
        entry = header + cs * 16 + ip
        size = segment[0]["size"]
        code = linked[entry:entry + size]
        original_offset = linear - 0x10000 + 0x1370
        original = raw[original_offset:original_offset + size]
        if len(code) != size or code != original:
            raise ValueError("Complete SDK linked segment differs from original EXE: " + name)
        results.append({"original_module_name": name, "original_module_sha256": module["sha256"],
                        "original_library_file_start": module["library_file_start"], "original_library_module_size": module["size"],
                        "vendor_command": command, "vendor_MZ_sha256": sha(linked), "vendor_map_sha256": sha(map_path.read_bytes()),
                        "original_EXE_IDA_linear_start": hex(linear), "original_EXE_logical_start": hex(linear - 0x10000),
                        "original_EXE_file_start": hex(original_offset), "size": size, "whole_segment_equal": True,
                        "linked_segment_sha256": sha(code), "object_unwritten_bytes": size - segment[0]["written_bytes"],
                        "scope": "Original-object vendor re-link only; no readable-source or source coverage"})
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__).read_bytes()),
               "input_EXE_sha256": EXE_HASH, "input_library_sha256": LIB_HASH,
               "module_inventory_sha256": sha(args.module_inventory.read_bytes()),
               "vendor_linker_sha256": sha(linker_path.read_bytes()), "results": results, "source_coverage_increment": 0}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"whole_original_module_bytes_identified": sum(r["size"] for r in results), "modules": len(results), "source_coverage_increment": 0}))


if __name__ == "__main__":
    main()
