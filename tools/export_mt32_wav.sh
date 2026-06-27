#!/usr/bin/env bash
# 產生 remake「MT-32 音源」用的遊戲資產:18 軌 MIDI 經真 MT-32 ROM(munt)render → 22050Hz mono WAV。
# remake 內建極簡 WAV 讀取 + 串流(零解碼依賴、跨平台);放在 <assets>/mt32/ 即啟用設定的「音源:MT-32」。
# 前置:munt-smf2wav image(見 docs/59)+ work/music/mt32rom/ ROM + work/music/export/midi/*.mid。
# 輸出:work/mt32/track_NN.wav(gitignore)。把整個 mt32/ 夾複製到遊戲 assets 目錄即生效。
set -e
cd "$(dirname "$0")/.."
M="$PWD/work/music"; OUT="$PWD/work/mt32"
mkdir -p "$OUT" "$M/_tmp"
docker image inspect munt-smf2wav >/dev/null 2>&1 || { echo "缺 munt-smf2wav image(見 docs/59)"; exit 1; }

echo "[1/2] munt MT-32 render → wav…"
for i in $(seq -w 0 17); do
  [ -f "$M/export/midi/$i.mid" ] || continue
  docker run --rm -v "$M":/m munt-smf2wav -m /m/mt32rom -i mt32 -f -o "/m/_tmp/m_$i.wav" "/m/export/midi/$i.mid" >/dev/null 2>&1 || echo "  軌$i render 失敗"
done
echo "[2/2] 降到 22050Hz mono(遊戲串流用)…"
for i in $(seq -w 0 17); do
  [ -f "$M/_tmp/m_$i.wav" ] && ffmpeg -y -loglevel error -i "$M/_tmp/m_$i.wav" -ar 22050 -ac 1 -c:a pcm_s16le "$OUT/track_$i.wav"
done
rm -rf "$M/_tmp"
echo "完成:$(ls "$OUT"/*.wav 2>/dev/null|wc -l) 軌 → $OUT/  ($(du -sh "$OUT" 2>/dev/null|cut -f1))"
echo "→ 把 work/mt32/ 整夾複製到遊戲 assets 目錄(與 DQ3.EXE 同層),設定選單『音源』即可選 MT-32。"
