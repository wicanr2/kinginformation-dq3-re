# dq3_remake_ebitan — 精訊 DQ3 的 Go / Ebiten port

> 目標:把 [`../dq3_remake`](../dq3_remake)(C99 + SDL2)的精訊 DQ3 remake 用 **Go + Ebiten(Ebitengine)** 重寫,
> 主要為了**乾淨的 Android/iOS/WASM 移植**(Ebiten 行動/觸控/OGG 內建,消掉 SDL2/NDK 路徑三大痛)。
> 評估與分階段 plan:[`../docs/62`](../docs/62-golang-ebiten-android-port-eval.md)。
>
> **原則**:困難的 RE 已在 C 版完成 —— 這裡是**翻譯已知邏輯**,不是重新發現。
> C 版 `dq3_remake/` 的原始碼 + `docs/` + `game_tester` 斷言 = 本 port 的**規格與對拍 oracle**。
> 素材(原版 `assets_raw/`)使用者合法持有、gitignore 不散布;需原版 `DQ3.EXE` 啟動(版權閘)。

## 結構

```
dq3_remake_ebitan/
├── go.mod / go.sum        # module(Ebiten v2)
├── main.go                # Ebiten Game shell(開窗/主迴圈/Layout 640×350)
├── internal/dq3data/      # ★ 純 Go 資料解析器(移植自 C;無引擎相依,可 headless 測)
│   ├── palette.go         #   DQ3.PAL 解碼(移植 dq3_pal_decode)
│   └── palette_test.go    #   對拍真實 DQ3.PAL(逐色驗證與 C 公式一致)
└── build.sh               # docker golang 建置(不污染 host):test + Ebiten compile-check
```

## 建置(docker,不污染 host)
```bash
bash dq3_remake_ebitan/build.sh   # 純 Go 單測 + Ebiten shell compile-check(Go 1.24 + GL/X11/ALSA)
```
桌面實際跑要有顯示器(`DQ3_ASSETS=/path/to/assets_raw go run .`)。

## 進度(對照 docs/62 分階段 plan)

- [x] **階段 1 骨架**:Go module + Ebiten 開窗/主迴圈/Layout(640×350);**管線驗證** —— 載入真實 `DQ3.PAL`
  用移植的 `dq3data.DecodePalette` 解碼、Ebiten 畫成 palette swatches。Ebiten v2.9.9 compile OK。
- [~] **階段 2 資料解析移植**(進行中):
  - [x] palette(DQ3.PAL)+ 對拍測試
  - [ ] BLK tile / sprite(BLS/SHP)/ text(D3TXT)/ CTY / monster / item / save … 逐一移植 + Go 測對拍 C
- [ ] **階段 3 渲染**:indexed fb → RGBA `*ebiten.Image`;tile/sprite/字型 blit;地表+城鎮畫對
- [ ] **階段 4 遊戲邏輯**:場景移動/碰撞/NPC、對話、選單、戰鬥(公式/AI/升級)、事件/傳送(對 game_tester 斷言移 Go 測)
- [ ] **階段 5 音訊**:MT-32 OGG(Ebiten `audio/vorbis` 內建,先)+ VOC 音效;SB-FM OPL2 之後補
- [ ] **階段 6 輸入 + 觸控 UI**:鍵盤(桌面/web)+ 虛擬方向鍵/A/B/選單(行動)
- [ ] **階段 7 Android**:`ebitenmobile bind` → `.aar` → Android Studio + 素材入 APK + 版權閘 → APK/AAB
- [ ] **階段 8 紅利**:同碼編 WASM(瀏覽器 demo)+ 桌面
