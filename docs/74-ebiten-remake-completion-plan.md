# 74 — Go/Ebiten remake 完成計畫：原版實機畫面盤點 → 玩家流程閉合

> 2026-08-12 發佈 checkpoint：v0.1.34 已由 `9d639d0` 正式發布；現行 hash、
> patch／full 素材邊界及驗證限制見 [`docs/131`](131-release-v0.1.34.md)。
>
> 2026-08-22 戰鬥勘誤：[`docs/148`](148-battle-single-action-queue-spec.md) 已證實本 EXE
> 每個存活 actor 每回合恰好一筆，沒有 Boss repeat-N。本文較早 milestone 中將其列為
> `unknown`／後續工作的段落只保留歷史，不得重新開啟或猜 JSON。
>
> **2026-08-22 唯一現況：**目前工作樹已在乾淨 Docker＋Xvfb 中，從標題以正式
> `InputState` 重播至 `THE END`（115.629 秒），campaign 恢復 E3。下文 monster77、
> monster51 與各補給失敗是促成本輪訂正的歷史 checkpoint，不再是 current blocker。
> 本結果只證明主線玩家流程；逐畫面、逐音效及硬體 timing 仍依各列維持 V1／V2／unknown。
>
> **2026-08-22 最後功能 worklist（不含回歸／發包）：**依序只做
> ①戰鬥道具、②原始怪物表實際使用且尚未閉合的 action、③剩餘合法野外咒文、
> ④商店賣出、⑤船進城載具交易、⑥必要選單與玩家可見支線語意。完成一項後才進下一項，
> 每項固定走「既有 RE → IDA Pro 9.4／IDAPython 補證 → spec → game-pack／runtime」。
> PCM hardware wall-clock、DAC、PIT、DMA 等平台時序只引用 Wiki／datasheet／成熟模擬器
> 規格並採可重現近似，不再列為遊戲 RE 或 remake 完成 gate。

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
5. **跨平台**：同一 Go game core 已建立本輪指定的 Linux AppImage、Windows ZIP 與 macOS ZIP；
   Android／WASM 不屬本輪 release gate。

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

> 閱讀閘門：下表的 E3 由本檔最上方最新完整 campaign 仲裁。`docs/153` 的 monster77
> 失敗與本檔後段 monster51 失敗都是歷史 checkpoint，已由 `docs/154..178` 的正式玩家
> 路線與交易訂正越過。逐畫面／逐音效 V3 仍依各列分開判定。

| 畫面族 | 原版證據 | Ebiten 現況 | 判定 |
|---|---|---|---|
| 年代／巨龍 cutscene | DOSBox、TIT/FIRST 資產 | `opening` 五張 PCX 已由 game-pack 載入；桌面／mobile 正式入口可無輸入播放、正式輸入跳過並交回標題 | E2／V1；素材 identity 為 D2，120 幀停留、排序／淡入淡出、TITP 位置與音效仍待 V3 |
| 標題 | DOSBox、影片、網路圖 | 標題／主選單／創角 lifecycle 已有 | E2；逐畫面仍待對拍 |
| attract 職業巡禮 | 影片、`docs/67` TITH–TITO | pack 八卡輪播、輸入中斷、runtime 圖已接 | E2／V1；能力條逐幀仍待 V3 |
| 主選單 | DOSBox、網路圖 | 已有新遊戲／載入框架 | 需輸入與版面 E2 |
| 主角姓名／性別 | DOSBox 正式輸入、IDA、同狀態 PNG | 共用元件與正式 trace；`FIRST.SCR`、record 407 的 13 個具名確認欄位、三層 raw EGA backdrop 與 `beveled_2px` frame 均由 pack 接入 | E3；能力確認固定 checkpoint 已 V3 靜態（AE 1,474／640×350），游標閃爍、palette register、能力條與整段 timing 仍待動態 V3 |
| 家中／母親 | DOSBox、影片 | sec4+rec82/83；handler54 transaction 已接 | E3；逐格護送動畫仍待 V3 |
| 王城謁見 | 攻略、影片、地圖 | 正式 region gate／精確獎勵／一次性已接 | E3；原版畫面仍待 V3 |
| 酒場／登錄所 | 攻略、D3TXT、地圖、EXE handler | 正式入口與四人隊正常輸入 trace 已閉合（2026-07-28） | E3 |
| 四人縱列 | 影片多處 | pack 對映 + 8 步 trail + 死者隊尾 + runtime 對拍已接 | E2；需同狀態 V3 |
| 城鎮／洞窟 | 影片、全 CTY render、DOSBox、IDA `sub_1BD97` | 通用 loader/render 已有；CTY `+0x11` raw 遭遇 gate 與步數計數器已接 | E2（遭遇 gate）；仍需事件與 entrance closure、同狀態 V3 |
| NPC／日夜 | DOSBox、RE | 四階段 runtime、日／夜 NPC table 與 story-flag filter 已有正式入口；IDA handler74／69／70／33 的四條 transaction 已接線，原始 clock／palette bank 靜態鏈已閉合 | E2／D3；精確可見 palette transition 與完整 route 仍待 V3；不再把五個已接旗標誤列為 runtime unknown |
| 地表／HUD | 影片、網路圖 | 四欄 H/M/等級 HUD 已由 pack 幾何／glyph 驅動 | E2；需入口／palette／同狀態 V3 |
| 指令窗 | 影片、攻略操作說明 | 2×3 六指令已有 | 需所有子選單與 Enter 語意 |
| 道具／裝備／狀況／咒文 | 影片、EXE field caster/handler | 詳細狀況窗已依 `DGROUP 0x3DA8`／D3TXT00 record 407 資料化，runtime 圖達 E2／V2；魯拉、烈米特、特黑洛斯、拉那魯達的 MP／gate／核心效果為 D2 | 狀況子選單／隊員詳情、道具／裝備逐窗與其餘工具咒仍需 E2／V3 |
| 戰鬥 | DOSBox／影片／原始怪圖 | 公式、多敵、狀態、boss queue 已有 | 呈現／訊息／cue 長尾 |
| 商店／旅社／教會／達瑪 | 原版資料、部分截圖 | 達瑪轉職已由正式流程閉合至 E3/V2；其餘多數已有 | 賣出、逐服務與達瑪原版同狀態 V3 仍缺 |
| 船 | 影片、DOSBox 截圖 | 取船鏈、正式登船與首次航行已接 | E3；航海畫面仍待 V3 |
| 不死鳥／飛行 | 攻略、影片、EXE／CTY70 | 六珠祭壇、復活、搭乘、飛行與降落已有正式 trace | 新遊戲 boot trace 已通過 P4-P6；主線 E3，畫面 V1 |
| 下降／下世界 | 影片、EXE | 巴拉摩斯後王座事件與自然下降入口已接 | 新遊戲 boot trace 已通過；主線 E3，畫面 V1 |
| 終盤連戰 | 影片、RE | 光之珠、隱藏樓梯、歐里狄加與三連戰已有正式入口 | 新遊戲 boot trace 已通過；主線 E3，畫面 V1 |
| THE END | 影片、TIT3 | 戰後回城、冊封與 ending scroll 已接 | 新遊戲 boot trace 已到 THE END；逐畫面與音效仍待 V3 |
| 音樂／音效 | 原版 MCX/VOC、錄音研究 | OPL2/OGG 有相當基礎 | 逐場景 cue matrix 待驗 |
| 觸控／Android | 無原版對照，屬 port UX | 輸入抽象與觸控已有；保存優先的 Android UX 規格見 [`docs/124`](124-android-ux-spec.md) | Docker APK build、生命週期與真機 touch path 分開驗收 |

## 3. 現況判斷

底層並非從零開始：

- 資產 parser、地圖、文字、NPC、道具、戰鬥、音樂、存檔與觸控均有大量成果。
- 主角命名／性別、開場家中、日夜、NPC 三層可見性、船、取船鏈、boss trigger 基礎、
  多敵、狀態咒文等已比 `docs/data/flow-gap-ebitan.md` 初版盤點進步。
- R-2 沙曼歐莎的靜態 spec 已定。

真正問題曾經是：

1. ~~沒有一條 production input 驅動的完整主線 trace~~（2026-08-09 已由
   `TestOpeningProductionInputTrace` 閉合，不再是 current blocker）。
2. 歷史 `spine_test.go` 直接呼叫內部函式，不能證明玩家可達；它仍是線索／回歸材料，不取代
   下方的正式 trace。
3. ~~production 的 `T/R/U/Z/Enter/Cancel` demo/debug 行為~~（R-4/R-5a 已移除）。
4. 六珠／不死鳥／飛行、下降與終盤正式事件切片已完成，且已由同一條不改狀態的
   production `InputState` trace 從新遊戲延伸至 THE END；目前不再把孤立切片當作 campaign
   證明，後續只需補畫面／音效的 V3 對拍與發佈驗收。
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

尚未完成的是能力確認畫面 stat panel 逐像素對拍與母親逐格移動動畫；不能因 transaction
與 opening trace 通過就把整個 P1（酒場、四人隊、正常出城）標成完成。

## 4. 證據與驗收等級

每個 GAP 使用以下狀態：

- **E0 — 線索**：攻略、影片目視或合理推測。
- **E1 — 原版證實**：原始資料、EXE RE 或 DOSBox 同操作證實。
- **E2 — runtime 接線 parity**：對明確命名的資料／狀態交易，Ebiten 正式 runtime
  已接到相同 writer、gate、consumer 與副作用。畫面證據另以 V 等級標示；E2 本身
  不表示逐像素或玩家可見 parity。
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
  CTY60／61 階段的原版同狀態 V2；CTY59 handler41 的
  Y=4 設施分支已由 IDA／原始文字與正式 `cmdTalk` 測試閉合為 D3／E2／V1，見 `docs/102`。
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

### P8 — Release（本輪指定桌面格式已產出）

- 無 debug full playthrough，分段使用正式 save checkpoint。
- 使用者實玩驗收。
- Linux／Windows／macOS desktop packages。
- Android 已另開保存優先的 port UX 工作；本輪若只有 Docker APK/AAB build，狀態仍為
  `build-only`，不得由桌面包推論真機 touch、lifecycle、音訊或存檔通過。WASM 仍不在本輪。
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
2. ~~閉合 `file 0x13a0` 開場 owner 及其直接 callees 的必要資料流，建立 opening state ledger。~~
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
`(152,238,352,96)`、文字 inset `(16,16)`、20 欄／4 行；`sub_1fb36` 的較大範圍只是
關閉時背景備份，schema 已將可見 rect 移入
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

### 2026-08-09 歷史收尾狀態（已由本檔後段現行 checkpoint 取代）

以下只保存當時 checkpoint 的驗收史；不得用其中「已完成／目前」字樣覆蓋 2026-08-22
現行 replay 狀態。

1. **主線事件串接（已完成）**：boot 起的同一條 trace 已正式完成商人交付、CTY41→42、CTY43
   關卡、拉之鏡、假王／怪力魔、CTY54 變身杖交換、船員之骨定位、幽靈船動態物件／入口，
   並由 CTY36 原始轉場取得愛的回憶、離船後恢復同座標停泊船、通過 `(76,54)` 奧莉薇亞
   海岬，再進 CTY55 取得蓋亞之劍並完成 save/load。IDA 已確認 raw item0x0f 經特殊 item
   table 派發至
   `(logical) 0x3d65`，唯一成功座標 `(65,109)`，成功 set world-state bit0x20 並由
   `DGROUP 0x3b96` 套用 direct viewport patch；`DGROUP 0x3a54` 是 world-state reload 的
   distinct persistent patch。`docs/106` 已完成 JSON／Go 遷移、IDA parity 與 direct/reload
   分離；正式 trace 已由 CTY56 南洞口穿過 sec1→2→3→2→3→sec0 北出口，進 CTY64
   以 handler49 取得銀寶珠，並通過 party-order inventory 與 save/load。其後同一條正式 trace
   已完成 CTY83 黃寶珠、蓋亞之劍交換、索瑪城奧爾特加事件、封咒／三連戰、光之珠弱化、
   回王城冊封與 `THE END`。
2. **設定資料追蹤**：每個 blocker 都先由 IDA／原始指令追
   `writer → table/state → consumer → visible effect`，並以 DOSBox 同狀態核對；不得用
   C remake、攻略或推測值直接補 production config。
3. **完整通關證明（已完成）**：`TestOpeningProductionInputTrace` 已在 Docker＋Xvfb
   由新遊戲以正式 `InputState` 不改狀態延伸至 `THE END`，並涵蓋 P4-P6 關鍵事件、戰鬥、
   資源交易及中途 save/load round-trip；本項達 campaign E3。終盤固定 `TIT3.P` 片尾已
   透過同一條正式 trace 產生 `dq3_remake_ebitan/docs/img/ending_the_end_runtime.png`，
   並與 `docs/title/ending.png` 逐像素一致；結局 cue=17 也由 `audio.json` 接到王城冊封
   後的正式時序。這只把終盤畫面／cue 升為 V3，不將其餘流程誤稱為逐畫面 V3。
4. **戰鬥文字串接（本輪已完成）／戰鬥長尾**：`interface.json.battle_texts` 已將 25 個
   battle role 對映到 `texts.json` 的 D3TXT00 原始 glyph/control stream；名稱、數字插值、
   睡眠／醒來、攻擊／傷害／沒打中、勝利／經驗／金錢、狀態／解毒、帕魯朋特及全滅均由
   pack 提供，缺資料時 production 啟動 fail closed。訊息階段另已改用 `battle_message`
   的原版 `(152,238,352,96)`／20×4 外框，隱藏左下指令框與右側空敵名框；指令／敵名
   panel 的 raw rect 亦已接入 pack，詳見 `docs/109`。Docker＋Xvfb
   的 `internal/...` 全套與 `game`（長 campaign trace 另行重播）其餘測試通過，受影響戰鬥
   PNG 已重產；欄位與證據見 `docs/108`。
   README 戰鬥 PNG 另已關閉非原版的 `combat_info` 診斷飄字後重產；仍需 RE／同狀態
   V3 的逐動作停頓、動畫／音效 cue、其他抗性與狀態、敵群 formation、當時未知的 Boss 多次行動、
   掉落及逐項咒文效果，不能把文字層完成誤稱為整個戰鬥完成。
5. **玩家可見 parity**：四人縱列與 HUD 已完成第一個 E2 runtime slice（`party_field_hud.png`；
   原版影片 `f000295`／`f000300`／`f000900` 作 oracle），仍需相同地圖／位置／晝夜的 V3；能力確認 stat panel、母親逐格演出、
   attract 的能力條逐幀／淡入淡出、
   商店／教會／旅社／達瑪、二選一小視窗原版幾何、日夜 palette、剩餘 BGM／SFX 及
   ending timing 尚未全部 V3；固定片尾 `TIT3.P` 與 ENDING cue 已完成 V3，但結局文字
   停頓／轉場 timing 仍待影片同狀態核對。二選一小視窗現行 geometry 只有 legacy renderer，尚無 D3
   結構證據，不得直接寫入 game-pack canonical JSON。
   教會復活確認的現行 runtime 證據為 `docs/img/church_revive_confirm.png`，仍待 DOSBox
   同狀態 V3。
6. **Release（本輪已完成指定桌面包）**：使用 Docker＋osxcross／既有 Ebitengine image
   產出 AppImage、Windows x86_64 ZIP、macOS Intel ZIP 與 Apple Silicon ZIP；發佈包不納入
   原版資產。Android、WASM 與真機觸控／lifecycle／音訊／存檔不屬於本輪三平台目標，未以
   未完成的產物宣稱支援。工作樹內受保護的空 `dq3_remake_ebitan/tmp_dump.go` 未刪除，
   build smoke 以不含該 scratch 的臨時副本執行；包裝與雜湊見 [`docs/114`](114-release-artifacts.md)。

後續工作以 V3 畫面／音效對拍、戰鬥與場景 cue 長尾為主；若重新開啟主線，
仍須從上述合法 production checkpoint 重播並記錄第一個實際 blocker，不得改用孤立 handler
或狀態注入取代玩家流程。

2026-08-09 另完成戰鬥指令標籤的資料化：`interface.json.battle_command_labels` 保存
`war`、`flee`、`defend`、`item`、`spell` 五組原版雙 glyph，`battle.go` 不再持有
這組玩家可見字模 table。這是 D2 engine/data 邊界切片；逐 label 的 IDA writer sidecar、
框線樣式、逐動作 timing／動畫／音效、抗性／formation／掉落與跨平台 release 仍未閉合。

2026-08-09 同批再完成地表命令窗字模資料化：`interface.json.field_command_labels`
保存 `title` 與 `talk`／`spell`／`status`／`item`／`equip`／`examine` 七組雙 glyph，
`cmdmenu.go` 不再持有版本專屬 raw glyph table；缺少欄位時 `NewGameWithPack` fail closed。
證據與 D2 限制見 [`docs/110`](110-field-command-labels-re.md)。這仍不等同戰鬥
逐動作 timing／動畫／音效、場景 cue 或跨平台 release 已完成。

2026-08-09 追加戰鬥場景資料化切片：`interface.json.battle_scene` 保存原版場景帶
`80..246`、怪物基線 `232` 與 `D3TXT00.FON` 游標 glyph `0x77`；`battle.go` 移除這組
DQ3 專屬常數，production 缺欄位即 fail closed。證據與 D2 限制見 [`docs/111`](111-battle-scene-layout-re.md)；框線
style pattern、spell／target baseline、逐動作 timing／動畫／音效、完整抗性／formation／
掉落與跨平台 release 仍是後續 V3 GAP。

2026-08-09 追加開場／創角 glyph 資料化切片：`interface.json.new_game_labels` 保存主選單、
性別、能力確認及 D3TXT00 records 451–456 的完整姓名畫面 glyph stream；`NewGameFlow`、
`NameInput`、`GenderSelect` 與酒館改由 pack 讀取，缺欄位時 `NewGameWithPack` fail closed。
姓名盤已恢復原版 45 raw cells、欄優先 cell 換算、raw35 功能列入口與第五列完成 gate，
不再把 38 格英數盤的直接 OK 當作正式流程。`D3TXT00.FON` 大小／SHA-256、role 對映、
D2 限制與 parity test 見 [`docs/112`](112-newgame-labels-re.md)。
同日再完成開場／創角 geometry slice：`interface.json.new_game_geometry` 保存 IDA
linear raw window（姓名 `0x29088`、功能列 `0x290a6`、注音／候選 `0x290c4/0x290e0`、
能力確認 `0x28b78/0x28b92/0x28bc6`）、9×5／16px 姓名盤步距及 640×350 像素 rect；
`NewGameWithPack` 缺欄位即 fail closed，renderer 不再把大黑框疊在標題城堡背景上，創角
階段改用原版 `FIRST.SCR` 黑底雙立繪。IDA writer／consumer、截圖量測、D2 限制與 parity
test 見 [`docs/113`](113-newgame-geometry-re.md)。本輪已以 Docker／Xvfb 正式 dump 更新
`docs/ng_menu.png`、`docs/ng_name.png`、`docs/ng_gender.png`、`docs/ng_confirm.png`；其中
`ng_name.png` SHA-256 為 `7d489f8f13d9eb1cb4e246efc9cef4c82a84f957720bb31ba912294dc9dcc39a`。
功能列字模、45 格輸入與正式功能焦點路徑已接入；palette、藍色選項框、逐幀游標／動畫與同狀態演出
仍不能宣稱 V3。

2026-08-10 再閉合開場外框色彩切片：IDA Pro 9.4 追查
`sub_10854 → sub_1f590 → sub_1fd30 → sub_1fdb1`，確認四邊 writer 與
`DS:0x727` 的 raw plane mask；DOSBox `v3_01_afterstats.png` 取樣外框為
`(255,223,255)`、內部黑。`interface.json.new_game_geometry.frame` 現以
`border_pattern=solid_2px` 資料化，`NewGameFlow`／酒館消費 pack frame，缺欄位或
不支援 pattern fail closed；這是 D2／runtime V2，palette、藍黑 `confirm_choice` pattern
與逐幀 EGA 演出仍保留為 V3 gap。完整 sidecar 見
[`docs/117`](117-newgame-frame-re.md)。

下一個視覺／音效 parity 工作：

1. 以 `confirm_choice` 已接的藍黑 backdrop／中央黑 content／lavender-膚色 2px frame 為基線，
   補 palette register、逐幀游標／能力條動畫的 writer／consumer evidence，並維持現有
   raw window／runtime hash 可回查。
2. 再處理戰鬥 frame style、spell／target rect、逐動作 timing／動畫／SFX 與完整抗性／
   formation／當時未知的 Boss 多次行動／掉落，所有設定先進 pack JSON 並補 D2/D3 parity。
   Boss 項目現已由 `docs/148` 證實不存在，不再執行。
3. 完成場景／設施／日夜 palette、剩餘 BGM／SFX、結局 timing；若另開 Android／WASM，
   必須從新的工作範圍與容器驗收開始。每批仍由合法 production checkpoint 重播，不用
   debug shortcut。

2026-08-10 遭遇 gate 勘誤與接線：IDA Pro 9.4 重查 `sub_1BD97`（IDA linear
`0x1bd97..0x1bdde`；`dec [0x52f4]` 的 file offset `0xd118`）確認 CTY 模式下
`cmp [0xd77],0; jz return`，因此 section header `+0x11` **0=安全、非0=可遇敵**。
原始 writer `sub_130cf`（file `0x443f` 起）把 `ah=[di+1]` 寫入 `[0xd77]`；完整輸入雜湊、
位址基準、Go 對映與保留的不確定性見 [`docs/115`](115-encounter-cty-gate-re.md)。
`Town.EncounterFlag`／`Scene.encounterFlag` 已保存 raw，正式地表／遭遇 CTY 移動改用
`DS:[0x52f4]` 對映的計數器；安全 CTY 與飛行不推進，save JSON 保存計數器。這關閉的是
遭遇 gate 的 E2／資料對映缺口，不把強制遭遇、完整 encounter pack JSON、戰鬥動畫／音效
或同狀態 V3 誤標為完成；本輪只跑受影響的針對性測試，未重跑 311 秒完整 trace。

2026-08-10 戰鬥 frame 契約勘誤與接線：IDA Pro 9.4 追完
`sub_1f590 → sub_1fd30 → sub_1fdb1`，確認 `0x031a`、`0x0b11`、`0x0912` 是
`win_rect` 的 raw flags／背景備份槽，不是三種 frame style。`sub_1fd30` 依共用
EGA writer 畫四邊；與原版 `dosbox/orochi_boss.png` 像素核對後，可見結果保存為
`interface.json` 各 battle window 的 `frame`（白色 1px border、黑色 interior，D2
pack evidence；推論說明見 [`docs/109`](109-battle-hud-rects-re.md)）。
`fillPackBox` 已讓戰鬥訊息、指令、敵名及隊伍狀態框消費 JSON，
`NewGameWithPack` 缺戰鬥 frame 即 fail closed；直接 fixture 才保留導覽用 fallback。
`internal/gamepack` 的 EXE/HUD parity、`game` 編譯、`TestBattleCommand`、
`TestMultiEnemy`／`TestHit` 均在 Docker 通過；本輪未重播 311 秒完整 trace，戰鬥逐動作
timing／動畫／音效、抗性／formation／掉落仍是 V3 長尾。

2026-08-10 敵人數量列基線接線：`sub_1b101` 的名稱／數量 writer 與
`dosbox/orochi_boss.png` 的玩家可見像素均閉合為同一個 16px row；
`interface.json.battle_enemy.count_inset_y=0` 明確表示相對名稱列的垂直偏移，
`battle.go` 不再把這項關係藏在共用 renderer。這筆證據是 strong+D2，raw numeric
writer 的 DX 內部轉換仍未升為 D3；spell／target rect、逐動作 timing／動畫／音效、
抗性／formation／掉落仍是戰鬥 V3 長尾。Docker＋Xvfb 的
`dq3_remake_ebitan/docs/battle_count_row_runtime.png`（SHA-256
`a9c49e091d25bc28d018cf5363c07f75dac4a8029a91cf78f9da60b179518eb8`）只作 debug 視覺
核對，不取代正式玩家輸入。

2026-08-10 詳細狀況窗切片接線：IDA Pro 9.4 以 `lea si, ds:3DA8h`→`sub_1F4E3` 閉合
原版 `DGROUP 0x3DA8`（file `0x19EE8`）的 352×192 視窗結構；flags、位置、尺寸與
D3TXT00 record 407 的 22×12 glyph/control stream 已寫入 `interface.json.field_status`
與 `texts.json`。Go renderer 只消費 pack 的 frame、標籤 glyph、性別／職業 glyph 及
具名 anchor，動態填入等級、HP／MP、五項能力、攻擊、防禦與經驗；缺契約、frame 或文字
時 `NewGameWithPack` fail closed。Docker＋Xvfb 的正式新遊戲 fixture 產出
`dq3_remake_ebitan/docs/status_detail_runtime.png`，狀態面板達 E2／V2；完整入口、子選單、
隊員詳情、道具／裝備逐窗與同狀態原版 V3 對拍仍未閉合。raw bytes、位址基準、推論等級與
限制見 [`docs/116`](116-field-status-panel-re.md)，欄位契約見 [`docs/84`](84-game-pack-json-contract.md)。

2026-08-10 開場能力確認欄位勘誤與資料接線：由 `docs/dosbox/v3_01_afterstats.png` 與
D3TXT00 record 407 的 glyph code 逐列解碼，確認右欄不是舊 renderer 使用的「速度／HP／MP」，
而是「運氣點數／最大HP／最大MP／攻擊力／守備力／經驗」。`interface.json.new_game_labels`
新增 `luck`、`max_hp`、`max_mp`；`NewGameFlow` 依正式 pack 讀取並將 `stats.LUCK`、最大 HP、
最大 MP 寫入對應列，`agility` 僅保留給詳細狀況窗。這修正設定資料 mismatch，證據仍為 D2、
runtime 仍為 V2；原版 EGA pattern、藍色選項框與同狀態逐像素對拍另列 V3，不宣稱整個開場
畫面已完成。

2026-08-10 開場 `confirm_choice` 靜態設定資料勘誤與 runtime 接線：同一正式 DOSBox
姓名／性別輸入量得 raw window `linear 0x28bc6` 的藍黑 checkerboard
`x=360..471,y=62..125`（`RGB(0,85,223)`、local odd phase），中央黑色內容
`x=376..455,y=78..109`，以及 lavender／膚色交錯的 `x=367..464,y=68..117` 雙色
2px frame。`interface.json.new_game_geometry` 新增 `confirm_choice_backdrop`、
`confirm_choice_content`、`confirm_choice_frame`；`FrameStyle` 的具名
`checkerboard_frame_2px` 僅接受已註冊 primitive，缺 accent 或未知 pattern 即 fail closed。
右欄六列 anchor 同批校正為 `y=126,142,158,174,190,206`。Docker＋Xvfb 針對性 tests、
PNG dump 與純函式 frame phase test 通過；runtime `docs/ng_confirm.png` 已更新。這是
D2／V2 的穩定畫面切片，palette register、游標閃爍／能力條淡入與逐幀演出仍列 V3，
不把一張 PNG 擴大解讀為完整創角 parity。raw／像素與歷史勘誤見 [`docs/118`](118-newgame-choice-backdrop-re.md)
及 [`docs/113`](113-newgame-geometry-re.md)。

2026-08-10 日夜第一性原理 audit（歷史快照；已由 2026-08-22 checkpoint 訂正）：當時 engine 有 `dnPhaseSteps=60` 的四階段 clock、
`DarkenPalette(worldPal, phase)`、住宿／黑暗之燈 reset、日／夜 NPC table 及 story flag
filter，正式 component／開場 trace 也已涵蓋日夜入口；因此不重寫既有機制。下一輪只追
原版 raw clock／palette register 的同狀態對拍，並對 `docs/71` 列出的 21 個條件 flag 逐一
閉合入口、writer、pack state、consumer 與玩家可見結果；已有動態 pack flag 不重複造表，
缺 caller／writer／consumer 的項目維持 `unknown`／fail closed。日夜 row 的原始判定更新為
「機制 E3、精確 palette／逐事件 flags audit 待 V3」，不是把機制缺口誤記為尚未實作。

2026-08-10 開機年代／巨龍過場正式接線：原先桌面／mobile `NewGame` 後直接進標題，與原版
第一幀不符；`interface.json.opening` 現保存 `TITA.P`、`TITB.P`、`TITD.P`、`TITE.P`、
`TITP.P` 五張 manifest-完整 PCX。`StartOpeningCutscene` 已接到桌面／mobile 正式入口，
無輸入按 pack 幀推進，任何正式輸入停止並把同一按鍵交回標題；`TestOpeningCutsceneBootAndSkip`、
`TestDQ3OpeningCutsceneContract`、Docker＋Xvfb 的 `TestDumpOpeningCutscene` 均通過。這關閉
「開機直接跳標題」的 E2／V1 缺口；素材 identity／靜態 runtime 為 D2／E2，120 幀只是可重播
設定，排序、TITP 位置、淡入淡出、開場音效與逐幀原版 timing 仍是 V3 長尾，完整邊界見
[`docs/120`](120-opening-cutscene-re.md)。

2026-08-10 戰鬥／日夜靜態 RE 閉環補充：IDA Pro 9.4 對同一份 `DQ3.EXE` 完成正式戰鬥
`sub_1BDDF → sub_1C08B → sub_1C34F → sub_1A973` 的 caller／queue／actor chain，確認
玩家／敵人速度排序、死亡／睡眠／逃跑／咒文／物理分派；`sub_1C425` 也確認經驗／金錢
分配、`DGROUP 0x2518 bit1` 掉落抑制，以及 D3MNS `+0x25/+0x26` 的掉落 gate／item raw id。
另在 `loc_190D7`（file `0xa447`）確認事件／AI-only runner 最多重建 12 輪，但每輪每個
queue entry 仍只消費一次；這不構成 boss 同回合多次行動證據。
`sub_1D7FB` 已閉合 D3MNS `+0x19..+0x1d` packed resistance 與 `0/68/180/255` 門檻；
`sub_1A7D5`／`sub_1BF35` 已閉合候選八筆、38 點 encounter budget、group/background/page
及 active position 的 formation 資料流。`sub_20770`／`sub_208A7`／`sub_208E2` 則把
戰鬥 `BP` cue、`DS:0x253e` 指標表、`FVOC.VCX`／`NVOC.VCX` 與 Sound Blaster driver
等待點接起；`sub_1C673`／`sub_1C642`／`sub_1B3C3` 的 tick／重畫 timing 也已保存。
以上是 E1／D2 靜態證據，並不等於 cue 波形、逐格動畫或同狀態畫面／聲音 V3。

同一輪完成日夜精確值：`sub_1EE23`（file `0x10193`）令 `DGROUP 0x251d` 在有效
overworld tick `0..0xef` 遞增，`0x78`（120）寫 `0x526c=0`，`0xf0`（240）歸零並
寫 `0x526c=1`；`sub_1EE76`（file `0x101e6`）依 `clock/0x14` 查 `DGROUP 0x25c5`
的原始 12-byte palette index `01 00 00 00 01 02 03 04 04 04 03 02`，由 `DQ3.PAL`
buffer `DS:0x3232` 選 bank 並上傳 DAC。`sub_1ECDC` 載入 `DQ3.PAL`／`MNSBK.PAL`，
`sub_131BD` 以同一 `[0x526c]` 選日／夜 NPC 表，再測 `[0x4f70]` story bit。
21 個條件 flag 的原版 SET literal→`sub_16EDF`→`[0x4f70]` ledger 已補至
[`docs/123`](123-static-battle-daynight-re.md)；原版 writer 全部 confirmed，但
`0x0d`、`0x11`、`0x24`、`0x1a`、`0x1b` 的 runtime 事件／consumer 仍 unknown，不能把
「原版有 writer」誤寫成 remake 已接。

本輪因此只關閉當時的靜態設定 mismatch；當時未把 Boss 同回合多次行動升格（正式鏈只證實
每個活躍 actor 每回合一次）；該負證據後來由 `docs/148` 完整閉合為不存在。此段也未把
當時的 `dnPhaseSteps=60`／`DarkenPalette` 近似改成
原版二值 selector。`docs/123` 與 [`docs/data/dq3-static-battle-daynight-evidence.json`](data/dq3-static-battle-daynight-evidence.json)
已保存帶來源 hash／位址基準／推論等級的非 production sidecar；剩餘工作是將達 D2／D3 gate
的欄位逐批遷入 versioned game-pack、補每個 formation／resistance descriptor、VOC cue
波形與同狀態 DOSBox V3、逐動作動畫／停頓，以及缺失 21 flag 的正式玩家入口；未知值保持
fail closed。

2026-08-11 靜態勘誤：以 `sub_1D7FB` 的原始 `cmp`／`inc BX` 重算 packed resistance
category，修正 Go `SpellChance` 的 first-`>=` 邊界誤讀與 `0x00` sentinel 遺失；
`spell=2/38/59`、怪物 121、越界 fail-closed component tests 在 Docker 通過。完整
effect 名稱／玩家可見抗性仍需逐項 descriptor／V3，未知資料不進 production pack。

2026-08-11 靜態戰鬥／flag deep dive：本輪只補證據，不宣稱新增 runtime 完成度。formation
候選表已按原始 `dec region; shl 5` 修正為區域 `1..27`、108 列完整 raw sidecar，13 個
固定編隊 record 也保留 count／未命名 header／pair bytes；`sub_1AB2C`→active `+0x03`→
`sub_1B1FE` 的 position writer／consumer 已達 strong，但螢幕座標仍 unknown。`DS:0x37c3`
的 60 筆 resistance descriptor 與 D3TXT00 `0x79..0xb4` 中文 record 對映已 confirmed，
不把名稱推成元素抗性。handler74／69／70／33 分別閉合 `0x0d`、`0x11`、`0x1a/0x1b`、
`0x24` 的原版事件副作用，CTY full-byte NPC records 也證實 `sub_131BD` generic consumer；
Go production writer／直接 GET caller／玩家可見事件在當時仍 unknown。當時 Boss repeat-N、
逐動作 SHP frame、PCM 波形與停頓維持 unknown／V3；Boss 項目後來由 `docs/148` 證實不存在，
其餘未知欄位不得進 game-pack。

2026-08-11 攻略×IDA 旗標身份勘誤：`0x1a/0x1b` 的現行身份是 CTY65 巴拉摩斯戰後
世界／NPC 可見性交易（handler70 `sub_1622A`），不是龍之女王／彩虹材料中繼；龍之女王
光之珠在 CTY67 handler52。`0x0d` 對應索瑪後期 CTY80、`0x11` 對應六珠祭壇 CTY70、
`0x24` 對應沙曼歐莎條件居民／handler33。這些是原版 writer／generic consumer 的
`confirmed` 路線身份，Go 正式入口仍 `unknown`，詳見 [`docs/119`](119-daynight-story-flag-audit.md)
與 [`docs/123`](123-static-battle-daynight-re.md)；不得把攻略路線直接當成 runtime 已接。

## 2026-08-11 E2 快速完成標準（使用者確認）

本 milestone 採用 E2／D2 的靜態與資料驅動交付標準，不以重新啟動 DOSBox 作為驗收前置。
驗收 oracle 限定為 IDA Pro 9.4 的非破壞性 listing／交叉參照、原始 EXE／DAT／SHP／VOC／CTY
bytes、可重建 decoder、Docker 內 deterministic Go trace、schema/reference/parity tests
及既有 reference PNG。戰鬥、事件、formation、抗性、掉落、SHP redraw、VOC cue 與 save/load
只納入 `confirmed` 或必要的 `strong` 證據；未命名 raw 欄位與無法閉合的玩家語意一律保留
`unknown`、raw sidecar 與 fail-closed，不用合理預設或 C remake 補值。

本標準在當時明確排除本 milestone 的 V3 承諾：當時未知的 Boss repeat-N、精確螢幕座標、逐動作
SHP frame、實際 PCM 播放波形／玩家停頓及五個缺失旗標的完整玩家可見語意。這些項目仍記錄
在證據 ledger，但不阻塞 E2 的 JSON／引擎／headless 驗收，也不得被 README 或 release
描述成已完成的原版 parity。

E2 的最小充分切片順序為：

1. 將已確認的戰鬥／formation／resistance／SHP／VOC／flag raw 契約遷入 versioned game-pack，
   補 schema、reference validation 與原始 bytes parity test。
2. 以共用引擎的 deterministic `InputState` trace 驗證 queue、掉落、抗性、formation、事件
   交易與 save/load；不以 debug shortcut 或直接 handler 呼叫代替正式入口。
3. 以 decoder 輸出及既有 PNG 做 headless contract 驗收；完成後只宣稱 E2／D2，不宣稱 V3
   畫面／聲音 parity。

### E2 第一個垂直切片（2026-08-11）

- 修正 `internal/dq3data.EncounterTables.Slot`：原版 `dec region; shl 5` 對應 raw region
  `1..27` 到 Go slot `0..26`，不再錯讀下一區；region `0` 與越界值 fail-closed。
- 新增 `internal/gamepack/packs/dq3_cht/data/battle.json`，保存 27×4 候選 raw table、13
  筆固定編隊、resistance threshold／class／60 筆 descriptor 與中文 record 對映。
- `NewGameWithPack` 現從 `data.battle` 的驗證 raw bytes 建立 encounter decoder；不再從
  `DQ3.EXE` 直接讀取版本專屬遭遇表。`gamepack` loader 檢查 source hash、raw length、
  descriptor cardinality 與固定 stride，未知資料不提供 fallback。
- `D3MNS.DAT`、`DQ3MNS.SHP`、`PACKBG.SCR`、`MNSBK.PAL`、`FVOC.VCX` 的 battle asset
  路徑／完整性也移入 manifest；SHP 同時記錄原始 bytes 與已產生 128/129 版本的受控
  alternate hash，兩者都必須精確匹配，不能接受任意檔案。
- Docker `go test ./internal/gamepack ./internal/dq3data` 與 Xvfb headless `go test ./game -run ^$`
  通過；這只證明 E2 loader／編譯契約，不代表 V3 畫面／音效 parity。

### E2 headless trace 驗收（2026-08-11）

未啟動 DOSBox；使用專案既有 `Dockerfile.test` 建立一次性工具鏈，並以唯讀 Go module cache、
UID/GID 對映及 `xvfb-run` 執行。下列結果是同一個 `assets_raw`／pack 的可重播驗收：

- `go test ./internal/... -count=1`：`battle`、`dq3data`、`gamepack`、音訊、物品、RNG、
  spell、stats、注音等套件全綠。
- `TestE2BattlePackEncounterTrace`：正式 `NewGame` 載入 `data/battle.json`，由 pack raw
  建立 region／candidate decoder，阿里阿罕 raw region 1 與 `[5,6,5,10]` 候選一致。
- `TestBaramosProductionInputTrace`、`TestMirrorProductionInputTrace`：事件入口、戰鬥命令、
  回合收集、勝利交易與後續對話均由正式 `InputState` 推進。
- `TestMirrorWinSaveRestore`、`TestSaveRoundTrip`、`TestSaveRecordsAndChecksGamePackIdentity`、
  `TestSavePreservesRememberedOverworldPosition`：事件完成後 save/load 與 pack identity 檢查通過。
- `TestBattle*`、`TestShanpaneOriginalMixedFormation`、`TestEnemyTurnEachAliveActs`、
  `TestGroupBattleAllMustDie`：速度 queue、逐 actor 行動、混合編隊、全滅與獎勵通過。

因此本輪可標記為「戰鬥資料／loader／headless 交易 E2、formation raw E2、畫面 V2（既有 PNG）」；
不提升為 V3。當時 Boss 同回合 repeat-N、逐動作 SHP frame、PCM 波形／停頓、五個 story flag
的完整 remake 玩家入口列為 `unknown`／後續工作；Boss 項目已由 `docs/148` 證實不存在，其餘
狀態依後續 checkpoint；formation 的 raw EGA projection 雖已接入
pack／renderer，仍不能把它包裝成同狀態 V3 截圖 parity 或 release 的完整戰鬥宣告。

補充：本輪重播時發現 CTY07／08 塔區的 section route fixture 會遇到原始遭遇表；已改由
雷貝正式道具店購買聖水、以道具選單驅敵，並在 route helper 以正式戰鬥輸入處理意外遭遇。
原始 CTY07→CTY08 transition chain 現可抵達 CTY08 sec3 並接上拿吉米老人；完整
`TestOpeningProductionInputTrace` 舊版後段曾在 CTY13 路徑 fixture 停止；2026-08-11 已確認
CTY13 sec2 的 raw encounter gate（`+0x11=10`）會開正式 battle modal，並在
`traceWalkToNoPortal` 補上 production battle／聖水輸入。從合法前段 trace 已抵達 CTY13
sec2、依序完成雙開關、取得 `0x56` 魔法鑰匙並通過 save/load（targeted E2）。這個修正不是
以 debug state 蓋掉；完整 campaign 在修正後尚未重播完畢，故仍不宣稱 E3。V3 畫面／音效
不足之處依 README 的 `V3-approx` 政策標示，不能改寫成逐像素或逐波形 parity。

### E2 formation／PCM 補證（2026-08-11）

IDA Pro 9.4 補出 `sub_1AAA1 → sub_1AAD5 → sub_1AB2C → sub_1B31A` 的完整 raw
writer／consumer：`0x26 - Σ(D3MNS +0x28 × count) + 2` 是第一筆 active raw position，
之後每隻以 `2 × D3MNS +0x28` 前進；`sub_1AB2C` 也逐個體擲 HP。這些常數已資料化為
`battle.json.formation_position`，並由 production renderer 依 EGA `0x54` stride／`0x102`
bottom 計算位置。VOC decoder 同步保留每 cue 的原始 sample rate、sample count 與可重建
duration；該歷史 checkpoint 當時把 host `PlaySFX` wall-clock 列為 unknown，現已依本檔
頂端停止線改採平台規格近似；action frame 仍只由新原版 oracle 驅動。完整位址、hash、推論等級與
`NVOC` cue21 `0x1f973` 勘誤見 [`docs/123`](123-static-battle-daynight-re.md) 與 sidecar。

## 2026-08-11 歷史 checkpoint：IDA story-flag runtime E2 已接線

前文「五個 flag 的 production writer／玩家入口仍 unknown」是接線前歷史紀錄；現行工作狀態
以本節、`docs/119` 的 E2 勘誤及 `events.json` schema `0.1.29` 為準。四條已由 IDA handler
與 CTY selector 閉合的固定 transaction 已接到共用 engine＋DQ3 pack：

1. handler74／CTY80 國王冊封：寫 `0x0d`、清 `0xf5`、重載場景。
2. handler69／CTY70 六珠祭壇：依 pack 動畫完成後寫 transient `0x11`、map cell `0x80`、
   停放拉米亞，並在重載前清 transient；save/load 重放 map patch。
3. handler70／CTY65 巴拉摩斯：固定編隊勝利後清 `0x29`、同時寫 `0x1a/0x1b`、重設
   day phase／step 並套 palette。
4. handler33／CTY44 `(13,27)`：正式命令入口開 `DI=0x0bff`（record 71），再清 `0x3d`、
   寫 `0x24/0x21`、重建條件 NPC。

`game/story_flags_test.go` 的 pack contract、正式 command trace、Phoenix map/save、Baramos
aftermath 與 ending writer tests 在 Docker＋Xvfb 通過；這是 E2 engine/data／交易閉環，不是
同狀態 DOSBox V3。當時 Boss repeat-N、逐動作 SHP frame、PCM wall-clock、精確 palette register
與完整 campaign E3 依 `docs/123` 保持 `unknown`／V3 gap；Boss 項目已由 `docs/148` 證實不存在，
其餘不應重新開啟已完成的 JSON／flag
切片。

## 2026-08-11 歷史正式重播勘誤：該 checkpoint campaign E3 通過

上一段關於「完整 campaign E3 仍 unknown」是接線後、重播前的暫時結論，已由更強的
production replay 推翻。下層拉米亞 path 的 helper 曾把 active battle 保留的 movement cooldown
當作普通 cooldown；修正為先用正式戰鬥輸入完成遭遇後，
`TestOpeningProductionInputTrace` 在 Docker＋Xvfb 由新遊戲重播到 `THE END` 通過（63.92 秒），
含巴拉摩斯、P4–P6、索瑪三連戰、冊封與關鍵 save/load。完整 `game` 測試因 Ebiten 圖形資源
累積而改為同一二進位的 18 個一次性批次；288 個頂層測試全數通過。

因此該 checkpoint 應標示為 **主線 campaign E3／runtime V1**。這個訂正不改變 V3 缺口：逐動作畫面、
PCM timing與精確 palette 仍是獨立證據／polish 工作；本段當時並列的 Boss repeat-N 已由
`docs/148` 證實在本 EXE 不存在。完整驗收與
前述暫時 blocker 的原因、訂正、PNG 雜湊見 [`docs/125`](125-acceptance-20260811.md)。
現行工作樹是否達 campaign E3，必須讀本檔最新 checkpoint，不得沿用本段歷史結果。

## 2026-08-12 歷史 checkpoint：能力確認 V3 靜態／原版靜態研究收斂

本節是 `docs/74` 的現行狀態表，優先於上述較早的 V2／unknown 文字：

1. 創角能力確認的正式固定輸入已與原版 `640×350` 同狀態對拍，AE=1,474（0.658%）。
   資料、原始定位、殘差與範圍見 [`docs/126`](126-newgame-confirmation-v3-static-comparison.md)。
   它是此單一 checkpoint 的 V3 靜態，不是整段開場或全遊戲 V3。
2. IDA 9.4 已重新審計戰鬥 queue／alternate runner、SHP redraw、VOC dispatch、formation、
   resistance、日夜與五個 story flag。可由靜態原始資料決定的 pack／runtime 欄位已閉合；
   當時 Boss repeat-N、動作 frame、PCM wall-clock 與可見 palette transition 沒有可安全填入的
   production 值；Boss 項目已由 `docs/148` 證實不存在，其餘保留 `unknown`／fail-closed。收斂與停止條件見
   [`docs/123`](123-static-battle-daynight-re.md)。
3. `dq3_cht` schema 已升為 `0.1.31`、content version `0.1.36`。新增資料契約只擴充共用
   frame／backdrop primitive；所有 DQ3 glyph、raw window、geometry 與 layer 值仍在 pack，
   沒有新增版本專屬 Go 常數。

下一個 V3 工作只能由新的可重播原版 frame／音訊 trace 驅動；在沒有該輸入時，不以更長的
靜態反組譯或視覺猜測重新開啟已收斂的研究。

## 2026-08-12 修正前 checkpoint：日邦格八頭大蛇四人隊 V2 勘驗（歷史）

本節保存修正前的可重播對照與 blocker 成因；現行結論以緊接的「固定編隊背景 selector V2
閉合」章節為準。本輪取得一條新且可重播的原版／remake 對照輸入：同源 YouTube 本機畫格明確顯示第一戰勝利
仍在洞窟、第二戰遭遇在沙漠；從標題至 `THE END` 的 Docker＋Xvfb 正式 remake trace 則在
兩戰各輸出 `command`／`message`／`end` 四人隊畫面。這排除舊單勇者 fixture 造成的 HUD
誤判，也確認八頭大蛇素材本身已可顯示。

當時結論是 **戰鬥呈現尚未 match V2 oracle，不能宣稱 V3**：當時的 `Game.startPackFormation` 把
formation byte1 的 `BackgroundRaw`（35／26）直接傳入 `DecodePackBG`，但 formation byte2 的
`PageRaw=5` 沒有任何 runtime consumer。這個已證實（confirmed）的 direct mapping 產生天空草地
與綠色背景，和已證實的洞窟／沙漠原版畫面不同；原版 selector 的 archive consumer 尚未閉合，
不能猜它等於目前的 page index。下一個最小垂直切片僅應閉合「兩個 raw byte → pack reference →
renderer selector → 兩張正式 trace 圖」；同時保留不同隊伍／數值／動態 phase 的限制，不把近似
畫面升格為同狀態 V3。完整證據見 [`docs/127`](127-orochi-v2-production-compare.md)。

## 2026-08-12 歷史 checkpoint：固定編隊背景 selector V2 閉合

上一節描述的是修正前的 actual blocker；它不再是 current plan。IDA Pro 9.4 已從固定編隊
header 的 `raw[1]`／`raw[2]`，閉合到 `DS:0x0d71`／`DS:0x0d73`、`PACKBG.SCR` 的完整
`0x13d80` archive page 與 `MNSBK.PAL` 的 `0x30` palette bank。共同 renderer 現只消費
`dq3_cht` 的 `battle.background` 契約與每個 formation 的
`{ "page_raw", "palette_bank_raw" }`，資產、幾何、範圍與 raw parity 不一致一律 fail-closed。

同一條正式 `InputState` 重播的第一、第二八頭大蛇戰，現在分別呈現原版影片可見的洞窟與沙漠；
此二者是 near-state V2，非同狀態 V3。由第一性原理，這個窄修正值得做，因為兩場是主線必經、
高顯著的玩家可見設定錯配；反之，campaign 已可到 `THE END`，且沒有新 discrepancy 支持時，
不應擴張成所有 terrain→page 的完整 RE。generic battle 只保留已有草地 reference 的 `0/0`
baseline，明確不宣稱通用地形 parity。完整證據、hash 與停止條件見
[`docs/128`](128-battle-background-selector-re.md)。

## 2026-08-12 歷史 checkpoint：必經單頭目固定背景已接線

八頭大蛇之外，主線必經的怪力魔、巴拉摩斯與索瑪三連戰原本都由舊的單敵
`startBossBattle` 入口啟動，雖然原版 `battle.json.encounter.fixed_records` 已保留它們的
record，renderer 卻會吃到 generic `0/0` 草地 baseline。IDA 9.4 的原始 caller 證明五筆 record
皆走 `sub_1BE89 → sub_1BF35 → sub_1BFD1 → sub_1C688 → sub_1C6E5`；本輪改由 pack 的唯一
single-monster record 解出 `{page_raw,palette_bank_raw}`，再交給共用 battle primitive。

- 怪力魔 `DS:0x4ee4`：31／6；巴拉摩斯 `DS:0x4edf`：39／6；怨靈／殭屍
  `DS:0x4efa`／`DS:0x4eff`：42／8；索瑪 `DS:0x4ef0`：45／8。
- 同怪物有兩筆 record 時不猜選：八頭大蛇仍只走其兩階段 `staged_boss` pack event。甘達特與
  巴哈拉達救援本已由 event formation 直接傳 selector，無需另開單敵 fallback。
- `TestBuiltinDQ3BattlePackMatchesOriginalRawTables`、
  `TestRequiredFixedBossBackgroundsUsePackSelectors` 與既有頭目正式輸入 trace 均在 Docker＋Xvfb
  通過；完整 `InputState` campaign replay 也重新到達 `THE END`。

這是 D2／runtime V1 的資料與 renderer 閉環；只有八頭大蛇有原版影片 near-state V2。其餘五戰
尚無同狀態原版畫格，不能升格 V2／V3。此修正也不表示 generic terrain selector 已完成；完整
原始位址、推論等級與停止條件見 [`docs/129`](129-required-boss-backgrounds-re.md)。

## 2026-08-12 歷史 checkpoint：四人 HUD 訂正與剩餘交付 gate

IDA 9.4 已從 `sub_17C83` 的 `DGROUP 0x3e9c` writer 與 `sub_18222` 的內容 writer，訂正地表
四人 HUD：window `(152,238,352,80)`、label `+16px`、姓名／數值 `+32px`、姓名最多四 glyph，
末列依 class raw id 選職業 glyph。schema 已升為 `0.1.32`、DQ3 content 為 `0.1.37`；Go 只新增
通用 `PartyHUDLayout`，所有 DQ3 座標、glyph 與 evidence 留在 pack。正式 `InputState.Confirm`
畫面、資料層測試及本輪 64.14 秒的新遊戲到 `THE END` 重播均通過。此切片為 D3／near-state
V2，不是逐像素 V3；見 [`docs/130`](130-party-field-hud-re.md)。

該 HUD／發版前 checkpoint 的主線正式輸入交易曾達 E3；現行 campaign 由本檔末端
checkpoint 仲裁。必經 Boss 背景 selector 已接線，畫面等級依
`docs/128`／`docs/129` 分別停在 V1 或 near-state V2，不能概括成 V3 完成。
v0.1.34 也已由 checkpoint `9d639d0` 正式發布。Boss repeat-N 已由 `docs/148` 證實在本
EXE 不存在，remake 的正式 queue 以每個存活 actor 一筆鎖定；逐動作 frame、PCM wall-clock
或 generic terrain 全表仍不列成可由現有證據安全實作的必做項。收尾狀態為：

1. 三平台 patch／本機完整版已由精確 checkpoint 重建；公開 patch 的 release、雜湊與 macOS
   僅靜態驗證限制見 [`docs/131`](131-release-v0.1.34.md)。此項已完成。
2. Android 最新 source 的 AAR／APK 已重建並通過 `aapt`；專用 emulator image 也已鎖定
   emulator 37.1.11／API 34 revision 14。KVM 驗證已完成 Android 開機與 APK 安裝，但
   `MainActivity.applyGameChrome()` 的啟動空指標已修正。production 音訊版在 headless
   emulator 仍阻塞於 host audio backend，`firstWindowDrawn=false`；隔離靜音探針則完整顯示
   DQ3 並通過觸控與 lifecycle。此環境 gate 尚未升格為正式 Android 動態通過，見
   [`docs/132`](132-android-emulator-validation.md)。
3. 已停用仍會產生歷史 C/SDL 產品的 AppImage／Windows 發包入口；新發包只能走
   `dq3_remake_ebitan` 的 Go／Ebitengine 工具鏈。
4. 推廣片 r2 已於 2026-08-12 重新擷取現行 AppRun，依 6 月 28 日的完整冒險弧線改用
   全幅字幕帶、金框卡片、資料卡與雙畫面；音訊維持 48 kHz 雙聲道 AAC，並以完整解碼、
   音量與長段靜音的失敗即關閉驗收。三平台包與影片的現行交付集中於
   `dist-all/v0.1.34/`；檔案、雜湊與重建入口見 [`docs/134`](134-promo-video-r2-dist-all.md)。

任何新的 V3 工作仍須先有可重播原版 frame／音訊 trace 或明確新 caller；沒有新 oracle 時，
只保留為證據限制或可選 polish，不再擴張靜態 RE。

## 2026-08-22 current checkpoint：日夜原始 palette bank E2 接線

上一輪斷言審計確認 Go `DarkenPalette` 仍是四相位合理近似，且把 phase3 誤當日間 NPC
selector。IDA Pro 9.4＋IDAPython 本輪從 `sub_1EE23`、`sub_1EE76`、`sub_1EE9B`、
`sub_1ECDC` 重新閉合 clock writer、12-byte table、五個 `DQ3.PAL` bank 與 DAC upload：

- clock `0..0xef`，`0x78..0xef` 全段為夜間；palette 每 `0x14` tick 更新；
- selector 是 `[1,0,0,0,1,2,3,4,4,4,3,2]`，每 bank 16 色；
- 黑暗之燈原版 writer 固定 clock `0x8c`，不再近似成 `0x78`。

該日夜切片當時將 schema 升為 `0.1.33`、DQ3 content `0.1.38`；cycle、selector、asset key、黑暗之燈 clock
與 D3 evidence 全在 pack，Go 只保留通用 clock／bank primitive。原始 EXE／PAL parity、
pack validation 與受影響 runtime tests 已通過。此切片為 D3／E2／V2；仍缺同地點、同 clock
的 DOSBox DAC／畫面，不能宣稱 V3。完整 spec 與可重建 IDAPython 匯出器見
[`docs/136`](136-daynight-palette-bank-spec.md)。

## 2026-08-22 current checkpoint：中毒鏈 E2，完整主線 E3 已恢復

IDA Pro 9.4 已閉合 `D3MNS +0x0e..+0x13` 的 bit38，經 `sub_19AD6` 均勻選 bit、
`sub_199DC` remap 為 action `0x32`，再由 DGROUP `0x394e` 間接表進入 logical
`0xa307` 毒氣 handler。成功條件為 `rng(256)<=100`，writer 設角色 `+0x38 bit0x40`；
正式 movement consumer `sub_1964E` 每步加 1 HP 傷害（船／飛行狀態跳過），教會
`sub_1712B` 固定收 5G 後清除該 bit。

本輪同時訂正兩項過期斷言：48-bit mask 並非全都直接對到玩家 spell descriptor；教會
解詛咒也不是另一個 `+0x38` flag，而是掃描角色 item words 的 bit0x4000。現行 Go 已由
該中毒切片當時以 game-pack schema `0.1.40`／content `0.1.45` 接上 bit38 特殊動作、持久角色／存檔、正常移動
毒傷與教會 5G 解毒；原始 bytes parity、pack validation、component 與 save round-trip tests
通過，故此垂直鏈為 E2。完整 `TestOpeningProductionInputTrace` 已以正式聖水／特黑洛斯航路
越過遠洋，並由正式驅毒草選人交易越過 CTY23 單人回程；本輪初次重播曾停在沙曼歐莎
monster89 必經戰全滅，後續已由下方 `docs/140` 補給切片訂正並跨過。該句保存中途歷史；
最新完整 campaign 已由正式玩家路徑通過，不曾以清旗標或狀態注入繞過。
上句保存當時停止線；後續已由 `docs/147` 閉合 ITEM `+4/+5 & 0x0e00` → 裝備
item word `bit0x4000` → 同部位換裝拒絕 → 教會 level×100 並移除所有命中裝備。
完整 bytes、五筆 raw ID 與現行 E2 接線見 [`docs/147`](147-cursed-equipment-church-service-spec.md)。

同日後續以 IDA Pro 9.4 閉合 item `0x42` dispatcher：原版先移除持有者物品，再由
`sub_1469F` 判定選定角色死亡與 `+0x38 bit0x40`；即使無效果也會消耗，成功才清 poison
並顯示 rec364。現行工作樹已加入 pack-owned `clear_condition`、正式選人 modal 與原始
table／handler／D3TXT parity tests；精確證據與 per-character inventory 架構差距見
[`docs/138`](138-antidote-item-condition-re.md)。這使該道具交易達 E2／V1；完整主線 E3
已由本節末端最新重播恢復。

CTY23 的收斂證據見 [`docs/139`](139-cty23-solo-return-poison-recovery.md)：失敗狀態確認
主角中毒且背包已有 `0x42`，故補上移動前的正式道具選人輸入，而非改怪物或 encounter。
同一重播已抵達後續沙曼歐莎假王戰；下一輪只追 monster89 必經戰的第一個規則／策略差異。

> 上句保存當時的研究前沿；現行結果已由緊接的 `docs/140` 段訂正，不再以 monster89
> 作為 current blocker。

monster89 靜態閉合見 [`docs/140`](140-monster89-prebattle-resupply.md)：原始怪物記錄的
action gate 為 0、48-bit mask 全零，沒有證據支持新增特殊 action 或降低 Boss 數值。
初次 replay 的戰前狀態是勇者 `87/277 HP、0 MP`，兩名同伴皆陣亡。IDA／raw CTY 後續證實
CTY43 雖有 type3 block，日／夜 NPC 表都沒有 facility consumer 指向它；舊文件的「可用教會」
已訂正。現行正式路線改由 CTY43 旅店恢復魯拉 MP，再赴已造訪 CTY2 教會／旅店補滿；同一條
trace 已勝 monster89、完成變身杖、幽靈船與愛的回憶。monster89 垂直鏈因此恢復 E3；當時完整
campaign 尚未恢復，後續已由本節末端最新重播訂正。怪物 46 的逐怪 MP 與 bit41 全隊睡眠已依 `docs/141` 接線，正式 trace
已越過 CTY36、蓋亞之劍、銀寶珠、返航 monster125 與黃寶珠。monster125 的 ID 高危代理
已依 `docs/142` 推翻。中途第一個 blocker 曾依序為 monster4×5 與 monster77×2；
`docs/143` 已用 IDA 閉合 bit34→action42 自療並讓正式普攻策略越過，後者再由
`docs/144` 證明是固定寄放持綠寶珠同伴造成的返航，而非逐怪規則缺口。

`docs/144..146` 後續以正式酒場寄放無珠角色、六珠祭壇後復隊、教會逐人解毒與巴拉摩斯
終盤回復策略閉合。最終根因是 trace 的 `finalBoss` 下界錯設 `0x7a`，使程式中明寫給
巴拉摩斯 `0x79` 的 herbs／healer 分支永遠不可達；訂正為 `0x79..0x7c` 後，Docker＋Xvfb
下 `TestOpeningProductionInputTrace` 由標題重播至 `THE END` PASS（188.68 秒），且
`TestTraceSelectHealSpellPolicy` 與 church consumer tests 通過。此結果恢復 campaign E3；
V3 畫面／音效不因此升格；curse item metadata 的後續訂正見 `docs/147`。

## 2026-08-22 current checkpoint：單次行動 queue 與逃跑音效等待

`docs/148` 已用 IDA Pro 9.4 枚舉三個 action queue builder caller、全部 player／enemy
consumer caller及 raw 函式指標候選。正式與 alternate runner 都只為每個存活 actor
建立一筆 queue entry，consumer 也只執行一次；本 EXE 沒有 Boss repeat-N 路徑。Go
`buildTurnOrder` 與 component test 已鎖定相同 cardinality。這項結論是 `confirmed`，但不把
逐動作 SHP frame、畫面停頓或其他音效 cue 一併升格。

`docs/149` 接著閉合兩條玩家可見音效鏈：玩家成功逃跑寫 `FVOC cue 13`，敵人成功逃跑
寫 `FVOC cue 21`，兩者顯示訊息後都呼叫 `sub_208E2` 等待完成。game-pack schema
該逃跑音效切片當時以 schema `0.1.40`／content `0.1.45` 保存 raw cue、等待契約與 D3 evidence；Battle 在訊息實際顯示時
播放 cue，並以 VOC 原始 sample duration 向上換算 60 TPS 來阻擋提前輸入。原始 callsite、
descriptor、pack validation 與 headless duration tests 已通過，故這兩個角色達 D3／E2。
Sound Blaster DMA wall-clock 與重取樣標準行為依平台規格近似，不再是 RE gap；逐幀動畫
同步仍須有同狀態 oracle 才能升 V3。

## 2026-08-22 current checkpoint：玩家一般物理攻擊雙 VOC 序列

[`docs/150`](150-player-physical-sfx-sequence-spec.md) 以 IDA Pro 9.4 閉合
`sub_1B4F6` 的正常玩家物理分支：cue 6 → record `0x14a` → `sub_208E2` → cue 11 →
`sub_208E2` → damage consumer。pack 的 `player_physical.steps` 保存 cue 6／11 與各自 D3
evidence；Battle 只有在前一步原始 VOC duration gate 完成後才播放下一步，leader／companion
都先排 `actor_attack` 再排 damage。原始 EXE／FVOC parity 與 component tests 已通過，故此
有限角色為 D3／E2。此切片當時尚未涵蓋敵方 attack 與共同 damage cue；兩者已由下一節
`docs/151` 獨立閉合。cue 6 的其他 call-site、critical 與逐幀動畫仍是
unknown／V3，不由兩個切片外推。

## 2026-08-22 current checkpoint：一般物理結果、敵方攻擊與個別死亡訊息

[`docs/151`](151-common-physical-result-sfx-spec.md) 以 IDA Pro 9.4／IDAPython 接續
`sub_1AC05`、`sub_1ACCE`、`sub_1AFC6` 與死亡分支，閉合敵方 cue4／record331、雙方成功
傷害 cue1／record332 或 333、共同 miss cue3／record335，以及敵人／玩家 HP 歸零後的
record336／357。該切片當時的 game-pack schema `0.1.40`／content `0.1.45` 新增通用 sound／text roles；
Battle 依玩家可見順序排 attack → result → 個別死亡 → 勝負摘要，cue completion gate 依
FVOC source duration 分別為 11／11／14 frame。

原始 EXE／D3TXT／FVOC parity、hit／miss／death component tests、完整 `internal/...`、排除
campaign 的其餘完整 `game` tests 與隔離 desktop build 均通過；正式
`TestOpeningProductionInputTrace` 另在乾淨 Xvfb session 由標題抵達 `THE END`（134.13 秒）。
整包單程序驗收會因先前 Ebitengine 測試資源累積觸發容器 OOM／GC 壓縮抖動，因此 current
gate 明確拆成上述互補集合；這是測試環境分類，不是產品缺陷。cue9 的
`word_272E1==3` 特殊分支、critical、咒文傷害、NVOC selector、硬體 wall-clock 與逐動作
SHP／畫面同步仍是 unknown／V3，不能由本切片擴張為完整戰鬥 V3。

## 2026-08-22 歷史 checkpoint：怪物 action3／持久麻痺 E2，當時 campaign pending

[`docs/152`](152-special-physical-paralysis-spec.md) 已以 IDA Pro 9.4／IDAPython 閉合
`D3MNS +0x0e` MSB-first bit3：怪物 3／48／51 經 `sub_19AD6 → sub_199DC` 選出
action raw3，`sub_1ACCE` 以 `attack/2+rng(attack/4)` 無視守備造成單體傷害，存活者才寫
角色 `+0x38 bit0x10` 並顯示 record201。cue4／record331 → cue9／record334 → damage →
paralysis 的順序、FVOC cue9 descriptor、死亡不麻痺均已有 raw parity 與 component test。

持久麻痺現在與 bit0x20 sleep 分離：命令與 action queue 略過、戰後保留、共享 0x28 個合法
場景步後全隊解除，滿月草 field handler 先消耗再判定，condition／倒數可 save/load。IDA
`0x1ccb2` 另證實 battle CureStatus 以 `DX=0x10`／record399 呼叫 `sub_1469F`，Go 已同步清
persistent paralysis；jump-table caller 尚無 IDA direct xref，該 dispatch 邊維持 `strong`。

本切片精確 tests 已通過。完整 `TestOpeningProductionInputTrace` 因原先的測試玩家策略不知道
持久麻痺，初次在怪物 48／51 遭遇全滅；後續只以正式購買／使用聖水、CureStatus、保守逃跑
與保留後續驗收所需的一瓶聖水修正 trace，不改怪物數值、座標、旗標或 RNG。乾淨
Docker＋Xvfb 重播曾由標題抵達 `THE END`（121.799 秒），該時點工作樹 campaign E3 恢復。
文件／註解稽核撤回一項未經 selector RE 就改動 RNG 次序的嘗試後，同一正式 trace 又以
134.912 秒通過；該次失敗與撤回保留為證據，不能靠調整 trace 掩蓋未證實的原版順序。
後續 `docs/153` 已以 IDA 9.4 證實正常分支確實在選 action bit 前抽存活目標；移除
all-party mask 的跳過近似後，targeted tests 通過，但目前完整 replay 在 CTY38→CTY70
航段的 monster77×2 全滅。多次以 CTY83／CTY38 正式商店、旅店、教會與 8／16 瓶聖水
嘗試仍停在同一狀態，已撤回這些未收斂的 trace 改動。該 checkpoint 的 campaign E3 因而 pending；
下一輪需把該 deterministic 長航路當成獨立玩家路線／補給切片，不改怪物或 RNG。

剩餘實作缺口的第一順位：現行 `bcItem` 仍只支援藥草，原版戰鬥道具清單／滿月草
transaction 尚未由 selector → inventory owner → handler → target UI 完整閉合；不得用
field item handler 猜接。其後順位固定依本檔頂端最後功能 worklist，不由舊 milestone 改寫。

## 2026-08-22 歷史 checkpoint：原野道具 owner E3，當時 campaign 後移至取船後航線

[`docs/156`](156-shop-personal-inventory-capacity-spec.md) 與
[`docs/157`](157-battle-drop-personal-inventory-capacity-spec.md) 已把商店購買與戰鬥掉落
改回每名角色含裝備共八格；全隊滿格時不再把物品無限塞入勇者 slice。這使舊 trace 隱藏的
正式操作缺口浮現：戰士滿格後無法購買青銅盾，而當時原野道具面板只能操作勇者。

[`docs/158`](158-field-item-owner-selection-drop-spec.md) 以 IDA Pro 9.4 閉合
`0x1372F..0x13B0F`：原版先選持有者，保存 `DS:062D/062F`，再於該角色八格執行使用、
給予、同 owner 重排或丟棄。Go 已接上同一 ownership；EXE bytes 與 component tests 通過。
正式 trace 已越過羅馬利亞購裝並繼續至取船後航線，故此有限角色達 E3。

該時點工作樹的完整 campaign E3 仍是 **pending**：當次乾淨 Docker＋Xvfb replay 在 CTY16
之後的 monster51 遭遇全滅，當時四人 HP 均為 0、聖水存量為 0。這是新的第一個玩家路徑
blocker；尚未證明是 engine 規則、補給策略或個人物品容量交易問題。下一輪必須先讀相關
encounter／補給 RE，寫獨立 spec，再決定實作；不得以較早的 `THE END` checkpoint 宣稱
現行工作樹完成，也不得改怪物數值或 RNG 掩蓋。

> **訂正：**以上 monster51 是當時的第一個 blocker，不是現況。後續 `docs/159..178`
> 逐項以「RE 文件 → spec → 正式玩家輸入實作」閉合隊伍物品持有、獎勵容量、原野補血與
> 抵達目標同一步的戰鬥 modal；最新完整重播結果以本檔頂端與下節為準。

## 2026-08-22 current checkpoint：主線 E3 恢復，RE／V3 邊界保留

[`docs/159`](159-party-owned-field-supply-trace-spec.md) 至
[`docs/178`](178-rainbow-tile-arrival-encounter-spec.md) 依序閉合隊伍道具供應與消耗、
共用八格獎勵 writer、金／銀／黃寶珠及終盤必要道具容量、野外回復咒文，以及航行抵達
目標格同一步遭遇的輸入所有權。黃寶珠的共用隊伍寫入鏈另以 IDA Pro 9.4／IDAPython
重查 `sub_18C0F` caller → 共用八格 writer → party inventory consumer；原始 EXE 為
115,282 bytes，SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。

乾淨 Docker＋Xvfb 的 `TestOpeningProductionInputTrace` 已從標題以正式 `InputState`
抵達 `THE END`，耗時 115.629 秒。沒有清狀態、座標注入、debug key、怪物數值或 RNG
近似；因此目前 campaign 為 E3。這不證明 DQ3.EXE 已逐行完整解讀，也不把戰鬥逐動作
SHP、所有選單或所有場景升為 V3；PCM wall-clock 已依平台規格近似並退出 RE 範圍，
其餘項目仍須在出現玩家可見差異時，以
獨立窄切片取得原版 oracle，不能重新把 remake 收尾變成無界反編譯。

同一工作樹的比例驗收另已通過 `go test ./internal/...` 全部套件、Docker＋Xvfb
`go test ./game` 全套（164.874 秒），以及只指定 `main.go`、不碰使用者受保護
`tmp_dump.go` 的隔離 Linux desktop build（13,373,336 bytes）。這些證明目前程式與資料
契約可建置且回歸一致；不取代 macOS 真機、原版逐畫面或人耳音訊驗收。

## 2026-08-23 current slice：戰鬥道具 owner／selector／消耗品 handler

[`docs/179`](179-battle-item-selector-runtime-spec.md) 以 IDA Pro 9.4 閉合
`sub_1C1D8 → sub_1B836 → DGROUP 0x0d50/0x0d4b → sub_1B95C`：原版從目前
actor record `+0x3a` 掃八個 item word，保存原始格序與目標，再由 near function pointer
分派；不使用全隊藥草計數。Go 已新增正式 `phItem`、逐 actor inventory/equipment slot、
target route、先消耗與戰後逐 owner 寫回；`0x41/42/43/44/45/48/4c` 的已證實 handler
參數進入 `battle.json.items[]`。schema `0.1.44`／content `0.1.49`。

本切片的 `go test ./internal/...` 與 Docker＋Xvfb `go test ./game -run Battle` 為比例驗收；
完整 campaign 未因本批重跑。`0x05/0x0a/0x0c/0x10/0x14/0x17/0x18/0x1a/0x1c/0x46`
共用 spell-action consumer 仍由下一個「實際使用的剩餘怪物 action」切片統一閉合，避免
在戰鬥道具分支另造一份猜測 mapping。它們目前缺 definition 時失敗即關閉，不冒稱完成。
