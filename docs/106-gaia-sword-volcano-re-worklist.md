# 蓋亞之劍火山事件：IDA 證據與待辦

> 狀態：2026-08-02 交接；只完成原版靜態追蹤，尚未落地 production JSON／Go，亦尚未完成
> DOSBox 玩家可見閉合。

## 本輪已確認

- 輸入：`work/DQ3.EXE`，115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：Docker 內 IDA Pro 9.4；位址口徑為
  `logical = IDA linear - 0x10000`、`file = IDA linear - 0xec90`。
- 道具選單 dispatcher `(logical) 0x3ccf..0x3d16` 對 raw item `<0x41` 搜尋
  `DGROUP 0x3666`。該表原始項目只有 `0x0f,0xff`，item `0x0f` 的 handler word 位於
  `DGROUP 0x3668`，值為 `(logical) 0x3d65`。
- handler `(logical) 0x3d65..0x3d7c` 只在世界座標
  `DGROUP 0x4f2f=0x41`、`DGROUP 0x4f31=0x6d`，即 `(65,109)` 成功；不符時跳至
  `(logical) 0x4307` 的共用「不能使用道具」路徑。
- 成功路徑依序呼叫 `(logical) 0x4ab5` 與 `(logical) 0x43c9`。後者建立八個暫態物件、
  執行 30 輪更新／繪製，之後以 `SI=DGROUP 0x3b96`、`AX=0x6f`、`DI=0x3d` 呼叫
  `(logical) 0x4d31` 地表 patch writer，接著 OR `DGROUP 0x4f44` bit `0x20`，並呼叫
  `(logical) 0x2511` 的 world-state 重建路徑。
- handler 沒有清除所選 inventory slot；目前靜態證據支持蓋亞之劍成功後不消耗。

以上均保留原始位址／運算元，沒有用 rename 取代定位。入口、item table、座標比較、
world-state writer 與呼叫順序標為 `confirmed static`；「八個物件代表熔岩／火山動畫」及
patch 的玩家可見意義，在原版實機對拍前只標為 `strong`。

## 推翻的舊近似

現行 `game/itemuse.go` 的「任何地表使用 item `0x0f` 即寫 remake flag `0x32`」不符合原版：

- 原版有唯一成功座標 `(65,109)`；
- persistent writer 是 `DGROUP 0x4f44 bit0x20`，不是自造 story flag；
- 原版還有暫態動畫、地圖 patch 與讀檔／重建 consumer。

在下列證據閉合前，不得只把座標改成 `(65,109)` 就宣稱修正完成。

## 下一輪工作順序

1. 由 `(logical) 0x4d31` 閉合 patch writer 的寬、高、row stride 與
   `DGROUP 0x3b96` 完整原始 tile bytes。
2. 追 `(logical) 0x2511` 對 world-state bit `0x20` 的讀檔／場景重建 consumer，確認相同
   patch 會在 save/load 後重套。
3. 從本機原版影片或 DOSBox 取得使用前、動畫中、使用後同狀態畫面，確認玩家站位、火山／
   熔岩變化、訊息、聲音與可通行區。
4. 將 raw item、唯一座標、world-state mask、完整 patch、動畫參數與不消耗語意移入有限
   `item_use_effects` JSON；Go 只擴充共用具名 primitive，不保留 DQ3 專屬 fallback。
5. 加 EXE bytes／table parity、錯誤座標 fail-closed、成功 transaction、重播、save/load
   與正式道具選單測試。
6. 從 CTY55 合法 checkpoint 延長 boot production trace，經火山進尼羅肯特，正常取得
   銀寶珠；事件後仍須能返回 CTY83 完成黃寶珠 E3。
