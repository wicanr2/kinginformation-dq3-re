#!/usr/bin/env bash
# DQ3 正式場景配樂的最小充分抽樣閘門。必須在含 ffmpeg／ffprobe 的 Docker 內執行。
set -euo pipefail

audio_root="${1:-dq3_remake_ebitan/mobile/assets/mt32}"
tracks=(00 01 02 03 06 14 17)
tmp_dir="$(mktemp -d /tmp/dq3-audio-sample-XXXXXX)"
trap 'rm -rf "$tmp_dir"' EXIT

printf 'track\tcodec\trate\tchannels\tduration_s\tmean_db\tpeak_db\tmax_silence_s\tsha256\n'
for track in "${tracks[@]}"; do
	file="$audio_root/track_${track}.ogg"
	test -s "$file" || { echo "錯誤：缺少或空白：$file" >&2; exit 1; }

	probe="$(ffprobe -v error -select_streams a:0 \
		-show_entries stream=codec_name,sample_rate,channels,duration \
		-of csv=p=0 "$file")"
	IFS=, read -r codec rate channels duration <<<"$probe"
	[[ "$codec" == vorbis && "$rate" == 32000 && "$channels" == 2 ]] || {
		echo "錯誤：track_${track} 格式為 $codec/$rate Hz/${channels}ch" >&2
		exit 1
	}

	volume_log="$tmp_dir/volume_${track}.log"
	silence_log="$tmp_dir/silence_${track}.log"
	ffmpeg -hide_banner -nostats -v info -i "$file" -af volumedetect -f null - 2>"$volume_log"
	ffmpeg -hide_banner -nostats -v info -i "$file" \
		-af silencedetect=noise=-50dB:d=3 -f null - 2>"$silence_log"
	mean="$(sed -n 's/.*mean_volume: \([-0-9.]*\) dB.*/\1/p' "$volume_log" | tail -1)"
	peak="$(sed -n 's/.*max_volume: \([-0-9.]*\) dB.*/\1/p' "$volume_log" | tail -1)"
	max_silence="$(sed -n 's/.*silence_duration: \([0-9.]*\).*/\1/p' "$silence_log" | sort -nr | head -1)"
	max_silence="${max_silence:-0}"
	test -n "$mean" && test -n "$peak"
	awk -v mean="$mean" -v peak="$peak" -v silence="$max_silence" 'BEGIN {
		exit !(mean >= -30 && mean <= -12 && peak >= -12 && peak <= -1 && silence <= 4)
	}' || {
		echo "錯誤：track_${track} 音量／靜音超界：mean=${mean}, peak=${peak}, silence=${max_silence}" >&2
		exit 1
	}
	ffmpeg -v error -i "$file" -map 0:a:0 -f null -
	hash="$(sha256sum "$file" | awk '{print $1}')"
	printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
		"$track" "$codec" "$rate" "$channels" "$duration" "$mean" "$peak" "$max_silence" "$hash"
done

