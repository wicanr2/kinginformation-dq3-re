# 色盤(indexed color)機制 + overworld sprite 渲染

> 起因:有人發現**地表上主角的臉是藍色**。追根究柢是「色盤」沒搞懂 → 本檔先把 DOS 色盤
> 機制講清楚,再記錄 bug 根因與修法,以及 overworld 船 sprite 的繪製。

## 一、什麼是色盤 / 索引色(indexed color)

DOS VGA(mode 13h,320×200×256 色)的畫面記憶體,**每個 pixel 存的不是 RGB,是一個編號
(0–255)**。真正的顏色放在硬體 **DAC** 的 256 格調色盤(palette)裡,每格一個 RGB。

顯示流程:

```
畫面記憶體[pixel] = 編號 n  ──查──▶  DAC[n] = (R,G,B)  ──▶  螢幕
```

舉例:某 pixel 存 `2`,DAC[2]=(252,216,156)→ 螢幕顯示膚色。

### 為什麼當年要這樣設計

| 好處 | 說明 |
|---|---|
| **省記憶體** | 存編號 1 byte/pixel;存 RGB 要 3 byte。320×200 = 64KB vs 192KB,VGA 才塞得下 |
| **寫入快** | 1 byte/pixel,搬圖、貼 sprite 都省頻寬 |
| **調色盤動畫(關鍵把戲)** | 改 DAC **一格**的 RGB,所有用那個編號的 pixel **瞬間全變色**,完全不用重畫 |

### 調色盤動畫:DQ3 海浪怎麼動的

地表海面 fill tile(BLK index 88)的 pixel 存編號 **2 與 10**。靜態 `DQ3.PAL` 裡
idx2=(252,216,156) 沙色、idx10=(156,116,0) 褐色。遊戲執行時**把 DAC 第 2、10 格在沙↔藍
之間循環** → 整片海面免費閃動,一個 pixel 都不用重畫。這就是 indexed color 最大的價值。

## 二、SDL 了為什麼還用索引色

`dq3_runtime.c`:framebuffer `g_fb` 仍是 **8-bit 索引**(每 pixel 一個 byte),
`g_pal[256]` 是還原的 DAC,`g_pal32[256]` 是預乘 ARGB 快取。present 時才把
`g_fb[i] → g_pal32[g_fb[i]]` 轉成 RGB 材質交給 SDL。

- **不是為了省空間**(present 時還是轉成全 RGB,沒省到)。
- **是為了忠實重現 DOS 渲染**——尤其調色盤動畫(改一格全變色)。若直接存 RGB,就重現不了
  「海面 palette cycling」這類當年的視覺效果。色盤在這裡不是包袱,是被還原的 DAC 本身。

## 三、Bug:地表主角臉變藍 + 修法

### 根因(一個索引衝突)

`dq3_field.c` 為了海面藍,做了 `s->pal[2]=blue; s->pal[10]=blue`(模擬 palette 動畫)。
但 dump 主角 sprite(`DQ3MST.BLS` entry0)發現 **膚色用的正是編號 2**(124 px,見
`hero_face_compare.png` 左=靜態膚色、右=被海藍覆蓋)。

→ 海面把編號 2 改藍,**連主角的臉一起染藍**(同一組 16 色盤,編號 2 不能同時是藍又是膚)。

### 為什麼原版沒這問題:BG / sprite 分盤

256 色的 mode 13h 裡,**背景 tile 和 sprite 用不同的調色盤區段**:tile 的「2 號」對到會
動畫的 DAC 格、sprite 的「2 號」對到另一格(固定膚色),兩者井水不犯河水。remake 先前把
tile 和主角塞進**同一組 16 色**(slot 0–15)才撞色。

### 修法:sprite 專用色盤 bank(slot 16–31)

對齊原版分盤(`dq3_scene` + `dq3_scene.c`):

- `dq3_scene` 加 `spal[16]`(sprite 色盤 bank)。
- `dq3_field_load`:`spal` = **海面覆蓋之前**的靜態色盤(idx2=膚色);`pal[2/10]` 才覆蓋成藍。
  城鎮(`dq3_town_load`)無覆蓋 → `spal = pal`。
- `dq3_scene_apply_palette`:推 32 格 —— slot 0–15 = tile 色盤、slot 16–31 = sprite bank。
- `dq3_scene_render`:主角/NPC 像素寫入時 `+ DQ3_SPRITE_PAL_BASE(16)` → 落在 sprite bank。
  背景 tile 維持 0–15(海藍照動)。

結果:`hero_face_check.png` 主角臉恢復膚色,海面照樣藍;城鎮 NPC 顏色不受影響
(`town_npc_check.png`)。

## 四、overworld 船 sprite 渲染

取船後船要**看得見**。船圖在 `DQ3.BLK` tile **51–58**(4 方向各 2 動畫格:51/52 下、
53/54 左、55/56 上、57/58 右;`ship_tiles.png`)。

- `dq3_scene_draw_tile_at(scene, fb, w, h, mx, my, tile_idx)`:用與 render 相同攝影機,在地圖
  格 (mx,my) 疊一個 BLK tile(整格不透明;船 tile 自帶海背景,疊海格無縫)。
- main.c `draw_ship_overlay`(在 `dq3_scene_render` 後):**登船** → 玩家格畫船(依 facing 取
  51/53/55/57);**未登船** → 停泊格畫 51。城鎮不畫。
- 船 tile 走**背景 tile 色盤**(slot 0–15)→ 海背景正確(這是 tile 不是 sprite,不 +16)。

畫面:`ship_parked.png`(岸邊停泊船)、`ship_sailing.png`(航行中,船隨鏡頭置中)。

## 畫面紀錄(docs/img/)

| 檔 | 內容 |
|---|---|
| `hero_face_compare.png` | 主角 sprite:靜態膚色(左)vs 被海藍覆蓋(右,bug)|
| `hero_face_check.png` | 修後地表主角臉=膚色(放大)|
| `ship_tiles.png` | DQ3.BLK 船 tile 51–58(4 方向)|
| `ship_parked.png` / `ship_sailing.png` | overworld 停泊 / 航行船 |
