#!/usr/bin/env bash
# 錄製 dq3_remake 實機畫面 → 帶成果註解的 demo MP4。
# 風格對齊作者另一專案(indiana-jones-...-cht):深藍底 #0a0e1a + 金字 #f0c000,
# 每段「成果說明卡 → 實機片段」,淡入淡出 concat,1280x720/25fps。
#
# 兩階段、不污染 host:
#   (1) 錄製:docker(SDL dummy 驅動)跑 game/battle/title + DQ3_RECDIR 把每次 present 的畫面 dump 成 fNNNNNN.ppm。
#   (2) 組片:host ffmpeg(讀 ppm 序列 + drawtext 中文卡片;免 ImageMagick)。
# 產物:work/video/dq3_remake_demo.mp4
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"; cd "$ROOT"
REC=work/rec; OUT=work/video; mkdir -p "$OUT"
FONT="${DQ3_DEMO_FONT:-/usr/share/fonts/opentype/noto/NotoSansCJK-Black.ttc}"
[ -f "$FONT" ] || { echo "缺中文字型 $FONT(設 DQ3_DEMO_FONT 指定)"; exit 1; }
command -v ffmpeg >/dev/null || { echo "host 需要 ffmpeg"; exit 1; }
BG="0x0a0e1a"; GOLD="0xf0c000"; CAP="0xf5d020"; WHITE="0xe8e8e8"
W=1280; H=720; FPS=25

# ---------- Phase 1:docker 內錄各場景幀 ----------
echo "== Phase 1:錄製實機畫面(docker / SDL dummy)=="
docker run --rm -v "$ROOT":/repo -v dq3build:/build dq3-remake bash -lc '
  set -e
  cmake -S /repo/dq3_remake -B /build -DCMAKE_BUILD_TYPE=Release >/dev/null 2>&1
  cmake --build /build -j >/dev/null 2>&1
  R=/repo/work/rec; rm -rf $R
  E="SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy"
  mkdir -p $R/title    && env $E DQ3_RECDIR=$R/title    DQ3_INPUT="............" /build/dq3_remake /repo/assets_raw title TITG.P >/dev/null 2>&1 || true
  mkdir -p $R/field_day  && env $E DQ3_RECDIR=$R/field_day  DQ3_DEBUG="party;warp:1:88:88"      DQ3_INPUT="rrrrrrrrrruuuuuuuuuulllllllllldddddddddd" /build/dq3_remake /repo/assets_raw game >/dev/null 2>&1 || true
  mkdir -p $R/field_night && env $E DQ3_RECDIR=$R/field_night DQ3_DEBUG="party;warp:1:88:88;dn:2" DQ3_INPUT="rrrrrrrruuuuuuuuddddddddllllllll" /build/dq3_remake /repo/assets_raw game >/dev/null 2>&1 || true
  mkdir -p $R/cmdmenu  && env $E DQ3_RECDIR=$R/cmdmenu  DQ3_DEBUG="party;warp:1:88:88"       DQ3_INPUT="c..d..d..d..u..u..e.." /build/dq3_remake /repo/assets_raw game >/dev/null 2>&1 || true
  mkdir -p $R/equip    && env $E DQ3_RECDIR=$R/equip    DQ3_DEBUG="party;item:0x0b;item:0x21;item:0x32;item:0x39;equip" DQ3_INPUT="..dd..uu..ee..bb.." /build/dq3_remake /repo/assets_raw game >/dev/null 2>&1 || true
  mkdir -p $R/battle   && env $E DQ3_RECDIR=$R/battle   DQ3_BATTLE_PARTY=1 DQ3_MON=5 DQ3_MON_N=3 DQ3_INPUT="..s..s..s..s..s..s..s..s..s..s" /build/dq3_remake /repo/assets_raw battle >/dev/null 2>&1 || true
  for d in title field_day field_night cmdmenu equip battle; do echo "   $d: $(ls $R/$d/*.ppm 2>/dev/null|wc -l) 幀"; done
'

# ---------- Phase 2:host ffmpeg 組片 ----------
echo "== Phase 2:組片(host ffmpeg)=="
TMP="$(mktemp -d)"; trap 'rm -rf "$TMP"' EXIT
esc(){ printf '%s' "$1" | sed "s/:/\\\\:/g; s/'/\\\\'/g"; }

# 成果說明卡:深藍底 + 金色標題 + 白色副標,淡入淡出。card <out.mp4> <dur> <title> <sub>
card(){ local out=$1 dur=$2 ti; ti=$(esc "$3"); local sb; sb=$(esc "$4"); local fo; fo=$(awk "BEGIN{print $dur-0.5}")
  ffmpeg -y -loglevel error -f lavfi -i "color=c=${BG}:s=${W}x${H}:r=${FPS}:d=${dur}" \
    -vf "drawtext=fontfile=${FONT}:text='${ti}':fontcolor=${GOLD}:fontsize=58:x=(w-tw)/2:y=h/2-70,\
drawtext=fontfile=${FONT}:text='${sb}':fontcolor=${WHITE}:fontsize=32:x=(w-tw)/2:y=h/2+20,\
fade=t=in:st=0:d=0.4,fade=t=out:st=${fo}:d=0.5" \
    -c:v libx264 -pix_fmt yuv420p -r ${FPS} "$out"; }

# 實機片段:ppm 序列 → 整數 2x 放大(保留像素)→ pad 720p → 角落浮水印 → 淡入淡出。
# clip <scene_dir> <out.mp4> <target_dur> <tag>
clip(){ local dir=$1 out=$2 dur=$3 tag; tag=$(esc "$4")
  local n; n=$(ls "$dir"/*.ppm 2>/dev/null | wc -l); [ "$n" -gt 0 ] || { echo "  (跳過空場景 $dir)"; return 1; }
  local ifr; ifr=$(awk "BEGIN{r=$n/$dur; if(r<1)r=1; print r}")
  local fo; fo=$(awk "BEGIN{print $dur-0.5}")
  ffmpeg -y -loglevel error -framerate "$ifr" -i "$dir/f%06d.ppm" \
    -vf "scale=iw*2:ih*2:flags=neighbor,scale=${W}:${H}:force_original_aspect_ratio=decrease:flags=neighbor,pad=${W}:${H}:-1:-1:color=black,\
drawtext=fontfile=${FONT}:text='${tag}':fontcolor=${CAP}:fontsize=26:x=24:y=h-46:box=1:boxcolor=0x000000aa:boxborderw=10,\
fps=${FPS},format=yuv420p,fade=t=in:st=0:d=0.4,fade=t=out:st=${fo}:d=0.5" \
    -an -c:v libx264 -pix_fmt yuv420p -r ${FPS} "$out"; }

LIST="$TMP/list.txt"; : > "$LIST"
add(){ echo "file '$1'" >> "$LIST"; }

echo "   烘製卡片 + 片段…"
card "$TMP/c00.mp4" 4.0 "Dragon Fighter III 傳說的終章" "精訊 1993 未發售版 · 反組譯 + C/SDL2 跨平台重製"; add "$TMP/c00.mp4"
clip "$REC/title"      "$TMP/s01.mp4" 3.0 "原版標題畫面 TITG.P(直接解碼自原始素材)"          && add "$TMP/s01.mp4"
card "$TMP/c02.mp4" 3.2 "地表:真實 DQ3CON.MAP" "原版起點 · 真實遭遇區 · 步數驅動晝夜";        add "$TMP/c02.mp4"
clip "$REC/field_day"  "$TMP/s02.mp4" 4.5 "地表行走(白天)"                                    && add "$TMP/s02.mp4"
card "$TMP/c03.mp4" 3.2 "晝夜系統" "白天→黃昏→黑夜→黎明 · 步數驅動 · palette 調暗";            add "$TMP/c03.mp4"
clip "$REC/field_night" "$TMP/s03.mp4" 4.0 "地表行走(黑夜 · palette 調暗)"                    && add "$TMP/s03.mp4"
card "$TMP/c04.mp4" 3.2 "野外指令窗" "對話 · 咒文 · 狀況 · 道具 · 裝備 · 調查";                 add "$TMP/c04.mp4"
clip "$REC/cmdmenu"    "$TMP/s04.mp4" 4.0 "C 鍵叫出命令選單"                                   && add "$TMP/s04.mp4"
card "$TMP/c05.mp4" 3.4 "裝備四槽:武器 / 鎧 / 盾 / 兜" "從 ITEM.DAT byte4 高位反推部位(精訊版 4 槽)"; add "$TMP/c05.mp4"
clip "$REC/equip"      "$TMP/s05.mp4" 4.0 "每名隊員 4 裝備槽總覽"                              && add "$TMP/s05.mp4"
card "$TMP/c06.mp4" 3.4 "戰鬥系統" "傷害公式逐一對反組譯確認 · 怪物 AI/數值全 RE · rng 升級";   add "$TMP/c06.mp4"
clip "$REC/battle"     "$TMP/s06.mp4" 4.5 "指令戰鬥(勇者 + 戰士 vs 史萊姆)"                   && add "$TMP/s06.mp4"
card "$TMP/c99.mp4" 5.0 "未發售的經典 · 三十年後跑起來了" "7 bug 全修 · 主線可破關 · Linux / Windows 已打包 · game_tester 80/80"; add "$TMP/c99.mp4"

echo "   concat → MP4(加無聲音軌,平台相容)…"
ffmpeg -y -loglevel error -f concat -safe 0 -i "$LIST" \
  -f lavfi -i anullsrc=channel_layout=stereo:sample_rate=44100 \
  -c:v libx264 -pix_fmt yuv420p -profile:v high -level 4.0 -crf 20 \
  -c:a aac -shortest -movflags +faststart "$OUT/dq3_remake_demo.mp4"

DUR=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$OUT/dq3_remake_demo.mp4" 2>/dev/null)
ls -lh "$OUT/dq3_remake_demo.mp4" | awk -v d="$DUR" '{print "== ✅ demo ->",$9,"("$5", "d"s) =="}'
