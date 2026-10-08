"""Docker-only build/run wrapper for the controlled RNG ABI probe in docs/25.

Copy only non-test dosgolem internal Go source and go.mod into a temporary module.
Original engine/input files remain read-only. No cmd/probe scratch is consumed.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import time


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--assembly-root", type=Path, required=True)
    parser.add_argument("--dosgolem-root", type=Path, default=Path("/dosgolem"))
    parser.add_argument("--compiler-controls", type=Path)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Run with the host UID/GID")
    args.output.mkdir(exist_ok=False)
    repository = Path(__file__).resolve().parents[1]
    engine = args.dosgolem_root
    source = repository / "tools/dosgolem_matching_rng_abi.go"
    original = repository / "assets_raw/DQ3.EXE"
    assembled = args.assembly_root / "RE-RNG-46-BYTES.EXE"
    assembly_receipt = args.assembly_root / "assembly-receipt.json"
    record = json.loads(assembly_receipt.read_text())
    if record["source_assembled_bytes"] != 46 or sha(original) != record["input"]["sha256"]:
        raise ValueError("Assembly receipt scope or original input differs")
    if sha(assembled) != sha(original):
        raise ValueError("Source-assembled local scaffold bytes differ")
    sources = {}
    with tempfile.TemporaryDirectory(prefix="dq3-rng-abi-") as temporary:
        root = Path(temporary)
        for path in sorted((engine / "internal").rglob("*.go")):
            if path.name.endswith("_test.go"):
                continue
            relative = path.relative_to(engine)
            destination = root / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, destination)
            sources[str(relative)] = sha(path)
        shutil.copyfile(engine / "go.mod", root / "go.mod")
        sources["go.mod"] = sha(engine / "go.mod")
        target = root / "cmd/matching-rng-abi/main.go"
        target.parent.mkdir(parents=True)
        shutil.copyfile(source, target)
        subprocess.run(["gofmt", "-w", str(target)], check=True, timeout=20)
        (args.output / "probe-source.go").write_bytes(target.read_bytes())
        environment = os.environ.copy()
        environment.update(GOCACHE=str(root / "build-cache"), GOMODCACHE=str(root / "module-cache"),
                           GOPROXY="off", GOTOOLCHAIN="local", CGO_ENABLED="0", GOMAXPROCS="2")
        started = time.monotonic()
        binary = args.output / "rng-abi-probe"
        with (args.output / "build.log").open("wb") as log:
            subprocess.run(["go", "build", "-p", "2", "-trimpath", "-o", str(binary),
                            "./cmd/matching-rng-abi"], cwd=root, env=environment,
                           stdout=log, stderr=subprocess.STDOUT, timeout=180, check=True)
        build_seconds = time.monotonic() - started
    output = args.output / "receipt.json"
    command = [str(binary), "-exe", str(original), "-rebuilt", str(assembled), "-output", str(output)]
    if args.compiler_controls is not None:
        controls = json.loads((args.compiler_controls / "receipt.json").read_text())
        for label in ("word", "pair"):
            control = next(case for case in controls["cases"] if case["case"] == label)
            code = args.compiler_controls / (label + "-code.bin")
            if control["compile_status"] != "compiled" or sha(code) != control["code_sha256"]:
                raise ValueError("Known C compiler control differs")
            command.extend(["-" + label + "-control", str(code)])
    with (args.output / "run.log").open("wb") as log:
        subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=120, check=True)
    result = json.loads(output.read_text())
    assert result["counts"] == {"core_all_u16_seeds": 65536, "bound10_all_u16_seeds": 65536,
                               "bound0_all_u16_seeds": 65536, "boundary_inputs": 84}
    assert result["negative_wrong_AX_remainder_assumption_rejected"]
    assert result["scope"]["normal_player_path"] is False
    metadata = {
        "original_sha256": sha(original), "assembly_receipt_sha256": sha(assembly_receipt),
        "producer_sha256": sha(Path(__file__)), "authored_go_source_sha256": sha(source),
        "formatted_go_source_sha256": sha(args.output / "probe-source.go"),
        "go_version": subprocess.check_output(["go", "version"], text=True).strip(),
        "engine_sources": sources,
        "engine_sources_canonical_sha256": hashlib.sha256(json.dumps(sources, sort_keys=True, separators=(",", ":")).encode()).hexdigest(),
        "engine_tree_scope": "non-test internal Go files and go.mod only; external cmd/probe scratch excluded",
        "build_seconds": build_seconds, "binary_sha256": sha(binary),
        "receipt_sha256": sha(output), "UID": output.stat().st_uid, "GID": output.stat().st_gid,
        "compiler_controls_receipt_sha256": sha(args.compiler_controls / "receipt.json") if args.compiler_controls is not None else None,
    }
    (args.output / "producer-meta.json").write_text(json.dumps(metadata, indent=2) + "\n")
    print(json.dumps({"counts": result["counts"], "CPU_probe_seconds": result["wall_seconds"],
                      "build_seconds": build_seconds, "receipt": str(output), "receipt_sha256": sha(output)}))


if __name__ == "__main__":
    main()
