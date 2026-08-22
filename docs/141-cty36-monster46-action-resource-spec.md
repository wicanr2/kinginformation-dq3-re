# CTY36 monster46 action／MP 與離船資源 blocker（2026-08-22）

> 歷史 blocker：bit9 MP 與 bit41 睡眠已接線，後續正式 trace 已由標題抵達 `THE END`；
> 下述失敗狀態是研究時間線，不是 current worklist，現況以 `docs/74` 最新 checkpoint 為準。

## 問題與停止線

現行 `TestOpeningProductionInputTrace` 已正常取得 CTY36 愛的回憶並由原始 transition
離船，但在航向奧莉薇亞海岬 `(76,54)` 時遇到 `monster46×4` 全滅。本切片只回答：

1. monster46／CTY36 原始設定是否抄錯；
2. monster46 的兩個 action 在原版如何分派；
3. remake 是否缺少會改變該戰鬥的原版 consumer；
4. 規則補齊後，正常玩家應如何用現有道具管理長路資源。

不研究逐幀動畫、PCM wall-clock、全怪物 action 表或全遭遇表；沒有證據時不改怪物
HP／攻防、群量、encounter gate 或 RNG。

## 輸入與可重建工具

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`
- `assets_raw/D3MNS.DAT`：5,330 bytes；SHA-256
  `48bf3ce78c425239363761c13c708a6e04f8daa628cd575f7e049bc0bc7bb4ed`
- `assets_raw/CTY36.DAT`：9,915 bytes；SHA-256
  `86291f8846dda91512f4a03585effd889d6668a982c30946d36164447b7df091`
- IDA Pro 9.4；16-bit DOS database，原始輸入唯讀。
- IDAPython：`tools/ida_dump_monster46_cty36.py`。
- sidecar：`work/monster46-cty36-ida.json`（研究輸出，不加入 Git）。

位址契約：`IDA linear = logical + 0x10000`；主 load image
`file = logical + 0x1370`。匯出器保留原始 `sub_*` 名稱、bytes、xref 與位址，不 rename
或寫回 binary。

## monster46 與 CTY36 raw 設定

monster46 記錄在 `D3MNS.DAT file 0x75e`，stride `0x29`：

```text
23 00 05 00 0a 00 00 00 28 59 00 34 00 8c
00 40 00 00 00 40 00 00 00 1b b4 d5 f4 33 a8 ff 00 00
00 5c 00 14 00 0a 53 e6 08
```

已由既有 loader／consumer 證實的欄位：HP=`35+rng(5)`、MP=10、攻40、防89、速52、
action gate=140、逃跑 threshold=27／rate=180、EXP92、gold20。兩個 section header 的
`+0x11` 都是 `0x1a`（26）；`sub_1BD97` 的 encounter consumer 對非零值繼續計數，
所以 CTY36 船內隨機遭遇是原版設定。推論等級：**confirmed**。

因此不得用降低 monster46 數值、刪除四隻群或把 CTY36 設成安全區修 trace。

## action mask 位序勘誤

raw mask 是 `00 40 00 00 00 40`。`sub_19AD6` 在每個 byte 由 `0x80` 向右測試，故
set bit 是 **9、41**，不是用 LSB-first 會得到的 14、46。初版 sidecar 曾犯此工具錯誤，
但尚未進 production；匯出器已改成 MSB-first 並由 raw bytes 重建。推論等級：
**confirmed／tooling correction**。

## bit9：一般希亞多與怪物 MP transaction

`sub_199DC` 對 bit0..15 直接令 `action=bit`，所以 bit9→action9。
`sub_19B4C`（IDA `0x19b4c`／logical `0x9b4c`／file `0xaebc`）：

1. 將 `action+0x79` 寫成文字 record，因此 action9→D3TXT00 rec130（希亞多）；
2. 以 `action*3` 查 DGROUP `0x37c3` descriptor，MP cost=3；
3. 比較 active monster MP `[bx+0x233d]`；足夠才扣除，否則顯示 MP 不足並消耗本次 action。

monster46 原始 MP=10，因此每個個體最多成功施放三次 cost3 的希亞多；本切片開始前的 Go
雖保存 enemy MP，普通 spell fallback 卻未扣 MP，形成無限施法，現已依下方規格修正。
推論等級：**confirmed**
（selector→descriptor→MP reader/writer→文字 consumer 閉合）。

## bit41：action0x35 的全隊睡眠 writer

bit41 由 DGROUP `0x3930[23]` remap 成 action `0x35`；DGROUP `0x394e` 的 handler word
指向 logical `0xa3ee`（IDA `0x1a3ee`／file `0xb75e`）。該 handler：

```text
IDA 0x1a3ee  DI=0x171 → D3TXT00 rec369「{怪物}吐出甜甜的味道。」
IDA 0x1a3f6  CL=party count；逐名呼叫 sub_19834 取得角色 record
IDA 0x1a406  test [SI+0x38],0x00a0；死亡或已睡眠則跳過
IDA 0x1a40d  RNG(256)
IDA 0x1a410  cmp AL,0xb4；roll>180 跳過
IDA 0x1a414  or byte [SI+0x38],0x20
IDA 0x1a418  DI=0x152 → D3TXT00 rec338「{角色}睡著了……」
```

這是 all-party、逐名判定的 battle sleep action，不是歷史 lookup 所寫的「因帕斯」，也
不是單體物理攻擊。本切片開始前的 Go 尚無此 pack action，現已依下方規格接線。
推論等級：**confirmed**（remap table、
handler pointer、原始 writer、RNG threshold、文字 record 與狀態 consumer 閉合）。

## Remake 規格

1. `battle.json.monster_actions` 新增 bit41／action53：
   `apply_condition`、sleep condition、`party_alive_unaffected`、`success_roll_max=180`。
2. `battle.conditions` 新增 battle-only sleep definition；不得附帶場景傷害或持久存檔。
3. battle text pack 新增 rec369 的穩定角色；成功訊息重用已存在且由 rec338 定錨的
   `enemy_sleep` role。
4. `executeMonsterAction` 依 condition ID 分派 poison（持久鏡像）或 sleep（本場
   `statusParalysis`）；未知 condition 仍 fail closed。
5. 一般 spell fallback 在執行效果前按原始 descriptor 檢查並扣每隻 enemy 的 MP；不足時
   顯示現有 `mp_insufficient` role並結束該 action，不落回物理。
6. component test 必須鎖定：bit41 全隊逐名睡眠／threshold、已睡眠或死亡跳過、monster46
   MP 10 只能支付三次 cost3 spell，以及第四次 MP 不足不造成傷害。

## 正式 trace 策略與驗收

目前失敗狀態為勇者 Lv35、`0/282 HP、16 MP`，兩名同伴皆死亡；monster46 四隻剩
`8/37/39/40 HP`。背包仍有六瓶聖水，表示不是缺少可用資源，而是從 CTY38 補給後沒有在
幽靈船／離船段使用它。

規則補齊並重播後，若第一個 blocker 仍是同一路段，正式 trace 可在取得愛的回憶後、離開
CTY36 前，透過道具面板使用玩家實際持有的聖水；不得直接寫 repel counter。驗收必須從
愛的回憶合法 save checkpoint 經原始 transition、正常登船、抵達 `(76,54)`，並繼續完成
奧莉薇亞事件與下一個 save/load。若 blocker 移到別處，記錄第一個新狀態，不為本切片調整
無關怪物。
## 正式流程重播結果（2026-08-22）

- 接上怪物 46 的 bit9 逐怪 MP 消耗與 bit41 全隊睡眠後，production trace 已自然通過
  CTY36 幽靈船、奧莉薇亞海岬、CTY55 蓋亞之劍與 CTY64 銀寶珠；不需降低怪物數值、
  移除 formation 或關閉 CTY36 遭遇。
- 新的第一個失敗點在 CTY64 出口：勇者只剩 1 MP，無法支付 Rura 的 8 MP。首次重播曾
  假設是共用 portal helper 保留最後一瓶聖水所致；在進洞前正式使用該瓶後，失敗仍完全
  相同，而且洞窟出口的 `heroMP >= MPCost(Rura)` 不變量沒有觸發。此假說已被推翻。
- 後續診斷一度在正式標題輸入前檢查剛建立的 `Game`，因而誤見初始 MP／能力；直接
  `Load()` 可完整還原 `MP=8` 與終盤能力。這是已訂正的測試診斷錯誤，不是 save
  writer／loader 缺陷。現行不變量改放在正式選擇「載入進度」之後。
- 正式載入後 MP 的確仍為 8；後續消耗發生在測試為驗證銀寶珠冪等性而再次走向 NPC
  的額外路段。CTY64 encounter gate 仍開啟，該路段抽到戰鬥並伴隨升級（最大 MP
  `44→45`），把現有 MP 花到 1。八瓶聖水此時亦已由前段正常路線耗完，不能假稱尚有
  最後一瓶可用。現刪除這段非主線、非 round-trip 必要的重走；存檔前 transaction 與
  正式載入後的 item／flag 驗證已充分證明持久化，不修改事件、RNG 或 MP。
- 證據等級：洞窟出口與存讀檔資源保存均為 **E3 runtime observation**；原版玩家可採
  其他補給順序，不能把這條 deterministic trace 策略擴張成原版唯一攻略。
- 刪除非必要的 CTY64 重複 NPC 路段後，正式 trace 已成功使用 Rura 離開並進入返航；
  新的第一個 blocker 是前往 CTY83 航段遭遇 monster125 時全滅。這是下一個資源／戰鬥
  策略切片，不會重開已閉合的怪物 46 或銀寶珠交易。
- 後續乾淨重播另走到合法 RNG 分支：尼羅肯特出口僅剩 6 MP，先於返航即被 Rura 8 MP
  不變量擋下。這表示 CTY38→幽靈船→海岬→CTY55→火山→尼羅肯特的補給跨度仍過長，
  不是 monster125 修正造成的產品退化。
- 現行路線在取得蓋亞之劍並完成 save/load 後，以正式 Rura 回已造訪 CTY38，從原始
  item／inn NPC 補聖水與住宿。首次方案再 Rura 回 CTY55，但正式目的地清單沒有 CTY55，
  已被 runtime 證據推翻。改走船回 CTY55 雖合法，卻在額外往返後又於尼羅肯特耗盡
  MP；蓋亞之劍此時已跨 save/load，沒有事件理由重進 CTY55。現行最小路線是正常出城、
  登上由 Rura 重定位至 CTY38 的船，直接航向火山使用格。所有移動、購買、住宿、MP
  消耗與船位重定位都走 production handler；不得改寫 MP、repel、RNG 或座標。
- 直航重播到火山後，尼羅肯特出口為 `MP=0`，但背包仍有 1 瓶聖水；這證明共用 portal
  helper 的「保留最後一瓶」政策才是洞窟戰鬥重新出現的直接原因。現於進洞前透過正式
  道具選單使用該瓶，使驅敵效果涵蓋洞窟。先前相同實驗仍失敗，是因讀檔後另有一段已
  刪除的非必要重複 NPC 路徑；不能把那次結果誤寫成最後一瓶對洞窟無效。
- 再次重播證明單瓶效果仍短於完整洞窟；用完後出口依然 `MP=0`。曾把「聖水不足後以
  Toheros 並保留 8 MP」加到全域 portal helper，但它改變前段無關迷宮的 RNG／資源序列，
  已撤回。現只在 `traceGaiaSwordToNirokenta` 的每個已知 section hop 前檢查並經 production
  咒文選單施放 Toheros；若 MP 不足就接受遭遇，不延長 timer 或改 encounter counter。
- MP 仍歸零的直接原因是 portal battle helper 只讓第一位角色選逃跑，後續隊員在同回合
  仍走一般治療策略；逃跑失敗數輪會耗光英雄 MP。尼羅肯特專用 policy 現也傳入 battle
  resolver：安全的一隻弱敵以普通攻擊收尾，其他敵人正式逃跑，但所有非逃跑命令都用
  普攻等待結算，不施法。一般迷宮的 battle policy 不變。
- 上述 policy 套用位置曾誤落在八頭大蛇 trace，引用不在作用域的 `preserveRura`；已還原
  該歷史路段並精確套入 `traceWalkThroughPortalWithRepelPolicy`。修正後正式 trace 越過
  尼羅肯特、monster125 與黃寶珠；最新 blocker 已移到商人城回阿里阿罕航程的
  monster4×5，不再是本文件的 monster46／尼羅肯特 MP 問題。
