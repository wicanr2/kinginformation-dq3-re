# 189 — 開場抽樣對拍與現行完成度仲裁

日期：2026-08-24。目的不是用兩張圖取代全遊戲 V3，而是檢查 `docs/188` 修正後的玩家可見
結果，並為本輪 release 留下可回查的已知差異。

## 證據與擷取口徑

- 原版來源：`dq3_real_video/YTDown_YouTube_Media_J_fozjiKTB8_001_1080p.mp4`，SHA-256
  `3c77abfdaddd3eb809c0cc1906c9c2d7698a0931061643d52494e183b3036802`。
- remake：commit `27a66ec8a8d6f4cb0b0405e73948b75e50a336ef`、schema `0.1.51`、content
  `0.1.57`，以 `TestDumpNewGameScreens` 的現行 renderer 輸出。
- 原版影片 frame 為 `640×480`，其中遊戲畫面約在 `y=40..439`；remake 邏輯畫布為
  `640×350`。並排檢查只把原版裁成 `640×400`、remake 底部補黑到 `640×400`，沒有縮放
  或把這項處理當成遊戲幾何證據。
- 原版與並排圖只保留在 `/tmp/dq3-parity-27a66ec/`，不加入 Git。下列 SHA-256 讓本機結果
  可辨識：rec80 原版 `ef912e69...1226`、remake `89329a39...a87`；rec78 原版
  `f3c5e9ea...c2e6`、remake `e7069357...e386`。

## 抽樣結果

| checkpoint | 對拍類型 | 已確認相符 | 玩家可見差異 | 判定 |
|---|---|---|---|---|
| rec80 王城指引 | near-state；原版約 `5:18` | 主角已由家門自動抵達王城入口；rec80 內容與交還操作順序一致 | 原版背景是阿里阿罕城鎮道路、水道、花圃與住宅；remake 顯示水池中的小型城堡圖塊，地圖／鏡頭狀態顯著不符。對話框位置、寬度及換行也不同 | 流程 E3；畫面 V2，存在高可見度 parity gap |
| rec78 首次謁見 | near-state；原版約 `5:34` | 國王、勇者及右側大臣的相對角色關係一致；勇者確實可見，先前「漏畫」判讀維持撤回 | 原版是 `640×400` 可見區，remake 為 `640×350`；色盤、牆／柱細節、對話框寬度、文字起始 glyph 與分頁不同 | 事件 E3；構圖 V2；不是 V3 |

## 完成度仲裁

1. `opening_escort` 的兩階段控制權、抵達順序、rec80 與 completion flags 已由 component、
   game-pack validator 與正式創角 trace 證實，維持 E3。
2. 抽樣直接否定「開場畫面已接近原版完成」；rec80 背景是新的玩家可見差異，後續若追求
   忠實視覺，應以同一 checkpoint 追 scene/page/camera owner，不可只靠截圖手調座標。
3. 王座勇者 renderer 沒有功能缺陷；剩餘是 viewport、palette、場景細節與文字 layout parity。
4. 本輪仍可建立功能 release，但 release note／README 必須標成「campaign E3、視覺整體
   V1／V2，含已知開場背景差異」，不得宣稱完整 V3 或像素級原版還原。

## 本輪重跑

- `TestOpeningEscortValidationRejectsBrokenArrivalContract`：PASS。
- `TestOpeningMotherEscortUsesVisibleFrames`：PASS。
- `TestOpeningKingAudienceRendersHero`：PASS。
- `TestOriginalOpeningEventTransactions`：PASS。
- `TestDumpNewGameScreens`：成功輸出本次所需開場與王座 PNG；其後在既有誘惑洞窟 fixture
  失敗，與本抽樣 checkpoint 無關，不把整支 dump 測試記為 PASS。
