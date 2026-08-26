# 195 — 場景音樂恢復與設定熱切換 polish

日期：2026-08-26。程式基準：v0.1.36 發行後的 `main` 工作樹；schema `0.1.53`、
content `0.1.59` 未變更。

## 玩家可見問題

既有 runtime 在戰鬥 modal 關閉後固定呼叫 `field` cue。這會讓城堡、城鎮及迷宮內的
戰鬥結束後錯播地表曲。系統設定把音樂切為關閉時會暫停目前 player，但重新開啟只改
enabled 狀態，沒有重新播放玩家當前場景。

這兩項是 remake 自身的路由缺口，不需要新增原版猜測：七個具名場景 cue 與 pack 軌號已由
[`docs/61`](61-music-scene-mapping.md)／`audio.json` 保存，戰鬥開始及各場景入口亦已有正式
consumer。修正只把「結束／重新啟用」接回同一個具名場景 selector。

## 修正

- `currentAudioCue()` 以 ending → battle → title → town scene → field 的 modal／場景優先序
  回傳穩定 cue ID；實際 track 仍由 game-pack 決定。
- 一般戰鬥、逃跑、敗北回城及劇情戰結束後，依結算後的實際 `inTown/curCty` 狀態恢復
  castle／town／dungeon／field，而非固定 field。
- 設定選單由 OFF 切回 ON 時立即恢復當前 title／battle／scene／ending cue。
- `gameAudio` 介面只抽象跨版本播放契約，正式實作仍是 `gaudio.Music`；測試可記錄 pack
  最終解析出的 track，不建立主機音訊裝置。
- 訂正 `settings.go` 過期註解：正常產品固定使用 OGG；`DQ3_FM` 只是明確診斷入口。

## 驗收與證據等級

- Docker＋Xvfb 的 audio／music／sound／settings／battle cue 比例回歸通過：
  `internal/gaudio`、`internal/gamepack`、`game`。
- 路由測試涵蓋 title、field、castle、town、dungeon、battle、ending；並確認迷宮恢復解析為
  pack track 3、城堡重新啟用解析為 pack track 1。
- 七個正式 OGG 再次通過完整解碼、Vorbis／32 kHz／雙聲道、音量、峰值與長靜音 gate；
  數值及 SHA-256 與 [`docs/193`](193-audio-sampling-polish-20260825.md) 相同。

本批把音訊技術抽樣與主要 runtime 轉場提升為 **E2**。沒有 Windows／macOS 實際裝置人耳
結果，也沒有把 EBG 18–23 或每個戰鬥動作停頓全部接成原版同狀態 oracle，因此維持非 V3；
DAC／PIT／DMA wall-clock 依既定停止線不重新研究。

## 自評影響

「音樂、音效與時序」由 78% 調為 84%。固定權重下總分由 89.65% 增至 90.25%，對外仍取整
為 90%；這次提升代表主要場景恢復路由閉合，不代表全音訊逐聲 exact。
