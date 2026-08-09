# 開場能力確認框線：IDA writer／RGB sidecar

本文件只封存能力確認與姓名／性別 modal 共用的**外框色彩**證據。它不把
`confirm_choice` 的藍色交錯圖樣誤稱為已完成，也不把 raw EGA plane mask 當作可跨版本
的 style ID。

## 輸入與位址基準

| 項目 | 值 |
|---|---|
| 輸入 | `DQ3.EXE`，115282 bytes |
| SHA-256 | `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| 工具 | IDA Pro `9.4.0.260610`，`ida-pro-9.4-ver3:latest`，一次性 Docker |
| IDA 位址空間 | DOS loaded linear `0x10000..0x2aee0`；DGROUP base `0x24dd0` |
| DGROUP 對照 | `file = 0x16140 + DS_off`；`DS:0x727` 為 file `0x16867`、linear `0x254f7` |
| 輸出 | IDA auto-analysis／IDC function dump／`.asm`；原始 EXE 保持唯讀 |

本批 IDA image 內沒有可用的 host Python 3.14 shared library，因此沒有把失敗的
IDAPython stdout 當作證據；結論來自 IDA database 的函式邊界、xref、IDC disassembly 與
可回查的 raw bytes。這項工具限制本身保留在交接紀錄中。

## caller → writer → consumer

### 1. 開場 raw window 確實走共用 writer

`sub_10854` 在姓名完成後依序：

1. 以 `lea si, byte_28BC6`／`sub_1F4E3` 顯示性別選擇；
2. 以 `lea si, byte_28B78`／`sub_1F590` 顯示能力主面板；
3. 以 `lea si, byte_28B92`／`sub_1F590` 顯示「這個人可以嗎？」提示／選項。

三個結構的 raw 欄位與 flags 已由 [`docs/113`](113-newgame-geometry-re.md) 保存：
`(flags=3, x, y, width, height)`。`sub_1F590` 先呼叫 `sub_1FCC6`，再呼叫
`sub_1FB36`、`sub_1FC57`，最後以 `sub_1FCE1` 將目前 window 推入背景／框線堆疊；
因此 `flags=3` 不是三種可任意替換的 style 名稱。

### 2. 四邊 frame writer

IDA linear `sub_1FD30`（`0x1fd30..0x1fdb0`）依序呼叫 `sub_1FDB1` 四次：

```text
top:    di=[si+2], ax=[si+4],     bp=[si+6], cx=0x10
bottom: di=[si+2], ax=[si+4]+[si+8]-0x10, bp=[si+6], cx=0x10
left:   di=[si+2], ax=[si+4]+0x10, bp=2, cx=[si+8]-0x20
right:  di=[si+2]+[si+6]-2, ax=[si+4]+0x10, bp=2, cx=[si+8]-0x20
```

`sub_1FDB1`（`0x1fdb1..0x1fe11`）設定 EGA graphics/sequencer registers，使用
`bh=0xaa` 的旋轉 bit pattern 寫入四邊；每次由 `mov bl, ds:727h` 取得當前 plane mask。
原始 EXE 的 `DS:0x727` byte 為 `0x05`（file `0x16867`），但該 byte 是原版 EGA 狀態
資料，不能直接當作另一代引擎的 RGB 或 style enumeration。

### 3. 玩家可見色彩

對 DOSBox 原版 [`v3_01_afterstats.png`](../dosbox/v3_01_afterstats.png)（原始 SHA-256
`0b55c114a44a9ff5b83f1601425b82cc607ffb1ef3d92c5f123dd047e73b8695`）在 640×350 logical
viewport 取樣：

- 左／右／上／下能力 panel 的外框可見 RGB 為 `(255,223,255)`；
- frame 內部為 `(0,0,0)`；
- `confirm_choice` 另有 `(0,85,223)` 與黑色交錯 pattern，並與文字／角色 sprite 重疊。

因此本輪只把可證實的外框 RGB 以 `interface.json.new_game_geometry.frame` 資料化；
共用 renderer 仍只執行一像素矩形 primitive，不知道 DQ3 座標或 raw mask。

## 推論等級與限制

| 結論 | 等級 | 限制 |
|---|---|---|
| `sub_1F590` → `sub_1FD30` → `sub_1FDB1` 是開場 window 的四邊 writer | `strong` | 有 IDA caller／callee／raw window 閉合；尚未以同一 InputState 做逐幀 DOSBox capture |
| 外框可見 RGB `(255,223,255)`、內部黑 | `D2`／`strong` | 原版 screenshot 與 static writer 交叉吻合；palette register 的完整時序未在本 sidecar 重建 |
| `confirm_choice` 的藍色交錯 pattern | `unknown`／V3 gap | 目前只有玩家可見取樣，尚缺完整 plane/latch writer→動畫 consumer 閉合 |

`NewGameWithPack` 現在要求 `new_game_geometry.frame`；直接 unit fixture 若未安裝 frame
仍可使用導覽 fallback。這個切片提升的是外框色彩資料化與 runtime V2，不能宣稱整個
開場畫面已達 V3。
