"""Docker-only primary C codegen experiment; docs/25 is the research index.

Preserve every authored C variant, compiler object and actual relocation.
No instruction injection, byte patching or original-code scaffolds. A byte
match is a candidate observation and does not approve source-unit ownership.
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


ROOT = Path(__file__).resolve().parents[1]
EXPECTED = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


FILL_PREFIX = """extern volatile unsigned char unknown_DS_265d[];
unsigned sub_32a3(void);
#pragma aux sub_32a3 "_*" value [cx] modify exact [bx cx];
unsigned sub_32a3(void) {
"""
SEARCH_PREFIX = """unsigned char sub_6fcf(unsigned incoming_ax, const unsigned char __near *source);
#pragma aux sub_6fcf "_*" parm [ax] [si] value [bl] modify exact [ax bx cx si];
unsigned char sub_6fcf(unsigned incoming_ax, const unsigned char __near *source) {
"""


def variants(unit):
    result = {"baseline": (ROOT / unit["source"]).read_text(encoding="ascii")}
    if unit["id"] == "sub_132A3":
        declarations = "unsigned count = 13; unsigned index = 8;\n"
        bodies = {
            "count_first": declarations + "do { unknown_DS_265d[index] = 255; index += 2; } while (--count); return count;",
            "register": "register " + declarations.replace("; unsigned index", "; register unsigned index") +
                        "do { unknown_DS_265d[index] = 255; index += 2; } while (--count); return count;",
            "for_count": "unsigned index = 8, count; for (count = 13; count; --count) { unknown_DS_265d[index] = 255; index += 2; } return count;",
            "goto_count": declarations + "again: unknown_DS_265d[index] = 255; index += 2; if (--count) goto again; return count;",
            "two_increments": declarations + "do { unknown_DS_265d[index] = 255; ++index; ++index; } while (--count); return count;",
            "count_index": "unsigned count; for (count = 13; count; --count) unknown_DS_265d[34 - 2 * count] = 255; return count;",
        }
        result.update({name: FILL_PREFIX + body + "\n}\n" for name, body in bodies.items()})
        result["nonvolatile"] = result["count_first"].replace("extern volatile", "extern")
    elif unit["id"] == "sub_16FCF":
        declarations = "unsigned char target = (unsigned char)(incoming_ax >> 8), index = 1; unsigned count = 6;\n"
        bodies = {
            "for_count": declarations + "for (; count; --count) { if (*source++ == target) break; ++index; } return index;",
            "goto_count": declarations + "again: if (*source++ == target) return index; ++index; if (--count) goto again; return index;",
            "register": "register " + declarations.replace("; unsigned count", "; register unsigned count") +
                        "do { if (*source++ == target) return index; ++index; } while (--count); return index;",
            "word_compare": "unsigned target = incoming_ax & 0xff00, count = 6; unsigned char index = 1; do { if (((unsigned)*source++ << 8) == target) return index; ++index; } while (--count); return index;",
            "predecrement": declarations + "do { if (*source++ == target) break; ++index; --count; } while (count); return index;",
        }
        result.update({name: SEARCH_PREFIX + body + "\n}\n" for name, body in bodies.items()})
    return result


FLAG_SETS = {
    "previous": ["-os", "-oi", "-ofr"],
    "no_reorder": ["-os", "-oi", "-of"],
    "loop": ["-os", "-oi", "-of", "-ol"],
    "loop_flow": ["-os", "-oi", "-of", "-ol", "-ok"],
    "loop_reorder": ["-os", "-oi", "-ofr", "-ol"],
    "expensive_loop": ["-os", "-oi", "-of", "-ol", "-oh"],
    "unroll": ["-os", "-oi", "-of", "-ol+"],
    "time_loop": ["-ot", "-oi", "-of", "-ol"],
    "disabled": ["-od", "-oi"],
}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--ida-inventory", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    raw = (ROOT / "assets_raw/DQ3.EXE").read_bytes()
    if len(raw) != 115282 or hashlib.sha256(raw).hexdigest() != EXPECTED:
        raise ValueError("Original input differs")
    inventory = json.loads(args.ida_inventory.read_text())
    if inventory["input"]["sha256"] != EXPECTED or inventory["tool"]["version"] != "9.4":
        raise ValueError("Original IDA inventory identity differs")
    manifest_path = ROOT / "tools/watcom_matching_manifest.json"
    manifest = json.loads(manifest_path.read_text())
    payload_path = Path("/opt/watcom/source-manifest.json")
    payload = json.loads(payload_path.read_text())
    if not payload["complete_archive_verified"] or payload["release"] != "2026-10-01-Build":
        raise ValueError("Compiler identity differs")
    for item in payload["files"]:
        if sha(Path("/opt/watcom") / item["path"]) != item["sha256"]:
            raise ValueError("Compiler/header payload differs")
    args.output.mkdir(exist_ok=False)
    compile_root = Path("/tmp/watcom-primary-codegen")
    compile_root.mkdir(exist_ok=False)
    help_result = subprocess.run(["wcc"], capture_output=True, text=True, timeout=20)
    (args.output / "compiler-help.txt").write_text(help_result.stdout + help_result.stderr)
    results = []
    started = time.monotonic()
    for unit in manifest["cases"]:
        if unit["id"] not in ("sub_132A3", "sub_16FCF"):
            continue
        address = int(unit["ida_linear_start"], 16)
        fn = next(fn for fn in inventory["functions"] if int(fn["ida_linear_start"], 16) == address)
        original = raw[int(unit["file_start"], 16):int(unit["file_start"], 16) + unit["size"]]
        if fn["chunks"] != [[hex(address), hex(address + unit["size"])]] or b"".join(bytes.fromhex(inventory["instructions"][ea]["file_bytes"]) for ea in fn["instructions"]) != original:
            raise ValueError("Original source-unit candidate extent differs")
        for variant, source in variants(unit).items():
            for flag_id, flags in FLAG_SETS.items():
                name = unit["id"].lower() + "_" + variant + "_" + flag_id
                source_path = args.output / (name + ".c")
                source_path.write_text(source, encoding="ascii")
                shutil.copyfile(source_path, compile_root / source_path.name)
                obj_path = args.output / (name + ".obj")
                command = ["wcc", "-bt=dos", "-ms", "-0", "-s", "-ecc", "-zld", *flags,
                           "-fo=" + obj_path.name, source_path.name]
                result = subprocess.run(command, cwd=compile_root, capture_output=True, timeout=20)
                (args.output / (name + ".log")).write_bytes(result.stdout + result.stderr)
                row = {"unit": unit, "variant": variant, "flag_id": flag_id, "source_sha256": sha(source_path),
                       "command": command, "compile_returncode": result.returncode, "artifact_stem": name}
                if result.returncode:
                    row.update(status="COMPILE_FAILED")
                else:
                    shutil.copyfile(compile_root / obj_path.name, obj_path)
                    row["object_sha256"] = sha(obj_path)
                    subprocess.run(["wdis", "-l=" + str(args.output / (name + ".lst")), str(obj_path)],
                                   stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT, check=True, timeout=20)
                    try:
                        code, fixups = resolve_ds_offsets(read_object(obj_path), unit["public_symbol"],
                            {symbol: int(offset, 16) for symbol, offset in unit["external_DS_offsets"].items()},
                            signed_addends=unit.get("encoded_addend_mode") == "signed16")
                    except (ValueError, UnsupportedOMF) as error:
                        row.update(status="REFUSED", reason=str(error))
                    else:
                        (args.output / (name + "-code.bin")).write_bytes(code)
                        row.update(status="MATCH" if code == original else "DIFF", code_hex=code.hex(), code_size=len(code),
                                   actual_fixups=fixups, original_hex=original.hex(), whole_range_equal=code == original)
                results.append(row)
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__)),
               "OMF_parser_sha256": sha(ROOT / "tools/omf_matching_probe.py"), "original_sha256": EXPECTED,
               "IDA_inventory_sha256": sha(args.ida_inventory), "candidate_manifest_sha256": sha(manifest_path),
               "compiler_payload_manifest_sha256": sha(payload_path), "results": results,
               "wall_seconds": time.monotonic() - started, "C_coverage_increment": 0,
               "scope": "Compiler experiment only; original types/compiler/unit classification unknown"}
    # Positive control for the vendor's actual LOOP emitter. Unlike the
    # original candidates, this authored arithmetic function has no DQ3
    # identity and must never enter original source coverage.
    control_source = args.output / "longshift_control.c"
    control_source.write_text(
        'unsigned long longshift_control(unsigned long value, unsigned count);\n'
        '#pragma aux longshift_control "_*" parm [dx ax] [cx] value [dx ax] modify exact [ax dx cx];\n'
        'unsigned long longshift_control(unsigned long value, unsigned count) { return value << count; }\n',
        encoding="ascii")
    shutil.copyfile(control_source, compile_root / control_source.name)
    command = ["wcc", "-bt=dos", "-ms", "-0", "-s", "-ecc", "-zld", *FLAG_SETS["previous"],
               "-fo=longshift_control.obj", control_source.name]
    compiled = subprocess.run(command, cwd=compile_root, capture_output=True, timeout=20, check=True)
    (args.output / "longshift_control.log").write_bytes(compiled.stdout + compiled.stderr)
    control_obj = args.output / "longshift_control.obj"
    shutil.copyfile(compile_root / control_obj.name, control_obj)
    control_code, control_fixes = resolve_ds_offsets(read_object(control_obj), "_longshift_control", {})
    (args.output / "longshift_control-code.bin").write_bytes(control_code)
    control_listing = args.output / "longshift_control.lst"
    subprocess.run(["wdis", "-l=" + str(control_listing), str(control_obj)],
                   stdout=subprocess.DEVNULL, stderr=subprocess.STDOUT, check=True, timeout=20)
    loop_lines = [line for line in control_listing.read_text().splitlines()
                  if re.search(r"^\s*[0-9a-fA-F]+\s+E2\s+[0-9a-fA-F]{2}\s+loop\b", line)]
    if not loop_lines:
        raise ValueError("Authored long-shift positive control did not emit LOOP")
    receipt["longshift_positive_control"] = {
        "source_sha256": sha(control_source), "object_sha256": sha(control_obj), "command": command,
        "code_hex": control_code.hex(), "actual_fixups": control_fixes, "loop_listing_lines": loop_lines,
        "scope": "Authored compiler arithmetic control; not an original source unit or compiler identity proof",
    }
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    summary = {}
    for unit in ("sub_132A3", "sub_16FCF"):
        group = [r for r in results if r["unit"]["id"] == unit]
        summary[unit] = {"cases": len(group), "matches": sum(r["status"] == "MATCH" for r in group),
                         "shortest": min((r["code_size"] for r in group if "code_size" in r), default=None),
                         "distinct_resolved_code": len({r["code_hex"] for r in group if "code_hex" in r})}
    print(json.dumps({"summary": summary, "wall_seconds": receipt["wall_seconds"]}))


if __name__ == "__main__":
    main()
