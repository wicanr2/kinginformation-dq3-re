# Boss / special 事件正式觸發點盤點(第一性原理)

> 使用者洞見(2026-06-26):boss 不是會走動的 NPC,是地圖上**固定一格(或數格)放著 sprite 的物件**,
> 玩家走過去按**空白鍵(examine)**就觸發開戰。所以「boss 正式觸發點」== 某 CTY/section/座標上
> 的 `kind=special` 事件格。純靜態可定位,不需實機探索。

## 權威來源

`docs/data/npc_dialogue.json` —— 已 RE 解碼的全 CTY NPC dump,每筆欄位:
`sec, bank, x, y, sprite(entry), sub, kind(talk/facility/special), dlg(事件 record 號), text`。
這是反組譯產物,比手刻 section parser 可靠(手刻易踩 section 偏移錯位陷阱,見記憶
`re-section-table-count-prefix-offby1`)。工具:`tools/list_special_events.py`。

- `kind=special` = examine 觸發事件(boss / 機關 / 劇情觸發點),全遊戲共 **64 個**。
- `dlg` = 事件 record 號;boss 事件如八頭大蛇 **CTY19 sec1 (35,12) dlg=45** 對上既接的 byte4=45。
- ⚠ `dlg` 是事件 id,**不保證等同** dq3_scripted 的 byte4(多數一致,個案需以對話/攻略確認)。
- ⚠ special **同時涵蓋** boss、機關(石門/按鈕)、give 道具劇情 NPC,不是全部都是 boss。

## 辨識 boss 的結構特徵(第一性原理)

1. **2×2(或多格)同 sprite 群 + 同 dlg**:boss 大 sprite 跨 4 格,4 筆 special 同 sprite/同 dlg。
2. **對話為「‥‥‥‥」(沉默)**:boss 開戰前不說話,examine 即進戰鬥。
3. sprite entry **不一定大**(八頭大蛇 b2=14);關鍵是 special + 結構,非 sprite 數值大小。

## 已確認的 boss 觸發點

| CTY | sec | 座標 | 結構 | 狀態 |
|---|---|---|---|---|
| 19 | 1 | (35,12) | 八頭大蛇 dlg=45 sprite=14 | ✅ 已接(main.c cur_cty==19 b4==45)|
| **14** | **1** | **(14,12)(14,13)(15,12)(15,13)** | **2×2 sprite=46 + dlg=58 + 對話全「‥‥‥‥」(沉默大 boss)** | **❓ 未接(強 boss 候選)** |

> CTY14 身分待定(對話僅沉默,無城名線索)。2×2 沉默大物件 = 典型 boss 擺位。
> 需以 reference 攻略 / overworld 位置 / 王座畫面 dump 確認是哪場 boss(巴拉摩斯城?怪力魔?)。

## 巴拉摩斯城 CTY 定位(未決)

- BBS 史料(docs/history/dq3-bbs-1994.md):「打贏巴拉摩斯後回阿里阿罕,索瑪現身震垮蓋亞那洞穴牆→跳下進異界」。
- npc_dialogue.json 提「巴拉摩斯」的 CTY:CTY00(阿里阿罕,破關後傳聞 ×4)、CTY16、CTY25(×4)。
  這些是**提到**巴拉摩斯的台詞,非巴拉摩斯**所在城**。
- 巴拉摩斯城本身待從 special 結構 + overworld 座標 + 王座畫面交叉定位(CTY14 為當前最強候選)。

## 全 64 special 清單

見 `tools/list_special_events.py`(無參數列全部;`CTYnn` 篩單城;`--text` 附對話)。
已接的 byte4/dlg:scripted 表 {7,12,16,25,31,44,49,50,52,84} + main.c 特例 {(15,25),(19,45),(82,64)}。
其餘為待考證(部分是已用其他機制接的劇情 NPC / 機關,非全部 boss)。
