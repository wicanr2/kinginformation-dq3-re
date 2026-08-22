# dq3_remake_ebitan

精訊版《勇者鬥惡龍 III》的 Go／Ebitengine remake，也是本 repo 唯一現行產品線。

2026-08-12 發版 checkpoint 的 Docker＋Xvfb 正式 `InputState` trace 曾由新遊戲重播至
`THE END`（63.92 秒），並通過完整 `game` 的 288 個頂層回歸測試，因此該 checkpoint 為
campaign E3。2026-08-22 未發版工作樹依 [`docs/153`](../docs/153-monster-action-target-rng-order-re.md)
修正怪物 action RNG 次序後，monster77／51 玩家路線 blocker 已由 `docs/154..178` 的正式
交易與輸入切片越過；最新乾淨 Docker＋Xvfb trace 以 115.629 秒抵達 `THE END`，現行
campaign 為 E3。這不否定仍未知語意，也不把畫面／音效升格為 V3。
現行未發版 game pack 為 schema `0.1.43`／content `0.1.48`；其中玩家／敵人成功逃跑
已依原版 cue 13／21 與完成等待鏈接成 D3／E2，硬體 wall-clock 與逐幀同步仍非 V3，詳見
[`docs/149-battle-flee-sfx-wait-spec.md`](../docs/149-battle-flee-sfx-wait-spec.md)。
玩家一般物理攻擊的 cue 6 → cue 11 有序完成等待另見
[`docs/150-player-physical-sfx-sequence-spec.md`](../docs/150-player-physical-sfx-sequence-spec.md)；
敵方 attack、雙方命中／miss 與個別倒下訊息另見
[`docs/151-common-physical-result-sfx-spec.md`](../docs/151-common-physical-result-sfx-spec.md)；
cue9 特殊分支與 action3／持久麻痺已由
[`docs/152-special-physical-paralysis-spec.md`](../docs/152-special-physical-paralysis-spec.md)
獨立閉合；critical、咒文傷害及硬體 wall-clock 仍未由這些切片外推。戰鬥道具 selector、
持有人、raw item dispatch 與目標 UI 仍未閉合，不能由滿月草 field handler 外推完成。
先前船／地表追跡逾時與
Kandar 塔 fixture 全滅均已釐清為測試路徑問題並修正；詳見
[`docs/125-acceptance-20260811.md`](../docs/125-acceptance-20260811.md)。這不代表原版 V3
畫面與音效 parity：創角能力確認的固定 checkpoint 已完成 V3 靜態對拍，其他逐畫面／
逐動作／PCM 對拍及 Android 動態驗收仍待完成，詳見
[`docs/126-newgame-confirmation-v3-static-comparison.md`](../docs/126-newgame-confirmation-v3-static-comparison.md)。
現況與工作順序以
[`docs/74-ebiten-remake-completion-plan.md`](../docs/74-ebiten-remake-completion-plan.md)
為準；不要從本文件的歷史清單推算完成度。

## 建置

原版素材不納入 repo。請自行提供合法持有的素材至 `assets_raw/`，或用
`DQ3_ASSETS` 指向素材目錄。

```bash
bash dq3_remake_ebitan/build.sh
```

桌面執行：

```bash
cd dq3_remake_ebitan
DQ3_ASSETS=/path/to/assets_raw go run .
```

遊戲設定採 versioned game pack。未指定時使用執行檔內建、經測試的 canonical
`dq3_cht`；開發時可改外部 JSON 而不重新編譯：

```bash
DQ3_ASSETS=/path/to/assets_raw \
DQ_GAME_PACK=/path/to/dq3_cht \
go run .
```

外部 pack 會嚴格檢查 schema、未知欄位、ID、範圍與 evidence；格式見
[`docs/84-game-pack-json-contract.md`](../docs/84-game-pack-json-contract.md)。
套用不同 content hash 的新格式存檔不會被靜默載入。

圖形測試需在 Xvfb 下建立 `game.test`，並從 `dq3_remake_ebitan/game/` 執行，否則
測試相對路徑不會找到 repo 的 `assets_raw/`。素材缺失造成的 `SKIP` 不算對拍通過。

Android 綁定入口在 `mobile/`，建置腳本是 `build-android.sh`；它與桌面版共用
`game/` core。Android 專屬 UX（保存 640×350 原畫、橫向沉浸式畫面、中文 App 身分、
生命週期與觸控安全區）見 [`docs/124`](../docs/124-android-ux-spec.md)。APK／AAB 的
Docker build 與真機 touch path 是獨立 gate；只有 build 成功不能宣稱實機驗收。

## 目錄角色

- `game/`：場景、事件、戰鬥、選單、存檔與 renderer。
- `internal/`：原始資料解析、公式、音訊及其他可獨立測試的模組。
- `mobile/`：Android／iOS 綁定骨架。
- `docs/`：由 production renderer 產生的檢查圖片。

## 證據規則

`dq3_remake/`、舊文件及既有 Go 程式只能作研究線索。實作參數或流程前，應以
`DQ3.EXE`、原始資料、DOSBox 同狀態實機，以及本機完整影片交叉驗證。每個功能應分別記錄：

- E：正常玩家輸入能否抵達；
- V：畫面／操作與原版的對拍程度；
- D：反組譯資料流的證據程度。

單元測試通過、內部 handler 可直接呼叫或單張截圖相似，都不能單獨標成「完成」。
