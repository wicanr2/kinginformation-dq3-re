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
| 敵方共用行動 dispatcher | confirmed：狀態／逃跑／物理／action-mask 主要分支與 raw 欄位；每個存活 actor 每回合恰好一筆 | 已資料化；single-entry queue 已鎖定 | 不涵蓋全部 action 語意、逐動作 frame／PCM timing；本 EXE 無 Boss repeat-N，見 `docs/148` |
| 戰鬥逃跑音效 | confirmed：玩家 cue13、敵人 cue21，兩條皆經 `sub_208E2` 完成等待 | pack／Battle D3／E2；VOC source-duration 輸入閘門 | 不等於硬體 wall-clock、重取樣波形或逐幀 V3，見 `docs/149` |
| 玩家一般物理攻擊音效 | confirmed：cue6 → record0x14a → wait → cue11 → wait → damage consumer | 有序 pack sequence／Battle D3／E2 | 只限 `sub_1B4F6` 正常分支；不涵蓋 cue6 其他 caller或 critical，見 `docs/150` |
| 一般物理結果與敵方攻擊 | confirmed：敵攻 cue4／record331；命中 cue1／record332 或 333；miss cue3／record335；死亡 record336／357 | pack role／Battle D3／E2；VOC source-duration gate | 不涵蓋 cue9 特殊分支、critical、咒文傷害、硬體 wall-clock 或逐幀 V3，見 `docs/151` |
| formation | confirmed raw position formula；pixel surface 為 strong | 已接線 | 不得稱逐像素 V3 |
| 日夜 | confirmed clock `0..0xef`、`0x78` boundary、12-byte palette selector | 五個原始 bank／12 段 selector D3／E2 | 尚無同 clock DOSBox DAC 畫面，不得稱 V3 |
| 五個 story flag | confirmed 原版 writer、event chain、generic NPC consumer | 有限 transaction E2 | E2 不等於 direct GET 存在或玩家可見 V3 |
| 必經 Boss 背景 | selector／archive-page-palette 證據依 `docs/128`、`129` | V1；八頭大蛇 near-state V2 | 不得概括為全部 Boss V3 |
| 正式主線 | `docs/153` 確認 RNG 次序；`docs/154..178` 閉合其後玩家路線與交易 blocker | 最新正式 `InputState` trace 以 115.629 秒由標題抵達 THE END，campaign E3 | 不等於全遊戲每畫面、音效、時序皆原版 parity |

## 仍未知或未閉合

- Boss 同回合 repeat-N 的早期 `unknown` 已由 `docs/148` 推翻：正式與 alternate runner、
  第三個 queue-builder caller、全部 actor-consumer direct xref 與 raw pointer 候選共同證實
  本 EXE 每個存活 actor 每回合恰好一筆，沒有待補的 repeat-N 資料來源。
- 玩家／敵人成功逃跑 cue 13／21、玩家一般物理 cue 6 → cue 11，以及 `docs/151` 的
  敵方 attack cue4、雙方共同 damage cue1、physical miss cue3 與個別死亡 record
  都已閉合；cue9／action3／持久麻痺亦由後續 `docs/152` 閉合。仍未知的是其他 critical、
  咒文 damage、每個動作實際 SHP frame、
  PCM 波形／host wall-clock 與玩家可見逐幀停頓。cue 6 在其他 call-site 出現，不能因
  本次有限閉合而獲得全域「攻擊音效」名稱。
- 日夜五個原始 palette bank 與 12 段 selector 已接線；同 clock DOSBox DAC／畫面對拍
  仍未閉合，因此玩家可見結果尚未達 V3。
- generic terrain → battle background 全表未閉合；沒有新玩家可見 discrepancy 時不追。
- formation 的 raw 公式已接線，但所有編隊的逐像素位置並未以同狀態原版畫面證實。
- 抗性 raw descriptor／threshold 已閉合；全部中文效果名稱與語意不可由 raw 值外推。
- 怪物 bit3→持久麻痺特殊物理、bit38→poison 與 bit41→全隊睡眠已由原版
  remap／handler 閉合；其餘 action bit
  並未逐一閉合。
  現行 `enemyAction` 對未由 pack 定義的 bit 仍保留歷史
  `MonsterSpellRec`→`spell.GetDef` 相容近似；這是 remake 行為，不是原版 exact 證據，
  也不能用來產生「完整怪物咒文表」。本輪只校正斷言，不擅自改變戰鬥規則。
- temporary solo challenge 的 completed writer 尚未由原版證實。
- 對話保留控制碼的部分語意與怪物記錄 `+0x27` 仍未閉合。教會 poison／curse 的舊合併
  敘述已由 `docs/137` 拆開訂正；後續 `docs/147` 又閉合 ITEM metadata → item word
  `bit0x4000` → level×100 教會移除，現行 poison 與 curse 兩條有限鏈皆為 E2。
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
- 教會／中毒後續訂正：`sub_19AD6`／`sub_199DC` 證實怪物 mask bit18..47 需經 remap，
  bit38 → action0x32 → logical0xa307 毒氣 handler；`sub_1964E` 證實每個合法移動步
  對中毒成員加 1 傷害，`sub_1712B` 固定 5G 清 bit0x40。這推翻「mask 全部直接對
  玩家 spell descriptor」及「poison／curse 都是 +0x38 flag」兩項斷言。後續已用
  當時以 schema `0.1.40`／content `0.1.45` 接成 E2；完整新遊戲 trace 後續已重新回綠，但該結果
  不把逐畫面／音效升格為 V3。詳見 `docs/137`、`docs/147`。
- 驅毒草後續訂正：item `0x42` 經 DGROUP `0x366c` 進 logical `0x3dc3`；`sub_14CF9`
  先把命中物品槽寫成 `0x00ff`，`sub_1469F` 才判斷死亡／`+0x38 bit0x40` 並選 rec341
  或 rec364。這推翻歷史 C「沒中毒不消耗」斷言；現行 pack、選人 modal 與文字 consumer
  為 E2／V1。原版 per-character inventory ownership 仍未閉合；滿月草持久麻痺與 field
  consumer 已由後續 `docs/152` 訂正，不能再列為全域未知。
- CTY23 重播訂正：全滅時主角 `conditionPoison` 已設且背包已有 `0x42`；正式移動 helper
  現在經道具清單與選人 modal 解毒並越過單人回程。後續補給、隊伍與巴拉摩斯策略切片
  曾讓當時完整 replay 再次抵達 `THE END`；CTY23 本身不再是 blocker，但 `docs/153`
  RNG 訂正後的現行 campaign 已於後段 CTY38→CTY70 航路重新 pending。
- 戰鬥逃跑音效後續訂正：IDA 9.4 已證實玩家成功逃跑 `cue 13`、敵人成功逃跑
  `cue 21`，兩條 caller 都在訊息後呼叫 `sub_208E2`。現行 pack／Battle 以原始 VOC
  sample duration 阻擋提前輸入，為 D3／E2；`PlaySFX` API 本身仍非阻塞，硬體
  wall-clock、重取樣波形與逐幀同步仍未閉合，詳見 `docs/149`。
- 玩家一般物理攻擊後續訂正：IDA 9.4 已證實 `sub_1B4F6` 正常分支依序 dispatch
  cue 6、顯示 D3TXT record `0x14a`、等待完成、dispatch cue 11、再等待後進入共同傷害
  consumer。現行 pack／Battle 保存並執行這個有序序列，為 D3／E2；文本原先將 record330
  誤註為 enemy action consumer，已更正為 player normal physical consumer。詳見 `docs/150`。
- monster89 後續訂正：IDA 9.4 與 `D3MNS.DAT` raw record 證實其 action gate 為 0、mask
  全零；戰敗不是未實作的特殊 Boss action。CTY43 type3 block 也沒有日／夜 NPC consumer，
  不是可用教會。正式路線改由 CTY43 旅店＋CTY2 教會，已跨過 monster89、幽靈船與愛的
  回憶、CTY36、蓋亞之劍與銀寶珠；怪物 46 訂正與後續 trace 證據見 `docs/141`。
