"""Docker-only fixed official Watcom archive extraction; entry tools/build/README.

The existing host Watcom image contains wcc386 but not the 16-bit wcc backend.
Verify the complete pinned archive, then preserve a per-file payload manifest.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import urllib.request
import zipfile


URL = "https://github.com/open-watcom/open-watcom-v2/releases/download/2026-10-01-Build/open-watcom-2_0-c-linux-x64"
HASH = "a961f3e02ce27bcd88428345dc57483457a623b6e718b6cee52c48d24ab8065f"
SIZE = 129081693


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if os.getuid() == 0:
        raise ValueError("Use host UID/GID")
    args.output.mkdir(exist_ok=False)
    archive = args.output / "official-archive.bin"
    digest = hashlib.sha256()
    count = 0
    request = urllib.request.Request(URL, headers={"User-Agent": "DQ3 matching research"})
    with urllib.request.urlopen(request, timeout=60) as response, archive.open("xb") as stream:
        while block := response.read(1024 * 1024):
            count += len(block)
            if count > SIZE:
                raise ValueError("Official archive exceeds pinned size")
            digest.update(block)
            stream.write(block)
    if count != SIZE or digest.hexdigest() != HASH:
        raise ValueError("Complete official archive identity differs")
    payload = args.output / "payload"
    payload.mkdir()
    records = []
    with zipfile.ZipFile(archive) as source:
        required = {"binl64/wcc", "binl64/wcc386", "binl64/wdis", "binl64/wlink", "binl64/wlib"}
        names = set(source.namelist())
        if not required <= names:
            raise ValueError("Official archive lacks required 16-bit/native tools")
        selected = sorted(required | {name for name in names if name.startswith("h/") and not name.endswith("/")})
        for name in selected:
            parts = Path(name).parts
            if name.startswith("/") or ".." in parts:
                raise ValueError("Unsafe vendor archive member")
            data = source.read(name)  # ZIP CRC independently checked by zipfile.
            destination = payload / name
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_bytes(data)
            if name in required:
                destination.chmod(0o755)
            records.append({"path": name, "size": len(data), "sha256": hashlib.sha256(data).hexdigest()})
    manifest = {"schema_version": 1, "release": "2026-10-01-Build", "URL": URL,
                "complete_archive_sha256": HASH, "complete_archive_size": SIZE,
                "complete_archive_verified": True, "files": records,
                "producer_sha256": hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
                "scope": "Official compiler candidate only; no original DQ3 compiler identity claim"}
    (payload / "source-manifest.json").write_text(json.dumps(manifest, indent=2) + "\n")
    print(json.dumps({"archive_verified": True, "files": len(records), "wcc": next(row for row in records if row["path"] == "binl64/wcc")}))


if __name__ == "__main__":
    main()
