# 62 — Go／Ebitengine Android port：歷史決策摘要

> 本文件原為 2026-06-28 動工前的估算。原文包含「困難 RE 已完成」、「C/docs 可作
> oracle」和 3–6 週完成等錯誤前提，已刪除以免污染現況判讀；完整內容仍在 Git 歷史。

仍有效的決策：

- Go／Ebitengine 是現行 remake 產品線，桌面與 Android 共用 `game/` core。
- `dq3_remake/` C／SDL2 版只作研究線索，不是規格或忠實度來源。
- Go 便於共用桌面、行動與 WASM 的渲染、輸入和音訊程式，但不會自動解決遊戲內容正確性。
- 原版 oracle 是 `DQ3.EXE`、原始資料、DOSBox 同狀態實機與本機完整影片。
- Android 的 APK／AAB、觸控、安全區、背景恢復、音訊與存檔仍需真機驗收。

現行完成定義、證據分級與工作順序只看
[`74-ebiten-remake-completion-plan.md`](74-ebiten-remake-completion-plan.md)。
