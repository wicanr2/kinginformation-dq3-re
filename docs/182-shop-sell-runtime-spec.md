# 商店賣出、售價與裝備交易 runtime 規格

狀態：2026-08-23；一般武防店／道具店賣出為 D3／E2。瑪依拉 `DGROUP 0x0b62=1`
的特殊店鋪入口保留為「選單／玩家可見支線」切片，不把一般商店完成度冒稱為該支線 E3。

## 證據

- `DQ3.EXE`：115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- IDA Pro 9.4；16-bit DOS linear address；原檔唯讀、原始名稱保留。
- 可重生 exporter：`tools/ida_dump_shop_sell.py`；sidecar 不入 Git。

`0x1776f..0x1793a` 的賣出鏈已證實：

1. 選隊員後列出該角色的八個 item words；選定後保存原始 slot pointer。
2. 一般售價讀 `ITEM +2/+3`，`0x17873 shr ax,1` 保存二分之一，再次右移得四分之一，
   相加後寫 `DGROUP 0x2593`，也就是逐步整數截斷的 `price/2 + price/4`。
3. 原價 0 在 `0x1786b..0x17870` 直接拒絕，不進確認。
4. 玩家確認後，`0x1789a..0x1789e` 才把選中 item word 寫 `0x00ff`，接著
   `sub_1895C` 加錢；取消不改物品、不加錢。
5. item word 可含裝備標記，所以裝備中物品也走同一 slot transaction；remake 的分離模型
   以清 equipment slot 表示，能力值由既有即時計算自然更新。

推論等級：`confirmed`。原始 bytes parity test 鎖定售價與 remove-before-credit 序列。

## 特殊 selector 的停止線

全程式 operand audit 只找到 `DGROUP 0x0b62` writer：`0x16315` 設 1，呼叫
`sub_17034(BX=0)` 後於 `0x1632d` 清 0。只有這個 scripted facility mode 才令 raw `0x6d`
在 `0x1784c..0x1785e` 取得固定 `0x57e4`（22500）售價。pack 已保存此 override，且以
`shop_kind=special` 避免污染所有一般道具店；特殊 NPC 入口與後續王者之劍可見流程仍待
支線切片接線，完成前不得宣稱該支線 E3。

## Remake 接線

- 售價 ratio、零價 gate 與特殊 override 位於 `facilities.json`；共用 Go 只執行有限公式。
- 正常店鋪貨架按取消進入賣出選人，再選個人物品並二次確認；取消任一頁不交易。
- 個人未裝備物品與四個裝備槽皆可出售；成功才移除並加錢。
- 玩家可見名稱仍取原始 D3TXT00 item record，沒有在 production Go 新增文字。

本切片不升格逐窗幾何 V3，也不以單元測試取代瑪依拉特殊 NPC 支線。
