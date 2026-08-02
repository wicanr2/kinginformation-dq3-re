# 船系統 RE + remake 落地(#2)

> 地表 gate 2 核心:取船後可跨海。本檔記錄「海 tile 如何辨識」的 RE 證據與 remake 設計。

## 關鍵 RE：船 mode1 的 attr consumer

overworld 走動碰撞查 `BLKBM.DAT`(tile index → attr word;`dq3_field.c` 載入,`dq3_scene_walkable`
判 `attr&1`=阻擋)。問題:海與山**步行都被擋**(都 `attr&1`),船要能走海、不能爬山,得區分兩者。

逐 tile 列 `BLKBM.DAT`(`tools/_big.py` / 臨時腳本)得指紋:

| 地形 | tile index | attr | 位元 |
|---|---|---|---|
| 草原/可走 | 0–7,18–25… | `0x0002` | bit1 only(可走)|
| **山/牆(步行擋)** | 66–74,77,48 | `0x0007` | bit0(BLK)+bit1+bit2,**無 0x20** |
| **海(步行擋、船可走)** | 88–117 | `0x0025` | bit0(BLK)+bit2+**bit5(0x20)** |
| 城鎮入口(可走觸發) | 51–58 | `0x1026` | bit5+bit1+bit12,不擋 |

上述指紋只能說明常見海面，舊結論「海的辨識只在 `attr0x20`」已被
IDA Pro 9.4 的直接 consumer 推翻。`DQ3.EXE` SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`，
`sub_12758` IDA linear `0x127d6..0x1282f`（file `0x3b46..0x3b9f`）在載具
mode1 中對一般 tile 依序分支：

1. `attr & 0x0002 == 0`：航行；
2. `attr & 0x0002 != 0 && attr & 0x0001 == 0`：呼叫上岸 consumer；
3. `attr & 0x0002 != 0 && attr & 0x0001 != 0`：阻擋。

因此 `0x25` 仍可航行、`0x0002` 可上岸、`0x0007` 會阻擋，但不能再用
`0x20` 當唯一規格。Go production 與 component test 已改用這個 confirmed
bit1／bit0 consumer。

### 上層地表 seam（confirmed）

同一函式 `sub_12758` IDA linear `0x12763..0x127ad`（file
`0x3ad3..0x3b1d`）先處理上層地表邊界。它不是普通 modulo：X 越左邊寫
`242`、越右邊寫 `2`；Y 越上邊寫 `204`、越下邊寫 `2`，保留原版捲動緩衝邊。
Go 的 production 方向輸入與 CTY58 正式航線尋路已使用此規則。下層世界的合併
座標 consumer 尚未獨立閉合，目前 fail closed，不將上層常數外推過去。

## remake 設計(`dq3_ship.{h,c}`)

純 remake 機制(原版取船劇情走 runner,早期 build 觸發未必接全,見 docs/47),用旗標流串:

- 狀態 `dq3_ship { owned, aboard, px, py, layer }`:`owned`=已取得(SHIP 里程碑 0x205);
  `aboard`=目前在船上;`px,py,layer`=未登船時船停泊的 overworld tile 與層(地表0/下層1)。
- `dq3_ship_navigable(scene,tx,ty)` 是舊 C remake 歷史實作；其 `attr0x20` 規則
  不再是原版 oracle。現行 Go 依上述 mode1 consumer 分流航行、上岸與阻擋。
- `dq3_ship_input(scene, ship, sc, layer)`:overworld 專用移動,回傳碼:
  - `0` BLOCKED · `1` WALKED(步行)· `2` SAILED(航行)· `3` DISEMBARK(上岸)· `4` BOARD(登船)
  - 步行時:目標格 == 停泊船格且 owned → 登船;否則照 `dq3_scene_walkable`。
  - 在船上:目標格可航行 → 航行;否則若是步行可走的陸地 → 上岸(船停在原水格);否則擋。
- 城鎮移動仍走 `dq3_scene_input`(船只在 overworld)。

## debug 口 + 存檔

- `ship`:給船(owned)+ 直接登船,停泊於玩家當前格。`ship:X:Y`:把船放在 (X,Y)。
- 取船里程碑由 `dq3_progress_set(DQ3_MS_SHIP)` 推進。
- 存檔:`dq3_ship` 併入 dq3_save(owned/aboard/座標/層)。

## 驗證

`tools/playthrough_check.sh`:登船→航行海格(WALKED→SAILED)、上岸(DISEMBARK)、
山格擋(船不上山)各一斷言。單元測試 `test_ship.c` 覆蓋 navigable 指紋與 board/disembark 狀態機。
