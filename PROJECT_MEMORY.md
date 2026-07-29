# DQ3 Go/Ebiten remake 接手記憶

> 更新：2026-07-29。接手先讀 `CLAUDE.md`、`CONTEXT.md`，再讀
> [`docs/74-ebiten-remake-completion-plan.md`](docs/74-ebiten-remake-completion-plan.md)。
> 本檔只保存不易過期的決策；逐項狀態不要在此重複維護。

## 目標與完成定義

現行產品是 `dq3_remake_ebitan/`。玩家須由全新遊戲開始，只用正式遊戲輸入，依精訊版流程抵達
THE END；關鍵事件的入口、設定資料、畫面、聲音、副作用與 save/load 均有原版證據。

以下都不算完成：編譯成功、單元測試全綠、直接呼叫內部事件函式、用 debug 鍵/env 串場景，
或只證明 parser/handler 存在。

## 證據與產品線

- 本機原版影片／DOSBox：玩家可見流程與操作 oracle。
- IDA、原始 EXE/CTY/文字/資產：caller、設定值與副作用 oracle。
- 攻略：路線與定位線索，不代替原始資料。
- `dq3_remake/` C/SDL 版：機制與 parser 參考，不是原版 oracle。
- 舊 RE/進度 Markdown：線索；若與近期證據或 code 衝突，以證據重查。

失敗原因、事件契約與驗收紀律集中在
[`docs/69-why-remake-looked-done-but-wasnt.md`](docs/69-why-remake-looked-done-but-wasnt.md)。

## 現況入口

- 唯一 current plan：[`docs/74-ebiten-remake-completion-plan.md`](docs/74-ebiten-remake-completion-plan.md)
- 原版流程 oracle：[`docs/66-original-flow-oracle.md`](docs/66-original-flow-oracle.md)
- 最新 production trace：[`docs/79`](docs/79-najimi-production-trace.md)
- 近期 IDA/影片證據：[`docs/75`](docs/75-phoenix-orbs-re.md)、
  [`docs/76a`](docs/76-baramos-gaia-re.md)、[`docs/76b`](docs/76-r5-endgame-realignment.md)、
  [`docs/77`](docs/77-r5b-castle-aftermath.md)

目前核心終盤切片已接通，boot 起的正式 trace 已到 CTY01 魔法球並通過存讀檔，但
**尚未完成從新遊戲開始的無 debug 全流程驗收**。下一步由該 checkpoint 正常走到誘惑洞窟，
以魔法球破牆並抵達羅馬利亞；不得跳回孤立 handler 或把後段事件 checkpoint 測試提升成
campaign E3。

## 固定工程方法

1. 每輪閉合一個垂直切片：正式入口 → 狀態交易 → 畫面/聲音 → 下一節點 → save/load。
2. 未知值 fail-closed；禁止用 C remake、經典 DQ3 慣例或「合理值」填 production。
3. 每個參數追原版「誰在何時寫入」，並記錄明確位址口徑。
4. component test 與 production-input trace 分級；只有正常玩家可達才可關閉主線 GAP。
5. 進度只維護 `docs/74`；歷史計畫保留短索引，細節從 Git 歷史取回。

## 工作樹注意

接手先執行 `git status --short`。未追蹤的 Android libs、scratchpad 與 `tools/dosbox_verify_*`、
`tools/tmp_*` 可能是使用者檔案；不得刪除、覆寫或誤納入 commit。
