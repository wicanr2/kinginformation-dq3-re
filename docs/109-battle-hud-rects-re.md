# DQ3 戰鬥 HUD 框線、原始旗標與列高：IDA 9.4 非破壞性證據

> 2026-08-09；本文件是 `DQ3.EXE` 的附加證據 ledger，不改寫原始 binary 或 IDA
> database。語意只附加在原始 linear address／DGROUP／file offset 旁，舊結論若被
> 推翻仍保留在引用文件中。

## 輸入與工具

| 欄位 | 值 |
|---|---|
| 輸入 | `assets_raw/DQ3.EXE` |
| 大小 | `115282` bytes |
| SHA-256 | `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| 主要工具 | IDA Pro `9.4.0.260610` (`/home/anr2/ida_94_official/dist` 的 Docker image) |
| IDA 位址基準 | DOS EXE loaded linear `0x10000..0x2aee0`；`seg016` linear base `0x24dd0` |
| file 對照 | 本文明列的 raw rect 以 `DQ3.EXE` file offset 表示；不可與 logical／linear 數值混用 |
| 分析限制 | IDAPython 在該授權 image 未載入，故本輪使用 IDA 產生的 `.asm`、xref／函式邊界與 raw-byte parity；`.i64`／`.asm` 留在 `/tmp`，不進 Git |

## 共用 `win_rect` 的可見尺寸勘誤（confirmed）

`sub_1f590` 先進入 `sub_1fb36`，其 `[si+6]+1`、`[si+8]+0x10` 是關閉時
備份背景的 buffer 範圍。真正畫可見邊框的是 `sub_1fd30`：

```text
top/bottom:  x=[si+2], y=[si+4] / [si+4]+[si+8]-0x10, width=[si+6]
left/right:  x=[si+2] / [si+2]+[si+6]-2, y=[si+4]+0x10, height=[si+8]-0x20
```

因此 production JSON 的可見 `width`／`height` 取 raw `+6`／`+8`，而不是把
`sub_1fb36` 的備份額外範圍算進去。共用 rect `DGROUP 0x3e6e`（file `0x19fae`）的
raw bytes `0b 01 13 00 ee 00 2c 00 60 00` 對應：

```text
x=0x13*8=152, y=0xee=238, width=0x2c*8=352, height=0x60=96
```

這推翻了 `docs/94` 舊版 `(360,112)` 的備份區解讀；舊 bytes、舊文件與反證鏈仍可回查，
新的 JSON／parity test 改用 `(352,96)`。

## 原始第一 word 的正確語意：旗標／備份槽，不是 frame style ID（confirmed）

`win_rect` 的第一個 word 會被 `sub_1f590`／`sub_1f4e3` 分成兩個 byte 使用：
低位元組傳給 `sub_1fb36`／`sub_1fbda` 作背景備份槽索引，高位元組的 bit 1 控制
`sub_1fcc6`／`sub_1fce1` 的視窗堆疊與 `sub_1fd30` 邊線清除。它不是三種可替換
的框線樣式；目前已回查的原始值 `0x031a`、`0x0b11`、`0x0912` 應保留為 raw
flags／backup slot，不能在 pack 或 Go 中命名成 style。

## 共用可見框線 frame（strong）

`sub_1f590 → sub_1fd30 → sub_1fdb1` 是共用的可見框線 writer。`sub_1fd30` 只
依 `[si+2]`／`[si+4]`／`[si+6]`／`[si+8]` 及 `DS:727` 的 EGA plane mask
畫四條邊；沒有依第一 word 選擇另一套 glyph pattern。與原版 DOSBox 戰鬥畫面
[`dosbox/orochi_boss.png`](../dosbox/orochi_boss.png) 的像素核對顯示，640×350
邏輯畫布上的玩家可見結果是連續 1px 白框、黑色內部。這項結果以 `strong` 保存：
原始 writer 與實機畫面一致，但尚未以同一個完整 input trace 對每一個 modal
逐畫面閉合。

`interface.json` 的 `frame` 物件只保存這個可見 RGB 結果（`border_rgb`、
`interior_rgb`）與上述證據；框線演算法仍是共用 engine primitive。production
`NewGameWithPack` 缺少戰鬥三個 frame 物件時 fail closed；直接 unit fixture 才
允許暫時使用導覽用的黑底白框。

本輪另外把數量字的可見列基線與原版畫面閉合：輸入
`dosbox/orochi_boss.png`（640×350，SHA-256
`746d7a95c2456fa6338438e6b61ff4fe120670304ebb0f98fdb98c36ef0cad55`）在敵名與數量
兩個區域都顯示相同的 16px 字模列（數字 `1` 的可見像素列同高）。這是畫面 D2
證據；它只證明玩家可見的 row relation，不把尚未完全解開的 EGA writer 內部座標轉換
誤升格為 D3。

本輪 Docker＋Xvfb 的 renderer dump [`battle_count_row_runtime.png`](../dq3_remake_ebitan/docs/battle_count_row_runtime.png)
（`TestDumpMultiEnemyBattle`，SHA-256
`a9c49e091d25bc28d018cf5363c07f75dac4a8029a91cf78f9da60b179518eb8`）也量得名稱／數量
可見列同為 `y=264..277`。這是 debug 視覺證據，不取代正式玩家輸入 trace，也不把
debug hook 當成 production 入口。

## 指令／選單框（confirmed + strong）

### 入口到 consumer

```text
sub_1c169 (逐隊員下令)
  → sub_1c1d8 (writer：填 DS:0x411a / 0x4126，使用 DS:0x4112)
  → sub_1f590 (畫 frame；sub_1fd30 畫可見邊框)
  → sub_1f908 (cursor；sub_215ee 畫 actor label)
  → sub_1f779 (正式鍵盤／游標輸入)
```

`DS:0x4112` 在 IDA `seg016` 的 linear address 是 `0x28ee2`，raw file offset 是
`0x1a252`。原始 bytes（保留 raw，不以自訂名稱取代）為：

```text
1a 03 | 12 00 | f8 00 | 10 00 | 60 00 | b9 01 | 00 00 | 00 00
00 00 | 00 00 | 00 00 | 03 00 | 14 00 | 08 01 | 00 00
```

可證實的幾何與列規則：

| 原始欄位／writer | JSON／畫面語意 | 等級 |
|---|---|---|
| `+2=0x12`, `+4=0xf8` | 左上角 `(144,248)` | `confirmed` |
| `+6=0x10` | 可見寬 `128px` | `confirmed` |
| `sub_1c1d8`：`+8=0x60`、`+0x14=4` | 四列時高 `96px` | `confirmed` |
| `sub_1c1d8`：`+8=0x50`、`+0x14=3` | 三列時高 `80px` | `confirmed` |
| `sub_1f908`：`+0x18=0x14`, `+0x1a=0x108` | 游標原點 `(160,264)`（相對框 `(16,16)`） | `confirmed` |
| `sub_1c1d8`：`add bp,4` 後呼叫 `sub_215ee` | actor label X 相對框 `+32px` | `confirmed` |
| `battleCmdLabels` 兩個 glyph 的欄距 | JSON `label_inset_x=32`、`secondary_label_inset_x=80`；由 raw menu record 與實機畫面共同校核 | `strong` |

### 指令標籤的資料化

原版 `sub_1c169` 的五種穩定指令角色使用下列 D3TXT00.FON glyph pair；這批先保留
`D2`，因為字模索引與玩家可見畫面已對上，但逐指令的 IDA writer 位址仍未獨立匯出
成 sidecar，不能把 `strong` 直接升成 D3。`interface.json.battle_command_labels`
保存同一對映，renderer 不再從 Go table 猜字：

| role | primary／secondary glyph | 原版字形 | inference |
|---|---:|---|---|
| `war` | `107／207` | 戰／鬥 | `D2` |
| `flee` | `629／630` | 逃／跑 | `D2` |
| `defend` | `203／204` | 防／禦 | `D2` |
| `item` | `402／1354` | 道／具 | `D2` |
| `spell` | `429／430` | 咒／文 | `D2` |

`NewGameWithPack` 缺少這組契約會 fail closed；直接 Battle fixture 若不安裝 pack labels
只是不繪製版本專屬文字，不能視為可發佈行為。後續若要升級至 D3，需從 IDA
保留原始函式／位址／bytes 的 sidecar 補齊每個 label writer，並與同狀態 DOSBox 畫面
核對，不得以現有 Go 常數回填證據。

`height_mode="rows_plus_base"`、`base_rows=2`、`row_height=16` 是已閉合的具名
engine primitive；Go 不保存 DQ3 的 3／4 或座標 fallback，pack 缺少此 panel 時
`NewGameWithPack` 失敗即關閉。

## 敵名／數量框（confirmed + strong）

### 入口到 consumer

```text
sub_1b053 / sub_1b0cb
  → writer: byte_28f00 / byte_28f1e 的 active-group height
  → sub_1f590 (畫 frame)
  → sub_1b101 (sub_213c4 畫 monster name；sub_21929 畫 count)
```

`byte_28f00` 的 IDA linear address 是 `0x28f00`，raw file offset `0x1a270`；
`byte_28f1e` 是同一結構的第二份 variant，raw file offset `0x1a28e`。第一份
raw bytes 為：

```text
11 0b | 24 00 | f8 00 | 20 00 | 60 00 | b4 01 | 02 00 | b5 01
b6 01 | 01 01 | 00 00 | 03 00 | 26 00 | 08 01 | 00 00
```

可證實的幾何與列規則：

| 原始欄位／writer | JSON／畫面語意 | 等級 |
|---|---|---|
| `+2=0x24`, `+4=0xf8` | 左上角 `(288,248)` | `confirmed` |
| `+6=0x20` | 可見寬 `256px` | `confirmed` |
| `sub_1b053/sub_1b0cb`：`word_28f08=(active_groups+2)<<4` | 高 `16*(active_groups+2)` | `confirmed` |
| `sub_1b101`：`word_28f02+4`、`word_28f04+0x10` | 名稱文字相對框 `(32,16)` | `confirmed` |
| `sub_1b101`：`bp+=0x0e` 後 `sub_21929` | 數量 X 相對框 `+144px` | `strong`（`sub_21929` 的 glyph baseline 仍需逐像素對拍） |
| `sub_1b101`：名稱與數量同一可見 16px row | JSON `count_inset_y=0`（相對 `text_inset_y`） | `strong`（IDA caller + `orochi_boss.png` D2 畫面） |

JSON 的 `height` 以一列 canonical fixture 為 `48px`，實際 production 依
`rows_plus_base` 動態套用 active group rows；未知列數不填 0，validator 會拒絕缺資料。

## 未閉合項目

- `0x031a`、`0x0b11`、`0x0912` 是 raw flags／備份槽，並非三種 frame style；其
  可見框線已由共用 `sub_1fd30/sub_1fdb1` 與 DOSBox 像素證據閉合，renderer 不應
  再把它們當成可替換樣式。
- `sub_21929` 的可見數量 row relation 已由 `count_inset_y=0` 資料欄位與
  `orochi_boss.png` 同列畫面閉合；數字 writer 內部 raw DX 到顯示像素的轉換仍只保留
  `strong`，不可把這項局部證據宣稱為整個戰鬥 V3。
- spell／target 子選單的其他原始 rect 尚未逐一建立 sidecar；本輪只把共用 command panel
  primitive 接到既有 renderer，未宣稱所有 battle phase 的 V3 已完成。
