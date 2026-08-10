# 開場能力確認 `confirm_choice` 藍黑棋盤證據

> 狀態：`strong`／D2；本文件只閉合玩家可見的背景矩形與寫入 phase，尚未把
> 游標逐幀時序、FIRST.SCR 全 palette 或整個能力頁升為 V3。

## 輸入與工具

- 原始輸入：`assets_raw/DQ3.EXE`，大小 115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 靜態定位：IDA Pro 9.4（Docker 內、linear 位址基準 `0x10000..0x2aee0`）；
  raw window `confirm_choice` 為 linear `0x28bc6`，欄位
  `(flags=3,x=43,y=46,width=12,height=64)`。原始名稱／位址保留，沒有以推測
  語意覆蓋 IDA 名稱。
- 動態 oracle：`dosbox-run:latest`、同一套正式姓名／性別輸入與 Xvfb 擷取，
  左上 `640x350` 為邏輯畫布。`confirm_name_start.png` 的 SHA-256
  `e8b4e4dd4975d45815b86275585d73ae085754a0b218aa2a6b2e36748cfdff7a`，與既有
  `docs/36_shots/04_name_input_zhuyin.png` 完全相同，表示輸入起點沒有改用 debug shortcut。

同一序列擷取的穩定能力確認狀態：

| 狀態 | 檔案 | SHA-256 |
|---|---|---|
| 游標在「否」後 | `work/confirm_probe/confirm_no_after_down.png` | `def0314cc53ebbb4b4452832ff8a2298e98da21c37bfdcabc1278ca493a67674` |
| 游標回到「是」後 | `work/confirm_probe/confirm_yes_after_up.png` | `6d28893436131c5ab288bc4140222c68a0a172999e88adddbd45ac1e618cb127` |
| 按「是」後的下一狀態 | `work/confirm_probe/confirm_after_yes.png` | `612e63649d8a246d1c9be894060f5be2d9343eea9d65e094d58d39491e1b4a55` |

`confirm_yes_initial.png`（`111ead8752ed54c936cfcf6279d05f0a0637db1d5b5dc36d27fa8f8d749a6c72`）
仍在能力條／背景切換時序中，不能拿它代替穩定同狀態 oracle。

同一張穩定圖也勘誤了能力右欄的文字 anchor：六列 ink 起點是
`y=126,142,158,174,190,206`，不是先前 JSON 的 `y=96`。這個舊值會把
「運氣／最大HP／最大MP／攻擊力／守備力／經驗」畫進藍框上方；現已將
`interface.json.new_game_geometry.stats_right_rows.y` 改為 `126`。這是由原版
玩家可見 glyph 的連續 16px row 與同一 panel rect 直接量得，不是以 remake 截圖猜測。

## 玩家可見閉合

在「是」與「否」兩張穩定圖中，均可量到同一個藍色 plane 覆蓋矩形：

```text
x = 360, y = 62, width = 112, height = 64
```

其座標範圍為 `x=360..471`、`y=62..125`。命中的 phase 顏色固定為
`RGB(0,85,223)`；未命中的 phase 是原本的黑色底圖。以矩形左上角為 phase 原點時，
`(local_x + local_y) & 1 == 1` 才寫入藍色，例如 y=62 時 x=361、363…為藍色，
y=63 時 x=360、362…為藍色。兩個游標狀態的藍色像素數均為 2016，表示這是
選項窗的靜態背景，不是只在選中項目時才出現的游標 highlight。

可見能力頁的 `confirm_choice` 外框仍是既有 pack rect
`x=367,y=68,width=98,height=50`；它疊在上述較大的背景矩形上。外框、游標與選項
字在此狀態均取 lavender `RGB(255,223,255)`，游標位置只改變約 10×9 glyph 區域：
「否」狀態的箭頭在第一列，「是」狀態的箭頭在第二列。藍色背景本身不隨游標改變。

外框內不是整片黑色：穩定圖在 `x=376..455,y=78..109` 才是中央
`80×32` 黑色內容矩形；其外圍仍保留藍黑 checker phase，再由 frame writer 疊出
lavender／膚色交錯邊線。這個內容矩形已另存為
`interface.json.new_game_geometry.confirm_choice_content`；若把整個
`confirm_choice` interior 填黑，會錯誤抹掉外圍藍色 phase。

因此 engine 新增的 `PatternFill` 只寫入命中的 checkerboard phase，未命中的像素
保留呼叫前的底圖；再用 `confirm_choice_frame` 指定的具名
`checkerboard_frame_2px` 非填充外框 primitive 疊回 `confirm_choice`。這保留了原版
EGA plane/latch 的玩家可見結果，不會用一個黑色 `fillBoxStyle` 把藍色 phase 清掉。

## 推論邊界

- **已證實（D2）**：raw window linear `0x28bc6` 與兩個同輸入 DOSBox 穩定狀態
  都指向同一個 112×64 藍黑 checkerboard；矩形、RGB 與 phase 已寫入
  `interface.json.new_game_geometry.confirm_choice_backdrop`。
- **強推論（D2／strong）**：寫入是 `sub_1f590` 所建立的選項 window 背景在
  `sub_1f63c` 重繪後的可見 consumer；本輪 IDA fresh auto-analysis 未重建舊資料庫
  中所有手工函式邊界，故不把未能直接匯出的 writer 指令標成 confirmed。
- **未知**：游標閃爍／逐幀動畫的完整 period、palette register 切換、能力條淡入及
  角色立繪與背景切換的精確時序。這些仍是開場 V3 gap，不可由兩張穩定截圖推導。

## Runtime 接線

`gamepack.PatternFill` 只接受引擎已註冊的 `solid`／`checkerboard_1px` primitive；
production `NewGameWithPack` 缺 `confirm_choice_backdrop`、`confirm_choice_content`、
`confirm_choice_frame` 或 pattern 不支援時 fail closed。`drawNewGameStats` 在能力文字前
先套用 backdrop，再把 pack 指定的 `confirm_choice_content` 填成黑色，最後以專屬 frame
描外框（中央內容外仍保留棋盤），並以該 frame 的 lavender 色畫能力確認文字／游標。這只改善確認頁的已量測狀態，
不宣稱姓名、性別、戰鬥或場景 palette 已達 V3。
