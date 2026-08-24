# 187 — 開場玩家路徑：OGG、注音、母親帶路與 NPC 動畫

日期：2026-08-24。此切片由玩家實際進入遊戲後的五項可見反證驅動，不重新開啟全域逆向。

## 結論

| 玩家反證 | 根因 | 現行修正 | 證據邊界 |
|---|---|---|---|
| 背景音樂有噪音 | 找不到 `mt32/track_00.ogg` 時自動安裝 `MBG.MCX` 的 OPL2 FM fallback | 正常產品只直接解碼 `track_NN.ogg`；`DQ3_FM` 僅供明確診斷 | runtime E2；不宣稱播放器 wall-clock V3 |
| `ㄨㄤˇ` 沒有「王」 | 生成表依標準讀音把 glyph186「王」列在 `ㄨㄤˊ`；`ㄨㄤˇ` 只有「往、枉」 | 標準候選不改，pack 另以 `user_report` D3 相容別名把「王」附加到 `ㄨㄤˇ` | 這是明示 UX 相容，不是原版或國語讀音斷言 |
| 組字注音被遮住 | renderer 把組字畫在姓名列第五格後，與右側功能欄重疊 | 新增 `name_composition` pack anchor，依原版 DOSBox 畫面畫在下方「輸入注音」框 | DOSBox 同畫面 D2／runtime E2 |
| rec81 後直接傳送 | `motherEscort` 直接切到 CTY00 sec0 `(8,38)` | `opening_escort` 由 pack 保存母親／玩家 tile frames；逐格走到右下樓梯後才執行 handler54 旗標與轉場 | 本機原版影片 4:53–5:12 證實可見順序；每格 6 frames 是可重播近似，不稱 wall-clock exact |
| NPC 靜止、移動面向錯 | `npcTick` 只改 tile；沒有寫 renderer facing／walk，交談轉向會殘留 | CTY ctrl 下／左／上／右轉為 renderer 下／上／左／右；待機每 18 幀切換兩幀，移動前同步面向 | 原始 ctrl mover 規則既有 D2；動畫節拍屬可見近似 |

## 資料與引擎邊界

- `interface.json.opening_escort` 保存 DQ3 專屬 CTY、section、母親／玩家路徑、停留幀與證據。
- `new_game_geometry.name_composition` 保存組字 anchor；`new_game_labels.zhuyin_candidate_aliases`
  保存版本字庫的相容候選。共用 Go 只處理有限 frame sequence、去重附加與 renderer。
- schema 為 `0.1.50`，DQ3 content 為 `0.1.56`。缺少這些資料時不猜 DQ3 座標或字模。

## 驗收與交付界線

具名測試鎖定 `ㄨㄤˇ` 包含 glyph186、標準候選順序不變、NPC 待機切幀與右移面向、母親
帶路有可見 hold frames 且最後才抵達 CTY00 sec0 `(8,38)` 並開 rec80。`dist-all/v0.1.36-local/`
早於本切片；在重新打包前不得描述為包含本輪修正。
