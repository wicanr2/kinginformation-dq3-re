"""Bounded, source-independent Inertia trials for DQ3 Issue #5, docs/25.

This is a diagnostics runner, not an original-behavior oracle. Generated C and
zero exit status do not establish byte matching or semantic validation.
IDA linear, MZ logical, file offsets and Inertia linear addresses stay distinct.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time


ROOT = Path(__file__).resolve().parents[1]
EXPECTED_HASH = "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"
SOURCE_COMMIT = "c555363b810d3a6df786e5d6511d1bb28fa82333"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--inertia-root", type=Path, default=Path("/opt/inertia"))
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use the host UID/GID")
    original = ROOT / "assets_raw/DQ3.EXE"
    if sha(original) != EXPECTED_HASH or original.stat().st_size != 115282:
        raise ValueError("Original executable identity differs")
    args.output.mkdir(exist_ok=False)
    actual_commit = subprocess.check_output(
        ["git", "-c", "safe.directory=" + str(args.inertia_root), "-C", str(args.inertia_root),
         "rev-parse", "HEAD"], text=True).strip()
    if actual_commit != SOURCE_COMMIT:
        raise ValueError("Inertia source commit differs")
    # Use the CLI's own loader, which converts the 0x1000 paragraph option to
    # linear 0x10000. The backend class default alone is a different contract.
    loader_check = (
        "import json,sys; from pathlib import Path; sys.path.insert(0, '/opt/inertia'); "
        "import inertia.frontend.x86_16.load_dos_mz; import inertia.frontend.x86_16.simos_86_16; "
        "from inertia.cli.project_loading import _build_project; "
        "p=_build_project(Path(sys.argv[1]),force_blob=False,base_addr=0x1000,entry_point=0x1000); "
        "o=p.loader.main_object; "
        "print(json.dumps({'mapped_base':o.mapped_base,'entry':o.entry,"
        "'rng_bytes':p.loader.memory.load(o.mapped_base+0xe6b9,16).hex()}))"
    )
    environment = os.environ.copy()
    environment.update(PYTHONHASHSEED="0", PYTHON_JIT="1", INERTIA_ENABLE_TAIL_VALIDATION="1",
                       INERTIA_VEX_BACKEND="python")
    with (args.output / "loader-check.stdout").open("w") as log, \
            (args.output / "loader-check.stderr").open("w") as error:
        checked = subprocess.run([sys.executable, "-c", loader_check, str(original)], cwd=args.output,
                                 env=environment, stdout=log, stderr=error, timeout=45, check=False)
    if checked.returncode:
        raise ValueError("Inertia MZ loader smoke failed; preserve logs and investigate")
    coordinates = json.loads((args.output / "loader-check.stdout").read_text())
    if coordinates["mapped_base"] != 0x10000 or coordinates["rng_bytes"] != original.read_bytes()[0xFA29:0xFA39].hex():
        raise ValueError("Inertia address base or RNG source bytes differ")
    targets = [("rng", 0xE6B9, 32), ("rng_bound", 0xE6C9, 48), ("npc_mover", 0x2025, 160)]
    results = []
    for label, logical, window in targets:
        directory = args.output / label
        directory.mkdir()
        address = coordinates["mapped_base"] + logical
        command = [sys.executable, str(args.inertia_root / "decompile.py"), str(original),
                   "--addr", hex(address), "--timeout", "45", "--max-memory-mb", "2048",
                   "--max-functions", "1", "--window", str(window), "--c-target", "msc-dos",
                   "--output-c-dir", str(directory / "generated"), "--dump-layers",
                   "--dump-layer-dir", str(directory / "layers"), "--ignore-local-sidecar-hints",
                   "--no-alternate-source-c", "-q"]
        started = time.monotonic()
        with (directory / "stdout.txt").open("w") as log, (directory / "stderr.txt").open("w") as error:
            try:
                process = subprocess.run(command, cwd=directory, env=environment,
                                         stdout=log, stderr=error, timeout=70, check=False)
            except subprocess.TimeoutExpired:
                exit_code = None
                status = "outer_timeout"
            else:
                exit_code = process.returncode
                status = "process_completed" if exit_code == 0 else "process_failed"
        generated = []
        for path in sorted(directory.rglob("*")):
            if path.is_file() and path.suffix.lower() in (".c", ".h", ".json"):
                generated.append({"path": str(path.relative_to(args.output)), "size": path.stat().st_size,
                                  "sha256": sha(path)})
        result = {"label": label, "address_spaces": {"logical": hex(logical),
                  "IDA_linear": hex(0x10000 + logical), "file": hex(0x1370 + logical),
                  "Inertia_linear": hex(address)}, "command": command, "process_status": status,
                  "exit_code": exit_code, "wall_seconds": time.monotonic() - started,
                  "generated_files": generated, "byte_match_status": "not_tested",
                  "semantic_validation_status": "unreviewed", "inference_level": "unknown"}
        results.append(result)
        print(json.dumps({"case": label, "status": status, "exit_code": exit_code,
                          "seconds": result["wall_seconds"], "artifacts": len(generated)}), flush=True)
    receipt = {"schema_version": 1, "input": {"path": str(original), "size": original.stat().st_size,
               "sha256": EXPECTED_HASH}, "tool": {"name": "Inertia", "source_commit": actual_commit,
        "python": sys.version, "uv_lock_sha256": sha(args.inertia_root / "uv.lock"),
               "runner_sha256": sha(Path(__file__)), "lifter_backend": "python",
               "backend_scope": "Explicit upstream reference mode; no Cython extension built"},
               "loader_check": coordinates, "cases": results,
               "limitation": "Exit code, generated C and tool self-validation cannot substitute for original byte matching"}
    output = args.output / "receipt.json"
    output.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"receipt": str(output), "cases": len(results), "uid": output.stat().st_uid}))


if __name__ == "__main__":
    main()
