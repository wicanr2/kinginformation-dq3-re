# 65 — ebitan 流程對齊原版:戰役計畫 + 進度(2026-07-03)

> 使用者判定:remake「程式都完成了」但**整個遊戲流程跟原版完全不一樣,沒有真的達到重製**。
> 本檔是對齊戰役的計畫與唯一進度真相(完成即更新,不留 stale marker;規則 63)。

## 根因(code 核實,非臆測)

- **流程層的 RE 其實早已攻克在 C remake**:`[0x722]` runner state machine(57 setter、`call [bx+0x3baa]`
  派發、region 命中,見 `docs/re-log-722-state-machine.md`);主線 8 里程碑「真實 NPC 觸發」;
  杜勝利 56 步攻略逐項 code 核實(`docs/data/walkthrough-flow-audit.md`,11 缺口→8 已解)。
- **ebitan 版(Go/Ebiten port)沒有移植流程層**,用簡化品替代:預建示範隊伍(≠露依達酒場創角)、
  地表中心出生(≠原版阿里阿罕起點)、「進城即設里程碑」(≠真實 NPC/事件觸發)、Z/U debug 鍵
  串終盤(≠自然劇情觸發)、無 runner region 事件、無遭遇區、無晝夜。
- 因此「玩起來完全不是原版」——差距不是子系統數值,是**流程驅動整層缺席**。

## 驗收定義

**一個玩家從新遊戲開始、不按任何 debug 鍵,能按原版流程(oracle)走到破關。**
oracle 優先序(2026-07-03 使用者校正後):**(0) DOSBox 原版實測(最高)** >
(1) 原版實機錄影(`dq3_real_video/`) > (2) 杜勝利/青衫攻略 > (3) RE docs/宣稱。
最終仲裁 = 使用者實測(玩過原版破關)。

> ⚠ **RE 宣稱不可信任為完備**(使用者判定,已被證實):flow-audit 只審了任務/道具/boss 層,
> 開場層、按鍵行為、NPC 密度等「玩起來的感覺」從未被 audit。每個機制修復前,
> 先用 DOSBox 原版實測該機制(工具鏈現成:`dq3-dosbox` image + `tools/dosbox_*.sh`,docs/36/14)。

## 使用者實測回報(2026-07-03,玩桌面 C remake)— 根因已定位

| # | 症狀 | 根因(code 佐證) | 影響版本 |
|---|---|---|---|
| U1 | 開場不是房間、母親沒帶去見國王 | C `main.c:1353`「互動開場:從阿里阿罕城鎮(CTY00 sec0)起步」;grep 母親/16歲/謁見 **0 命中** —— 開場劇情(docs/36 §1-6/7:旁白+家室內)**C 也沒做**;ebitan 更糟(地表中心) | **兩版** |
| U2 | 城內任何地方按 Space 被傳出城 | C `main.c:1982` demo 殘留鍵。**已仲裁(DOSBox 實測)**:原版 Space=開命令窗(2×3 六指令)+ 左下 HP/MP 小窗 → C 修法=Space 改開命令窗;ebitan 命令窗語意正確,補 HP/MP 小窗、出城方式仍待第二輪實測 | C(ebitan 小補) |
| U3 | 城鎮 NPC 沒原版多 | **RE 定案(docs/68)**:原版掃 NPC 無 `b2<4` 判斷=**顯示全部**;兩版 remake 都誤加 `b2<4 continue` 丟棄(全 CTY 共 **1661 個** b2<4 村民);C cache 上限又誤設 8(原版 13)。修法=移除丟棄+b2<4 載 DQ3LIN.BLS(C-3) | **兩版**(移植 bug) |

## Ground truth 資產

| 資產 | 位置 | 用途 |
|---|---|---|
| 原版實機錄影(1080p,1.1GB) | `dq3_real_video/*.mp4` | 流程 oracle(開場/謁見/事件順序);影格抽到 `dq3_real_video/frames/`(gitignored,版權不入庫) |
| 杜勝利 56 步攻略 | `references/walkthroughs/dusheng_dq3.txt` | 主線步驟 ground truth |
| 青衫攻略 | `reference/勇者鬥惡龍3:傳說的終章.html` | 交叉來源 |
| C remake 流程層 | `dq3_remake/src/main.c` + re-log-722 + flow-audit | **移植 spec**(宣稱先抽驗再信) |
| DQ3.EXE disasm | docs/05-17 | C 也缺的部分回頭 RE |

## 階段與進度

### A — 原版流程 oracle 建立 ⏳
- [x] A1 開場抽格完成(600 格/前 10 分鐘每秒);全片抽格進行中(每 2s)
- [x] A2 流程 oracle 文件 → `docs/66-original-flow-oracle.md`(開場 11 項 + 主線 60 列 + 精訊版差異 8 項 + 待核對 10 條)
- [x] A3 開場影格判讀完成、時間軸併入 docs/66 §1b(職業巡禮=attract 演出,由 docs/36 互動路徑仲裁);中後段細節於 C 階段按需釘

#### A3 初步發現:原版開場序列(影格判讀,時間=影片秒)
> 錄影 = 風林製作全程實況(含字幕指引)。標題 = TITG.P(已有)。

| 影片時間 | 畫面 | 對 remake 的意義 |
|---|---|---|
| 0:00 | 標題(DRAGON FIGHTER III 傳說的終章 ©1993 精訊) | 兩版都有 ✓ |
| ~0:45 | 暗場三人立繪(attract/開場演出) | **兩版皆無** |
| ~1:29 | 回標題(attract 循環) | 兩版皆無 |
| ~2:00–4:30 | **職業巡禮/創角序列**:PRIST→MAGICIAN→FIGHTER→WORTHY→PLAYER,每職業男女大立繪 + STR/DEX/CON/INT/LUK 五維能力條動態捲動 | **兩版皆無此畫面**(C 用酒場選單創角;ebitan 用預建隊)。立繪素材位置未知 → 需 RE(rulebook 64:已有截圖 oracle + 已破 4-plane/PCX 解碼器 → 掃資料檔比對) |
| ~4:30 | 過場(標題背景、無 logo) | — |
| ~5:00 | 進遊戲:**室內場景**(家/旅店;藍框視窗) | ebitan 從地表中心開始 ✗;C 待核 |
| ~8:00 | **四人縱列隊伍**(caterpillar)行走於神殿/城內 | **ebitan 只畫主角一人** ✗;C 待核 |

> 已解:WORTHY=賢者、PLAYER=遊玩者(docs/67 全鑑定);職業巡禮=attract 演出(docs/36 仲裁)。
> 殘:命名/性別在影片哪秒、謁見時間點(C-2 需要時再釘)。

### B — 差異盤點 ⏳
- [x] B1 FLOW-GAP 清單 → `docs/data/flow-gap-ebitan.md`(**78 GAP:A=33 B=12 C=14 ok=15**;port-C≈40/new-RE≈10/wire≈8)。分野:寶箱/USE 已對齊;NPC 給物鏈+boss 觸發系統性缺席(runner 0x722 層 0 起點);終盤連戰只缺入口;不死鳥 0 起點;商店不能賣;主角無命名
- [x] B2 C remake 宣稱抽驗 ✅:game_tester 實跑 **93/93 全綠**(2026-07-03);5 項流程宣稱 grep 屬實(里程碑綁 scripted NPC 檢查 main.c:1745、諾魯特 0x5b/0x215 @1623、sokoban gate @1847、王者之劍 transform @1559、酒場創角 @1496)→ **C remake 可當移植 spec**(Fable 抽驗 2 條逐字吻合)

### new-RE:attract/開場美術定位 ✅(docs/67)
- TIT*.P 20 檔全解:8 職業立繪(TITH-O,WORTHY=賢者)、巨龍 cutscene(TITB)、三人組(TITD)、
  1993 卡(TITE)、紋章(TITP)、無 logo 底(TITF)、THE END(TIT3);FIRST.SCR=plane-major raw。
  → attract 實作無素材障礙,剩輪播節奏影格對位。殘:TIT1/TIT2/TITA/TITC 鑑定。

### C — 修復批次(玩家遭遇順序 × B1 分析;每批 sonnet 實作、Fable 核實、分段請使用者實測)
| 批 | 內容 | 依賴 |
|---|---|---|
| **C-1 ✅** | 主選單→注音命名(預設注音,必填)→性別;共用元件;載入進度明確化;Fable 視覺核驗修 3 glyph 缺陷(commit a6d2de7) | 完成 |
| **C-2 ✅(部分)** | 創角完成→進 sec4 家+開場旁白 rec82→83(視覺核驗對齊原版)→母親帶出門到 sec0 城鎮。**節約做法:Opus 自己做、無大 agent**。TODO:能力值確認畫面、對母親對話才護送(現簡化為旁白後自動)、謁見金裝數字 | 核心完成 |
| C-3 | 酒場正式入口(露依達櫃台 NPC)+ 移除預建示範隊;U2/U3 修:出城方式(待 DOSBox 仲裁)、Space 行為、NPC B2<4 不丟 | C-1 |
| C-4 | examine→boss 觸發 dispatch + NPC 給物/交換鏈(port-C:boss-trigger-points + scripted 擴充 + 里程碑真實 gate) | — |
| **C-5a ✅** 取船鏈 | 救人(0x211,C-4a)→古布達 gate 給黑胡椒→波魯多加王(CTY37,座標吻合C)授船+msShip;dump 確認船地表可用 | 完成 |
| **C-5b ✅(晝夜)** | 晝夜系統:地表60步/相位·日夜NPC表word0/word1·palette調暗·ラナルータ/黑暗之燈接線·isNight()(commit 5955a70;日夜dump核 NPC24→6)。**遭遇區分區、商店賣出**另拆待做 | 晝夜完成 |
| C-6 | 六珠/不死鳥坐騎、下降自然觸發、終盤入口 wire(拔 Z/U) | C-4 |
| C-7 | attract 演出(TIT*.P 輪播,docs/67)+ 四人縱列 caterpillar + THE END 畫面 | 影格節奏對位 |
> C remake 端的 U2(Space demo 殘留 main.c:1982)另修一刀(等 DOSBox 確認原版 Space 行為後)。

### D — 驗收
- [ ] 無 debug 鍵全程可破關(照 docs/66 走)+ 使用者實測確認

## 紀律

- **完整性 > 投報(HARD,rulebook 83)**:不准以「低投報」跳過任何流程步驟;卡關換方法(靜態/錄影 oracle/實機對照)。
- **單元綠 ≠ 玩得通**:GAP 關閉標準是「正常玩家路徑可達 + oracle 一致」,非測試綠。
- **版權**:影格/截圖不入公開版控(dq3_real_video/ 已 gitignore)。
- **模型分工(rulebook 45)**:haiku=抽格/機械盤點;sonnet=文件合成/移植/code 盤點;Fable=協調/RE/oracle 判讀/獨立核實。

## FLOW-GAP 清單

→ `docs/data/flow-gap-ebitan.md`(78 GAP 全表,單一真相;本檔只留批次進度)。

## 待修既有 bug(subagent 揪出)
- ~~**scripted 給物無「已持有不重給」guard**~~ **✅ 修好(commit f0273b5)**:`scriptedTalk` 加
  `give!=255 && hasItem(give) → 開後話不重給`(對齊 C `dq3_inv_find<0`),消除 milestone==0 列的道具無限複製。
