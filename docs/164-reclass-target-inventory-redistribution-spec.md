# 164 — 達瑪轉職目標的正式背包重分配規格（2026-08-22）

## 問題

共用 grant writer 讓《領悟之書》合法落入有空格的角色後，正式 trace 必須把它交給預定
轉職的第一名同伴。該同伴此時有五件未裝備物品與三件裝備，已滿原版八格，而且沒有
已證實可丟棄的藥草。直接丟棄舊裝備會猜測 `docs/158` 尚未閉合的完整不可丟 metadata。

## 證據與推論等級

- `docs/92`：handler39 只檢查所選角色自己的八格，賢者轉職需要該角色持有 `0x4a`。
  **confirmed**。
- `docs/158`：原版 rec421 給予可由任一 owner 移交到任一隊員；不同 owner 時受八格容量
  限制，成功才清來源，同 owner 只重排。**confirmed**。
- `docs/157`：裝備與未裝備物合計八格。**confirmed**。
- 本輪 runtime 顯示目標未裝備 `[0x1e,0x22,0x51,0x56,0x5b]` 且有三件裝備；這只是
  remake replay 狀態，不用來推論原版道具可否丟棄。

上述 owner／writer／consumer 已由 IDA Pro 9.4 sidecar 與 bytes 閉合，本切片不重新分析
executable，也不提升 `docs/158` 的未知 drop metadata。

## 規格

1. 若領悟之書已在轉職目標手上，不做無意義同 owner 重排。
2. 若目標有空格，經正式 owner selector／rec421／target selector 直接給予。
3. 若目標滿格，先在其他角色找一件藥草，以正式丟棄騰出一格；再把目標的一件未裝備
   物品正式移交到該空格，最後正式把領悟之書交給目標。
4. 不直接改 slice、不丟棄未證實道具、不改裝備、不增加容量。
5. 每次跨 owner 交易鎖來源 raw ID 數量減一、目標加一；不能要求來源完全不含重複 ID。

## 驗收

- 目標最後持有 `0x4a`，全隊仍只持有一本。
- 每名角色裝備加未裝備物不超過 pack `personal_inventory_slots`。
- handler39 由正式對話、選人、選賢者與確認輸入完成，並通過 save/load。
