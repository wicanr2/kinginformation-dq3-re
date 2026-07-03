# 73 — 「兩版都缺」缺口修正 worklist(2026-07-03)

> 來源 `docs/72`(3 subagent 盤點 + 協調者核實)。聚焦 **ebitan**(可玩 remake 主目標);
> C 端同缺者標「C 另補」。協調者(Opus)派 sonnet 逐批實作,**每批獨立核實才收**
> (重跑完整套件 game+internal 全綠 + 視覺/數量 dump + 抽查關鍵邏輯,rulebook 45/65)。
> 順序:先乾淨高值、後大工程;各批多動 game.go 故**序列執行**(不並行,避免衝突)。

| 批 | 缺口 | 範圍 | model | 狀態 |
|---|---|---|---|---|
| **W1 ✅** | A2 文字插值(道具名/數值/名字) | ebitan dialogue.go 插值 + 事件填 var(commit 52ce817;record191『力量增加銅劍。』視覺驗證) | sonnet+協調者接完 | 完成 |
| **W2 ✅** | A3 怪物群隻數權重預算 | 多敵重構+史萊姆×6(commit 0269012;6隻橫排 dump 驗證)| sonnet | 完成 |
| **W3 ✅** | A1 玩家狀態/buff/debuff 咒文 | 5 修正狀態+玩家/敵施放(commit 13be592;拜基魯多×2/拉里荷睡眠 驗證)| sonnet | 完成 |
| **W4 ✅** | A4 酒場兩段式招募(登錄名冊≠入隊) | ebitan roster 名冊 vs party 分離,不再即建即入隊/頂替 | sonnet | 完成(commit 5507a0a)|
| **W5 ✅** | D. C oracle NPC 三 bug 回補(b2<4丟棄/8-slot/story-flag過濾) | **C 端** dq3_scene.c;移植 ebitan 已修的三項,恢復 oracle 忠實 | sonnet | 完成(commit f27b51d)|
| — | B1 沙曼歐莎怪力魔→變身杖 | 需先 R-2 RE(座標)| — | 延後(R 系列)|
| — | C1 EBG 音效 cue / C2 攻方狀態減半 0xd4f | 低信心/RE 未定 | — | 延後 |

## 共通驗收(每批)
- 完整套件 `go test ./game/ ./internal/... -count=1` **game+internal 全綠**、gofmt/vet clean。
- 視覺/數值 dump 肉眼或數字核對原版行為(rulebook 65)。
- 不碰他批範圍;不 commit(交協調者核實後 commit)。
- 誠實回報:改法、驗證、待人工項,不膨脹。

## 收尾(2026-07-04)

W1-W5 全數完成、獨立核實、commit(52ce817/0269012/5507a0a/f27b51d/13be592)。docs/72 高信心「兩版都缺」缺口(A1-A4)ebitan 全補;C oracle NPC(D 段)W5 回補。**延後項**:B1 沙曼歐莎怪力魔(需 R-2 RE 座標)、C1 EBG 音效、C2 攻方減半 0xd4f(RE 未定)。A1-A4 的 C 版對等(除 D)未回補,若要 C 對拍需另補。

## ⏸ 暫停點(2026-07-04,下週續)

**已完成並 push**:W1-W5 全數(見上表)。過期 markdown 斷言已校正(docs/60/68/71/72)。

**暫停時有 2 個 subagent 在背景跑**(結果留待下週驗收,未 commit):
- **R-2 RE 調查**(a1107e729cf7c9fc8):產沙曼歐莎怪力魔 spec —— 假國王 CTY44 座標 + 拉之鏡 0x61 USE reveal 機制。**不改碼**,回報是 spec 文字;下週我判讀後寫進 docs/70 + 派實作。
- **C-parity A2+A3**(accf33a1197859355):C 版回補文字插值 + 怪物群權重隻數。**寫 C 版未 commit**;下週我獨立核實(dq3-remake docker build+test)後 commit。

**下週 resume 步驟**:
1. 讀兩個 subagent 回報(task 通知 / output 檔)。
2. C-parity → 核實 C build+測試綠 → commit。
3. R-2 spec → 判讀「已知 vs 待定」→ spec 齊則派 ebitan 實作 subagent(reveal 邏輯 + 怪89 boss + 變身杖 0x62;isNight() 已備)。
4. 續 R 系列:R-3 不死鳥+六珠+飛行坐騎(最大)、R-4 巴拉摩斯城、R-5 龍王城+終盤 wire(拔 Z/U/Enter)。
5. 其餘延後項:C1 EBG 音效、C2 攻方減半 0xd4f、戰鬥訊息列渲染(需字型策略)。
