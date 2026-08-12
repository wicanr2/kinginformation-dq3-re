#!/usr/bin/env bash
# 只封裝現行 Go/Ebitengine Linux binary；不得用於歷史 C/SDL 產品。
set -euo pipefail

: "${APPDIR:?請指定已完成的 AppDir}"
: "${OUTPUT:?請指定 AppImage 輸出路徑}"

APPIMAGETOOL=${APPIMAGETOOL:-/repo/work/.tools/appimagetool}
RUNTIME_FILE=${RUNTIME_FILE:-/repo/work/.tools/runtime-x86_64}
ARCH=${ARCH:-x86_64}

test -d "$APPDIR"
test -x "$APPDIR/AppRun"
test -x "$APPDIR/usr/bin/dq3-remake"
test -x "$APPIMAGETOOL"
test -s "$RUNTIME_FILE"
mkdir -p "$(dirname "$OUTPUT")"
rm -f "$OUTPUT"

log="${OUTPUT}.appimagetool.log"
rm -f "$log"
if ARCH="$ARCH" "$APPIMAGETOOL" --runtime-file "$RUNTIME_FILE" \
  "$APPDIR" "$OUTPUT" >"$log" 2>&1; then
	:
else
	status=$?
	tail -80 "$log" >&2 || true
	exit "$status"
fi
test -s "$OUTPUT"
chmod 0755 "$OUTPUT"
printf 'AppImage 已建立：%s（%s bytes）\n' "$OUTPUT" "$(stat -c %s "$OUTPUT")"
