#!/usr/bin/env bash
# 將已驗證的現行三平台包與推廣片集中到唯一交付目錄。
# 這是 Docker-only 的檔案收集器；不重新編譯，也不把原版素材加入 Git。
set -euo pipefail

if [ ! -f /.dockerenv ]; then
  echo "錯誤：交付物收集只能在 Docker 容器內執行。" >&2
  exit 2
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

RELEASE="${DQ3_RELEASE:-v0.1.34}"
SOURCE_RELEASE="${DQ3_SOURCE_RELEASE:-dist/$RELEASE}"
DEST="${DQ3_DIST_ALL_ROOT:-dist-all/$RELEASE}"
PROMO_SOURCE="${DQ3_PROMO_SOURCE:-$DEST/promo/dq3-remake-promo-20260812-r2.mp4}"
PROMO_TARGET="$DEST/promo/$(basename "$PROMO_SOURCE")"

case "$DEST" in
  dist-all/*) ;;
  *) echo "錯誤：交付目錄必須位於 dist-all/ 下：$DEST" >&2; exit 2 ;;
esac

for tree in patch full; do
  test -d "$SOURCE_RELEASE/$tree" || {
    echo "錯誤：缺少已驗證的 $SOURCE_RELEASE/$tree。" >&2
    exit 2
  }
done
test -s "$PROMO_SOURCE" || {
  echo "錯誤：缺少已驗證推廣片 $PROMO_SOURCE。" >&2
  exit 2
}

mkdir -p "$DEST/promo"
# 只重建兩個封裝樹，保留同一版 promo/source 的實際錄影輸入。
rm -rf "$DEST/patch" "$DEST/full"
mkdir -p "$DEST/patch" "$DEST/full"
cp -a "$SOURCE_RELEASE/patch/." "$DEST/patch/"
cp -a "$SOURCE_RELEASE/full/." "$DEST/full/"
# appimagetool log 是建置診斷，不是使用者交付物；只保留 AppImage 本身。
find "$DEST/patch" "$DEST/full" -type f -name '*.AppImage.appimagetool.log' -delete
if [ "$PROMO_SOURCE" != "$PROMO_TARGET" ]; then
  cp -p "$PROMO_SOURCE" "$PROMO_TARGET"
fi

MANIFEST="$DEST/SHA256SUMS.txt"
(
  cd "$DEST"
  find patch full promo -type f ! -name SHA256SUMS.txt ! -name RELEASE.txt -print \
    | sort \
    | while IFS= read -r file; do sha256sum "$file"; done
) > "$MANIFEST"

cat > "$DEST/RELEASE.txt" <<EOF
DQ3 Go／Ebitengine release: $RELEASE
現行交付目錄：dist-all/$RELEASE
三平台：Linux x86_64 AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP
推廣片：promo/$(basename "$PROMO_TARGET")
封裝來源：$SOURCE_RELEASE
驗收雜湊：SHA256SUMS.txt
EOF

test -s "$MANIFEST"
test -s "$DEST/RELEASE.txt"
echo "完成：$DEST"
echo "封裝檔案數：$(find "$DEST/patch" "$DEST/full" -type f | wc -l)"
echo "推廣片：$PROMO_TARGET"
