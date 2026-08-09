# 蓋亞之劍火山事件：IDA 證據與待辦

> 狀態：2026-08-09 已落地 production JSON／Go，並由 boot 起的正式輸入 trace
> 閉合「蓋亞之劍→尼羅肯特洞窟→CTY64 銀寶珠」及其後 CTY83／P4–P6／THE END；主線
> campaign 為 E3，runtime 截圖與音效仍是 V1／V-gap，未宣稱逐畫面 V3。

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

- `DGROUP:3B96`（file `0x19cd6`）是使用當幀的 direct viewport table；
  `DGROUP:3A54`（file `0x19b94`）是 world-state reload consumer 的 persistent table。
  IDA `sub_143C9` 的三個 far helper（`sub_2074E`、`sub_20754`、`sub_20770`）只處理計時／
  動畫／圖形資源，沒有寫世界碰撞資料；兩張 table 不能合併。輸入、工具版本與位址基準
  詳見 Docker sidecar `/tmp/dq3-map-sidecar5/map.txt`、`/tmp/dq3-far-sidecar5/far.txt`。

以上均保留原始位址／運算元，沒有用 rename 取代定位。入口、item table、座標比較、
world-state writer 與呼叫順序標為 `confirmed static`；「八個物件代表熔岩／火山動畫」及
patch 的玩家可見意義，在原版實機對拍前只標為 `strong`。

## 推翻的舊近似

現行 `game/itemuse.go` 的「任何地表使用 item `0x0f` 即寫 remake flag `0x32`」不符合原版：

- 原版有唯一成功座標 `(65,109)`；
- persistent writer 是 `DGROUP 0x4f44 bit0x20`，不是自造 story flag；
- 原版還有暫態動畫、地圖 patch 與讀檔／重建 consumer。

在下列證據閉合前，不得只把座標改成 `(65,109)` 就宣稱修正完成。

## 已完成的 production slice

1. `item_use_effects` 已保存唯一座標 `(65,109)`、world-state `0x20`、動畫參數、
   persistent `map_patch`（3A54）與 `direct_map_patch`（3B96）；loader／schema／EXE parity
   test 已拒絕缺值與錯誤範圍。
2. save/load 只重建 persistent table；成功使用當幀另套 direct table，避免把兩個原版
   consumer 誤合併。
3. 正式 trace 走原始地表入口 `(43,140)` → CTY56 sec1 → sec2 → sec3 → sec2 → sec3
   → CTY56 sec0 `(3,2)` → 北出口 `(43,138)` → CTY64 `(51,133)`；沒有座標注入或直接
   `traceAdventureTravelToCty` 直航。
4. CTY64 handler49 的銀寶珠交易使用 party-order inventory consumer；save/load 與重複
   對話皆已驗收。完整 `TestOpeningProductionInputTrace` 通過。

## 後續非主線 blocker

- `TestOpeningProductionInputTrace` 已從新遊戲正式輸入通過 CTY83、P4–P6、索瑪城三連戰、
  王城冊封與 `THE END`；本文件的銀寶珠事件不再代表尚未完成 campaign。
- 原版影片已確認「火山後先走下方陸路、穿尼羅肯特洞窟，再向東北神殿」；目前仍需把
  動畫／音效與同狀態 runtime 對拍由 V1 提升至 V3。網路攻略只作路線定位，不能代替
  EXE／實機規格。
