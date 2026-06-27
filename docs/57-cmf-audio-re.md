# 精訊 DQ3 音樂(SB FM / CMF)反組譯經驗

> 計畫見 [`56-audio-fm-plan.md`](56-audio-fm-plan.md);工作清單 [`../dq3_remake/AUDIO_WORKLIST.md`](../dq3_remake/AUDIO_WORKLIST.md)。
> 本檔記錄 **Phase 1(定位 + 抽出 CMF 音樂資料)** 的 RE 過程、findings 與判定 —— 成功失敗都留。
> **IP**:全程只反組譯遊戲**自身**二進位格式(使用者正版檔),非重現受版權樂曲。

## 目標

讓 remake 播精訊**自己編的** SB FM 音樂:從原版遊戲資料 RE 出音樂序列 → OPL2 合成器播放。
第一步是「資料到底在哪、長怎樣」——本檔即此步的偵察紀錄。

## 偵察鏈(2026-06-27)

1. **素材夾掃音訊檔** → 命中 `SBCM.LIB`、`SBCM.MAP`、`FVOC.VCX`、`NVOC.VCX`。
2. **`SBCM.MAP` = linker map**(`Start/Stop/Length/Name/Class`,CODE/STACK/DATA/`MUSIC_TEXT` 段)→ `SBCM.LIB` 是**驅動 object/lib**,非曲子。
3. **`SBCM.LIB` 符號** → `CMFDRV` / `CMFSEG` / `cmfasm.ASM` / `_ct_card_here` / `_ct_io_addx` / `_ct_music_status` /
   `_ct_scan_card` / `_ctvd_*`(CT-VOICE 數位音)/ `_ctvm_*`。⇒ 這是標準 **Creative CT-VOICE + CT-MUSIC(CMFDRV)驅動套件**。
4. **EXE 字串** → `Creative Sound Blaster Card`、`fvoc.vcx`、`nvoc.vcx`、`player.dat`。⇒ 用 Creative SB;音效走 VOC。
5. **`FVOC.VCX`/`NVOC.VCX`** → 開頭是 offset 索引表(`0x5c,0x4e6,0x976,…`)的 **VOC PCM 音效 bank**(數位音效,非音樂)。
6. **`PLAYER.DAT`(200B)** → 玩家初始狀態表(`0x0200,0x5839,…` 重複 `0101 0100 8b2e`),**非音樂**。
7. **OPL 埠** → EXE 內 word `0x228`(SB FM reg 埠)出現 7 處,但:
   - `mov dx,0x228`(`ba 28 02`)**0 命中** → 埠是 runtime 經 `_ct_io_addx` 變數間接寫(SB 偵測後填)。
   - 多數 `0x228` 落在**遞增資料表**(`0x0223,0x0224,…,0x0228,0x0229,…`,被 disasm 誤判成 `add` 指令)→ 非埠引用。
8. **全檔(含 EXE)搜 CMF magic `CTMF`** → **0 命中**;EXE 無 `music`/`CMF`/`song`/`BGM` 字串。

## Phase 1 判定(誠實)

**在 shipped 檔案裡找不到標準 CMF(`CTMF` magic）音樂資料。** 兩種可能,尚未分清:

- **(A) 此未完成 build 沒做完音樂**:Creative CMF 驅動已 link、VOC 音效已備,但 **CMF 曲序從未加入** ——
  與缺 128/129 怪 sprite 同類的「半成品」缺口。若如此,**無精訊音樂可抽**(原始曲為椙山版權,不抄)。
- **(B) 音樂內嵌 EXE 非標準形式**:CMF body 去掉 `CTMF` 檔頭後內嵌、或精訊自改格式 → 需**反組譯 CMFDRV 的
  `ct_play_music` 呼叫端**,看遊戲是否曾餵音樂資料指標給驅動、指向哪。

> rule 62:搜尋落空 ≠ 絕不存在。判定 (A) 前須跑完 (B) 的 driver-trace。

## ★ 更正:音樂**確實存在且運作中**(使用者親證 + YouTube 精訊版錄影 + RE 獨立佐證,2026-06-27)

使用者親耳聽過、且 YouTube 有精訊版遊戲錄影有 BGM → **排除 (A)「沒做完」,確認 (B):音樂真的在、內嵌 EXE**。
RE 也獨立佐證音樂播放器**實際運作**:

- EXE **重設 8253 timer**:`mov al,0x36; out 0x43`(2 處:file 0x13fc3 / 0x155c0)→ 設定音樂節拍中斷率。
- **`set_timer_count(ax)` 函式**(logical **0x2c52**):`push ax; mov al,0x36; out 0x43; pop ax; out 0x40; mov al,ah; out 0x40; ret`
  —— 把節拍值寫進 8253 channel 0(改 timer 頻率 = 音樂 tempo,典型 FM 音樂驅動作法)。
- 其後一個 ISR(`not ax; mov al,0x20; out 0x20; iret` = 發 EOI 給 PIC)。

⇒ EXE 內**有完整的計時音樂播放機制**,音樂資料內嵌(非 standalone CMF 檔、無 CTMF magic、非 overlay)。
**(A) 已排除。** 缺的只是「資料 base 在哪」——需從 timer ISR / 音樂事件處理器反追資料指標(deep trace)。

## 下一步(Phase 1 續)— 定位內嵌音樂資料

- 反組譯 EXE 內 **CT-MUSIC 初始化**(`ct_scan_card`/OPL reset)與 **play 呼叫**:
  - 若遊戲**從未呼叫** play(只 init 驅動)→ 確認 (A):音樂未完成,據實回報、Phase 2+ 暫緩或改原創 BGM。
  - 若有 play(data 指標)→ 跟指標到音樂資料 base,抽出驗解析 → 進 Phase 2(OPL2 core + parser)。
- 交叉佐證:用 DOSBox 跑原版(若環境可)聽是否真有 BGM 還是只有音效——但本環境無 DOSBox,以靜態為主。

## 參考:CMF / OPL2(公開規格,供 Phase 2)

- **CMF header**:`CTMF`(4) + ver(2) + instrument_offset(2) + music_offset(2) + ticks/quarter(2) + tempo(2) +
  title/author/remarks offsets + channel-in-use(16) + instrument_count(2) + basic_tempo(2)。
- **Instrument block**:每樂器 16 byte = 2 operator 的 OPL2 參數(mod/car:char、scale/level、attack/decay、
  sustain/release、wave)+ feedback/connection。
- **Music block**:MIDI-like delta-time 事件流(note on/off、program change 對應 instrument)。
- **OPL2(YM3812)**:2-op FM × 9ch;reg index→addr 埠、值→data 埠;模擬有公開精簡 core 可移植。

## 教訓

- **音訊驅動 vs 音樂資料分清**:`*.LIB`/linker map = 驅動 object(程式),不是曲子;曲子是另一份資料(CMF)。
  別把驅動 lib 當音樂在那找音符。
- **間接埠**:Creative 驅動的 IO 埠 runtime 偵測填變數(`_ct_io_addx`)→ 搜 `mov dx,埠` 會落空,屬正常,非「沒在用」。
- **資料誤判成碼**:遞增 word 表(`0x0223..`)會被反組譯器讀成 `add` 指令,內含的 `0x228` 是巧合非埠 —— 反組譯找埠引用時要先分辨碼/資料段。
- 與「缺 sprite」同源:未發售 build 的**缺口要先確認是「沒做」還是「沒接好/沒找到」**(同 `re-content-exists-not-build-incomplete` 的反面:這次可能真的是原版沒做完)。
