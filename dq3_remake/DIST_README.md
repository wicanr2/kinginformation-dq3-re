# Dragon Fighter III(精訊版 DQ3)現代重製版 — 交付說明

本重製以現代 C99 + SDL2 還原精訊未發售版《勇者鬥惡龍 III》(Dragon Fighter III, 1993)的遊戲玩法,
跨平台(Linux / Windows / macOS / Android,皆經 SDL2)。原始遊戲素材有版權,**不隨包附帶**,
需自行從原版 `DQ3.EXE` 等檔案準備。

## 內容物

| 路徑 | 說明 |
|---|---|
| `bin/dq3_remake` | Linux x86-64 主程式(動態連結 SDL2)|
| `bin/dq3_*_test` | 單元測試執行檔(數值/戰鬥/系統)|
| `verify/` | 驗證套件(`game_tester.sh` / `playthrough_check.sh` / `mainline_check.sh`)|
| `run.sh` | 啟動腳本(指定 assets 目錄)|
| `DIST_README.md` | 本說明 |

## 準備素材(必要)

把原版檔案放到一個目錄(下稱 `assets_raw/`),至少需要:
`DQ3.EXE`、`DQ3.PAL`、`DQ31.BLK`/`DQ3N.BLK`、`DQ3CON.MAP`/`DQ3UND.MAP`、`CTY*.DAT`、
`D3TXT*.TXT`、`D3MNS*`、`ITEM.DAT`、`DQ3MAN.BLS`、`ENDTXT.TXT` 等(完整清單見原版發行檔)。

> 素材版權屬精訊/原作,本專案僅還原程式邏輯,不重新散布素材。

## 執行

```sh
# Linux(需安裝 SDL2 runtime:libsdl2-2.0)
./run.sh /path/to/assets_raw
# 等同
SDL_VIDEODRIVER=x11 bin/dq3_remake /path/to/assets_raw game
```

啟動後進入遊戲主迴圈:地表移動、進城、對話、戰鬥、商店、創角(露依達酒場)、存檔。

### 操作

方向鍵移動;Enter 對話/調查/確定;Space 翻頁;C 野外指令窗(つよさ/道具/咒文);
B 設施;F10 離開(含自動存檔)。

## 驗證(交付前已全綠)

容器或本機(需 SDL dummy driver + assets)執行:

```sh
SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy verify/game_tester.sh /path/to/assets_raw bin/dq3_remake
```

涵蓋:15 單元測試 + 35 系統孤立斷言 + 主線一條龍(進度 1→9 破關)+ 全 89 城載入零崩潰 +
索瑪二階段(光之珠)+ 結局捲動 + sub2 給物 NPC。

## 從原始碼建置(其他平台)

原始碼:`https://github.com/wicanr2/kinginformation-dq3-re`(`dq3_remake/`)。需 CMake + SDL2 dev。

```sh
cmake -S dq3_remake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build -j
```

- **Windows**:用 MSVC/MinGW + SDL2 dev,或 vcpkg 取 SDL2;CMake 同上。
- **macOS**:`brew install sdl2`,CMake 同上。
- **Android**:SDL2 Android 專案殼包 `dq3_remake/src` + `include`,assets 放 APK assets,主迴圈同 `main.c::run_game`。

## 重製狀態(誠實揭露)

- **完整可玩核心**:創角→地表(雙層 overworld)→城鎮(89 城/NPC/對話/轉場)→互動戰鬥
  (怪物 AI/咒文/裝備/升級,原版 7 bug 全在 C 層修正)→商店/旅社/教會/存檔。主線里程碑
  8 段真實 NPC 觸發,可一條龍推進到破關(索瑪戰 + 結局)。
- **未發售版固有缺口**:精訊版為未上市早期 build,部分劇情觸發/文本原始資料即未接好
  (例:結局 ENDTXT 文字未定稿、部分 runner 事件無靜態觸發)。本重製以自有旗標流補上主線連接,
  不 1:1 還原半成品的劇情 runner VM。詳見 `docs/47-remaining-work.md`。
