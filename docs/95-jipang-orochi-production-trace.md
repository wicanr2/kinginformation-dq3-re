# 95 — 日邦格八頭大蛇兩階段事件

> 狀態：原版設定與控制流 D3；從新遊戲起的正式玩家輸入 trace 與 save/load 已閉合至 E3；
> 四張 Ebitengine runtime 圖為 V2，尚未做逐像素同狀態 DOSBox V3。

2026-08-12 的四人隊正式 trace 已取代舊單勇者 fixture 作為戰鬥畫面證據；比對發現兩戰
背景尚未 match 原版影片，故仍只是 V2／near-state，詳見 [`docs/127`](127-orochi-v2-production-compare.md)。

## 先前錯誤

舊 `bossTrigger` 把 CTY19 第一戰直接視為事件完成，同時給草薙大劍 `0x14`、紫寶珠
`0x69` 並寫一個近似旗標。這與原版兩次戰鬥、戰間自動轉場、可重複的 Yes 分支及戰後寶箱
都不一致。機制可以戰鬥不代表設定資料、入口與副作用已 match。

## 原版閉合鏈

IDA Pro 9.4 在一次性 Docker 容器分析原始 `DQ3.EXE`；logical 與 file offset 的換算仍採
`file = logical + 0x1370`。

| 階段 | 原版入口／consumer | 玩家可見與狀態副作用 |
|---|---|---|
| 第一戰 | handler45 logical `0x5cb6`（file `0x7026`）；CTY19 sec1 固定 NPC `(35,12)` | 編隊 DGROUP `0x4ed0`：`01 23 05 4b 01`，即背景35、page5、怪75×1 |
| 第一戰後 | handler45 先以 `sub_167c5` 消費 DGROUP `0x3cee`，再呼叫旗標 helper 與兩次玩家移動 consumer | movement table `01 03 00 01 03 01 ff` 令 NPC slot0 向右兩步；其後 clear `0x44`、set `0x20`，玩家方向3向右兩步，轉至 CTY21，顯示 D3TXT04 rec70 |
| 第二次交談 | handler36 logical `0x5979`（file `0x6ce9`）；CTY21 sec0 NPC `(18,7)` | rec66 提問；Yes 顯示 rec67 後返回且可重談；No 顯示 rec68 並開戰 |
| 第二戰 | handler36；編隊 DGROUP `0x4ed5` | `01 1a 05 4b 01`，即背景26、page5、怪75×1；`DGROUP 0x2518` OR `6` 抑制掉落 |
| 第二戰後 | handler36 | rec69；clear `0x20`、`0x45`，set `0x1f`；同格寶箱才可取得紫寶珠 `0x69` |

原版完整實況亦呈現相同順序：`f000745` 第一戰、`f000746` 掉落通知、`f000748` 自動到宮殿
並顯示 rec70、`f000765` 選項、`f000768` 第二戰、`f000774` 戰後寶箱。影片只作流程與畫面
oracle，數值仍由 EXE／DAT 仲裁。

## 怪物掉落欄位校正

勝利結算 logical `0xc425`（file `0xd795`）以 `DGROUP 0x2321` 的 formation 第一組怪物
raw ID 乘記錄長度 `0x29`，從 D3MNS runtime record 的 `+0x25` 讀判定閾值，以
`roll(256) <= threshold` 判定，再從 `+0x26` 讀道具；`DGROUP 0x2518 bit1` 會略過
掉落。怪75為 `+0x25=0xff`、`+0x26=0x14`，所以第一戰由正常掉落取得草薙大劍；第二戰
由 handler36 抑制，不會再掉一把。`+0x27` consumer 尚未閉合，程式與文件保持
`unknown_27`，不再沿用舊的 `DropRate` 誤名。

## Remake 對應與驗收

- schema `0.1.10` 的 `staged_boss_events` 保存兩個 NPC selector、兩組編隊、旗標交易、
  兩步路徑、掉落政策、寶箱 gate 與 rec66–70 text ID。
- Go 僅提供有限兩階段 boss 狀態機與 `original`／`suppress` 掉落 primitive。
- `TestOpeningProductionInputTrace` 已從標題新遊戲一路正常抵達 CTY19，經正式「對話」、
  戰鬥、移動、選項與「調查」取得兩件道具，再於 CTY21 save/load 驗證旗標與 inventory。
- component tests 另鎖定 Yes 分支不開戰且可重談、戰敗不交易旗標，以及第二戰不重複掉落。
- runtime 圖：`docs/img/jipang_orochi_first_battle.png`、`jipang_orochi_first_post.png`、
  `jipang_orochi_second_battle.png`、`jipang_purple_orb_obtained.png`。尚待 DOSBox 同狀態 V3。

## 2026-08-12 四人隊正式畫面勘驗

舊 PNG 是 `NewGame` 加載 fixture，只有主角，不能用來判斷原版四人 HUD 或正式遭遇狀態。
本輪在 Docker＋Xvfb 以 `TestOpeningProductionInputTrace` 從標題正常重播至 `THE END`，
僅在既有的第一、第二八頭大蛇戰的 `command`／`message`／`end` phase
輸出 640×350 PNG。取樣先讓 renderer 的短暫受擊 flash 消退，避免 high-speed trace 因未 Draw
而把純紅受擊覆蓋誤當作怪物色盤。

結果確認四人狀態列、命令框與敵名框均由正式流程渲染；怪物也不是缺色或透明。然而同源
原版影片第一戰結束畫面是洞窟、第二戰遭遇畫面是沙漠，remake 分別顯示天空草地與綠色背景。
上述文字是 2026-08-12 早期 trace 的歷史缺口，已由後續 IDA 9.4 archive loader 閉合。現行
staged-boss path 將 raw byte1／byte2 作為 `page_raw`／`palette_bank_raw`，由 pack 提供完整
`PACKBG.SCR` stride 與 `MNSBK.PAL` bank；第一、第二戰正式 trace 分別顯示洞窟、沙漠，達
near-state V2。這不是 generic terrain parity，也不是 V3。原始 raw、勘誤原因與最新 hashes
見 [`docs/128`](128-battle-background-selector-re.md)。
