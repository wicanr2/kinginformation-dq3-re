# 怪物 action 前置目標 RNG 次序（2026-08-22）

## 結論與範圍

本切片只回答正常怪物行動分支中，cast gate、存活隊員目標、raw action bit 與 handler
的 RNG 次序。原版正常分支會先判定 cast gate；只有 gate 成功才抽一名存活隊員，之後才選 action bit；
即使最後選到 bit41 全隊睡眠，前置目標 RNG 仍已消耗。raw 呼叫次序為 `confirmed`；
`sub_1AB83` 的「存活單體隊員 selector」語意因迴圈與 party record consumer 為 `strong`。

這不閉合 alternate battle context、bit9 下游全部 target semantics、其餘 48-bit action
語意或玩家可見 V3。未閉合部分不得由正常分支外推。

## 輸入與非破壞性工具

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/D3MNS.DAT`：5,330 bytes；SHA-256
  `48bf3ce78c425239363761c13c708a6e04f8daa628cd575f7e049bc0bc7bb4ed`。
- IDA Pro 9.4；16-bit DOS database；`IDA linear = logical + 0x10000`，
  `file = logical + 0x1370`。
- 可重建 IDAPython：`tools/ida_dump_monster46_action_rng.py`。
- gitignored sidecar：`work/monster46-action-rng-20260822.json`。

匯出器保留原始 `sub_*`、位址、bytes、xref 與函式邊界；未 rename、未修改原始 EXE。

## 正常分支的已證實順序

`word_25B45 & 2 == 0` 時：

```text
IDA 0x1aa6c  sub_1E6B9：RNG256 cast gate
IDA 0x1aa6f  cmp AL,[SI+0x0d]
IDA 0x1aa71  jnb physical
IDA 0x1aa73  call sub_199DC
  IDA 0x199ea  call sub_1AB83
    IDA 0x1ab95  call sub_1E6C9：以 party count 為 bound，死亡者重抽
  IDA 0x199f3  call sub_19AD6
    IDA 0x19b15  call sub_1E6C9：以 set-bit count 選 raw action
  remap → indirect handler
```

monster46 `+0x0d=140`，只有 `AL < 140` 進 action selector。其 raw mask
`00 40 00 00 00 40` 以 MSB-first 解為 bit9、bit41。bit41 經
`DGROUP 0x3930[23]` remap 為 action `0x35`，handler 是 logical `0xa3ee`／
IDA `0x1a3ee`／file `0xb75e`。以上為 `confirmed`。

bit41 handler `0x1a3ee..0x1a432` 自身沒有 bounded single-target RNG；它逐名呼叫
`sub_1E6B9` 取得 RNG256，`AL <= 0xb4` 時寫角色 `+0x38 bit0x20`。因此「全隊 handler
不需要目標」不能反推前置 selector 沒有抽目標。以上 raw data flow 為 `confirmed`。

## Alternate branch 與停止線

`word_25B45 & 2 != 0` 會先進 `sub_19207`，其 bounded RNG 在 IDA `0x19230`；目前只
能確認 active record／queue selection，不能把它命名成相同的 party target selector，
等級維持 `strong／unknown`。bit9 進 `sub_19B4C` 已確認，但 `sub_1D881` 的間接效果
consumer 尚未完整閉合，不能宣稱 bit9 的全部 target semantics。

## Remake 規格

1. 正常 `enemyAction` 先執行 cast gate。只有 gate 成功且 action mask 非空時，才依
   現行相容 actor list 消耗一次 target RNG，之後選 raw action bit；不得因所有 set bit 恰為
   `party_alive_unaffected` 而跳過。
2. cast gate 失敗時直接走物理分支，不得在 gate 前預抽 target。all-party handler
   可忽略 gate 成功後已抽出的 target 值，但不得回退 RNG state。
3. component test 使用只有 bit38 的原始 action mask，鎖定「cast gate → target →
   action selector → 每名 condition roll」的 state；monster46 混合 mask 另鎖定前置
   target RNG 仍存在。
4. 不在此切片改動歷史 `MonsterSpellRec` fallback；未閉合普通／特殊 action 的
   fail-closed 遷移另開 spec。

## 驗收與證據等級

- IDA sidecar 必須非空並含兩個輸入 hash、位址基準、`sub_199DC`／`sub_1AB83`／
  `sub_19AD6` xref 與 raw ranges。
- targeted selector tests、完整 `game` 分片、`internal/...` 與正式 production trace
  都需通過；若 deterministic trace 因正確 RNG 次序改變而失敗，先記錄第一個玩家狀態，
  不得修改怪物資料或用狀態注入掩蓋。
- 正常分支 call order：`confirmed`；`sub_1AB83` 單體存活隊員語意：`strong`；
  alternate branch 與 bit9 下游完整語意：`unknown`。

## 實作與目前回歸結果

Go 已移除「若所有 set bit 都是 all-party action 就跳過 target RNG」的近似。
2026-08-22 後續勘誤發現，首次接線又把 `sub_1A973` 的 gate 與 `sub_199DC` 內 target
順序寫反：該版在 gate 前預抽 target，使物理分支的 RNG state 漂移。IDA sidecar 保留的
`0x1aa6c`／`0x1aa71`／`0x1aa73` 與 `0x199ea`／`0x199f3` 足以推翻該實作；現行訂正為
`cast gate → target → action bit`，並以 bit38 原始 mask component test 鎖定完整 RNG state。

歷史：首次「gate 前預抽 target」錯序實作的完整 trace 在 CTY38→CTY70 的航路
遇 monster77×2 全滅。曾以 CTY83／CTY38 正式商店、旅店、教會與 8／16 瓶聖水重播，仍
在同一航段耗盡資源；這些未收斂的 trace 變動已撤回。因此本切片 production 規則保留，
當時 campaign E3 改列 pending。後續只重設該玩家航路／補給切片，不回退原版 RNG，
也不修改怪物、encounter、HP、座標或旗標。

上段的 monster77 與補給實驗只屬錯序版本，不再是 current blocker。訂正為
`gate → target → bit` 後，新的第一個失敗提前至羅馬利亞附近 Lv15
練級戰的 monster15×3；原因是 trace 仍以 raw monster ID 當成強度門檻。該玩家
策略勘誤與 current 驗收見 [`docs/154`](154-cast-gate-rng-campaign-recovery-spec.md)。

2026-08-22 最終現況訂正：monster15、monster77 與後續 monster51 都是本日歷史
checkpoint。`docs/154..178` 已依相同原版 RNG 順序完成正式玩家路線與交易修正；最新
完整 trace 以 115.629 秒抵達 `THE END`。本文件證實的 call order 不變，現行 campaign
狀態由 `docs/74` 仲裁。
