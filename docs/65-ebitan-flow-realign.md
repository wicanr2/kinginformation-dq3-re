# 65 — Ebiten 流程對齊戰役（已封存）

> 本檔原為 2026-07-03 的執行中計畫。其 U1/U2/U3、78-gap 統計、C-1～C-7 待辦與「唯一進度
> 真相」宣告都已過期；長篇內容已刪除以免後續誤判，原文仍在 Git 歷史。

這一階段留下的耐用結論只有：

- remake 當時「機制很多、流程不像原版」的主因，是玩家流程層及設定資料沒有對齊。
- 完成必須由新遊戲經正式輸入、不用 debug 到 THE END。
- 原版影片/DOSBox 決定可見行為，IDA 與 raw data 決定參數、caller 和副作用。
- C remake 只能當移植線索，不能直接當原版 oracle。

目前權威文件：

- 失敗經驗與固定方法：[`69-why-remake-looked-done-but-wasnt.md`](69-why-remake-looked-done-but-wasnt.md)
- 原版流程 oracle：[`66-original-flow-oracle.md`](66-original-flow-oracle.md)
- 現況與完成計畫：[`74-ebiten-remake-completion-plan.md`](74-ebiten-remake-completion-plan.md)
- 已完成切片／續做點：[`73-gap-fix-worklist.md`](73-gap-fix-worklist.md)
