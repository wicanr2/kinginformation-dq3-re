# 99 — 最終鑰匙、魯拉船重定位與提頓綠色寶珠 production trace

> 日期：2026-08-02（Asia/Taipei）
> 範圍：CTY40 最終鑰匙 checkpoint → 魯拉 → 巴哈拉達附近登船 → 夜間提頓牢門 → 綠色寶珠 → save/load
> 原版輸入：`DQ3.EXE` SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`

## 結論

本切片已達 **E3**：`TestOpeningProductionInputTrace` 從新遊戲開始，只送正式
`InputState`，經旅店、航行、道具與咒文選單抵達 CTY40；取得最終鑰匙後正式施放魯拉，
由原版目的地表同步重定位船，航行至提頓入口外使用黑暗之燈，再由道具選單使用最終鑰匙
開 tier-3 牢門，與夜間 NPC 交談取得綠色寶珠，最後通過 save/load。

已推翻兩個舊近似：

1. 調查門時掃描全隊最高階鑰匙並自動開門；原版是玩家從 rec421 道具動作選單明確選擇
   一把鑰匙「使用」。
2. 魯拉只移動隊伍；原版持有船時會以同一目的地 index 重定位船。

## 證據與推論等級

| 結論 | 原版位置 | 等級 | 閉合情形 |
|---|---|---|---|
| 魯拉目的 CTY 順序 | file `0x16c00`；DGROUP `0x0ac0` | `proven` | handler logical `0x4c4b..0x4c65` 以目的 index 讀 CTY，再由 `cty_loc` 寫玩家 X/Y。 |
| CTY47 arrival X 特例 | logical `0x4c69..0x4c75` | `proven` | 目的 index 10 在一般 X−1 後另加 2，等於 CTY X+1。 |
| 魯拉重定位船 | logical `0x4c7e..0x4c9c`；file `0x16c14`；DGROUP `0x0ad4` | `proven` | world-state bit0 set → 同 index 讀 20 組座標 → 寫 `DGROUP 0x4f3c/0x4f3e`；巴哈拉達 index8 寫 `(104,126)`，runtime 可由玩家正式登船。 |
| 魯拉重定位另一載具 | logical `0x4c9f..0x4cba` | `strong inference` | world-state bit1 set 時寫玩家落點 `(+1,-1)` 至 `DGROUP 0x4f40/0x4f42`；與現有不死鳥狀態相符，但本切片未做同狀態 DOSBox 對拍。 |
| 鑰匙由所選道具決定 tier | DGROUP `0x366a`；logical `0x3f65/0x3f6e/0x3f77` | `proven` | item `0x55/0x56/0x57` 各寫 tier `1/2/3` 到 `DGROUP 0x2593`，再進共同門 consumer logical `0x4906`。 |
| 開門不消耗鑰匙 | logical `0x4906..0x4a48` | `proven` | consumer 只比較面向 BLKBM attr bits6–7 與所選 tier，成功改 door tiles／播放 cue，沒有清 inventory slot。 |
| CTY20 夜間給綠色寶珠 | NPC subtype2 `byte4=35`；DGROUP `0x3bb4 + 35×2` → logical `0x593a` | `proven` | flag `0x3e` set → rec38 → item scratch `0x66` → 全隊八格 consumer；儲存成功後才 clear flag。flag clear 時播放 rec39。 |
| 全隊背包順序 | logical `0x6856` | `proven` | 依勇者、同伴順序各掃八格 `0x00ff`；長流程中勇者已滿，綠色寶珠合法落入第一位有空格的同伴。 |
| 全隊皆滿時的追加失敗對話 | logical `0x5037` 與全域文字 pointer `0x013a` | `unknown` | 現行 transaction 已 fail closed（不給物、不清 flag），但追加訊息的完整文字／時序尚未做玩家可見對拍。 |

IDA database、人工函式邊界、影片抽格與旁車輸出只存在 `/tmp/dq3-teidon-ida/`，不加入
Git。人工語意沒有覆寫原始自動名稱；每項結論保留 IDA linear、logical／DGROUP 位址與
推論等級。

## 原版影片

本機完整影片 `dq3_real_video/YTDown_YouTube_Media_J_fozjiKTB8_001_1080p.mp4`
約 `1740–1756` 秒顯示：

1. CTY40 取得最終鑰匙後開咒文選單施放魯拉；
2. 返回巴哈拉達一帶後前往提頓；
3. 在牢門前開道具選單，選最終鑰匙「使用」；
4. 門打開後與夜間犯人交談，取得綠色寶珠；
5. 再次交談顯示送往雷亞姆蘭特祭壇的後續文字。

影片角色使用修改過的 `M999`，只證明操作、場景與可見結果，不能當 MP 數值 oracle。
正式 trace 因此在最後旅店補滿後，以正常逃跑指令保留魯拉所需 8 MP。

## game-pack 資料

schema `0.1.14` 新增：

- `rura_navigation.destinations`：20 筆 CTY、名稱 record、玩家 arrival offset、正規化
  `{x,y,layer}` 船落點；由 EXE file `0x16c00/0x16c14` parity test 逐筆鎖定。
- `item_use_effects[].door_key_tier` 與 `open_facing_locked_door`：三把鑰匙只提供 tier，
  共用 Go 狀態機負責面向門 gate。
- `npc_item_reward_events`：提頓 NPC selector、flag `0x3e`、item `0x66` 與 rec38/39 的
  stable text ID。

Go 不再保存魯拉 CTY 順序、CTY47 特例、船落點或三把鑰匙 raw ID table；缺少目的地、
未知引用或非法座標時 loader／runtime 失敗即關閉。

## 驗收

- `TestDQ3RuraNavigationMatchesOriginalEXE`：20 筆 CTY／船落點／arrival offset 與原始 EXE 相等。
- `TestRuraRelocatesOwnedShipFromPack`：CTY15 魯拉後玩家 `(100,119)`、船 `(104,126)`。
- `TestPackDoorKeyRequiresFormalItemUse`：調查不開門；最終鑰匙正式使用才開，且不消耗。
- `TestPackDoorKeyFailsClosedBelowRequiredTier`：魔法鑰匙面對 tier3 門不改 tile。
- `TestTedonGreenOrbOriginalNightNPC`：夜間 selector、rec38/39、item／flag transaction。
- `TestTedonGreenOrbFullPartyInventoryDoesNotConsumeFlag`：全滿時不給物、不清 flag。
- `TestOpeningProductionInputTrace`：boot 起正式輸入、魯拉／船／黑暗之燈／牢門／NPC、全隊
  背包與 save/load 全部通過。

本批 runtime 圖：

- `docs/img/teidon_final_key_door_closed.png`
- `docs/img/teidon_final_key_door_open.png`
- `docs/img/teidon_green_orb_dialogue.png`

流程證據為 E3；本批 runtime 圖是 V2。原版同狀態逐像素對拍與全滿背包追加失敗對話仍是
後續 V3／呈現長尾，不影響已閉合的成功 transaction。
