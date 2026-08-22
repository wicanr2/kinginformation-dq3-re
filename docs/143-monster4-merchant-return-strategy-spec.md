# 商人城返航 monster4×5 戰鬥策略（2026-08-22）

> 歷史 blocker：本切片已完成，後續正式 trace 已由標題抵達 `THE END`；下述全滅狀態只
> 保存策略訂正過程，現況以 `docs/74` 最新 checkpoint 為準。

## 問題與停止線

正式 `TestOpeningProductionInputTrace` 已完成黃寶珠 transaction 與 save/load。祭壇所需珠子
之一仍在酒場名冊同伴身上，玩家必須回阿里阿罕正式重新招募；當時勇者只有 1 MP，無法施放
魯拉，改由商人城出口登船後，返航遇到 monster4×5 並全滅。

本切片只回答：monster4 是否有缺失的攻擊 action，或現行 deterministic 玩家策略把可安全
收尾的群體誤判成必須反覆逃跑。不得改 D3MNS、formation、RNG、角色 HP／MP、座標或遭遇表；
也不得用現行尚未證實精確目的地的蓋美拉翅膀近似跳過航程。

## 輸入與工具

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/D3MNS.DAT`：5,330 bytes；SHA-256
  `48bf3ce78c425239363761c13c708a6e04f8daa628cd575f7e049bc0bc7bb4ed`。
- IDA Pro 9.4＋IDAPython；腳本 `tools/ida_dump_monster4_return.py`。
- sidecar：`work/monster4-return-ida.json`（不加入 Git）。
- 位址契約：`IDA linear = logical + 0x10000`；`file = logical + 0x1370`。

原始檔唯讀；IDA database 與工作副本只在一次性容器 `/tmp`。匯出保留原始函式名、raw
bytes、xref、位址與推論等級，未 rename 或修改原始 binary。

## monster4 原始記錄與 action

monster4 位於 D3MNS file `0x00a4`，41-byte record：

```text
2d000500b80b0000122d003200c8000000002000000000630022000080ff00000016012600054bc805
```

| 欄位 | 值 | 證據等級 |
|---|---:|---|
| HP | `45 + rng(5)`，上限 50 | confirmed raw＋既有 loader consumer |
| MP | 3000 | confirmed raw＋戰鬥 MP consumer |
| attack／defense／agility | 18／45／50 | confirmed raw＋戰鬥 consumer |
| action gate | 200 | confirmed raw＋`sub_199DC` caller |
| action mask | `00 00 00 00 20 00`，MSB-first bit34 | confirmed |
| flee threshold／rate | 99／0 | confirmed raw；玩家逃跑公式另見 `docs/123` |
| EXP／gold | 278／38 | confirmed raw |

IDA 的 DGROUP `0x3930` remap 將 bit34 映成 action42。descriptor 位於 DGROUP `0x3841`
（IDA `0x28611`），raw `07 ff 0b`；handler table DGROUP `0x394e` entry（IDA `0x2874a`）
指向 logical `0xa1b4`／IDA `0x1a1b4`／file `0xb524`。該 handler 與先前 monster125 的
自療 action 共用，讀 descriptor 做 7 MP gate，將 active enemy HP 增加至自身 max cap。
精訊文字 rec163 對應「比荷瑪」。推論等級：**confirmed**（selector→remap→descriptor→
MP reader/writer→self HP writer/cap→文字 consumer）。它沒有對玩家造成新的特殊傷害。

## Remake 規格

1. 不修改 production battle、monster JSON 或 D3MNS decoder；現行 action42 自療與逐怪 MP
   路徑已足以表達原版此 action。
2. `traceCanSafelyFinishEncounter` 不再硬限制「只能一隻」。只要所有存活敵都同時滿足：
   勇者敏捷不低於敵人、勇者 attack 不低於敵 max HP 的兩倍、勇者 defense 不低於敵 attack
   的兩倍，便可用正式普通攻擊逐隻收尾。
3. 此判定只選玩家命令，不直接寫勝負或 HP；`preserveMP` 期間所有等待角色仍選普通攻擊，
   不施放回復咒文。
4. 新增 monster4×5 的 component test，並從黃寶珠 checkpoint 重播正常航程。若仍全滅，
   必須記錄實際每回合傷害與資源，再收窄公式；不得加入 monster ID 白名單。

這是 deterministic trace 的保守玩家策略，不是原版攻略 AI，也不把 component 綠燈當作
campaign E3。

## 正式重播結果

- monster125 單體與 monster4×5 群體的 component test 通過；群體判定沒有寫入勝負、HP、
  RNG 或怪物 ID 白名單。
- 完整 trace 不再停於 monster4×5，但同一段商人城返航的下一個合法 RNG 分支遇到
  monster77×2；當時英雄 `0/306 HP、2 MP、poison`，兩名同伴皆死亡，因而全滅。
- 這推翻「只要逐隻補怪物安全判定即可完成返航」的假說。真正前沿是商人建城／黃寶珠前後
  的隊伍與物品交易：應確認正常玩家能否在放下建城商人前把祭壇珠子移交，或在返航前由
  合法設施恢復隊伍；不得繼續為每個航海 encounter 增加白名單。
- monster4 action42 的 production 行為不需修改；本切片只修正了過窄的測試玩家策略。
