# 變身杖、船員骨頭、幽靈船與愛的回憶 production trace

更新：2026-08-02

## 結論

本切片已從新遊戲正式輸入路徑閉合：CTY54 拒絕／接受交換 → 船員骨頭三段定位文字 →
地表 80×80 buffer 驅動的幽靈船位置更新 → 動態座標入口 CTY36 → 船內原始轉場／鑰匙門 →
section1 `(18,55)` 取得愛的回憶 `0x64` → save/load。

先前把 `[0x5053/0x5055]=(150,90)` 寫進 Go 的 `overPx/overPy` 是錯誤：兩者是幽靈船物件
座標，不是玩家離城返回座標。schema `0.1.20` 已把兩種狀態分離；幽靈船位置獨立存檔，
CTY54 返回點保持原值。CTY36 愛的回憶也已從 Go `treasures` table 遷入 game-pack JSON。

下一個主線 blocker 是地表 `(76,54)` 的歐里比雅詛咒事件；愛的回憶 item-use handler
本身只顯示 record596，真正的旗標交易由該座標 consumer 自動執行。

## 證據身分

- 輸入：`assets_raw/DQ3.EXE`，115282 bytes。
- SHA-256：`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 主要工具：IDA Pro 9.4；位址空間為 IDA linear。
- 本載入映像換算：`file = IDA linear - 0xec90`。
- 原始資料：`CTY54.DAT`、`CTY36.DAT`、`D3TXT06.TXT` records56–62、
  `D3TXT00.TXT` records596–598、740–744。
- IDA 工作資料庫與 sidecar 位於 `/tmp/dq3-ghostship-ida/`，不進 Git；語意註記不取代
  原始位址、bytes 或運算元。

## writer → state → consumer 閉環

| 原始定位 | 原始行為 | 附加語意 | 等級 |
|---|---|---|---|
| linear `0x15c23..0x15c89` | flag0x43、item0x62、兩組 choice、records56–59/62 | 無杖詢問與有杖交換是兩個分支 | confirmed |
| linear `0x15c8b..0x15ca9` | `[si]=0x63`、`[5053]=150`、`[5055]=90`、`[4f44]|=4`、clear0x43 | 原格換成船員骨頭並啟用幽靈船物件 | confirmed |
| linear `0x12d94..0x12da7` | `tile=((DS:0b5a&3)<<1)+0x33`，讀 `[5053/5055]` | 幽靈船 tile 由原版 RNG state 低兩位選 `0x33/35/37/39`，不是固定動畫 timer | confirmed |
| linear `0x12d17..0x12d5b` | 將世界座標換成 active buffer 相對座標；buffer 固定 `80×80` | 只在物件位於目前 world buffer 時覆寫 tile | confirmed |
| linear `0x12db5..0x12dde` | `RNG(0x12)`，查 DGROUP `0x3a0c` 18 組座標；候選在 buffer 內才回寫 `[5053/5055]` | 幽靈船不是每幀亂走；只在 world buffer rebuild 且舊位置不在 buffer 時嘗試一次 relocation | confirmed |
| linear `0x12e1e..0x12eb7` | X offset 保持10–69、Y offset保持8–71；超界以玩家座標減40重建 | world buffer 滾動 gate 與 relocation 呼叫時機 | confirmed |
| linear `0x1413f..0x14197` | records740–744、兩次 absolute delta 寫 DS2593 | item0x63 依序顯示東／西 X 距離與南／北 Y 距離，不消耗 | confirmed |
| linear `0x126a5..0x126bb` | 玩家 X/Y 等於 `[5053/5055]` 時 `BP=0x24` | 動態座標直接進 CTY36 | confirmed |
| CTY36 sec1 file `0x13cb` | `01 64 00 8b`；tile `(18,55)` subid0 | type1 寶箱給愛的回憶0x64並 clear present flag0x8b | confirmed |
| linear `0x14198..0x141a0` | item0x64 只顯示 global record596 | 手動使用不交易詛咒旗標 | confirmed |
| linear `0x126bd..0x1272b` | `(76,54)`、flag0x35、搜尋0x64、records597/598 | 有回憶時自動 clear0x35；無回憶時播放強制水流動作 | 後續已於 `docs/105` 閉合 |

DGROUP `0x3a0c` 的 18 組候選保留重複項，因重複即原版抽樣權重：

```text
(150,90) (170,110) (163,100) (156,96) (170,100) (169,90)
(153,109) (150,110) (165,98) (100,177) (120,179) (110,180)
(109,185) (118,159) (120,173) (153,109) (150,110) (130,188)
```

## game-pack 與驗收

- schema `0.1.20`／content `0.1.24` 新增 `tracked_world_objects`。world-state mask、入口 CTY、
  RNG tile table、80×80 buffer gate、18 組 relocation candidates、tracker item 與五個 text ID
  全在 JSON；Go 只保留有限 loader、window/rebuild primitive 與入口交易。
- `choice_item_exchange_events.activate_world_object` 取代錯誤的 `success_world_position`；
  交換成功不再改玩家 `overPx/overPy`。
- 存檔新增具 stable ID 的動態世界物件座標；舊 `0.1.19` world-state 存檔只能由明確
  activation 記錄遷移，不猜合理預設。
- component tests 鎖定拒絕／接受、玩家與物件座標分離、原版東西南北記錄、18 候選
  buffer gate、動態入口、CTY36 raw event 與愛的回憶 present flag。
- `TestOpeningProductionInputTrace` 從新遊戲只送正式輸入，經船員骨頭、出城、登船、航行、
  CTY36 原始 portal、鑰匙門與調查取得愛的回憶，並完成 save/load。
- runtime V1 圖：[`幽靈船地表物件`](img/ghost_ship_overworld.png)、
  [`愛的回憶寶箱`](img/ghost_ship_loves_memory_chest.png)。本機完整實況已用於路線定位；
  原版同狀態畫格與 palette 尚未達 V2/V3，故兩圖不能宣稱視覺 parity。

## 後續狀態

`(76,54)` 的五次強制移動、records597／598、flag0x35、道具不消耗、CTY55 與蓋亞之劍
均已由下一條正式切片閉合；見 `docs/105`。手動使用 item0x64 的 record596 仍只是提示，
不是解除詛咒 transaction。
