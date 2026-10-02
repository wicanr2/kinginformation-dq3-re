# 192 — 開場 rec80／rec79 對話序列與逐字時序證據

日期：2026-08-25。本文件勘誤 `docs/188` 把開場城門提示描述成單一 rec80 的舊結論；
原始位址與舊證據仍保留，新增證據只附加語意，不以 rename 取代定位。

2026-10-02補記：生日record82／家中record83的`0xfffc`是同一文字consumer內嵌等待，
不是EOF後再等待一鍵；`0xffff`於生日row3捲動後直接返回，原版第18次Enter已進家中。
冷啟動動態閉合、原始定位與正式紅測試見[docs/188](188-opening-escort-to-castle-spec.md)最新DRAFT。
本文件的舊版面與「逐頁關閉」不能代替該續頁規格，也不外推其他尚未重驗record。

2026-10-02追加時間勘誤：生日／房間自然dosgolem觀測的PIT除數是12428，約96Hz。
先前以BIOS預設18.2Hz換算每字3更新的假設已推翻。此consumer的1tick閾值仍成立，
本次正式修正只套用已觀測生日／房間契約；其他record不外推。
來源、位址、READY審查與回歸入口見[docs/188](188-opening-escort-to-castle-spec.md)。

## 輸入與工具

- 原版：`dq3_remake_ebitan/mobile/assets/DQ3.EXE`，115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 文字：`D3TXT01.TXT` 與 `D3TXT00.FON`；由既有 `dq3data.Text` oracle 解碼。
- IDA Pro 9.4，image `ida-pro-9.4-ver3:py312-x11-v4`；原檔唯讀，暫存 `.i64`／JSON
  位於 `/tmp`，不加入 Git。可重建 exporter：`tools/ida_dump_opening_handler54.py`。
- 玩家可見 oracle：本機原版影片
  `dq3_real_video/YTDown_YouTube_Media_J_fozjiKTB8_001_1080p.mp4` 約 5:17–5:24；
  影片是 1440×1080／60 fps 的縮放轉碼，僅用於事件順序與同狀態畫面，不反推原生座標。

## writer → consumer 閉環

| 原始定位 | 原始行為 | 附加語意 | 等級 |
|---|---|---|---|
| IDA linear `0x101bf..0x101c2`／file `0x152f..0x1532` | `sub_167F6` 後呼叫 `sub_15002` | 建立共用視窗並消費第一段文字流；D3TXT oracle 與影片對應 rec80 | confirmed |
| linear `0x101c5..0x101ce`／file `0x1535..0x153e` | `DS:0x0b34=1`、`DI=0x0c08`、呼叫 `sub_21414` | 不移動角色，緊接著消費第二段文字流；影片與 D3TXT 對應 rec79（含主角姓名插值） | confirmed |
| linear `0x101d3..0x101da`／file `0x1543..0x154a` | `sub_1683A`、`SI=DS:0x3e6e`、`sub_1F604` | 第二段結束後關閉／恢復共用視窗 | strong |
| linear `0x101dd..0x101f5`／file `0x154d..0x1565` | `DS:0x4f1f=1,2,2`，各呼叫 `sub_194C3` | 兩段對話之後才執行剩餘自動移動 | strong |
| linear `0x101f8..0x10204`／file `0x1568..0x1574` | set flag `0x17`、clear flag `0x50`、清 `DS:0x0b34` | opening 完成交易 | confirmed |

`sub_21414`（linear `0x21414..0x215ee`／file `0x12784..`）逐 word 判斷
`0xffff/0xfffe/0xfffd/0xfffc` 與六類插值碼；一般 glyph 經 `sub_211B6`／`sub_21163`
後，在 `DS:0x259e != 1` 時等待 timer `DS:0x0005 >= 1`。這證實原版存在逐 glyph gate，
但 timer 的 PIT／wall-clock 細節依專案停止線不再深挖，只能實作並標示
`hardware-spec approximation`，不可冒稱逐週期 V3。

## Remake 契約與驗收

- `opening_escort.dialogue_records` 是有序且至少一筆的 record 陣列；DQ3 pack 為
  `[80,79]`。引擎在同一 `dialogue_frame_index` 依序顯示，最後一筆關閉前不得交易旗標，
  也不得執行剩餘 arrival frames。
- record 內容、換行、插值與 glyph index 直接取原始 D3TXT／FON，不複製成 Go 字串。
- 20×4、16×16 cell、inset 與視窗外框沿用 `docs/94` 已閉合的 D3 window structure；
  影片縮放不能推翻 EXE 幾何。
- `TestOpeningMotherEscortUsesVisibleFrames` 鎖定 rec80 → rec79 → flags → remaining walk
  的順序與原始 record bytes。

本批閉合的是開場兩段分頁／record owner 與逐 glyph gate 的存在。record 版面與 glyph
位置可達 V3；逐 glyph 的可見速度在採用公開 timer 規格近似並完成同片抽樣前維持 V2，
不以硬體逐週期研究阻塞 remake。
