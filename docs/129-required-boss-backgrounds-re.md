# 129 — 必經頭目固定背景的 IDA 接線

> 日期：2026-08-12。狀態：所有必經、由固定單隻編隊啟動的頭目戰，已改由原版
> `fixed_records` 的 header `raw[1]`／`raw[2]` 選擇背景 archive page 與 palette bank。
> 這是 D2 靜態資料接線與 runtime V1；除日邦格八頭大蛇兩戰外，尚無同狀態原版畫格，
> 不宣稱 V2／V3 視覺 parity。

## 輸入、工具與位址口徑

| 項目 | 值 |
|---|---|
| 原始輸入 | `assets_raw/DQ3.EXE`，115,282 bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| 背景輸入 | `PACKBG.SCR` SHA-256 `4e226cd23c38bbbce9db974da559933a2e84669b3abb1559ab9755c85ed800d0`；`MNSBK.PAL` SHA-256 `e2985503b3eb5256bcaf44f2f1b0a0b1ab124e66873c2de210e0d25ffd75110d` |
| 工具 | Docker 一次性 `ida-pro-9.4-ver3:latest` 的 IDA Pro 9.4.0.260610；可重建 IDC 匯出為 [`tools/ida_dump_mandatory_boss_background.idc`](../tools/ida_dump_mandatory_boss_background.idc) |
| 位址 | IDA linear = logical + `0x10000`；primary file = logical + `0x1370`；`seg000:xxxx`、`DS:xxxx` 與 file offset 不混用 |

原始 EXE 以唯讀方式掛入容器；工作副本、IDA database 與 listing 都只留在 `/tmp`，未加入 Git。

## 原版固定 record → 共用背景 consumer

IDA 匯出保留的原始 caller 顯示，下列五條都以 `lea si, DS:4E..` 呼叫
`sub_1BE89`：

```text
sub_14312  → DS:4EE4  ; 怪力魔
sub_1622A  → DS:4EDF  ; 巴拉摩斯
sub_165E7  → DS:4EFA  ; 巴拉摩斯怨靈
sub_1661E  → DS:4EFF  ; 巴拉摩斯殭屍
sub_164CD  → DS:4EF0  ; 索瑪

sub_1BE89 → sub_1BF35 → sub_1BFD1 → sub_1C688 → sub_1C6E5
```

`sub_1BF35` 的原始指令將 `[si+2]` 寫至 `DS:0D73`、將 `[si+1]` 寫至
`DS:0D71`；`sub_1BFD1` 再直接呼叫 `sub_1C688`。後者以
`DS:0D73 * 0x30` 選擇 `MNSBK.PAL` bank，並在 fixed-battle 的既有 direct branch
把 `DS:0D71` 交給 `sub_1C6E5`。`sub_1C6E5` 以 page × `0x13d80` seek
`PACKBG.SCR`，讀取可見 `0xcf80` field。完整 archive 格式見 [`docs/128`](128-battle-background-selector-re.md)。

| 玩家必經戰 | 原始 record | raw bytes | page／bank | 靜態推論等級 |
|---|---|---|---|---|
| 沙曼歐莎怪力魔 | `DS:0x4ee4`，file `0x1b024` | `01 1f 06 59 01` | 31／6 | raw、caller、writer、archive consumer：已證實；此 event 的 direct-branch mode：強推論 |
| 巴拉摩斯 | `DS:0x4edf`，file `0x1b01f` | `01 27 06 79 01` | 39／6 | 同上 |
| 巴拉摩斯怨靈 | `DS:0x4efa`，file `0x1b03a` | `01 2a 08 7a 01` | 42／8 | 同上 |
| 巴拉摩斯殭屍 | `DS:0x4eff`，file `0x1b03f` | `01 2a 08 7b 01` | 42／8 | 同上 |
| 索瑪 | `DS:0x4ef0`，file `0x1b030` | `01 2d 08 7c 01` | 45／8 | 同上 |

「direct-branch mode」仍只標強推論，是因目前輸入沒有逐一同狀態的原版記憶體／畫格 trace；
它不影響 raw header、writer、archive seek 或 palette-bank consumer 的已證實定位。任何未來反證
必須保留這份 record／位址索引後再勘誤，不能以改名或重寫歷史掩蓋。

## runtime 接線與失敗即關閉

`gamepack.FixedBattleFormationForSingleMonster` 只從 `battle.json.encounter.fixed_records` 解出
單群、單隻、唯一的原始 record，再傳給共用 `startFormationWithBackground`。它不含 DQ3 page、
palette、怪物名稱或畫面 fallback。

- 若 monster 沒有對應 record，或同一 monster 有兩個固定 record，回傳失敗；runtime 不會任選
  一筆或退回草地。例如八頭大蛇 `0x4b` 有兩場不同 record，仍只能走 `staged_boss` 的明確
  formation。
- 甘達特首次、巴哈拉達守衛／二次甘達特與八頭大蛇兩戰，本來已各由 data event 直接持有
  formation selector；不經這個唯一單隻 accessor。
- 因此所有必經頭目使用 pack raw selector，而 generic terrain encounter 仍僅保留已知
  `0/0` baseline；本輪沒有反推或宣稱完成 terrain→page 的通用 selector。

## 驗收與界線

- `TestBuiltinDQ3BattlePackMatchesOriginalRawTables` 現逐 byte 對照五個 record，並驗證八頭大蛇
  的重複 record 不可被任意選取。
- `TestRequiredFixedBossBackgroundsUsePackSelectors` 以正式 `NewGame` 載入 pack，逐一啟動
  怪力魔、巴拉摩斯、怨靈、殭屍、索瑪，確認 renderer decoded field 與同一個 pack selector 的
  page／palette bank 一致。
- `TestMirrorProductionInputTrace`、`TestBaramosProductionInputTrace`、`TestZomaNaturalTalkEntry`、
  `TestZomaSeq` 與日邦格／甘達特 targeted trace 都由 Docker＋Xvfb 通過；正式 campaign replay
  另由本輪重新跑到 `THE END`。

這一批消除了必經單頭目「明明有原始固定背景 record，卻從舊單敵入口落回 generic grass」的
設定錯配。它不補逐動作 timing、PCM、palette transition，也不把沒有原版同狀態畫格的五戰升格
為 V2 或 V3；這些仍遵守 [`docs/123`](123-static-battle-daynight-re.md) 的停止條件。
