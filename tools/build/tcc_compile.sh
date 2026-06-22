#!/usr/bin/env bash
# 在 docker(DOSBox + Turbo C 2.01)內把單一 .c 編成 .obj,dump 回 host。
# 不污染 host:DOSBox/TCC 全在容器內;只把工作目錄掛載進去。
#
# 用法: tools/build/tcc_compile.sh <src.c> [tcc_flags...]
#   預設 flags = -ml -c (large model, compile-only)。可覆寫(如 -mm -c / -ms -c)。
# 產出: <srcdir>/<name>.<model>.obj
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SRC="${1:?need src.c}"; shift || true
SRCDIR="$(cd "$(dirname "$SRC")" && pwd)"
SRCBASE="$(basename "$SRC")"
NAME="${SRCBASE%.*}"
FLAGS="${*:--ml -c}"
TAG="$(echo "$FLAGS" | grep -oE '\-m[smclh]' | head -1 | tr -d '-' || echo m)"

WORK="$(mktemp -d)"
cp "$SRCDIR/$SRCBASE" "$WORK/"
# DOS 8.3:來源直接放 D: 根目錄,輸出 .OBJ 也在 D:。
cat > "$WORK/go.bat" <<EOF
@echo off
d:
C:\\BIN\\TCC.EXE -IC:\\INCLUDE -LC:\\LIB $FLAGS D:\\$SRCBASE > D:\\TCC.OUT
EOF

docker run --rm \
  -e SDL_VIDEODRIVER=dummy -e SDL_AUDIODRIVER=dummy \
  -v "$WORK":/work dq3-turboc \
  bash -lc '
    set -e
    cat > /work/dosbox.conf <<CFG
[sdl]
output=surface
[dosbox]
machine=svga_s3
[autoexec]
mount c /tc
mount d /work
d:
go.bat
exit
CFG
    timeout 120 dosbox -conf /work/dosbox.conf -exit >/work/dosbox.out 2>&1 || true
    ls -la /work
  '

echo "=== TCC console output ==="
cat "$WORK/TCC.OUT" 2>/dev/null || true
echo "=== outputs in $WORK ==="
ls -la "$WORK"
if ls "$WORK"/*.OBJ >/dev/null 2>&1; then
  OUT="$SRCDIR/${NAME}.${TAG}.obj"
  cp "$WORK"/*.OBJ "$OUT"
  echo "OBJ -> $OUT"
fi
echo "workdir: $WORK"
