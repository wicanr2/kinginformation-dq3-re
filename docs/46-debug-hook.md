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

## 待擴充

- `party`(快速建測試隊伍,驗戰鬥/商店/裝備)、`event:N`(直接跑 scripted_event N)、`battle:monster`。
- 真實劇情觸發(如 event 86 的旗標條件)仍需 runner RE;debug 口提供「跳過劇情直接到該狀態」的捷徑。
