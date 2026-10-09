"""Docker-only bounded call-preservation controls; docs/25-match-progress.md.

These are unapproved C representations, not source-coverage claims. Original
instructions, ABI pragmas and fixed compiler commands remain visible.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys


ROOT = Path(__file__).resolve().parents[1]
INPUT_SHA = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
CLOBBERS = "[ax bx cx dx si di bp es]"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--ida-inventory", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    original = ROOT / "assets_raw/DQ3.EXE"
    inventory = json.loads(args.ida_inventory.read_text())
    if sha(original) != INPUT_SHA or inventory["input"]["sha256"] != INPUT_SHA:
        raise ValueError("Original input differs")
    if inventory["tool"]["version"] != "9.4":
        raise ValueError("Use reviewed IDA 9.4 evidence")
    args.output.mkdir(exist_ok=False)
    source_root = args.output / "sources"
    source_root.mkdir()
    common = "extern volatile unsigned unknown_DS_259c;\n"
    controls = [
        ("restore_void", "sub_15037", 0x5037, 23, "_sub_5037", {"_unknown_DS_259c": "0x259c"},
         common + f'''void sub_21414(unsigned record);
#pragma aux sub_21414 "_*" far parm [di] modify exact {CLOBBERS};
void sub_5037(void);
#pragma aux sub_5037 "_*" modify exact {CLOBBERS};
void sub_5037(void) {{ unsigned saved = unknown_DS_259c; unknown_DS_259c = 1; sub_21414(0x13a); unknown_DS_259c = saved; }}
''', "_sub_21414", 0x264),
        ("restore_ax_return", "sub_15037", 0x5037, 23, "_sub_5037", {"_unknown_DS_259c": "0x259c"},
         common + f'''unsigned sub_21414(unsigned record);
#pragma aux sub_21414 "_*" far parm [di] value [ax] modify exact {CLOBBERS};
unsigned sub_5037(void);
#pragma aux sub_5037 "_*" value [ax] modify exact {CLOBBERS};
unsigned sub_5037(void) {{ unsigned saved = unknown_DS_259c; unsigned result; unknown_DS_259c = 1; result = sub_21414(0x13a); unknown_DS_259c = saved; return result; }}
''', "_sub_21414", 0x264),
        ("register_envelope", "sub_1E8A9", 0xe8a9, 20, "_sub_e8a9", {},
         f'''void sub_2197f(void);
#pragma aux sub_2197f "_*" far modify exact {CLOBBERS};
void sub_e8a9(void);
#pragma aux sub_e8a9 "_*" modify exact [es];
void sub_e8a9(void) {{ sub_2197f(); }}
''', "_sub_2197f", 0x7cf),
    ]
    cases = []
    for name, identity, offset, size, symbol, ds, content, far_name, target in controls:
        function = next(f for f in inventory["functions"]
                        if int(f["ida_linear_start"], 16) == offset + 0x10000)
        if function["original_name"] != identity:
            raise ValueError("Original function identity differs")
        source = source_root / (name + ".c")
        source.write_text(content, encoding="ascii")
        cases.append({"id": name, "original_name": identity, "source": str(source.resolve()),
                      "public_symbol": symbol, "ida_linear_start": hex(offset + 0x10000),
                      "logical_start": hex(offset), "file_start": hex(offset + 0x1370), "size": size,
                      "external_DS_offsets": ds, "compiler_profile": "cdecl-size-calls-unframed",
                      "call_placements": {"caller": {"segment": 0, "offset": offset}, "near": {},
                                          "far": {far_name: {"segment": 0x111b, "offset": target}}},
                      "scope": "Unapproved compiler ABI control; never automatically counted as source coverage"})
    manifest = args.output / "manifest.json"
    manifest.write_text(json.dumps({"schema_version": 1, "input_sha256": INPUT_SHA,
                                    "cases": cases}, indent=2) + "\n")
    compiled = args.output / "compiled"
    with (args.output / "compile.log").open("wb") as log:
        subprocess.run([sys.executable, str(ROOT / "tools/run_watcom16_abi.py"),
                        "--candidate-manifest", str(manifest), "--ida-inventory", str(args.ida_inventory),
                        "--output", str(compiled)], stdout=log, stderr=subprocess.STDOUT,
                       timeout=120, check=True)
    receipt = json.loads((compiled / "receipt.json").read_text())
    results = []
    for case in receipt["results"]:
        row = {"case": case["case"], "original_name": case["source_unit"]["original_name"],
               "status": case["status"], "original_bytes": case["source_unit"]["size"],
               "compiled_bytes": case.get("code_size"),
               "byte_exact": case.get("original_compare", {}).get("byte_exact", False),
               "code_sha256": case.get("code_sha256")}
        results.append(row)
    result = {"producer_sha256": sha(Path(__file__)), "input_sha256": INPUT_SHA,
              "IDA_inventory_sha256": sha(args.ida_inventory),
              "compiler_receipt_sha256": sha(compiled / "receipt.json"), "results": results,
              "formal_C_coverage_increment": 0, "compiler_modified": False,
              "scope": "Fixed compiler/profile controls; any future exact result still needs independent review"}
    (args.output / "receipt.json").write_text(json.dumps(result, indent=2) + "\n")
    print(json.dumps({"results": results, "formal_C_coverage_increment": 0}))


if __name__ == "__main__":
    main()
