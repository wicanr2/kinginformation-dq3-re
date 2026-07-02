# 評估:改用 Go / Ebiten 重寫以移植 Android(2026-06-28)

> 問題(使用者):現行 C99 + SDL2 的 remake,若改用 **Go + Ebiten(Ebitengine)** 重寫,移植 Android 的難度如何?
> 結論先講:**若 Android(甚至 iOS / 瀏覽器)是目標,Go/Ebiten 很可能比 SDL2/NDK(原 Tier 4)好走**——
> 因為困難的 RE 已做完,Go 版是「翻譯已知邏輯」而非「重新發現」;而 Ebiten 的行動裝置 + 觸控 + OGG 音訊是一等公民,
> **正好消掉 SDL2/NDK Android 路徑最痛的三塊**(NDK/SDLActivity、三 ABI vorbis 交叉編譯、觸控 UI)。

## 1. 現況規模(估工作量的基準)

- **~9,356 行非測試 C**(69 個 .c / 41 個 .h,約 12k 含測試)。**資料驅動**:遊戲邏輯/數值/表全從 `DQ3.EXE` RE 出。
- 最大子系統:`battlescene`(801)、`scene`(522)、`audio`(428)、`opl2`(232)+`cmf`(197)+`voc`(77)。
- **關鍵**:困難的智力工作(反組譯 DQ3.EXE、破格式、對數值公式)**已完成並文件化**(docs/)。
  Go 版是把「已確立的解析器 / 公式 / 表」用另一語言重寫 —— **翻譯,不是發現**。

## 2. 為什麼 Ebiten 對 Android 有利

| 痛點(SDL2/NDK 路徑,原 Tier 4)| Ebiten 怎麼解 |
|---|---|
| NDK + SDLActivity + Gradle/JNI 接原始碼 | `ebitenmobile bind` 直接把 Go 編成 `.aar` → Android Studio 專案引用,**無需手接 NDK/SDL** |
| **三 ABI libvorbis 交叉編譯**(arm64/armeabi-v7a/x86_64)地獄 | Ebiten 內建 `audio/vorbis` **純 Go OGG 解碼** → MT-32 OGG 免任何原生庫、免交叉編譯 |
| 觸控 UI 要自己合成 SDLK_* 餵回事件迴圈 | Ebiten `TouchIDs()`/`Touch` **一等公民**,直接讀觸控座標 |
| 素材進 APK + `SDL_RWops` 讀 | Go `embed.FS` 內嵌素材,或行動框架的 asset 讀取 |
| 存檔/設定可寫路徑各平台不同 | Ebiten 行動端有對應 storage API(已用 `dq3_user_dir` 抽象過,概念一致)|

**額外紅利**:同一份 Go/Ebiten 碼還能編到 **桌面(Win/Mac/Linux)** + **WASM(瀏覽器直接玩)** —— 一次寫、四平台。

## 3. 逐子系統移植難度

| 子系統 | C 現況 | Go/Ebiten 難度 | 說明 |
|---|---|---|---|
| **資料解析**(CTY/BLK/sprite/text/palette/monster/item/save)| 純 byte 解析 | 🟢 易(繁瑣)| Go `encoding/binary`+slice 很乾淨,1:1 翻譯;RE'd 格式照搬 |
| **遊戲邏輯**(戰鬥公式/RNG/升級/AI/事件/場景移動碰撞/對話/選單)| 純邏輯 | 🟡 中 | 邏輯照翻;RE'd 數值/表/公式的正確性直接繼承(docs 當 oracle)|
| **渲染**(indexed palette framebuffer → RGB + tile/sprite blit)| 自繪 fb | 🟡 中 | 保留 indexed fb,每幀用 palette 轉 RGBA 畫成 `*ebiten.Image`;Ebiten 管縮放/present |
| **音訊 — MT-32 OGG** | libvorbis(桌面)/ 缺(Android)| 🟢 易 | Ebiten `audio/vorbis` 純 Go,**Android 免庫**——反而比現在簡單 |
| **音訊 — SB FM(OPL2/CMF 合成)** | 自寫 DSP(~430 行)| 🟠 中–高 | 唯一較硬的一塊:把 OPL2 FM 合成器移成 Go(或找 Go OPL2 lib);**但可延後**(預設走 MT-32 OGG)|
| **音效 VOC** | 解碼→PCM | 🟢 易 | 解成 PCM 餵 Ebiten audio |
| **輸入 + 觸控** | SDL 鍵盤 | 🟡 中 | Ebiten 鍵盤易;觸控是新做但 Ebiten 原生支援(比 SDL 路徑簡單)|
| **CJK 字型**(自建 16×16 glyph)| 自繪 | 🟢 易 | glyph blit 照翻 |
| **檔案/素材/存檔路徑** | stdio + `dq3_user_dir` | 🟢 易 | `embed.FS` + 行動 storage |

## 4. 工作量估計

- **總量**:~9k LOC 的「理解過的邏輯」重表達成 Go + Ebiten glue。因為是翻譯非發現,**估 3–6 週專注工時**(單人,熟 Go)。
- **最硬**:OPL2 FM 合成器移植(DSP)——但**可先跳過**,Android 版預設用 MT-32 OGG(Ebiten 原生),SB-FM 之後補。
- **最省**:資料解析 + 數值公式(照 docs/現有 C 翻),幾乎機械化。
- 對比原 Tier 4(SDL2/NDK Android):NDK/SDLActivity/三 ABI vorbis/觸控從零 = 高風險長尾;Ebiten 把這段變成「內建」。

## 5. 風險 / 取捨

- **這是重寫不是移植**:C→Go 兩種範式,~9k 行要重表達。但邏輯已懂 → 風險是「工時」不是「未知」。
- **雙 codebase**:若桌面 C 版與 Go 版都維護會分裂。**建議**:Android/行動/web 若是目標,**Go/Ebiten 成為主線**;
  C/SDL2 版保留為 **RE 參考 + 已交付的桌面包**(docs 仍是 oracle)。
- **忠實度不變**:remake 本來就非 byte-identical 原版(是 RE 後的重表達);Go 版沿用同一套 RE'd 邏輯/表,忠實度來源不變。
- **OPL2 移植**是唯一技術不確定點;先用 MT-32 OGG 出 Android，SB-FM 當 Tier-4 級後補。

## 6. 分階段 plan(建議)

1. **骨架**:Go module + Ebiten 開窗 + 主迴圈 + 一個場景(地表)畫出來。驗證 Ebiten 渲染管線。
2. **資料解析移植**:CTY/BLK/sprite/palette/text/monster/item/save 的 Go parser(對現有 C + docs 逐一翻),寫 Go 單測對拍。
3. **渲染**:indexed fb → RGBA → `*ebiten.Image`;tile/sprite/字型 blit。地表 + 城鎮畫對。
4. **遊戲邏輯**:場景移動/碰撞/NPC、對話、選單、戰鬥(公式/AI/升級)、事件/傳送。對現有 game_tester 斷言移成 Go 測。
5. **音訊**:MT-32 OGG(Ebiten vorbis,先)+ VOC 音效;SB-FM OPL2 之後補。
6. **輸入 + 觸控 UI**:鍵盤(桌面/web)+ 虛擬方向鍵/A/B/選單(行動)。
7. **Android**:`ebitenmobile bind` → `.aar` → Android Studio 專案 + 素材入 APK + 版權閘 → APK/AAB。
8. **紅利**:同碼編 WASM(瀏覽器 demo)+ 桌面(Win/Mac/Linux)。

## 7. 建議

- **要 Android/行動/web** → Go/Ebiten 是更好的路(消掉 NDK/三 ABI vorbis/觸控三大痛;OGG/觸控內建)。工時可控(邏輯已懂)。
- **只要桌面** → 維持現行 C/SDL2(已打包 Linux/Win/AppImage,macOS 走 CI)——不需重寫。
- 折衷起手:先做階段 1–2(骨架 + 一個場景 + 資料解析),**用最小成本驗證 Ebiten 管線 + Android build 真的順**,再決定要不要全量重寫。
- 現行 C 版的 docs/ + game_tester 斷言 = Go 版的**規格與對拍 oracle**,重寫時不會沒方向。
