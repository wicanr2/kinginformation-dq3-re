# DOSBox 黃金對照:原版實機 vs 靜態解碼

第一次在 DOSBox 把精訊版 DQ3(`DRAGON FIGHTER III / 傳說的終章`)原版 `DQ3.EXE` 跑起來,
逐張截圖建立「黃金對照」,驗證本專案前面各 docs 從 `DQ3.EXE` + 素材檔靜態解出的 render
是否與實機一致。結論:**已驗證的畫面(開場年代過場、標題、主選單、注音姓名輸入)與我們的
靜態解碼逐項吻合**,palette、佈局、文字、字型皆對得上;town / 世界圖 / 戰鬥畫面卡在
「注音姓名輸入」這道必填關卡,headless 自動推進未能在時限內突破,誠實標為受阻(見文末)。

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

## 受阻:town / 世界圖 / 戰鬥未截到(誠實標記)

目標清單中的**城鎮(室內)、世界地圖、戰鬥(史萊姆)**三畫面**本次未截到**,原因:

- 上述三者都在**注音姓名輸入完成之後**才會進入;姓名輸入是必填關卡,Escape 不能跳過。
- headless 自動推進時,方向鍵 / Enter / Tab / 直接打字(`xdotool type`)都只能在注音
  鍵盤格內移動紅游標,**無法跨進右欄「完成」**把名字定案。該畫面的輸入 handler 對
  按鍵的對應與標準方向鍵不同(游標被限制在注音表內),要可靠導到「完成」需先反組譯
  `DQ3.EXE` 的姓名輸入 handler(哪個鍵切到功能列 / 確認),屬 gameplay 自動化範疇,
  超出本次「建環境 + 黃金對照」目標,於時限內有界嘗試後停手。

→ 後續要截 town / world / battle,兩條路其一:(a) 先反組譯姓名輸入 handler,得知切換到
「完成」的鍵,補完自動推進腳本;(b) 用 DOSBox 存檔狀態(savestate)或既有進度存檔
(`載入進度`)略過開場直接進遊戲。本次已驗證的標題 / 開場 / 選單 / 字型路徑,已足以
確認我們對 `FIRST.SCR`、`TIT*.P`、CHINA.FON、選單系統的解碼正確;town / world / battle
的對照(對 `docs/maps/cty00_town.png`、`world_con.png`、`docs/13` 戰鬥)留待解開姓名
輸入關卡後補做。

## 與我們 render 的對照總結

| 實機畫面 | 我們的 render | 對照結論 |
|---|---|---|
| 開場「1989」過場 | `docs/title/scene_1989.png`(TITA.P) | ✅ 黑底金字、字體、色階吻合 |
| 標題 DRAGON FIGHTER III | `docs/title/title_logo.png`(TITG.P) | ✅ logo / 副標 / © / 漸層天空 / 城堡剪影逐項吻合 |
| 主選單 勇者鬥惡龍 | `docs/03` / `docs/12`(選單系統) | ✅ 選單框 / ▶游標 / 中文字型一致 |
| 注音姓名輸入 | (新發現,尚未 RE) | ✅ 揭露中文版採注音輸入法;待反組譯 handler |
| 城鎮(室內) | `docs/maps/cty00_town.png` | ⏸ 受阻(卡姓名輸入,未截到) |
| 世界地圖 | `docs/maps/world_con.png` | ⏸ 受阻 |
| 戰鬥(史萊姆) | `docs/13` / `references/game3.png` | ⏸ 受阻 |

未發現需修正的解碼偏差:已驗證畫面的 palette(PCX 自帶 16 色)、planar 重組、中文字型
渲染均與實機一致。海面 palette cycling、CTY 城鎮貼圖等動態 / 室內畫面的對照,須待突破
姓名輸入後才能在實機檢驗。

## 產出檔案

- `tools/dosbox.conf` — DOSBox 設定(svga_s3 / nosound / autoexec 進遊戲)。
- `tools/dosbox_run.sh` — 容器內驅動:Xvfb + DOSBox + xdotool 推進 + import 截圖
  (title / 開場 / 選單序列;有時間上界)。
- `tools/dosbox_ingame.sh` — 進階序列(嘗試完成姓名 → 進遊戲;受阻於姓名輸入)。
- `work/dosbox/live_*.png` — 實機 raw 截圖(版權,**不入版控**;留本機對照用)。
