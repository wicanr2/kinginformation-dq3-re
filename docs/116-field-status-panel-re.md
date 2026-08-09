# 地表狀況詳細窗：record 407 與 `DGROUP 0x3DA8` 非破壞性證據

> 2026-08-10；本文件是附加證據 ledger。原始 `DQ3.EXE`、原始位址與 raw bytes
> 不移動、不覆寫；Go／JSON 只保存可回查的語意與推論等級。

## 輸入、工具與位址口徑

| 欄位 | 值 |
|---|---|
| 輸入 | `assets_raw/DQ3.EXE` |
| 大小 | `115282` bytes |
| SHA-256 | `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| 主要工具 | IDA Pro `9.4.0.260610`，`/home/anr2/ida_94_official/dist`，只讀路徑掛入一次性 Docker |
| 輔助證據 | IDA 匯出的 `/tmp/dq3-ida-style/DQ3.EXE.asm`、`docs/data/d3txt_codes.json`、`scratch_re/status407b.png` |
| 位址基準 | 原始 `DGROUP` 結構欄位另標 `DGROUP`；file offset 換算為 `DGROUP + 0x16140`；IDA logical／linear 不與 file 數值混列 |
| 限制 | `.i64`、授權、解包 binary 與原版素材均不進 Git；本輪未將自訂名稱寫回 IDA database |

## Window structure（confirmed）

IDA 匯出在 `logical 0x18338..0x1834d` 保留下列 writer→consumer 片段：

```text
lea si, ds:3DA8h
call sub_1F4E3
...
lea si, ds:3EDEh
call sub_1F604
```

`DGROUP 0x3DA8`（file `0x19EE8`）的原始 bytes 為：

```text
01 03 | 13 00 | 2e 00 | 2c 00 | c0 00 | 97 01 | 00 00 | 00 00
00 00 | 4e 83 | 00 00 | 00 00 | 00 00 | 02 03 | 2d 00 | 0e 00
16 00 | 30 00 | 2d 02 | 00 00
```

`sub_1F4E3`／共用 `sub_1F590` 對這份結構消費的欄位閉合如下：

| raw 欄位 | 玩家可見／流程語意 | 推論等級 |
|---|---|---|
| `+0=0x0301` | window flags／背景備份槽，不是顏色 style | `confirmed` |
| `+2=0x13` | x=`0x13×8=152` | `confirmed` |
| `+4=0x2e` | y=`46` | `confirmed` |
| `+6=0x2c` | width=`0x2c×8=352` | `confirmed` |
| `+8=0xc0` | height=`192` | `confirmed` |
| `+0x0A=0x197` | D3TXT00 record `407` 的框與標籤 | `confirmed` |
| `+0x0C..` | 由 status consumer 在視窗內填入角色欄位 | `strong`；仍需同一輸入逐幀 DOSBox 對拍 |

因此 `interface.json.field_status.window` 使用 `(152,46,352,192)`、22×12 個
16px 字格；沒有使用截圖目測的巨大全畫面黑框。`flags` 保留在本 ledger／raw
證據，不被誤命名成可替換的 frame style。

## record 407（confirmed）

`D3TXT00.TXT` record 407（`docs/data/d3txt_codes.json`，D3TXT00/FON 的原始 glyph
索引）是完整的 22×12 格：外框、性別／等級／HP／MP、力量／速度／耐力／聰明度／
運氣點數、最大 HP／MP、攻擊／守備／經驗標籤均在資料流中。`0xFFFE` 換行控制碼
原樣保存，未改成 Go 字串排版規則。

`interface.json.field_status` 另外保存 `name`、`class`、`sex` 與各數值 anchor；
`texts.json` 保存 record 407 的 Unicode transcription 與 canonical `glyph_codes`。
引擎只執行：

```text
pack text ID → record407 grid/frame
角色 record（heroStat／heroExp／heroHP／heroMP）→ pack anchors
裝備 writer／consumer → attack、defense derived values
```

英雄職業、性別值的 glyph 也由 pack 提供；production Go 不再寫入可見中文或
`H`／「進度階段」等替代標籤。缺少 `field_status`、frame 或 text ID 時，
`NewGameWithPack` 失敗即關閉。

## Runtime 視覺證據（D2）

Docker＋Xvfb 以既有 `TestDumpNewGameScreens` 的正式 `renderFrame` 產出
[`status_detail_runtime.png`](../dq3_remake_ebitan/docs/status_detail_runtime.png)。
這張圖只證明 renderer 已消費 JSON／record407 並填入目前角色值，不取代原版
同輸入 trace；`scratch_re/status407b.png` 仍是原版畫面輔助證據。

## 未閉合與下一步

- `sub_1F4E3` 的 row／callback 互動與 `看全體的情形`、`重新排序` 子選單尚未接回
  正式 InputState；本輪只閉合「詳細狀況窗」的 writer→text→數值 consumer。
- hero 以外的隊員詳細狀況、裝備／道具清單仍需各自追 `DGROUP` 結構與同狀態畫面，
  不可把本輪一個視窗外推成所有 field modal 已完成。
- 只有補上同一輸入、存讀檔與下一節點證據後，該畫面才能從 E2／V2 升至 E3／V3。
