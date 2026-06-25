#!/usr/bin/env bash
# PPM → PNG 批次轉換(docker pillow;dq3_remake DQ3_DUMP 出的 PPM 轉成可存檔 PNG)。
# 用法:tools/ppm2png.sh <dir>   轉 <dir>/*.ppm → *.png 並刪 ppm。
set -eu
DIR="${1:-docs/img}"
docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim bash -c "
  uv venv -q /tmp/v && . /tmp/v/bin/activate && uv pip install -q pillow
  python3 - <<PY
from PIL import Image
import glob,os
for p in sorted(glob.glob('$DIR/*.ppm')):
    Image.open(p).save(p[:-4]+'.png'); os.remove(p); print('→',p[:-4]+'.png')
PY"
