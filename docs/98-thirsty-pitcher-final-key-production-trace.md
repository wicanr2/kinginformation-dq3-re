# 乾渴壺、海中淺灘與最終鑰匙 production trace

> 更新：2026-08-02（Asia/Taipei）
> 範圍：item `0x5e`、地表 `(0x92,0x35)`、world-state `0x08`、CTY40、最終鑰匙 `0x57`

## 1. 結論

舊 Go/C prototype 的「乾渴壺可在任意地表使用，然後設自造 flag `0x33`」是錯誤近似。
原版 item handler 只在船上、地面世界 `(146,53)` 且 story flag raw `0x12` 為 clear 時成功。
成功後不消耗乾渴壺，而是設定原版 `DGROUP 0x4f44 bit0x08`，執行四輪 palette mode 4，
再把 `(144,50)` 起的 `5×4` 世界 tile 區塊換成 EXE 內固定表。新區塊在 `(146,52)` 顯出
CTY40 入口，玩家向上移動進祠堂，調查原始 event0 取得最終鑰匙 `0x57`。

## 2. 原始輸入與位址

- `DQ3.EXE` SHA-256：
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`
- item-use pointer table：`DGROUP 0x366a` entry 29，raw item `0x5e` →
  `(logical) 0x4012`、IDA linear `0x14012`、`(file) 0x5382`。
- map-state consumer：`(logical) 0x2c9c`、IDA linear `0x12c9c`。
- replacement table：`DGROUP 0x3a54`、IDA linear `0x28824`、`(file) 0x19b94`。
- `CTY40.DAT` section0 event0：`01 57 00 47`；事件格 `(18,17)`、subid0。

IDA Pro 9.4 在一次性 Docker 容器內分析；原始 EXE／DAT 唯讀。`.i64`、log 與人工語意只
存在 `/tmp` sidecar，沒有 Rename 原始符號、改原始 bytes 或加入 Git。

## 3. handler 閉合

`(logical) 0x4012..0x4062`：

```text
test story flag raw 0x12；set 時走共用失敗訊息
cmp DGROUP 0x4f3b,1            ; 必須是船
cmp DGROUP 0x4f2f,0x0092       ; world X
cmp DGROUP 0x4f31,0x0035       ; world Y
call clear field/menu windows
or  DGROUP 0x4f44,0x0008       ; persistent world state
mov DGROUP 0x25d3,4            ; palette mode
mov cx,4 / call palette cycle
copy DGROUP 0x3a54 to world (0x90,0x32), 5 bytes × 4 rows
clear story flag raw 0x12
redraw world field
```

`(logical) 0x2c9c` 是 world-state consumer：bit `0x08` set 時呼叫 `(logical) 0x2d78`，
後者用相同 `(0x90,0x32)` 與 `DGROUP 0x3a54` 重新套用 patch。因此這不是 remake milestone
flag；它必須保存為原版 world-state，讀檔後由同一資料重建地圖。

原始 20-byte row-major patch：

```text
58 58 60 58 58
58 68 42 66 58
5e 42 7c 42 5c
58 5a 5a 5a 58
```

## 4. 推論等級 ledger

| 結論 | 推論等級 | 依據 |
|---|---|---|
| raw item `0x5e` 派發至 handler `0x4012` | `proven` | item table word、handler caller 與原始 bytes |
| 成功格只接受船上 `(146,53)` | `proven` | vehicle/X/Y 三個 gate 與失敗分支 |
| 成功設定 world-state bit `0x08` 且不消耗 | `proven` | handler writer；沒有 inventory-clear caller；讀檔 consumer 重套 patch |
| 地圖替換為上述 `5×4` tiles | `proven` | `sub_14d31` 的 5-byte×4-row writer、EXE table與 world-state consumer |
| `(146,52)` 是 CTY40 入口 | `proven` | patch tile、原始 `cty_loc[40]`、正式向上輸入進 CTY40 |
| CTY40 event0 給最終鑰匙並 clear flag `0x47` | `proven` | CTY40 raw event、事件格、共用 inventory consumer與正式調查 |
| palette mode 4 的逐幀顏色已與 DOSBox 完全相同 | `unknown` | mode/cycle writer 已證實，但 remake 尚未重播四輪 palette animation |

## 5. game-pack 與正式輸入

schema `0.1.13` 的 `item_use_effects` 新增有限 `reveal_world_map_patch`：vehicle、layer、
成功座標、story flag gate、world-state mask、完整 tile patch 與原始動畫參數均在 JSON。
Go 只保留共用 gate、world-state transaction、row-major patch 與 save/load rebuild；舊
`itemuse.KindOf(0x5e)`、任意地表成功及自造 flag `0x33` 已刪除。CTY40 event0 也由 legacy
Go `treasures` table 遷入 `treasure_events`。

延長後的 `TestOpeningProductionInputTrace`：

1. 從 CTY76 乾渴壺合法 checkpoint 正常返回 CTY39、離城並登船。
2. 只送正式船舶方向輸入航行至 `(146,53)`，路徑避開其他 CTY 入口。
3. 由命令窗／道具窗／rec421「使用」選擇乾渴壺，驗證 world-state 與入口 tile。
4. 正常向上進 CTY40，由「調查」取得 `0x57`、clear `0x47`。
5. 保存、回標題讀檔；驗證 CTY40、最終鑰匙、world-state 與地圖 patch 全部保持。

## 6. 畫面證據與已知差異

| 狀態 | 圖片 | 等級 |
|---|---|---|
| 船停在原版使用格、祠堂尚未顯現 | [`thirsty_pitcher_before.png`](img/thirsty_pitcher_before.png) | V2 |
| 使用後 `5×4` patch 與入口顯現 | [`thirsty_pitcher_revealed.png`](img/thirsty_pitcher_revealed.png) | V2 |
| 進入 CTY40、最終鑰匙事件尚未完成 | [`final_key_shrine_before.png`](img/final_key_shrine_before.png) | V2 |
| CTY40 正式調查取得最終鑰匙 | [`final_key_obtained.png`](img/final_key_obtained.png) | V2 |

本機完整原版影片 `f000848.jpg`／`f000849.jpg` 可核對使用格、礁石構圖與正式道具窗，
`f000850.jpg` 已進祠堂；原片不加入 Git。取得鑰匙後中心 tile 變成紅色十字也是原版可見
結果，不是 remake 的破圖。原版海水是藍色且事件有快速 palette transition，
目前 remake 地表水面仍呈棕色、未重播四輪動畫，因此不得標 V3。這是玩家可見 parity GAP，
不能用正確 transaction 掩蓋。

## 7. 下一個 blocker

- 從 CTY40 checkpoint 正常返回提頓，夜間以最終鑰匙開牢門並取得綠寶珠。
- 另追世界水面 tile／palette animation，使本事件由 V2 升至 V3。
- 本切片不代表全新遊戲至 THE END 已完成。
