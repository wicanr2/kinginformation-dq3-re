# 73 — 「兩版都缺」缺口修正 worklist(2026-07-03)

> 來源 `docs/72`(3 subagent 盤點 + 協調者核實)。聚焦 **ebitan**(可玩 remake 主目標);
> C 端同缺者標「C 另補」。協調者(Opus)派 sonnet 逐批實作,**每批獨立核實才收**
> (重跑完整套件 game+internal 全綠 + 視覺/數量 dump + 抽查關鍵邏輯,rulebook 45/65)。
> 順序:先乾淨高值、後大工程;各批多動 game.go 故**序列執行**(不並行,避免衝突)。

| 批 | 缺口 | 範圍 | model | 狀態 |
|---|---|---|---|---|
| **W1 ✅** | A2 文字插值(道具名/數值/名字) | ebitan dialogue.go 插值 + 事件填 var(commit 52ce817;record191『力量增加銅劍。』視覺驗證) | sonnet+協調者接完 | 完成 |
| **W2** | A3 怪物群隻數權重預算 | ⚠ ebitan 戰鬥現單敵,需重構 battle.go 成 N 敵(C dq3_battlescene 已多敵可參考)+ weight 隻數 | sonnet | 進行中 |
| W3 | A1 玩家狀態/buff/debuff 咒文(拉里荷/拜基魯多/史卡拉…rec143-160) | ebitan 狀態效果基礎設施(battle) + spell defs 擴充 + 玩家施放 + 敵對玩家下狀態 | sonnet | 待(最大)|
| **W4 ✅** | A4 酒場兩段式招募(登錄名冊≠入隊) | ebitan roster 名冊 vs party 分離,不再即建即入隊/頂替 | sonnet | 完成(commit 5507a0a)|
| **W5 ✅** | D. C oracle NPC 三 bug 回補(b2<4丟棄/8-slot/story-flag過濾) | **C 端** dq3_scene.c;移植 ebitan 已修的三項,恢復 oracle 忠實 | sonnet | 完成(commit f27b51d)|
| — | B1 沙曼歐莎怪力魔→變身杖 | 需先 R-2 RE(座標)| — | 延後(R 系列)|
| — | C1 EBG 音效 cue / C2 攻方狀態減半 0xd4f | 低信心/RE 未定 | — | 延後 |

## 共通驗收(每批)
- 完整套件 `go test ./game/ ./internal/... -count=1` **game+internal 全綠**、gofmt/vet clean。
- 視覺/數值 dump 肉眼或數字核對原版行為(rulebook 65)。
- 不碰他批範圍;不 commit(交協調者核實後 commit)。
- 誠實回報:改法、驗證、待人工項,不膨脹。
