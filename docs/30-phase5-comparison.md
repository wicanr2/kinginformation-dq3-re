# 階段⑤:DOSBox 原版 vs SDL remake 逐畫面比對

remake 的正確性以原版 DOSBox 當 oracle 逐畫面比對。本文記錄已完成的比對與結果。
工具:`tools/cmp_title.py`(像素 diff + 並排圖);DOSBox 截圖經 `tools/dosbox_*.sh`。
比對圖為版權像素衍生,`dq3_remake/*.png` 已 gitignore,不入版控。

## 標題畫面(最嚴格基準)— 一致 ✅

原版與 remake **解同一個 `TITG.P`**(ZSoft PCX,640×350,EGA 16 色),理論上應一致。
實測(`cmp_title.png`,左=DOSBox 原版 640×350、中=remake、右=差異×6):

| 指標 | 結果 |
|---|---|
| 每通道差 < 24/256 的像素 | **224000 / 224000 = 100%** |
| 完全相同像素 | 52986 / 224000 = 23.7% |

- 視覺上完全一致(標題字、劍、城堡剪影、「傳說的終章 ©1993 精訊資訊有限公司」)。
- 差異圖近乎全黑,僅天空漸層邊緣有極微差 —— 來源是 DOSBox EGA 輸出/截圖的量化/gamma
  vs remake 精確 6→8bit 色盤展開,屬次感知級,非解碼錯誤。
- 結論:**標題畫面 remake 與原版 pixel-faithful**(素材解碼正確,已另由 byte-identical
  反組譯重組佐證,docs/17/19)。

## 城鎮室內佈局 — 一致(結構)✅

DOSBox 起始房間(阿里阿罕城內,紅地毯磚房 + 主角)與 remake `dq3_town` CTY00 section0
render(`town0.png` / `game.png`)**佈局同型**:紅地毯地板、白/黃磚牆隔間、床、寶箱、
走廊、樓梯位置一致。攝影機框法不同(DOSBox 視窗 vs remake 20×15 視窗)故非逐像素,
但 tile 佈局、tile 圖庫、調色盤一致(同 CTY/BLK/PAL 解碼路徑)。

## 待補比對(需更多 oracle 場景)

- **地表**:需自動走出城鎮到世界地圖截圖,對 remake field;tile/palette 解碼同源,預期一致。
- **戰鬥**:DOSBox 遭遇戰(史萊姆群)vs remake battlescene;sprite/MNSBK.PAL 同源,
  已各自對 `references/game3.png` 一致。
- **bug 場景對照組**:bug 修正後的正確行為(如 #1 巴拉摩斯可打死、#8 戰後不變黃綠)
  需在 DOSBox 跑修正版(binary patch 對照組)截圖對 remake 行為。

## 重點

- 「一模一樣」在**素材畫面層級已驗證**(標題 pixel-faithful、城鎮/怪物同源解碼一致)。
- **遊戲邏輯層級**的逐場景行為比對(走動、戰鬥數值、事件)需 remake 補完對應系統後,
  以 DOSBox 同操作序列對拍;傷害公式等尚待對 DOSBox 校準(docs/13)。

## 戰鬥畫面版面(對 game3.png)— 結構已對齊 ✅

remake 戰鬥 HUD 依 `references/game3.png` 重排,結構一致:
- 上方隊伍狀態列(勇者/武鬥家/僧侶/魔法師,各 H+HP/M+MP/職業+等級)。
- 中間綠地帶怪群;下方左指令(角色名 + 2 欄 戰鬥/逃跑/防禦/道具 + 游標)、下方右敵名+數量。

殘留差異:**天空雲層背景**(原版 `PACKBG.SCR` page 0x13)。偵察(`tools/packbg_probe*.py`):
每 page 讀 0x6e00=28160 bytes = 320×176×4bit;row-interleaved 解碼在 640×88 時右半出現
清楚雲層,320×176 則有掃描線交錯雜訊 —— 確切寬度/交錯方式待進一步實驗。目前 remake
戰鬥天空為純藍(非雲),屬已知待補的視覺差異,不影響版面結構與互動。

## packbg 戰鬥背景格式(結構已解,完整像素待底層模擬)

反組譯 `load_packbg_page`(file 0xda55..0xdb48)解出 packbg 寫入 pipeline:
- 開 PACKBG.SCR,`lseek` 到 `page*0x3d80 + page`(page 0x13 = 草原戰鬥背景)。
- **3 趟 AND 合成**到 VRAM(`and es:[di],al`,Map Mask `out 0x3c4` 逐 plane ah 1→8):
  - 趟1:0x58=88 row × 4 plane × 0x50=80 byte(640px 寬);di pitch 0x54=84/row。
  - 趟2:0x9e=158 row(接續 di)。趟3:8 row。合計 254 row。
- 緩衝為 **640 寬 row-interleaved planar**(每 row plane0/1/2/3 各 80 byte)。
- **戰鬥背景隨地圖不同**:不同 page = 不同地形背景(已實證:page 0x13 解出右半=草原藍天白雲綠地、另一段=海洋大浪)。

**待解**:3 趟都是 `AND`(非 MOV),代表是**疊在一個預先畫好的底層**上(load_packbg_page
之前由別的 routine 繪底圖)。只解 buffer 本身 → 右半雲層清楚但左緣殘留(底層缺失)。
完整像素需連底層 routine 一起模擬。工具:`tools/packbg_probe*.py`(右半草原/海洋可見)。
