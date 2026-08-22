# 精訊版 DQ3 專案指引

現行交付目標是完成 `dq3_remake_ebitan/` 的 Go／Ebitengine remake。`dq3_remake/`
的 C／SDL2 程式和 `re/` 的 C 重建是歷史研究與比對材料，不代表原版規格，也不是目前
要完成的產品。

開始工作前依序讀：

1. `PROJECT_MEMORY.md`
2. `CONTEXT.md`
3. `README.md`
4. `docs/74-ebiten-remake-completion-plan.md`
5. 與當前子系統直接相關的 RE 文件

## 事實與完成度

- 原版是 16-bit real-mode、large-model DOS 程式；不是 Borland Pascal 或保護模式程式。
- 不可把 C remake、舊 Markdown、測試綠燈或可執行切片當成原版 oracle。
- 機制存在不代表設定、資料、gate、時序、畫面與副作用已和精訊版一致。
- 反組譯結論需追通 `writer → table/state → consumer → 玩家可見副作用`，並標明
  `file offset`、`logical` 或 `DGROUP` 位址口徑。
- 只有從新遊戲開始、不用 debug shortcut，依正式玩家輸入抵達 THE END，且關鍵存讀檔、
  畫面與事件對拍通過，才可宣稱 remake 完成。

原版驗證、DOSBox、RE、IDA、搜尋、測試與遊戲執行一律只能在一次性 Docker 容器內進行；
主機不得直接啟動工作負載。RE 必須優先把 `/home/anr2/ida_94_official/dist` 唯讀掛入
專用 image，使用 IDA Pro 9.4＋IDAPython；`tools/dis.sh`／Ghidra 只作批次定位或交叉驗證。
不得把原版素材、IDA database／授權或發佈包加入 Git。

專案 repo：<https://github.com/wicanr2/kinginformation-dq3-re>
