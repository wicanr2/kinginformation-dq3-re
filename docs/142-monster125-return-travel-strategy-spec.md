# 銀寶珠返航 monster125 戰鬥策略（2026-08-22）

> 歷史 blocker：本切片已完成，後續正式 trace 已由標題抵達 `THE END`；下述全滅狀態只
> 保存策略訂正過程，現況以 `docs/74` 最新 checkpoint 為準。

## 問題與範圍

正式 `TestOpeningProductionInputTrace` 已從 CTY64 使用 Rura 返回 CTY38 並登船，在前往
CTY83 的正常航路遇到單隻 monster125 後全滅。現行 helper 對 `monster_id >= 80` 一律
反覆選逃跑；本切片判定 monster125 是否真是高危敵人，以及應修 production 規則、原始
資料，還是只修 deterministic 玩家策略。

停止線：不改 D3MNS 數值、encounter candidate、RNG、逃跑公式或角色狀態；不擴張為
全怪物 action 語意表，也不研究動畫／PCM timing。

## 輸入、工具與位址契約

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/D3MNS.DAT`：5,330 bytes；SHA-256
  `48bf3ce78c425239363761c13c708a6e04f8daa628cd575f7e049bc0bc7bb4ed`。
- IDA Pro 9.4＋IDAPython；腳本 `tools/ida_dump_monster125_return.py`。
- sidecar：`work/monster125-return-ida.json`（不加入 Git）。
- 位址：`IDA linear = logical + 0x10000`；EXE `file = logical + 0x1370`。

匯出保留原始函式名、bytes、xref、位址與推論等級；沒有 rename 或修改原始 EXE／DAT。

## D3MNS record

monster125 位於 `D3MNS.DAT file 0x1405`，41-byte raw record：

```text
2d000900180000001e33003200c808000000400000000025645bf47f8aff000000310134000505e60d
```

已由既有 loader／IDA consumer 定錨的欄位：

| 欄位 | 值 | 等級 |
|---|---:|---|
| HP | `45 + rng(9)` | confirmed |
| MP | 24 | confirmed |
| attack／defense／agility | 30／51／50 | confirmed |
| action gate | 200 | confirmed |
| action mask | `08 00 00 00 40 00`，MSB-first bit4、bit33 | confirmed |
| flee threshold／rate | 37／100 | confirmed raw；完整雙方逃跑公式沿用 `docs/123` |
| EXP／gold | 305／52 | confirmed |

舊 `docs/38` 把 HP 顯示成 54，是把 base 與 random range 相加後的上界，不是每隻固定
HP；後續表格必須明寫範圍，不能再把上界冒充 base。

## action4 與 action41

`sub_19AD6` 由每 byte 的 `0x80` 向右掃 mask；`sub_199DC` 對 bit0..15 直接使用 bit，
故 bit4→action4。DGROUP `0x37c3 + 4*3 = 0x37cf` descriptor raw `06 23 1d`；
`sub_19B4C` 以第一 byte 檢查／扣除 active monster MP，因此為 6 MP 的一般 descriptor
路徑。文字 record 是 `0x79+4=125`「比吉拉瑪」。推論等級：confirmed。

bit33 經 DGROUP `0x3930[15]` remap 為 action41；descriptor 位於 DGROUP `0x383e`，
raw `05 55 0b`。handler table DGROUP `0x394e` 指向：

```text
logical 0xa1b4 / IDA linear 0x1a1b4 / file 0xb524
```

handler 先取得 descriptor、經 `sub_19BC3` 做 MP gate，再以 base／RNG 增加 active enemy
`[di+7]` HP，並由 `[di+9]` 上限封頂，最後使用 record `0x151` 的恢復訊息。這是 5 MP
的自身回復 action；現行 spell descriptor rec162「比荷依米」的 MP／base 與 raw
`05 55` 一致。推論等級：confirmed（selector→descriptor→MP gate→HP writer／cap→文字）。

## Remake 判定與實作規格

monster125 是低 HP、低攻防的一般怪，不是 Boss，也沒有新的未接特殊 action。現行 runtime
已能用普通 descriptor 處理 bit4 傷害與 bit33 自療，且前一切片已補上逐怪 MP 扣除；沒有
證據支持改 production JSON 或戰鬥公式。

真正缺陷位於測試玩家策略：`traceAdventureTravelToCty` 把 `monster_id >= 80` 當成高危
代理，反覆逃跑；怪物 ID 不是戰力排序。修正規格：

1. 保留既有「明確要求逃跑」與長途 MP 保存模式。
2. 若敵群數量很小、每隻最大 HP 不高於勇者目前 attack，且敵 attack 的兩倍不高於勇者
   defense，視為可用一輪普通攻擊安全收尾；即使 ID 高也不反覆逃跑。
3. 可安全收尾時仍使用 production 戰鬥命令，只選普通攻擊以保存 MP；不直接判勝、不寫 HP。
4. 完整 trace 必須從銀寶珠 save/load 經 Rura、登船、實際 encounter，勝利後繼續進 CTY83。

這是 remake 的 E3 玩家策略，不是「原版 AI 建議」或唯一攻略。若同一判定誤傷其他路線，
應收窄可安全擊倒條件，不得退回 monster ID 白名單或改怪物數值。

## 2026-08-22 正式重播結果

- `traceCanSafelyFinishEncounter` 的針對性測試通過；正式 trace 已越過返航途中的 monster125，
  證實「高 ID 一律逃跑」才是測試玩家策略錯誤，沒有修改原始怪物數值或 production 規則。
- 後續黃寶珠 transaction 與 save/load 已完成。因祭壇珠子之一仍在酒場名冊同伴身上，trace
  必須回阿里阿罕正式重新招募；當下英雄 MP 只有 1，不能施放魯拉，因此改走商人城出口、
  正式登船與一般航行，不注入 MP／座標／造訪表。
- 新的第一個 blocker 是該返航遇到 monster4×5 時全滅。這是下一個長途資源／隊伍恢復策略
  切片；monster125 的資料與兩個 action 已閉合，不應再重開或以 monster ID 白名單處理。
- 完整 campaign 仍未回綠；本結果是正式玩家路徑 runtime observation，不是原版唯一攻略。
