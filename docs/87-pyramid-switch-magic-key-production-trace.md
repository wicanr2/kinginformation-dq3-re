# 87 — 金字塔雙開關與魔法鑰匙 production trace

> 狀態：E3／V2。更新：2026-07-29。
>
> 範圍：由諾亞尼爾甦醒後的合法 checkpoint，以正式 `InputState` 經阿莎拉慕、沙漠祠堂、
> 依席斯抵達 CTY13 金字塔，解開石門、取得魔法鑰匙並完成存讀檔。

## 1. 本輪問題

舊引擎已能載入金字塔與一般事件，但玩家由 CTY13 sec2 `(14,23)` 出生後無法走到
`(13,6)` 的魔法鑰匙寶箱。第一個真正 blocker 是 `(13–14,10–11)` 的 2×2 石門，
不是碰撞或入口座標錯誤。

本輪不用攻略值或 C remake 補設定；入口、開關、旗標、陷阱、文字、音效與寶箱都由
原始 EXE／CTY／D3TXT 重新閉合。

## 2. 原版資料與 consumer

### 2.1 CTY13 sec2

section base 是 CTY file `0x1a1d`：

- `section+4 = 0x0019`，特殊 handler table 位於 file `0x1a36`，前三筆為
  `15 16 17`，即 handler 21／22／23。
- event subid0 是 `{type=0,param=0x62,p2=0x46}`；flag `0x46` set 時，
  `(13,10)/(14,10)/(13,11)/(14,11)` 四格石門存在。
- event subid1 是 `{type=1,item=0x56,p2=0x73}`，即魔法鑰匙。
- event subid2 是 `{type=1,item=0x41,p2=0x74}`。
- transition subid3 是 `{cty=13,section=1,x=13,y=28}`。
- 普通地板開關：
  - `(25,25)`：subid1／handler21；
  - `(4,25)`、`(23,25)`：subid2／handler22；
  - `(2,25)`：subid3／handler23。

### 2.2 DQ3.EXE

三個入口分別是：

- handler21：logical `0x55f9`／file `0x6969`；
- handler22：logical `0x5659`／file `0x69c9`；
- handler23：logical `0x5682`／file `0x69f2`。

閉合的狀態交易：

1. 三者先顯示 D3TXT03 rec86 並詢問是／否；否不改狀態。
2. 選是顯示 rec87。暫存 byte `DGROUP 0x67a` 初值 0；第一鍵只寫 1。
3. 已武裝時再按 handler21／22，顯示 rec88，使用 transition subid3 落到
   CTY13 sec1 `(13,28)`。
4. 已武裝時按 handler23，顯示 rec89；file `0x6a3b` 以 `BX=0x46`
   清 story flag，呼叫 file `0x4748` 重建事件格，並以 raw `BP=0x0b` 播音效。
5. 通用 event-tile rebuild 看到 flag `0x46` 已清，將四格石門 tile `+1` 並清 event subid，
   玩家因而可通行。

因此正確攻略「最右後最左」是原版的一個成功序列，但引擎實作依原始 consumer 保留較寬鬆
語意：任一宣告開關可作第一鍵，第二鍵為 completion subid3 才成功；其他第二鍵觸發陷阱。

### 2.3 玩家可見文字

D3TXT03 已逐 word 遷入 `texts.json`：

| record | 文字 ID | 用途 |
|---:|---|---|
| 86 | `dq3:text.pyramid.switch.prompt` | 是否按牆上按鍵 |
| 87 | `dq3:text.pyramid.switch.pressed` | 主角按下按鍵 |
| 88 | `dq3:text.pyramid.switch.trap` | 陷阱 |
| 89 | `dq3:text.pyramid.switch.success` | 前方石塊鬆動 |

## 3. remake 對應

資料放在 `internal/gamepack/packs/dq3_cht/data/events.json` 的
`two_step_floor_switch_gates`。Go 只實作通用有限狀態機：

- pack 宣告場景、開關 tile/subid/raw handler、completion subid；
- pack 宣告 clear flag、陷阱 transition/destination、文字 ID、SFX cue；
- pack 宣告解鎖寶箱的原始 event/item/present flag；
- 缺欄位、非法範圍、斷開的 treasure 或不存在的文字引用，loader 直接拒絕。

魔法鑰匙舊有 `{13,2,1,1,86,115}` Go table entry 已刪除；取得流程現在只消費 pack
`unlocked_treasure`，採原版 `present flag set=可取、clear=已取` 語意。

## 4. 驗證

- `TestDQ3PyramidSwitchGateMatchesOriginalEXECTYAndText`
  - 鎖定 CTY handler/event/transition bytes；
  - 鎖定 EXE 三個 handler 與 clear flag 指令 anchor；
  - 四段文字逐 word 對 D3TXT03。
- `TestTwoStepFloorSwitchGateDeclineDoesNotArm`
- `TestTwoStepFloorSwitchGateSuccessOpensGateAndTreasure`
- `TestTwoStepFloorSwitchGateWrongSecondSwitchTraps`
- `TestOpeningProductionInputTrace`
  - 從標題／新遊戲一路走到金字塔；
  - 正式踩最右、最左開關並選是；
  - 不踩轉場格的同區路徑仍逐步送入 `Game.step`；
  - 正式命令窗選「調查」取得魔法鑰匙；
  - save/load 後 item `0x56` 保留、flag `0x73/0x46` 仍清、石門仍可通行。

runtime 圖：

- `dq3_remake_ebitan/docs/pyramid_switch_prompt.png`
- `dq3_remake_ebitan/docs/pyramid_magic_gate_open.png`
- `dq3_remake_ebitan/docs/pyramid_magic_key.png`

三張圖由 production renderer 重產兩次，byte-for-byte 相同；已目視確認牆上按鍵對白、
開啟的石門與「魔法鑰匙」取得通知。V3 仍需原版 DOSBox 同狀態截圖逐像素／版面仲裁。

## 5. 未完成

這只把 boot 起連續 trace 推進到魔法鑰匙 checkpoint，不代表金字塔後或整個 campaign 完成。
下一輪從此存檔使用正式輸入穿過魔法鑰匙門，記錄下一個玩家實際 blocker。
