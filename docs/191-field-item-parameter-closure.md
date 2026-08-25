# 190 — 野外道具參數閉合：藥草、聖水、祈禱之戒

日期：2026-08-25。輸入為 `assets_raw/DQ3.EXE`，大小 115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；主要工具為
IDA Pro 9.4。可重建匯出器是 `tools/ida_dump_antidote_item.py`，sidecar 只放 `/tmp`。

| 道具 | 原始定位與資料流 | 結論 | 等級 |
|---|---|---|---|
| 藥草 `0x41` | table linear `0x2843A` → handler `0x13DB9`，`BP=0x1E` → helper `0x146EB` | rejection sampling 產生 30–39 HP，再封頂 | confirmed |
| 聖水 `0x44` | table `0x28440` → handler `0x13DDD`；field writer `DS:0x52F6=0x28` | 野外驅敵 40 步；舊 64 為無 EXE 證據的近似 | confirmed |
| 祈禱之戒 `0x48` | table `0x28448` → handler `0x13E40`，`BP=0x16` → helper `0x1473F` | rejection sampling 22–31 MP；`RNG(256) <= 0x40` 才損壞 | confirmed |

位址皆為 IDA linear；原始函式名、bytes、xref 與輸入雜湊保留在 sidecar，本文語意不取代原始定位。
現行 `events.json` 以有限 primitive 保存三組參數，正式玩家路徑不再讀 classic／舊 C 數值。
`sub_146EB`／`sub_1473F` 都先以 `DS:0x0741` 取得所選隊員，再寫 `DS:0x259C` 給角色結構
consumer；藥草的 `sub_14CF9` 則在治療檢查前移除持有者 `+0x3A` 的原始物品格。因此野外
藥草與祈禱之戒現在都走正式隊員選擇 modal，持有者與效果目標互相獨立；藥草即使選到死亡
或滿血目標也不退還。此項完成 D3→E2，但逐 record 成功／失敗文字仍未升格 V3。
