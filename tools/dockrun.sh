#!/usr/bin/env bash
# 在 docker uv venv 內跑 Python 工具,不污染 host。
# 用法: tools/dockrun.sh tools/some_script.py [args...]
# 專案根目錄掛載到 /work;PIL 等套件用 uv 即時安裝在容器內。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
exec docker run --rm \
  -v "$ROOT":/work -w /work \
  ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q pillow && python $*"
