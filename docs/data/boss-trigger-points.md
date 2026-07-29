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
| 14 | 1 | 守衛 `(14,13)(15,13)(13,14)(16,14)`；返回 boss `(14,14)(15,14)` | handler58／61 scripted NPC；handler59／60 scene tile | ✅ 原版兩階段 formation、movement、旗標與正式流程已閉合（`docs/89`） |

### CTY14 巴哈拉達救援（2026-07-30 IDA／原始資料更正）

舊文件曾把事件簡化成單格 examine、兩場 boss token 與自造完成旗標；那不是原版流程，
錯誤數值已刪除。
原版是四個有限階段：

1. handler58 四守衛問答；拒絕加入後 formation
   `01 18 01 39 04`，即 `monster57×4`。
2. 玩家踩 subid1／handler59 格後播放呼救，NPC 依
   `020000020300ff` movement 移動。
3. 玩家踩 subid2／handler60 牆上開關，兩名俘虜依
   `030300ff`、`020000050100ff` movement 會合。
4. handler61 返回 boss formation
   `04 18 01 38 01 39 01 39 01 39 01`，即
   `monster56×1 + monster57×3`；接受求饒後 SET `0x25`、CLEAR `0x14/0x34`。

守衛、approach、俘虜與 boss 可見性使用原版 story flags `0x80/0x81/0x82/0x14`。
CTY15 handler25 再以 flag `0x36` 控制黑胡椒 `0x5c` 的一次性交易。完整 caller、
formation、transition、D3TXT、影片與正式 InputState trace 見 `docs/89`。

## 巴拉摩斯城 CTY 定位(2026-07-28 已解)

- 地表 `(43,130)` 進 CTY66；CTY66 `(30,11)` 跨到 CTY65 `(10,13)`。
- CTY65 sec0 `(8,3)` 是 sprite37、sub2 handler70。
- IDA `sub_1622A` 載入 `ds:4EDF={01,27,06,79,01}`，即怪121/`0x79` 單戰。
- 勝利 CLR 原版 story flag `0x29`；remake 同步保留 `0x213` 作既有里程碑相容。
- 舊「須打兩次」源自精訊版巴拉摩斯打不死 bug，不得移植成 remake 設計。

## 全 64 special 清單

見 `tools/list_special_events.py`(無參數列全部;`CTYnn` 篩單城;`--text` 附對話)。
已接的 byte4/dlg:scripted 表 {7,12,16,25,31,44,49,50,52,84} + main.c 特例 {(14,58)甘達特巢穴,(15,25),(19,45),(82,64)}。
其餘為提示/居民/設施對話(special-events-audit 已逐一判讀,0 review 真缺口),非 boss。
