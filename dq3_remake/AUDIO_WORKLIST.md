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
  音樂格式 = CMF 變體(OPL2 FM)。
  > ⚠ 偵察期兩個推測**事後證實是錯的**(已修正):①「EXE 載入清單無音樂檔」→ 音樂在 `MBG.MCX`(只是沒以明文檔名靜態字串出現,動態組名/封包 seek);②「全檔無 CTMF magic ⇒ 內嵌 EXE」→ 是精訊自訂 CMF 變體(本來就沒 CTMF magic),且在**外部 MBG.MCX**。
- [x] **(B) 確認:音樂真的在**(使用者親證 + YouTube 精訊版錄影 + RE 定位)。
  > ⚠ 先前拿來佐證的「EXE 重設 8253 timer + 計時 ISR = 音樂節拍」**追錯 ISR**(那是 VGA 螢幕/文字子系統,與音樂無關)。音樂為真的正解見下方 Phase 1 完成(MBG.MCX）。
- [~] **Phase 1 續(進行中,deep trace)**:從 timer ISR / 音樂事件處理器反追到內嵌音樂資料 base。
  目標:定位 ① 音樂事件流(序列)base、② instrument(OPL2 樂器)表 base、③ 曲目索引表(多首)。
  具體步驟(逐步做,別跳;每步 commit 進度到 docs/57):
  - [x] 1a. timer init 位置:file 0x13fc3 / 0x155c0;`set_timer_count(ax)` = logical 0x2c52。
  - [x] 1b/1c. 追了 INT8/timer hook → **此線追錯**(該 ISR 是 VGA 螢幕/文字子系統)。
    副產:★解出 **段→file 換算 `file=seg×16+off+0x1370`**(已驗證)→ 跨段 trace 解鎖。
  - [x] **1 完成(2026-06-27)**:`tools/omf_lib.py` 解 SBCM.LIB → byte-match 把 CMFDRV/cmfasm/midirtn 定位進 EXE
    → 找 `_sbfm_play_music`(file 0x140ee)2 個呼叫者 → 解出音樂**從 `MBG.MCX` seek+read 載入**(非內嵌 EXE)。
    **`MBG.MCX` = 76-byte dword 偏移表 + 18 條 CMF 變體音軌(OPL2 FM)。** 全程見 docs/57。
- [x] `tools/extract_cmf.py`:拆 MBG.MCX → `work/music/track_00..17.bin`(gitignore);印各軌 offset/len/header。18 軌 OK。
  - ~~1d–1h(原計畫:從 ISR 反追內嵌 EXE 的事件/樂器/曲目表)~~ **作廢**:基於「內嵌 EXE」錯誤前提;
    Phase 1 已由 OMF byte-match + MBG.MCX 偏移表直接解決(曲目表 = MBG.MCX 前導 dword 表,18 軌)。
- [ ] `tools/extract_cmf.py`:依 1f/1h 定位 dump 各曲 → `work/music/song_NN.bin`(gitignore)+ 印 instrument 數/事件數驗證。

### Phase 2 — OPL2 core + CMF parser(離線驗聽感)
- [x] **事件格式解碼 + 離線驗證(2026-06-27)**:`tools/cmf_render.py` 解析軌事件(MIDI-like + 尾隨 delta +
  **running status**)+ 2-op FM 合成 → wav。track00=2021 事件/1007 音符/~120s、track02=997/494/~49s,**旋律完整正確**。
  格式規格見 docs/57「Phase 2 起步」。⚠ track17 等少數軌 note-on/off 配對失敗(編碼變體,待查)。
- [x] `dq3_opl2.{c,h}`:精簡 OPL2(YM3812)FM 核心(9×2-op、4 波形、ADSR、fnum/block;暫存器層 + s16 render)。近似但 AdLib 音色明確。
- [x] `dq3_cmf.{c,h}`:軌 parser(MIDI-like + running status)+ tick 排程器(note→fnum/block、program、預設 FM patch)。
- [x] `tools/opl_render.c`:離線 track→wav 驗證 OPL2 音色(track00 peak18073/rms6226/97%非靜音)。
  ⚠ 待精校:各軌 instrument→OPL patch(目前預設 patch)、精確 tempo、track17 等變體軌。

### Phase 3 — SDL audio 串接 remake ✅
- [x] `dq3_audio.{c,h}`:載 MBG.MCX(18軌)+ SDL audio callback 驅動 OPL2/CMF + 執行緒安全 + headless no-op。
- [x] API:`dq3_audio_play/stop/set_enabled/set_volume/play_scene`;dummy 驅動下不出聲不崩。
- [x] CMake 接 opl2/cmf/audio + libm;runtime 加 SDL_INIT_AUDIO;17/17 測試零回歸。

### Phase 4 — 場景配曲 + 設定 ✅
- [x] 場景配曲:主迴圈每幀依 in_town/layer/ship 選軌(地表/城鎮/下層/船);戰鬥入口 boss vs 一般曲。
- [x] 設定:config 加 `music`/`music_vol`(load/save/default);主程式依設定 enable/volume。
  ⚠ 待補:遊戲內設定選單 UI 的音樂開關/音量項(目前走設定檔)。

### Phase 5 — VOC 音效 ✅
- [x] `dq3_voc.{c,h}`:FVOC/NVOC.VCX 解碼(dword 偏移表 + 標準 VOC sound block,8-bit unsigned PCM)→ s16。FVOC 22 + NVOC 24。
- [x] `dq3_audio`:8 voice SFX mixer + `dq3_audio_sfx/se` + 語意音效對應;戰鬥指令接攻擊/咒文/確認音效。
  ⚠ 待精校:語意音效→實際 VCX id 對應(初步猜測)、更多接點(選單/開門/道具/受傷)。

## 驗收
- [x] 音樂資料可解析(MBG.MCX 18 軌,旋律完整)。
- [x] 離線 render wav(FM 音色、非雜訊)。
- [x] remake 地表/戰鬥/城鎮聽到精訊自己的曲;headless / 17 測試零回歸;設定檔可關音樂。
- [ ] (精校)各軌實際 instrument 音色、精確 tempo、語意音效對應、遊戲內設定 UI。
