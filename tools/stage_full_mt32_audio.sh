#!/usr/bin/env bash
# 將本機合法持有的 MT-32 預錄音軌與 ROM 加入「完整版」staging；不得用於公開 patch。
set -euo pipefail

if [ ! -f /.dockerenv ]; then
  echo "錯誤：MT-32 私有素材只能在 Docker 內整理。" >&2
  exit 2
fi

: "${FULL_ASSETS_DIR:?請指定完整版 staging 的 assets_raw 目錄}"
TRACK_DIR="${MT32_TRACK_DIR:-/repo/work/music/export/mt32}"
ROM_DIR="${MT32_ROM_DIR:-/repo/work/music/mt32rom}"

test -d "$FULL_ASSETS_DIR"
test -s "$ROM_DIR/MT32_CONTROL.ROM"
test -s "$ROM_DIR/MT32_PCM.ROM"
for track in $(seq -w 0 17); do
  test -s "$TRACK_DIR/track_${track}.ogg"
done

install -d "$FULL_ASSETS_DIR/mt32" "$FULL_ASSETS_DIR/mt32-rom"
for track in $(seq -w 0 17); do
  install -m 0644 "$TRACK_DIR/track_${track}.ogg" "$FULL_ASSETS_DIR/mt32/track_${track}.ogg"
done
install -m 0644 "$ROM_DIR/MT32_CONTROL.ROM" "$FULL_ASSETS_DIR/mt32-rom/MT32_CONTROL.ROM"
install -m 0644 "$ROM_DIR/MT32_PCM.ROM" "$FULL_ASSETS_DIR/mt32-rom/MT32_PCM.ROM"

(
  cd "$FULL_ASSETS_DIR"
  sha256sum mt32/track_*.ogg mt32-rom/MT32_CONTROL.ROM mt32-rom/MT32_PCM.ROM
) > "$FULL_ASSETS_DIR/MT32-PRIVATE-ASSETS.sha256"

echo "完成：已加入 MT-32 音軌與 ROM（僅限本機完整版）。"
