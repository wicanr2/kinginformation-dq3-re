# Remake DEBUG 口(headless 腳本驗證)

> 用環境變數 `DQ3_DEBUG` 在遊戲啟動後直接觸發事件/設定狀態,配 `DQ3_DUMP` 一幀截圖,
> **不用真的玩到那一段就能驗證**(scripted verification)。實作:`main.c` run_game setup 後解析。

## 用法

```
DQ3_DEBUG="<指令>;<指令>;…" DQ3_DUMP=out.ppm  dq3_remake <assets_dir> game
```
- 指令以 `;` 分隔,依序套用。
- 設 `DQ3_DUMP` 時:套完 debug → 直接渲染當前場景 → 寫 PPM → 結束(跳過 demo walk)。
- 不設 `DQ3_DUMP`:套完 debug → 進互動迴圈(從該狀態開始玩)。

## 指令

| 指令 | 效果 |
|---|---|
| `descent` | scripted_event 86 下降:切下層 overworld(field_under)、置玩家於下層城 CTY77 入口附近 (84,68)、設 `DQ3_FLAG_DESCENDED`。共用 `do_descent()`(U 鍵亦同)|
| `ascend` | 切回地表 overworld |
| `warp:CTY:X:Y` | 載入 CTY 的 section0,置玩家於 (X,Y)(進城/迷宮;含 load_field_hero)|
| `party` | 建測試隊(勇者/戰士/僧侶/魔法使者 4 人,名=glyph 1-4)→ 名冊+隊伍 |
| `event:N` | 直接跑 scripted_event N(0x56=下降 do_descent;其餘 → dq3_scripted_event_run,如 0x53 彩虹合成)|
| `flag:N` | 設劇情旗標 N(支援 0x 十六進位)|
| `item:N` | 道具 N 入背包 |
| `gold:N` | 金錢設為 N |
| `prog` | 顯示主線進度階段(旗標流 #1):`進度階段 k/9 = <里程碑名>`|
| `prog:N` | 設里程碑 0..N 完成(START→THIEF_KEY→…→ZOMA)。RAINBOW/DESCEND 鏡射既有 EXE 旗標,故合成/下降事件會自動推進 |
| `opos:X:Y` | overworld 定位玩家於當前層 (X,Y)(測海岸/航行用)|
| `ship` | 取得船 + 登船於玩家當前格(設 SHIP 里程碑)|
| `ship:X:Y` | 船停泊於 (X,Y)(owned,不登船);走到該格踏上即登船 |
| `use:N` | 使用消耗品 N(#3):藥草治第一個受傷隊員 / 聖水驅敵 / 蓋美拉翅膀回地表 |
| `hurt:N` | 設隊長 cur_hp=N(測藥草治療封頂用)|
| `zoma` | 大魔王索瑪終戰(怪 0x7c);勝利 → 破關 + 結局(ENDTXT)。`DQ3_BATTLE_SCRIPT` 控回合 |
| `finale` | 直接觸發破關→結局(設 ZOMA 里程碑 → 進度 9/9)。驗主線終局,不必真打贏索瑪 |
| `dlg:bank:rec` | 渲染任意對話 record(D3TXT0<bank> 第 rec 筆)。配 `DQ3_INPUT="."` 讓對話框畫進末幀 |

scripted_event 86(下降)已正式化:`DQ3_SEVENT_DESCENT`(0x56)+ `DQ3_FLAG_DESCENDED`(0x13a)+
`do_descent()`(場景效果在 main.c,因需 field/layer);原版 runner 劇情觸發待 RE,debug/U 鍵代觸發。

## 例

```bash
# 下層 overworld 截圖(驗證下降/雙層)
DQ3_DEBUG="descent" DQ3_DUMP=/tmp/under.ppm dq3_remake assets_raw game

# 進羅馬利亞 + 給錢(驗證 CTY2 載入 + 商店)
DQ3_DEBUG="gold:5000;warp:2:20:10" DQ3_DUMP=/tmp/romaly.ppm dq3_remake assets_raw game

# 設彩虹合成旗標 + 給彩虹水滴
DQ3_DEBUG="flag:0x139;item:0x75" DQ3_DUMP=/tmp/x.ppm dq3_remake assets_raw game
```

## 即戰果:抓到真 bug

加 debug 口第一次 `warp:N` 就 segfault → 定位到 `dq3_town_load` 的 **NULL-deref**:
`s->dlg_bank` / `s->section` 寫在 `s = calloc(...)` **之前**(s 還是 NULL)。此 bug 會讓**實際遊戲
每次進城都崩**,但 headless 單元測試(dialogue/door test)都**手動建 scene、不走 dq3_town_load**,
所以從未觸發。debug 口一試即現 → 已修(兩行移到 calloc 後)。

> 教訓:headless 驗證口要能打到「真實載入路徑」,不能只測手搭的替身(呼應 verification-fidelity)。

## 腳本輸入 DQ3_INPUT(驅動互動迴圈)

`DQ3_INPUT="<keys>"` 讓互動迴圈由腳本驅動(headless playthrough),每字 = 一次 poll:
`u/d/l/r`=方向、`e`=Enter、`s`=Space、`c/b/t`=C/B/T 鍵、`y`=U 鍵(下降)、`.`=閒置、`q`=結束。
字串耗盡 → 自動結束(`g_quit`);設 `DQ3_DUMP` 則 dump 末幀。`DQ3_DEBUG` 先定位再 `DQ3_INPUT` 驅動。

```bash
# 走到店員開店(debug 擺玩家在店員正上方 → Enter)
DQ3_DEBUG="warp:0:27:32" DQ3_INPUT="e" DQ3_DUMP=/tmp/shop.ppm dq3_remake assets_raw game
# → stderr: 設施:武器/防具店(CTY0 sec0 k1,7 品項)
```

> 注意:debug 定位(warp/descent/ascend)會跳過開場 CTY00 載入(`debug_placed` 旗標),否則被覆蓋。
> NPC **會擋移動**(`dq3_scene_input`:`dq3_scene_npc_at(tx,ty)>=0 → return 0`);互動測試走到 NPC 旁、面向它按 Enter。

## 主線里程碑驗證:`tools/playthrough_check.sh`

一批 debug+input 里程碑 pass/fail(全城載入/話す/開店/建隊/下降/合成事件/跨城):
```bash
tools/playthrough_check.sh <assets_dir> <dq3_remake_bin>   # PASS=N FAIL=0(項數動態成長)
```
全綠 —— 證明各系統(載入/對話/設施/隊伍/scripted event/傳送…)串接無誤。整合於 `game_tester.sh`(現 93/93)。

## 待擴充

- `party`(快速建測試隊伍,驗戰鬥/商店/裝備)、`event:N`(直接跑 scripted_event N)、`battle:monster`。
- 真實劇情觸發(如 event 86 的旗標條件)仍需 runner RE;debug 口提供「跳過劇情直接到該狀態」的捷徑。

## 亂數模式(DQ3_RNG)

`DQ3_RNG=real` → 真實亂數(xorshift32,週期 2³²,分布佳);未設 / 其他 = DOS 忠實
(鏡射 EXE 0xfa39,16-bit state 週期 ≤65536,重現原版亂數規律)。影響 NPC 走動 +
戰鬥隨機(damage/逃跑/挑目標/敵咒)。預設 DOS 以保持與原版一致。
