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
- 可使用 `tools/dis.sh`、Capstone、Ghidra，以及本機 IDA Pro 9.4。
- 不可用 DQ3 慣例、「合理值」、C remake 或攻略猜 production 設定。
- 機制存在不代表位置、初值、旗標、gate、時序、畫面與消耗已 match。
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

- 長期架構是共用 Go／Ebitengine 核心加 `dq1_cht`、`dq2_cht`、`dq3_cht` versioned
  game pack；JSON canonical 欄位契約見 `docs/84-game-pack-json-contract.md`。
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
- 有重大修改才 commit + push；commit 訊息描述已閉合的垂直切片，不誇大後續流程。

## 工作樹保護

現有未追蹤的 `android/app/libs/`、`scratchpad*`、`tools/dosbox_verify_*`、`tools/tmp_*`
通常是使用者資料。除非使用者逐項明確要求，禁止刪除、覆寫或加入 commit。不要用
`git reset --hard`、`git checkout --` 或廣域清理命令處理髒工作樹。

`CLAUDE.md` 保留為短入口；術語與知識索引由 `CONTEXT.md` 維護，不在本檔複製易過期的
位址表或逐項進度。
