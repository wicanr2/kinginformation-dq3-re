# DOSBox 原版 oracle 驗證紀錄(忠實度)

> 專案北極星:remake「與原版一模一樣,使用 dosbox 驗證」。本檔記錄交付前的 oracle 比對結果。
> 版權畫面像素本身留本地(`dosbox/` gitignore),此處只記方法與結論。

## 方法

- 原版:`dq3-dosbox` 容器(Xvfb + DOSBox)跑 `work/dq3_fixed_game`(patched 原版),
  `tools/dosbox_run.sh` 腳本化按鍵 + `import` 截圖。
- remake:同畫面以 headless dump(`DQ3_DUMP`)渲染 PPM→PNG。
- 並排比對(ImageMagick `convert`)。

## 結果

### 標題畫面(TITG.P)— 逐像素一致 ✅

`dosbox/oracle_title_compare.png`(原版上 / remake 下):
- logo「DRAGON FIGHTER III」、副標「傳說的終章」、「© 1993 精訊資訊有限公司」完全一致。
- 漸層天空、城堡綠色剪影、洋紅地面、劍+寶珠 logo 配色/版面一致。
- remake 為原生 2×(640×400);原版截圖為 DOSBox letterbox 放大(1024×768)故水平略擠,
  內容像素與調色盤一致。

### 既有 oracle 紀錄(本專案歷程)

- `dosbox/warp_*.png`(2026-06-23):全 CTY warp 進城逐張實機截圖,作為城鎮版面/色盤 oracle。
- `tools/dosbox_to_overworld.sh`:原版自動進遊戲(注音輸入序列 RE)→ 起始房間/城鎮主角截圖,
  作為 sprite/palette 校正基準(docs/29)。
- RE 鐵證:整檔反組譯→nasm 重組 **sha256==原版(byte-identical 100%)**(docs/17)。

## 交付前自動化驗證(game_tester)

`tools/game_tester.sh` 一鍵全綠(**23/23**):
- 15 單元測試(數值/戰鬥/道具/存檔…)
- playthrough_check 35 項(各系統孤立斷言)
- mainline_check(主線一條龍 → 進度 9/9 破關)
- **全 89 城載入零崩潰**
- 新內容:索瑪二階段(光之珠)/ 結局捲動 / sub2 給物 NPC(水槍·光之珠)
- 存檔→讀檔 roundtrip(狀態回復)

⇒ 標題逐像素一致 + RE byte-identical + 23/23 自動驗證 + 89 城零崩潰 = 忠實度與穩定性交付基準。
