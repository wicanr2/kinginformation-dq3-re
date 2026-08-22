# 一般物理攻擊結果、VOC 與死亡訊息契約

> 狀態：2026-08-22。敵方一般物理入口、雙方一般物理成功傷害、共同 miss 與死亡文字
> 分支為 `confirmed`；remake 目標為 D3／E2。本切片不接 `word_272E1==3` 的 cue9 分支、
> critical／特殊 effect、咒文傷害、NVOC bank selector 或硬體 wall-clock。

## 問題與停止線

`docs/150` 已閉合玩家正常物理攻擊的 cue6 → record330 → wait → cue11 → wait，但刻意
停在 `sub_1ACCE` 前。現行 Go 還有四個玩家可見差異：

1. 敵人一般物理攻擊沿用玩家 `actor_attack`，因此錯讀 record330；原版是 record331。
2. 敵人打玩家後沿用 record332；該 record 的 `0xffed` 是敵人名稱插值，原版玩家受傷
   使用 record333 的 `0xfffb`。
3. 敵人 attack cue4、成功物理傷害 cue1 與共同 miss cue3 都沒有接入。
4. 一般物理使敵人／玩家 HP 歸零後，原版分別顯示 record336／357；現行只直接進勝利或
   全滅摘要。

cue9 位於 `sub_1ACCE` 的 `word_272E1==3` 路徑，且帶另一組 cue4 → wait → cue9／
record334 控制流；其 action mode 與玩家可見語意尚未獨立閉合。本規格保留為 `unknown`，
不因同在 common damage function 就併入一般物理。

## 輸入、工具與非破壞性輸出

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/D3TXT00.TXT`：原始 record table；record330–357 由 typed decoder 讀取。
- `assets_raw/FVOC.VCX`：72,531 bytes；SHA-256
  `47be27d2244e314ca6f2a21d3fa61a95d4e823f9085ed3d21808ac25d66b39c5`。
- `assets_raw/NVOC.VCX`：143,156 bytes；SHA-256
  `53c710f356b728e97287bfb64af7f64cf2067b28f83a9dfe32f99a6eb61f4156`。
- Docker 內 IDA Pro 9.4／IDAPython；匯出器 `tools/ida_dump_physical_sfx.py`。
- gitignored sidecar `work/physical-sfx-20260822.json`；schema
  `dq3.ida_physical_sfx.v1`，保留原始函式名、bytes、CFG、xref、VOC descriptor、工具版本、
  輸入 hash 與逐筆 inference level，不 rename、不修改原始 EXE 或 database。
- 主段換算：`logical = IDA linear - 0x10000`；`file = logical + 0x1370`。

## 原版控制流

### 敵方一般物理入口

`sub_1A973` 在 IDA `0x1aa65`／`0x1aa79` 呼叫 `sub_1AC05`。入口固定為：

```text
IDA 0x1ac0b  mov bp,4
IDA 0x1ac0e  call sub_20770
IDA 0x1ac13  mov di,0x14b
IDA 0x1ac16  call sub_21414
...
IDA 0x1acb8／0x1acbf → sub_1ACCE
```

因此 cue4 與 record331 的 call-site／順序為 `confirmed`。`sub_1AC05` 本身沒有在文字後
立即等待；一般敵方命中路徑進 `sub_1ACCE` 後，才於 IDA `0x1adcb` 等 cue4 完成。
remake 的訊息佇列把這個 completion gate 綁在 enemy-attack 訊息上，保持玩家可見順序與
輸入阻擋，但不宣稱內部傷害計算發生的 CPU 時點與 DOS 完全相同。

### mode 分流與成功傷害

`sub_1ACCE` 對 `word_25B45 & 2` 的 direct xref 已由 IDA database 匯出。兩個分支可由
原始 record 控制碼反向驗證受擊者：

```text
bit2 != 0（敵人是受擊者）
  0x1adae cue1 → sub_20770
  0x1adb9 sub_1B3C3
  0x1adc2 record0x14c (332；0xffed enemy-name control)
  0x1ade4 顯示 → 0x1ade9 sub_208E2

bit2 == 0（玩家是受擊者）
  0x1adc8 sub_1F470 → 0x1adcb sub_208E2（等入口 cue4）
  0x1add0 cue1 → sub_20770
  0x1add8 sub_18222
  0x1addb record0x14d (333；0xfffb player-name control)
  0x1ade4 顯示 → 0x1ade9 sub_208E2
```

故成功一般物理傷害的 cue1、雙方 record 與 completion wait 為 `confirmed`。record332
繼續由現行 `actor_damage`（敵人受傷）role 擁有；新增 `party_damage` 指向 record333。

### 共同 miss

一般傷害公式的 miss branches 進 IDA `0x1aeb9`，呼叫 `sub_1AFC6`：

```text
0x1afc6 mov bp,3
0x1afc9 call sub_20770
0x1afce mov di,0x14f        ; record335
0x1afd7 call sub_21414
0x1afdc call sub_208E2
```

這個 subroutine 同時由兩種 mode 的 common damage path 呼叫，所以 cue3／record335／wait
是共同 miss result，為 `confirmed`。它不證明所有咒文 miss 或任意 runtime failure 都應播放
cue3；pack role 限定為 `physical_miss`。

### 死亡文字

傷害訊息完成後，saved HP `BP<=0` 進 `0x1af86`。IDA `0x1af94` 再測同一 mode bit：

- `bit2==0`：`0x1af9c` 設玩家 death bit，`0x1afa4` 選 record357／`0x165` 並顯示。
- `bit2!=0`：`0x1afad` 設 enemy-result state，`0x1afb3` 選 record336／`0x150`，顯示後
  呼叫 `sub_1B23F` 並遞減 alive enemy count。

兩筆 record 都沒有新 VOC dispatch；因此只新增文字 role，不猜死亡 cue。群戰只在實際
HP 由正值降至零時排一次個體死亡訊息，之後才接勝利或全滅摘要。

## 原始文字與 VOC descriptor

| role | record | raw 開頭／用途 |
|---|---:|---|
| `enemy_attack` | 331／`0x14b` | `0xffed,149,623,624,57`；敵人名稱插值 |
| `actor_damage` | 332／`0x14c` | `0xffed,...,0xfffa,...`；敵人受傷／數值 |
| `party_damage` | 333／`0x14d` | `0xfffb,...,0xfffa,...`；玩家受傷／數值 |
| `physical_miss` | 335／`0x14f` | 無名稱插值；共同 miss |
| `enemy_defeated` | 336／`0x150` | `0xffed,...`；單一敵人倒下 |
| `actor_died` | 357／`0x165` | `0xfffb,...`；單一玩家倒下 |

FVOC raw descriptor：

| cue | file offset | block length | source rate | samples | source duration | 60 TPS ceil |
|---:|---:|---:|---:|---:|---:|---:|
| 1 | `0x4e6` | 1163 | 6493 Hz | 1161 | 178,807,947 ns | 11 |
| 3 | `0x12b6` | 1511 | 6493 Hz | 1509 | 232,404,127 ns | 14 |
| 4 | `0x18a2` | 1155 | 6493 Hz | 1153 | 177,575,850 ns | 11 |

descriptor 與 source duration 為 `confirmed`；可聽名稱、NVOC selector、Sound Blaster DMA
wall-clock、重取樣波形與逐幀畫面同步仍為 `unknown`／V3。

## Remake 契約

- `battle.json.sound_cues` 新增 `enemy_physical_attack`、`physical_hit`、`physical_miss`；
  raw cue 全由 pack 擁有，三者皆需 D3 evidence 與 completion wait。
- `interface.json.battle_texts` 新增 `enemy_attack`、`party_damage`、`enemy_defeated`、
  `actor_died`；`texts.json` 保存實際字串、glyph/control words 與 record evidence。
- 玩家／同伴正常物理命中：既有 attack sequence → `actor_damage + physical_hit` →
  HP 歸零時 `enemy_defeated` → victory messages。
- 敵人正常物理命中：`enemy_attack + enemy_physical_attack` →
  `party_damage + physical_hit` → HP 歸零時 `actor_died` → 必要時 party defeated。
- 雙方一般物理 miss：仍先排各自 attack 訊息，再排 `physical_miss + cue3`。
- Go 只實作 message/cue roles 與有限狀態機；不得寫 DQ3 cue／record 常數或缺資料 fallback。

## 驗收

- IDA sidecar 必須包含 cue3、`sub_1AC05`／`sub_1ACCE`／`sub_1AFC6`／死亡 blocks、
  `word_25B45` xref、工具版本、輸入 hashes 與 inference level。
- pack validation 與 EXE／D3TXT／FVOC raw parity 鎖定 cue1／3／4、records331／332／333／
  335／336／357、mode 分流與等待。
- component tests 驗證玩家／同伴／敵人 hit、miss、death 的 message role、cue order、等待幀數，
  且死亡訊息只排一次。
- 正式新遊戲 `InputState` trace 仍能抵達 `THE END`；完整 `game`／`internal/...` 與 desktop
  build 不退步。
- 停止線：cue9、critical／特殊 effect、咒文 damage、NVOC selector、硬體 wall-clock 與
  SHP animation 另開 spec；本切片不得宣稱完整戰鬥 V3。

## 2026-08-22 實作與驗收結果

- 本切片落地時的 game-pack schema `0.1.40`／content `0.1.45` 已加入三個 sound roles、
  並由 strict validator 與 reference validation 失敗即關閉。
- 玩家／同伴／敵方的一般物理 hit、miss、death 已由正式 Battle action path 接線；敵方
  record331 與玩家受傷 record333 不再誤用 record330／332，個別死亡訊息只在 HP 由正值
  降為零時排一次。
- `TestBuiltinDQ3CommonPhysicalResultMatchesOriginal` 已直接鎖定 DQ3.EXE call-site bytes、
  D3TXT records331／332／333／335／336／357 與 FVOC cue1／3／4 descriptors。
- cue completion component tests 已鎖定 cue4／1／3 各 11／11／14 frame；hit／miss／death
  佇列與勝利訊息順序 tests 通過。
- `go test ./internal/... -count=1` 全綠；排除已單獨驗收 campaign 的其餘完整
  `game` tests 全綠（63.138 秒）；`TestOpeningProductionInputTrace` 在乾淨 Xvfb session
  由標題抵達 `THE END`（134.13 秒）；隔離 Linux desktop build 通過。
- 將整個 `game` 套件與 campaign 放在同一測試程序會因先前 Ebitengine 測試資源累積造成
  容器 OOM，加入低 `GOMEMLIMIT` 又會使 campaign GC 壓縮抖動至逾時；因此 current gate
  採上述互補集合。這是驗證環境限制，不改變各集合的 assertion，也不冒稱單程序全綠。
