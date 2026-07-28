# 78 — 野外咒文 RE 契約（2026-07-28）

本文件只記可重用證據與未決事項。目標是避免後續又從 C remake 的近似值反推原版。
位址皆為原始 `DQ3.EXE` file offset；分析以 `tools/dis.sh` 與 IDA Pro 9.4 交叉確認。

## 真正的咒文施放路徑：已證實（D2）

- `0x1c9ee`：咒文選單；`0x1cb3c`：field caster。後者以
  `spellID = rec-0x79`、stride 3 查 DGROUP `0x37c3`：
  byte0=MP、byte1=base、byte2=target/effect flags，先檢查並扣施法者 `[member+0x18]` MP。
- field effect pointer table 位於 DGROUP `0x38cc`，以 `spellID-0x28` 索引。
- 真正魯拉 handler 是 file `0xe0da`；它通過場景 gate 後呼叫
  `0xe288 → 0x5e66 → 0x5f9b`（16-bit near-call IP 回繞）。
- `0x5e66` 建立目的地選單；`0x5f9b` 執行傳送。相同目的地效果也會被回城道具共用；
  舊文件把 `0x5140` 道具派發入口直接稱作魯拉入口，口徑錯誤。
- DGROUP `0x0ac0` 的目的地 index → CTY：
  `0,1,2,3,4,6,12,16,15,17,47,21,39,43,38,79,81,84,85,86`。
- `0x5e66` 讀 `[CTY + 0x506a]` 的造訪 bit，只有已解鎖目的地進選單。
- `0x5f29..0x5f32` 使用 `0x1ef + destination index` 作 D3TXT00 正式地名 record。
- `0x5fbb` 以 CTY 查 `cty_loc`，寫 world `X=ctyX-1, Y=ctyY`。
- destination index 10（CTY47）在 `0x5fd9` 另將 X 加 2，最後為 `ctyX+1`。
- 因此「回城內最後位置」及「魯拉、烈米特共用 RETURN_TOWN」都不是原版契約。

原版 caster 在 MP 足夠且通過施法者前置檢查後，會先播放共通施法流程並扣除 MP，之後才
派發 effect handler。目的地取消或 handler 的場景失敗沒有一般退款分支。因此 Ebiten
現行「效果成功／目的地選定後才扣 MP」是待修正差異，不得寫成原版契約。已造訪城鎮會存檔。
舊存檔只安全遷移必然起點 CTY0 與讀檔時所在城鎮，不猜測其他旅行紀錄。

## MP 與場景 gate：已證實（D2）

| rec | 咒文 | MP | handler | 已證實效果 |
|---|---|---:|---|---|
| 172 | 魯拉 | 8 | `0xe0da` | 地表可用；CTY 需 section header `+0x10 bit0`，再進共用 20 城選單 |
| 173 | 烈米特 | 8 | `0xe10c` | 必須在 CTY scene 且 header `+0x10 bit1`，成功後回地表 remembered position |
| 176 | 特黑洛斯 | 4 | `0xe19c` | 設世界 repel flag，`[0x52f6]=0x28`（40 步） |
| 177 | 拉那魯達 | 12 | `0xe1ab` | CTY scene 內拒絕；地表切換 day/night state 並重載 palette |
| 178 | 雷姆歐魯 | 15 | `0xe1df` | 設 world flag `0x8000`、`[0x52f7]=0x19`（25 個有效移動步） |
| 179 | 阿巴卡姆 | 0 | `0xe1f2` | 設 door permission 4，呼叫 `0x14906` 開面向的 tier 1..3 門 |

Ebiten 已新增 CTY `MapFlags` parser，移除「用 dungeon BGM 猜烈米特可用」的舊 gate，
並將 MP 與特黑洛斯步數改為上表原值。雷姆歐魯已接入選單、25 步 consumer 與
party／vehicle occupant renderer；原版 expiry routine file `0xaaca` 在 timer 歸零時清除
world flags `0xc000`。目前為靜態資料流與 Go production-input 回歸（D2），仍待 DOSBox
同狀態畫面升 D3。阿巴卡姆的 door consumer 會先改中心門，再由 file `0x14a49` 掃相鄰
8 格的連續門片；非 CTY、面向非門或其他失敗會顯示 D3TXT00 rec `0x15a`。Ebiten 已接
0 MP 全門權限、相鄰門片 mutation 與全域失敗訊息。

## 尚未證實／尚未完成（不得標 faithful）

- handler 中 `0xd75`、world flag `0x4f46 bit0x80` 等較少見魯拉禁止狀態，Ebiten 尚無
  一一對應的世界狀態欄位。
- 因帕斯、托拉瑪那與帕爾普恩特尚未完整接入 Ebiten 選單。
- 原版冒險之書是否保存 `[0x4f46]/[0x52f7]` 這類暫態場景咒文狀態尚未追完；Go save
  目前明確視為 runtime-only，讀檔會清除雷姆歐魯 timer。
- 拉那魯達的四相位 clock 與原版二值 `[0x526c]` 的逐步映射仍需 DOSBox 對拍。

## 後續 IDA 工作方式

1. 保存 IDB，不再每輪只產生孤立的線性 disassembly。
2. 將 `0x1c9ee/0x1cb3c/0xe0da/0xe10c`、`0x37c3/0x38cc`、`0x0ac0/0x506a`
   命名，建立 caller/xref。
3. 續追 rec174/175/180 的 field handler 與可見副作用；rec178/179 補 DOSBox 同狀態驗證。
4. 每個結論記錄 writer → table → consumer → 可見副作用；只有完整鏈才升 D3。
5. 用 DOSBox 同狀態輸入作最終仲裁，再鎖 production-input 與 save/load 測試。

工具：`/home/anr2/ida_94_official/ida-pro-9.4/idat`；原版：
`assets_raw/DQ3.EXE`。授權檔及 IDA 發佈包不得加入 repo。
