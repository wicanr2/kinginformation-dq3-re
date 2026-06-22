#!/usr/bin/env bash
# rebuild.sh — matching-decompilation full-rebuild:把目前所有「已 byte-match 的 C 函式」
# splice 進 docs/17 的 byte-identical 重組,輸出 work/RE-DQ3.EXE,sha256 與原版比。
#
# 框架正確性的基準(inv.2):0 個 match C 函式時,RE-DQ3.EXE 必須 == 原版 DQ3.EXE。
# 隨 match 函式增加(每個 exact-match 的 bytes 本就 == 原版,inv.1),sha256 維持相同,
# 逐步「以 C source 取代 db bytes」逼近 100% matching-decomp。
#
# 兩階段(避免 docker 內再起 docker):
#   階段 1(host):每個 re/match/*.c 用 msc_compile.sh(dq3-msc image)編成 .obj。
#   階段 2(docker uv+nasm):splice_rebuild.py --no-compile 抽 .obj 函式 bytes、
#                          只 splice exact-match 者、產 .asm;nasm 組成 EXE;sha256 比對。
#
# 用法: tools/build/rebuild.sh
# docker 同步前景執行,不污染 host。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
WORK="$ROOT/work/csplice"
mkdir -p "$WORK"
MATCH_DIR="$ROOT/re/match"

echo "=== 階段 1: 編譯 re/match/*.c (MSC 5.1, host docker dq3-msc) ==="
shopt -s nullglob
CFILES=("$MATCH_DIR"/*.c)
if [ ${#CFILES[@]} -eq 0 ]; then
  echo "  re/match/ 無 .c — 0 函式模式(驗證 splice 框架基準)"
else
  for c in "${CFILES[@]}"; do
    # 抽 @match marker 的 model(預設 /AS)
    MODEL=$(grep -oE '@match[^*]*model=(/A[SMCLH])' "$c" | grep -oE '/A[SMCLH]' | head -1 || true)
    MODEL="${MODEL:-/AS}"
    echo "  -> $(basename "$c")  model=$MODEL"
    bash "$ROOT/tools/build/msc_compile.sh" "$c" /c "$MODEL" /Ox >/dev/null 2>&1 || \
      echo "     (編譯失敗,跳過 — 不影響其餘函式與 0 函式基準)"
  done
fi

echo "=== 階段 2: splice + nasm 重組 + sha256 (docker uv+nasm) ==="
docker run --rm -v "$ROOT":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim bash -c '
  set -e
  apt-get update -qq >/dev/null 2>&1 && apt-get install -y -qq nasm >/dev/null 2>&1
  uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone >/dev/null 2>&1
  cd /work/tools/build
  python3 splice_rebuild.py /work/work/csplice/dq3.asm --no-compile
  cd /work
  nasm -f bin -o work/csplice/RE-DQ3.EXE work/csplice/dq3.asm
  cp work/csplice/RE-DQ3.EXE work/RE-DQ3.EXE
  echo "--- sha256 ---"
  sha256sum assets_raw/DQ3.EXE work/RE-DQ3.EXE
  A=$(sha256sum assets_raw/DQ3.EXE | cut -d" " -f1)
  B=$(sha256sum work/RE-DQ3.EXE | cut -d" " -f1)
  if [ "$A" = "$B" ]; then
    echo "RESULT: PASS — RE-DQ3.EXE byte-identical 原版 (splice 框架正確)";
  else
    echo "RESULT: FAIL — sha256 differs(有 exact-match 函式但仍不符,或 splice 框架有誤)";
    cmp assets_raw/DQ3.EXE work/RE-DQ3.EXE || true;
    exit 1;
  fi
'
echo "=== 覆蓋率 ==="
cat "$WORK/coverage.txt" 2>/dev/null || echo "  (無 coverage.txt)"
echo "輸出: work/RE-DQ3.EXE"
