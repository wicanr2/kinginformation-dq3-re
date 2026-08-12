# 開場能力確認框線：IDA writer／RGB sidecar

本文件封存能力確認與姓名／性別 modal 共用的**外框色彩與邊線遮罩**證據。它不把
`confirm_choice` 內部的藍色交錯選擇圖樣誤稱為已完成，也不把 raw EGA plane mask 當作
可任意擴充的 style ID；pack 只引用已註冊的 engine primitive。

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
資料，不能直接當作另一代引擎的 RGB 或 style enumeration。以 logical frame 的每一條邊
重播 `bh=0xaa` 後，像素層可表達為局部座標 `(x+y)%2==1` 的
`checkerboard_1px` mask；未命中的像素保留底圖，這正是新遊戲 pack 目前引用的已註冊
engine primitive。

### 3. 玩家可見色彩

對 DOSBox 原版 [`v3_01_afterstats.png`](../dosbox/v3_01_afterstats.png)（原始 SHA-256
`0b55c114a44a9ff5b83f1601425b82cc607ffb1ef3d92c5f123dd047e73b8695`）在 640×350 logical
viewport 取樣：

- 左／右／上／下能力 panel 的外框可見 RGB 為 `(255,223,255)`；
- frame 內部為 `(0,0,0)`；
- `confirm_choice` 另有 `(0,85,223)` 與黑色交錯 pattern，並與文字／角色 sprite 重疊。

因此本輪把可證實的外框 RGB 與 `border_pattern=checkerboard_1px` 以
`interface.json.new_game_geometry.frame` 資料化；共用 renderer 只執行名稱化的
一像素邊線 primitive，不知道 DQ3 座標或 raw mask。`confirm_choice` 內部的藍黑選擇
圖樣仍需另外追 selection writer／palette 時序。

## 推論等級與限制

| 結論 | 等級 | 限制 |
|---|---|---|
| `sub_1F590` → `sub_1FD30` → `sub_1FDB1` 是開場 window 的四邊 writer，`bh=0xaa` 對應局部 checker mask | `strong` | 有 IDA caller／callee／raw window 與 bit pattern 閉合；尚未以同一 InputState 做逐幀 DOSBox capture |
| 外框可見 RGB `(255,223,255)`、內部黑 | `D2`／`strong` | 原版 screenshot 與 static writer 交叉吻合；palette register 的完整時序未在本 sidecar 重建 |
| `confirm_choice` 的藍色交錯 pattern | `unknown`／V3 gap | 目前只有玩家可見取樣，尚缺完整 plane/latch writer→動畫 consumer 閉合 |

`NewGameWithPack` 現在要求 `new_game_geometry.frame` 及已註冊的 pattern；直接 unit
fixture 若未安裝 frame 仍可使用導覽 fallback。這個切片提升的是外框遮罩資料化與
runtime V2，不能宣稱 palette、`confirm_choice` 藍色選擇圖樣或整個開場畫面已達 V3。

## 勘誤（2026-08-11）：能力大面板改以原圖像素閉合

前述把四邊 writer 的 `bh=0xaa` 直接投影成 `checkerboard_1px`，是尚未把玩家可見
logical pixels 與 EGA writer 的兩列／兩欄邊界分開的強推論，已被同一張
`dosbox/v3_01_afterstats.png` 像素取樣推翻。原圖在 `StatsLeft`、`StatsEquipment`、
`StatsRight` 三個 rect 的外框呈現連續的 `(255,223,255)`，上／下各兩列、左／右各兩欄；
內部保持黑色。當時 pack 先改用 `solid_2px`，並由 engine 的具名 primitive 實作，
不改變 `confirm_choice` 的藍黑棋盤與 `checkerboard_frame_2px`。原有 `checkerboard_1px`
判讀保留為已推翻歷史證據；`confirm_choice` 的逐幀 palette／動畫仍是 V3 gap。

## 勘誤（2026-08-12）：`beveled_2px` 與固定畫面 V3

`solid_2px` 是 `checkerboard_1px` 被推翻後的中間近似，不是現行規格。以正式輸入的
`verify_open_10_story_01.png` 再與同狀態 remake 對拍後，三個大面板的四個**最外角**應
保留底圖，其餘兩像素寬邊仍為連續 lavender。因此 `interface.json` 現採註冊的
`beveled_2px` primitive；它不是 DQ3 分支，而是由 `GeometryRect.frame_edge_widths` 可調的
共用 frame primitive。

原始 `sub_1FD30 → sub_1FDB1` writer、visible RGB 與 raw window 仍保持原定位；只有像素
投影的舊強推論被更新。此變更使能力確認固定 checkpoint 達 V3 靜態對拍（AE=1,474），
但不推出每一 frame 的 EGA latch、palette 或游標 timing。完整輸入 hash、量測與殘差見
[`docs/126`](126-newgame-confirmation-v3-static-comparison.md)。
