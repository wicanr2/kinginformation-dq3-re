#!/usr/bin/env bash
# 把精訊 DQ3 的 18 軌音樂(MBG.MCX,使用者正版遊戲自身資料)離線 render 成兩個版本的 OGG,
# 供使用者私人保存 + 台灣電玩史紀錄。輸出 work/music/export/{sb,mt32}/(gitignore,不入庫散布)。
#   sb   = 精訊原生聲霸卡 OPL2 FM(我們的 dq3_opl2 合成器)
#   mt32 = 把事件流轉 MIDI、用 GM SoundFont 合成的 Roland/MT-32 風(精訊當年未做此音源)
# 需:docker(gcc / fluidsynth)+ host ffmpeg + work/music/sf2/TimGM6mb.sf2 + work/music/track_*.bin
set -e
cd "$(dirname "$0")/.."
EXP=work/music/export
SF=work/music/sf2/TimGM6mb.sf2
mkdir -p "$EXP/sb" "$EXP/mt32" work/music/_tmp

echo "[1/5] 編譯 opl_render(OPL2 FM)…"
docker run --rm -v "$PWD":/work -w /work gcc:13-bookworm \
  cc -O2 -I dq3_remake/src tools/opl_render.c dq3_remake/src/dq3_opl2.c dq3_remake/src/dq3_cmf.c -lm -o /work/work/music/_tmp/opl_render

echo "[2/5] SB:render 18 軌 wav…"
for i in $(seq -w 0 17); do
  work/music/_tmp/opl_render "work/music/track_$i.bin" "work/music/_tmp/sb_$i.wav" 0 >/dev/null 2>&1 || true
done

echo "[3/5] MT-32:track → MIDI…"
for i in $(seq -w 0 17); do
  python3 tools/cmf_to_midi.py "work/music/track_$i.bin" "work/music/_tmp/$i.mid" >/dev/null 2>&1 || true
done

echo "[4/5] MT-32:MIDI → wav(fluidsynth + GM SoundFont)…"
docker run --rm -v "$PWD/work/music":/m debian:bookworm-slim bash -c '
  apt-get update -qq >/dev/null 2>&1 && apt-get install -y -qq fluidsynth >/dev/null 2>&1
  for i in $(seq -w 0 17); do
    [ -f /m/_tmp/$i.mid ] && fluidsynth -ni -g 1.0 -F /m/_tmp/mt32_$i.wav -r 44100 /m/sf2/TimGM6mb.sf2 /m/_tmp/$i.mid >/dev/null 2>&1 || true
  done'

echo "[5/5] 轉 OGG(q5 VBR)…"
for i in $(seq -w 0 17); do
  [ -f work/music/_tmp/sb_$i.wav ]   && ffmpeg -y -loglevel error -i "work/music/_tmp/sb_$i.wav"   -c:a libvorbis -q:a 5 "$EXP/sb/track_$i.ogg"
  [ -f work/music/_tmp/mt32_$i.wav ] && ffmpeg -y -loglevel error -i "work/music/_tmp/mt32_$i.wav" -c:a libvorbis -q:a 5 "$EXP/mt32/track_$i.ogg"
done
rm -rf work/music/_tmp
echo "完成:"
echo "  SB   → $EXP/sb/   ($(ls "$EXP/sb"/*.ogg 2>/dev/null | wc -l) 軌)"
echo "  MT32 → $EXP/mt32/ ($(ls "$EXP/mt32"/*.ogg 2>/dev/null | wc -l) 軌)"
