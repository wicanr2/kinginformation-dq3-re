# 186 — 角色方向、交談轉向、HELP／系統設定與音樂收尾

日期：2026-08-23。這是玩家進入遊戲後提出的實際反證所驅動的窄切片，不重新開啟全域 RE。

## 玩家可見反證與修正

| 問題 | 根因 | 修正 | 證據等級 |
|---|---|---|---|
| 按上顯示向左、按左顯示向上 | BLS raw entry 為下／左／上／右，renderer 直接套用引擎下／上／左／右碼 | `LoadCharSprite` 在 decoder 邊界以 `{0,2,1,3}` 重排，所有角色共用同一邏輯方向 | 玩家實測 confirmed；decoder test E2 |
| NPC 對話不面向主角 | `cmdTalk` 命中 NPC 後沒有 facing writer | 交談入口寫入 `playerFacing ^ 1`，即下↔上、左↔右，再進既有 scripted／generic consumer | 玩家實測 confirmed；正式 command test E2 |
| 沒有 HELP／遊戲中系統設定 | `S` 只在標題處理，沒有 F1 input role | `F1` 開 pack-owned HELP；`F2`／`S` 在一般地表開既有 Settings；Confirm／Cancel 關閉 | remake UX D2／E2，不冒稱原版 UI |
| 一般完整版沒有音樂 | desktop 只有 `DQ3_MT32` 才傳音樂 FS，FM 又要求 `DQ3_FM`，正常啟動必然靜音 | 完整版若有 `assets_raw/mt32/track_00.ogg` 即預設 MT-32 OGG；否則建立 assets audio backend 並讀 `MBG.MCX` 作 SB-FM fallback | runtime wiring E2；人耳／硬體 wall-clock 不由本測試升格 |

HELP 的視窗、字模列與 evidence 在 `interface.json.help`，schema `0.1.49`、content `0.1.55`；
Go 只保留通用 modal 與 renderer。欄位說明已追加至 `docs/84`。

## 驗收

- Docker＋Xvfb 具名測試：NPC 四方向轉向、F1/F2 開關、BLS 上／左重排、builtin pack load 全過。
- 排除使用者受保護空白 `tmp_dump.go` 的容器隔離副本可完成桌面根套件純編譯。
- `internal/gaudio` OGG decoder test 通過；本機 MT-32 `track_00..17.ogg` 均存在。
- `dist-all/v0.1.36-local/full/` 四個完整版均包含 18 軌 OGG、`MT32_CONTROL.ROM` 與
  `MT32_PCM.ROM`。Linux AppImage 解包與 ALSA null sink＋Xvfb 8 秒 smoke 通過；三個 ZIP
  CRC 與私有素材數量通過。Windows／macOS 只做靜態架構／封包驗證，macOS 未真機驗收。

## 散布邊界

MT-32 ROM、原版素材與預錄音軌只保留於本機完整版，不加入 Git、不放公開 patch 或 GitHub
Release。ROM SHA-256 保存於 `dist-all/v0.1.36-local/MT32-ROM-SHA256.txt`；本文件不收錄 ROM。
