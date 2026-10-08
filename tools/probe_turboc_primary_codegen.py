"""Docker-only existing Turbo C primary-loop control, indexed in docs/25.

Use native C register pseudovariables without assembly or byte injection.
Unknown OMF group frames remain refused; no runtime or coverage is approved.
"""

import argparse
from datetime import datetime, timezone
import hashlib
import json
import os
from pathlib import Path
import subprocess
import zipfile

from omf_matching_probe import UnsupportedOMF, read_object, resolve_ds_offsets


ROOT = Path(__file__).resolve().parents[1]
ARCHIVE_HASH = "2c87f988605ae9ed70e5fef35b9854de87e36ccdb021caec35ab2424ca4b5553"
COMPILER_HASH = "19650666dcaa03e3f68efd9beeb57821ba4ed4d84c88a52b0e989bbfc97e07ba"
ORIGINAL_HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
SOURCES = {
    "ordinary": "extern volatile unsigned char unknown_DS_265d[]; void sub_32a3(void) { register unsigned count=13, index=8; do { unknown_DS_265d[index]=255;index+=2;}while(--count); }\n",
    "pseudoreg": "void sub_32a3(void) {_CX=13;_BX=8;do {((volatile unsigned char near *)0x265d)[_BX]=255;_BX+=2;}while(--_CX);}\n",
}
FLAGS = {"register": "-r", "jump": "-r -O", "speed": "-r -O -G", "reg_opt": "-r -O -G -Z"}
SOURCE_EPOCH = int(datetime(2026, 10, 1, tzinfo=timezone.utc).timestamp())


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--ida-inventory", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    original_path = ROOT / "assets_raw/DQ3.EXE"
    raw = original_path.read_bytes()
    if len(raw) != 115282 or sha(original_path) != ORIGINAL_HASH:
        raise ValueError("Original identity differs")
    inventory = json.loads(args.ida_inventory.read_text())
    if inventory["input"]["sha256"] != ORIGINAL_HASH or inventory["tool"]["version"] != "9.4":
        raise ValueError("IDA input/version differs")
    unit = next(fn for fn in inventory["functions"] if fn["ida_linear_start"] == "0x132a3")
    original = raw[0x4613:0x4624]
    if unit["chunks"] != [["0x132a3", "0x132b4"]] or b"".join(bytes.fromhex(inventory["instructions"][ea]["file_bytes"]) for ea in unit["instructions"]) != original:
        raise ValueError("Full original candidate extent differs")
    archive_path = ROOT / "tools/build/tc201.zip"
    compiler_path = ROOT / "tools/build/tc_extract/Disk2/TCC.EXE"
    if sha(archive_path) != ARCHIVE_HASH or sha(compiler_path) != COMPILER_HASH:
        raise ValueError("Existing compiler input differs")
    with zipfile.ZipFile(archive_path) as archive:
        if compiler_path.read_bytes() != archive.read("Disk2/TCC.EXE"):
            raise ValueError("Existing compiler differs from pinned archive")
    args.output.mkdir(exist_ok=False)
    commands = ["@echo off", "d:", r"C:\DISK2\TCC.EXE > HELP.TXT"]
    cases = []
    for variant, content in SOURCES.items():
        for flag_id, flags in FLAGS.items():
            stem = "C" + format(len(cases), "03d")
            source = args.output / (stem + ".C")
            source.write_text(content, encoding="ascii")
            # TCC records the input's DOS file time in COMENT class E9.
            # Make the actual input reproducible, never mask object metadata.
            os.utime(source, (SOURCE_EPOCH, SOURCE_EPOCH))
            command = r"C:\DISK2\TCC.EXE -ms -c " + flags + " " + stem + ".C > " + stem + ".LOG"
            commands.append(command)
            cases.append({"stem": stem, "variant": variant, "flags": flags, "flag_id": flag_id,
                          "source_sha256": sha(source), "compiler_command": command})
    commands.append("echo complete > DONE.TXT")
    (args.output / "go.bat").write_text("\r\n".join(commands) + "\r\n", encoding="ascii")
    config = ("[sdl]\noutput=surface\n[dosbox]\nmachine=svga_s3\n[cpu]\ncycles=max\n[autoexec]\n"
              + "mount c " + str(ROOT / "tools/build/tc_extract") + "\nmount d " + str(args.output)
              + "\nd:\ncall go.bat\nexit\n")
    conf_path = args.output / "dosbox.conf"
    conf_path.write_text(config)
    with (args.output / "dosbox.log").open("wb") as stream:
        result = subprocess.run(["dosbox", "-conf", str(conf_path), "-exit"],
                                env={**os.environ, "SDL_VIDEODRIVER": "dummy", "SDL_AUDIODRIVER": "dummy"},
                                stdout=stream, stderr=subprocess.STDOUT, timeout=90)
    if result.returncode or not (args.output / "DONE.TXT").is_file():
        raise ValueError("Compiler batch did not finish")
    for case in cases:
        obj_path = args.output / (case["stem"] + ".OBJ")
        if not obj_path.is_file():
            case.update(status="COMPILE_FAILED")
            continue
        case["object_sha256"] = sha(obj_path)
        try:
            code, fixes = resolve_ds_offsets(read_object(obj_path), "_sub_32a3", {"_unknown_DS_265d": 0x265d})
        except (ValueError, UnsupportedOMF) as error:
            case.update(status="REFUSED", reason=str(error))
        else:
            (args.output / (case["stem"] + "-code.bin")).write_bytes(code)
            case.update(status="MATCH" if code == original else "DIFF", code_hex=code.hex(), original_hex=original.hex(),
                        code_size=len(code), actual_fixups=fixes, whole_range_equal=code == original)
    receipt = {"schema_version": 1, "producer_sha256": sha(Path(__file__)),
               "OMF_parser_sha256": sha(ROOT / "tools/omf_matching_probe.py"), "original_sha256": ORIGINAL_HASH,
               "IDA_inventory_sha256": sha(args.ida_inventory), "archive_sha256": ARCHIVE_HASH, "compiler_sha256": COMPILER_HASH,
               "compiler_banner": (args.output / "HELP.TXT").read_text(errors="replace"), "generated_source_epoch": SOURCE_EPOCH,
               "original_IDA_linear_range": ["0x132a3", "0x132b4"], "original_logical_range": ["0x32a3", "0x32b4"],
               "original_file_range": ["0x4613", "0x4624"], "results": cases, "C_coverage_increment": 0,
               "scope": "Existing compiler experiments; ordinary group frame remains unsupported; original compiler unknown"}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"cases": [(c["variant"], c["flag_id"], c["status"], c.get("code_size")) for c in cases]}))


if __name__ == "__main__":
    main()
