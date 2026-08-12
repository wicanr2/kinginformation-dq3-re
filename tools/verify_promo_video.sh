#!/usr/bin/env bash
# 推廣片失敗即關閉驗收：不能只檢查 audio stream 存在，還要證明內容不是數位靜音。
set -euo pipefail

if [ ! -f /.dockerenv ]; then
  echo "錯誤：推廣片驗收只能在 Docker 容器內執行。" >&2
  exit 2
fi

VIDEO="${1:-}"
test -s "$VIDEO" || { echo "錯誤：找不到影片 $VIDEO。" >&2; exit 2; }
command -v ffmpeg >/dev/null 2>&1 || { echo "錯誤：容器缺少 ffmpeg。" >&2; exit 2; }
command -v ffprobe >/dev/null 2>&1 || { echo "錯誤：容器缺少 ffprobe。" >&2; exit 2; }

video_streams="$(ffprobe -v error -select_streams v -show_entries stream=index -of csv=p=0 "$VIDEO" | wc -l)"
audio_streams="$(ffprobe -v error -select_streams a -show_entries stream=index -of csv=p=0 "$VIDEO" | wc -l)"
[ "$video_streams" -eq 1 ] || { echo "錯誤：預期一條影像串流，實際 $video_streams。" >&2; exit 1; }
[ "$audio_streams" -eq 1 ] || { echo "錯誤：預期一條音訊串流，實際 $audio_streams。" >&2; exit 1; }

width="$(ffprobe -v error -select_streams v:0 -show_entries stream=width -of csv=p=0 "$VIDEO")"
height="$(ffprobe -v error -select_streams v:0 -show_entries stream=height -of csv=p=0 "$VIDEO")"
audio_codec="$(ffprobe -v error -select_streams a:0 -show_entries stream=codec_name -of csv=p=0 "$VIDEO")"
sample_rate="$(ffprobe -v error -select_streams a:0 -show_entries stream=sample_rate -of csv=p=0 "$VIDEO")"
channels="$(ffprobe -v error -select_streams a:0 -show_entries stream=channels -of csv=p=0 "$VIDEO")"

[ "$width" = 1280 ] && [ "$height" = 720 ] || {
  echo "錯誤：畫面必須為 1280×720，實際 ${width}×${height}。" >&2
  exit 1
}
[ "$audio_codec" = aac ] || { echo "錯誤：音訊必須為 AAC，實際 $audio_codec。" >&2; exit 1; }
[ "$sample_rate" = 48000 ] || { echo "錯誤：音訊取樣率必須為 48000 Hz，實際 $sample_rate。" >&2; exit 1; }
[ "$channels" = 2 ] || { echo "錯誤：音訊必須為雙聲道，實際 $channels。" >&2; exit 1; }

volume_log="$(mktemp /tmp/dq3-promo-volume-XXXXXX.log)"
silence_log="$(mktemp /tmp/dq3-promo-silence-XXXXXX.log)"
cleanup() {
  rm -f "$volume_log" "$silence_log"
}
trap cleanup EXIT

ffmpeg -hide_banner -nostats -v info -i "$VIDEO" -map 0:a:0 -af volumedetect -f null - 2>"$volume_log"
mean_volume="$(sed -n 's/.*mean_volume: \([-0-9.]*\) dB.*/\1/p' "$volume_log" | tail -1)"
max_volume="$(sed -n 's/.*max_volume: \([-0-9.]*\) dB.*/\1/p' "$volume_log" | tail -1)"
test -n "$mean_volume" && test -n "$max_volume" || {
  echo "錯誤：無法量測音軌音量。" >&2
  exit 1
}
awk -v mean="$mean_volume" -v peak="$max_volume" 'BEGIN { exit !(mean > -35 && peak > -12) }' || {
  echo "錯誤：音軌近似靜音（mean=${mean_volume} dB，max=${max_volume} dB）。" >&2
  exit 1
}

ffmpeg -hide_banner -nostats -v info -i "$VIDEO" -map 0:a:0 \
  -af silencedetect=noise=-50dB:d=3 -f null - 2>"$silence_log"
long_silence="$(sed -n 's/.*silence_duration: \([0-9.]*\).*/\1/p' "$silence_log" \
  | awk '$1 >= 3 { print $1; exit }')"
if [ -n "$long_silence" ]; then
  echo "錯誤：音軌含超過三秒的連續數位靜音。" >&2
  exit 1
fi

# 完整解碼，排除封裝、時間戳或截斷錯誤。
ffmpeg -v error -i "$VIDEO" -map 0:v:0 -map 0:a:0 -f null -

duration="$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$VIDEO")"
echo "通過：${width}×${height}，${duration}s，AAC ${sample_rate}Hz ${channels}ch，mean=${mean_volume}dB，max=${max_volume}dB。"
