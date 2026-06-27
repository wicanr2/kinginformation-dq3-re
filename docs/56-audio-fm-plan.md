# 精訊 DQ3 音樂(SB FM / OPL2 CMF)還原計畫

> 目標:讓 `dq3_remake` 播出**精訊版自己編的遊戲音樂**——不是憑記憶抄椙山浩一的譜(那有版權、不做),
> 而是**從使用者合法持有的原版遊戲資料反組譯 + 抽出 CMF 音樂序列 + 用 OPL2 FM 合成器播放**
> (同本專案抽地圖/sprite/劇本的做法:資料 gitignore 不散布,執行又有原版 `DQ3.EXE` 閘)。
> 音效(VOC 數位音)為次要,先聚焦音樂。

## 1. 已確認的事實(RE 偵察,2026-06-27)

| 項 | 證據 |
|---|---|
| 精訊有自編 SB FM 音樂 | `DQ3.EXE` 寫 **0x228 / 0x229**(Sound Blaster 的 OPL FM 暫存器埠;reg 7 處 / data 5 處)→ 跑起來用 OPL2 發聲 |
| 音樂驅動 = CMF | `SBCM.LIB`(linker map `SBCM.MAP`)= **`SBFM_CMF_DRV`**(Creative SB FM CMF 驅動);符號 `_ct_music_status`、`_ct_io_addx`、`_sbfm_status_addx`、`_sbfm_fade…` |
| 音效 = VOC | `FVOC.VCX` / `NVOC.VCX`(Creative VOC 數位音;EXE 字串 `fvoc.vcx`/`nvoc.vcx`、`Creative Sound Blaster Card`)|
| 音樂資料位置 | ★**未定**:素材夾無 CTMF 檔頭(標準 CMF magic);`PLAYER.DAT`(200B)為玩家初始狀態非音樂 → CMF 曲序疑**內嵌 DQ3.EXE**(linker `MUSIC_TEXT` 段連入),待 RE 定位 |

## 2. 背景:CMF 與 OPL2(公開規格)

- **OPL2(YM3812)**:AdLib / Sound Blaster 的 FM 音源晶片,2-operator FM × 9 channels(或 6+5 打擊)。
  暫存器寫法:先寫 reg index 到 addr port(SB FM = 0x228)、再寫值到 data port(0x229)。模擬有公開實作。
- **CMF(Creative Music Format)**:檔頭 `CTMF` + 版本 + instrument block offset + music block offset +
  ticks/tempo + 樂器數;instrument = 16-byte OPL2 operator 參數;music block = MIDI-like delta-time 事件流。
  規格公開,parser 不難。

## 3. 架構(remake 目前**無音訊引擎**,全新子系統)

```
PLAYER/EXE 內 CMF 資料 ──parse──▶ {instruments[], events[]}
                                        │
SDL audio callback(44100, s16)         ▼
   每 callback 取 N 樣本 ◀── OPL2 emulator core(render N 樣本)
                                        ▲
   音樂排程器(依事件 delta-time 在正確 tick 寫 OPL reg / note on-off)
                                        │
   場景配曲(main loop):地表曲 / 戰鬥曲 / 城鎮曲 / boss / 結局…  ◀── dq3_audio API
```

元件:
- `dq3_opl2.{c,h}` —— OPL2 FM 核心(公開精簡實作移植,render 取樣)。
- `dq3_cmf.{c,h}` —— CMF parser(instruments + event stream)+ 排程器(tick → OPL reg 寫入)。
- `dq3_audio.{c,h}` —— SDL audio callback 串接 + 對外 API(`dq3_audio_play(song_id)` / `stop` / `fade`)。
- `tools/extract_cmf.py` —— 從 DQ3.EXE 定位 + 抽出 CMF 音樂資料(產 gitignore 的 `work/music/*.cmf` 供驗證)。

## 4. IP / 版權立場(釘住)

- **做**:反組譯遊戲**自身**的音樂資料、用 OPL2 合成器播放(資料來自使用者正版遊戲檔,gitignore、不入庫、
  不散布;remake 又有原版 `DQ3.EXE` 指紋閘)。同本專案 RE 地圖/sprite/劇本。
- **不做**:憑記憶/知識抄寫椙山浩一的旋律進合成器(那是複製受版權保護的樂曲)。本計畫**不**含任何手抄譜。

## 5. 風險 / 未決

- 音樂資料 offset 未定(可能內嵌 EXE);若**精訊未完成 build 根本沒附曲序**(如同缺 128/129 sprite),
  則「無精訊音樂可抽」—— 步驟 1 會先確認這點。
- OPL2 cycle-accurate 模擬非必要;聽感正確即可(用成熟精簡 core)。
- 配曲對應(哪首曲配哪場景)需另 RE EXE 的選曲邏輯,或先以合理對應暫接。

## 6. 分步(詳見 `dq3_remake/AUDIO_WORKLIST.md`)

1. **定位 + 抽出** EXE 內 CMF 音樂資料(`tools/extract_cmf.py`)→ 確認資料存在 + 可解析。**先做這步 de-risk。**
2. OPL2 core + CMF parser → 離線 render 一首成 wav 驗聽感。
3. SDL audio callback 串接 remake + `dq3_audio` API。
4. 場景配曲(地表/戰鬥/城鎮/boss/結局)+ 設定選單音量/開關 + 存檔不受影響。
5. (次要)VOC 音效。
