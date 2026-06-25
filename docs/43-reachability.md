# 世界可達性 / 可破關結構分析

> 用已抽出的資料(轉場圖 + 地表進城表)回答「世界是否連通、離能破關還差什麼」。
> 工具 `tools/analyze_reachability.py`(資料驅動,可重生)。

## 結果

| 指標 | 值 |
|---|---|
| 現存 CTY 檔 | 89 |
| 地表進城(cty_loc map=0)| 73 |
| 下層進城(map=1)| 15 |
| 純迷宮(map=0xff,只靠轉場)| 1 |
| **從 CTY0 可達** | **73 / 89** |
| 壞掉的轉場(目標無檔)| 0 |

**不可達 16 城**:`36, 77, 78, 79, 80, 81, 82, 84, 85, 86, 87, 88, 89, 90, 92, 93`
—— 全部是**下層世界(アレフガルド/map=1)**的 CTY(含 93=合成祠堂、90 區=終盤)。

## 根因:地表 → 下層 沒有 CTY 轉場邊

轉場圖(各 section `+0xc` 表)**內部 0 壞邊**,但**地表與下層之間沒有 `dest_cty` 邊**。
DQ3 的下層世界不是靠「踩某城門 → 載某 CTY」進入,而是靠**地表 overworld 的洞**:

- docs/35 §五:出城轉場 `destSec==0xff` + `[0x5051]=1/3` → **`Y += 0x12c(300)`= 切到下層 overworld layer**;
  `destSec==0xfe` 為特殊(`[0x5051]=1`)。
- 流程:地表 overworld 走進洞 → 掉到**下層 overworld**(dq3und.map)→ 在下層走到下層城的
  `cty_loc`(map=1)進城。是 **overworld 地形機制**,不是 CTY 轉場邊,故不出現在轉場圖。

## remake 現況與 blocker

- remake 出城 handler(`main.c` `dsec==0xff`)只 `cur = field`(回**地表**),**未支援下層 layer**
  (Y+300 / 載 dq3und.map / 切下層 overworld)。
- `find_cty_at_map(px,py,map)` 已帶 `map` 參數,但只用 `map=0`(地表);下層進城未接。
- ⇒ **remake 目前進不去下層世界 → 16 城 + 終盤不可達 = 不可破關。**

## 下層 overworld 接入(已接基礎)

轉場 `dest_sec` 編碼分層(資料驗證):**`0xff`=出地表 overworld、`0xfe`=出下層 overworld**
(0xfe 全 10 筆來自下層 CTY 77/81/82/87-90/92/93,spawn=下層 overworld 座標)。remake 已接:

- `dq3_field_load_map(dir, "DQ3UND.MAP")`:下層 overworld(244×167,同 DQ3CON.MAP 格式)。
- 轉場 handler:`dsec==0xfe` → 懶載下層 field、切 `layer=1`、置玩家於 spawn;`0xff` → 地表。
- 進城用 `find_cty_at_map(px,py,layer)`(地表/下層各自的 cty_loc）。
- dev 鍵 `U`:地表↔下層切換(測試用)。
- ⚠ **剩:首次地表→下層「下降」**——0xfe 全來自下層內部,地表無 0xfe/無 dest_cty 邊到下層,
  故首次下降是**地表 overworld 的洞 tile 事件**(待 RE;目前用 `U` 鍵代)。接上後 16 城結構可達。

## 可破關路線圖(剩餘 gate)

1. ~~下層 overworld 接入~~ **已接基礎(上節)**;剩首次下降洞事件 RE。
2. **地形 / 船 gate**(本分析**未模型化**):地表 overworld 內非全可走,早期被海阻隔,需取船才能跨洲。
   完整可達性要疊「overworld tile 可走性 + 船」層。屬下一層分析(需 overworld tile map 可走性)。
3. **劇情 scripted-event gate**:取盜賊鑰匙、魔法球、胡椒換船、彩虹水滴(#2 已)等 runner 0xabb2
   事件(docs/31)。多數是進度旗標,部分子型2 NPC 已接主對話(docs/42)。
4. **終盤**:下層世界打大魔王索瑪(在不可達的 90 區一帶)。需 1 解鎖。

## 限制誠實揭露

- 地表 hub 採**樂觀假設**(所有 map=0 城互通),**未扣地形 / 船**;真實早期被海限制在起始洲。
  所以「73 可達」是**結構上限**,非實際早期可達。實際可達性待疊 overworld 可走性(gate 2)。
- 本分析只看**結構連通**(轉場邊 + 進城表),足以定位「下層完全斷開」這個首要 blocker。
