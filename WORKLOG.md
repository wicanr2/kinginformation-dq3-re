# DQ3 工作歷程

## 2026-10-01 — Issue #1：dosgolem 開場對拍

起始 HEAD `0c6f780`，未追蹤使用者檔案全數保留。主機 gh 登入成功，遠端原先沒有 Issue；
自動核准審查首次拒絕建立，使用者再明確授權後建立並回讀
[Issue #1](https://github.com/wicanr2/kinginformation-dq3-re/issues/1)。

dosgolem 上游 `2f44a68` 從真實 EXE 入口探測，原版輸入 hash 與工具版本見
[docs/196](docs/196-dosgolem-opening-sequence-parity.md)。未修正探測黑畫面由空白補齊的
DOS 檔名造成；可丟棄副本修正後 DOS tests PASS，原版自然執行至六張卡片與標題／巡禮。
確認 remake 漏 1990 卡且誤用候選紋章，規格審查達 READY。

環境勘誤：容器登入 shell 覆寫 Go PATH，改 `bash -c`；Xvfb-run 作為 PID1 只剩
啟動等待，依實際程序狀態停止後改有界 bash／Xvfb＋trap，相同開場測試及五張擷取 PASS。
第一次 `-shots` 的原始色號檔誤命名 `.png`，後改 `-dump-at` 由 dosgolem 平面擷取重生。
這些是工具／驗證環境問題，不記為 remake 缺陷。

首次正式第六幕比較失敗，證實 TITF 只有背景，原版另有 DFP 動態疊圖；規格退回 DRAFT，
以 IDA 9.4 追原始 caller、writer、directory、兩種平面 consumer、座標及翻頁。
獨立解碼與原版全部 129 次翻頁一致後達 READY，再接入共用引擎與資料包。
schema `0.1.54`／content `0.1.60` 的正式無輸入序列、原版前五幕色號及第六幕 129 個
位置的色號／RGB 共 30,016,000 個比較像素通過；時間採公開 PIT 公式近似，沒有深入 ISR。

全部 internal、受影響開機／存讀檔與桌面建置通過。完整 game 有母親帶路斷言及 Lv1
盜賊鑰匙 fixture 全滅兩項失敗；以修改前 `0c6f780` 在同一容器與實際素材重現相同失敗，
另登記 [Issue #2](https://github.com/wicanr2/kinginformation-dq3-re/issues/2)，不能冒稱完整回歸全綠。
本批未建立發行包，未提交原版素材或 IDA database。commit／push 與最後清理核對於 Issue #1 記錄。
