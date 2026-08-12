# 130 — 地表四人 HUD 幾何與欄位 writer 閉環

> 日期：2026-08-12。狀態：地表命令窗的隊伍 HUD 已由錯誤的寬框修回原版
> `DGROUP 0x3e9c` 幾何，姓名、H／M、職業與等級也依原始 writer 的欄內位置由
> game pack 驅動。此切片是 D3／runtime V2；原版影片是 near-state，不宣稱 V3。

## 第一個玩家可見差異

修正前 runtime `party_field_hud.png` 使用 `(48,244,448,80)`，四欄被平均撐寬，姓名、
H／M label、數值與末列又共用近似位置。原版完整實況
`dq3_real_video/frames/f000900.jpg` 的正式地表命令畫面則顯示中央 4×80px 欄：姓名壓在
上框、H／M 分列，末列為單字職業 glyph 加等級。這是主線反覆出現且一眼可見的差異。

影片原始畫格是 480×360 的縮放／黑邊錄影，因此只作玩家可見形狀閉環；canonical
邏輯座標取自下列 EXE writer，不用壓縮影片反推像素。

## 輸入與位址口徑

| 項目 | 值 |
|---|---|
| 原始輸入 | `assets_raw/DQ3.EXE`，115,282 bytes |
| SHA-256 | `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| 工具 | 一次性 Docker 內的 IDA Pro 9.4.0.260610 |
| 可重建匯出 | `tools/ida_dump_party_field_hud.idc`；sidecar 只在 `/tmp`，不加入 Git |
| 位址 | `linear=logical+0x10000`；`file=logical+0x1370`；`DS` 明列為 DGROUP |

原始檔唯讀掛載；IDA 工作副本、database、listing 與 sidecar 均只在 `/tmp`。腳本不改名
函式或資料，只匯出原始名稱、位址、caller 與指令。

## 入口 → 結構 writer → 內容 writer → 玩家畫面

地表快捷命令 owner `sub_17C83` 在 `seg000:7CCB..7CFD`：

```text
mov al, ds:5077h
mov word ptr ds:3E9Eh, 13h
mov word ptr ds:3EA0h, 0EEh
mov ds:3EA8h, ax
mov bl, 0Ah
mul bl
add ax, 4
mov ds:3EA2h, ax
lea si, ds:3E9Ch
call sub_1F590
call sub_18222
```

四人時得到 `(x,y,width,height)=(0x13×8,0xee,(4×0x0a+4)×8,0x50)`，也就是
`(152,238,352,80)`。`DGROUP 0x3e9c`（file `0x19fdc`）原始 bytes 為：

```text
0d 01 13 00 ee 00 2c 00 50 00 91 01 04 00 92 01
93 01 22 82 fe 00 0a 00
```

`sub_1F590` 消費這份結構；multi-column record 從 base `+2` byte 開始、每欄遞增
`0x0a` byte。`sub_18222` 再提供動態內容：

- `DS:716 = base+4` byte 後由 `sub_215EE` 畫姓名，且 `cmp cx,4` 明確限制四 glyph；
- H／M 數值由 `sub_219D2` 於相同 `base+4` byte anchor 寫入；
- 末列以 `0x6a + class_raw` 交給 `sub_211B6` 畫單一職業 glyph，再於 `base+4`
  寫等級；
- 每欄結束 `base += 0x0a` byte，故欄距固定 80px。

以上 raw、writer 與 consumer 均為 `confirmed`。影片 `f000900` 顯示同一正式命令入口及
相同資訊層級，將玩家可見閉環升為 D3；因人物數值與捕捉狀態不同，畫面只標 near-state
V2，不稱逐像素 V3。

## remake 接線與驗收

`interface.json.party_hud` 現保存 window 幾何、label 起點、姓名／數值相對位移、姓名上限與
各 class 的原始單字 glyph。共用 renderer 只消費 `PartyHUDLayout`，沒有 DQ3 座標、職業名或
`0x6a` 公式。缺 glyph map 或欄內位置會在 pack 載入時失敗即關閉。

`TestDQ3PartySpriteAndHUDContract` 鎖定原始 pack 值；
`TestPartyHUDUsesOriginalPackColumnAnchors` 鎖定姓名截斷與四列 anchors；
`TestPartyHUDOpensThroughProductionInput` 與 `TestDumpPartyFieldPNG` 均以正式
`InputState.Confirm` 開啟命令窗，而非直接設 `cmd.open`。新 runtime PNG 仍為 V2 fixture，
不取代原版同狀態 oracle。
