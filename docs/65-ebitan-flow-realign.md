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
每個流程步驟三重來源對齊:(1) 原版實機錄影(`dq3_real_video/`) (2) 杜勝利/青衫攻略
(3) RE 證據(re-log-722 / disasm)。最終仲裁 = 使用者實測(玩過原版破關)。

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
- [ ] A1 錄影盤點+抽格(haiku):ffprobe、全片每 2s 一格 + 開場前 10 分鐘每秒一格 → `dq3_real_video/frames/`
- [ ] A2 流程 oracle 文件(sonnet):杜勝利 56 步 × 青衫 × 既有 docs → `docs/66-original-flow-oracle.md`
      (開機→破關逐步:地點/觸發/獲得/解鎖/出處;開場序列專節)
- [ ] A3 影格核對(Fable):用 A1 影格驗 A2 的開場序列與關鍵節點,填 video timestamp

### B — 差異盤點 ⏳
- [ ] B1 ebitan 實際流程 vs oracle 逐步比對(code 為準)→ FLOW-GAP 清單(本檔附表,嚴重度排序)
- [ ] B2 C remake 宣稱抽驗:重跑 game_tester(93/93?)+ 抽 grep 3–5 個 audit 宣稱 → 決定 C 當 spec 的可信度

### C — 逐 GAP 修 ebitan(C 為 spec;C 也缺的回頭 RE)
- 順序 = 玩家遭遇順序:開場/創角 → 出生點/初期場景 → 遭遇區/晝夜 → runner region 事件/真實里程碑觸發
  → boss/道具 gate → 終盤自然觸發(移除 Z/U 依賴)
- 每 GAP:sonnet 移植(C spec 指路)→ Fable 獨立核實(diff/測試/headless 截圖)→ commit;分段請使用者實測

### D — 驗收
- [ ] 無 debug 鍵全程可破關(照 docs/66 走)+ 使用者實測確認

## 紀律

- **完整性 > 投報(HARD,rulebook 83)**:不准以「低投報」跳過任何流程步驟;卡關換方法(靜態/錄影 oracle/實機對照)。
- **單元綠 ≠ 玩得通**:GAP 關閉標準是「正常玩家路徑可達 + oracle 一致」,非測試綠。
- **版權**:影格/截圖不入公開版控(dq3_real_video/ 已 gitignore)。
- **模型分工(rulebook 45)**:haiku=抽格/機械盤點;sonnet=文件合成/移植/code 盤點;Fable=協調/RE/oracle 判讀/獨立核實。

## FLOW-GAP 清單(B1 產出後填)

| # | 流程步驟 | 原版行為(出處) | ebitan 現況 | 嚴重度 | 狀態 |
|---|---|---|---|---|---|
| — | (待 B1) | | | | |
