# 189 — 開場抽樣對拍與現行完成度仲裁

2026-10-02追加勘誤（MOTHER-FINISH-38-INPUTS）：下方rec80／rec79的兩段文字、
(21,19)提示定位及V3聲明已被原版冷啟動38次正式輸入的收據推翻。
原版101CE只有單一record80；提示為主角(22,19)、母親(22,18)，後三步到(21,17)
才交易旗標。現行有限規格與原始位址見[docs/188](188-opening-escort-to-castle-spec.md)。
舊表保留形成史，城鎮完整RGB仍未通過。

日期：2026-08-24。目的不是用兩張圖取代全遊戲 V3，而是檢查 `docs/188` 修正後的玩家可見
結果，並為本輪 release 留下可回查的已知差異。

## 證據與擷取口徑

- 原版來源：`dq3_real_video/YTDown_YouTube_Media_J_fozjiKTB8_001_1080p.mp4`，SHA-256
  `3c77abfdaddd3eb809c0cc1906c9c2d7698a0931061643d52494e183b3036802`。
- remake：原抽樣使用 2026-08-25 schema `0.1.52`／content `0.1.58`；同日後續勘誤升至
  schema `0.1.53`／content `0.1.59`，以
  `TestDumpNewGameScreens` 的現行 renderer 輸出；尚未提交，故不虛構 commit 定位。
- 原版影片來源經縮放後為 `640×480`，遊戲的邏輯畫布仍是 `640×350`，置於約
  `y=65..414`；舊版文件誤裁為 `640×400`，不得再據此宣稱 viewport 差異。
- 原版與 remake 圖只保留在 `/tmp`，不加入 Git。SHA-256：rec80 原版
  `57ec3d523a8e02645e7026e5e539ac3ad83706322efb4f911a492dc8f631461b`、remake
  `79de34dbca2465f8e020ba9b9280f7b62de3a8e24dedc486ef03fe9dbdd12f49`；rec78 原版
  `232662d0fdc66b2c5a7d150cdd13743df980d43ac2a5c2b27653f028836d0f50`、remake
  `5ee26e53ec948d2253ef2b863b2f22016759d7ad8c66e88b79d7db475ece1d2e`。

## 抽樣結果

| checkpoint | 對拍類型 | 已確認相符 | 玩家可見差異 | 判定 |
|---|---|---|---|---|
| rec80／rec79 王城指引 | same-route；原版約 `5:17–5:24` | 對話觸發點為 `(21,19)`；IDA、原始 D3TXT 與影片閉合兩段順序，關閉後續走至城門；BLK 色號對齊 | 主角 sprite 外觀與水平鏡頭仍不同；每格 6 frames 與逐 glyph wall-clock 仍為近似 | 流程 E3；文字 record／cell layout V3，整體動態畫面 V2 |
| rec78 首次謁見 | near-state；原版約 `5:34` | 原版與 remake 都是 `640×350` 邏輯畫布；國王、勇者及大臣相對構圖接近 | 原版場景左右各約 32 px inset，remake 鋪滿寬度；角色 sprite、文字起始 glyph 與分頁仍待同頁抽樣 | 事件 E3；構圖 V2；不是 V3 |

## 完成度仲裁

1. `opening_escort` 的兩階段控制權、抵達順序、rec80→rec79 與 completion flags 已由 component、
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
- `TestDumpNewGameScreens`：成功輸出本次所需 rec80／rec78 PNG；其後在既有誘惑洞窟 fixture
  失敗，因此整支 dump 測試仍誠實記為 FAIL，只有前述輸出經雜湊與目視抽驗。
