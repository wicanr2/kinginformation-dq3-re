#!/usr/bin/env bash
# 在 docker(DOSBox + Microsoft C 5.1)內把單一 .c 編成 .obj,dump 回 host。
# 用法: tools/build/msc_compile.sh <src.c> [cl_flags...]
#   預設 flags = /c /AS /Ox  (compile-only, small model, full optimize)。
#   模型: /AS small /AM medium /AC compact /AL large /AH huge。
# 產出: <srcdir>/<name>.<model>.msc.obj
set -euo pipefail
SRC="${1:?need src.c}"; shift || true
SRCDIR="$(cd "$(dirname "$SRC")" && pwd)"
SRCBASE="$(basename "$SRC")"
NAME="${SRCBASE%.*}"
FLAGS="${*:-/c /AS /Ox}"
TAG="$(echo "$FLAGS" | grep -oiE '/A[SMCLH]' | head -1 | tr -d '/' | tr 'A-Z' 'a-z' || echo as)"

WORK="$(mktemp -d)"
cp "$SRCDIR/$SRCBASE" "$WORK/"
cat > "$WORK/go.bat" <<EOF
@echo off
set PATH=C:\\BIN
set LIB=C:\\LIB
set INCLUDE=C:\\INCLUDE
set TMP=D:\\
d:
C:\\BIN\\CL.EXE $FLAGS D:\\$SRCBASE > D:\\CL.OUT 2>&1
EOF

docker run --rm \
  -e SDL_VIDEODRIVER=dummy -e SDL_AUDIODRIVER=dummy \
  -v "$WORK":/work dq3-msc \
  bash -lc '
    set -e
    cat > /work/dosbox.conf <<CFG
[sdl]
output=surface
[dosbox]
machine=svga_s3
[autoexec]
mount c /msc -t dir
mount d /work
d:
go.bat
exit
CFG
    timeout 120 dosbox -conf /work/dosbox.conf -exit >/work/dosbox.out 2>&1 || true
  '

echo "=== CL output ==="
cat "$WORK/CL.OUT" 2>/dev/null || true
echo "=== outputs ==="
ls -la "$WORK"
if ls "$WORK"/*.OBJ >/dev/null 2>&1; then
  OUT="$SRCDIR/${NAME}.${TAG}.msc.obj"
  cp "$WORK"/*.OBJ "$OUT"
  echo "OBJ -> $OUT"
fi
echo "workdir: $WORK"
