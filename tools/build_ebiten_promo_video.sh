#!/usr/bin/env bash
# 以現行 Go／Ebitengine runtime 與本機實際 MT-32 render 製作推廣片。
# 版型遵循 game-promo-video-ffmpeg：靜態畫面採固定畫格＋淡入淡出，避免
# zoompan 造成無界 frame explosion；推廣片只在 Docker 內產生並失敗即關閉。
set -euo pipefail

if [ ! -f /.dockerenv ]; then
  echo "錯誤：推廣片建置只能在 Docker 容器內執行。" >&2
  exit 2
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

for command_name in ffmpeg ffprobe convert montage awk sha256sum; do
  command -v "$command_name" >/dev/null 2>&1 || {
    echo "錯誤：容器缺少 $command_name。" >&2
    exit 2
  }
done

SERIF_FONT="${DQ3_PROMO_SERIF_FONT:-/usr/share/fonts/opentype/noto/NotoSerifCJK-Bold.ttc}"
SANS_FONT="${DQ3_PROMO_FONT:-/usr/share/fonts/opentype/noto/NotoSansCJK-Bold.ttc}"
OUT="${DQ3_PROMO_OUT:-dist-all/v0.1.34/promo/dq3-remake-promo-20260812-r2.mp4}"
AUDIO_ROOT="${DQ3_PROMO_AUDIO_ROOT:-work/music/export/mt32}"
OPENING_CAPTURE="${DQ3_PROMO_OPENING_CAPTURE:-dist-all/v0.1.34/promo/source/opening_runtime.mp4}"

test -f "$SERIF_FONT" || { echo "錯誤：缺少繁中字型 $SERIF_FONT。" >&2; exit 2; }
test -f "$SANS_FONT" || { echo "錯誤：缺少繁中字型 $SANS_FONT。" >&2; exit 2; }

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
TMP="$(mktemp -d /tmp/dq3-promo-r2-XXXXXX)"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT INT TERM

W=1280
H=720
FPS=30
SEGMENT_SECONDS=4
TOTAL_SECONDS=72
# 從現行 runtime 畫面抽出的夜藍、海綠、金黃三色主題；不改動原始畫格。
BG="#07111f"
BG2="#1c3850"
GOLD="#f2c94c"
WHITE="#f4f4f4"
CAPTION="#f8d85a"

card_png() {
  local output="$1" title="$2" subtitle="$3"
  convert -size "${W}x${H}" "gradient:${BG}-${BG2}" \
    -fill "rgba(7,17,31,0.60)" -draw "rectangle 110,150 1170,570" \
    -stroke "$GOLD" -strokewidth 3 -fill none -draw "rectangle 110,150 1170,570" \
    -gravity center -fill "$GOLD" -font "$SERIF_FONT" -pointsize 66 \
    -annotate +0-55 "$title" \
    -fill "$WHITE" -font "$SANS_FONT" -pointsize 30 -annotate +0+45 "$subtitle" \
    -fill "$CAPTION" -font "$SANS_FONT" -pointsize 20 -annotate +0+132 "精訊版／繁體中文／Go・Ebitengine" \
    "$output"
}

scene_png() {
  local output="$1" image="$2" caption="$3" layout="${4:-boxed}" image2="${5:-}"
  case "$layout" in
    full)
      convert "$image" -resize "${W}x${H}^" -gravity center -extent "${W}x${H}" \
        -fill "rgba(0,0,0,0.70)" -draw "rectangle 0,610 ${W},${H}" \
        -gravity south -fill "$CAPTION" -font "$SERIF_FONT" -pointsize 28 \
        -annotate +0+28 "$caption" "$output"
      ;;
    dcard)
      convert -size "${W}x${H}" "gradient:${BG2}-${BG}" \
        \( "$image" -resize "1080x600>" -bordercolor "$GOLD" -border 4 \) \
        -gravity center -geometry +0-34 -composite \
        -gravity south -fill "$CAPTION" -font "$SERIF_FONT" -pointsize 28 \
        -annotate +0+22 "$caption" "$output"
      ;;
    split)
      test -s "$image2" || { echo "錯誤：雙畫面版型缺少第二張圖。" >&2; exit 2; }
      montage "$image" "$image2" -tile 2x1 -geometry 620x500+8+8 \
        -background "$BG" "$TMP/split.png"
      convert -size "${W}x${H}" "gradient:${BG}-${BG2}" \
        \( "$TMP/split.png" -bordercolor "$GOLD" -border 4 \) \
        -gravity center -geometry +0-30 -composite \
        -gravity south -fill "$CAPTION" -font "$SERIF_FONT" -pointsize 27 \
        -annotate +0+24 "$caption" "$output"
      ;;
    boxed)
      convert -size "${W}x${H}" "gradient:${BG}-${BG2}" \
        \( "$image" -resize "960x560>" -bordercolor "$GOLD" -border 4 \) \
        -gravity center -geometry +0-34 -composite \
        -gravity south -fill "$CAPTION" -font "$SERIF_FONT" -pointsize 28 \
        -annotate +0+22 "$caption" "$output"
      ;;
    *)
      echo "錯誤：未知推廣片版型 $layout。" >&2
      exit 2
      ;;
  esac
}

static_video() {
  local output="$1" image="$2" duration="$3"
  ffmpeg -y -loglevel error -loop 1 -framerate "$FPS" -i "$image" -t "$duration" \
    -vf "fade=t=in:st=0:d=0.25,fade=t=out:st=$(awk -v d="$duration" 'BEGIN { print d-0.35 }'):d=0.35,format=yuv420p" \
    -an -c:v libx264 -preset veryfast -threads 2 -crf 18 -pix_fmt yuv420p -r "$FPS" "$output"
}

LIST="$TMP/video-list.txt"
: > "$LIST"
append_clip() { printf "file '%s'\n" "$1" >> "$LIST"; }

card_png "$TMP/00-title.png" "傳說的終章" "精訊版 DQ3：一段台灣 DOS RPG 的保存與重生"
static_video "$TMP/00-title.mp4" "$TMP/00-title.png" 4
append_clip "$TMP/00-title.mp4"

# 這段是目前 v0.1.34 AppRun 在 Xvfb 中以正式輸入擷取的實機畫面。
ffmpeg -y -loglevel error -i "$OPENING_CAPTURE" -t 12 \
  -vf "scale=${W}:${H}:force_original_aspect_ratio=decrease:flags=neighbor,pad=${W}:${H}:(ow-iw)/2:(oh-ih)/2:color=black,drawtext=fontfile=${SANS_FONT}:text='現行 Go／Ebitengine 實際運行':fontcolor=${CAPTION}:fontsize=27:x=24:y=26:box=1:boxcolor=0x000000b8:boxborderw=10,format=yuv420p" \
  -an -c:v libx264 -preset veryfast -threads 2 -crf 18 -pix_fmt yuv420p -r "$FPS" "$TMP/01-runtime.mp4"
append_clip "$TMP/01-runtime.mp4"

# 每種主要版型至少出現一次；雙畫面段落把八頭大蛇兩個已核對背景並列。
scenes=(
  "dq3_remake_ebitan/docs/opening_home_rec82.png|母親帶勇者踏上旅程|boxed|"
  "dq3_remake_ebitan/docs/opening_king_rec78.png|謁見國王，接受最初使命|dcard|"
  "dq3_remake_ebitan/docs/img/party_field_hud.png|四人隊伍與原版地表 HUD|full|"
  "docs/img/teidon_dark_lamp_night.png|日夜、城鎮與條件事件|boxed|"
  "docs/img/ship_first_sailing.png|取得船隻，航向世界各地|full|"
  "docs/img/eginbear_push_puzzle_solved.png|推石解謎與隱藏通路|dcard|"
  "docs/img/merchant_revolution_yellow_orb_obtained.png|商人城發展與黃寶珠事件|boxed|"
  "docs/img/jipang_orochi_first_battle.png|八頭大蛇：洞窟 → 沙漠|split|docs/img/jipang_orochi_second_battle.png"
  "docs/img/jipang_orochi_second_battle.png|八頭大蛇第二戰：沙漠背景|full|"
  "dq3_remake_ebitan/docs/baramos_battle.png|巴拉摩斯決戰|boxed|"
  "docs/monsters/restored_128_129.png|補回原始檔遺失的 128／129 號怪物|full|"
  "dq3_remake_ebitan/docs/phoenix_revived.png|六顆寶珠與不死鳥復活|dcard|"
  "dq3_remake_ebitan/docs/zoma_final_battle.png|跨越兩個世界，迎戰大魔王索瑪|full|"
)
index=1
for entry in "${scenes[@]}"; do
  IFS='|' read -r image caption layout image2 <<< "$entry"
  png="$TMP/$(printf '%02d' "$index")-scene.png"
  clip="$TMP/$(printf '%02d' "$index")-scene.mp4"
  scene_png "$png" "$image" "$caption" "$layout" "$image2"
  static_video "$clip" "$png" "$SEGMENT_SECONDS"
  append_clip "$clip"
  index=$((index + 1))
done

scene_png "$TMP/99-ending.png" \
  dq3_remake_ebitan/docs/img/ending_the_end_runtime.png \
  "保存台灣 DOS RPG 的技術、文字與記憶" full
static_video "$TMP/99-ending.mp4" "$TMP/99-ending.png" "$SEGMENT_SECONDS"
append_clip "$TMP/99-ending.mp4"

ffmpeg -y -loglevel error -f concat -safe 0 -i "$LIST" \
  -an -c:v copy -movflags +faststart "$TMP/video.mp4"

# 五首本機 MT-32 render 形成 72 秒冒險弧線；跨段各淡化一秒，避免數位靜音。
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
