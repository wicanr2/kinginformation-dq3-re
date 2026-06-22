# Bug #8:戰鬥後畫面變黃/綠且不恢復(調色盤未還原)

玩家回報:精訊版原版 DOS 在**戰鬥中畫面變成黃色/綠色,脫離戰鬥後恢復不了**。
這是 `bugs.md` 七個之外、由實機回報補充的第 8 個 bug,根因為**調色盤(DAC)未還原**。

## 現象與根因類別

精訊版用兩套調色盤(docs/04):
- `DQ3.PAL`(80 色)— 地表/城鎮 tile。
- `MNSBK.PAL`(160 色)— 怪物/戰鬥背景。

戰鬥進場時把 DAC 切成 `MNSBK.PAL`(畫怪物與戰鬥背景 packbg page 0x13);**離場路徑只重載
地圖 tile 圖庫,從不把 `DQ3.PAL` 重新套回 DAC**。於是回到地表/城鎮後,tile 仍以怪物調色盤
顯示 → 整體偏黃/綠,且因地表平常不會再主動重載調色盤,**色偏一直留著無法恢復**。

## 反組譯佐證(信心:中高 — 結構/類別層級)

戰鬥進出場三段(seg0,`tools/re_disasm.py`):

| 函式 | file off | 作用 | palette 還原? |
|---|---|---|---|
| `battle_enter_screen` sub_bfd1 | 0xd341 | 載戰鬥背景(packbg 0x13)、設視訊 mode、進場 | 切到怪物色系 |
| `battle_redraw_field` sub_bcf2 | 0xd062 | 勝利後重繪場景(NPC/角色重排) | 無 |
| `battle_leave_screen` sub_c03f | 0xd3af | `call 0xfe64`(load_blk 重載地圖 tile)+ 視訊/檔案/鍵盤還原 | **無** |

對這三段全段反組譯掃 `DQ3.PAL` 緩衝(DS:0x3232)、`MNSBK.PAL`(0x3352)、palette 指標
(0x25d1)、palette flag(0x25d5)的引用:**三段皆零命中**。離場只 `call 0xfe64` 重載 tile
圖庫,沒有任何把 `DQ3.PAL` 重套回 DAC 的呼叫。這證實「離場不還原 DAC palette」的結構性缺漏,
與回報的「脫離戰鬥後恢復不了」一致。

> 精確的「進場把 DAC 設成怪物色」那一條指令落在 packbg/VGA 載入路徑(seg 0x10xx 視訊模組),
> 該模組另含 DOS critical-error(int 24h)handler 等,位址需配 DOSBox 實機逐步確認;
> 本文先坐實**類別根因**(離場未還原 DAC),精確 byte-level 修點留 DOSBox oracle 回合。

## 修法

**原版 EXE(若日後做 binary patch)**:在 `battle_leave_screen` 重載 tile 後,補一次
「載 `DQ3.PAL` → 套回 DAC」。需要的不是改值,是**增加一段呼叫**;受 docs/22 同樣的
code cave 限制(resident 主碼無安全 cave),較可能要 C 層才穩。

**SDL2 remake(本專案主線,已從設計上免疫)**:
- 每個場景(`dq3_scene`)**自帶 palette**,進場/返回時必定重套(`dq3_scene_apply_palette`)。
- 戰鬥場景日後用 `MNSBK.PAL` 衍生 palette;戰鬥**結束返回地表/城鎮時**,驅動端對目的場景
  呼叫一次 `dq3_scene_apply_palette` → DAC 立即還原成 `DQ3.PAL`,不可能殘留怪物色。
- 此契約寫進 scene 介面:**任何場景切換都必須重套目的場景 palette**,從根本杜絕 #8。

## 與既有 palette 議題的關係

docs/04 記過:地表海面 idx2/10 由 EXE 執行期改寫 DAC(海浪動畫)。原版這種「DAC 是全域
可變狀態、誰用誰改、用完不一定還原」的模式,正是 #8 的溫床。remake 改為**每場景顯式重套
palette**(palette 不是全域隱式狀態),海面 cycling 也改成在地表場景內局部驅動,互不污染。

## 驗證

- 靜態:已確認三段戰鬥進出場碼無 `DQ3.PAL` 重套(上表)。
- 動態(待 DOSBox oracle 回合):原版進一場戰鬥→勝利離場→截地表畫面,應見黃/綠色偏;
  remake 同流程應正常(返回即重套 `DQ3.PAL`)。
