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
| ✅ | B1 沙曼歐莎怪力魔→變身杖 | EXE `0x5682..0x5732` + D3MNS | Ebiten 正式 item trace/save-load | 2026-07-28 完成事件切片 |
| ✅ | R-3 六珠祭壇→拉米亞飛行 | CTY70 + EXE `0x7425..0x7599` + 本機原版影片 | 正式調查／對話／搭乘／降落 trace | 2026-07-28 完成事件切片 |
| ✅ | R-4 巴拉摩斯→索瑪現身→下降 | EXE/CTY65/72/77 + 原版影片 | 正式交談、戰鬥、王座事件與 CTY 轉場 | 2026-07-28 完成事件切片 |
| ✅ | R-5a 龍女王／彩虹橋／索瑪最終戰接線 | IDA `sub_15E02`,`loc_14243`,`sub_164CD..1661E` + CTY67/90 + 原版影片 | 光之珠 flags、精確橋座標/tile、CTY90 sec4→5、正常話す三 formation；移除 T/R/Z/Enter/Cancel 後門 | 2026-07-28 完成並測試 |
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

**暫停前那 2 個 subagent 都已完成並收尾**(tree 乾淨):
- ✅ **R-2**：精確 handler 已追到；玩家 gate `(14,7)`、無 facing，正式 item-use、怪89、
  勝敗回滾、變身杖、白天與存讀檔均已完成。
- ✅ **C-parity A2+A3**:C 版文字插值 + 權重隻數已獨立核實(18/18 測試綠)+ commit 8391d37。另修到更深 bug(C dq3_dialogue.c 平行渲染器實際遊戲從沒插值,連 VAR_NAME 都死的)。

**下週 resume 步驟**:
1. R-3 已完成；精確更正與驗證見 `docs/75-phoenix-orbs-re.md`。
2. R-4/R-5a/R-5b 已完成核心資料接線：王座隱藏樓梯、歐里狄加 rec69/70 與
   e0→0e 旗標交易、戰後回拉達多姆 handler74 冊封。後續補歐里狄加固定戰鬥動畫、
   非戰鬥傳送咒文選單，並納入無 debug 全流程通關稽核。
3. 其餘延後項:C1 EBG 音效、C2 攻方減半 0xd4f、戰鬥訊息列渲染(需字型策略)、C-parity A1(玩家狀態咒文回補 C)。
4. 建議 playtest:C 的 VAR_NAME 主角名對話現形(C-parity 修了原本靜默壞掉的路徑)。
