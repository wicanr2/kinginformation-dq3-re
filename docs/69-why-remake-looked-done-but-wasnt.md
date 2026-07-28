# 69 — 為什麼多輪反組譯仍做不出忠實 remake

> 本文件是這個問題的**唯一權威經驗摘要**。它不記錄會過期的完成百分比；現況與下一步只看
> [`74-ebiten-remake-completion-plan.md`](74-ebiten-remake-completion-plan.md)。

## 結論

先前不是「反組譯得不夠久」，而是反組譯成果沒有形成從原版證據到 production gameplay 的閉環：

```text
原版實機輸入/畫面
  → EXE 真實 caller、資料表、初始狀態
  → 明確事件契約
  → Go production 入口與設定資料
  → 同輸入、同狀態、同畫面/副作用驗收
```

過去大量時間花在 parser、公式、handler 與通用機制；這些多半真的存在、也能通過單元測試，但
**機制正確不等於餵給機制的資料正確**。出生 section/座標、隊伍、裝備、旗標、NPC 可見性、
事件 index、對話 record、formation、轉場目的地只要有一項用猜的，遊戲一開始就會和原版不同。

## 多輪失敗的具體原因

1. **把機制 parity 誤當遊戲 parity。** loader 能載 CTY、runner 能 dispatch、戰鬥能排 formation，
   不代表新遊戲使用正確 CTY、自然觸發正確 runner、或 boss 使用正確 formation。
2. **從衍生 remake 移植，而非從原版取值。** C 版的簡化、debug token、合成旗標和猜測初值被
   Go 版繼承；錯誤因「已有實作」而取得不該有的權威。
3. **靜態找到 handler，卻沒追自然 caller。** 「函式存在」被當成「玩家可達」；觸發座標、
   facing、前置旗標、輸入 owner 與呼叫時機沒有一起解出。
4. **位址基準錯誤造成漂亮但錯的結論。** logical/file/loaded segment/overlay 混用，曾把
   `logical 0x164cd / file 0x1783d` 漏掉高位 `0x1`，錯認成下降事件；實際是索瑪終戰流程。
5. **把搜尋不到當不存在。** jump table、間接 call、資料驅動 dispatch 沒有一般 xref；
   `grep 0 hits` 只代表查法沒命中，不能證明原版沒有功能。
6. **測試驗證自己的假設。** 直接呼叫 `descend()`、`startZomaSeq()` 或自行設定 flag 的測試，
   只能證明內部函式自洽，不能證明玩家從正式輸入可到達。
7. **debug 捷徑遮住流程斷裂。** `Z/U/T/R`、Enter/Cancel demo path、debug env 與 synthetic
   `progressSet` 把孤立場景串成「可破關」，正常玩家路徑其實不存在。
8. **沒有以原版逐段驗收。** 編譯、單測、headless component dump、素材解碼成功都曾被當完成；
   一次從新遊戲開始的同輸入對拍，就能立刻發現出生點、Space 語意、NPC 數量等根本差異。
9. **完成標準停在子系統，而非交易完整性。** 只查畫面或 flag，沒同時核對事件前置、消耗/取得、
   對話、場景變化、存檔持久化及下一事件可達性。
10. **文件沒有單一真相。** dated audit、舊 worklist 和新計畫並存，舊的「未接」「唯一 blocker」
    會在下一輪被重新載入，導致重查已完成項、復活已推翻的結論並浪費 context。

## 為什麼「反組譯很久，參數仍不一致」

反組譯時間主要累積在**廣度與內部機制**，不是逐事件做「設定資料 reconciliation」。沒有建立一張
表逐欄回答「原版誰在何時寫入這個值、Go 現在寫什麼、證據在哪」，時間再久也只會增加更多
已知函式，不會自動修正 `NewGame()`、地圖入口或事件表裡的猜測常數。

後來有效的做法，是針對玩家即將遇到的切片，在 IDA 由真實 caller 往下追，並用本機原版影片、
DOSBox、CTY/EXE raw data 交叉核對。這才陸續修正沙曼歐莎、不死鳥、下降、彩虹橋與索瑪入口。

## 後續固定方法

### 證據優先序

1. 同版本本機原版實機行為／完整影片（決定玩家看見與輸入順序）。
2. IDA 的 caller/xref/decompile，加原始 EXE、CTY、文字及資產表（決定值與副作用）。
3. 攻略（決定路線並協助定位，不代替程式資料）。
4. C remake 與既有 RE 文件（線索，不是 oracle）。
5. 猜測、經典 DQ3 慣例（只能標待證，不可進 production）。

來源互相衝突時不投票；回到同一狀態、同一輸入做可重現仲裁。

### 每個事件必填契約

- 玩家入口：地圖、section、座標範圍、facing、按鍵/移動。
- 前置狀態：道具、隊伍、日夜、flags、runner index。
- 原版 caller 與資料來源：IDA 位址口徑、xref、CTY/表 offset。
- 玩家可見結果：對話 bank/record、動畫、tile/NPC、BGM/SFX、戰鬥 formation。
- 狀態交易：取得、消耗、gold、flags、位置、失敗/取消分支。
- 持久化：事件前後 save/load 結果。
- 下一步：完成後玩家是否能自然抵達下一個事件。

### 實作與驗收 gate

- 禁止用「合理值」補未知參數；未知即列為待追。
- production test 必須從正式按鍵與地圖入口進入；helper direct-call 只能當單元測試。
- 每個主線切片至少保留一條 production-input trace，並對關鍵畫面做 framedump。
- 最終完成定義：全新遊戲、不用 debug 鍵/env/內部函式，依精訊版流程到 THE END；關鍵 gate
  前後可存讀，事件副作用和原版一致。
- 進度只維護一份 current plan；歷史快照縮成索引，細節需要時從 Git 歷史取回。

## 接手者五問

在宣稱任何流程「完成」前，回答：

1. 玩家不用 debug，從哪個正式輸入進來？
2. 每個設定值的原版 write/caller 或 raw table 證據在哪？
3. 測試是否可能只是在驗證 remake 自己的猜測？
4. 事件所有副作用及 save/load 都核過嗎？
5. 用相同狀態和輸入對過原版畫面與下一步可達性嗎？

任一題答不出來，只能說「機制存在」或「部分接線」，不能說 remake 已完成。
