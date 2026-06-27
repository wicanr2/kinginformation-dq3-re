#!/usr/bin/env bash
# 精訊 DQ3《傳說的終章》全曲紀念長片(MT-32 版):18 首音樂完整各播一次,
# 畫面用各場景片段一路循環穿插切換,每首前加一張曲序卡片。約 14 分鐘。
# 配樂用 work/music/export/mt32/ 的 OGG(真 MT-32 ROM 經 munt render)。個人紀念,輸出 work/video/(gitignore)。
# 前置:已擷取的場景幀在 work/rec/(record_memento.sh 跑過會留);沒有則先跑那支。需 host ffmpeg。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"; cd "$ROOT"
REC=work/rec; OUT=work/video; OGG=work/music/export/mt32
FONT="${DQ3_DEMO_FONT:-/usr/share/fonts/opentype/noto/NotoSansCJK-Black.ttc}"
[ -f "$FONT" ] || { echo "缺中文字型 $FONT"; exit 1; }
command -v ffmpeg >/dev/null || { echo "缺 host ffmpeg"; exit 1; }
[ -d "$OGG" ] || { echo "缺 $OGG(先跑 tools/export_music_mt32.sh)"; exit 1; }
BG="0x0a0e1a"; GOLD="0xf0c000"; CAP="0xf5d020"; WHITE="0xe8e8e8"
W=1280; H=720; FPS=25
mkdir -p "$OUT"; TMP="$(mktemp -d)"; trap 'rm -rf "$TMP"' EXIT
esc(){ printf '%s' "$1" | sed "s/:/\\\\:/g; s/'/\\\\'/g"; }

# 場景片段池(各 ~3.5s,循環穿插用);依序:標題/地表/夜/戰鬥/倉庫番/建城前/建城後/父親/九頭龍/索瑪
SCENES=(title field night battle sokoban townpre townpost hydra father zoma)
TAGS=("原版標題" "地表(白天)" "走入黑夜" "指令戰鬥" "倉庫番推石" "建城前·空草原" "建城後·繁榮" "勇者之父 歐里狄加" "五頭龍大王" "大魔王 索瑪")
echo "== 建場景片段池 =="
: > "$TMP/pool.txt"
for idx in "${!SCENES[@]}"; do
  d="$REC/${SCENES[$idx]}"; n=$(ls "$d"/*.ppm 2>/dev/null | wc -l); [ "$n" -gt 0 ] || continue
  ifr=$(awk "BEGIN{r=$n/3.5; if(r<1)r=1; print r}")
  tag=$(esc "${TAGS[$idx]}")
  ffmpeg -y -loglevel error -framerate "$ifr" -i "$d/f%06d.ppm" \
    -vf "scale=iw*2:ih*2:flags=neighbor,scale=${W}:${H}:force_original_aspect_ratio=decrease:flags=neighbor,pad=${W}:${H}:-1:-1:color=black,\
drawtext=fontfile=${FONT}:text='${tag}':fontcolor=${CAP}:fontsize=26:x=24:y=h-46:box=1:boxcolor=0x000000aa:boxborderw=10,\
fps=${FPS},format=yuv420p" -t 3.5 -an -c:v libx264 -pix_fmt yuv420p -r ${FPS} "$TMP/p$idx.mp4"
  echo "file '$TMP/p$idx.mp4'" >> "$TMP/pool.txt"
done
ffmpeg -y -loglevel error -f concat -safe 0 -i "$TMP/pool.txt" -c:v libx264 -pix_fmt yuv420p -crf 22 "$TMP/pool.mp4"
POOLD=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$TMP/pool.mp4")
echo "   池長 ${POOLD}s"

# 曲序卡片標籤(g_scene_track 推測角色;未知留空)
declare -A ROLE=( [0]="地表" [2]="城鎮" [3]="城堡" [5]="地城" [6]="戰鬥" [9]="船" [14]="魔王決戰" [16]="結局" [17]="標題" )

card(){ local out=$1 dur=$2 ti; ti=$(esc "$3"); local sb; sb=$(esc "$4"); local fo; fo=$(awk "BEGIN{print $dur-0.4}")
  ffmpeg -y -loglevel error -f lavfi -i "color=c=${BG}:s=${W}x${H}:r=${FPS}:d=${dur}" \
    -vf "drawtext=fontfile=${FONT}:text='${ti}':fontcolor=${GOLD}:fontsize=64:x=(w-tw)/2:y=h/2-60,\
drawtext=fontfile=${FONT}:text='${sb}':fontcolor=${WHITE}:fontsize=30:x=(w-tw)/2:y=h/2+30,\
fade=t=in:st=0:d=0.3,fade=t=out:st=${fo}:d=0.4" -c:v libx264 -pix_fmt yuv420p -r ${FPS} "$out"; }

echo "== 逐軌組片(card + 場景循環填到該軌長度)=="
LIST="$TMP/list.txt"; : > "$LIST"; : > "$TMP/audio.txt"
CARD=2.5
for i in $(seq -w 0 17); do
  ii=$((10#$i)); f="$OGG/track_$i.ogg"; [ -f "$f" ] || continue
  dur=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$f")
  role="${ROLE[$ii]:-}"; sub="樂曲 $((ii+1)) / 18"; [ -n "$role" ] && sub="$sub　·　$role"
  card "$TMP/card$i.mp4" "$CARD" "♪ $(printf '%02d' $((ii+1))) / 18" "$sub"
  filld=$(awk "BEGIN{d=$dur-$CARD; if(d<1)d=1; print d}")
  ffmpeg -y -loglevel error -stream_loop -1 -i "$TMP/pool.mp4" -t "$filld" -an -c:v libx264 -pix_fmt yuv420p -r ${FPS} "$TMP/fill$i.mp4"
  echo "file '$TMP/card$i.mp4'" >> "$LIST"; echo "file '$TMP/fill$i.mp4'" >> "$LIST"
  echo "file '$ROOT/$f'" >> "$TMP/audio.txt"
  printf "   軌%s %.0fs\n" "$i" "$dur"
done

echo "== concat 影像 + 鋪 18 軌 OGG =="
ffmpeg -y -loglevel error -f concat -safe 0 -i "$LIST" -c:v libx264 -pix_fmt yuv420p -crf 21 "$TMP/silent.mp4"
ffmpeg -y -loglevel error -f concat -safe 0 -i "$TMP/audio.txt" -c:a libvorbis -q:a 5 "$TMP/all.ogg"
VD=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$TMP/silent.mp4")
ffmpeg -y -loglevel error -i "$TMP/silent.mp4" -i "$TMP/all.ogg" \
  -filter_complex "[1:a]afade=t=in:st=0:d=1.5,afade=t=out:st=$(awk "BEGIN{print $VD-3}"):d=3[a]" \
  -map 0:v -map "[a]" -c:v copy -c:a aac -b:a 192k -shortest -movflags +faststart "$OUT/dq3_fullmusic_mt32.mp4"
D2=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$OUT/dq3_fullmusic_mt32.mp4" 2>/dev/null)
ls -lh "$OUT/dq3_fullmusic_mt32.mp4" | awk -v d="$D2" '{print "✅ ->",$9,"("$5", "d"s ≈ "int(d/60)"分)"}'
