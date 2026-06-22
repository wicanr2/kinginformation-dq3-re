#!/usr/bin/env bash
# 在 docker uv venv 內跑反組譯 Python 工具 (capstone),不污染 host。
# 用法: tools/dockrun_cap.sh tools/re_loaders.py disasm 0xfe64 60
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
exec docker run --rm \
  -v "$ROOT":/work -w /work \
  ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone && python $*"
