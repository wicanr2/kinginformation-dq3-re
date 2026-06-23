# 世界地點圖:cty_loc 表 → 地圖標記(攻略反推)

把 `load_cty` 查表(DGROUP `0x748`,docs/13/data)的「地表座標 → CTY 號」反推到世界地圖上,
配青衫攻略反查各地點(尤其祠堂)。用途之一:定位 #2 彩虹水滴合成祠堂(scripted-event 83,
docs/31),不必硬解事件 VM。

## 座標系與比例

| 世界 | 地圖檔 | tile 尺寸 | full PNG | local_y 換算 |
|---|---|---|---|---|
| 上層(地表)| `DQ3CON.MAP` | 244×205 | `world_con_full.png` 3904×2460 | `y`(直接)|
| 下層(アレフガルド)| `DQ3UND.MAP` | 244×167 | `world_und_full.png` 3904×2004 | `y − 269` |

- 像素:`px = X · (PNGw/244)`、`py = local_y · (PNGh/tiles_h)`。
- **統一垂直座標**:`[0x4f31]`(玩家 Y)上層 0..204、下層 269..435(下層佔 269..435 = 167 列;
  205..268 為空隙)。據此把 cty_loc entry 分到對應世界。
- entry 格式 `{X(byte), 0, Y(u16)}`;`load_cty` 比對 `Y==[si+2] 且 X∈{[si],[si]+1}`,
  entry 索引 = CTY 號(`CTYNN.DAT`)。idx 94–97 為表尾雜值。

## 標註成果

- `docs/maps/world_con_cty.png` — 上層 76 個入口(idx 0–76 + 83)。
- `docs/maps/world_und_cty.png` — 下層 15 個入口(idx 77–93,`y−269`)。
- 重複座標對(同點兩 CTY,典型「入口/內部」):idx 79=80 `(106,366)`、90=91 `(110,373)`。

> 標記是**所有 CTY 入口**(城鎮/城/祠堂/洞窟一起),非只有祠堂。要逐點區分「祠堂 vs 城鎮」
> 需逐點看地圖 sprite 或 CTY→名稱表;目前未做。

## 青衫攻略關鍵地點(反查依據)

| 物品/事件 | 攻略描述 | 世界 |
|---|---|---|
| **彩虹水滴合成(#2)** | **利姆達爾鎮往南航行的小島·神聖祠堂**(原版 BUG → 銀寶珠 0x6b)| 下層 |
| 太陽之石 0x72 | 拉達多姆城東邊廚房左數第 2 格 → 往南到二樓 | 下層 |
| 雲雨之杖 0x73 | 拉達多姆城往南過橋 → 往東南的祠堂 | 下層 |
| 乾渴壺 0x5e | 藍色地毯上開啟通道取得 | — |
| **最終鑰匙 0x57** | 姆歐魯村北方海域淺礁,中央下方用乾渴壺吸海水 → 浮現祠堂 | 下層 |
| 彩虹水滴用途 | 利姆達爾往西北走到盡頭 → 造橋 | 下層 |

## #2 合成祠堂定位

- 合成祠堂在**下層、利姆達爾南方小島**。下層南區標記為候選(idx 84/85/92/93、79/80・90/91 重複對)。

### A′:EXE patch 傳送 + DOSBox 截圖(已驗證)

無 debugger build 也能定位 —— **patch 新遊戲起始 CTY 即可開機直接進入任一 CTY 內裝**:

- 新遊戲 init(file 0x13a0):設座標 → `[0x4f2d]=1`(城鎮)→ `[0x256c]=0`(CTY0)→ load town。
- **patch file `0x1400`**(`mov word [0x256c],0` 的立即值)→ 候選 CTY 號 → 開機進該 CTY。
- 工具:`tools/dosbox_warp_cty.sh <idx>`(dq3-dosbox 容器內,走名字輸入 oracle → 截圖)。
- 截圖存 `docs/maps/warp_shots/`。已辨識:
  - **CTY79/80**(同座標,村莊,護城河+綠籬)= **利姆達爾**候選。
  - **CTY84** = 旅館型(床+寶箱+NPC)。**CTY92/93** = 城鎮/城堡型。
  - 合成祠堂(小型祠堂內裝)尚未在已成功載入的候選中現身。
- **限制**:DOSBox 按鍵自動化(XTEST)在負載下偶發卡名字輸入(~半數),逐顆需重試;城鎮內裝
  視覺相近,祠堂/城鎮之分需更穩的逐顆載入或「給材料 → 觸發合成」確認。

### 收尾下一步(擇一)

1. **給材料 + 觸發確認**(決定性):patch 同時注入 太陽之石+雲雨之杖到隊伍道具欄,進候選 CTY
   走到 NPC 調べる;**只有真祠堂會觸發合成**(太陽之石消失 → 銀寶珠),截道具欄前後比對。
2. **靜態 CTY-NPC 格式**:解 CTY 檔的 NPC→event-id 綁定,掃下層 CTY 找 action=83(scripted-event)。

## 再生

```
tools/overlay_cty_map.py     # PIL(docker uv venv);輸出 docs/maps/world_{con,und}_cty.png
tools/dosbox_warp_cty.sh <i> # 容器內:patch [0x256c]=i → 開機進 CTYi 內裝,截圖比對
```

## 下層(アレフガルド)城鎮對照 — 建築 sprite 驗證 + mikesrpgcenter 標籤

> 修正:下層 `local_y = raw_Y − 300`(原誤用 −269,下層整體偏 31 格)。以**地圖建築 tile
> sprite** 為 ground truth 驗證:16 個下層 cty_loc 全部精準落在建築 sprite 上(towns=tile
> 0x76/0x77、其他建築 0x08/0x0c/0x11/0x1b/0x78/0x7c/0x9c…),非隨意圈點。

疊到 NES Alefgard 標籤圖(`docs/maps/alefgard_cty_overlay.png`),逐點對上:

| CTY | local 座標 | Alefgard 標籤 | 精訊名/備註 |
|---|---|---|---|
| 78 | (64,26) | Cave NW of Brecconaly | |
| 82/81 | (142,25)/(166,34) | Tower W of Kol / Kol | 東北 |
| 87 | (89,36) | Brecconaly 北 | |
| 77 | (84,67) | Brecconaly | |
| 79/80 | (106,66) | Castle of Zoma 一帶 | town sprite |
| 91/90 | (110,73) | Castle of Zoma | |
| 88 | (91,81) | Cave SW of Brecconaly | |
| **86** | (163,96) | **Rimuldar** | **利姆達爾**(marker 落在標籤上)|
| 89 | (166,68) | Rimuldar 北/Rainbow Bridge | |
| 84 | (87,113) | Haukness | town sprite |
| 85/92 | (135,126)/(139,136) | Cantlin / Staff of Rain | 雲雨之杖一帶 |
| **93** | (170,133) | **Rainbow Drop** | **#2 合成祠堂(彩虹水滴)**;利姆達爾南方,建築 0x7c |

**結論:利姆達爾 = CTY86,#2 合成祠堂(Rainbow Drop)= CTY93**(下層 SE,利姆達爾南方),
與攻略「利姆達爾往南航行到小島神聖祠堂」一致。座標 = cty_loc CTY93 = 地表 raw(170, 0x01B1)
→ 下層 local(170,133)。

> A′(DOSBox 傳送)對 CTY80/85/93 渲染同畫面 → A′ 視覺對部分 CTY 不可靠(檔 md5 各異,
> 確為不同城);改以**地圖建築比對**為準。
