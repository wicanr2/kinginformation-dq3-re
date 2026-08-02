# DQ3 Go/Ebitengine remake 開發標準

本 repo 的現行產品是 `dq3_remake_ebitan/`。目標是忠實完成精訊版 DQ3，而不是延伸歷史
`dq3_remake/` C/SDL 或 `re/` C 重建。後兩者只能作線索，不能作原版規格。

## 每輪開始

1. 執行 `git status --short`，保留所有使用者未追蹤／未提交檔案。
2. 依序讀 `PROJECT_MEMORY.md`、`CONTEXT.md`、`README.md`、
   `docs/74-ebiten-remake-completion-plan.md`，再讀當前子系統的直接證據文件。
3. 以 `docs/74` 為唯一 current plan；舊 Markdown 是歷史線索。若文件、程式與原版證據衝突，
   先重查證據並修正過期文件，不累積互相矛盾的摘要。
4. 從上一個合法 production checkpoint 重播，記錄第一個玩家實際遇到的 blocker。

## 原版證據規則

證據優先序：

1. 本機 DOSBox 同操作、完整實況影片的玩家可見結果。
2. 原始 `DQ3.EXE`、CTY、D3TXT、ITEM、怪物及其他資料。
3. 攻略與網路資料只供定位。
4. C remake、舊測試與舊文件不可代替原版 oracle。

逆向必須閉合：

```text
入口/caller → writer → table/state → consumer → 玩家可見效果與副作用
```

- 位址一律標明 `(file)`、`(logical)` 或 `DGROUP`；換算規則見 `CONTEXT.md`。
- 凡有 RE／反組譯需求，**必須優先使用 `/home/anr2/ida_94_official/dist` 的 IDA Pro
  9.4**，以其 database、交叉引用、型別與反編譯結果作主要分析工具；不得先用 Ghidra
  產生結論後才補查 IDA。`tools/dis.sh`／Capstone 適合可重現的批次位址掃描，Ghidra
  只在 IDA 無法處理或需第二套工具交叉驗證時補用。IDA 授權、database 與輸出不得加入
  Git。
- 即使使用上述本機 IDA 安裝，也只能把明確需要的唯讀路徑掛入一次性 Docker 容器執行；
  禁止直接在 host 啟動 IDA、反編譯、索引或其他專案作業。容器工作完成後依本檔 Docker
  生命週期規則立即停止並清除。
- **IDA 分析必須非破壞且可追溯**：原始 EXE／DAT 永遠唯讀並保留原位置；每次分析先記錄
  輸入路徑、檔案大小與 SHA-256，再把工作副本、`.i64`／`.idb`、`.asm`、腳本輸出放在
  `/tmp` 或 gitignore 的工作目錄。不得原地解包、patch、rename、覆寫或把 IDA database、
  授權及解包 binary 加入 Git。
- **語意只能附加，不得取代原始定位**：IDA 的 rename、comment、type、function boundary 與
  decompiler interpretation 都是 annotation。證據筆記與可重建的 sidecar／ledger 必須以
  原始 symbol 或位址為 key，同時保留 raw bytes／指令、新增語意與推論等級；不得只留
  重新命名後的結論。若為操作方便將 annotation 匯入工作 database，仍必須保留原始
  身分與 sidecar，且 database 本身不是證據。若修正語意，保留舊斷言為「已推翻」並附反證；
  訂正所需證據不得比原斷言更弱，不移動或改寫原始位址使既有引用失效。
- 每一項 IDA annotation 與文件結論都必須帶**推論等級**：`confirmed`（原始 bytes／明確
  writer-consumer／可重現實機閉合）、`strong`（多條獨立靜態證據但尚缺玩家可見閉合）、
  `hypothesis`（待驗假說）、`unknown`（未知）。不得因 IDA 自動命名、反編譯可讀或攻略吻合
  就標成 confirmed；production JSON 仍遵守下方 D2／D3 gate。
- 推論等級只對明確的輸入 binary 成立；每筆必須帶輸入路徑、檔案大小、SHA-256、IDA
  linear address／位址基準、原始範圍及 consumer。不得將一個 EXE 的 compiler、資料佈局或函式
  語意外推到另一個 EXE；缺 caller／writer／consumer 閉環時不得標為 `confirmed`。
- 查資料語意時以 IDA database 的 xref、型別與 caller/callee graph 為主，不以 grep 攤平
  `.asm` 代替交叉參考。xref 只保證直接參考；若讀取很多但 writer 異常少，必須再追
  address-taken、指標別名、segment:offset 與間接寫入。讀／寫分類應使用 IDA xref type，
  不用助憶碼字串或 operand 位置猜測。
- Headless IDA 腳本必須把結果寫入明確 sidecar 檔並驗證檔案非空、含輸入 hash／位址基準；
  不得只相信 exit code 或 stdout。優先沿用已驗證可工作的 IDC／IDAPython 路徑；某種腳本
  介面失效時記錄限制，不得靜默缺資料仍宣稱分析完成。
- 位址與函式的已知語意應在 dump／匯出流程中由 sidecar 自動附註，並連同推論等級顯示；
  不可只靠分析者記得先搜文件。新語意回填索引前先搜既有文件與程式，零命中只能表示
  索引尚無資料，不能證明該位址從未被研究。
- 判讀呼叫參數前先辨識 compiler、memory model、calling convention、segment register 與
  字串格式；Turbo Pascal 的左至右壓棧／callee cleanup、C 的慣例或 far pointer 不得互套。
- 不可用 DQ3 慣例、「合理值」、C remake 或攻略猜 production 設定。
- 機制存在不代表位置、初值、旗標、gate、時序、畫面與消耗已 match。
- 同一怪物、NPC 或場景在不同階段出現時，不得合併成單一 trigger 或用一次性獎勵近似；
  必須逐階段追入口、選項、編隊、掉落抑制、移動／轉場、旗標交易及事後可互動物件，再以
  game-pack 的有限狀態事件描述。原始資料欄位在 loader 與 consumer 未閉合前一律保留
  `unknown_XX`，不可因舊文件或數值看似合理就命名。
- 原版素材、影片、IDA database／授權與發佈包不得加入 Git。

## 實作與驗收

每批只閉合一個垂直切片：

1. 寫 evidence/spec ledger，列出已知與未知。
2. 實作正式玩家入口，不以 debug key、env、直接事件函式或狀態注入串流程。
3. 加 component tests，鎖定 gate、狀態交易及失敗不消耗。
4. 從上一 checkpoint 以正式 `InputState` 建 production trace。
5. 驗證事件後仍可往下一節點，並做 save/load round-trip。
6. 產出或更新受影響的 runtime PNG，與原版同狀態目視核對。
7. 更新 `docs/74`、`PROJECT_MEMORY.md` 及必要的單一主題 RE 文件。

### 共用 game pack 與 JSON

- **遊戲引擎與遊戲資料必須分離**。`game/` 與共用 `internal/` 引擎只能知道跨版本的資料
  契約、載入／驗證、輸入、繪圖、存檔及具名狀態機；不得知道 DQ3 專屬人物、地名、文字、
  record、座標、旗標或數值。所有版本專屬內容由 versioned game pack 提供；新增 DQ1／DQ2
  時應只需加入或選擇 pack，不得複製引擎或在共用 Go 程式加入遊戲版本分支。
- 引擎不得修改 pack 內容，也不得在缺資料時自行推導「合理預設」。pack 只描述資料與引用，
  不得嵌入任意程式碼或複製引擎流程。兩者的邊界由 schema、reference validation、
  `content_version` 與 canonical hash 鎖定。
- 長期架構是共用 Go／Ebitengine 核心加 `dq1_cht`、`dq2_cht`、`dq3_cht` versioned
  game pack；JSON canonical 欄位契約見 `docs/84-game-pack-json-contract.md`。
- **禁止把版本專屬設定新增成 Go 常數或 table**：CTY／section／座標、tile subid、story
  flag、對話 record、玩家可見文字、選單標籤、系統訊息、怪物／編隊、寶箱 gate、商店
  價格、掉落與音訊 cue 都必須放在對應 game-pack JSON。Go 只保留跨版本的 loader、
  validator 與具名狀態機／effect primitive。
  為反組譯診斷暫時硬寫的值不得進 commit；切片收尾前必須遷入 JSON、更新 `docs/84` 欄位
  說明，並加原始 EXE／DAT parity test。缺欄位或未知引用一律 fail closed，不設 Go fallback。
- 視窗的外框座標／尺寸、文字 inset、欄數、每頁行數與換頁容量也是版本專屬資料；必須從
  原版視窗結構及其 consumer 閉合後放進 game-pack JSON。不得用截圖目測、C remake 尺寸或
  目前螢幕剩餘空間猜值，也不得在共用 Go renderer 留 DQ3 專屬 fallback。
- production Go 禁止直接寫玩家會看到的中文、日文或英文句子；畫面流程只能引用穩定的
  text ID。各 game pack 的文字 JSON 保存實際字串與原版 record/evidence，版面寬度、換頁、
  選項與插值參數使用具名欄位，不把控制碼或排版邏輯偷塞進字串。開發用 assertion、log 與
  測試說明不屬於遊戲內容，但不得在 runtime 畫面作為缺字 fallback。
- 新增或更改遊戲設定前，先判斷它屬於 engine behavior、game data 或大型 binary asset。
  可調內容不得繼續新增為散落的 Go table；引擎狀態機也不得為了資料化而改成任意 JSON code。
- production JSON 的初值至少要有 D2 evidence，會改變流程的值需 D3。未知值 fail closed，
  不用 0、合理預設或 C remake 填空。
- 每批遷移必須有 schema/reference validation、原始 EXE／DAT parity test，以及正式玩家
  input trace。Go decoder 保留為 oracle，不能因 JSON 化刪除。
- Desktop 外部 pack 可免重編譯；嵌入 Android／Web 的 pack 仍需重包，除非另有具版本、
  hash／簽章驗證的外部更新機制。

完成度用：

- E0 線索；E1 原版證實；E2 runtime parity；E3 玩家流程閉合。
- V0 無畫面證據；V1 可顯示；V2 有 runtime 對拍；V3 同狀態 parity。

只有從新遊戲開始、不用 debug shortcut，以正式輸入到 THE END，關鍵事件、畫面、聲音與
存讀檔均通過，才可宣稱 remake 完成。單元測試全綠、孤立 handler 或後段 checkpoint
不等於 campaign 完成。

## 測試、container 與提交

- 圖形測試在具備 `assets_raw` 的正確工作目錄，以 `go test -c` 後配合 Xvfb 執行；
  素材缺失造成的 SKIP 不算驗收。
- 每次至少跑受影響的 tests；重大切片收尾跑完整 `game`、`internal/...` 與 desktop build。
- Docker 驗證預設使用 `docker run --rm` 臨時 container；不用時不得留存。若工具必須
  建立常駐 container，工作結束要先 `docker stop` 再移除該明確 container。不得連帶刪
  image、volume 或專案資料。
- 提交前跑 `git diff --check`、檢查圖片與文件代表現行程式，並確認沒有使用者 scratch、
  Android local libs、原版資產或 RE database。
- 提交前另跑 `rg` 檢查本批新增 Go 程式是否出現版本專屬 raw ID／座標／record／flag；
  每個命中都必須是 parser、validator、原始格式 oracle 或測試證據，不能是 production fallback。
- 有重大修改才 commit + push；commit 訊息描述已閉合的垂直切片，不誇大後續流程。

## 工作樹保護

現有未追蹤的 `android/app/libs/`、`scratchpad*`、`tools/dosbox_verify_*`、`tools/tmp_*`
通常是使用者資料。除非使用者逐項明確要求，禁止刪除、覆寫或加入 commit。不要用
`git reset --hard`、`git checkout --` 或廣域清理命令處理髒工作樹。

`CLAUDE.md` 保留為短入口；術語與知識索引由 `CONTEXT.md` 維護，不在本檔複製易過期的
位址表或逐項進度。
