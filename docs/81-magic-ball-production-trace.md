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

## CTY30 後續路線校正

先前把 CTY30 section 2 的入口／出口不連通解讀成「可能缺落穴或傳送機制」，這是錯誤假設。
原始資料直接給出答案：

- CTY30 使用 BLKBM5。
- section 2 的 `(6,11..12)`、`(14,11..12)`、`(22,11..12)` 是 tile 14–17。
- BLKBM5 對這些 tile 的 attr 是 `0x0041`：bit0 阻擋、bits6–7 表示 tier-1 鎖門。
- 每格 hi-map low 5 bits 都是 9；既有原版開門 consumer 會把門片改成 tile 9。
- 從入口 `(15,2)` 走上方橫廊到左門，使用仍持有的盜賊鑰匙開 `(6,11..12)` 後，
  出口 `(4,36)` 可由正常碰撞路徑抵達。
- 跨入 CTY31 section 1 後，入口 `(2,4)` 與通往 section 0 的出口 `(8,4)` 之間另有
  BLKBM1 attr `0x0041` 的 tier-1 門 `(4,4)`；同樣須以正式「調查」開啟。
- DQ3.EXE file `0x49d1..0x49f2` 在跨 CTY 時寫新 CTY/section，將 CTY index×4 加到
  `DGROUP 0x748`，再把該 `cty_loc` 的 X/Y 寫入 `[0x4f2f]/[0x4f31]`。remake 已同步
  更新 remembered-world 座標，故 CTY31 出洞會落在 `(45,70)`，不再錯回 CTY30
  `(168,164)`。

production trace 現已用正式「調查」命令開門，經 CTY31 正常出洞，再步行進入 CTY02
羅馬利亞並完成 save/load。沒有使用 debug 位移或直接切 section。

現行抵達畫面：
[`romaly_arrival.png`](../dq3_remake_ebitan/docs/romaly_arrival.png)。
