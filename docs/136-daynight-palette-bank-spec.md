# 136 — 原版日夜 clock／palette bank 接線規格

> 日期：2026-08-22。範圍只包含地表／城鎮 16 色 palette bank 選擇與黑暗之燈的
> clock writer；NPC 日夜雙表沿用 `docs/60`。本規格不宣稱 DOSBox DAC capture 或
> 同狀態逐像素 V3。

## 1. 輸入、工具與非破壞性契約

- 輸入：`assets_raw/DQ3.EXE`，115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- palette：`assets_raw/DQ3.PAL`，240 bytes，SHA-256
  `178a9ca809a33108d8e2a6430796eae8909d98e514fe543fc1923139110389c4`。
- 工具：Docker image `ida-pro-9.4-ver3:py312-x11-v2`，IDA Pro 9.4＋IDAPython。
- 可重建匯出器：[`tools/ida_dump_daynight_palette.py`](../tools/ida_dump_daynight_palette.py)。
  它保留原始 `sub_*` 名稱、IDA linear address、bytes、caller 與輸入 hash，不 rename、
  不寫回原 EXE；`.id0/.id1/.nam/.til` 與完整 sidecar 留在 `/tmp`。
- 位址：IDA linear = logical + `0x10000`；主 seg0 file = logical + `0x1370`。

## 2. IDA 9.4 證據

### 2.1 clock writer（confirmed）

`sub_1EE23`（IDA `0x1ee23`、logical `0xee23`、file `0x10193`）：

```text
0x1ee3d  inc word ptr ds:251Dh
0x1ee41  cmp word ptr ds:251Dh, 78h
0x1ee48  mov byte ptr ds:526Ch, 0
0x1ee4e  cmp word ptr ds:251Dh, 0F0h
0x1ee56  mov word ptr ds:251Dh, 0
0x1ee5c  mov byte ptr ds:526Ch, 1
0x1ee62  mov ax, ds:251Dh
0x1ee65  mov bl, 14h
0x1ee67  div bl
0x1ee69  cmp ah, 0
0x1ee6e  call sub_1EE76
```

因此 clock 是 240 tick 循環；`0x78` 起為夜間，`0xf0` wrap 為日間。只有
`clock % 0x14 == 0` 才更新 palette。

### 2.2 palette selector（confirmed）

`sub_1EE76`（IDA `0x1ee76`、file `0x101e6`）及不直接 upload 的
`sub_1EE9B`（IDA `0x1ee9b`、file `0x1020b`）執行同一選擇：

```text
segment = clock / 0x14
bank = byte ptr [DGROUP 0x25c5 + segment]
source = DGROUP 0x3232 + bank * 0x30
sub_1EE76 -> sub_20A3A(source)  // upload 16×RGB6 = 0x30 bytes
```

`DQ3.EXE` file `0x18705` 的 12 bytes 為：

```text
01 00 00 00 01 02 03 04 04 04 03 02
```

故 12 個 20-tick segment 的 bank 順序是
`[1,0,0,0,1,2,3,4,4,4,3,2]`。這是原版 table 與 consumer 的
writer→table→consumer 閉環，等級為 `confirmed`。

### 2.3 palette asset loader（confirmed）

`sub_1ECDC`（IDA `0x1ecdc`、file `0x1004c`）讀 `0xf0` bytes 到
`DGROUP 0x3232`。`0xf0 / 0x30 = 5`，與 `DQ3.PAL` 的五個 16 色 bank 完全相符。
remake 必須直接選原始 bank，不能再以 RGB 百分比推導近似色。

### 2.4 黑暗之燈（confirmed）

既有 `docs/93` 已閉合 item `0x5f` handler：日間使用時逐次增加 clock，最後固定寫
`0x008c`（140）並刷新 palette。現行 `day_night_phase:2 + step reset` 只得到 clock
120，不能精確選到原版 segment 7／bank 4；pack 必須保存 `day_night_clock:140`。

## 3. production 資料契約

`events.json.day_night_cycle` 保存：

- `clock_ticks:240`、`night_start_tick:120`；
- `palette_segment_ticks:20`；
- `palette_entries_per_bank:16`；
- `palette_bank_indices:[1,0,0,0,1,2,3,4,4,4,3,2]`；
- `palette_asset_key:"world_palette"`；
- D3 evidence（原始 EXE 位址、consumer 與本文件）。

`manifest.json.assets.world_palette` 鎖定 `DQ3.PAL` 的 size／SHA-256。共用 Go 引擎只知道
clock、segment、bank 與 palette slice，不知道 DQ3 的表值。缺欄位、表長不符、bank 越界、
asset hash 不符一律 fail closed。

`item_use_effects` 的黑暗之燈新增 `day_night_clock:140`；引擎以 cycle 驗證 clock，
再換算既有持久化 `(dnPhase,dnStep)`，不新增 DQ3 raw 常數。

## 4. runtime 與驗收

1. 每個合法地表步仍推進一 tick；跨 20-tick segment 時立即換 bank。
2. `isNight` 由 clock 是否落在 `[120,240)` 判斷；NPC 日／夜表仍只在二值邊界切換。
3. render 前由目前 clock 選 16 色 bank，確保剛進城、轉場與讀檔都不殘留舊 palette。
4. 測試重讀 pack JSON 與真實 `DQ3.PAL`，鎖定 12 個 segment 的 bank 起點顏色、
   clock 119／120／239／wrap 及黑暗之燈 140。
5. 正式 input trace 抽跑黑暗之燈 transaction 與 save/load；本切片最高為 D3／E2。
   沒有 DOSBox 同 clock DAC screenshot 前仍標 V2，不宣稱 V3。
