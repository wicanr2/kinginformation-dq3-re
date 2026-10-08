"""Docker-only Wasm addition to the verified Watcom16 payload; docs/25.

Revision r2 adds the official MASM-compatible assembler. The r1 source and
payload stay immutable; no host runtime, original game or private input.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import zipfile


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--original-source", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    archive = args.original_source / "official-archive.bin"
    if sha(archive) != "a961f3e02ce27bcd88428345dc57483457a623b6e718b6cee52c48d24ab8065f":
        raise ValueError("Complete official archive differs")
    original = args.original_source / "payload"
    manifest = json.loads((original / "source-manifest.json").read_text())
    if manifest["release"] != "2026-10-01-Build" or not manifest["complete_archive_verified"]:
        raise ValueError("Original payload identity differs")
    for record in manifest["files"]:
        if sha(original / record["path"]) != record["sha256"]:
            raise ValueError("Original payload file differs")
    args.output.mkdir(exist_ok=False)
    payload = args.output / "payload"
    shutil.copytree(original, payload)
    with zipfile.ZipFile(archive) as packed:
        data = packed.read("binl64/wasm")
    target = payload / "binl64/wasm"
    if target.exists():
        raise ValueError("r1 unexpectedly already contains Wasm")
    target.write_bytes(data)
    target.chmod(0o755)
    manifest["files"].append({"path": "binl64/wasm", "size": len(data), "sha256": sha(target)})
    manifest["revision"] = "2.0-20261001-r2"
    manifest["supersedes_for_assembly"] = "dq3-watcom16:2.0-20261001-r1"
    manifest["preparation_producer_sha256"] = sha(Path(__file__))
    (payload / "source-manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(json.dumps({"files": len(manifest["files"]), "wasm_sha256": sha(target), "revision": manifest["revision"]}))


if __name__ == "__main__":
    main()
