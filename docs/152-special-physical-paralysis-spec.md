# 怪物動作 3「致命一擊」與持久麻痺契約

> 狀態：2026-08-22。原版 `mask bit 3 → action_raw 3 → 無視守備的特殊物理傷害 →
> 存活目標 +0x38 bit0x10 麻痺 → 戰鬥命令 gate → 場景 0x28 步解除／滿月草清除`
> 的 damage／writer／consumer 已由 IDA Pro 9.4 與原始資料閉合，推論等級為
> `confirmed`；jump-table 到 handler 的間接 dispatch 邊仍是 `strong`。本切片目標為 D3／E2；
> cue9 的原始 sample duration 可重播，但 Sound Blaster wall-clock、重取樣波形與逐幀
> SHP 動畫仍為 `unknown`，不得宣稱 V3。

## 問題與舊誤解

`docs/151` 當時將 `word_272E1==3`、cue9 與 record334 留為獨立未知是正確的停止線。
本切片開始前的 Go 會把未由 pack 接管的 mask bit3 經歷史 `MonsterSpellRec(3)` 誤送到
玩家 spell record124，因而把原版特殊物理動作近似成吉拉。另一方面，當時的
`statusParalysis` 實際代表原版 `bit0x20` 睡眠／混亂，會依亂數自然甦醒；它不能再兼作
原版 `bit0x10` 持久麻痺。

本規格只修正有完整 writer／consumer 證據的 bit3，不把其餘 action mask 一併猜成咒文。

## 輸入、工具與非破壞性證據

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/D3MNS.DAT`：5,330 bytes；130 筆 × 41 bytes；SHA-256
  `48bf3ce78c425239363761c13c708a6e04f8daa628cd575f7e049bc0bc7bb4ed`。
- `assets_raw/D3TXT00.TXT`：18,680 bytes；SHA-256
  `38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b`。
- `assets_raw/FVOC.VCX`：72,531 bytes；SHA-256
  `47be27d2244e314ca6f2a21d3fa61a95d4e823f9085ed3d21808ac25d66b39c5`。
- Docker 內 IDA Pro 9.4.0.260610／IDAPython；匯出器
  `tools/ida_dump_special_physical.py`，gitignored sidecar
  `work/special-physical-20260822.json`。
- sidecar 保留原始函式名、IDA linear、bytes、CFG、xref、相對 `+0x38` operand 掃描、
  action descriptor 與 VOC descriptor；沒有 rename、comment 或原地修改 EXE／database。
- 位址換算：`logical = IDA linear - 0x10000`；主段
  `file = logical + 0x1370`。

## 原版 action selector 與使用者

`sub_19AD6` 以 MSB-first 掃描怪物 record `+0x0e..+0x13` 的 48-bit action mask，選中
set bit 的 raw index 放入 `DH`。`sub_199DC` 在 IDA `0x199fb..0x19a0c` 對 `<0x12`
的 direct index 原值保存到 `word_272E1`；因此第一個 mask byte 的 `0x10` 是
`mask_bit=3`／`action_raw=3`，不是玩家 critical flag。

原始 `D3MNS.DAT` 只有三筆設此 bit，而且三者都沒有其他 action bit：

| monster id | 原版名稱 | `+0x0d` gate | `+0x0e..+0x13` | `+0x08` attack |
|---:|---|---:|---|---:|
| 3 | 金屬怪 | 50 | `10 00 00 00 00 00` | 150 |
| 48 | 獵人蜂 | 255 | `10 00 00 00 00 00` | 38 |
| 51 | 牛羊怪 | 80 | `10 00 00 00 00 00` | 44 |

`word_27287 & 0x10` 位於共同傷害 IDA `0x1ad5d..0x1ad69`，只會將一般傷害乘二；
那是另一條玩家 critical 分支。不能用它命名或實作本 action。

## 傷害、文字與音效順序

`sub_1ACCE` 於 IDA `0x1acf0` 比較 `word_272E1,3`。命中時把防禦值 `DX` 清零並跳到
`0x1aebe`：

```text
BX = attacker attack
BX >>= 1
AX = BX >> 1
DX = rng(AX)
damage = BX + DX
若受擊角色正在防禦，damage >>= 1
HP = max(0, HP - damage)
```

因此未防禦時為 `attack/2 + rng(attack/4)`（各次位移均為整數截斷），完全忽略受擊者
守備。這條路不執行一般物理 cue1，也不套 `word_27287` critical 倍率。

玩家受擊的玩家可見順序為：

```text
cue4 + record331 敵方攻擊 → sub_208E2 等待
cue9 + record334「發出致命的一擊。」→ sub_208E2 等待
record333 玩家受到 damage
若 HP > 0：record201「{角色}被痺了！」並 OR +0x38 bit0x10
若 HP == 0：進 record357 死亡分支，不寫麻痺
```

FVOC cue9：file offset `0x4cd2`、type1、block length 2,978、rate divisor107、
source rate 6,711 Hz、2,976 samples、source duration 443,451,050 ns，60 TPS 向上取整為
27 frames。這只證明原始 descriptor／sample duration，不替聲音命名，也不宣稱 host
wall-clock 相同。

## 麻痺的 battle、field、item 與 save consumer

### 戰鬥命令 gate

`sub_1BF7F`（IDA `0x1bf94`）與 `sub_1C1D8`（IDA `0x1c1e9`）都直接測角色
`[si+0x38] & 0x00b0`；`0xb0 = death 0x80 | sleep 0x20 | paralysis 0x10`。命中時不為該角色
建立／顯示正常命令。因此麻痺者不行動；它不走 `sub_1B4F6` 的 `bit0x20` 亂數甦醒鏈。

`sub_1CB3C` 另於 `0x1cb4e` 單獨測 `bit0x10`，若繞過正常命令 gate 嘗試施法則顯示
record489「{角色}還在痲痺，不能使用咒文。」。這是防禦性 consumer，不代表麻痺只禁止
咒文。

IDA 進一步閉合戰鬥內的解除 consumer：`0x1ccb2` 以 `DX=0x10`、`BP=0x18f`
（D3TXT record399「{角色}的痲痺恢復了。」）呼叫同一個 `sub_1469F`。它與相鄰
`0x1cca8` 的 `DX=0x40`／record364 解毒 handler 同屬戰鬥咒文效果 dispatch；結合既有
spell record166–168 分類，`CureStatus` 必須清除持久麻痺 bit0x10，而非只清戰鬥暫態的
sleep bit0x20。raw 指令與 xref 限制已追加於同一份 IDA sidecar；handler 的直接 caller
經 jump table 進入，IDA 沒有直接 code xref，故該 dispatch 邊仍標為 `strong`，但
`DX=0x10 → sub_1469F → record399` consumer 本身為 `confirmed`。

### 場景倒數解除

`sub_1C954` 保留角色 `bit0x10`，若任一角色帶此位元則設定 global `word_29D16`
bit`0x0200` 與 `byte_2A0C8=0x28`。正式場景步進 `sub_193E3 → sub_19530` 每次遞減該 byte；
歸零呼叫 `sub_19810`，逐名清除 `+0x38 bit0x10`，再清 global bit`0x0200`。
同一函式稍後呼叫已閉合的每步中毒 consumer `sub_1964E`，故 `0x28` 的單位是合法場景
移動步，不是戰鬥回合或畫面 frame。原版使用一個全隊共用倒數，歸零時清除全部成員麻痺。

### 滿月草

item dispatcher table 的 raw `0x45` entry（DGROUP `0x3672`／file `0x197b2`）指向
logical `0x3e15`。該 handler 先呼叫 `sub_14CF9` 消耗選中 inventory word，再以
`DX=0x10`、`BP=0x16c` 呼叫 `sub_1469F`：存活且帶 bit0x10 的指定角色會清除該 bit；
死亡或未麻痺仍已消耗，走 record341「但是什麼也沒有發生。」。成功沿用原版
record364「{角色}體內的毒解了。」；文字看似不專屬麻痺，但 raw handler／record 已確認，
remake 不自行改寫成較合理的新句子。

### save/load

原版角色 record 的 bit0x10 是跨場景狀態。Go 使用自有 `conditionSet`，不得冒充原版 bit
layout，但必須保存 paralysis condition 與全隊剩餘倒數；save/load 後不得把麻痺誤變睡眠、
毒或直接消失。

## Remake spec

- game-pack `conditions` 新增 `common:condition.paralysis`：persistent、legacy mask`0x10`、
  shared field expiry 40 steps、無 field damage。
- `monster_actions` 新增 mask bit3／action raw3，kind 限定為
  `special_physical_condition`、scope `party_one_alive`、傷害 primitive
  `ignore_defense_half_plus_random_quarter`、只有目標傷後仍存活才套 paralysis。
- `sound_cues` 新增 cue9 role；`battle_texts`／`texts.json` 新增 record334、201 role。
- 引擎依 action 的 target scope 決定是否先抽單體目標；不能因 pack 中同時存在 all-party
  bit38／41 就跳過 bit3 的單體目標 RNG。
- 戰鬥開始把 persistent paralysis 映成獨立 battle bit；命令收集與 action runner 都跳過
  麻痺角色，不套睡眠甦醒 RNG。戰鬥結束再同步回 persistent condition。
- 戰鬥 `CureStatus` 同時清除暫態 sleep／confusion 與 persistent paralysis；正式 trace
  遇到會持久麻痺的敵人時，未麻痺且學會該咒文的角色可先解除隊友，再依風險選擇逃跑。
- 正常移動每步遞減 shared paralysis countdown；歸零一次清除全隊 paralysis。新 battle
  產生麻痺時將倒數設為 pack 的40。倒數納入 Go save round-trip。
- `item_use_effects` 新增 raw item0x45 的 `clear_condition` transaction，沿用選人 modal 與
  「先消耗、後判定」primitive，但 condition 必須由 pack lookup，不再硬限制 poison。
- schema/reference validation 對上述 kind、scope、formula、condition、文字、cue、40-step
  契約失敗即關閉；Go 不新增 DQ3 raw item、record、mask bit、cue 或步數 fallback。

## 驗收與停止線

- IDA sidecar 鎖定 selector、damage block、record/cue call-sites、`0xb0` command gate、
  `0x0200/0x28 → sub_19810` field chain、滿月草 handler 與所有輸入 hash。
- parity tests 直接鎖 DQ3.EXE action descriptor、三筆 D3MNS mask、records201／334／341／364／
  489、FVOC cue9 descriptor 及關鍵 call-site bytes。
- component tests 鎖 target RNG、忽略守備傷害邊界、防禦減半、死亡不麻痺、訊息／cue順序、
  麻痺跳過回合、戰後同步、40步全隊解除、滿月草成功／無效先消耗與 save/load。
- 正式 `InputState` campaign trace 不得退步；若原路徑未隨機遇到三隻使用者，另以 production
  battle entry／正式 Battle action path 做 deterministic component trace，不以直接改 HP 後
  冒稱主線 oracle。
- 停止線：玩家 critical、其他 mask bit、cue9 可聽名稱、SHP frame、硬體播放時序另立規格。
