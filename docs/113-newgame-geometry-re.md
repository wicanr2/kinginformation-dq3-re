# 開場／創角幾何反組譯與對拍

本文件封存 `interface.json.new_game_geometry` 的來源與限制。初版的
`checkerboard_1px`、`solid_2px` 與「尚未 V3」段落都保留為時間序列；2026-08-12 的
現行 canonical 幾何是 `beveled_2px`、record 407 的十三個具名能力欄位與三層 raw EGA
backdrop，固定能力確認畫面已達 V3 靜態對拍。完整現況與殘差見
[`docs/126`](126-newgame-confirmation-v3-static-comparison.md)，不要以本文件早期段落覆蓋。

## 輸入與工具

- 輸入：`assets_raw/DQ3.EXE`，115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：IDA Pro 9.4（`/home/anr2/ida_94_official/dist`，在一次性 Docker image
  `ida-pro-9.4-ver3:latest` 執行）；database 與批次 listing 留在 `/tmp`，未加入 Git。
- IDA listing：`/tmp/dq3-ida-batch.asm`。IDA linear range `0x10000..0x2aee0`，
  `seg0`／logical offset 與 file offset 不可混用；本文件的 `linear:` 標籤就是 IDA
  linear address。
- 可重生輸出：listing header 同時記錄輸入 SHA-256、format、base address 與 entry
  point；IDAPython 在此 image 缺少 host Python 3.14 library，故本批只採用 IDA
  auto-analysis／IDC 匯出，不宣稱有 IDAPython sidecar。

## caller → writer → consumer

### 姓名／注音輸入

`sub_10854`（logical `0x0854`）呼叫 `sub_10d17`（`0x0d17`）建立姓名 modal；
`sub_10d17` 先 `lea si,byte_29088` 呼叫 `sub_1f590`，再依 `sub_10dc8`／
`sub_10e8e`／`sub_10f5b` 分派功能列、英數與注音盤。

| IDA linear raw 結構 | raw 欄位 `(flags,x,y,width,height)` | consumer／用途 |
|---|---|---|
| `0x29088` | `(1,19,46,32,144)` | 姓名主窗；`sub_1f590` → `sub_1fb36` EGA frame writer |
| `0x290a6` | `(1,41,78,12,112)` | 五列功能列；`sub_10dc8` → `sub_1f779` navigation |
| `0x290c4` | `(1,19,190,12,48)` | 注音盤／切換窗；`sub_10f5b` → `sub_1f590` |
| `0x290e0` | `(1,19,190,34,48)` | 組字候選窗；`sub_112e4` → `sub_1f590` |

`sub_10f5b` 明確寫入 `word_274d2=9`、`word_274d6=0x15`、`word_274d8=0x5e`；
`sub_1123c` 將欄位換成 `x=0x15+2*column`、`y=0x5e+0x10*row`，所以玩家可見
姓名盤是 9 欄、16px 水平／垂直步距。`sub_11172` 同樣把游標位置交給滑鼠／
`sub_1123c` consumer；這不是依英數字串長度猜出的排列。

`sub_1f590` 讀 `byte_29088+0x0a=0x1c3`（D3TXT00 rec451），其 raw 起點
`(0x13,0x2e)` 使標題 anchor 為 `(233,46)`；`sub_10e55` 的 `di=0x1f+2*len`、
`dx=0x3e` 使姓名列文字 anchor 為 `(249,62)`，游標每字 16px。rec452／453 的
逐列 glyph、五列功能文字與 rec456 組字提示已另存 `new_game_labels`；raw35
（欄優先 `cell=col*5+row` 的 cell43）才設定功能列焦點，不能將 45 格壓成舊的 38 格。

### 性別／能力確認

`sub_10854` 在姓名完成後依序呼叫 `sub_1f4e3`（性別選擇）、`sub_1f590` on
`byte_28b78`（能力主面板）、`sub_1f590` on `byte_28b92`（提示／選項），並由
`sub_1f63c` 重繪背景；因此三組 raw window 不能合併成一個全螢幕 modal。

| IDA linear raw 結構 | raw 欄位 `(flags,x,y,width,height)` | 截圖中量到的 640×350 rect |
|---|---|---|
| `0x28b78` | `(3,19,46,44,192)` | 左能力 `x=159,y=52,w=145,h=98`、裝備 `x=159,y=148,w=145,h=82`、右能力 `x=303,y=52,w=194,h=178`（最終 V3 靜態 pack） |
| `0x28b92` | `(3,45,14,22,48)` | `confirm_prompt: x=367,y=20,w=162,h=34` |
| `0x28bc6` | `(3,43,46,12,64)` | `confirm_choice: x=367,y=68,w=98,h=50` |

截圖來源（均為 DOSBox 原版 1024×768 擷取；左上 640×350 是邏輯畫布）：

- `docs/36_shots/04_name_input_zhuyin.png`，SHA-256
  `e8b4e4dd4975d45815b86275585d73ae085754a0b218aa2a6b2e36748cfdff7a`；姓名主窗
  框線量測為 `(159,52)..(400,181)`，功能列分隔在 `x=319/320`，注音模式窗為
  `(159,196)..(240,229)`。
- `docs/36_shots/07_gender_select.png`，SHA-256
  `abbdfc0b3fba317e9a0823e9fdc4ac7b483faebea6c039597e25a96d7e178288`；性別 raw
  window 的同狀態框線 anchor 取 `(344,46)`，文字起點 `(368,62)`，列距 16px。
- `dosbox/v3_01_afterstats.png`，SHA-256
  `0b55c114a44a9ff5b83f1601425b82cc607ffb1ef3d92c5f123dd047e73b8695`；能力面板
  三分割與提示／選項 rect 由調色盤 index 13 的水平／垂直線段量測，不以目前
  remake screenshot 反推。

## JSON 與推論等級

`interface.json.new_game_geometry` 同時保存像素 rect、文字／格盤 anchor、共用外框
RGB、`border_pattern` 與上述 `raw_windows`。`new_game_labels` 另保存 rec451–456 的逐列 45 格與五列
功能文字；
`NameInput` 以 `raw35 → cell43 → function list`、第五列完成的正式輸入路徑消除舊的
直接完成捷徑。整個物件目前標為 `D2`：raw window writer、格距公式、字模 record、
外框 RGB 與 `bh=0xaa` 邊線遮罩已交叉吻合，但尚未為每一個 panel 建立同一輸入序列的
逐幀 DOSBox capture，也尚未閉合 `confirm_choice` 的藍色選擇圖樣。因此 production 可
使用這些幾何與字模，但不能宣稱創角畫面 V3。

能力確認右欄的欄位語意另經 `dosbox/v3_01_afterstats.png` 與 record 407 逐 glyph
解碼校正：六列是「運氣點數／最大HP／最大MP／攻擊力／守備力／經驗」，不是早期
renderer 使用的「速度／HP／MP」。`agility` 只屬詳細狀況窗 role；欄位與 glyph
對映見 [`docs/112`](112-newgame-labels-re.md)。這項修正仍是 D2／V2，不會把尚未閉合
的框線 pattern 提升為 V3。

若後續證據推翻任一矩形，應保留本文件與 JSON 的舊斷言，追加勘誤及新 sidecar；不
得移動 `raw_windows` 位址或只留下改名後的語意。

## 2026-08-10 勘誤：靜態 confirm_choice 已另立資料契約

上文保留的是本文件原先「藍色選擇圖樣尚未閉合」時的歷史狀態。本輪以同一正式姓名／
性別輸入重跑 DOSBox 穩定畫面後，將可見但不涉及逐幀時序的部分拆成三個 pack 欄位：
`confirm_choice_backdrop`（`x=360,y=62,w=112,h=64`、`RGB(0,85,223)`、奇偶 phase）、
`confirm_choice_content`（`x=376,y=78,w=80,h=32` 的黑色內容區）及
`confirm_choice_frame`（`x=367,y=68,w=98,h=50`、`checkerboard_frame_2px`、
lavender／膚色 accent）。能力確認右欄六列的實際 glyph 起點也校正為
`y=126,142,158,174,190,206`；舊 JSON 的 `y=96` 會把文字畫入選擇框，已由
`interface.json` 與 parity test 修正。完整 raw／像素證據與推論等級見
[`docs/118`](118-newgame-choice-backdrop-re.md)。

這項勘誤只把穩定畫面的幾何、色彩與可重現 frame primitive 接入 runtime（D2／V2）；
palette register 切換、游標閃爍、能力條淡入及逐幀動畫仍保留原文件的 V3 限制，不能把
靜態 PNG 對拍擴大解讀成整段創角演出已完成。

## 2026-08-12 現行勘誤：固定確認頁 V3 靜態完成

前段最後一句的「D2／V2」是當時的暫時狀態，已由較強的同輸入畫面比較取代。現在的
`interface.json` 是唯一 canonical owner：

- `stats_name=(200,46)`、`stats_hero=(216,62)`、`stats_sex=(168,78)`、
  `stats_sex_value=(248,78)`、`stats_cloth=(200,158)`；
- 13 個 `stats.*.{label,value}` 取代一維 row anchor，值欄保存 `digits`；
- 左／裝備面板以 `frame_edge_widths.right=1` 描述 shared seam，右面板寬度為 194；
- creation／prompt／choice raw backdrop 依 `draw_order=0/1/1` 在 foreground 前後合成。

這些值由 `sub_10854 → sub_1834E`、raw windows、EGA writer 與同狀態 PNG 共同約束。
完整 hash、AE=1,474、反證與 V3 範圍在 [`docs/126`](126-newgame-confirmation-v3-static-comparison.md)。
逐幀游標、palette register 和整段創角 timing 仍是獨立的未完成工作。
