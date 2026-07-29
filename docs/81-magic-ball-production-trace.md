# 81 — 魔法球破牆：原版 RE 與 production trace

> 2026-07-29；位址皆為 `DQ3.EXE` file offset。

## 原版契約

IDA／Capstone 交叉追蹤到 `0x58cd..0x5990` 的魔法球 handler：

- 必須在城鎮模式、CTY30 section 0。
- 玩家必須位於 Y=13、X=8 或 9；原版沒有面向條件。
- 使用後清除 story flag `0x51`，並走共通的選中道具移除路徑。
- 共通地圖事件 rebuild 位於 `0x46df..0x4747`：事件 flag 清除時，將事件 tile 改為
  `tile+1`，並清除 hi-map 的低五位 sub-id。

CTY30 section 0 的四格牆均指向 event `(type=1, param1=1, flag=0x51)`，因此精確結果是：

| 座標 | 使用前 | 使用後 |
|---|---:|---:|
| (8,14), (9,14) | 24 | 25 |
| (8,15), (9,15) | 22 | 23 |

本機完整實況約 577–582 秒也顯示玩家在兩尊像之間，以正式道具選單使用魔法球、2×2
爆煙、牆面開啟；之後的道具窗確認魔法球已消耗。

## 本輪落地與驗收

- `itemuse` 新增原版道具 `0x58`，只在上述 gate 成功時消耗。
- 場景載入與事件後都使用共通 flag-driven tile rebuild，不為四個座標硬編碼結果。
- component tests 鎖定成功、錯誤位置不消耗，以及真實 CTY30.DAT 的四格變化。
- `TestOpeningProductionInputTrace` 從新遊戲沿正常輸入取得盜賊鑰匙、雷貝魔法球，步行至
  CTY30，以正式道具選單破牆，並驗證消耗、flag、tile 與 save/load。
- 現行 runtime 圖：
  [`temptation_magic_ball_open.png`](../dq3_remake_ebitan/docs/temptation_magic_ball_open.png)。

## 尚未宣稱完成

魔法球 transaction 已達 E3，但尚未因此宣稱抵達羅馬利亞。CTY30 section 2 的入口與出口
位於不連通區域，表示仍缺原版洞窟機關／落穴等 traversal 語意。下一輪須先追該機制，再把
production trace 延伸至 CTY31 與羅馬利亞，禁止以 debug 位移或直接改 section 代替。
