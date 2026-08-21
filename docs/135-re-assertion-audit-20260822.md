# DQ3 反組譯斷言審計（2026-08-22）

## 結論

DQ3.EXE **沒有被完整語意解讀**。現有工程已對 remake 必經玩家路徑閉合多個有限的
入口／writer／資料／consumer 鏈；同時仍保留若干原版未知與玩家可見 V3 缺口。

`docs/17` 的整檔 SHA-256 相同只證明 MZ 區段與原始 bytes 可無損重生。load image 大量
使用 `db` 原樣保存，所以該結果不能證明每個 byte 的程式／資料分類、欄位語意、caller、
副作用或玩家可見結果。`docs/19` 的 C byte-match 也只完成方法與單函式 PoC，未覆蓋全檔。

本審計不重開全域反組譯。依「可玩性優先、證據最小充分」原則，只有會改變現行玩家
體驗或交付 gate 的新 discrepancy，才開獨立窄任務並優先用 IDA Pro 9.4 追查。

## 證據基準

- 原始輸入：`assets_raw/DQ3.EXE`，115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 主要 RE 工具：IDA Pro 9.4.0.260610；IDA linear = logical + `0x10000`，主 load image
  file = logical + `0x1370`。個別文件若採不同位址空間，仍以該文件標記為準。
- bytes 保真：`docs/17`；有限戰鬥／日夜／story flag 靜態鏈：`docs/123` 及
  `docs/data/dq3-static-battle-daynight-evidence.json`；現行 remake 狀態：`docs/74`。

## 現行狀態表

| 範圍 | 原版證據 | remake 狀態 | 不得擴張的斷言 |
|---|---|---|---|
| MZ 結構與輸出 bytes | confirmed：整檔可 byte-identical 重生 | 研究工具可重生 | 不等於全檔語意已理解 |
| seg0 候選指令解碼 | strong：12,949 條助記符 round-trip 未見已證實解碼錯誤 | 供定位 | 不等於函式語意、資料欄位與間接 xref 完整 |
| 敵方共用行動 dispatcher | confirmed：狀態／逃跑／物理／咒文主要分支與 raw 欄位 | 已資料化 | 不涵蓋 boss repeat-N、全部咒文語意、逐動作 frame／PCM timing |
| formation | confirmed raw position formula；pixel surface 為 strong | 已接線 | 不得稱逐像素 V3 |
| 日夜 | confirmed clock `0..0xef`、`0x78` boundary、12-byte palette selector | 五個原始 bank／12 段 selector D3／E2 | 尚無同 clock DOSBox DAC 畫面，不得稱 V3 |
| 五個 story flag | confirmed 原版 writer、event chain、generic NPC consumer | 有限 transaction E2 | E2 不等於 direct GET 存在或玩家可見 V3 |
| 必經 Boss 背景 | selector／archive-page-palette 證據依 `docs/128`、`129` | V1；八頭大蛇 near-state V2 | 不得概括為全部 Boss V3 |
| 正式主線 | remake 正式 input trace 到 THE END | transaction E3 | 不等於全遊戲每畫面、音效、時序皆原版 parity |

## 仍未知或未閉合

- boss 同回合 repeat-N 的原版觸發來源仍為 `unknown`；正式 queue 只證明每個 entry
  呼叫一次 actor consumer。
- 每個動作實際 SHP frame、PCM 波形／host wall-clock、cue 語意與玩家可見停頓未達 V3。
- 日夜五個原始 palette bank 與 12 段 selector 已接線；同 clock DOSBox DAC／畫面對拍
  仍未閉合，因此玩家可見結果尚未達 V3。
- generic terrain → battle background 全表未閉合；沒有新玩家可見 discrepancy 時不追。
- formation 的 raw 公式已接線，但所有編隊的逐像素位置並未以同狀態原版畫面證實。
- 抗性 raw descriptor／threshold 已閉合；全部中文效果名稱與語意不可由 raw 值外推。
- temporary solo challenge 的 completed writer 尚未由原版證實。
- 對話保留控制碼的部分語意、church poison／curse 持久狀態，以及怪物記錄 `+0x27`
  仍未閉合。
- NPC `b2<4` 原版會讀到 stale buffer，玩家可見結果未定義；remake 使用 entry 0 是
  可玩性 fallback，不是忠實復原。

以上未知不自動阻塞目前 remake。若新原版 oracle 顯示玩家可見差異，才建立獨立 ledger，
以 `confirmed／strong／hypothesis／unknown` 分級追入口到 consumer；不得再用「合理值」
或舊 C remake 補成 production 事實。

## 本輪訂正

- `docs/17`、`docs/19`：移除「byte-identical 等於每個 byte 語意正確」的錯誤推論。
- `docs/20`：patch 能力限縮為有獨立資料流證據的精確位址，回歸聲明限於抽測路徑。
- `docs/37`、`docs/16`：把「完整／全部／完全一致」限縮為明確的 dispatcher 或目視抽樣。
- `docs/74`：E2 與 V 等級分開；主線 E3、Boss selector 接線與 V1／V2 不再混稱完成。
- `docs/123` 與 JSON sidecar：保留五 flags 的歷史 unknown，同時記錄後續有限 E2 接線；
  不把 E2 冒稱原版 direct GET 或 V3。
- Go 註解：移除已過期的「命令尚未移植」、日夜 exact、NPC fallback 忠實與 formation V3
  暗示；未改 runtime 行為。
- 日夜窄切片後續訂正：以 IDA 9.4／IDAPython 閉合 `sub_1EE23`、`sub_1EE76`、
  `sub_1EE9B`、`sub_1ECDC`，將五個原始 `DQ3.PAL` bank 與 12 段 selector 接入 pack；
  移除舊 C／Go 的無證據 RGB 縮放，並修正 clock `0x78..0xef`（phase 2／3）皆選夜間
  NPC 表。這只新增該有限鏈的 D3／E2，不改變「DQ3.EXE 未被完整語意解讀」的總結。
