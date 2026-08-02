# 74 — Go/Ebiten remake 完成計畫：原版實機畫面盤點 → 玩家流程閉合

> 建立：2026-07-28（Asia/Taipei）
>
> 目標：完成 `dq3_remake_ebitan/`，讓玩家不用 debug 鍵或 debug 環境變數，從新遊戲依精訊版正常
> 流程玩到 THE END。本文是接手後的執行計畫；歷史分析見 `docs/65`、`docs/69`、`docs/70`、
> `docs/73`，接手摘要見根目錄 `PROJECT_MEMORY.md`。

## 0. 完成定義

「完成 remake」必須同時滿足：

1. **玩家可達**：新遊戲起，全程只使用正式遊戲輸入，能走到結局。
2. **流程忠實**：主線入口、事件順序、gate、勝敗分支及道具／旗標副作用對齊精訊版。
3. **玩家可見忠實**：關鍵畫面、UI、NPC 可見性、日夜、隊伍呈現、轉場、音樂及音效有原版證據。
4. **可保存**：關鍵事件前後存檔／讀檔保持一致，不靠記憶體內暫態才能繼續。
5. **跨平台**：同一 Go game core 通過桌面與 Android；WASM 是後續 release target，不阻塞玩法完成。

以下不算完成：

- 測試直接呼叫 `descend()`、`startZomaSeq()` 或內部事件函式。
- 用 `Z`、`U`、`T`、`R`、Enter 等 developer/demo 捷徑抵達場景。
- 用 `DQ3_SHIP`、`DQ3_ENTER`、`DQ3_BATTLE` 等 debug env 取得正常玩法中拿不到的狀態。
- parser／公式／handler 單元測試全綠，但玩家入口不存在。
- raw asset 圖可解碼，但 production runtime 沒有顯示。

## 1. 原版畫面證據 inventory

### 1.1 本機完整實況

| 項目 | 內容 |
|---|---|
| 檔案 | `dq3_real_video/YTDown_YouTube_Media_J_fozjiKTB8_001_1080p.mp4` |
| 長度／大小 | 3538.37 秒（58:58）、1,097,105,763 bytes |
| 來源 | YouTube `J_fozjiKTB8` |
| 標題 | `PC 精訊版 勇者鬥惡龍3 快速通關攻略 Dragon Fighter 3 (Walkthrough)` |
| 作者 | 懷舊遊戲攻略頻道（`@fedaeltw`） |
| 已抽格 | `frames/`：每 2 秒一張，共 1,769 張；`frames_opening/`：開場每秒一張，共 600 張 |
| 覆蓋 | 標題至 THE END；主線場景、城鎮、地表、迷宮、隊伍、選單、船與終盤 |

證據限制：

- 影片有「風林製作」浮水印及後製指引字幕。
- 角色常見 H999/M999 等修改值，屬快速攻略／修改狀態。
- 有剪輯與快速通關選路。
- 因此可作 **場景外觀、UI 結構、隊伍排列、事件順序與路線 oracle**；不可單獨作為
  傷害、經驗、金錢、難度、隨機率或精確 timing oracle。

抽樣已確認：

| 約略時間 | 畫面證據 |
|---|---|
| 0:00 | 精訊 `DRAGON FIGHTER III 傳說的終章` 標題 |
| 5:00 | 阿里阿罕室內，主角單人起步 |
| 10:00 | 四人縱列穿越通道 |
| 15:00 | 水域／橋梁／寶箱場景 |
| 20:00 | 洞窟／boss NPC 場景 |
| 25:00 | 日邦格室內與隊伍互動 |
| 30:00 | 地表六指令選單、道具窗、四人 HUD |
| 35:00 | 四人縱列地表特殊場景 |
| 45:00 | 城鎮／城堡地圖與四人縱列 |
| 50:00 | 深色水域地表與完整指令／道具 UI |
| 55:00 | 終盤神殿與劇情對話 |
| 58:58 | `THE END / TO BE CONTINUE DRAGONFIGHTER I` |

### 1.2 本機 DOSBox 第一手 oracle

- `docs/36_shots/`：開場、標題、主選單、注音／英數姓名、功能列、性別、家中、CTY00。
- `dosbox/`：大量原版啟動畫面、道具、NPC 對話、城鎮、船、導航及事件截圖。
- `work/dosbox/`：版權截圖，不入版控。
- 可重現工具：`tools/dosbox_*.sh`、`tools/dosbox.conf`。
- 已有文件：`docs/14`、`docs/29`、`docs/30`、`docs/36`。

第一手 DOSBox 的權重高於網路影片；影片負責找場景與路線，DOSBox 負責對同一狀態、同一輸入做仲裁。

### 1.3 網路補充來源

1. [快速通關實況](https://www.youtube.com/watch?v=J_fozjiKTB8)：與本機影片同源。
2. [杜勝利精訊版攻略保存頁](https://vv0817.neocities.org/gametxt/09_dq3/09_dq3)：
   明載 `Space=遊戲選單`、`Enter=調查／交談`、F5/F6 存讀檔，並描述母親→國王→酒場二樓登錄
   →一樓招募的正常流程。
3. [精訊版回顧與四張實機圖](https://home.gamer.com.tw/artwork.php?sn=6041594)：
   標題、主選單、世界地圖 HUD、王城畫面；可交叉核對外觀。
4. [精訊開發者時代背景訪談](https://www.sfoxstudio.com/10464/how-did-i-get-into-the-game-8/)：
   說明精訊版是自行開發且未完成，不應假設 FC 原版每個細節都已存在於精訊版。

網路文章只作外部佐證；玩法的最終仲裁仍是本機原版、原始資料與使用者實測。

## 2. 玩家可見畫面盤點

| 畫面族 | 原版證據 | Ebiten 現況 | 判定 |
|---|---|---|---|
| 年代／巨龍 cutscene | DOSBox、TIT/FIRST 資產 | 未完整 production 接線 | GAP |
| 標題 | DOSBox、影片、網路圖 | 已有 | 有基礎，需 lifecycle 對拍 |
| attract 職業巡禮 | 影片、`docs/67` TITI–TITN | 素材定位，未播放 | GAP |
| 主選單 | DOSBox、網路圖 | 已有新遊戲／載入框架 | 需輸入與版面 E2 |
| 主角姓名／性別 | DOSBox 逐鍵截圖 | 已有共用元件 | 核心已接，需完整 trace |
| 家中／母親 | DOSBox、影片 | sec4+rec82/83；目前自動護送 | 部分，互動語意不忠實 |
| 王城謁見 | 攻略、影片、地圖 | 缺完整正式 gate／獎勵核對 | GAP |
| 酒場／登錄所 | 攻略、D3TXT、地圖、EXE handler | 正式入口與四人隊正常輸入 trace 已閉合（2026-07-28） | E3 |
| 四人縱列 | 影片多處 | production 主要仍單一主角 sprite | 高可見 GAP |
| 城鎮／洞窟 | 影片、全 CTY render、DOSBox | 通用 loader/render 已有 | 需事件與 entrance closure |
| NPC／日夜 | DOSBox、RE | 三層可見性與晝夜已有 | 機制有，逐事件 flags 待接 |
| 地表／HUD | 影片、網路圖 | 地圖、主角、panel 已有 | 需四人 HUD／入口／palette |
| 指令窗 | 影片、攻略操作說明 | 2×3 六指令已有 | 需所有子選單與 Enter 語意 |
| 道具／裝備／狀況／咒文 | 影片、EXE field caster/handler | 魯拉、烈米特、特黑洛斯、拉那魯達的 MP／gate／核心效果為 D2 | 其餘工具咒與逐窗仍需 E2 |
| 戰鬥 | DOSBox／影片／原始怪圖 | 公式、多敵、狀態、boss queue 已有 | 呈現／訊息／cue 長尾 |
| 商店／旅社／教會／達瑪 | 原版資料、部分截圖 | 達瑪轉職已由正式流程閉合至 E3/V2；其餘多數已有 | 賣出、逐服務與達瑪原版同狀態 V3 仍缺 |
| 船 | 影片、DOSBox 截圖 | 航行系統與取船鏈已有 | 需正常流程 E3 |
| 不死鳥／飛行 | 攻略、影片、EXE／CTY70 | 六珠祭壇、復活、搭乘、飛行與降落已有正式 trace | 已完成事件切片；全流程仍待稽核 |
| 下降／下世界 | 影片、EXE | 巴拉摩斯後王座事件與自然下降入口已接 | 已完成事件切片 |
| 終盤連戰 | 影片、RE | 光之珠、隱藏樓梯、歐里狄加與三連戰已有正式入口 | 已完成事件切片 |
| THE END | 影片、TIT3 | 戰後回城、冊封與 ending scroll 已接 | 已有 production trace；逐畫面仍待 V3 |
| 音樂／音效 | 原版 MCX/VOC、錄音研究 | OPL2/OGG 有相當基礎 | 逐場景 cue matrix 待驗 |
| 觸控／Android | 無原版對照，屬 port UX | 輸入抽象與觸控已有 | 正式流程完成後驗收 |

## 3. 現況判斷

底層並非從零開始：

- 資產 parser、地圖、文字、NPC、道具、戰鬥、音樂、存檔與觸控均有大量成果。
- 主角命名／性別、開場家中、日夜、NPC 三層可見性、船、取船鏈、boss trigger 基礎、
  多敵、狀態咒文等已比 `docs/data/flow-gap-ebitan.md` 初版盤點進步。
- R-2 沙曼歐莎的靜態 spec 已定。

真正問題是：

1. 沒有一條 production input 驅動的完整主線 trace。
2. `spine_test.go` 直接呼叫內部函式，不能證明玩家可達。
3. ~~production 的 `T/R/U/Z/Enter/Cancel` demo/debug 行為~~（R-4/R-5a 已移除）。
4. 六珠／不死鳥／飛行、下降與終盤正式事件切片已完成；尚缺的是從新遊戲開始、不改狀態的
   單一路徑通關稽核，不能用各切片測試取代。
5. 畫面證據沒有形成逐界面的 current matrix，raw decode、component dump 和正常流程畫面容易混稱。
6. `build.sh` 內 `go test | tail` 未啟用內層 `pipefail`，可遮蔽測試失敗；圖形測試應以
   明確啟動的 Xvfb 加 `go test -c` 執行，避免 `xvfb-run` wrapper 卡住而誤判程式失敗。

### 3.1 使用者校正：首要根因是設定資料未 match

2026-07-28 使用者再次指出：遊戲一開始就與原版不同；大量機制其實已還原，但餵給機制的
初始設定、位置、旗標與事件資料不正確。這比「缺入口」更精確。

因此後續不得再使用以下推理：

- section header 有 spawn，所以新遊戲一定使用該 spawn。
- C remake 有某個初值，所以它就是原版初值。
- 攻略說國王給裝備，所以可以在 `NewGame()` 直接預先給裝備。
- 某 milestone 最終會完成，所以可在初始化時先設完成。
- 某 event handler 已存在，所以用合理座標接上即可。

正確做法是追原版 EXE **誰在何時寫入**：

- current CTY／section；
- player X/Y、overworld remembered X/Y；
- party count／member records；
- equipment／inventory／gold；
- story flags／NPC visibility；
- dialogue bank／record；
- scene resource variant／palette；
- input owner／event runner；
- transition destination；
- BGM／SFX cue。

### 3.2 開場第一輪 EXE trace 已確認的 mismatch

工具：專案既有 `tools/dis.sh`（Docker + Capstone），直接反組譯原版 `DQ3.EXE` file
`0x13a0` 起的新遊戲初始化；位址口徑為 file offset。

| 原版 write（file） | 原版語意 | Ebiten 現況 | 判定 |
|---|---|---|---|
| `0x13a0 [0x4f2f]=0x99` | remembered overworld X | `NewGame`:`0x99` | match（2026-07-28） |
| `0x13a6 [0x4f31]=0xae` | remembered overworld Y | `NewGame`:`0xae` | match（2026-07-28） |
| `0x13ac/0x13b2` | 另一組 current/saved world X/Y 同為 `0x99/0xae` | `overPx/overPy=0x99/0xae` | match（2026-07-28） |
| `0x13b8 [0x4f50]=4` | current/saved section = 4 | 另有自造 progress 狀態 | 需逐欄對帳 |
| `0x13be [0x256a]=4` | CTY00 section 4 | `startOpening()` 載 CTY00 sec4 | match |
| `0x13c4 [0x4f52]=0` | CTY/scene side state | 尚未建立同欄證據 | 待追 |
| `0x13f6 [0x4f2d]=1` | scene/layer state | Ebiten `inTown=true`，但非逐欄移植 | 待對帳 |
| `0x13fc [0x256c]=0` | CTY index 0 | `curCty=0` | match |
| `0x140d [0xb34]=1` | opening/input/render state | 無對應 evidence contract | 待追 |
| `0x1425 [0x4f33]=5` | 玩家室內 X = 5 | `openingHomeX=5` | match（2026-07-28） |
| `0x142b [0x4f35]=5` | 玩家室內 Y = 5 | `openingHomeY=5` | match（2026-07-28） |

另外：

- 原版角色初始化 `(file 0x1c01..0x1c4e)` 已再追通：
  - 清角色 record 狀態區；
  - 八格 inventory 先全部寫 `0x00ff`；
  - `[member+0x15]=1`（Lv1）；
  - 第一格寫 `0x801e`，即 equipped bit `0x8000` + `0x1e` 布衣；
  - 初始化／學咒文後把 max HP/MP 複製到 current；
  - 最後 `[0x722]=1`，明確是**單人隊伍**。
- 以上出生狀態 mismatch 已於 2026-07-28 修正並由 `TestOriginalNewGameInitialState` 鎖定：
  `heroGold=0`、單人、只裝備布衣 `0x1e`、不提前設 `msStart`。
- 沿 caller/runner 繼續追到 file `0x140a..0x1633`：
  - 家中對白順序 `rec82→83→81`；
  - handler54 自動帶至 CTY00 sec0 `(8,38)`、`set flag17/clear flag50`、播 rec80；
  - 王座 handler56 播 rec78，GIVE `00,01,01,03,1f,1f`，加 `0x32=50G`，
    `clear flag17/set flag18`。
- Ebiten 已按同一 transaction 接線，`TestOriginalOpeningEventTransactions` 從真實 CTY 素材驗證
  母親落點、story flags、王座 region、精確獎勵與一次性。

2026-07-28 後續已補：

- `sub_ed3c/sub_fa57` 的 Lv1 持久能力生成與「這個人可以嗎？」Yes/No loop；
- 能力表 writer 證實順序 `STR,VIT,AGI,HP,MP,INT,LUCK`，撤銷誤改 VIT 的舊 MP patch；
- 從標題經正式 `InputState`、碰撞與 portal 一路走到王座的 E3 trace。

尚未完成的是能力確認畫面人物立繪逐像素對拍與母親逐格移動動畫；不能因 transaction
與 opening trace 通過就把整個 P1（酒場、四人隊、正常出城）標成完成。

## 4. 證據與驗收等級

每個 GAP 使用以下狀態：

- **E0 — 線索**：攻略、影片目視或合理推測。
- **E1 — 原版證實**：原始資料、EXE RE 或 DOSBox 同操作證實。
- **E2 — runtime parity**：Ebiten 以正式 renderer/input 在固定狀態重播，畫面／狀態與 oracle 相符。
- **E3 — 玩家流程閉合**：從上一個正常節點只用正式輸入抵達，事件後可繼續且 save/load round-trip 通過。

主線 GAP 只有 E3 才能標完成。畫面可另標 V0–V3：

- V0 無畫面。
- V1 raw asset／fixture。
- V2 Ebiten runtime screenshot。
- V3 同狀態原版／Ebiten 對照核實。

設定資料另標 D0–D3：

- **D0**：remake hardcode／推測值，無原版來源。
- **D1**：攻略、影片或 C remake 間接線索。
- **D2**：原始資料或 EXE 單一 write/read 已證實。
- **D3**：完整 writer → consumer → scene/event side effect 已追通，且 DOSBox 同狀態驗證。

所有 production 初值與主線 event config 至少 D2；會改變流程的值必須 D3。

### 4.1 每個差異的強制 RE 追蹤鏈

```text
影片／DOSBox 發現畫面或行為差異
  → 固定同一個玩家 checkpoint
  → 找 EXE scene/event owner
  → 逐指令列出 writes、calls、tables、resource IDs
  → 回原始 CTY/D3TXT/ITEM/D3MNS/flag image 驗值
  → 追 consumer，證明值如何影響畫面／流程
  → 寫成 typed Go config/state
  → 正式 input trace 抵達
  → 同狀態 screenshot/state/save 對拍
```

反編譯規則：

- 凡有 RE／反組譯需求，優先在一次性 Docker 容器內使用
  `/home/anr2/ida_94_official/dist` 的 IDA Pro 9.4；不得先用 Ghidra 產生結論後才補查
  IDA。`tools/dis.sh`／Capstone 用於可重現的批次位址掃描，Ghidra 只在 IDA 無法處理
  或需要第二套工具交叉驗證時補用；禁止直接在 host 執行任何專案作業。
- `tools/re_disasm.py` 是 file offset；所有文件位址明標 `(file)` 或 `(logical)`。
- 含 indirect jump／switch 時直接讀原始跳表；Ghidra 需要
  `JumpTable.writeOverride()`，不可只信自動 decompile。
- decompiler 行數異常膨脹或大量 unreachable 時視為不可信，回原始指令。
- 未知設定 fail closed，不以 generic fallback 靜默取代。

## 5. 執行計畫

### P0.5 — 精訊版 DQ 共用 game pack

使用者於 2026-07-29 決定把遊戲設定外部化，讓已證實的參數可改 JSON 而不需重新編譯
desktop 執行檔，並讓同一 Ebitengine 核心後續承載精訊版 DQ1／DQ2。canonical 欄位與
邊界規格見 `docs/84-game-pack-json-contract.md`。

執行約束：

- DQ3 是第一個 pack 與 fidelity 基準；不可先為了抽象漂亮而中斷目前的玩家流程閉合。
- 優先遷移正阻塞流程的教會／旅館／商店，再處理成長、咒文、怪物、遭遇與事件。
- 原始 DAT／EXE decoder 保留並成為 JSON parity oracle；資料搬家不等於重新猜一次數值。
- JSON 必須版本化、嚴格拒絕未知欄位、驗證 ID／範圍／資源，並攜帶 evidence provenance。
- Desktop 支援外部 pack 後才可宣稱修改 JSON 免重編譯；Android／Web 嵌入資料不作此宣稱。
- DQ1／DQ2 以 capabilities 表達實際差異，不建立假的 DQ3 職業、隊伍或交通工具資料。

### P0 — 基建與單一真相

交付：

- 修正 `build.sh` fail masking；拆出快速單測、Xvfb game test、長時間 visual dump。
- 找出 game test hang／長測試，所有測試需有 timeout 與進度輸出。
- 建立 `docs/data/ebiten-player-flow-matrix.md`：`docs/66` 每一步對應入口、code owner、E/V 等級。
- 建立 debug inventory：`T/R/U/Z/Enter/Cancel` 及所有 `DQ3_*` env，分類為 test-only 或 production leak。
- 建立 deterministic input trace harness：按正式 `InputState` 推進，不直接呼叫劇情函式。
- 建立 `original-state-ledger`：每個初值／事件值列 EXE write、data source、consumer 與 D 等級。
- 先完成開場 `0x13a0` scene owner 的 full trace，不再沿用 header spawn fallback。

Gate：

- CI 中任何測試失敗必須非零退出。
- trace 能從標題走到家中，再保存 screenshot/state checkpoint。

### P1 — 新遊戲至四人隊出發

範圍：

1. 反組譯完整新遊戲 init 與 scene owner，逐欄移植 CTY、section、`(5,5)`、world remembered
   `(0x99,0xae)`、party、inventory、gold、flags、dialogue/input state。
2. cutscene／標題／主選單最小正確 lifecycle。
3. 主角姓名、性別、**能力值確認畫面**與其 Yes/No loop。
4. 家中旁白的 exact dialogue lifecycle。
5. 反組譯母親 NPC 的 event owner；依原版輸入與 flag 實作護送，不再旁白後自動瞬移。
6. 王城謁見、父親劇情、國王給金錢與裝備的 exact transaction；獎勵前後分開存檔驗證。
7. [x] R-1：以 EXE sub2 jump table + CTY NPC identity 定位 CTY00 酒場／登錄所正式入口。
8. 登錄角色 → 一樓招募／分離 → 隊伍最多四人。
9. [x] 移除 placeholder 同伴；開局單人，登錄／招募後才成四人隊。
10. 反組譯 caterpillar owner、歷史座標 buffer 與 draw order；不是只在主角後方手調三個位置。
11. 核對正常出城／進城操作；移除 Enter/Cancel demo 行為。

E3 trace：

`boot → title → new game → name → gender → home → mother → king → registration →
recruit 3 → leave Aliahan → overworld`

### P2 — 前段主線閉合

依 `docs/66` 玩家順序，不以系統類別跳做：

- [x] 盜賊鑰匙：2026-07-29 已由既有 boot trace 延伸，經 CTY07 地道、原始 transition
  table 跨至 CTY08 sec3、正式交談取得 0x55，並通過 save/load。
- [x] 魔法球：2026-07-29 從 0x55 checkpoint 正常返回 CTY07 地道、步行至 CTY01，
  以正式「調查」開右上屋 door(28,8)，進 sec1 與 handler7 交談取得 0x58；0x55 不消耗，
  checkpoint save/load 通過。
- [x] 誘惑洞窟、羅馬利亞首次謁見：正式使用魔法球、穿越 CTY30／31、抵達 CTY02，
  rec45→rec15 與 save/load 已閉合（`docs/81`、`docs/82`）。
- [x] 甘達特／金皇冠／臨時王位：正式練級北行、擊敗甘達特、求饒、取冠、夜間回城住宿、
  還冠強制 Yes/No、臨時角色圖、地下競技場兩層辭位、兩端 save/load 與離城均達 E3；
  事件與文字已由 `temporary_role` game-pack primitive 資料化（`docs/82`、`docs/85`）。
- [x] 諾阿尼魯／紅寶石／覺醒粉：2026-07-29 已由羅馬利亞辭位 checkpoint 正式步行至
  CTY04／05／11；原始 treasure present flag、女王原地換物、CTY-only 使用 gate、
  `SET 0x26 / CLEAR 0x31`、rec599、即時 NPC 場景重載、save/load 與讀檔後離村均達 E3。
  版本專屬 selector、item、flags 與文字已遷入 `quest_item_chain_events`（`docs/86`）。
- [x] 金字塔按鈕／魔法鑰匙：原版右→左開關、陷阱、石門、寶箱、正式輸入與
  save/load 已閉合（`docs/87`）。
- [x] 波魯多加首次晉見／國王的信：2026-07-29 已由魔法鑰匙 checkpoint 正式步行，
  經依席斯、羅馬利亞祠堂、CTY16 旅店切回白天後進 CTY37；handler26 的 flag `0x37`、
  item `0x5b`、D3TXT04 rec24／25、對話關閉後交易與 save/load 已達 E3。事件內容由
  `staged_vehicle_exchange` game-pack primitive 提供，舊 remake flag `0x215` 已廢除
  （`docs/50`）。
- [x] 諾魯德密道：2026-07-29 已由 boot trace 正式施放 Rura 至已訪問城鎮並步行到
  CTY62；持信從 `(38,6)` 交談，handler50 只讓 NPC 走到 `(50,21)` 後交回控制，
  玩家再自行走到 `(50,15)` 的 subid1／handler57 格。rec87–90、兩段 NPC raw movement、
  國王信不消耗、`clear 0x16/0x4c + set 0x1c`、事件後 save/load 均已閉合；事件由
  `guided_passage` game-pack primitive 提供（`docs/88`）。
- [x] 巴哈拉達救人／黑胡椒：2026-07-30 已由諾魯德東口 checkpoint 正式步行至 CTY15
  與 CTY14，經原始 transition `(26,3)`、tier2 魔法門、四守衛、handler59／60 三段 NPC
  movement、甘達特 `monster56×1 + monster57×3`、求饒及古布達一次性交易取得 `0x5c`。
  原版 flags／formation／movement／D3TXT、runtime PNG、事件後 save/load 均已閉合；
  事件由 schema `0.1.5` `hostage_rescue` game-pack primitive 提供（`docs/89`）。
- [x] 回波魯多加取船與首次航行：2026-07-30 已由黑胡椒 checkpoint 經正式標題讀檔、
  巴哈拉塔旅店補給、魯拉、CTY16／37 transition、handler26 交換、取船後 save/load、
  正式登船與航行閉合。item `0x5c` 消耗、flag `0x38` 保留、flag `0x2c` 清除、船停泊
  `(25,73,L0)` 及 runtime PNG 均已驗證；事件沿用 `staged_vehicle_exchange` game-pack
  primitive（`docs/90`）。

每段包含：

- EXE writer/consumer trace 與 D2/D3 設定 ledger。
- 正常入口和地圖可達性。
- NPC/dialogue/event gate。
- boss 勝敗。
- 道具／旗標。
- NPC visibility。
- save/load。

Gate：不用 `DQ3_SHIP` 可自然取得船並航行。

### P3 — 中段事件與六珠來源

按攻略閉合：

- 日邦格八頭大蛇兩戰。
- 隱身草與耶進貝亞石塊解謎。
- 乾渴壺／最終鑰匙。
- 提頓夜間綠寶珠。
- [x] 勇氣神殿單人 gate／藍寶珠：handler37 四人→單人、CTY23 `0x67`、handler62 復隊、
  試煉途中及復隊後 save/load 均由 boot production trace 閉合；見 `docs/100`。
- [x] 海盜村密道／紅寶珠：CTY27 `ctrl bit0x40` 可推入口、hidden transition、
  event0 `0x68`／flag `0x3f`、魯拉船重定位、正式航行與 save/load 均由 boot production
  trace 閉合；見 `docs/101`。畫面尚為 V1。
- [~] 商人建城／黃寶珠：已用 IDA 9.4 推翻舊 C／Go 反向 portal 表；同座標
  `(210,64)` 的 CTY58→59→60→61→83 ordered gate 已由 game-pack JSON 載入，初始
  `flag0x23 clear` 正確進 CTY58。三個 stage gate 已分別閉合到最終鑰匙 `0x47`、
  假王事件 `0x42`、蓋亞之劍 `0x48`，不是自製計時器。商人交付已由酒場登錄商人開始，
  正式航行進 CTY58，閉合兩次確認、共用預存所、身份保存與 save/load；
  runtime PNG 為 V1。CTY83 event `01 6a 00 4a`、handler48 record42→依 present
  flag0x4a 選擇 record43、建城者姓名插值、男女 sprite flag、黃寶珠 JSON transaction
  與 save/load component 已閉合（D3／E2／V1）；舊 Go treasure row 的反向 flag 語意已
  刪除。同一條 boot production trace 已由商人交付 checkpoint 自然重進 CTY60，並已
  走 CTY41→42、CTY43 關卡、CTY44、CTY24 拉之鏡與假王戰。尚缺蓋亞之劍 gate 後進
  CTY83 的 E3 trace、
  CTY59 handler41 設施分支及原版同狀態 V2；見 `docs/102`。
- [x] 沙曼歐莎 R-2：
  `道具選單使用拉之鏡 → CTY44 sec1 玩家站(14,7) → 夜晚 → rec97/98 → 怪89 →
  勝後0x62+clear21/set22+白天；敗／逃 clear10/restore42`。
- 變身杖 → 船員之骨 → 幽靈船 → 愛的回憶 → 蓋亞之劍 → 銀寶珠。

R-2 的 handler 與 flag writes 已由 EXE file `0x5682..0x5732` 定案；CTY24 event
`01 61 00 9f` 已由舊 Go table 遷入 game-pack JSON。IDA 9.4 再證實離城 consumer
file `0x49f9..0x4a40` 直接使用 transition X/Y，不讀 facing；舊 remake 的兩格推出已刪除。
campaign trace 已從 CTY60 正式走 CTY41→42、以 `(213,123)` 離場、經 CTY43 跨 CTY
portal 進沙曼歐莎，取得拉之鏡並 save/load，再完成夜間假王戰與變身杖交易，詳見
`docs/103`。

Gate：六顆寶珠均可由正常玩家取得，且存讀檔保持來源事件完成狀態。

### P4 — 不死鳥與飛行坐騎

先 E1 定案：

- 不死鳥祠堂 CTY／section／座標。
- 六珠放置是逐顆或一次性交付。
- 祭壇對話、動畫、音樂、旗標。
- 拉米亞 sprite、移動及起降規則。

Go contracts：

1. `MountState`：未取得／停泊／乘坐。
2. traversal：飛越陸海與不可降落 tile。
3. takeoff/landing：正式 input、位置與 facing。
4. town/portal：飛行時入口判定。
5. render：拉米亞 sprite、caterpillar 隱藏／恢復。
6. encounter：飛行中是否遇敵。
7. save：坐騎取得、位置、乘坐態 round-trip。
8. touch：與桌面同一抽象 action。

Gate：

- 六珠 → 祭壇 → 拉米亞 → 起飛 → 飛越障礙 → 合法降落 → 進城 → 存讀檔，全程 E3。

### P5 — 巴拉摩斯與自然下降

- 定位巴拉摩斯城 overworld 入口、CTY、boss tile。
- 飛行抵達。
- 城內寶箱及 boss 兩戰規則。
- 勝後阿里阿罕王城事件、索瑪現身、flag。
- 蓋亞那洞穴跳坑自然觸發下降。
- 自然路徑完成後移除 `U` production 捷徑。

Gate：從拉米亞起飛至下層世界，不使用 U。

### P6 — 愛列夫加特與終盤

- 龍王城／光之珠。
- 太陽之石、雲雨之杖、彩虹水滴取得鏈。
- 王者之劍的商店賣出／交易鏈。
- 彩虹橋位置與 walkability mutation。
- 父親／五頭龍場景與復原 sprite。
- 索瑪神殿正式入口。
- 城內大魔人維持原版一般隨機遭遇；終盤 formation 為怨靈 → 殭屍 → rec72 → 索瑪。
- 光之珠弱化、勝敗、洛特授勳、ENDTXT、TIT3。
- 自然入口完成後移除 `Z` production 捷徑。

Gate：由下層世界正常移動至 THE END，不使用 Z 或直接呼叫 boss queue。

### P7 — 視覺、聲音與支線 parity

建立 UI evidence matrix，逐項 V3：

- attract 職業巡禮與能力條 timing。
- 四人 HUD、狀況／裝備／咒文／道具窗。
- 戰鬥訊息、受擊／狀態演出。
- 日夜 palette 與 NPC。
- 商店賣出、教會、旅社、達瑪。
- 地表海面動畫、船、拉米亞。
- 全場景 BGM cue、EBG/VOC SFX。
- THE END timing。

支線也按正常入口驗收；不得用「不影響破關」永久跳過。

### P8 — Release

- 無 debug full playthrough，分段使用正式 save checkpoint。
- 使用者實玩驗收。
- Linux／Windows／macOS desktop packages。
- Android APK/AAB 真機：安全區、觸控、背景恢復、音訊、存檔。
- WASM smoke（若資產授權／載入策略允許）。
- developer hooks 以 build tag 或明確 developer mode 隔離。
- 公開包不含原版資產，啟動時驗證玩家提供的合法資產。

## 6. 工作順序與依賴

```text
P0 trace/build
 └─ P1 新遊戲/酒場/四人隊
     └─ P2 前段/自然取船
         └─ P3 中段/六珠/R-2
             └─ P4 不死鳥/飛行
                 └─ P5 巴拉摩斯/下降
                     └─ P6 愛列夫加特/終盤
                         └─ P7 parity
                             └─ P8 release
```

R-2 可在 P1/P2 同期作獨立功能切片，但不得因它的 spec 已齊就跳過新遊戲到中盤的可達性。

## 7. 每批固定交付物

每一批必須同時交付：

1. Spec：原版來源、位置、輸入、旗標、道具、勝敗及未知點。
2. RE ledger：writer、consumer、table/resource、file/logical 位址、D 等級。
3. Code：production route，不只 debug route。
4. Component tests。
5. Input trace：從上一個正常節點抵達。
6. State assertions：事件前後與 save/load。
7. Runtime screenshots。
8. Oracle comparison。
9. Worklist 更新：E/V/D 等級，不寫模糊「完成」。

## 8. 第一個實作 sprint

在開始 R-2 之前先做：

1. 修正 `build.sh` pipefail 與 game test timeout。
2. ~~完整反組譯 `file 0x13a0` 開場 owner 及其直接 callees，建立 opening state ledger。~~
   已追至 handler56(file `0x1633`)。
3. ~~先修正已證實 mismatch：室內 `(5,5)`、remembered world `(0x99,0xae)`、單人隊伍、
   布衣 `0x1e`、不提前設 `msStart`、不提前給國王獎勵。~~ 已完成並加回歸。
4. 建 input trace harness。
5. 建立 `title → name → gender → ability confirmation → home` trace。
6. `home → mother → king` transaction 已接；下一步補正式 input trace 與逐格演出。
7. [x] 反組譯 CTY00 酒場入口與原版出城操作；完整 production input trace 已通過。
8. 接 `king → tavern → recruit → overworld`。
9. [x] 以同一 RE/trace 方法完成 R-2 item-use vertical slice；EXE 精確 gate、正式戰鬥輸入、
   勝敗交易、存讀檔與 runtime 圖均完成。
10. 下一 sprint 依 P3 攻略順序盤點六珠來源，不直接跳到不死鳥祭壇。

這個順序直接處理過去最大失敗模式：先確保玩家真的走得到，再增加更多孤立機制。

## 9. 下一輪 breakdown（2026-07-28）

下一輪先修「機制存在但原版設定未 match」，不新增無證據的近似功能。

### A. 場景咒文交易與剩餘 handlers

1. [x] 修正共通 caster：MP 足夠後、effect handler 前立即扣除；取消魯拉目的地或場景
   gate 失敗不退款。既有取消／失敗測試已更新。
2. [x] 為 rec174／175／178／179 建場景咒文 ledger。rec180 descriptor flags=`0x01`，
   已證實是 battle-only 巴魯朋特，不屬於 field ledger；其 16-slot table 與四個唯一 handler
   已另記於 `docs/78`。
3. [x] 先落地證據鏈完整者：
   - rec178／Remoaru：15 MP、25 個成功移動步的隱形狀態與 renderer consumer。
   - 阿巴卡姆：0 MP、面向門的 tier-4 權限、3×3 連續門片 mutation 與失敗 rec `0x15a`。
4. [x] 因帕斯已確認 `bp=event type` 與原始 rec253／254；多拉瑪那已追通 `0x2000`
   writer、`0x0400` 連續區 guard 與兩類傷害地板 consumer。因帕斯色彩動畫仍為 V-gap。
   巴魯朋特已確認原版只有 5 個有效 slot（4 種唯一效果），其餘 11 格是無效訊息；
   battle core 已依 EXE 接入。
5. [x] 對本批落地咒文加入 production menu、正常輸入、MP／timer／restore 測試，重產並
   目視檢查 `dq3_remake_ebitan/docs/field_spell_menu.png`。原版暫態 timer 的 save 格式仍
   待 RE，不把 Go runtime-only policy 冒充原版契約。
6. [x] 戰鬥命令改為原版逐隊員收集：依 D3TXT00 rec441..444 顯示「會咒／可逃」組合，
   跳過死亡與失能隊員；每位隊員使用自己的 MP、咒文、敏捷與防禦狀態。收完命令後依
   `sub_d6bf` 將全隊與每隻敵人以敏捷擲值混排，敵人不再固定等全隊行動完才出手。
   production-input、四種選單、跳過條件及快慢敵我先後均有回歸測試。
7. [x] 單體戰鬥行動加入敵／我方目標選擇與取消回退；攻擊、單體傷害、荷依米系、
   藥草及單體輔助咒不再固定第一隻敵或施法者自己。咒文目標以原版 descriptor flags
   `0x0d/0x15/0x1d/0x09/0x0b/0x11/0x13` 重建，並修正拉里荷、瑪荷頓、瑪努莎、
   史克魯多、比荷瑪順、薩梅哈的群體範圍；拜基魯多／史卡拉改為逐隊員 buff。

### B. 玩家流程 audit

1. 由目前最早的合法 checkpoint 起跑 production-input trace，記錄第一個不可達、錯誤 gate
   或錯誤設定，不以 debug key 越過。
2. 對該 blocker 追 `writer → state/table → consumer → visible effect`，以 DOSBox／影片交叉驗證。
3. 一批只閉合一個玩家可達的垂直切片，並驗證事件前後 save/load；完成後更新 E／V／D，
   不寫單一「完成」。

### C. 批次驗收

1. 全部 Go tests 在能讀到 `assets_raw` 的正確工作目錄執行；素材缺失 `SKIP` 不算通過。
2. 重跑受影響的 DOSBox 同狀態案例，保存輸入與可見結果。
3. 檢查 runtime 圖是否仍代表現行程式；有重大畫面修改就更新舊圖。
4. `git diff --check`、確認不納入原版素材／IDA 檔／使用者 scratch files，再做一次重大
   commit 與 push。

## 10. 今日收尾與剩餘工作（更新 2026-08-02）

boot 起的同一條正式 trace 已由首次航行繼續至加爾那之塔《領悟之書》`0x4a`、
CTY17 達瑪神殿轉職，再閉合 CTY20《黑暗之燈》`0x5f`。
IDA Pro 9.4 證明原版玩家睡眠／混亂在行動前 `roll<=0x64` 清除、怪物側
`roll<0x64` 清除，修除怪物 43 令全隊永久不能行動的航海死循環。CTY18 sec1
`01 4a 00 b7` 已遷入 schema `0.1.6` `treasure_events`，正式路徑繞完塔內不同連通區，
命令窗調查後立即切換開箱 tile，並完成 save/load；證據與 runtime 圖見 `docs/91`。
schema `0.1.7` 另加入 `item_actions` 與 `reclass_events`。IDA 證明實際入口為
CTY17 sec0 handler39，而非舊文件的 CTY49；Lv20、勇者禁止、五個基本職業、遊玩者／
個人持有領悟之書才可選賢者，以及成功交易均已閉合。正式 trace 由領悟之書 checkpoint
正常離塔、航行、練所選同伴至 Lv20、使用魔法鑰匙、rec421 給予、完成賢者轉職並
save/load；證據與更新圖片見 `docs/92`。
schema `0.1.8` 再加入有限的 `item_use_effects`。IDA 證明 raw `0x5f` 經
DGROUP `0x366a` 派發表抵達 `(logical) 0x4063`：只允許地表且當前為白天，成功時設為
黑夜並重設時鐘步數，不消耗道具。正式 trace 已由達瑪 checkpoint 航行至 CTY20，
白天調查寶箱、出村後由 rec421 使用、完成 save/load，並繼續航行抵達 CTY19；
證據與更新圖片見 `docs/93`。

本輪插入的玩家可見 parity audit 已修正兩個會污染後續截圖判讀的共用錯誤。IDA Pro 9.4
閉合 DGROUP `0x3e6e` 的 caller→rect→consumer，確認原版對話外框是
`(152,238,360,112)`、文字 inset `(16,16)`、20 欄／4 行；schema `0.1.9` 已將其移入
`interface.json`。另由 `sub_1b1fe → sub_1b31a/sub_1b37c → sub_1b2af` 證實
`DQ3MNS.SHP` 透明度來自四色 plane 後的獨立 RLE AND-mask，不能把黑色色號 0 當透明。
component/parity tests 與 README 所引用的受影響 runtime PNG 已重產；逆向證據見 `docs/94`。
這批之後已由 CTY19 合法 checkpoint 閉合日邦格八頭大蛇兩階段事件。IDA Pro 9.4
證明 handler45 第一戰先消費 DGROUP `0x3cee` 的 NPC slot0 右移兩步，再 clear `0x44`、
set `0x20` 並讓玩家右移兩步轉至 CTY21 rec70；handler36 的 Yes 分支可重談，No 分支才
以另一背景開第二戰並由 `DGROUP 0x2518 bit1` 抑制掉落。勝利後 clear `0x20/0x45`、
set `0x1f`，再由同格寶箱取得紫寶珠。怪75的草薙大劍則由 D3MNS `+0x25/+0x26`
原始掉落，不是事件直接給予。schema `0.1.10`、正式 InputState trace、save/load 與四張
runtime 圖均已更新；見 `docs/95`。

schema `0.1.11` 已接續閉合隱形草與愛丁貝亞守衛。IDA Pro 9.4 證明 item `0x5d`
handler `(logical) 0x3fff` 會清所選物品槽、寫 25 步 timer 並 set `DGROUP 0x4f46`
的隱形 bits；CTY39 sec0 raw handler29 則在可見時把 slot0 守衛沿 X 追到玩家同欄，
bit `0x8000` set 時略過追蹤。正式 trace 由原始 CTY38 貨架買入，在 `(14,38)` 由
rec421 使用，從右欄越過守衛進 CTY76，並完成 save/load。兩張新版 runtime 圖與完整
證據見 `docs/96`。本輪同時推翻「CTY47/75 就是隱形草貨架」的舊路線假設；CTY47/75
是朗錫爾／勇氣神殿雙入口，實際含 `0x5d` 的原始商店表是 CTY38。

schema `0.1.12` 再閉合 CTY76 推石與乾渴壺。IDA Pro 9.4 證明玩家碰撞 consumer
`(logical) 0xde2a`（IDA linear `0x1de2a`）是檢查 runtime NPC `ctrl bit0x40`，不是舊 C prototype 猜的
`B2==40`；隱形 bits `0xc000` 仍在時反而禁止推動。handler30 `(logical) 0x586b`
固定檢查 NPC slot0–2 的 Y byte 是否全為 5，成立才 clear flag `0x3b` 並刷新 passage。
正式 trace 由 CTY76 checkpoint 只送方向鍵走完 119 步解路，再由命令窗調查 event0
取得乾渴壺 `0x5e`，完成 save/load。兩張 runtime 圖與證據見 `docs/97`。

schema `0.1.13` 已接續閉合海中淺灘與最終鑰匙。IDA Pro 9.4 證明 item `0x5e` handler
`(logical) 0x4012` 只接受船上 `(146,53)`，成功 set 原版 world-state bit `0x08`，以
DGROUP `0x3a54` 的 20-byte 表覆寫 `(144,50)` 起 `5×4` 地表並保留乾渴壺。正式 boot
trace 已由 CTY76 checkpoint 返回地表、登船航行、用正式道具窗顯出 CTY40 入口、向上進祠堂、
調查取得最終鑰匙 `0x57`，並驗證地圖與道具 save/load。三張 runtime 圖與證據見 `docs/98`；
原版 palette transition 與藍色水面仍是 V3 GAP。

schema `0.1.14` 已接續閉合最終鑰匙的原版輸入 owner、魯拉載具副作用與提頓綠色寶珠。
三把鑰匙不再由「調查」掃描全隊最高 tier；玩家必須由 rec421 道具選單明確使用，成功
開門且不消耗。IDA Pro 9.4 的 Rura handler `(logical) 0x4c2b..0x4cf8` 證實持有船時會
以 DGROUP `0x0ad4` 的平行表重定位船；CTY15 的船落點為 `(104,126)`。同一條 boot trace
已由 CTY40 正式魯拉、登船航行、在提頓入口外使用黑暗之燈、以最終鑰匙開 tier3 牢門，
再由夜間 handler35 把綠色寶珠交給第一個有空格的隊員並完成 save/load。見 `docs/99`。

schema `0.1.15` 已接續閉合蘭西爾勇氣試煉與藍寶珠。IDA Pro 9.4 證明 handler37
`(logical) 0x59e4` 保存 active party count、強制單人、寫 world `(82,165)` 並設 mode bit
`0x80`；handler62 `(logical) 0x608a` 原樣復隊並清 mode bit。舊文件「接受即 set flag0x13」
及反向 `owPortal` 已推翻；flag0x13 writer 仍 unknown，不在 production 合成。同一條 boot trace
已經正式航行、住宿切白天、以最終鑰匙開門、接受挑戰、走 CTY23、取得藍寶珠、途中
save/load、返回 CTY75 復隊及再次 save/load；見 `docs/100`。

海盜村密道與紅寶珠切片已在當時閉合；現行 schema／content 版本只以
`PROJECT_MEMORY.md` 與 pack manifest 為準，不在歷史 checkpoint 重複保存舊版號。
IDA Pro 9.4 證明 NPC runtime `ctrl bit0x40` 的推動 gate 與 tile subid transition consumer；
CTY27 raw 證明 `(26,9)` 可推物件、hidden transition→sec1 `(5,9)` 及 event0
`{item=0x68,flag=0x3f}`。正式 trace 必須先魯拉 CTY15 觸發原版船重定位，才可登船航行至
CTY27；取得後已完成 save/load 及入口物件 visibility 驗證。原版同狀態影片畫格尚未定位，
故畫面只標 V1；見 `docs/101`。

下一輪依下列順序接手：

1. **主線最高優先**：boot 起的同一條 trace 已正式完成商人交付、CTY41→42、CTY43
   關卡、拉之鏡、假王／怪力魔、CTY54 變身杖交換、船員之骨定位、幽靈船動態物件／入口，
   並由 CTY36 原始轉場及寶箱取得愛的回憶後完成 save/load。下一輪從這個合法 checkpoint
   閉合地表 `(76,54)` 奧莉薇亞海岬、蓋亞之劍與 CTY83 黃寶珠；詳見 `docs/104`，不得直接
   從後段 component checkpoint 起跑。
   不得因 P3 後段已有孤立 component tests，就跳過兩節點之間的玩家路徑、資源與存檔 gate。
2. **設定資料追蹤**：每個 blocker 都先由 IDA／原始指令追
   `writer → table/state → consumer → visible effect`，並以 DOSBox 同狀態核對；不得用
   C remake、攻略或推測值直接補 production config。
3. **完整通關證明**：P3 至 P6 雖有多個正式入口事件切片，仍缺
   `new game → THE END` 不改狀態、不中途呼叫內部函式的連續 campaign trace，以及各關鍵
   checkpoint 的 save/load round-trip。
4. **戰鬥長尾**：睡眠／混亂恢復時序與 99／100／101 邊界已閉合，但訊息的角色／怪物
   名稱代入、逐動作停頓、動畫／音效 cue、其他抗性與狀態、敵群 formation、
   boss 多次行動、掉落及逐項咒文
   效果仍需 RE 與同狀態驗證。
5. **玩家可見 parity**：四人縱列與 HUD、能力確認立繪、母親逐格演出、attract、
   商店／教會／旅社／達瑪、二選一小視窗原版幾何、日夜 palette、剩餘 BGM／SFX 及
   ending timing 尚未全部 V3。二選一小視窗現行 geometry 只有 legacy renderer，尚無 D3
   結構證據，不得直接寫入 game-pack canonical JSON。
   教會復活確認的現行 runtime 證據為 `docs/img/church_revive_confirm.png`，仍待 DOSBox
   同狀態 V3。
6. **Release**：完成正式流程後才做跨桌面平台包裝、Android 真機的觸控／lifecycle／音訊／
   存檔驗收，以及選配 WASM smoke；公開包不得納入原版資產。

下一輪第一項工作應是執行玩家流程 audit 並記錄「第一個實際 blocker」，不是再挑一個孤立
機制實作。
