"""Docker-only MSC 5.1 ABI controls, scoped to the mounted candidate, docs/25.

Known C inputs make these compiler experiments, not original-language evidence.
Every failed keyword/source remains a failure; no flags are silently retried.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import time

from omf_matching_probe import read_object, select_function


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use the host UID/GID")
    args.output.mkdir(exist_ok=False)
    cases = {
        "word": "unsigned abiword(unsigned limit) { return limit; }\n",
        "pair": "unsigned long abipair(unsigned limit) { return ((unsigned long)limit << 16) | (unsigned long)(limit >> 1); }\n",
        "fast": "unsigned _fastcall abifast(unsigned limit) { return limit; }\n",
    }
    commands = ["@echo off", "set PATH=C:\\BIN", "set INCLUDE=C:\\INCLUDE\\INCLUDE",
                "set LIB=C:\\LIB", "set TMP=D:\\", "d:"]
    for stem, text in cases.items():
        (args.output / (stem + ".c")).write_text(text, encoding="ascii")
        commands += ["C:\\BIN\\CL.EXE /c /AS /Ox " + stem + ".c > " + stem + ".log",
                     "if errorlevel 1 echo FAIL > " + stem + ".err"]
    commands += ["echo DONE > DONE.TXT"]
    (args.output / "go.bat").write_text("\r\n".join(commands) + "\r\n", encoding="ascii")
    config = ("[sdl]\noutput=surface\n[cpu]\ncycles=fixed 100000\n[mixer]\nnosound=true\n"
              "[autoexec]\nmount c /msc\nmount d " + str(args.output) + "\nd:\ncall go.bat\nexit\n")
    (args.output / "dosbox.conf").write_text(config, encoding="ascii")
    started = time.monotonic()
    with (args.output / "dosbox.log").open("wb") as log:
        process = subprocess.run(["dosbox", "-conf", str(args.output / "dosbox.conf"), "-exit"],
                                 stdout=log, stderr=subprocess.STDOUT, timeout=90, check=False)
    if process.returncode or not (args.output / "DONE.TXT").exists():
        raise ValueError("Compiler batch did not finish")
    results = []
    for stem, text in cases.items():
        obj_path = args.output / (stem.upper() + ".OBJ")
        marker = args.output / (stem.upper() + ".ERR")
        failure = marker.exists() and marker.read_text().strip() == "FAIL"
        record = {"case": stem, "source_sha256": hashlib.sha256(text.encode()).hexdigest(),
                  "flags": "/c /AS /Ox", "compile_status": "failed" if failure else "compiled",
                  "compiler_stdout": (args.output / (stem.upper() + ".LOG")).read_text(errors="replace")}
        if failure:
            record["object_present"] = obj_path.exists()
        else:
            obj = read_object(obj_path)
            code, public = select_function(obj, "_abi" + stem)
            if any(f["segment_index"] == public["segment_index"] for f in obj["fixups"]):
                raise ValueError("ABI echo control unexpectedly requires fixups")
            binary = args.output / (stem + "-code.bin")
            binary.write_bytes(code)
            record.update(public=public, code_hex=code.hex(), code_size=len(code),
                          obj_sha256=obj["sha256"], code_sha256=hashlib.sha256(code).hexdigest())
        results.append(record)
    if not all(x["compile_status"] == "compiled" for x in results[:2]):
        raise ValueError("Required ordinary C controls failed")
    tool_inputs = {p.name: hashlib.sha256(p.read_bytes()).hexdigest() for p in Path("/msc/BIN").glob("*.EXE")}
    receipt = {"scope": "ABI of the mounted MSC candidate only; original compiler remains unknown",
               "cases": results, "compiler_inputs": tool_inputs, "wall_seconds": time.monotonic() - started,
               "producer_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest()}
    (args.output / "receipt.json").write_text(json.dumps(receipt, indent=2) + "\n")
    print(json.dumps({"cases": [(x["case"], x["compile_status"], x.get("code_size")) for x in results],
                      "seconds": receipt["wall_seconds"]}))


if __name__ == "__main__":
    main()
