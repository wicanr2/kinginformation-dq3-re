#!/usr/bin/env bash
# 精訊 DQ3《傳說的終章》紀念影片:節錄各場景 → 兩支 MP4(SB FM 版 / MT-32 版配樂)。
# 配樂用 work/music/export/{sb,mt32}/ 的 OGG(精訊自己的曲子,SB=OPL2 FM、MT-32=真 ROM 經 munt)。
# 個人紀念用,輸出 work/video/(gitignore,不散布)。需 host ffmpeg + docker(dq3-remake)。
# 兩階段:(1) docker 錄各場景 ppm 幀;(2) host ffmpeg 組片 + 各鋪一套 OGG。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"; cd "$ROOT"
REC=work/rec; OUT=work/video; mkdir -p "$OUT"
FONT="${DQ3_DEMO_FONT:-/usr/share/fonts/opentype/noto/NotoSansCJK-Black.ttc}"
[ -f "$FONT" ] || { echo "缺中文字型 $FONT"; exit 1; }
command -v ffmpeg >/dev/null || { echo "host 需要 ffmpeg"; exit 1; }
BG="0x0a0e1a"; GOLD="0xf0c000"; CAP="0xf5d020"; WHITE="0xe8e8e8"
W=1280; H=720; FPS=25

# ---------- Phase 1:docker 內錄各場景幀 ----------
if [ "${SKIP_CAPTURE:-0}" != "1" ]; then
echo "== Phase 1:錄製實機畫面(docker / SDL dummy)=="
docker run --rm -v "$ROOT":/repo -v dq3build:/build dq3-remake bash -lc '
  set -e
  cmake -S /repo/dq3_remake -B /build -DCMAKE_BUILD_TYPE=Release >/dev/null 2>&1
  cmake --build /build -j >/dev/null 2>&1
  R=/repo/work/rec; rm -rf $R
  E="SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy"
  B=/build/dq3_remake; A=/repo/assets_raw
  mk(){ mkdir -p $R/$1; }
  mk title    && env $E DQ3_RECDIR=$R/title    DQ3_INPUT="............" $B $A title TITG.P >/dev/null 2>&1 || true
  mk field    && env $E DQ3_RECDIR=$R/field    DQ3_DEBUG="party;warp:1:88:88"      DQ3_INPUT="rrrrrrrrrruuuuuuuuuulllllllllldddddddddd" $B $A game >/dev/null 2>&1 || true
  mk night    && env $E DQ3_RECDIR=$R/night    DQ3_DEBUG="party;warp:1:88:88;dn:2" DQ3_INPUT="rrrrrrrruuuuuuuuddddddddllllllll" $B $A game >/dev/null 2>&1 || true
  mk battle   && env $E DQ3_RECDIR=$R/battle   DQ3_BATTLE_PARTY=1 DQ3_MON=5 DQ3_MON_N=3 DQ3_INPUT="..s..s..s..s..s..s..s..s..s..s" $B $A battle >/dev/null 2>&1 || true
  mk sokoban  && env $E DQ3_RECDIR=$R/sokoban  DQ3_DEBUG="warp:76:5:11:0"          DQ3_INPUT="u.u.l.l.u.u.r.r" $B $A game >/dev/null 2>&1 || true
  mk townbuild && env $E DQ3_RECDIR=$R/townbuild DQ3_DEBUG="party;merchant;warp:83:16:4:0" DQ3_INPUT=".u.u.e.e.e.e." $B $A game >/dev/null 2>&1 || true
  mk townbuilt && env $E DQ3_RECDIR=$R/townbuilt DQ3_DEBUG="party;flag:0x216;warp:83:16:5:0" DQ3_INPUT="d.d.r.r.l.l.u.u" $B $A game >/dev/null 2>&1 || true
  mk hydra    && env $E DQ3_RECDIR=$R/hydra    DQ3_BATTLE_PARTY=1 DQ3_ST_LEVEL=45 DQ3_MON=129 DQ3_MON_N=1 DQ3_INPUT="..s..s..s..s..s..s..s..s" $B $A battle >/dev/null 2>&1 || true
  mk father   && env $E DQ3_RECDIR=$R/father   DQ3_BATTLE_PARTY=1 DQ3_ST_LEVEL=45 DQ3_MON=128 DQ3_MON_N=1 DQ3_INPUT="..s..s..s..s..s..s" $B $A battle >/dev/null 2>&1 || true
  mk zoma     && env $E DQ3_RECDIR=$R/zoma DQ3_DUMP=/tmp/z.ppm DQ3_BATTLE_PARTY=1 DQ3_ST_LEVEL=50 DQ3_MON=124 DQ3_MON_N=1 DQ3_INPUT="..s..s..s..s..s..s..s..s..s..s" $B $A battle >/dev/null 2>&1 || true
  for d in title field night battle sokoban townbuild townbuilt hydra father zoma; do echo "   $d: $(ls $R/$d/*.ppm 2>/dev/null|wc -l) 幀"; done
'
fi

# ---------- Phase 2:host ffmpeg 組片(無聲)----------
echo "== Phase 2:組片(host ffmpeg)=="
TMP="$(mktemp -d)"; trap 'rm -rf "$TMP"' EXIT
esc(){ printf '%s' "$1" | sed "s/:/\\\\:/g; s/'/\\\\'/g"; }
card(){ local out=$1 dur=$2 ti; ti=$(esc "$3"); local sb; sb=$(esc "$4"); local fo; fo=$(awk "BEGIN{print $dur-0.5}")
  ffmpeg -y -loglevel error -f lavfi -i "color=c=${BG}:s=${W}x${H}:r=${FPS}:d=${dur}" \
    -vf "drawtext=fontfile=${FONT}:text='${ti}':fontcolor=${GOLD}:fontsize=56:x=(w-tw)/2:y=h/2-70,\
drawtext=fontfile=${FONT}:text='${sb}':fontcolor=${WHITE}:fontsize=31:x=(w-tw)/2:y=h/2+20,\
fade=t=in:st=0:d=0.4,fade=t=out:st=${fo}:d=0.5" -c:v libx264 -pix_fmt yuv420p -r ${FPS} "$out"; }
clip(){ local dir=$1 out=$2 dur=$3 tag; tag=$(esc "$4")
  local n; n=$(ls "$dir"/*.ppm 2>/dev/null | wc -l); [ "$n" -gt 0 ] || { echo "  (跳過空場景 $dir)"; return 1; }
  local ifr; ifr=$(awk "BEGIN{r=$n/$dur; if(r<1)r=1; print r}")
  local fo; fo=$(awk "BEGIN{print $dur-0.5}")
  ffmpeg -y -loglevel error -framerate "$ifr" -i "$dir/f%06d.ppm" \
    -vf "scale=iw*2:ih*2:flags=neighbor,scale=${W}:${H}:force_original_aspect_ratio=decrease:flags=neighbor,pad=${W}:${H}:-1:-1:color=black,\
drawtext=fontfile=${FONT}:text='${tag}':fontcolor=${CAP}:fontsize=26:x=24:y=h-46:box=1:boxcolor=0x000000aa:boxborderw=10,\
fps=${FPS},format=yuv420p,fade=t=in:st=0:d=0.4,fade=t=out:st=${fo}:d=0.5" \
    -an -c:v libx264 -pix_fmt yuv420p -r ${FPS} "$out"; }

LIST="$TMP/list.txt"; : > "$LIST"; add(){ echo "file '$1'" >> "$LIST"; }
card "$TMP/c00.mp4" 4.5 "Dragon Fighter III 傳說的終章" "精訊 1993 未發售版 · 反組譯 + C/SDL2 重製 · 紀念影片"; add "$TMP/c00.mp4"
clip "$REC/title"    "$TMP/s00.mp4" 3.0 "原版標題畫面"                       && add "$TMP/s00.mp4"
card "$TMP/c01.mp4" 3.0 "勇者啟程" "原版起點 · 步數驅動晝夜";               add "$TMP/c01.mp4"
clip "$REC/field"    "$TMP/s01.mp4" 4.5 "地表行走(白天)"                    && add "$TMP/s01.mp4"
clip "$REC/night"    "$TMP/s02.mp4" 3.6 "晝夜系統:走入黑夜"                 && add "$TMP/s02.mp4"
card "$TMP/c03.mp4" 3.0 "戰鬥" "傷害公式 · 怪物 AI/數值全 RE";              add "$TMP/c03.mp4"
clip "$REC/battle"   "$TMP/s03.mp4" 4.0 "指令戰鬥"                          && add "$TMP/s03.mp4"
card "$TMP/c04.mp4" 3.4 "倉庫番解謎" "推三顆大石開密道(當年解謎還原)";     add "$TMP/c04.mp4"
clip "$REC/sokoban"  "$TMP/s04.mp4" 4.0 "sokoban 推石"                      && add "$TMP/s04.mp4"
card "$TMP/c05.mp4" 3.4 "帶商人建城" "帶商人同伴 → 見老人 → 留下建城 → 繁榮"; add "$TMP/c05.mp4"
clip "$REC/townbuild" "$TMP/s05.mp4" 3.4 "帶商人見老人·觸發建城"           && add "$TMP/s05.mp4"
clip "$REC/townbuilt" "$TMP/s06.mp4" 3.0 "新城鎮建成:商店與居民"          && add "$TMP/s06.mp4"
card "$TMP/c07.mp4" 4.2 "當年缺 sprite 的決戰" "歐里狄加(父親)· 五頭龍大王 —— 缺圖曾使原版當機,今復原"; add "$TMP/c07.mp4"
clip "$REC/father"   "$TMP/s07.mp4" 3.4 "勇者之父 歐里狄加(怪128)"        && add "$TMP/s07.mp4"
clip "$REC/hydra"    "$TMP/s08.mp4" 4.0 "五頭龍大王(怪129)"               && add "$TMP/s08.mp4"
card "$TMP/c09.mp4" 3.6 "大魔王 索瑪" "光之珠驅散黑暗 → 二階段決戰";        add "$TMP/c09.mp4"
clip "$REC/zoma"     "$TMP/s09.mp4" 4.5 "最終決戰:大魔王索瑪(怪124)"      && add "$TMP/s09.mp4"
card "$TMP/c99.mp4" 5.5 "そして伝説へ… 傳說,於此展開" "未發售的經典 · 三十年後完整跑起來了"; add "$TMP/c99.mp4"

echo "   concat 影像…"
ffmpeg -y -loglevel error -f concat -safe 0 -i "$LIST" -c:v libx264 -pix_fmt yuv420p -crf 20 "$TMP/silent.mp4"
DUR=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$TMP/silent.mp4")
echo "   影片長度 ${DUR}s"

# ---------- Phase 3:各版本鋪 OGG 配樂 ----------
# 配樂選軌(隨影片弧線):地表(0)→ 戰鬥(6)→ boss(14)→ 結局(16)
TRACKS=(00 06 14 16)
bed(){ local ver=$1 dir="work/music/export/$1"
  [ -d "$dir" ] || { echo "  缺 $dir,跳過 $ver 版"; return; }
  local cl="$TMP/${ver}_st.txt"; : > "$cl"
  for t in "${TRACKS[@]}"; do [ -f "$dir/track_$t.ogg" ] && echo "file '$ROOT/$dir/track_$t.ogg'" >> "$cl"; done
  [ -s "$cl" ] || { echo "  $ver 無 OGG,跳過"; return; }
  ffmpeg -y -loglevel error -f concat -safe 0 -i "$cl" -c:a libvorbis -q:a 5 "$TMP/${ver}_st.ogg"
  local FO; FO=$(awk "BEGIN{print $DUR-3}")
  ffmpeg -y -loglevel error -i "$TMP/silent.mp4" -stream_loop -1 -i "$TMP/${ver}_st.ogg" \
    -filter_complex "[1:a]atrim=0:$DUR,afade=t=in:st=0:d=2,afade=t=out:st=$FO:d=3,volume=0.9[a]" \
    -map 0:v -map "[a]" -c:v copy -c:a aac -b:a 192k -shortest -movflags +faststart "$OUT/dq3_memento_${ver}.mp4"
  local D2; D2=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$OUT/dq3_memento_${ver}.mp4" 2>/dev/null)
  ls -lh "$OUT/dq3_memento_${ver}.mp4" | awk -v d="$D2" -v v="$ver" '{print "   ✅",v,"版 ->",$9,"("$5", "d"s)"}'
}
echo "== Phase 3:鋪配樂(兩版本)=="
bed mt32          # 只做 MT-32(SB 版已棄)
echo "完成 → $OUT/dq3_memento_mt32.mp4"
