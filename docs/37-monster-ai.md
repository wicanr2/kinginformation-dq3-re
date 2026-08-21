# 精訊版 DQ3 怪物 AI(敵方行動邏輯)反組譯

精訊版 DQ3 戰鬥中,每隻怪物每回合怎麼決定要做什麼?本文把 `DQ3.EXE` 的敵方行動決策
已靜態閉合共用敵方行動 dispatcher 的主要分支:「狀態檢查 → 逃跑判定 → 物理 / 咒文」
決策樹,加上四個藏在
`D3MNS.DAT` 怪物資料裡的 AI 參數。資料驅動驗證:史萊姆只會打、荷依米史萊姆會補血、
邪惡巫師會放六種咒文、金屬史萊姆很愛逃、boss 在此 gate 不逃——資料值與已知體感相符。
這不涵蓋 boss repeat-N、逐動作 frame／PCM timing、全部咒文索引語意或 V3 對拍。

> 位址為 **file offset**(`tools/re_disasm.py` 基準,同 docs/18 codecave note「file 0xc22e」);
> 反組譯用 docker capstone(16-bit)。傷害 / 咒文效果公式見 [docs/13](13-exe-battle.md)。

## 一、每隻怪物的回合:決策樹

敵方行動決策函式入口 **file 0xbcf0**(`[0x24a5]` = 當前行動怪物的活躍 struct;
`si` = 該怪的 `D3MNS.DAT` 記錄 = `d3mns_buf(0xd78) + monster_id*0x29`)。逐步:

```
1. HP 檢查        [actor+0x2334] <= 0          → 跳過(已死)
2. 狀態旗標 [actor+0x233f]:
     bit0 set                                  → 跳過(無法行動:麻痺/睡死等)
     bit7 set(睡眠 / 混亂):
         rng(256) < 100(~39%)                → 清 bit7,印「醒來」(msg 0x163),跳過該回合
         否則                                   → 印「still asleep」(msg 0x15e),跳過
3. 逃跑判定(僅怪物側 [0xd75]&2==0):
     若 我方強度 [0x5094] >= D3MNS[+0x17]       → 進逃跑擲骰:
         rng(256) <= D3MNS[+0x18]              → 逃走!(HP=0、群數 -1、印 msg 0x15b)
4. 行動選擇(file 0xbdc7):
     dl = D3MNS[+0x0d]   ; 施咒機率
     rng(256) < dl   → 0xad4c  施放咒文 / 特技
     rng(256) >= dl  → 0xbf75  物理攻擊
```

關鍵:**第 3、4 步完全由 `D3MNS.DAT` 的四個 byte 欄位驅動**——同一套程式碼,靠資料表現出
史萊姆、法師、會逃的金屬怪、不逃的 boss 等不同行為。

## 二、D3MNS.DAT 的 AI 參數(本文新定位)

`D3MNS.DAT` = 130 筆 × 0x29 byte(docs/16)。AI 用到的欄位:

| 偏移 | 語意 | 反組譯佐證(file) | 說明 |
|---|---|---|---|
| `+0x0d` | **施咒機率**(/256) | `0xbdc7 mov dl,[si+0xd]` → rng cmp | rng < 此值 → 放咒;0=純物攻、220=幾乎每回合放咒 |
| `+0x0e`~`+0x13` | **已知咒文 bitmask**(6 byte = 48 bit) | `0xae8e si=[0x24a9]+0xe;` 掃 48 bit | 每 bit = 會某咒;放咒時從中**均勻隨機**挑一個 |
| `+0x17` | **逃跑觸發閾值**(對我方強度 `[0x5094]`) | `0xbd71 al=[si+0x17]; cmp [0x5094],al; jb skip` | 我方強度 ≥ 此值才會考慮逃;boss=99 形同不逃 |
| `+0x18` | **逃跑機率**(逃跑抗性) | `0xbd7c dl=[si+0x18]; rng cmp` | 觸發後 rng ≤ 此值就逃;0=即使該逃也不逃 |

> `+0x18`(逃跑抗性)docs/16 已標;`+0x0d` / `+0x0e..0x13` / `+0x17` 為本文新定位。

### 實測值(資料驅動驗證)

| id | 怪名 | +0d 施咒 | +17 逃閾 | +18 逃率 | 已知咒文 bit | 行為 |
|---|---|---|---|---|---|---|
| 5 | 史萊姆 | **0** | 7 | 180 | （無） | 純物理攻擊 |
| 6 | 大烏鴉 | 0 | 7 | 180 | （無） | 純物理 |
| 1 | 荷依米史萊姆 | **220** | 30 | 180 | [32] | 幾乎每回合放補血咒(治療型)|
| 8 | 人面蝶 | 140 | 9 | 180 | [31] | 常放咒 |
| 30 | 魔法香菇怪 | 80 | 27 | 180 | [9, 32] | 2 種咒文 |
| 100 | 邪惡巫師 | **200** | 47 | **0** | [1,12,19,26,29,42] | 6 種咒文、從不逃(真法師)|
| 2 | 金屬史萊姆 | 80 | **3** | 128 | [0] | 閾值極低 + 50% 逃率 → **超愛逃** |
| 3 | 金屬怪 | 50 | 3 | 180 | [3] | 同樣早逃高機率 |
| 120 | 劍魔 | 60 | 47 | 80 | [22] | 偶爾放咒、少逃 |
| 128 | 歐里狄加 | 80 | **99** | **0** | [33] | boss:幾乎不逃 |
| 129 | 五頭龍大王 | 80 | 99 | 0 | [47] | boss:不逃 |

**對照實機**:金屬史萊姆(逃閾 3 + 逃率 128)= DQ3 著名的「金屬怪一遇到就想逃」;
boss(逃閾 99 + 逃率 0)= 從不逃跑;史萊姆 / 大烏鴉(施咒 0)= 只會用爪牙打;荷依米史萊姆
(施咒 220 + 補血咒)= 邊打邊補。這些是資料與共用分支的抽樣交叉驗證，
不是全部怪物行為的實機完備證明。

## 三、施咒:從 bitmask 均勻隨機挑一個(file 0xae46)

放咒分支(0xad4c)先呼叫 **0xae46** 選一個咒文:

```
count = popcount(D3MNS[+0x0e..+0x13])    ; 數該怪會幾種咒(48-bit mask)
if count == 0: 退回物理攻擊
pick = rng(0xfa39) % count               ; 均勻隨機第 pick 個 set bit
spell_id = bit 位置(0..47),小幅修正(in [0x10,0x12) 時 +2)
[0x2511] = spell_id                      ; 設為要施放的咒
→ 套用效果(傷害 base/2+rng / 回復 / 狀態;魔甲對咒文傷害減半 #7b,file 0xc22e)
```

所以怪物**不挑「最佳」咒,而是從會的咒裡隨機選一個**(這也是 DQ 系列一貫的怪物 AI 風格:
無權重、純隨機;真正的強度差異來自「會哪些咒 + 施咒機率」)。咒文索引(bit 位置)經放咒
handler 對到 `[0x2511]`(spell_id),咒文名 rec ≈ spell_id 區(對映細節見 docs/13 咒文系統,
怪物咒文表與玩家咒文表的 index 對應待逐一坐實)。

## 四、物理攻擊(file 0xbf75)

非放咒 → 物理攻擊:印「○○のこうげき」(msg 0x14b)→ 選我方目標 → 傷害
`(atk − def)/2 + rng((atk−def)/4)`(docs/13 真公式;atk=D3MNS +0x08、def 取我方守備)。
防禦中的我方受擊減半;魔法鎧甲只減咒文傷害,不減物理。

## 五、與「行動次數 / 多段攻擊」

### 2026-08-10 靜態勘誤（本節新結論優先）

IDA Pro 9.4 以原始 `DQ3.EXE`（115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`）重播正式戰鬥鏈：

```text
sub_1BDDF → sub_1C08B → sub_1C34F → sub_1A973 / sub_1B3F3
```

`sub_1C34F` 只建立一筆 queue record／actor entry，`sub_1C08B` 對該 entry 直接呼叫
一次 actor consumer；回合結束後才重新建立下一筆 queue。這條 caller→queue→consumer
沒有發現依怪物資料再呼叫 `sub_1A973` N 次的外層 repeat loop。因此目前可證實的是
「每個活躍 actor 每回合一次」，不是「強敵每回合 1～2 次」。

> **被推翻的舊句子（保留作歷史）**：本節早期曾寫「部分強敵每回合 1～2 次，由外層
> 回合迴圈依怪物資料決定行動次數」。那是把攻略／慣例投射到未閉合欄位的假說，沒有
> caller／writer／consumer 證據，現標為 `unknown`，不得作 production 規格。

因此 D3MNS `+0x04`、`+0x1f`、`+0x25`、`+0x26` 仍不能命名為「行動次數」。若日後在
另一條正式入口找到 repeat count，必須附原始位址、輸入 hash、writer、state、consumer
與玩家可見時序；在此之前維持 `unknown`／fail closed。完整 queue／timing／boss ledger
見 [`docs/123`](123-static-battle-daynight-re.md)。

## 六、玩家側睡眠／混亂 consumer（2026-07-30 IDA 補證）

首次航海正式 trace 遭遇怪物 43 後，舊 remake 會在全隊中異常但仍有 HP 時無限空轉。
IDA Pro 9.4 重新追查後確認，先前「睡眠／混亂持續整場、不會自行解除」的摘要錯誤：

```text
敵方狀態效果 writer
  IDA 0x1ae76..0x1ae81（file 0xc9e6..0xc9f1）
  → 角色 struct +0x38 OR 0x20

角色行動前 consumer
  IDA 0x1b4f6（file 0xc866）
  → test +0x38 bit0x20
  → rng(256)
  → roll <= 0x64：清 bit0x20，顯示 D3TXT00 rec354，跳過本回合
  → roll >  0x64：顯示 D3TXT00 rec349，跳過本回合
```

怪物側則是既有 `IDA 0x1a973`（file `0xbce3`）：bit0x80 命中時
`roll < 0x64` 清除並顯示 rec355，否則顯示 rec350；兩條分支都跳過當回合。玩家與怪物
在邊界 `roll==100` 的比較運算不同，不能合併成一個 `<=` 或 `<` helper。

玩家可見副作用已由 `TestBattleStatusWakeThresholdMatchesOriginalBranches` 鎖定
99／100／101 邊界，`TestBattleActorWakeConsumesCurrentTurn` 鎖定「醒來仍消耗本回合」；
從新遊戲至加爾那之塔的正式 trace 另證明四人被怪物 43 施放異常後仍能繼續航行。

## 七、remake 對照

`dq3_battlescene` 目前的敵方 AI 是**簡化版**:每隻存活敵人 1/4 機率放「示範咒文」、否則物理
(`do_turn` 敵方反擊段)。要忠實還原原版,應改成讀復原後的 `D3MNS.DAT` AI 欄位:

- 逃跑:`我方平均等級 ≥ D3MNS[+0x17]` 時,`rng(256) ≤ D3MNS[+0x18]` 逃走。
- 施咒 vs 物攻:`rng(256) < D3MNS[+0x0d]` 放咒,否則物攻。
- 放咒:從 `D3MNS[+0x0e..+0x13]` bitmask 均勻隨機挑一咒(對映 `dq3_spelldef`)。
- 怪物睡眠／混亂：bit7 每回合 `roll<100` 醒；玩家側對應 +0x38 bit0x20，
  每次行動前 `roll<=100` 醒。兩者醒來當回合都不行動。

如此金屬怪會逃、巫師會連放咒、史萊姆只會打、boss 不逃——行為差異全來自資料,不用為每隻
怪寫程式。對應 `dq3_monster` 解 D3MNS 時可一併讀出這四個 AI 欄位 + bitmask。

## 反組譯佐證(關鍵指令,file offset)

| file | 指令 | 意義 |
|---|---|---|
| 0xbcf0 | `mov [0x24a5],bx` | 設當前行動怪物 |
| 0xbd03 | `[bx+0x2334]<=0 → ret` | 死亡跳過 |
| 0xbd2c | `test [bx+0x233f],1` | 狀態 bit0 無法行動 |
| 0xbd34/0xbd39 | `test ...,0x80; rng cmp 0x64` | 睡眠 bit7,~39% 醒 |
| 0xbd71 | `[0x5094] cmp [si+0x17]` | 逃跑觸發(我方強度 vs +0x17) |
| 0xbd7c | `dl=[si+0x18]; rng cmp` | 逃跑機率 +0x18 |
| 0xbdc7 | `dl=[si+0xd]; rng cmp → 0xad4c/0xbf75` | 施咒 vs 物攻(+0x0d) |
| 0xae46 | popcount + `rng%count` 挑 bit | 從 bitmask 隨機選咒 |
| 0xad4c | `[0x2511]=spell_id` → 套效果 | 放咒分支 |
| 0xbf75 | msg 0x14b → 物理傷害 | 物攻分支 |

> 工具:`tools/re_battle_dis.py` / capstone;AI 欄位值由 `D3MNS.DAT` 直接 dump 驗證
> (`scratch` 內 aifields.py / aimask.py)。
