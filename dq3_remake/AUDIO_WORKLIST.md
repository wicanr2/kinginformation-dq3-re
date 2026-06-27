# dq3_remake 音訊(SB FM / OPL2)WORKLIST

> 目標:remake 播出**精訊版自己的 SB FM 音樂**(從正版遊戲資料 RE 抽出 + OPL2 合成器播放)。
> 全盤計畫見 [`../docs/56-audio-fm-plan.md`](../docs/56-audio-fm-plan.md)。**IP 立場**:只 RE 遊戲自身資料、
> 不抄受版權樂曲;資料 gitignore 不散布、執行有原版 `DQ3.EXE` 閘。紀律同主 WORKLIST(docker、code 為真相、每階段 commit）。

## 現況(2026-06-27)

remake **目前無音訊引擎**。已 RE 偵察確認:精訊用 OPL2 FM(EXE 寫 0x228/0x229)、驅動 `SBFM_CMF_DRV`(CMF)、
音效 VOC(FVOC/NVOC.VCX)。音樂 CMF 資料位置**未定**(疑內嵌 DQ3.EXE),步驟 1 先確認。

## 階段

### Phase 1 — 定位 + 抽出 CMF 音樂資料(de-risk,先做)
- [x] **RE 偵察完成**(2026-06-27,過程見 [`../docs/57-cmf-audio-re.md`](../docs/57-cmf-audio-re.md)):
  音訊系統 = Creative CT-VOICE + CT-MUSIC(`CMFDRV`)驅動(`SBCM.LIB`)+ VOC 音效(`FVOC/NVOC.VCX`);
  音樂格式 = CMF(OPL2 FM)。**EXE 載入清單只有 VOC,無音樂檔;全檔無 `CTMF` magic;0x228 多屬資料表非埠。**
- [x] **(A) 已排除、(B) 確認**(使用者親證 + YouTube 精訊版錄影有 BGM + RE 獨立佐證):音樂**真的在、運作中**——
  EXE 重設 8253 timer(`mov al,0x36;out 0x43` ×2)+ `set_timer_count(ax)` 函式(logical 0x2c52)+ 計時 ISR。
  資料**內嵌 EXE**(非 CMF 檔、無 CTMF magic、非 overlay)。
- [ ] **Phase 1 續(進行中,deep trace)**:從 timer ISR / 音樂事件處理器反追 → 音樂事件流 + instrument base。
  - 已找:timer init(file 0x13fc3 / 0x155c0)、`set_timer_count`(logical 0x2c52)。
  - 待找:ISR 安裝點(INT8 或 SB-IRQ)→ ISR 內 OPL reg 寫入 → 事件指標變數 → 音樂資料 base。
- [ ] `tools/extract_cmf.py`:定位後 dump 音樂資料 → `work/music/`(gitignore)+ 解析驗證。

### Phase 2 — OPL2 core + CMF parser(離線驗聽感)
- [ ] `dq3_remake/src/dq3_opl2.{c,h}`:OPL2(YM3812)FM 核心(公開精簡實作移植;render s16 取樣)。
- [ ] `dq3_remake/src/dq3_cmf.{c,h}`:CMF parser(header/instruments/events)+ tick 排程器(→ OPL reg 寫入)。
- [ ] 離線工具 render 一首 → wav,人工聽感確認(像 AdLib/聲霸卡 FM)。

### Phase 3 — SDL audio 串接 remake
- [ ] `dq3_remake/src/dq3_audio.{c,h}`:SDL audio callback（44100 s16）+ OPL2 render + 排程推進。
- [ ] API:`dq3_audio_play(song)` / `dq3_audio_stop()` / `dq3_audio_fade()`；dummy 驅動下不出聲不崩(headless 安全)。
- [ ] CMake 接 SDL audio;game_tester 在 dummy audio 下零回歸。

### Phase 4 — 場景配曲 + 設定
- [ ] 場景配曲:地表 / 城鎮 / 戰鬥 / boss / 結局…(先合理對應，後可 RE EXE 選曲邏輯精校)。
- [ ] 設定選單加「音樂開關 / 音量」；存檔格式不受影響（音訊狀態不入存檔或加版本防呆）。

### Phase 5 —(次要）VOC 音效
- [ ] FVOC/NVOC.VCX(Creative VOC PCM)解碼 → SDL 混音;戰鬥/選單音效接點。

## 驗收
- 步驟 1 產出可解析的 CMF 資料（或據實判定無附曲序）。
- 離線 render 的 wav 聽感正確（FM 暖音色、非雜訊）。
- remake 在地表/戰鬥/城鎮聽到精訊自己的曲；headless / game_tester 零回歸；設定可關音樂。
