# 175 — 龍女王光之珠容量交易規格（2026-08-22）

## 玩家阻塞點

正式主線已越過黃寶珠、六珠祭壇、巴拉摩斯與王城事件，抵達 CTY67 龍女王。對話後
得到 `party item=false, flag0x4e=true, flag0x19=false`：光之珠沒有寫入，兩個不可逆
旗標也沒有推進。

這與原版共同物品 writer 在全隊滿格時失敗的語意一致。阻塞發生在終盤正式補給之後，
不能藉由放寬背包或不經 writer 直接設定光之珠／旗標處理。

## RE 證據

- [`docs/76`](76-r5-endgame-realignment.md)：CTY67 section0 NPC `(14,24)` 使用 handler52；
  IDA `sub_15E02` 在取得 item `0x65` 後 CLEAR flag `0x4e`、SET flag `0x19`，並播放
  D3TXT06 record71。
- [`docs/161`](161-common-party-item-grant-writer-spec.md)：fresh IDA direct xref 列出
  `sub_15E02` 在 `0x15E13` 寫 `DS:2593=0x65` 並呼叫共同 writer；writer 依勇者到同伴
  每人八格掃描，滿格不寫。
- 本輪重新產生的 IDA Pro 9.4＋IDAPython sidecar：
  `/tmp/dq3-yellow-orb-ida-20260822/party-item-grant-callers.json`。輸入
  `assets_raw/DQ3.EXE` 115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；位址為
  IDA linear，保留原始函式名、call bytes、附近 raw window 與推論等級。

推論等級：handler52 的 item／旗標順序及共同 writer caller 為 **confirmed**；本輪
runtime 的滿格失敗為 **confirmed runtime**。原版逐框對話／音效仍不由本規格提升為 V3。

## Production trace 規格

1. 從巴拉摩斯合法 save/load checkpoint 經王城與正式拉米亞飛行抵達 CTY67。
2. 與龍女王談話前，全隊至少預留一個合法空格：優先正式丟棄藥草，其次只丟
   ITEM.DAT 證明為裝備且未穿戴的備品。
3. 不丟劇情物品、不修改 pack 容量、不直接寫 inventory／flag。
4. 由 handler52 正常對話觸發；驗收必須查全隊 `hasPartyItem(0x65)`，因原版 writer
   允許光之珠合法落在同伴。
5. 只有光之珠落在勇者時，才以正式「給予」移交同伴；若 writer 已寫入同伴，不做
   多餘搬運。索瑪 consumer 本來就掃全隊。
6. 成功交易必須同時滿足 item存在、flag `0x4e` clear、flag `0x19` set，並繼續正常
   玩家路徑至下層世界。

## 驗收界線

- 完整 boot production trace 越過 CTY67，光之珠仍由正式交易取得並供索瑪弱化使用。
- 全隊任何角色不得超出八格；安全集合不足時失敗即關閉並另立 blocker。
- 本切片不更改 production handler、Boss 數值、戰鬥策略或 game-pack。
