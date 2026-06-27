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

## Trace 進度(2026-06-27,step 1b-1c)

- **timed-playback 函式**(file 0x13f1c):`getvect(8)` 存舊 → `cli; in 0x21; mask=0xfe`(只留 IRQ0 timer)→
  `set_timer_count(0x1b58)`(≈170Hz)→ `setvect(8, cs:0x29f)`(裝 ISR)→ **busy-wait**(`or ax,ax; je/jne` 等 ISR 改 ax)→
  還原 mask + timer。⇒ 這是**同步阻塞播放**(短 jingle/音效跑完才返回);連續 BGM 應另有非阻塞安裝點。
- 呼叫 `set_timer_count`(0x2c52)者:logical 0x12bcb / 0x12bf5(即此函式內)。
- ★**多段陷阱**:ISR = `cs:0x29f`,此 `cs` 是**音樂驅動自己的段**(CMFDRV 從 SBCM.LIB 連入,獨立 segment),
  **不是主碼段**。直接反組譯主段 file 0x160f(=主段邏輯 0x29f)得到的是遊戲碼(give-item/memory alloc),錯段。
  → 須先解出**驅動段的 file base**(查 EXE relocation / segment 配置 / `mov ax,seg; mov ds/es,ax` 載入驅動段的常數),才能正確反組譯 ISR。

## ★ 解鎖:load-relative 段位址 → file offset 換算(已驗證,2026-06-27)

多段 large-model EXE 反組譯的關鍵換算 —— **`file = seg×16 + off + 0x1370`**(0x1370 = EXE header 結束 / 載入映像在檔內起點):
- 驗證:INT8 ISR 裝成 `cs:0x29f`,install 碼在 file 0x13f1c;反推 install 段 = load-relative **seg 0x1000**
  (0x1000×16+0x29f+0x1370 = file 0x1160f,disasm 落在連貫 ISR、`sti;iret` 結尾 ✓)。
- 再驗:`lcall 0x1104:0xac` → file 0x1245c、`lcall 0x1032:0xb2` → file 0x11742,兩者都 disasm 出連貫程式碼 ✓。
- 用途:`9a lo hi seglo seghi` = far call seg:off;拿 seg/off 套公式即得 file,**跨段 trace 不再卡**。
- ⚠ 64K wrap:`re_codecave_scan.py` 的 `sXXXX` label = (file−0x1370)&0xffff(段內 offset,wrap at 64K);
  同一 label 可能對應多個 window(±0x10000),要用所在函式的 window 對齊(別像先前反到差 0x10000 的錯段)。

## ★ 修正:追到的 timer/INT8 ISR 是**螢幕/文字**子系統,非音樂(2026-06-27)

跟進 ISR(file 0x1160f,seg 0x1000:0x29f)+ 其 far call 後確認:
- ISR `mov es,0xa000`(VGA 記憶體);far target B(file 0x11742)`mov al,[si]; cmp al,0x24('$') 終止; … mul 84(stride); shr bx,1 ×3; add bx,[0x4f09]` = **VGA 位址計算 + '$' 結尾字串描繪**。
- ⇒ 這 timer 重排 + INT8 hook 是**定時把字串畫到 VGA 的螢幕子系統**,不是 FM 音樂驅動。
- **修正前一節**:「8253 timer reprogram = 音樂節拍」的佐證**追錯 ISR**(此 timer 服務螢幕,非音樂)。
  音樂**仍存在**(CMFDRV 已連入 + 使用者親證 + YouTube 錄影),但要從**另一條**重追:
  CMFDRV 的 OPL 寫入(0x228/0x229 via `_ct_io_addx`)+ SB 自己的中斷,**不是這個 0xa000 螢幕 timer**。

## 已排除的路徑(別重試,2026-06-27)

- ❌ **掃 `out dx,al`(0xEE)分群**:全檔 392 處、172 群,散在 VGA/CRTC/DSP 各處,雜訊太高不收斂。OPL 寫入埋在其中但無法靠數量挑出。
- ❌ **遊戲 `DQ3CON.MAP` / `DQ3UND.MAP` 當 linker map**:這兩個**不是 link map**,是二進位遊戲資料(N/X tile 填充樣式),副檔名誤導。真 link map 只有 `SBCM.MAP`(驅動自身,符號為 unresolved externals,無遊戲端絕對位址)。
- ⇒ **沒有遊戲端 .MAP 可給符號絕對位址**;得靠反組譯定位驅動碼。

## 下一步(最有價值)— 用 SBCM.LIB 在 EXE 裡定位 CMFDRV 驅動

- `SBCM.LIB`(OMF 物件庫)含 `_SBFM_PLAY_MUSIC` / `_SBFM_INSTRUMENT` / `CMFB` / `_SBC_SCAN_CARD` 等驅動碼。
  linker 把它整段連進 DQ3.EXE → **驅動的機器碼在 LIB 與 EXE 內 byte-identical**。
- 作法:從 LIB 取 `_SBFM_PLAY_MUSIC`(或 `_SBC_SCAN_CARD` 的 SB 偵測 `out reg4,0x60/0x80; in status`)的一段獨特 byte signature → 在 DQ3.EXE 搜 → 命中 = 驅動在 EXE 的位置。
- 定位驅動後:① 反組譯 `_SBFM_PLAY_MUSIC` 看它怎麼收音樂資料指標(CMF 解析);② 用段→file 公式找**誰呼叫它**(遊戲選曲處)+ 傳入的指標 = 內嵌音樂資料 base;③ 找指標表 = 曲目表。

## (參考)CMFDRV/OPL 線通用指引(取代追錯的 timer 線)
- 從 `_ct_io_addx`(SB IO base 變數)被寫的位置 → 找寫 OPL data 埠(base+1,值 = reg/value)的迴圈 = CMF 播放器。
- 找 CMFDRV 的 `ct_play_music` / `_ct_music_status` 對應 code:誰把音樂資料指標餵給它 → 指標 base = 內嵌音樂資料。
- 用上面的段→file 公式跟 CMFDRV 段的 far call(SBCM.LIB 連入的獨立段)。

## (作廢)舊下一步 — 續 step 1c→(基於追錯的 timer 線,保留供對照)
- 解驅動段 file base:找載入驅動段選擇子的 `mov ax,<seg>`(如前段見 `mov ax,0x14dd; mov ds,ax`、`mov ax,0x1ade`)→
  對映 EXE segment 表 → 算驅動段 file 起點 → 反組譯 ISR(driver_seg:0x29f)。
- 找**非阻塞 BGM 安裝**(連續播放):另一處 `setvect(8,…)` 不接 busy-wait、直接 return。
- 自 ISR 找 OPL 寫入(reg 0xA0-0xB8 note)+ 事件指標變數 → 反追音樂資料 base + 曲目表。

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
