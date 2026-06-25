# 世界可達性 / 可破關結構分析

> 用已抽出的資料回答「世界是否連通、離能破關還差什麼」。
> 工具 `tools/analyze_reachability.py`,**權威來源 = `docs/maps/map_graph.json`**
> (`extract_map_graph.py` 以**轉場 tile subid 正確界定**抽出;轉場表無 count 前綴,須以
> 實際門 tile 的 subid 取項,**不可盲讀固定 N 個 entry**)。

## ⚠ 一個踩過的雷(留為教訓)

本分析中途一度用「盲讀每 section 32 個 4-byte entry」parse 轉場表 → **溢出表尾讀到相鄰
資料(layout/tile)= 假邊**(冒出 CTY6→CTY80 之類幻邊),誤算成「89/89 全可達」。
轉場表**無 count、長度由實際門 tile subid 決定**;盲讀固定筆數必溢出。正中
[[re-disasm-offset-base-mismatch]] / rule 62 的 offset-bound 陷阱。**改回吃正確界定的
map_graph.json 後,結果與最初一致**(下層不可達)。教訓:分析器吃**正確界定**的資料,
盲讀固定筆數 = 自造 garbage。

## 結果(正確界定)

| 指標 | 值 |
|---|---|
| 現存 CTY | 89 |
| 地表進城(cty_loc map=0)| 73 |
| 下層進城(map=1)| 15 |
| 純迷宮(map=0xff)| 1 |
| **從 CTY0 經轉場可達** | **73 / 89** |

**不可達 16 城**:`36, 77-82, 84-90, 92, 93` —— 全部下層世界(アレフガルド)。
轉場圖 0 壞邊;**地表↔下層之間無轉場邊**(下層內部互通:79↔80、89→87)。

## 下層接入機制(已 RE)+ 真正的 blocker

層由**玩家 Y 座標**決定(RE 0x3fac:`Y≥0x12c(300)` 用 DQ3UND.MAP 讀 Y−300;兩圖檔柄常開)。
下層 = 同 overworld 座標空間的 **Y+300**。轉場 `dest_sec` 編碼(RE 0x488f/0x49f9):
`0xff`=出地表;**`0xfe`=設 `[0x5051]=1` → 出場 Y+=300 = 下層**。

- 已 RE + remake 已接:`dest_sec==0xfe` → 切下層 field;`find_cty_at_map(...,layer)`;dual-layer。
- **但 0xfe 門全來自下層 CTY 內部**(77/81/82/87-89/92/93),是「下層洞 → 下層 overworld」;
  **地表→下層的首次「下降」不在 +0xc 轉場圖裡**。
- ⇒ **gate 1 真正的 blocker = 首次「下降」觸發**(ギアガの大穴掉進アレフガルド)。RE 追到底:
  - 唯一把 Y+300 的是 0xfe 轉場(0x4a4f);file 0x3804 的 `[0x5051]=1` 只是**依玩家 Y 同步層旗標**
    (`if Y≥300`),非觸發。
  - 0xfe 門全在下層 CTY → 首次下降**不經 +0xc 轉場圖**,而是 **scripted 事件用 warp 執行器 0xd1f9
    直接把玩家送進アレフガルド**(子型2 NPC byte4 33-36 正是 0xd1f9 warp,docs/42)。屬
    runner 0xabb2 scripted-event 家族(靜態難追,同祠堂 #2,docs/31)→ 動態分析或逐 warp 表收尾。

## Gate 2:地表步行/船 可達性(已分析,`tools/analyze_terrain.py`)

從阿里阿罕 flood-fill 地表步行(`attr bit0=0` 可走、海 tile 88 阻擋):

| 條件 | 可達地表城 / 73 |
|---|---|
| 純步行 | **8**(阿里阿罕洲:0,1,7,9,25,28,29,30)|
| 步行 + `+0xc` 轉場洞穴 | **11**(+8,31,50)|
| **有船**(海亦可走,從起點)| **仍 8** |

**決定性發現:阿里阿罕洲的海是封閉的 —— 有船也逃不出阿里阿罕。** 這與原版一致:阿里阿罕是孤立洲,
**離開只能靠 旅の扉(travel door)**,船是後面在主大陸(波魯多加,胡椒換船)才取得。

⇒ **旅の扉 不在 `+0xc` 轉場圖**(CTY0 無轉場邊到主大陸 CTY2 羅馬利亞)→ 與「下降」同樣是
**scripted warp(執行器 0xd1f9 / runner 0xabb2)**。

## ★ 重新聚焦:scripted warp 系統 = 可破關的連接組織

把 gate 1/2 合起來看,**靜態 `+0xc` 轉場圖只管「局部門/階梯」**;真正把世界串起來的是
**scripted warp 系統(0xd1f9 + runner 0xabb2 跳表,docs/31)**,它承載:

- **旅の扉**(離開阿里阿罕孤立洲 → 主大陸)— 早期第一個 gate。
- **下降**(地表 → アレフガルド)。
- 多數**劇情 warp**(取船後的跨海點、終盤入口等)。

子型2 NPC byte4 33-36 已知用 0xd1f9 warp(docs/42)。**這套系統是到「實際可破關」最高槓桿的單一 blocker**
—— 比個別 ship/terrain 更根本。

### ★ 一部分 scripted warp 是靜態可解的(已接):type-2 examine + warp 表 0x4ea0

**更正「全靠動態」的悲觀**:scripted warp 有一條**純靜態**路徑 —— section `+8` examine 事件
**type-2** → `warp[param] = [param*3 + 0x4ea0]`(`{dest_cty, X, Y}`,113 非空項)→ 切場景
(原版餵 0xd1f9)。實證:CTY13 洞 type-2 → 羅馬利亞(CTY2)、CTY24/26/66 → CTY5、CTY82/87/88 → CTY8。

- `tools/gen_warp_data.py` 抽 0x4ea0 → `dq3_warp.{c,h}` + `dq3_warp_get(param)`。
- remake examine type-2 已接:`dq3_warp_get(param)` → 載 dest CTY 於 (X,Y)(洞穴/旅の扉出口)。
- 剩:`dest≥100` 的 overworld warp(回地表特定點)+ NPC 觸發的 0xd1f9 warp(子型2 byte4 33-36)
  + 純 runner 0xabb2 事件(阿里阿罕首個 旅の扉、下降)—— 這些仍需逐 warp/動態收尾。

## 可破關路線圖(剩餘 gate)

1. **下降事件(gate 1,進行中)**:RE file 0x3804 區的 overworld 下降(設 Y+300 / [0x5051]),
   接成地表特定點 → 下層。remake 暫用 dev 鍵 `U` 代。dual-layer + 0xfe 已就緒,只差觸發。
2. **地形 / 船 gate**(未模型化):地表 overworld 非全可走,早期被海阻隔,需取船跨洲。
3. **劇情 scripted-event gate**:盜賊鑰匙、魔法球、胡椒換船、彩虹水滴(#2 已)等(docs/31/42)。
4. **終盤**:下層打索瑪(CTY90 一帶,需 1)。

## 誠實揭露

- 「73 可達」採地表 hub 樂觀假設(地表城互通),**未扣地形/船**;實際早期被海限制在起始洲。
- 下層 16 城經轉場不可達,需 gate 1 下降事件;mechanism 已備,觸發待 RE。
