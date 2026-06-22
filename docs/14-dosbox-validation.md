# DOSBox 黃金對照:原版實機 vs 靜態解碼

第一次在 DOSBox 把精訊版 DQ3(`DRAGON FIGHTER III / 傳說的終章`)原版 `DQ3.EXE` 跑起來,
逐張截圖建立「黃金對照」,驗證本專案前面各 docs 從 `DQ3.EXE` + 素材檔靜態解出的 render
是否與實機一致。結論:**已驗證的畫面(開場年代過場、標題、主選單、注音姓名輸入、性別選擇、
開場劇情、城鎮室內)與我們的靜態解碼逐項吻合**,palette、佈局、文字、字型、城鎮 tile 皆對得上。
其中靠 `docs/15` / `re/nameinput.c` 的注音輸入法反組譯,自動化推過「注音姓名輸入」這道必填
關卡,首次截到城鎮室內並驗證 CTY 解碼。**世界地圖 / 戰鬥**兩畫面仍受阻於室內出口尋路
(盲走未命中出口 tile),誠實標記(見文末)。

實機 raw 截圖屬原版商業遊戲畫面(版權 Enix / 精訊),**留在 `work/dosbox/` 不入版控**;
本文以文字描述對照結論,需要圖時引用本專案已 commit 的 render(`docs/title/`、`docs/maps/`)。

## 環境:Docker + DOSBox + Xvfb(不污染 host)

host 無 dosbox,改在容器內跑。Image `dq3-dosbox`(Ubuntu 22.04 + dosbox 0.74-3 +
Xvfb + imagemagick + xdotool):

```dockerfile
FROM ubuntu:22.04
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -qq && \
    apt-get install -y -qq dosbox xvfb x11-utils imagemagick xdotool && \
    rm -rf /var/lib/apt/lists/*
```

執行(原版素材 `assets_raw/` 唯讀 mount 進 `/game` 當遊戲碟,截圖落到 `work/dosbox/`):

```bash
docker run --rm \
  -v $PWD/assets_raw:/game:ro \
  -v $PWD/tools:/work/tools:ro \
  -v $PWD/work/dosbox:/work/dosbox \
  dq3-dosbox bash /work/tools/dosbox_run.sh title
```

容器內流程(`tools/dosbox_run.sh` / `tools/dosbox_ingame.sh`):

1. `Xvfb :99 -screen 0 1024x768x24` 起虛擬 framebuffer,`DISPLAY=:99` 讓 DOSBox 的
   SDL surface 有 display(headless 無實體螢幕)。
2. `dosbox -conf /work/tools/dosbox.conf` 背景啟動;conf 的 `[autoexec]` 自動
   `mount c /game` → `dq3` 跑起主程式。
3. `import -window root` 從 root window 抓 framebuffer 截圖;`xdotool key --window <DOSBox>`
   送按鍵推進畫面。**每步 sleep 與整體 `timeout` 皆有上界**(單次 run ≤ 240s),
   無 sentinel 無限輪詢。

DOSBox 設定要點(`tools/dosbox.conf`,改自 `work/.jsdos/dosbox.conf`):

| 項目 | 值 | 理由 |
|---|---|---|
| `machine` | `svga_s3` | 與原始 jsdos bundle 一致;DQ3 用 VGA 320×200 |
| `mixer nosound` | `true` | headless 無音效卡,避免 ALSA 卡住(ALSA 報錯無害可忽略) |
| `output` | `surface` | 軟體 surface,Xvfb 下最穩 |
| `memsize` | `16` | 同 jsdos |
| `[autoexec]` | `mount c /game` / `c:` / `dq3` | 自動進遊戲 |

xdotool 送鍵須先 `xdotool windowfocus <win>`(headless 無 window manager,不 focus
會出現 `XGetInputFocus ... focused window of 1` 而按鍵不進視窗);加上後按鍵可靠進入。
ALSA `cannot find card '0'` / `Unknown PCM default` 為無音效卡的正常噪音,不影響畫面。

## 截到的畫面與對照結論

實機推進序列:開機 → 開場年代過場 → 標題 → 主選單 → 注音姓名輸入(卡關)。

### 1. 開場年代過場「1989」 ✅ 吻合

開機數秒,黑底中央淡入金色 serif 字「1989」。與我們的 render
`docs/title/scene_1989.png`(= `TITA.P`,PCX 解碼)**一致**:同樣黑底、置中、金色/暖
灰漸層 serif 字體。確認 `title_pcx.py` 對 `TITA.P` 的 PCX(v3.0、4-plane、自帶 16 色
palette)解碼正確,連淡入色階都對得上。

> `scene_1993.png`(= `TITD.P`,三人灰階群像)是開場稍後的一拍;本次自動推進的
> Enter 連按越過了該幀未截到,但 1989 過場 + 標題雙吻合已足以佐證開場解碼。

### 2. 標題畫面 ✅ 逐項吻合(第一手確認)

實機標題與 render `docs/title/title_logo.png`(= `TITG.P`)**逐項吻合**:

- `DRAGON FIGHTER` 立體 logo + 寶珠權杖(藍寶石)+ 環繞藍色弧線 + `III`
- 中文副標「傳說的終章」位置、字體一致
- 「ⓒ 1993 精訊資訊有限公司」藍字,位置與字色一致
- 背景:橘→白漸層天空、綠色山丘與城堡剪影、底部粉紫色地帶 — 配色完全相同

意義:`docs/title/README.md` 原本只用第三方攻略截圖(`references/game1.png`)佐證
title_logo;本次提供**第一手原版實機**確認,palette(取自 PCX header 自帶 16 色,
非 DQ3.PAL)與 row-interleaved planar 重組皆正確。trim 後實機標題內容區 = 640×340
(DOSBox 對 320×170 可見區做 2×),render 為 640×350,同尺度同內容。

### 3. 主選單 ✅ 吻合 docs/11~12 的選單系統

標題後 Enter 進主選單:框內「勇者鬥惡龍 /▶遊戲開始 / 載入進度」,兩側為紅披風持劍勇者
與藍髮持刀女戰士角色立繪(鳥山明風)。選單框、▶游標、中文字型與 `docs/03`(文字格式)、
`docs/12`(指令選單)描述的選單繪製一致;角色立繪屬 `TIT*.P` 群像系列(`docs/title`)。

### 4. 注音姓名輸入「輸入姓名」 ✅ 揭露中文化技術細節

「遊戲開始」後進入姓名輸入畫面,上方標題框「輸入姓名」與左右翻頁箭頭,中央為**注音
符號鍵盤**(ㄅㄆㄇㄈ … ㄦ + OK),右欄功能鍵「英數 / 前進 / 後退 / 取消 / 完成」,
下方「輸入注音」即時顯示已選注音;背景仍是勇者 + 女戰士立繪。

這是中文版未發售 DQ3 的一個**值得記錄的本地化技術點**:姓名輸入採**注音(Zhuyin)
拼音輸入法**,而非日版的假名表或西式字母表 —— 為台灣玩家量身打造的中文輸入子系統,
內建注音鍵盤 + 功能列。可作為「挖掘中文版相關技術」(CLAUDE.md 目標)的一筆素材;
後續可從 `DQ3.EXE` 反組譯此姓名輸入 handler(注音→字碼映射、CHINA.FON 取字)補進 docs。

## 突破姓名輸入 → 性別選擇 → 開場劇情 → 城鎮室內 ✅

接 `docs/15`(注音輸入法反組譯)+ `re/nameinput.c` 的精確按鍵語意,自動化推過姓名輸入
這道必填關卡,首次截到**城鎮室內**(town interior)。

### 推過姓名輸入的可靠按鍵序列(對 re/nameinput.c)

關鍵在於把注音鍵盤的 grid 當成 **1-D ring**(raw 0–44,Up=−9 / Down=+9 / Left=−1 /
Right=+1,皆 mod 45),cell = `col*5 + row`(`row=raw/9`、`col=raw%9`)。實測起始游標
固定在 raw 0(ㄅ),於是可純位移到達指定格:

1. **進功能列**:raw0 → `Down×3 Right×8` = raw35 = cell43(`NI_CELL_TOFN` 切功能列)→ `Space`。
2. **切英數模式**:功能列起始在 row1(英數),`Space` 即切到英數鍵盤(免組字,打一字最省事)。
3. **打一個字**:英數 grid → `Up×3 Left×8` 回 raw0(左上字元格「0」)→ `Space`,名字列出現「0」。
4. **完成**:raw0 → `Down×3 Right×8` = TOFN → `Space` 進功能列 → `Down×4`(row1→row5「完成」)
   → `Space`。名字非空 + 選「完成」→ `ni_dispatch` 放行,離開姓名輸入。

實機逐步截圖驗證每個轉場都對上 `re/nameinput.c`:選 cell43 確實進功能列、功能列 row1
切換鍵盤(英數 grid 變成 `0 5 A F K… / OK`、功能列首列改顯「注音」)、選字元格名字列出現
該字、選「完成」後畫面換成**「▶男性 / 女性」性別選擇選單**——姓名輸入關卡確認被推過。

> 注意:此自動化對**時序敏感**(headless 無 window manager,xdotool 送鍵前須 `windowfocus`,
> 且 DOSBox 取鍵與我們 sleep 的相對時序偶爾錯一拍導致某次 Space 落空)。多數 run 成功,
> 少數 run 會卡回姓名格;重跑即可。腳本 `tools/dosbox_nameinput.sh` 以固定 raw-index 位移
> 實作上述序列。

### 截到的新畫面

- **性別選擇「▶男性 / 女性」**:角色建立 modal(`re/nameinput.c` 的 `sub_0854`)在姓名後
  接的性別選單,證實已離開姓名輸入。
- **開場劇情旁白**:黑底對話框「這是 016 歲生日那天的事。」(名字「0」+ 年齡顯示在劇情中)。
- **城鎮室內(town interior)** ✅ 吻合 `docs/maps/cty00_town.png`:橘紅地磚、米黃/白磚牆、
  床、木箱、櫥櫃、NPC 與主角 sprite,搭配對話框「＊「今天是相當重要的一天呢。」實機的
  tile 字彙、配色、牆面樣式、床/家具 sprite 與我們 CTY00.DAT 解出的 render(section 0–4)
  **一致** → **CTY 城鎮 tile 解碼 + palette 驗證通過**。這是開場的主角故鄉(一棟封閉建築)。

## 仍受阻:世界地圖 / 戰鬥(誠實標記)

**世界地圖、戰鬥(史萊姆)兩畫面本次仍未截到**,原因:

- 開場主角故鄉是一棟**封閉室內建築**;走到世界地圖要踏上特定「出口 tile」(樓梯 / 門)。
- headless 盲走(各方向長距離 sweep + 清沿途對話)只能在室內各房移動,**未命中出口 tile**;
  四周藍色是「室外」但室內走不過去。城鎮無隨機遇敵,故戰鬥也未觸發(要先到地表草地)。
- 要可靠走到出口需**依截圖回授做尋路**(辨識出口 tile 座標再導航),屬逐格互動式自動化,
  時限內有界嘗試後停手,不無限掛著。

→ 後續截 world / battle 兩條路:(a) **截圖回授尋路**到出口 tile 走出城,到地表草地走動觸發
遭遇(`docs/13` 的步數遭遇計數器);(b) 用 DOSBox **savestate** 或既有進度存檔(主選單
「載入進度」)略過開場直接進地表。本次已把**靜態解碼的城鎮 tile 路徑**(CTY00.DAT →
`docs/maps/cty00_town.png`)在實機驗證通過;世界地圖(`world_con.png`)與戰鬥(`docs/13` /
DQ3MNS.SHP 史萊姆 sprite)的實機對照留待走出城後補做。

## 與我們 render 的對照總結

| 實機畫面 | 我們的 render | 對照結論 |
|---|---|---|
| 開場「1989」過場 | `docs/title/scene_1989.png`(TITA.P) | ✅ 黑底金字、字體、色階吻合 |
| 標題 DRAGON FIGHTER III | `docs/title/title_logo.png`(TITG.P) | ✅ logo / 副標 / © / 漸層天空 / 城堡剪影逐項吻合 |
| 主選單 勇者鬥惡龍 | `docs/03` / `docs/12`(選單系統) | ✅ 選單框 / ▶游標 / 中文字型一致 |
| 注音姓名輸入 | `docs/15` / `re/nameinput.c` | ✅ 揭露中文版採注音輸入法;按鍵序列實機驗證對上反組譯 |
| 性別選擇 / 開場劇情 | `docs/03`(文字)/ `docs/11`(流程) | ✅ 角色建立 modal 後接性別選單 + 旁白,流程對上 |
| 城鎮(室內) | `docs/maps/cty00_town.png`(CTY00.DAT) | ✅ tile 字彙 / 配色 / 牆面 / 家具 sprite 一致 |
| 世界地圖 | `docs/maps/world_con.png` | ⏸ 受阻(卡室內出口尋路,未截到) |
| 戰鬥(史萊姆) | `docs/13` / `references/game3.png` | ⏸ 受阻(需先到地表草地觸發遇敵) |

未發現需修正的解碼偏差:已驗證畫面的 palette(PCX 自帶 16 色 / CTY tile palette)、planar
重組、中文字型渲染、城鎮 tile 貼圖均與實機一致。世界地圖海面 palette cycling、戰鬥 sprite
等動態畫面的對照,須待走出城 / 觸發戰鬥後在實機檢驗。

## 產出檔案

- `tools/dosbox.conf` — DOSBox 設定(svga_s3 / nosound / autoexec 進遊戲)。
- `tools/dosbox_run.sh` — 容器內驅動:Xvfb + DOSBox + xdotool 推進 + import 截圖
  (title / 開場 / 選單序列;有時間上界)。
- `tools/dosbox_ingame.sh` — 進階序列(早期嘗試;在解開姓名輸入前受阻,保留作對照)。
- `tools/dosbox_nameinput.sh` — **推過注音姓名輸入**(raw-index 位移序列,對 re/nameinput.c)
  → 性別選擇 → 開場劇情 → 城鎮室內,並做室內探索 / 嘗試觸發戰鬥(有時間上界)。
- `work/dosbox/live_*.png` — 實機 raw 截圖(版權,**不入版控**;留本機對照用):
  `live_00_opening_1989` / `live_01_title` / `live_02_mainmenu` / `live_03_name_entry` /
  `live_04_gender_select` / `live_05_story_intro` / `live_06_town_interior` /
  `live_07_town_explore`。
