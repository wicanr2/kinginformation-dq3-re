# 取船劇情:波魯多加國王獻胡椒換船(真實 NPC 觸發)

> #2 船系統的「劇情接點」:把取船從 debug `ship` 換成**資料驅動的真實 NPC 觸發**。
> 本檔記錄一條一條核對出來的 RE 證據(附畫面),以及修正過程 —— **先前誤判 CTY40,後更正為 CTY37**。

## 教訓:別憑「bank 有城堡」瞎湊

第一版錯把 **CTY40**(另一座王座城堡,overworld (146,52))當波魯多加,理由只是「bank-4 候選裡
它最臨海、是城堡」。**畫面與敘述對不上**(該城國王只講洞窟提示,港口截圖是一片空海)→ 推倒重來。
正解靠**錨點文字**:D3TXT00 rec502 = 波魯多加(glyph `[659,377,357,368]`),掃到 **D3TXT04 rec17
=「歡迎到波魯多加城」**;再掃 bank-4 各 CTY 的 NPC 對話 rec,**只有 CTY37 的 NPC 引用 rec17** →
**波魯多加 = CTY37**(且 cty_loc CTY16/CTY37 同座標 (26,72) = 入口/內部成對)。

> 對齊 docs/33 未解項「CTY index ↔ 城鎮 id」:本檔以「NPC 對話 rec 反查城名」補上波魯多加一筆。

## RE 確鑿事實(逐筆核對)

| 項目 | 值 | 證據 |
|---|---|---|
| 波魯多加 CTY | **CTY37**(overworld (26,72))| NPC 引用 rec17「歡迎到波魯多加城」;王座廳畫面 `cty37.png` |
| 國王 NPC | section0 **(9,6)**,sub2 scripted-event,byte4=26 | NPC 清單(count+records×7)|
| 對話分支 | rec26「怎麼了?我在等黑胡椒。」/ rec28「胡椒太好吃了…好想睡。」| 渲染 `portoga_king_wait.png` / `portoga_getship.png` |
| 黑胡椒 item | **0x5c**(rec93,glyph `[302,303,304]`)| rec26 內「黑胡椒」glyph 比對道具名 |
| 港口海格 | overworld **(25,73)**(城南鄰可航行海格)| BLKBM attr `(attr&1)&&(attr&0x20)`(docs/48)|

★ 與 CTY40 的純對話國王不同,**波魯多加國王是 sub2 scripted-event NPC,對話本身就分「等胡椒/已收胡椒」
兩段** —— 即**這個未發售 build 其實有接取船觸發**(先前 docs/47 的「取船觸發沒接」是看錯城的誤判,已更正)。

## remake 落地(main.c)

對話波魯多加國王(CTY37 (9,6))時依胡椒持有分支:
- **已取船** → rec28(吃胡椒後話)。
- **持黑胡椒(0x5c)** → 消耗胡椒、`ship.owned=1`、停泊 (25,73) 地表、`dq3_progress_set(SHIP)`、rec28 + 取船訊息。
- **無胡椒** → rec26(等胡椒),不授船。

之後玩家出城走到 (25,73) 踏上即登船(dq3_ship_input,docs/48)跨海。

## 黑胡椒從哪來:斷鏈 + remake 補洞

同錨點法追胡椒來源:glyph `[302,303,304]`(黑胡椒)出現在 D3TXT04 rec24/26/27(波魯多加國王)
+ **D3TXT03 rec91**。渲染 rec91(`pepper_hint.png`)= **CTY15 (4,15) NPC**:「黑胡椒?下方的店裡
就有賣了。」畫面裡下方確實有兩間店。

但**逐一查證:胡椒 item 0x5c 在資料裡無任何取得來源**——
- 掃全 89 城所有店家區塊([type][count][items]):**無一店進 0x5c**(CTY15 自己道具店賣的是
  41/42/44/43/45/46 標準消耗品)。
- 寶箱/隱藏物表:`{30,1,1,1,0x41,0x5c}` 的 0x5c 是**旗標 id 不是道具**(item=0x41 藥草)。
- 0x5c 是 price-0 劇情物(ITEM.DAT cat0x08)。

⇒ **斷鏈**:國王要胡椒、NPC 提示「店裡有賣」、店也在,但**早期 build 沒把胡椒進到任何店的貨架**
(典型未發售缺口,docs/47 前提再一例)。

**remake 補洞**(為可破關):把黑胡椒補進 **CTY15 道具店**(= 提示 NPC 指的「下方的店」),
在 main.c 開店時動態 append(不改自動生成的 `dq3_shopdata.c`)。⇒ 完整可玩鏈:
**CTY15 道具店買黑胡椒 → 獻波魯多加國王(CTY37)→ 得船**。

**待 follow-up**:overworld **船 sprite 繪製**(停泊格畫出 boat)為 polish;若日後定位到原版胡椒
真正的販售/給予機制,可移除此補洞。

## 畫面紀錄(docs/img/)

| 檔 | 內容 |
|---|---|
| `cty37.png` | 波魯多加城王座廳(紅毯/垂簾/戴冠國王/金柱)|
| `portoga_king_wait.png` | 對話國王(9,6),**無胡椒**:「怎麼了?我在等黑胡椒。」|
| `portoga_getship.png` | **獻胡椒**:「胡椒太好吃了…好想睡。」= 授船瞬間 |
| `pepper_hint.png` | CTY15 提示 NPC:「黑胡椒?下方的店裡就有賣了。」(下方可見兩間店)|

## 驗證

```bash
# 無胡椒 → 等胡椒(不授船)
DQ3_DEBUG="warp:37:9:7" DQ3_INPUT="ue" dq3_remake assets_raw game   # → 波魯多加國王:等黑胡椒
# 帶胡椒 → 取船
DQ3_DEBUG="item:0x5c;warp:37:9:7" DQ3_INPUT="ue" dq3_remake assets_raw game
# → ★ 取船:獻黑胡椒給波魯多加國王 → 得船,停泊港口 (25,73)(SHIP 里程碑)
```
新增 debug 口 `dlg:bank:rec`(渲染任意對話 record,本次定位城名的關鍵工具)。
