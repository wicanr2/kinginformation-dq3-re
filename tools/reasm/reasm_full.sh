#!/usr/bin/env bash
# 完整 RE 重組驗證(docker 內,不污染 host):
#   DQ3.EXE → exe_to_nasm.py 產 .asm → nasm -f bin 組回 → sha256 比對原版。
# PASS = 我們產生的反組譯原始碼,經獨立組譯器 nasm 組出與原版完全相同的二進位。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
mkdir -p "$ROOT/work/reasm"
docker run --rm -v "$ROOT":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim bash -c '
  set -e
  apt-get update -qq >/dev/null 2>&1 && apt-get install -y -qq nasm >/dev/null 2>&1
  uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone >/dev/null 2>&1
  python3 tools/reasm/exe_to_nasm.py work/reasm/dq3.asm
  nasm -f bin -o work/reasm/dq3.rebuilt work/reasm/dq3.asm
  echo "--- sha256 ---"
  sha256sum assets_raw/DQ3.EXE work/reasm/dq3.rebuilt
  A=$(sha256sum assets_raw/DQ3.EXE | cut -d" " -f1)
  B=$(sha256sum work/reasm/dq3.rebuilt | cut -d" " -f1)
  if [ "$A" = "$B" ]; then echo "RESULT: PASS — nasm 重組 byte-identical (100%)";
  else echo "RESULT: FAIL — sha256 differs";
       cmp assets_raw/DQ3.EXE work/reasm/dq3.rebuilt || true; fi
'
