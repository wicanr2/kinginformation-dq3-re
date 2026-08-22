# 怪力魔（monster 89）戰前補給與原版證據

> 歷史 blocker：補給路線已由後續切片閉合，正式 trace 已由標題抵達 `THE END`；下方
> 「仍未閉合」只保存當時前沿，現況以 `docs/74` 最新 checkpoint 為準。

狀態：原版事件入口與 monster 89 原始記錄已由 IDA Pro 9.4 閉合；現行正式 trace
的失敗原因是戰前資源耗盡，不是 Boss 參數或特殊行動不一致。補給路線仍未閉合；
本文件記錄已排除的錯誤方案，不宣稱 monster 89 的逐 frame／音效已達 V3。

## 輸入、工具與位址契約

- `DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `D3MNS.DAT`：5,330 bytes；SHA-256
  `48bf3ce78c425239363761c13c708a6e04f8daa628cd575f7e049bc0bc7bb4ed`。
- 工具：IDA Pro 9.4；分析腳本 `tools/ida_dump_monster89.py`、
  `tools/ida_dump_cty43_facility_route.py`；受版控外的完整輸出為
  `work/monster89-ida.json`、`work/cty43-facility-route-ida.json`。
- IDA linear = logical + `0x10000`；EXE file = logical + `0x1370`。以下位址均明寫
  位址空間，不能互換。

## 原版事件與怪物記錄

IDA 原始函式 `sub_14312`（IDA linear `0x14312`；logical `0x4312`；file `0x5682`）
依序檢查 flag `0x42`、夜晚、CTY44 section1 與玩家座標 `(14,7)`，再進入原版戰鬥
consumer。這一段的入口、旗標與玩家可見事件交易為 `confirmed`。

`D3MNS.DAT` monster 89 的 41-byte 記錄位於 file `0x0e41`，raw bytes 為：

```text
40 01 50 00 0c 00 00 00 50 b4 00 3c 00 00 00 00 00 00 00 00
14 18 2f 64 ab e0 60 c0 ff 00 00 00 c4 09 69 00 00 62 c8 0b
```

已由既有欄位 consumer 證實的值包括 HP base `320`、HP random range `80`、MP `12`、
attack `80`、defense `180`、agility `60`、EXP `2500`、gold `105`。action gate 為 `0`，
48-bit action mask 全零；因此沒有證據支持替 monster 89 新增特殊 action 或降低 production
數值。以上 raw bytes 為 `confirmed`；尚未逐一閉合的尾端欄位仍不得僅憑表格位置命名。

## Remake runtime 診斷

2026-08-22 從上一個合法 checkpoint 重播 `TestOpeningProductionInputTrace`，在正式使用拉之鏡
後、進戰前觀察到：

```text
hero=87/277, MP=0
companions=[level34 HP0/232, level31 HP0/185]
```

戰敗時 monster 89 尚有 `236/321` HP。這證明目前第一個玩家 blocker 是長途路線未補給；
不能把它寫成 monster 89 原版參數未知或戰鬥 engine 缺 action。這項 runtime 結果為 E2
診斷證據，不是原版 V3 oracle。

## CTY43／44 facility 路由的 IDA／raw 訂正

- `CTY43.DAT`：4,159 bytes；SHA-256
  `9f0b1741267be10fb5b5bf7fd587dacb0ce1aa2216f78a620b54d2ee8e2086d6`。
- `CTY44.DAT`：7,040 bytes；SHA-256
  `dc7808883c4d54b9f7a79876d7cedb854538ba004dd84703134f29b32b3b306d`。
- 原版地表入口 lookup `sub_125FE`（IDA linear `0x125fe`；file `0x396e`）依 table 順序
  配對；CTY43 與 CTY44 都是 `(190,123)`，先命中 CTY43。CTY44 必須由 CTY43 section0
  transition index4 raw `2c 00 0d 1d`，即 `CTY44 sec0 @(13,29)` 進入。
- facility dispatcher 是原始 `sub_1702F`（IDA linear `0x1702f`；file `0x839f`），由
  NPC interaction subtype 與 `byte4` 索引目前 section 的 facility pointer table。

CTY43 section0 的 facility blocks raw 為 index0 旅店 `00 14 00`、index1 道具店
`02 06 41 42 44 43 45 47`、index2 教會 `03`。但 consumer 必須同時存在於 NPC 表：

- 日表只有 `(1,9)` subtype3／`byte4=0`，以及 `(31,23)` subtype3／`byte4=1`。
- 夜表只有 `(1,9)` subtype3／`byte4=0`。
- 日／夜共同的 `(12,4)` 記錄 raw `0c 04 0b 00 02 24 00` 是 subtype0／`byte4=2`，
  原版會把 `2` 當對話 record，不會進 facility dispatcher。

因此「CTY43 有可用教會」被推翻：已證實的是檔內有孤立 type3 block，但沒有任何日／夜
NPC consumer 指向它。這項結論為 `confirmed`（原始 CTY records＋IDA dispatcher）；不能再把
facility block 清單直接寫成玩家可用設施清單。

## 補給路線實驗與實作規格

曾兩次嘗試由 CTY24 回到 CTY43，依序經 CTY44 portal、CTY43 邊界離場再重進後，使用
正式 facility selector 尋找教會；兩次完整 trace 都在 CTY43 section0 回報沒有可達的
facility type3 NPC。上述 raw NPC consumer 證據進一步證明這不是尋路或 region 問題，而是
原版根本沒有指向 index2 的教會 NPC。該實驗已撤回，沒有留在 production trace，也沒有改
monster 89 數值。

同一條 production trace 先前已在 CTY2 羅馬利亞經可達的教會與旅店完成多次正式補給，
而且目前 save 的 `visitedTowns` 明確含 CTY2。最初候選 CTY22 雖曾由船路抵達並使用設施，
但重播證明它不在魯拉目的地清單；這個候選已排除。

第一次在 CTY60 後遠端繞 CTY2 補給，途中仍曾在 CTY42 後遭 monster83×4 全滅；再加兩段
正式聖水後雖抵達 Boss，進戰時仍只有勇者 `154/277 HP、1 MP`，兩名同伴皆陣亡。這證明
補給時點太早，而非補給 transaction 無效。新的最小正式策略改為：

1. 沿既有船／旅人之門／步行路線取得拉之鏡並完成 save/load。
2. 正常離開 CTY24，步行進已造訪的 CTY43，使用已由 raw NPC consumer 證實的 index0 旅店；
   旅店不復活同伴，但可恢復存活勇者的魯拉 MP。
3. 正式離城後魯拉到 CTY2 西側地表，正常步行進城，以教會復活死者，再住宿補滿全隊。
4. 正式離城並魯拉回已造訪的 CTY43；在地表用黑暗燈，正常進 CTY43、經 portal 到 CTY44
   完成假王事件。
5. 戰前記錄全隊 HP／MP；monster89 production data、encounter 與 repel counter 均維持不變。

這條策略只使用正式 `InputState` 與已由同一 trace／原始 NPC consumer 證明可達的玩家設施，
不直接注入角色狀態、金錢、位置、repel counter 或戰鬥結果。

## 第一次近端補給重播結果與連續性訂正

近端補給後的完整重播已跨過 monster89 並完成變身杖交易。進戰前勇者為 `277/277 HP、34 MP`，
但 CTY43／44 最後路段仍使兩名同伴降到 `8/232 HP` 與死亡；故 Boss 雖勝利，後續流程只住
CTY43 旅店時不會復活死亡角色。最新第一個 blocker 移到幽靈船入口前 monster125×2 全滅。

依這個實際失敗狀態，補上兩個正式玩家動作：

1. 魯拉回 CTY43 西側地表後、進城前使用一瓶玩家實際持有的聖水，保護最後一段 CTY43／44
   遭遇路程；不直接寫 repel counter。
2. monster89 勝利後先沿既有 CTY43 旅店恢復魯拉 MP，再正式魯拉 CTY2、步行進城、教會
   復活、住宿，才回到原本 CTY39／幽靈船路線。

這是 monster89 垂直鏈的「勝利後仍可往下一節點」驗收，不是為 monster125 改怪物數值。

第二次重播加入 monster89 戰後 CTY2 教會／旅店後，已依正式輸入繼續完成幽靈船與愛的回憶，
不再在 monster125 停住；最新第一個 blocker 移至後續 CTY36 路段的 monster46×4 全滅。
因此本文件的 monster89 補給／連續性切片已達 E3；這只是該必經事件的垂直鏈，不代表現行
完整 campaign E3 已恢復。下一輪從 CTY36 前最後一個合法補給 checkpoint 繼續，不再修改
monster89 或重開 CTY43 教會研究。

停止線：CTY43 的孤立教會 block 已閉合，不再追不存在的 NPC route；本切片不延伸研究
monster 89 動畫 frame、PCM wall-clock 或 Boss repeat-N，也不重開全域戰鬥 RE。
