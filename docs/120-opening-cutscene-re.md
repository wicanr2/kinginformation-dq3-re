# 120 — 開機年代／巨龍過場接線（2026-08-10）

## 結論

精訊版開機不是直接顯示 `TITG.P`：原版實機順序先出現年代／巨龍／角色群像，再進標題。
過去 Go/Ebitengine 只載入 title、attract、`FIRST.SCR` 與 `TIT3.P`，因此玩家第一幀就看
標題，這是實際的可見缺口，不是單純文件描述問題。

本輪把已定位的 PCX 以 versioned pack `opening` sequence 接入共用 engine：

| 順序 | pack asset | 原版內容 | 停留 | 證據等級 |
|---:|---|---|---:|---|
| 1 | `opening_year` → `TITA.P` | `1989` 年代卡 | 120 frames | strong |
| 2 | `opening_dragon` → `TITB.P` | 單色巨龍／勇者 | 120 frames | strong |
| 3 | `opening_party` → `TITD.P` | 三人灰階群像 | 120 frames | strong |
| 4 | `opening_year_1993` → `TITE.P` | `1993` 年代卡 | 120 frames | strong |
| 5 | `opening_crest` → `TITP.P` | 太陽翼紋章候選過場 | 120 frames | hypothesis |

素材完整性由 `manifest.json` 保存檔名、大小與 SHA-256；PCX 仍由共用
`dq3data.DecodePCX` 解碼，renderer 不知道 DQ3 檔名或年份文字。`OpeningSequence` 只提供
`skip_on_input`、frame asset key、停留幀與 evidence，沒有把畫面內容或任意程式碼塞入 JSON。

## 正常入口與驗收

- 桌面 `main.go` 與 mobile `init` 在 `NewGame` 後呼叫 `StartOpeningCutscene()`。
- 無輸入時依 pack frame 推進；任何正式方向／確認／取消／設定輸入會結束過場，並把同一
  按鍵交回標題 splash，因此不會吞掉玩家進入主選單的第一個 `Confirm`。
- `renderFrame` 在 attract 與 title splash 前消費 opening frame；測試 fixture 若只要直接
  操作標題，可不呼叫 start method，不會暗中注入劇情狀態。
- `TestDQ3OpeningCutsceneContract` 鎖定五張 frame、manifest integrity 與 120-frame pack
  值；`TestOpeningCutsceneBootAndSkip` 鎖定無輸入推進及正式 Confirm skip；完整
  `TestOpeningProductionInputTrace` 也由同一個 start entry 進入，沒有直接改寫 game state。
- `TestDumpOpeningCutscene` 在 Docker＋Xvfb 以同一個 `renderFrame` 輸出五張 runtime PNG（輸出
  目錄為 gitignored 的 `work/opening_probe/`）：`opening_year.png`
  SHA-256=`b0eb26d8361d4056b6a9849671fd38a5e3631e16198e08b85dee4a588c3b2901`、
  `opening_dragon.png`=`68454e872d662ffc3eb3a43015346cd3f4ceeb813516381fe60a6f2660e3c81f`、
  `opening_party.png`=`84ba13977250590caf8819c33e7716b2cbcb29abe08e11402d17d92fea487a02`、
  `opening_year_1993.png`=`9e19a67fb081f8cf3b1e7c10c717f4e15f7fcb8a5cb9bf23dd3ecb15c0d8715d`、
  `opening_crest.png`=`f514f904062ca29b07103441aa8d8e1f67c0bd1c91d600ac6aeaaab21458414b`。
  這些 hash 只證明 runtime 重播穩定，不是原版 timing／排序的證據。

## 推論邊界

- `TITA.P` 對 `1989`、`TITD.P` 對三人群像、`TITB.P` 對巨龍／勇者及 `TITE.P` 對 `1993`
  有原版實機／已解出參考圖支持；PCX identity 可標 **strong**。
- 本輪沒有可逐幀回放的完整原版影片 frame index；120 frames 是可重播的保守 pack 值，
  不能標成原版 timing confirmed。`TITP.P` 是否位於標題前最後一張仍是 **hypothesis**，
  需下一輪用本機完整影片逐幀核對後，才可升降序或移除。
- 因此此切片關閉「開機直接跳標題」的 E2／V1 缺口，但不是完整 cutscene V3；淡入／淡出、
  音效 cue 與精確 skip 行為仍列 `docs/74` 的視聽長尾。
