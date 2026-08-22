# 詛咒裝備來源與教會解詛咒垂直切片（2026-08-22）

## 問題與停止線

`docs/137` 已證實教會 `sub_171CC` 會掃描角色八個 item word 的
`bit0x4000`，但當時尚未找到該 bit 的資料來源，因此 remake 正確地失敗即關閉。
本切片只閉合：

```text
ITEM.DAT metadata → 裝備選擇 writer → item word bit0x4000
→ 同部位不可自行換下 → 教會 level×100 交易 → 移除所有命中物品
```

它不把 ITEM `+5` 的每個 bit 命名成戰鬥效果，也不宣稱所有道具特殊能力皆已解讀。

## 輸入、工具與位址契約

- `assets_raw/DQ3.EXE`：115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`
- `assets_raw/ITEM.DAT`：896 bytes，SHA-256
  `7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d`
- 工具：IDA Pro 9.4＋IDAPython；Docker image
  `ida-pro-9.4-ver3:py312-x11-v2`
- 位址：IDA linear；本文 logical = IDA linear − `0x10000`；第一程式 segment 的
  file = logical + `0x1370`。
- 可重建匯出器：`tools/ida_dump_cursed_item_provenance.py`；sidecar
  `work/cursed-item-provenance-20260822.json`（gitignored）。匯出保留原始 `sub_*`、
  bytes、xref、ITEM raw record 與推論等級，不 rename、不修改原始檔。

## 已證實：ITEM metadata 形成 cursed item word

`sub_17ED9`（IDA `0x17ed9`）先將每個角色 item word 複製到 `DX`，再於
`0x17ef5` 以 `and dx,3fffh` 去掉 runtime 的 `bit0x8000`（已裝備）與
`bit0x4000`（詛咒鎖定），留下 ITEM raw ID。其後：

1. `0x17f04`：raw ID × 7，形成 ITEM record offset。
2. `0x17f09`：讀 `ITEM +4`，旋轉／遮罩後只列出目前部位的物品。
3. `0x18040`：讀 `ITEM +6` 的職業 bitmask，`0x18053..0x1806a` 判斷角色能否裝備。
4. `0x18044`：以 `mov cx,[bx+1fdh]` 讀 ITEM `+4/+5` little-endian metadata word。
5. 玩家確認新裝備後，`0x18098..0x1809b` 先 OR item word `bit0x8000`。
6. `0x1809d..0x180a4` 測試 metadata word 與 `0x0e00`；命中時
   `0x180a6..0x180a9` 再 OR item word `bit0x4000`。

因此「ITEM `+5` 的 bits `0x02/0x04/0x08` 任一設置，該物品裝備時會形成
cursed item word」為 **confirmed**。這是 writer、原始 metadata、runtime word 與
後續 consumer 的完整鏈；不是依攻略或名稱猜出的清單。

## 已證實：五筆原始物品

ITEM `+5 & 0x0e != 0` 恰有五筆：

| raw ID | D3TXT00 rec | 現行 glyph→Unicode 導覽標籤 | ITEM raw 7 bytes | `+5 & 0x0e` |
|---:|---:|---|---|---:|
| `0x1b` | 28 | 雙刃劍 | `64 00 88 13 20 08 c0` | `0x08` |
| `0x1d` | 30 | 破壞之劍 | `6e 00 f0 55 20 02 c0` | `0x02` |
| `0x37` | 56 | 不幸的頭盒 | `00 23 00 00 60 04 c0` | `0x04` |
| `0x38` | 57 | 文鬼面具 | `00 ff 02 00 60 06 ff` | `0x06` |
| `0x3d` | 62 | 悲嘆盾 | `00 23 0b 00 80 0a c0` | `0x0a` |

record 對應是原始 `D3TXT00.TXT` 的 `record = raw ID + 1`。表中文字只沿用現行
glyph→Unicode sidecar 供導覽；例如「頭盒／文鬼」仍受歷史字形語意限制，不能反過來當成
ITEM metadata 證據。raw record、glyph words 與五筆篩選結果為 **confirmed**；個別 Unicode
字義依 `docs/02`／`docs/03` 的限制，不在本切片升格。

## 已證實：自行換裝會被拒絕

`sub_17ED9` 在建立同部位候選時，若看到 item word `bit0x4000`，便於
`0x17f22` 寫暫存 `DS:0x0678=1`。選擇窗返回後，`0x18003..0x18010` 先檢查該值；
非零時顯示 D3TXT00 rec `0x0f2` 並直接離開，不執行換裝 writer。

因此 cursed 裝備不能由一般裝備面板換下，為 **confirmed**。這個 gate 只對目前部位
候選生效；不可擴張成角色所有裝備面板永久鎖死。

## 已證實：教會交易

`sub_171CC`（IDA `0x171cc`／logical `0x71cc`／file `0x853c`）：

- `0x171fc..0x1720c` 掃角色 `+0x3a` 起八個 word；沒有 `bit0x4000` 時顯示
  rec `0x12b`，不收費。
- `0x17220..0x17228` 讀 level `+0x15`，乘 `100`，只建立一次服務費。
- 確認付款成功後，`0x17254..0x17266` 再掃八格；每個命中 word 都直接寫
  `0x00ff`，不是清 bit 後放回背包。
- 最後播放 cue `0x21`、顯示 rec `0x130`，呼叫 `sub_18197` 重算裝備能力。
- 拒絕或金幣不足不改物品、不扣款。

這段交易為 **confirmed**。若同一角色同時有多件 cursed 裝備，單次付款會移除全部。

## Remake 契約

1. `dq3data.Items` 保留 ITEM 原始格式 oracle，提供 metadata word／原始詛咒判定；
   production 不從攻略或名稱判斷。
2. game pack 保存五個 raw ID、每級 100G、原始位址與 D3 evidence；共用 Go 不硬寫
   DQ3 item ID、價格或 record。
3. Go 的四個裝備槽等價於原版八個 item word 中設了 `bit0x8000` 的項目；詛咒狀態由
   「目前裝備 raw ID 是否在 pack 集合」重建，故既有 save/load 不需新增另一個易失同步的 bit。
4. 一般裝備面板不得替換 cursed 的同部位裝備。教會選定角色後，費用為 level×100；
   付款成功才將所有 cursed 槽設為空，非 cursed 槽不變。
5. pack 缺服務、清單、費率或 evidence 時必須 fail closed；不得設合理預設。

## 驗收

- ITEM 896-byte／128×7 parser parity，鎖定五個且只有五個 raw ID。
- game-pack schema、reference、raw ITEM parity test。
- 正式裝備面板 primitive：裝上 cursed item 後，同部位替換不交易。
- 教會正式 `InputState`：無詛咒、拒絕、金幣不足、成功移除全部、只扣一次費用。
- save/load round-trip 保留 cursed equipment，載入後仍受換裝與教會 consumer 約束。
- 受影響 tests、`git diff --check` 與 Docker 容器清理。

完成以上即把此垂直鏈標為 E2；沒有同狀態 DOSBox 畫面與 cue timing，仍是 V1，不能宣稱
教會畫面／音效 V3。

## 實作與驗收結果

本切片完成時的 game pack schema `0.1.37`／content `0.1.42` 已保存五筆 raw ID；後續
`docs/149` 切片當時升為 `0.1.38`／`0.1.43`，該音效欄位不改變本交易。工作樹現行版本
以根 README 與 manifest 為準。pack 保存
`gold_per_level: 100` 與 D3 evidence。`dq3data.Items` 的原始格式 oracle、一般換裝拒絕、
教會正式三選單交易及 save/load consumer 均已接線。

Docker＋Xvfb 驗收：

- parser／pack／換裝／教會／存檔 targeted tests：PASS；
- `go test ./game ./internal/... -count=1`：PASS，`game` 172.811 秒；
- 完整 `TestOpeningProductionInputTrace` 包含在上述 `game` 套件，未使用 debug shortcut，
  現行 campaign E3 維持通過。

因此本切片為 D3／E2／V1；V3 限制不變。
