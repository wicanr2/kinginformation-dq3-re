# 船直接進城／停泊／出城復船交易

狀態：2026-08-23，D2／E2。這是載具狀態交易，不是船碰撞或取得船的重做；航行 attr
consumer 見 `docs/48`，授船見 `docs/50`，動態幽靈船入口見 `docs/104`。

## 問題

現行一般 world entrance 在船仍為 `shipAboard=true` 時直接呼叫 `enterTownCty`，只保存玩家
world 座標，沒有像已閉合的 tracked world object 一樣把船停到入口並切回徒步。結果是城內
仍殘留乘船狀態；下層出城 restoration 又被 `layer==0` 限制，會把玩家放回水格卻無法復船。

## 證據與推論等級

- `DQ3.EXE` SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `sub_12758` linear `0x127d6..0x1282f` 的 mode1 attr consumer 已證實航行／上岸／阻擋；
  詳見 `docs/48`。
- 原版 tracked world entry 的 `entry_vehicle_mode=disembark_ship_if_aboard` 已由
  `docs/104` 閉合，現行 runtime 亦保存入口座標、停泊並下船。
- 將同一有限交易套用到一般 CTY world entrance 為 `strong`：它閉合相同 mode1 狀態，且
  上下層 runtime 測試可重播；本切片未取得一條獨立的普通 CTY caller→writer IDA 閉環，
  因此不冒稱 D3／原版逐指令 exact。

## Remake 交易

只有城鎮載入成功後才提交：

1. 若目前在船上，把 `shipX/shipY` 寫成入口 world tile。
2. 清 `shipAboard`，城內使用徒步狀態。
3. 保留 `overPx/overPy`，正常載入 CTY spawn。
4. 出城回到同一 world tile；只要 owned 且座標等於停泊船即恢復 `shipAboard`。
5. 第 4 步同時適用 layer0 與 layer1，不把上層限制錯套到下層世界。

針對性測試以正式 `enterTownCty`／`exitTown` 分別抽測地表 CTY38 與下層 CTY85，並與
既有 mode1／tracked world object tests 一起通過。本批不宣稱逐畫面 V3。
