# DQ3 remake 接手記憶與完成評估

> 更新：2026-07-28（Asia/Taipei）
>
> 用途：接手本專案時先讀此檔，再讀 `CLAUDE.md`、`CONTEXT.md`、`docs/65-ebitan-flow-realign.md`、
> `docs/66-original-flow-oracle.md`、`docs/70-re-flow-completion-plan.md`、
> `docs/73-gap-fix-worklist.md`、`docs/74-ebiten-remake-completion-plan.md`。本檔記錄目前可信狀態與決策，
> 不取代各專題 RE 文件。

## 1. 專案真正目標

完成精訊資訊未發售《勇者鬥惡龍 III》中文版的保存、反組譯與現代 remake。remake 的最終驗收不是
「可編譯」、「單元測試全綠」或「靠 debug 捷徑可到結局」，而是：

> 玩家從新遊戲開始，不使用 debug 鍵或環境變數，按原版正常流程走到結局；關鍵流程、操作、
> 畫面與聲音有原版 DOSBox／錄影／攻略／RE 證據。

版權素材不進公開 repo；執行時由玩家提供合法原版資產。

## 2. 目錄與產品線

- `re/`：DQ3.EXE 的 C 反組譯／重建與 SDL shim，屬研究證據。
- `dq3_remake/`：C99/SDL2 remake。子系統很多、測試成熟，可作 Ebiten 移植參考，但流程曾被證明
  不完整，不能把它當無條件 oracle。
- `dq3_remake_ebitan/`：Go/Ebiten remake，現行應接手的主要產品；目標涵蓋桌面、Android、
  未來 WASM。
- `docs/`：RE、格式、oracle、流程與歷史知識庫。
- `assets_raw/`、錄影與部分輸出：合法原版／版權材料，通常 gitignore。

## 3. Ground truth 優先序

發生衝突時依序採信：

1. 使用者對原版的實測與驗收。
2. 原版 DOSBox 可重現行為／原版完整錄影。
3. 杜勝利、青衫等攻略的一手流程描述。
4. DQ3.EXE 靜態 RE 與原始資料。
5. C remake 現行行為與測試。
6. 舊 README、進度宣稱與 debug-driven 測試。

程式碼證明「目前做了什麼」；原版 oracle 證明「應該做什麼」。兩者不可混用。

## 4. 目前可信進度

### 已有堅實基礎

- 原始素材格式、地圖、字型、文字、怪物、NPC、事件、戰鬥等已有大量 RE 與解析器。
- EXE 重組曾做到 byte-identical；C remake 有廣泛子系統與測試。
- Ebiten 已有標題／命名／性別／能力確認、開場家中核心、城鎮與地表、NPC、命令窗、對話、隊伍／酒場元件、
  戰鬥、道具裝備、商店設施、船、晝夜、存檔、觸控與音訊等大量基礎。
- Ebiten 已修復 NPC 三層載入問題：不丟 `b2<4`、日夜表、story flag visibility。
- W1–W5 已完成：文字插值、加權怪群、多敵、玩家狀態咒文、酒場名冊／入隊分離、C NPC parity。
- R-2 沙曼歐莎規格已定位：CTY44；夜間假王 sec1 `(14,6)`；真王 sec4 `(17,31)`；
  拉之鏡 `0x61` 揭露怪力魔 89；勝後變身杖 `0x62`、設旗標、切白天。

### 仍不能宣稱完成

- 正常玩家路徑尚不能從新遊戲一路自然破關。
- 酒場正式入口與移除預建 placeholder 隊伍仍未收尾（R-1/C-3）。
- R-2 只有 spec，尚未實作及原版行為驗收。
- R-3 不死鳥祠堂、六珠 gate、飛行坐騎是最大剩餘子系統。
- R-4 巴拉摩斯城、boss 與自然下降仍待定位／接線。
- R-5 龍王城、光之珠、彩虹橋、索瑪神殿自然入口與後門移除仍待完成。
- `U`、`Z`、Enter 進城、城內 Cancel 出城等 debug／demo 後門仍存在；自然路徑完成前不可先拔，
  最終驗收前必須移除或隔離到非 production build。
- 開場出生資料與事件 transaction 已由 EXE 定錨：sec4 `(5,5)`、單人／布衣、rec82→83→81→
  母親自動帶至 sec0 `(8,38)`／rec80、王座給 50G + `00,01,01,03,1f,1f` 後才完成 `msStart`。
  能力值確認與標題→王座 production-input trace 已補；仍有 fidelity 缺口：確認畫面人物立繪、
  母親逐格移動動畫、attract 演出、
  四人縱列、THE END。
- 遭遇分區、商店賣出、部分戰鬥呈現／音效與低信心機制仍有缺口。

## 5. 工作樹與基準注意事項

2026-07-28 讀取時，DQ3 repo 有使用者未追蹤檔案：

- `dq3_remake_ebitan/android/app/libs/`
- `scratchpad_check.go`
- `tools/dosbox_verify_exit2.sh`
- `tools/dosbox_verify_exit3.sh`
- `tools/dosbox_verify_exit_town.sh`
- `tools/dosbox_verify_opening.sh`
- `tools/dosbox_verify_talk_mother.sh`
- `tools/dosbox_verify_talk_mother2.sh`
- `tools/tmp_parse_cty00_sec4.py`
- `tools/tmp_render_sec4.py`

不得刪除、覆寫或誤納入 commit。接手時先重查 `git status --short`。

最新已提交暫停點為 `4880946`；明確 resume 點在 `docs/73`：先實作 R-2，再做 R-3～R-5。

## 6. 從 FD2 與《青色枷鎖的詛咒》採用的工程方法

參考來源：

- `/home/anr2/cht/fd2/CLAUDE.md`、`README.md`、`remake/README.md`
- FD2 `docs/knowledge-base/42`、`56`、`57`、`91`
- `/home/anr2/cht/golden_box/curse_of_the_azure_bonds/CLAUDE.md`、`CONTEXT.md`、`PLAN.md`、`README.md`

採用以下做法：

1. **三條完成度分開**：資料／規則證據、玩家可達流程、玩家可見 audiovisual parity 分開追蹤，
   禁止把測試數量換算成遊戲完成百分比。
2. **垂直切片優先**：每輪閉合一條正常輸入可抵達的完整小流程：
   入口 → 狀態變化 → 畫面／聲音 → 返回／後續節點 → 存讀檔。
3. **可重播 trace**：為關鍵流程保存 deterministic input/state trace；測試直接呼叫內部函式只能算
   component test，不能算玩家流程驗收。
4. **證據分級**：
   - E0：推測／攻略線索。
   - E1：靜態 RE 或原始資料證實。
   - E2：remake runtime 在相同狀態可重播，並有 oracle 對照。
   - E3：正常玩家路徑抵達，跨場景／存檔回歸通過。
   只有 E3 可關閉主線流程 GAP。
5. **未知 fail-closed**：未知旗標、事件分支或副作用不得用「合理的 DQ3 行為」默默猜接。可保留明確
   boundary、蒐集 oracle，再實作；若為恢復未完成原版而有意設計新行為，必須標成 remake policy。
6. **畫面證據矩陣**：標題、命名、城鎮、對話、戰鬥、酒場、商店、結局等逐界面記錄 oracle、
   runtime screenshot、缺口與證據等級，避免 raw decode 圖被誤稱 runtime parity。
7. **單一權威 worklist**：`docs/65` 與 `docs/70` 目前有重疊／過期敘述；後續應合併成一份
   current gate matrix，歷史文件只留背景。

## 7. 重新評估：建議完成路線

### Phase 0：先建立誠實基線

- 固定可重現 build/test 指令與目前結果。
- 新增 debug dependency inventory，列出所有 production input/env 後門及其自然替代路徑。
- 建立主線 60 步 coverage matrix：每步標 `unreachable / component-only / E2 / E3`。
- 先補最小 input trace harness；不要再用 `spine_test` 直接呼叫內部函式代表可破關。

### Phase 1：閉合新遊戲至出發的第一條 E3 slice

- R-1/C-3：定位並接上露依達酒場入口。
- 移除預建示範隊，完成登錄 → 招募 → 入隊。
- 核對原版正常出城方式，移除 Cancel/Enter demo 行為的 production 依賴。
- 收斂母親護送／謁見的最小忠實流程。
- 驗收 trace：標題 → 命名 → 家 → 母親 → 國王 → 酒場 → 四人隊 → 正常出城。

### Phase 2：按攻略順序逐段閉合中盤

- 不應直接跳到 R-2 後一路只做終盤；先把 `docs/66` 每個既有任務 gate 做成玩家可達 trace。
- R-2 怪力魔作為下一個獨立 vertical slice 實作，但需補：
  道具選單使用拉之鏡 → 位置／朝向／夜晚 gate → 對話 → boss → 勝敗分支 → 旗標／獎勵 →
  白天／NPC 可見性 → save/load round-trip。
- 對已宣稱完成的取船等流程，用正常輸入 trace 抽驗；未達 E3 者降級，不重做已證實底層。

### Phase 3：最大風險 R-3 單獨立項

- 先完成六珠來源 coverage，逐顆證明可由正常玩家取得。
- 再定位不死鳥祠堂與祭壇事件。
- 飛行坐騎拆成移動模型、地形限制、起降、城鎮入口、儲存、觸控六個 contract。
- 以原版同位置／同操作的 DOSBox 或錄影對照，不以 C remake 自造行為作唯一 spec。

### Phase 4：終盤自然接線

- R-4：巴拉摩斯城 → boss → 世界事件 → 自然下降。
- R-5：龍王城／光之珠 → 彩虹橋 → 索瑪神殿 → 終戰 → 結局。
- 每完成一條自然入口才移除對應 `U`／`Z`／Enter 後門。
- 最終保留 debug 能力時，須用 build tag 或明確 developer mode 隔離，不接受 production 按鍵。

### Phase 5：完整 playthrough 與 audiovisual parity

- 建立可重播的無 debug 全流程 smoke trace，允許多段 checkpoint，但每段由真實 save 接續。
- 使用者實玩是 release gate。
- 再處理 attract、四人縱列、THE END、音效、戰鬥訊息、palette 等 fidelity 缺口。
- 桌面穩定後才把 Android/WASM 當 release gate；跨平台不能替代玩法 parity。

## 8. 風險與工期判斷

這不是「剩 R-2～R-5 四個事件」的小收尾。底層完成度高，但缺的是高耦合 campaign runtime 與完整
玩家驗收閉環。最大風險依序為：

1. 飛行坐騎與其世界／入口／存檔整合。
2. 六珠取得鏈與分散 story flags 的完整可達性。
3. 原版未完成、攻略與 RE 不一致時的產品政策。
4. 缺少 deterministic 玩家 trace 導致再次出現「測試綠但玩不到」。
5. 兩套 remake 與多份進度文件互相漂移。

在未完成 Phase 0 coverage matrix 前，不給虛假的完成百分比或精準日期。工程量應以「E3 主線步驟關閉數」
燃盡，而非 commit、測試或 handler 數量衡量。

## 9. 下一個安全動作

1. 保留現有 dirty/untracked files。
2. 跑 Ebiten完整測試建立 baseline（目前開局／元件鏈 targeted tests 已綠；完整 suite 待跑）。
3. 原版新遊戲初始化、能力生成／確認、母親／國王 transaction 已落地；
   `title→name→gender→stats-confirm→home→mother→castle→king` production-input trace 已通過。
   2026-07-28 另證實成長表順序為 `STR,VIT,AGI,HP,MP,INT,LUCK`，舊 docs/23 與 C/Go
   將 pair1 認成 MP 的 patch 已撤銷；真 MP 是 pair4 (`0x1a4ae/f`)。主角七項能力現在持久化，
   Lv1 與升級使用 `sub_fa57` RNG transaction，守備依 `sub_9521` 為耐力＋防具。
4. 實作 R-2 前先補一條玩家層測試，從道具選單實際使用拉之鏡，不直接呼叫事件函式。
5. 完成 R-2 E3 slice 後更新本檔與唯一 worklist，再開始 R-1 或 R-3。
