#!/usr/bin/env bash
# 把精訊 DQ3 的 18 軌音樂(已轉成的 MIDI)經真正的 MT-32 ROM 用 munt render → wav → OGG。
# 前置:① docker image munt-smf2wav 已 build(見 tools/Dockerfile.munt / docs/59);
#       ② work/music/mt32rom/ 內有 MT32_CONTROL.ROM + MT32_PCM.ROM(使用者自有 ROM,gitignore);
#       ③ work/music/export/midi/*.mid 已由 tools/cmf_to_midi.py 產生。
# 輸出:work/music/export/mt32/track_NN.ogg(gitignore,個人保存,不入庫散布)。
set -e
cd "$(dirname "$0")/.."
M="$PWD/work/music"
OUT="$M/export/mt32"
mkdir -p "$OUT" "$M/_tmp"

if ! docker image inspect munt-smf2wav >/dev/null 2>&1; then
  echo "缺 image munt-smf2wav,請先:docker build -f tools/Dockerfile.munt -t munt-smf2wav tools/"; exit 1
fi

echo "[1/2] MT-32 ROM render(munt)18 軌 → wav…"
for i in $(seq -w 0 17); do
  mid="$M/export/midi/$i.mid"
  [ -f "$mid" ] || continue
  docker run --rm -v "$M":/m munt-smf2wav -m /m/mt32rom -i mt32 -f -o "/m/_tmp/mt32_$i.wav" "/m/export/midi/$i.mid" >/dev/null 2>&1 || echo "  軌 $i render 失敗"
done

echo "[2/2] 轉 OGG(q5 VBR)…"
for i in $(seq -w 0 17); do
  w="$M/_tmp/mt32_$i.wav"
  [ -f "$w" ] && ffmpeg -y -loglevel error -i "$w" -c:a libvorbis -q:a 5 "$OUT/track_$i.ogg"
done
rm -f "$M"/_tmp/mt32_*.wav; rmdir "$M/_tmp" 2>/dev/null || true
echo "完成:MT-32 → $OUT/ ($(ls "$OUT"/*.ogg 2>/dev/null | wc -l) 軌)"
