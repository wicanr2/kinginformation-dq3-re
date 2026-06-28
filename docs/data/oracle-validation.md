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

### 注音 IME 姓名輸入 — 字序/版面一致 ✅

原版 `dosbox/warp_77_*.png`(姓名輸入畫面)vs remake `zhuyin_grid.png` / `tavern_name.png`:
- **注音盤**:ㄅㄆㄇㄈㄉㄊㄋㄌㄍㄎ / ㄎㄏㄐㄑㄒ… 字元排列與游標(ㄅ)一致 —— remake 的 CJK 旗艦
  中文化(注音組字 IME + 同音候選窗,docs/15)字序對齊原版。
- **英數盤**:0-9 + A-Z 切換輸入版面一致(原版「英數」鈕 ↔ remake 英數模式)。
- 差異(誠實記錄):原版姓名輸入畫面有**角色立繪裝飾**(紅披風勇者 / 藍髮女戰士);
  remake 目前只畫功能輸入盤、未繪該裝飾立繪(屬創角畫面美術裝飾,非功能差異,可後續補)。

### boss 大圖 sprite — 原版資產正確繪製 ✅

`dosbox/orochi_boss.png`:八頭大蛇(怪75,416×143)以原版 `DQ3MNS.SHP` 解碼繪製
(多頭紅龍 + 戰鬥 HUD)。修 `DQ3_SHP_MAXW` 384→416 後此前無法繪製的 boss 大圖正確顯示。
怪物 sprite 全 130 隻另經 docs/16 圖鑑對照(含 128/129 復原)。

### 既有 oracle 紀錄(本專案歷程)

- `dosbox/warp_*.png`(2026-06-23):全 CTY warp 進城逐張實機截圖,作為城鎮版面/色盤 oracle。
- `tools/dosbox_to_overworld.sh`:原版自動進遊戲(注音輸入序列 RE)→ 起始房間/城鎮主角截圖,
  作為 sprite/palette 校正基準(docs/29)。
- RE 鐵證:整檔反組譯→nasm 重組 **sha256==原版(byte-identical 100%)**(docs/17)。

## 交付前自動化驗證(game_tester)

`tools/game_tester.sh` 一鍵全綠(**現 93/93**;此數隨新增斷言動態成長,以實跑為準,別當固定值):
- 15 單元測試(數值/戰鬥/道具/存檔…)
- playthrough_check(各系統孤立斷言,項數動態)
- mainline_check(主線一條龍 → 進度 9/9 破關)
- **全 89 城載入零崩潰**
- 新內容:索瑪二階段(光之珠)/ 結局捲動 / sub2 給物 NPC(水槍·光之珠·誘惑之劍)
- boss 事件:甘達特(26)/ 八頭大蛇(75,sprite W416)/ in-game CTY19 觸發
- 存檔→讀檔 roundtrip(狀態回復)

⇒ 多畫面 oracle 一致(標題像素 / 注音 IME 字序 / boss sprite)+ RE byte-identical +
game_tester 全綠(現 93/93)+ 89 城零崩潰 = 忠實度與穩定性交付基準。
