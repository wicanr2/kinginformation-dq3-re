# 推廣片 r2 與 `dist-all/` 交付規格（2026-08-12）

## 現行結論

本輪依 `game-promo-video-ffmpeg` 技能重新錄製與剪輯推廣片。現行交付只認
`dist-all/v0.1.34/`：

- `patch/` 與 `full/`：Linux x86_64 AppImage、Windows x86_64 ZIP、macOS x86_64／arm64 ZIP。
- `promo/dq3-remake-promo-20260812-r2.mp4`：72 秒、1280×720、30 fps、H.264＋AAC。
- `promo/source/opening_runtime.mp4`：由正式 v0.1.34 AppRun 在 Docker／Xvfb 以正式輸入擷取的 12 秒開場來源。
- `SHA256SUMS.txt` 與 `RELEASE.txt`：同一交付樹的檔案清單、雜湊與平台界線。

`dist/` 與 `work/` 只保留歷史／中間產物；它們不再是下載或推廣片入口。完整包仍含
原版素材而只供使用者本機合法持有，不納入 Git；公開 patch 不含原版素材。

## 本輪相對前一支片的三項可查差異

1. **版型**：由單一黑底文字框改為全幅字幕帶、金框卡片、置中資料卡與八頭大蛇雙畫面，
   每個段落仍使用固定畫格＋淡入淡出，沒有 `zoompan` frame explosion。
2. **敘事**：保留 12 秒實機開場，後續依母親／王城／地表／日夜／航海／解謎／建城／
   兩場八頭大蛇／巴拉摩斯／128／129／不死鳥／索瑪／THE END 排成完整冒險弧線。
3. **聲音**：沿用本機實際 MT-32 render，五段曲目以一秒 `acrossfade` 串成 72 秒，
   以 `volumedetect`、`silencedetect` 與完整影音解碼驗收；不是只檢查音訊串流存在。

## 可重現流程

1. 在 Docker／Xvfb 執行 `tools/capture_ebiten_promo_opening.sh`，輸出
   `dist-all/v0.1.34/promo/source/opening_runtime.mp4`。
2. 在同一類型的推廣片 Docker image 執行 `tools/build_ebiten_promo_video.sh`；缺少
   任一 runtime 圖、實際音訊或字型即失敗。
3. 以 `tools/verify_promo_video.sh` 驗收影像尺寸、AAC 48 kHz 雙聲道、非靜音、無長段數位
   靜音及完整解碼。
4. 執行 `tools/collect_dist_all.sh`，從已驗證的 `dist/v0.1.34/` 收集三平台 patch/full，
   並更新 `SHA256SUMS.txt`。所有容器使用一次性 `--rm`、目前 UID/GID 與外部逾時；結束後
   `docker ps` 不得留下專案工作容器。

## 證據限制

影片是推廣片交付，不把 runtime 靜態圖擴張為全遊戲逐幀 V3 parity；畫面與事件狀態仍以各
自的 runtime／原版證據文件為準。音訊技術驗收證明封裝可聽與非靜音，不等同於完成所有原版
PCM 波形與玩家可見 timing 的逆向閉環。
