# 47 — C/SDL remake 剩餘工作舊盤點（已封存）

> 本檔曾是 C/SDL remake 的階段性盤點，包含已推翻的「早期 build 未接」「debug 串接即可」、
> 舊測試數量和已完成待辦。這些內容不再代表 Go/Ebiten 專案現況；原文保留於 Git 歷史。

目前專案目標是完成 `dq3_remake_ebitan/`。請改讀：

- [`74-ebiten-remake-completion-plan.md`](74-ebiten-remake-completion-plan.md) — current plan。
- [`73-gap-fix-worklist.md`](73-gap-fix-worklist.md) — 近期完成與續做點。
- [`69-why-remake-looked-done-but-wasnt.md`](69-why-remake-looked-done-but-wasnt.md) — 為何舊的
  「可破關／完成」判定不成立。

C/SDL 版仍可作 parser、資料結構與機制的參考實作，但不得把它的初值、事件接線、debug path
直接視為精訊原版規格。
