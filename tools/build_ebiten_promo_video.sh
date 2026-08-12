#!/usr/bin/env bash
# 由現行 Go／Ebitengine runtime 證據圖與本機 MT-32 render 組成推廣片。
#
# 本腳本必須在含 ffmpeg、ffprobe 與 Noto CJK 字型的 Docker image 內執行。
# 原始遊戲素材與本機音樂只作輸入；輸出位於 gitignored 的 work/video/。
set -euo pipefail

if [ ! -f /.dockerenv ]; then
  echo "錯誤：推廣片建置只能在 Docker 容器內執行。" >&2
  exit 2
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

command -v ffmpeg >/dev/null 2>&1 || { echo "錯誤：容器缺少 ffmpeg。" >&2; exit 2; }
command -v ffprobe >/dev/null 2>&1 || { echo "錯誤：容器缺少 ffprobe。" >&2; exit 2; }

FONT="${DQ3_PROMO_FONT:-/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc}"
OUT="${DQ3_PROMO_OUT:-work/video/dq3_remake_demo.mp4}"
LEGACY="${DQ3_PROMO_LEGACY:-work/video/dq3_remake_demo_20260628_legacy.mp4}"
AUDIO_ROOT="${DQ3_PROMO_AUDIO_ROOT:-work/music/export/mt32}"
OPENING_CAPTURE="${DQ3_PROMO_OPENING_CAPTURE:-work/video/dq3_ebiten_opening_runtime.mp4}"

test -f "$FONT" || { echo "錯誤：缺少繁中字型 $FONT。" >&2; exit 2; }

TRACK_TITLE="$AUDIO_ROOT/track_17.ogg"
TRACK_FIELD="$AUDIO_ROOT/track_00.ogg"
TRACK_BATTLE="$AUDIO_ROOT/track_06.ogg"
TRACK_BOSS="$AUDIO_ROOT/track_14.ogg"
TRACK_ENDING="$AUDIO_ROOT/track_16.ogg"

required_audio=(
  "$TRACK_TITLE" "$TRACK_FIELD" "$TRACK_BATTLE" "$TRACK_BOSS" "$TRACK_ENDING"
)
required_images=(
  dq3_remake_ebitan/docs/opening_home_rec82.png
  dq3_remake_ebitan/docs/opening_king_rec78.png
  dq3_remake_ebitan/docs/img/party_field_hud.png
  docs/img/teidon_dark_lamp_night.png
  docs/img/ship_first_sailing.png
  docs/img/eginbear_push_puzzle_solved.png
  docs/img/merchant_revolution_yellow_orb_obtained.png
  docs/img/jipang_orochi_first_battle.png
  docs/img/jipang_orochi_second_battle.png
  dq3_remake_ebitan/docs/baramos_battle.png
  docs/monsters/restored_128_129.png
  dq3_remake_ebitan/docs/phoenix_revived.png
  dq3_remake_ebitan/docs/zoma_final_battle.png
  dq3_remake_ebitan/docs/img/ending_the_end_runtime.png
)

for input in "${required_audio[@]}" "${required_images[@]}" "$OPENING_CAPTURE"; do
  test -s "$input" || { echo "錯誤：缺少推廣片輸入 $input。" >&2; exit 2; }
done

mkdir -p "$(dirname "$OUT")"
if [ -s "$OUT" ] && [ ! -e "$LEGACY" ]; then
  cp -p "$OUT" "$LEGACY"
fi

TMP="$(mktemp -d /tmp/dq3-promo-XXXXXX)"
cleanup() {
  rm -rf "$TMP"
}
trap cleanup EXIT

W=1280
H=720
FPS=30
SEGMENT_SECONDS=4
TOTAL_SECONDS=72
BG="0x070b15"
GOLD="0xf2c94c"
WHITE="0xf4f4f4"
CAPTION="0xf8d85a"

card() {
  local output="$1" duration="$2" title="$3" subtitle="$4"
  ffmpeg -y -loglevel error \
    -f lavfi -i "color=c=${BG}:s=${W}x${H}:r=${FPS}:d=${duration}" \
    -vf "drawtext=fontfile=${FONT}:text='${title}':fontcolor=${GOLD}:fontsize=62:x=(w-tw)/2:y=h/2-82,drawtext=fontfile=${FONT}:text='${subtitle}':fontcolor=${WHITE}:fontsize=30:x=(w-tw)/2:y=h/2+18,fade=t=in:st=0:d=0.35,fade=t=out:st=$(awk -v d="$duration" 'BEGIN { print d-0.45 }'):d=0.45" \
    -an -c:v libx264 -preset veryfast -crf 18 -pix_fmt yuv420p -r "$FPS" "$output"
}

scene() {
  local output="$1" image="$2" caption="$3"
  ffmpeg -y -loglevel error -loop 1 -t "$SEGMENT_SECONDS" -i "$image" \
    -vf "scale=${W}:700:force_original_aspect_ratio=decrease:flags=neighbor,pad=${W}:${H}:(ow-iw)/2:(oh-ih)/2:color=black,drawtext=fontfile=${FONT}:text='${caption}':fontcolor=${CAPTION}:fontsize=27:x=24:y=26:box=1:boxcolor=0x000000b8:boxborderw=10,format=yuv420p" \
    -an -c:v libx264 -preset veryfast -crf 18 -pix_fmt yuv420p -r "$FPS" "$output"
}

LIST="$TMP/video-list.txt"
: > "$LIST"
append_clip() {
  printf "file '%s'\n" "$1" >> "$LIST"
}

card "$TMP/00-title.mp4" 4 \
  "傳說的終章" \
  "精訊 1993 未發售作品 · Go／Ebitengine remake"
append_clip "$TMP/00-title.mp4"

ffmpeg -y -loglevel error -i "$OPENING_CAPTURE" -t 12 \
  -vf "scale=${W}:${H}:force_original_aspect_ratio=decrease:flags=neighbor,pad=${W}:${H}:(ow-iw)/2:(oh-ih)/2:color=black,drawtext=fontfile=${FONT}:text='現行 Go／Ebitengine 實際運行':fontcolor=${CAPTION}:fontsize=27:x=24:y=26:box=1:boxcolor=0x000000b8:boxborderw=10,format=yuv420p" \
  -an -c:v libx264 -preset veryfast -crf 18 -pix_fmt yuv420p -r "$FPS" "$TMP/01-runtime.mp4"
append_clip "$TMP/01-runtime.mp4"

scenes=(
  "dq3_remake_ebitan/docs/opening_home_rec82.png|母親帶勇者踏上旅程"
  "dq3_remake_ebitan/docs/opening_king_rec78.png|謁見國王，接受最初使命"
  "dq3_remake_ebitan/docs/img/party_field_hud.png|四人隊伍與原版地表 HUD"
  "docs/img/teidon_dark_lamp_night.png|日夜、城鎮與條件事件"
  "docs/img/ship_first_sailing.png|取得船隻，航向世界各地"
  "docs/img/eginbear_push_puzzle_solved.png|推石解謎與隱藏通路"
  "docs/img/merchant_revolution_yellow_orb_obtained.png|商人城發展與黃寶珠事件"
  "docs/img/jipang_orochi_first_battle.png|八頭大蛇第一戰 · 洞窟背景"
  "docs/img/jipang_orochi_second_battle.png|八頭大蛇第二戰 · 沙漠背景"
  "dq3_remake_ebitan/docs/baramos_battle.png|巴拉摩斯決戰"
  "docs/monsters/restored_128_129.png|補回原始檔遺失的 128／129 號怪物"
  "dq3_remake_ebitan/docs/phoenix_revived.png|六顆寶珠與不死鳥復活"
  "dq3_remake_ebitan/docs/zoma_final_battle.png|跨越兩個世界，迎戰大魔王索瑪"
)

index=1
for entry in "${scenes[@]}"; do
  image="${entry%%|*}"
  caption="${entry#*|}"
  clip="$TMP/$(printf '%02d' "$index")-scene.mp4"
  scene "$clip" "$image" "$caption"
  append_clip "$clip"
  index=$((index + 1))
done

scene "$TMP/99-ending.mp4" \
  "dq3_remake_ebitan/docs/img/ending_the_end_runtime.png" \
  "保存台灣 DOS RPG 的技術、文字與記憶"
append_clip "$TMP/99-ending.mp4"

ffmpeg -y -loglevel error -f concat -safe 0 -i "$LIST" \
  -an -c:v copy -movflags +faststart "$TMP/video.mp4"

# 片長 72 秒；以五首本機 MT-32 render 建立清楚的冒險弧線，四個接點各交叉淡化 1 秒。
ffmpeg -y -loglevel error \
  -i "$TRACK_TITLE" -i "$TRACK_FIELD" -i "$TRACK_BATTLE" -i "$TRACK_BOSS" -i "$TRACK_ENDING" \
  -filter_complex "[0:a]atrim=0:16,asetpts=PTS-STARTPTS[a0];[1:a]atrim=0:18,asetpts=PTS-STARTPTS[a1];[2:a]atrim=0:14,asetpts=PTS-STARTPTS[a2];[3:a]atrim=0:14,asetpts=PTS-STARTPTS[a3];[4:a]atrim=0:14,asetpts=PTS-STARTPTS[a4];[a0][a1]acrossfade=d=1:c1=tri:c2=tri[x1];[x1][a2]acrossfade=d=1:c1=tri:c2=tri[x2];[x2][a3]acrossfade=d=1:c1=tri:c2=tri[x3];[x3][a4]acrossfade=d=1:c1=tri:c2=tri,loudnorm=I=-16:TP=-1.5:LRA=11,aformat=sample_rates=48000:channel_layouts=stereo[a]" \
  -map "[a]" -t "$TOTAL_SECONDS" -c:a aac -profile:a aac_low -b:a 192k "$TMP/audio.m4a"

ffmpeg -y -loglevel error -i "$TMP/video.mp4" -i "$TMP/audio.m4a" \
  -map 0:v:0 -map 1:a:0 -c:v copy -c:a copy -t "$TOTAL_SECONDS" \
  -movflags +faststart "$TMP/final.mp4"

tools/verify_promo_video.sh "$TMP/final.mp4"
mv "$TMP/final.mp4" "$OUT"
sha256sum "$OUT"
echo "完成：$OUT"
