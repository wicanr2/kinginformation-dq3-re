# 161 — 共用劇情物品取得 writer 規格（2026-08-22）

## 問題

戰鬥掉落已使用原版每人八格 writer，但多個劇情、寶箱與 NPC 獎勵仍直接
`append(g.inventory, ...)`。正式 trace 在八頭大蛇第一戰前觀測勇者為 9 件未裝備物品、
加兩件裝備，共 11 格；這推翻「只有戰鬥掉落需要容量 gate」的舊實作。

## 輸入與非破壞性 IDA 證據

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- IDA Pro 9.4＋IDAPython；image `ida-pro-9.4-ver3:py312-x11-v3`。
- exporter：`tools/ida_dump_party_item_grant_callers.py`；sidecar：
  `/tmp/dq3-ida-party-grant-20260822/out/party-item-grant-callers.json`。
- 位址：IDA linear；`logical=linear-0x10000`、`file=logical+0x1370`。
- 匯出保留原始函式名、xref、call bytes 與 caller raw window；未 rename、未修改 EXE／IDA
  database。

IDA direct xref 證實 `sub_1684E`（音效後落入 `sub_16856`）並非戰鬥專用。caller 包含：

| caller | 寫入 `DS:2593` | 原始用途定位 |
|---|---:|---|
| `sub_1024C` `0x10278..0x102A5` | `0,1,1,3,0x1f,0x1f` | 阿里阿罕國王六件起始裝備 |
| raw `0x142D2..0x142D8` | `0x74` | 劇情取得物 |
| `sub_1521F` `0x15245` | `0x58` | 魔法球 |
| `sub_1537F` `0x153BA` | `0x55` | 盜賊鑰匙 |
| `sub_15719` `0x15750` | `0x5c` | 劇情取得物 |
| `sub_15771` `0x157E5` | `0x5b` | 劇情取得物 |
| `sub_1589F` `0x158BD` | `0x4b` | 劇情取得物 |
| `sub_1593A` `0x15960` | `0x66` | 劇情取得物 |
| `sub_15D70` `0x15D96` | `0x6b` | 劇情取得物 |
| `sub_15E02` `0x15E13` | `0x65` | 光之珠 |
| `sub_16379` `0x1639B` | `0x73` | 雲雨之杖 |
| `sub_16646` `0x1666C` | `0x10` | 劇情取得物 |
| `sub_18C0F` `0x18C33` | `0x4c` | 寶箱／場景取得物 |
| `sub_1C425` `0x1C537` | runtime `AX` | 戰鬥掉落 |

`sub_16898` 只是直接呼叫 `sub_16856`；`sub_1684E` 先呼叫 `sub_20770` 再 fall-through。
兩者最終都依隊長至同伴掃 `+0x3a` 八個 word，成功寫第一個 `0x00ff`，全隊滿格設
`DS:0726=1` 且不寫。以上 caller、writer 與 raw item 為 **confirmed**；尚未逐一閉合表中
未命名劇情的完整旗標交易，維持定位用途，不由本規格猜名。

## Remake 規格

1. production 中所有「取得新物品」入口必須透過 `grantPartyItem`；不得直接 append 勇者。
2. 成功後才清 present flag、顯示取得通知或推進不可逆交易；全隊滿格時 fail closed。
3. 國王六件獎勵按原始順序逐件呼叫 writer；正常四人初始容量足夠。若異常資料導致中途
   滿格，不得越界追加。
4. 既有物品交換的「移除來源 owner」另立 RE；本規格不把 hero-only `removeItems` 批次改成
   party search，以免猜錯交換持有者。
5. 測試鎖定每名含裝備八格、第一空位順序，以及代表 caller bytes；正式 trace 中任何角色
   不得再超過 pack capacity。

## 驗收

- component：寶箱／NPC reward 在勇者滿格時寫入同伴；全隊滿格時不清旗標、不冒稱取得。
- production：八頭大蛇第一戰前逐人 `itemCount <= 8`，草薙大劍仍由正常掉落取得。
- current campaign 從標題重播；下一個 blocker 另立 spec。
