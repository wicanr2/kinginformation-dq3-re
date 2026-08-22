# 戰鬥逃跑 VOC 與等待契約

> 狀態：2026-08-22，D3／E2。這一輪只閉合玩家逃跑成功與敵人逃跑成功；其他物理、
> 受傷及咒文 cue 保留 raw call-site，不在語意未完整對映前批次接線。

## 問題

原版戰鬥會在特定 action 播放 VOC，並於後續呼叫 `sub_208E2` 等待 Sound Blaster driver。
本切片接線前，remake 的 `Battle` 完全沒有 SFX consumer，只有地表選單與少數事件呼叫
`PlaySFX`，因此逃跑訊息可立即被略過，且戰鬥本身無原版 cue；下方實作與驗收結果已訂正
這個歷史差距。

## 輸入、工具與非破壞性輸出

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/FVOC.VCX`：72,531 bytes；SHA-256
  `47be27d2244e314ca6f2a21d3fa61a95d4e823f9085ed3d21808ac25d66b39c5`。
- `assets_raw/NVOC.VCX`：143,156 bytes；SHA-256
  `53c710f356b728e97287bfb64af7f64cf2067b28f83a9dfe32f99a6eb61f4156`。
- Docker 內 IDA Pro 9.4；腳本 `tools/ida_dump_battle_sfx_wait.py`。
- gitignored sidecar：`work/battle-sfx-wait-20260822.json`。
- 腳本只讀輸入，保留原始 `sub_*`、IDA linear 位址、bytes、xref、VOC descriptor 與
  推論等級；不 rename、不寫 database comment、不修改原始檔。
- 主段位址：`logical = IDA - 0x10000`、`file = logical + 0x1370`；segmented driver
  函式保留 IDA linear，不混成主段 file offset。

## 已證實的 dispatch → driver → wait

共用鏈為：

```text
caller writes BP raw cue
  → sub_20770 (IDA 0x20770)
  → DS:0x253e[(BP<<2)]
  → sub_22CF5 command 6
  → sub_22F60 → sub_236B4
  → sub_208E2 waits driver completion
```

IDA 對 `sub_22CF5` 的直接 caller 只有 `sub_20770:0x207aa`；driver dispatch 由
`sub_22F60` 跳至 `sub_236B4`。`sub_208E2` 有多個明確戰鬥 call-site，證明它不是單純
固定 tick delay，而是播放後的 driver completion gate。

## 玩家逃跑成功

`sub_1B3F3` 依玩家 command kind 進 `sub_1B47E`。其成功分支：

1. `0x1b4d3`：`mov bp,0x0d`；
2. `0x1b4d6`：呼叫 `sub_20770`；
3. `0x1b4db..0x1b4de`：輸出原始 record `0x15c`；
4. `0x1b4e3`：呼叫 `sub_208E2`；
5. `0x1b4ef`：將 battle result state 寫成 `2`。

因此玩家逃跑成功使用 raw cue `13`，而不是敵人逃跑的 cue `21`。`FVOC.VCX` cue13：
offset `0x8498`、type1、block length `3521`、rate byte `107`、codec0、3519 samples、
source rate 6711 Hz、source duration 524,362,986 ns。

## 敵人逃跑成功

`sub_1A973` 的怪物逃跑成功分支：

1. `0x1aa37`：`mov bp,0x15`；
2. `0x1aa3a`：呼叫 `sub_20770`；
3. `0x1aa3f..0x1aa42`：輸出原始 record `0x15b`；
4. `0x1aa47`：呼叫 `sub_208E2`；
5. 返回 action queue。

因此敵人逃跑成功使用 raw cue `21`。`FVOC.VCX` cue21：offset `0x11146`、type1、
block length `2568`、rate byte `165`、codec0、2566 samples、source rate 10989 Hz、
source duration 233,506,233 ns。

## Remake 契約

- versioned `battle.json` 保存兩個穩定角色：`player_fled`→raw cue13、
  `enemy_fled`→raw cue21；兩者 `wait_for_completion=true` 並附 D3 evidence。
- Go engine 不得硬寫 raw cue。`Battle` 只知道穩定角色、播放介面及「完成前不可接受下一次
  訊息確認」的狀態機。
- `gaudio` 即使沒有可用 host audio device，也要保存 VOC source duration；播放可以靜音
  降級，但 deterministic message gate 不能因 headless 環境消失。
- 每個帶 cue 的 `battleMessage` 在真正顯示時才播放，並以 source duration 向上換算成
  60 TPS 的最少等待幀。這是可重建的 E2 scheduling approximation；不宣稱 Sound Blaster
  DMA、Ebitengine backend 或真實喇叭 wall-clock 完全相同。
- 等待期間的 Confirm 必須被忽略；倒數結束後才能切下一筆訊息或離開戰鬥。

## 驗收與限制

- pack validation：兩個角色必須存在、raw cue 非負、等待 gate 開啟且 evidence 合法。
- raw parity：pack cue13／21 與上述 IDA call-site bytes 一致；FVOC descriptor 一致。
- component：fake audio 證明播放的 cue、向上取整幀數與 Confirm gate。
- normal path：正式 battle bootstrap 注入同一 `gaudio.Music`；完整 campaign trace 不退步。
- 本輪不宣稱：cue 的擬聲中文名稱、NVOC／FVOC 硬體選擇語意、其他 action cue 對映、實際
  PCM 波形、host wall-clock 或逐幀 V3。這些需各自的 caller 或原版錄音／frame trace。

## 驗收結果

- IDA sidecar 已由 IDA Pro 9.4／IDAPython 重新產生；玩家／敵人 writer、dispatch、wait
  call-site 都同列保留原始位址、bytes、推論等級與證據來源。輸入三檔雜湊與位址空間驗證通過。
- `TestBattleFleeSFXWaitGate`、pack EXE／FVOC parity、`internal/gaudio` 與
  `internal/gamepack` tests 通過。
- 正式 `TestOpeningProductionInputTrace` 由標題抵達 `THE END`（140.93 秒）；完整
  `game` 套件在 1 CPU／4 GiB Docker＋Xvfb 通過（162.799 秒），完整 `internal/...` 通過。
- 初版 headless 實作雖只需 duration，仍把所有 VOC 複製成雙聲道 PCM，造成完整 suite
  記憶體放大；現已修成無音訊後端只保存 duration、不配置不可播放 PCM，並有回歸測試。
  可用音訊後端仍保存並播放 PCM，規則時序不因靜音降級消失。
