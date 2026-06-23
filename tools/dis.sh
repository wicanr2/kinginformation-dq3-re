#!/usr/bin/env bash
# capstone 16-bit 反組譯捷徑(docker,不污染 host)。用法: tools/dis.sh <file_off_hex> [count]
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
exec docker run --rm -v "$ROOT":/work -w /work \
  ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone && python tools/re_disasm.py $*"
