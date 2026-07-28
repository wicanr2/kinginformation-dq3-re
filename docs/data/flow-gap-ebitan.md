# Ebiten FLOW-GAP 初版盤點（已封存）

> 這是 2026-06-27 的歷史快照，**不可用來判斷目前程式狀態**。原 78 列長表已刪除以避免
> stale 結論反覆進入 context；完整內容仍可從 Git 歷史取得。

初版盤點曾正確指出流程層未移植，但其中大量敘述已被後續實作或 IDA/原版影片證據推翻，例如：

- 新遊戲、酒場、日夜、NPC、船與主線事件仍為零起點；
- 終盤只能靠 `Z/U/T/R` 或 synthetic progress；
- 六珠、不死鳥、下降、彩虹橋、索瑪入口尚未接線；
- 以 C remake 或 grep 命中数直接判定原版行为。

目前請依序讀：

1. [`../69-why-remake-looked-done-but-wasnt.md`](../69-why-remake-looked-done-but-wasnt.md) — 失敗原因與防再犯方法。
2. [`../74-ebiten-remake-completion-plan.md`](../74-ebiten-remake-completion-plan.md) — 唯一 current plan。
3. [`../73-gap-fix-worklist.md`](../73-gap-fix-worklist.md) — 已完成事件切片及短期續做點。
4. [`../66-original-flow-oracle.md`](../66-original-flow-oracle.md) — 原版流程證據。
5. [`../75-phoenix-orbs-re.md`](../75-phoenix-orbs-re.md)、
   [`../76-baramos-gaia-re.md`](../76-baramos-gaia-re.md)、
   [`../76-r5-endgame-realignment.md`](../76-r5-endgame-realignment.md)、
   [`../77-r5b-castle-aftermath.md`](../77-r5b-castle-aftermath.md) — 近期 IDA/影片交叉追蹤結果。

需要回答「當時為何會判成缺口」時，才查看本檔的 Git 歷史；不要把歷史計數複製回 current plan。
