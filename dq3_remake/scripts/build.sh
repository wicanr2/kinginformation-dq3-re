#!/usr/bin/env bash
# 在 docker(SDL2 + cmake)內編 dq3_remake,並做兩種 headless 驗證:
#   (1) dummy video driver:跑載入+解碼,dump PPM(不開視窗)。
#   (2) Xvfb:跑完整視窗 present 路徑,用 import 截真實視窗 PNG。
# 不污染 host:全在容器內;掛載 repo。
#   dq3_remake/scripts/build.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

# 確保 image 存在(不存在則建)
if ! docker image inspect dq3-remake >/dev/null 2>&1; then
  docker build -t dq3-remake "$ROOT/dq3_remake/scripts"
fi

docker run --rm -v "$ROOT":/repo dq3-remake bash -lc '
  set -e
  mkdir -p /build && cd /build
  cmake -S /repo/dq3_remake -B /build -DCMAKE_BUILD_TYPE=Release >/tmp/cmake.log 2>&1 \
    || { tail -30 /tmp/cmake.log; exit 1; }
  cmake --build /build -j >/tmp/make.log 2>&1 || { tail -50 /tmp/make.log; exit 1; }
  echo "=== build OK: $(stat -c %s /build/dq3_remake) bytes ==="

  # 數值/升級系統單元測試(#4/#5/#6 修正驗證;非 0 即失敗)
  echo "--- 數值系統單元測試 ---"
  /build/dq3_stats_test /repo/assets_raw || { echo "stats test FAILED"; exit 1; }
  /build/dq3_combat_test /repo/assets_raw || { echo "combat test FAILED"; exit 1; }
  /build/dq3_monster_test /repo/assets_raw || { echo "monster test FAILED"; exit 1; }
  /build/dq3_battle_test || { echo "battle test FAILED"; exit 1; }
  /build/dq3_inventory_test || { echo "inventory test FAILED"; exit 1; }
  echo "=== unit tests OK ==="

  # (1) dummy:解碼驗證 → PPM
  SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy DQ3_DUMP=/tmp/titg.ppm \
    /build/dq3_remake /repo/assets_raw TITG.P
  pnmtopng /tmp/titg.ppm > /repo/dq3_remake/titg.png 2>/dev/null || \
    cp /tmp/titg.ppm /repo/dq3_remake/titg.png
  echo "=== headless dump OK ==="

  # (1b) field 地表:繪一幀 + 腳本化走動(驗證捲動/碰撞)
  echo "--- field: 初始幀 ---"
  SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy DQ3_DUMP=/tmp/field0.ppm \
    /build/dq3_remake /repo/assets_raw field
  pnmtopng /tmp/field0.ppm > /repo/dq3_remake/field0.png 2>/dev/null || \
    cp /tmp/field0.ppm /repo/dq3_remake/field0.png
  echo "--- field: 走動 16 步後 ---"
  SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy DQ3_DUMP=/tmp/field1.ppm \
    DQ3_WALK="RRRRRRRRDDDDDDDD" \
    /build/dq3_remake /repo/assets_raw field
  pnmtopng /tmp/field1.ppm > /repo/dq3_remake/field1.png 2>/dev/null || \
    cp /tmp/field1.ppm /repo/dq3_remake/field1.png
  echo "=== field headless dump OK ==="

  # (1c) town 城鎮:CTY00 section0 + DQ31.BLK(對照已驗證 cty00_sect0_verify.png)
  echo "--- town: CTY00 section0 ---"
  SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy DQ3_DUMP=/tmp/town0.ppm \
    DQ3_CTY=CTY00.DAT DQ3_SECT=0 DQ3_BLKN=1 \
    /build/dq3_remake /repo/assets_raw town
  pnmtopng /tmp/town0.ppm > /repo/dq3_remake/town0.png 2>/dev/null || \
    cp /tmp/town0.ppm /repo/dq3_remake/town0.png
  echo "--- town: 走動 ---"
  SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy DQ3_DUMP=/tmp/town1.ppm \
    DQ3_CTY=CTY00.DAT DQ3_SECT=0 DQ3_BLKN=1 DQ3_WALK="DDDRRR" \
    /build/dq3_remake /repo/assets_raw town
  pnmtopng /tmp/town1.ppm > /repo/dq3_remake/town1.png 2>/dev/null || \
    cp /tmp/town1.ppm /repo/dq3_remake/town1.png
  echo "=== town headless dump OK ==="

  # (1d) battle 互動戰鬥:史萊姆×3,腳本「戰戰戰戰」打到結算
  echo "--- battle: 史萊姆 ×3 腳本 FFFF ---"
  SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy DQ3_DUMP=/tmp/battle.ppm \
    DQ3_MON=5 DQ3_MON_N=3 DQ3_BATTLE_SCRIPT="FFFFFFFF" DQ3_SEED=12345 \
    /build/dq3_remake /repo/assets_raw battle 2>&1 | sed "s/^/    /"
  pnmtopng /tmp/battle.ppm > /repo/dq3_remake/battle.png 2>/dev/null || cp /tmp/battle.ppm /repo/dq3_remake/battle.png
  echo "=== battle headless dump OK ==="

  # (1e) game 串接:地表走動→遭遇戰鬥→進城鎮(換場 + 重套 palette,bug #8)
  echo "--- game: 地表→戰鬥→城鎮 ---"
  SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy DQ3_DUMP=/tmp/game.ppm \
    /build/dq3_remake /repo/assets_raw game 2>&1 | sed "s/^/    /"
  pnmtopng /tmp/game.ppm > /repo/dq3_remake/game.png 2>/dev/null || cp /tmp/game.ppm /repo/dq3_remake/game.png
  echo "=== game headless dump OK ==="

  # (1f) text 對話視窗:阿里阿罕 NPC 對話(CJK 字模)
  echo "--- text: D3TXT01 rec 1 ---"
  SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy DQ3_DUMP=/tmp/text.ppm DQ3_TXT=D3TXT01.TXT DQ3_REC=1 \
    /build/dq3_remake /repo/assets_raw text 2>&1 | sed "s/^/    /"
  pnmtopng /tmp/text.ppm > /repo/dq3_remake/text.png 2>/dev/null || cp /tmp/text.ppm /repo/dq3_remake/text.png
  echo "=== text headless dump OK ==="

  # (2) Xvfb:完整視窗路徑 + 截圖(timeout 包住,有界,不掛起)
  export DISPLAY=:99
  Xvfb :99 -screen 0 1280x720x24 >/tmp/xvfb.log 2>&1 &
  XPID=$!
  sleep 1
  ( timeout 6 /build/dq3_remake /repo/assets_raw TITG.P >/tmp/run.log 2>&1 ) &
  RPID=$!
  sleep 3
  import -window root /repo/dq3_remake/titg_window.png 2>/tmp/import.log || \
    echo "(import 截圖略過: $(cat /tmp/import.log))"
  kill $RPID 2>/dev/null || true
  kill $XPID 2>/dev/null || true
  echo "=== Xvfb window screenshot done ==="
'
echo "輸出: dq3_remake/titg.png(解碼)、dq3_remake/titg_window.png(視窗截圖) — gitignored"
