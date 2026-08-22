# 怪力魔勝利獎勵滿載回滾規格（2026-08-22）

## 問題

正式 campaign 已擊敗 monster89，原版完成旗標 `0x22` 與白天切換也發生，但全隊每人八格
皆滿，`grantPartyItem(0x62)` 失敗，現行 `settleMirrorBattle` 仍不可逆地完成事件。這會永久
失去變化之杖，違反共用取得 writer 的「成功後才推進交易」規格。

舊 trace 另以 hero-only `hasItem(0x62)` 驗收，亦不符合已證實的隊長至同伴 writer。

## IDA Pro 9.4 非破壞性證據

- 輸入：`assets_raw/DQ3.EXE`，115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：IDA Pro 9.4＋IDAPython，image `ida-pro-9.4-ver3:py312-x11-v4`。
- exporter：`tools/ida_dump_samanosa_mirror_reward.py`。
- sidecar：`work/samanosa-mirror-reward-ida-20260822.json`（gitignored）。
- 位址：IDA linear；`logical=linear-0x10000`、`file=linear-0xec90`。
- 原始 EXE 唯讀；工作 database 只在一次性容器 `/tmp`；未 rename、未修改 binary。

`sub_14312`（IDA linear `0x14312..0x143c9`）的勝利尾端為：

```text
0x143AF call sub_1EE9B           ; 戰後取得／容量 writer 鏈
0x143B2 call sub_118B0           ; 戰後畫面／狀態更新
0x143B5 cmp  byte ptr ds:726h,1  ; 共用 writer 的全隊滿載旗標
0x143BA jz   0x14384             ; 滿載：clear 0x10、restore 0x42、return
0x143BC mov  bx,21h
0x143BF call sub_16EF4           ; 成功才 clear 0x21
0x143C2 mov  bx,22h
0x143C5 call sub_16EDF           ; 成功才 set 0x22
```

入口在 IDA `0x1436A` 先把 `DS:0726` 清零。滿載跳到 `0x14384..0x14390`，其資料流與
敗／逃分支相同：clear flag `0x10`、restore flag `0x42`、return，不寫成功旗標。這與
`docs/161` 已證實的共用八格 writer（滿載設 `DS:0726=1` 且不寫物品）形成完整
writer→full-state→branch→flag consumer 閉環。

推論等級：**confirmed**。`sub_1EE9B` 內部的每個顯示 helper 語意不是本切片所需，不另行
猜名；玩家可見交易已由 monster89 勝利、0x62 grant、`DS:0726` 分支與旗標 writer 閉合。

## Remake 規格

1. monster89 勝利時先嘗試以 `grantPartyItem(0x62)` 寫入第一個有空位的現役角色。
2. 全隊滿格時走與敗／逃相同的事件回滾：clear `0x10`、restore `0x42`、維持夜晚；不得
   clear `0x21`、set `0x22`、切白天、顯示取得通知或冒稱完成。
3. 成功寫入後才 clear `0x21`、set `0x22`、切白天並顯示 record99。
4. 所有驗收使用 `hasPartyItem(0x62)`，不得只看勇者。
5. 完整 campaign 在使用拉之鏡前，以正式 rec421「丟棄」操作至少騰出一格；不得由測試
   直接刪除背包或放寬每人八格。

## 驗收

- component：勇者滿格、同伴有空位時獎勵寫同伴並完成事件。
- component：全隊滿格時不給物品且完整回滾假王事件。
- production trace：正式預留一格、擊敗 monster89、全隊持有 `0x62`、成功旗標與白天交易
  同時成立，然後繼續至下一個玩家 blocker。
