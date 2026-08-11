# 晝夜系統(2026-06-27;★含一次結論更正)

> 使用者問:「遊戲是否有黑夜系統(野外走路會白天/黃昏/黑夜/黎明轉換)?」

## 結論(更正後)

| 問題 | 答案 |
|---|---|
| 原版有晝夜**狀態**? | **有**(影響事件:夜晚進村開牢門、商店、NPC)|
| 城鎮 NPC 白天/黑夜差異的底層機制? | **每 section 兩份 NPC 清單**(word0=白天、word2=黑夜),由 `[0x526c]` 選載 → 詳見 [`docs/60`](../60-daynight-npc-mechanism.md)(已 RE + remake 已還原)|
| 走動會**自動循環**晝夜(白天→黃昏→黑夜→黎明)? | **會**,且是**步數計數驅動**(非時間)|
| 還能手動切換? | 能:**ラナルータ 拉那魯達(咒 rec177)切換 + 黑暗之燈(道具 0x5f)強制變夜** |
| remake 現況 | **已實作**(2026-06-27,見下)|

## 2026-08-10 靜態精確值補充（本節優先於舊版未決描述）

本節以同一份 `assets_raw/DQ3.EXE` 的 IDA Pro 9.4 靜態 listing 補上 clock／palette
的原始數值；它只升級原版資料流的 E1／D2 證據，沒有把現行 Go 的四階段近似宣稱為
逐像素 V3。

- 輸入：115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `sub_1EE23`（IDA linear `0x1ee23`、logical `0xee23`、file `0x10193`）在有效
  overworld update 中遞增 `DGROUP:0x251d`；`0x4f2d==1` 或 `0x5051==1` 時停止。
- clock 範圍是 `0..0xef`；到 `0x78`（120）寫 `DGROUP:0x526c=0`，到 `0xf0`
  （240）清零並寫 `0x526c=1`。因此是 240 個有效 tick 的二值 day/night selector，
  不是每一個 render frame 都計數。
- `sub_1EE76`（file `0x101e6`）以 `clock/0x14` 查 DGROUP `0x25c5`，其 12 bytes
  （file `0x18705`）為 `01 00 00 00 01 02 03 04 04 04 03 02`；再由
  `DS:0x3232 + index*0x30` 選 DQ3 palette bank，呼叫 `sub_20A3A` 上傳 16 色 DAC。
- `sub_1EE9B`（file `0x1020b`）做相同 bank 選擇但不直接上傳；`sub_1ECDC` 將
  `DQ3.PAL` `0xf0` bytes 載入 `DS:0x3232`，將 `MNSBK.PAL` `0x1e0` bytes 載入
  `DS:0x3352`，戰鬥背景再由 `sub_1C688` 以 `DS:0x3352+[0xd73]*0x30` 選取。
- `sub_131BD`（file `0x452d`）以 `[0x526c]==1` 選 section header `+0` 日表，否則
  選 `+2` 夜表，並用 record byte5 測 `[0x4f70]` story bit。

以上是原版 clock／palette 的精確靜態設定；現行 `dnPhaseSteps=60`、`DarkenPalette`
仍是 remake 的可玩近似與資料介面，需同一地點／同一 clock 的 DOSBox 畫面及 VGA/DAC
對拍後才可升 V3。完整戰鬥／音效／21 flags ledger 見
[`docs/123`](../123-static-battle-daynight-re.md)。

### ★ 結論更正紀錄(誠實留痕)
- 先前(本檔初版)誤判「走動不會自動循環,只靠咒/道具」,**論證錯誤**:我用「黑暗之燈存在 →
  夜晚不會自動來」當柵欄反證。**這個推論有漏洞** —— 黑暗之燈仍有用(立即強制變夜、室內/特定場景
  用),所以它存在**不能**反證自動循環。
- **使用者(玩過破關,ground truth)指正:DQ3 野外走路就是用步數計數自動變晝夜。** 採信並更正。
- 教訓:第一性原理論證若建立在「某設計若存在則 X 必不成立」的反證上,要先確認該設計**沒有其他
  存在理由**;黑暗之燈有「立即/室內變夜」的獨立用途,反證不成立(對齊柵欄原則:拆柵欄前先確認
  自己比建造者更懂它為何存在)。

## 舊版追查紀錄（截至 2026-06-27；未決句子由上節靜態補充修正）

- **遭遇步數計數器** `[0x52f4]` 已 RE(`sub_bd97`,file 0xd118 `dec byte [0x52f4]`,每步遞減→歸 0 起戰)。
  晝夜步數計數器是**同類的每步機制**,但在 overworld 處理鏈(`update_hud` 0xa8a0 / 地表 handler
  `sub_255b` 0x38cb / 移動 handler)裡;逐指令未完全定位到那一格(多層 handler,~9 輪反組譯未鎖死)。
- **機制已確立**(步數驅動)= 使用者 ground truth + 與遭遇計數器同源的每步設計。這段舊紀錄
  當時尚未定位確切值；2026-08-10 已由上方靜態補充證實 clock `0..0xef`、`0x78/0xf0`
  邊界及 12-byte palette index，剩餘是同狀態畫面／DAC 的 V3 對拍。

## remake 實作(2026-06-27)

- **相位狀態**(`dq3_scene.c`):`g_dn_phase` 0=白天 1=黃昏 2=黑夜 3=黎明;`dq3_scene_set/get_daynight`。
- **palette 調暗**:`dq3_scene_apply_palette` 依相位縮放 bg RGB(夜 42/44/70%、黃昏 82/62/58%、
  黎明 72/74/92%);sprite slot 用「淺一半」暗化 → 角色夜間仍看得見。海面 cycling(`animate_sea`)
  尾呼 apply_palette → 自動跟著變暗。
- **步數驅動**(`main.c`):地表每走一步 `g_dn_step++`,每 `DN_PHASE_STEPS=60` 步推進一相位
  (使用者指定:白天→黑夜 120 步、中間黃昏;全 4 相位循環 240 步)
  並重套 palette。接在與聖水倒數/中毒同一每步點。
- **手動切換**:野外「咒文」**拉那魯達(rec177)** → toggle 白天↔黑夜;**黑暗之燈(0x5f)** 道具
  只可在地表使用 → 強制變夜(不消耗,可重用)。IDA 完整派發、地表 gate、clock writer 與
  production trace 見 [`docs/93`](../93-teidon-dark-lamp-production-trace.md)。
- **驗證**:headless dump 地表 phase0 vs phase2 pixel diff 494437、夜景明顯調暗;game_tester 80/80。

### 已完成 / 待精校
- ✅ **晝夜存檔持久化**:save v6 起 `dq3_save_pos.daynight` 持久;讀檔還原相位(main.c 797/1128)。
- ✅ **夜 gated 事件**：提頓《黑暗之燈》、最終鑰匙牢門與 handler35 綠色寶珠均已由
  Go/Ebitengine boot production trace 閉合；正式航行靠岸後在入口外使用黑暗之燈，重新載入
  CTY20 夜表，再以 rec421 使用最終鑰匙並交談取得寶珠。見 `docs/93`、`docs/99`。
- 步數已設使用者指定(白天→黑夜 120 步)；現行 Go 四 phase 仍是 engine 近似，原版二值
  `[0x526c]` 與 240 tick clock 的靜態值已在本文件上方補正。
- 各相位 palette 的原版 bank 選擇已 `confirmed`；runtime RGB／DAC 與現行 PNG 的同地點、
  同 clock 逐格比色仍是 V3，不能用近似暗化取代。
