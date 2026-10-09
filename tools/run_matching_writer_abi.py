"""Docker-only controlled AX-live writer oracle; docs/25-match-progress.md.

Copy the non-test dosgolem internal module into /tmp. The original EXE,
compiler artifacts and engine source stay read-only. No campaign claim.
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


ROOT = Path(__file__).resolve().parents[1]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def code_from_receipt(path, case_id, exact):
    receipt = json.loads(path.read_text())
    case = next(r for r in receipt["results"] if r["case"] == case_id)
    code = path.parent / (case_id + "-code.bin")
    if (case["status"] != "RESOLVED" or case["source_has_assembly_instructions"]
            or case["original_compare"]["byte_exact"] != exact
            or sha(code) != case["code_sha256"]
            or sha(ROOT / case["source_unit"]["source"]) != case["source_sha256"]):
        raise ValueError("Source-compiled code identity differs")
    return code


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--positive", type=Path, required=True)
    parser.add_argument("--negative", type=Path, required=True)
    parser.add_argument("--dosgolem-root", type=Path, default=Path("/dosgolem"))
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    good = code_from_receipt(args.positive, "sub_1cfe8", True)
    bad = code_from_receipt(args.negative, "bad_drop_ax", False)
    original = ROOT / "assets_raw/DQ3.EXE"
    source = ROOT / "tools/dosgolem_matching_writer_abi.go"
    if sha(original) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c":
        raise ValueError("Original EXE differs")
    args.output.mkdir(exist_ok=False)
    sources = {}
    with tempfile.TemporaryDirectory(prefix="dq3-writer-abi-") as temporary:
        root = Path(temporary)
        paths = sorted((args.dosgolem_root / "internal").rglob("*.go"))
        paths = [p for p in paths if not p.name.endswith("_test.go")]
        paths.append(args.dosgolem_root / "go.mod")
        for path in paths:
            relative = path.relative_to(args.dosgolem_root)
            data = path.read_bytes()
            target = root / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(data)
            sources[str(relative)] = hashlib.sha256(data).hexdigest()
        for relative, digest in sources.items():
            if sha(args.dosgolem_root / relative) != digest:
                raise ValueError("Engine source changed during its snapshot")
        target = root / "cmd/matching-writer-abi/main.go"
        target.parent.mkdir(parents=True)
        shutil.copyfile(source, target)
        subprocess.run(["gofmt", "-w", str(target)], check=True, timeout=20)
        (args.output / "probe-source.go").write_bytes(target.read_bytes())
        environment = os.environ.copy()
        environment.update(GOCACHE=str(root / "cache"), GOMODCACHE=str(root / "modules"),
                           GOPROXY="off", GOTOOLCHAIN="local", CGO_ENABLED="0", GOMAXPROCS="2")
        binary = args.output / "writer-abi-probe"
        started = time.monotonic()
        with (args.output / "build.log").open("wb") as log:
            subprocess.run(["go", "build", "-p", "2", "-trimpath", "-buildvcs=false",
                            "-o", str(binary), "./cmd/matching-writer-abi"], cwd=root,
                           env=environment, stdout=log, stderr=subprocess.STDOUT,
                           timeout=180, check=True)
        build_seconds = time.monotonic() - started
    output = args.output / "receipt.json"
    with (args.output / "run.log").open("wb") as log:
        subprocess.run([str(binary), "-exe", str(original), "-good", str(good),
                        "-bad", str(bad), "-output", str(output)], stdout=log,
                       stderr=subprocess.STDOUT, timeout=90, check=True)
    result = json.loads(output.read_text())
    if (result["counts"] != {"all_u16_AX": 65536, "boundary_vectors": 384}
            or result["full_VM_memory_samples"] != 256
            or result["discarded_AX_control"]["AX_return_mismatches"] != 65568
            or result["scope"]["normal_player_path"] or result["scope"]["RNG_executed"]):
        raise ValueError("Controlled oracle receipt scope or counts differ")
    metadata = {
        "producer_sha256": sha(Path(__file__)), "authored_go_source_sha256": sha(source),
        "formatted_go_source_sha256": sha(args.output / "probe-source.go"),
        "original_sha256": sha(original), "positive_receipt_sha256": sha(args.positive),
        "negative_receipt_sha256": sha(args.negative), "engine_sources": sources,
        "engine_sources_canonical_sha256": hashlib.sha256(json.dumps(sources, sort_keys=True,
                                                     separators=(",", ":")).encode()).hexdigest(),
        "engine_scope": "snapshot of non-test internal Go files and go.mod; cmd scratch excluded",
        "go_version": subprocess.check_output(["go", "version"], text=True).strip(),
        "build_seconds": build_seconds, "binary_sha256": sha(binary),
        "receipt_sha256": sha(output), "UID": output.stat().st_uid, "GID": output.stat().st_gid,
    }
    (args.output / "producer-meta.json").write_text(json.dumps(metadata, indent=2) + "\n")
    print(json.dumps({"counts": result["counts"], "AX_negative_mismatches": 65568,
                      "CPU_seconds": result["wall_seconds"], "receipt_sha256": sha(output)}))


if __name__ == "__main__":
    main()
