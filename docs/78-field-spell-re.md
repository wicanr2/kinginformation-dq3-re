# 78 — 野外咒文 RE 契約（2026-07-28）

本文件只記可重用證據與未決事項。目標是避免後續又從 C remake 的近似值反推原版。
位址皆為原始 `DQ3.EXE` file offset；分析以 `tools/dis.sh` 與 IDA Pro 9.4 交叉確認。

## 魯拉：已證實（D2）

- `0x5dfc`：魯拉入口；`0x5e55` 建立目的地選單；`0x5f9b` 執行傳送。
- DGROUP `0x0ac0` 的目的地 index → CTY：
  `0,1,2,3,4,6,12,16,15,17,47,21,39,43,38,79,81,84,85,86`。
- `0x5e66` 讀 `[CTY + 0x506a]` 的造訪 bit，只有已解鎖目的地進選單。
- `0x5f29..0x5f32` 使用 `0x1ef + destination index` 作 D3TXT00 正式地名 record。
- `0x5fbb` 以 CTY 查 `cty_loc`，寫 world `X=ctyX-1, Y=ctyY`。
- destination index 10（CTY47）在 `0x5fd9` 另將 X 加 2，最後為 `ctyX+1`。
- 因此「回城內最後位置」及「魯拉、烈米特共用 RETURN_TOWN」都不是原版契約。

Ebiten 已接正式命令窗 → 咒文 → 目的地，成功選定後才扣 MP；已造訪城鎮會存檔。
舊存檔只安全遷移必然起點 CTY0 與讀檔時所在城鎮，不猜測其他旅行紀錄。

## 尚未證實（不得標 faithful）

- 魯拉、烈米特、特黑洛斯、拉那魯達的精訊版 MP 值。目前 8/6/2/8 來自舊 C
  remake／經典資料，屬 D1。
- 烈米特的完整 handler、允許場景 gate、入口座標來源與副作用。目前 Ebiten 的
  「迷宮 music 類型內才可用，回 remembered overworld」是可玩近似。
- 特黑洛斯的原版步數、弱怪判定；目前 `repel=64` 為既有 remake 設定。
- 拉那魯達的原版時間狀態交易；目前只切換既有 day/night phase。

## 後續 IDA 工作方式

1. 保存 IDB，不再每輪只產生孤立的線性 disassembly。
2. 將 `0x5dfc/0x5e55/0x5f9b`、`0x0ac0`、`0x506a` 命名，建立 caller/xref。
3. 從咒文 dispatch table 反查 rec172/173/176/177 的 handler 與 MP consumer。
4. 每個結論記錄 writer → table → consumer → 可見副作用；只有完整鏈才升 D3。
5. 用 DOSBox 同狀態輸入作最終仲裁，再鎖 production-input 與 save/load 測試。

工具：`/home/anr2/ida_94_official/ida-pro-9.4/idat`；原版：
`assets_raw/DQ3.EXE`。授權檔及 IDA 發佈包不得加入 repo。
