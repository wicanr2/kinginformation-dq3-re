# 156 — 商店購買選人與每人八格道具容量規格（2026-08-22）

## 問題與停止線

正式主線越過 CTY55 後，失敗診斷顯示勇者 `inventory` 已累積數百個 raw `0x41`（藥草）。
raw `0x41` 身分本身正確；異常來自 remake 商店成功時無條件 `append`，沒有使用其他任務
獎勵已採用的個人容量 consumer。本切片只閉合商店購買的選人、容量、寫入與扣款順序；
不研究賣出、裝備推薦、價格表或商店逐框版面。

## 輸入、工具與位址契約

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- IDA Pro 9.4＋IDAPython；image `ida-pro-9.4-ver3:py312-x11-v2`。
- exporter：`tools/ida_dump_shop_inventory_capacity.py`。
- sidecar：`work/shop-inventory-capacity-20260822.json`（研究輸出，不加入 Git）。
- 位址為 IDA linear；`file=linear-0xec90`。原始名稱、bytes 與位址保留，未 rename binary。

## 已證實的原版交易

`sub_17747`（IDA `0x17747`／file `0x8ab7`）讀取先前選定的 1-based 隊員索引
`DS:0722`，減一、乘二，再由 `DS:4F15[index]` 取得該角色記錄：

```text
0x17752 add si,3Ah
0x17755 mov cx,8
0x17758 mov ax,[si]
0x1775A cmp ax,00FFh
0x1775D jz  0x17766
0x1775F inc si
0x17760 inc si
0x17761 loop 0x17758
0x17763 mov al,0FFh ; 八格皆滿
0x17766 mov ax,[2591h]
0x17769 dec ax
0x1776A mov [si],ax ; 空格寫入所選商品 raw id
0x1776C mov al,0
```

武防店 caller `0x1769C..0x176A5` 與道具店 caller `0x17B9D..0x17BA6` 都先恢復
`DS:0722`，呼叫 `sub_17747`，再比較 `AL==0xFF`。只有成功分支才載入價格
`DS:2593` 並呼叫 `sub_18906`；道具店滿格分支顯示 record `0x126`。因此原版不是把商品
無限加到勇者，也不會在指定角色滿格時先扣錢或自動轉送另一角色。

推論等級：**confirmed**（兩種商店 caller、選人索引、八格掃描、空值 sentinel、商品
writer、滿格返回與成功後扣款順序閉合）。`sub_17C31` 是道具店選單 UI consumer；逐框
幾何不在本切片停止線。

## Remake 規格

1. 選定商品後進通用隊員選擇狀態；方向／觸控選擇勇者或同伴，取消回貨架。
2. 每人的容量取 `game-pack interface.item_actions.personal_inventory_slots`；裝備與未裝備物品
   合計占用容量。不得在 Go 另寫 DQ3 專屬 8。
3. 確認目標後，只有該角色有空位且金錢足夠時才加入其 personal inventory 並扣款。
4. 目標已滿時不扣款、不加入、不自動轉送；維持選人狀態供玩家改選。缺 pack 容量時
   fail closed。
5. 商店、寶箱、NPC reward 與掉落最終都不得讓任何個人 inventory 超過契約容量；本切片
   先修正式商店 writer，其他 writer 由後續全域 audit 個別驗證。
6. production trace 的購買 helper 必須真的送出商品確認→隊員確認兩段輸入，不直接呼叫
   transaction 或擴大背包。

## 驗收

- EXE parity：鎖定 `0x17747` 的八格／`0xFF`／writer bytes，以及兩種商店 caller 的
  success-before-charge 分支。
- component：勇者有空格成功、勇者滿格不扣款、改選同伴成功、裝備占容量、取消回貨架。
- 正式主線重播：inventory 長度不得再無界成長；再判斷 monster65 是否仍為真實 blocker。
