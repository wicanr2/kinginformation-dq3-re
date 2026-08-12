# 126 — 創角能力確認：同狀態 V3 靜態對拍

> 2026-08-12。本文件只宣告「創角能力確認的固定畫面」達成 V3 靜態對拍；不把單張
> 畫面擴大宣稱為整段創角、游標閃爍、淡入、音效或全遊戲的 V3。

## 範圍與可重播輸入

原版由正式操作進入：新遊戲 → 名稱 `0` → 男性 → 確認。沒有使用記憶體注入、debug
入口或改寫原始檔。原版邏輯畫布為 `640×350`，數值為：力量 8、耐力 4、速度 4、
HP 12、MP 9、聰明度 7、運氣點數 8、守備力 6、經驗 0。

| 項目 | 值 |
|---|---|
| 原始輸入 | `assets_raw/DQ3.EXE`，115,282 bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| 原版畫面 | Docker DOSBox 正式輸入產生的 `verify_open_10_story_01.png`；先明確裁左上 `640×350` logical viewport，再比對；僅在 `/tmp`，未加入 Git |
| remake 畫面 | `TestDumpNewGameConfirmV3StaticFixture` 以同一名稱、性別與能力值輸出 |
| 執行環境 | 一次性 Docker、Xvfb、`dq3-ebitan-test:20260811`；測試使用 `DQ3_ASSETS=/workspace/assets_raw` |
| DOSBox 外殼 PNG SHA-256（1024×768） | `40d6dbaaebdc83fc0d9e13246d84ab6d1f1fefcd2d60c95f223e3ea07db4bd2c` |
| 最終原版 logical PNG SHA-256（640×350） | `f5d55df53067aeb5dd4374a7ae0d4a63a7f9616b6b5c96713461637e72ad7e2a` |
| 最終 remake PNG SHA-256 | `1f75efc5e6e5c19f5038b1583a71bbb03a5132122b8f4a88401f3bc3193dbf02` |
| 差分 PNG SHA-256 | `4d8a58b8f785c85aca88bfc9e17df4fe7d1e787ec33c8370f89d5dc12269d401` |

原版素材、原版截圖、IDA database 與差分圖都只在工作目錄保存，沒有進入版本庫。

## 最終量測

以 ImageMagick `compare -metric AE` 對兩張 `640×350` PNG 逐像素比較，最終
絕對誤差（AE）為 **1,474**。以 224,000 個畫布像素換算為 **0.658%**；這是目前受控
變體中最低值。差分仍存在，故不使用「完全逐像素一致」的說法。

此 checkpoint 的 V3 含義是：原版與 remake 有相同正式前置流程、相同邏輯畫布、相同
玩家可見資料，以及可重跑的同狀態畫面比較。它**不**涵蓋下列時間性事實：

- 名稱游標的閃爍 phase、按鍵 repeat、人物立繪或背景切換的 frame timing；
- EGA latch／palette register 在每一個中間 frame 的狀態；
- 對話、選項與音效的 wall-clock 停頓。

剩餘 1,474 個不同像素不以任意 DQ3 常數或肉眼位移掩蓋。它們包含原版 EGA 寫入／遮罩
的細部殘差與單張捕捉無法區分的即時狀態；在另有同 phase frame trace 前，維持為
`strong`／V3 殘差，而非猜測性 JSON 值。

## 原始證據與資料化實作

| 原始定位 | 原始行為 | 結論／等級 |
|---|---|---|
| `sub_10854`，IDA linear `0x10854` | 依序以 `0x28b78` 顯示能力主面板、呼叫 `sub_1834E` 寫動態資料、再顯示 `0x28b92` 提示與選項 | `confirmed` caller→writer→玩家畫面鏈 |
| `sub_1834E`，`seg000:834E` | 寫姓名、職業、性別及十列能力；數字以固定寬度 decimal field 落到欄位右端 | `confirmed`；欄位 anchor、digits 與 label glyph 全存於 pack |
| raw `0x28b78`／`0x28b92`／`0x28bc6` | 原始 window `(flags,x,y,width,height)` 分別為 `(3,19,46,44,192)`、`(3,45,14,22,48)`、`(3,43,46,12,64)` | `confirmed` raw 結構；保留 byte-addressed 值，不以像素座標覆蓋 |
| `sub_1F590 → sub_1FC57` | 共用 EGA window 背景 writer；同狀態畫面可見底圖為黑色 | `confirmed` writer、D2 可見結果；terminal cell projection 是量測資料，不外推成通用規則 |
| `sub_1FD30 → sub_1FDB1` | 四邊 EGA frame writer | `confirmed` writer；`beveled_2px` 可見投影由本 checkpoint V3 靜態核對 |
| `sub_1F63C` | choice redraw 走獨立路徑，沒有直接呼叫 `sub_1F590` | `strong`；choice raw backdrop 維持最小量測 projection，不誤稱同一 writer 完全閉合 |

實作只增加跨版本的 engine primitive；所有 DQ3 資料都在
`internal/gamepack/packs/dq3_cht/data/interface.json`：

- `new_game_labels` 的 record 407 glyph stream、`?`、性別與 choice cursor；
- `new_game_geometry.stats` 的十三個具名 `label`／`value` 欄位，包含固定數字寬度；
- `GeometryRect.frame_edge_widths` 的共用 seam 資料；
- 三筆 `window_backdrops` 的 raw window 參照、draw order 與 terminal EGA cell projection；
- `FIRST.SCR` 的原始 palette RGB 值。

因此 Go renderer 不含 DQ3 專屬 record、座標、flag 或玩家文字。此固定 checkpoint 當時的
資料契約是 `schema_version: "0.1.30"`、DQ3 pack `content_version: "0.1.35"`；現行契約已因
固定編隊背景 selector 升為 `0.1.31`／`0.1.36`，不改變本頁能力確認的證據結論。欄位說明見
[`docs/84-game-pack-json-contract.md`](84-game-pack-json-contract.md)。

## 反證：record 407 不是可直接貼上的 bitmap template

`sub_213C4`（IDA linear `0x213c4`）會解讀 record word、`0xffff`／`0xfffe` 等控制值，
再經字模 writer 畫出模板。曾以 record 407 直接重播為 renderer template；同一輸入的
AE 反而為 6,105（2.725%），明顯比最終 1,474 差。這個實驗已移除，不留在 production
code 或測試中。

因此本頁採用的是「已證實的原始資料流 + 可驗證的具名 pack 幾何」，不是把控制 record
誤當成跨版本可直接繪製的畫面模板。

## 驗收與後續邊界

已通過：

```text
go test ./internal/gamepack
DISPLAY=:93 DQ3_ASSETS=/workspace/assets_raw \
  DQ3_DUMP_NG_V3=1 NG_V3_OUT=<temporary-output> \
  go test ./game -run '^TestDumpNewGameConfirmV3StaticFixture$' -count=1
```

此頁更新後，`docs/112`、`docs/113`、`docs/117` 的舊 V2 描述只保留為時間序列；現行
規格以本文件及 `interface.json` 為準。若要繼續開場的動態 V3，下一個必要輸入是可控
tick 的原版 frame trace，而不是繼續以靜態偏移猜測。
