# 用 munt(MT-32 模擬器)把 DQ3 音樂 render 成 Roland MT-32 版

> 精訊版 DQ3 當年只做了聲霸卡 OPL2 FM(見 [`58`](58-taiwan-jingxun-music-preservation.md)、[`57`](57-cmf-audio-re.md)),
> 沒有 MT-32 / Roland 音源。但音樂事件流本身是 MIDI-like,轉成標準 MIDI(`tools/cmf_to_midi.py`)後,
> 可用真正的 MT-32 ROM 透過 **munt** 離線 render 成「如果當年有 Roland 會是什麼樣子」的版本。
> 本檔記錄 munt 的 Dockerfile 與編譯/使用經驗,給後人省去踩坑。

## 為什麼是 munt,不是 fluidsynth / GM SoundFont

| 方案 | 問題 |
|---|---|
| fluidsynth + GM SoundFont | ① GM SoundFont **不是 MT-32 音色**(只是通用 MIDI),音色不對;② `mt32emu-smf2wav` 以外的離線 render 在某些環境會 real-time 串流不停(實測 fluidsynth `-F` / `-a file` 會把 78 秒的曲子寫成 **1.27 GB** 不終止)。 |
| **munt `mt32emu-smf2wav`** | ✅ 用**真正的 MT-32 ROM**(CONTROL + PCM)發聲 = 道地 MT-32 音色;✅ SMF→WAV 是**離線批次轉檔**,跑完就結束,不會串流暴衝。 |

munt 是 DOSBox / ScummVM 內建的 MT-32 模擬器,最權威的開源實作(<https://github.com/munt/munt>)。

## 踩坑經驗

1. **Debian/Ubuntu apt 沒有 munt 套件**(`apt-get install munt` / `mt32emu-smf2wav` 都 `Unable to locate package`)→ 只能從源碼編。
2. **要分兩步 build**:先編 `mt32emu`(核心庫,`-Dlibmt32emu_SHARED=TRUE` 出 .so)+ `cmake --install` + `ldconfig`,
   再編 `mt32emu_smf2wav`(CLI 工具,連結剛裝的庫)。順序反了會找不到 libmt32emu。
3. **`mt32emu_smf2wav` 額外依賴 pkg-config + glib2**:只裝 `git cmake build-essential` 會在 smf2wav 的 cmake config 卡
   `Could NOT find PkgConfig` / `Could NOT find GLIB2 (>= 2.32)`。要補 `pkg-config libglib2.0-dev`
   (核心庫 `mt32emu` 不需要,但 CLI 工具需要)。
4. **ROM 命名**:把 CONTROL ROM 命名為 `MT32_CONTROL.ROM`(64KB)、PCM ROM 命名為 `MT32_PCM.ROM`(512KB)放同一目錄,
   munt 以內容/大小自動偵測。經典 MT-32 用 `MT32_CONTROL.1987-10-07.v1.07.ROM`;CM-32L(超集,音色更多)用 CM32L 那組。
5. **ROM 不入庫**:ROM 屬使用者自有硬體韌體,放 `work/`(gitignore),不隨原始碼散布。

## Dockerfile

完整檔見 [`../tools/Dockerfile.munt`](../tools/Dockerfile.munt)。重點:

```dockerfile
FROM debian:bookworm-slim
RUN apt-get update -qq && apt-get install -y --no-install-recommends \
        git cmake build-essential ca-certificates pkg-config libglib2.0-dev
RUN git clone --depth 1 https://github.com/munt/munt.git /munt && \
    cmake -S /munt/mt32emu -B /b-lib -DCMAKE_BUILD_TYPE=Release -Dlibmt32emu_SHARED=TRUE && \
    cmake --build /b-lib -j4 && cmake --install /b-lib && ldconfig && \
    cmake -S /munt/mt32emu_smf2wav -B /b-smf -DCMAKE_BUILD_TYPE=Release && \
    cmake --build /b-smf -j4 && cmake --install /b-smf
ENTRYPOINT ["mt32emu-smf2wav"]
```

build:`docker build -f tools/Dockerfile.munt -t munt-smf2wav tools/`

## ROM 需求(retro CHT MT-32 技巧)

munt 要靠**真正的 Roland 韌體 ROM** 發聲。一套完整 ROM = **CONTROL ROM + PCM ROM**:

| 機種 | CONTROL ROM | PCM ROM | 備註 |
|---|---|---|---|
| **MT-32(經典)** | 64 KB(`MT32_CONTROL.ROM`) | 512 KB(`MT32_PCM.ROM`) | v1.07(1987)最常見;另有 v1.04 / patched |
| **CM-32L / LAPC-I** | 64 KB(`CM32L_CONTROL.ROM`) | **1 MB**(`CM32L_PCM.ROM`) | MT-32 超集,多 33 個音色;1989 v1.02 |

- **檔名**:munt 以**內容/大小自動偵測**(不靠檔名),放同一目錄即可;MAME-versioned 命名如
  `mt32_ctrl_1_07.rom`、`cm32l_pcm.rom` 也可。`-i mt32` / `-i cm32l` 指定機種。
- **版本差異影響音色**:v1.07 vs CM-32L 的低音/打擊不同;遊戲若針對特定機種混音,選對版本較還原。
- ⚠ **版權與取得**:MT-32/CM-32L ROM 是 **Roland 版權韌體**,非自由散布。
  正規取得 = **從自己擁有的實機 dump**(retro 保存圈合法途徑);**本專案不內含、不散布 ROM**,
  使用者自備放 `work/music/mt32rom/`(gitignore)。下載連結不在此記錄(避免協助散布版權檔)。

## 使用

```bash
# 把一個 MIDI 經 MT-32 ROM render 成 wav(-m 指向放 ROM 的目錄,-s 輸出檔)
docker run --rm -v "$PWD/work/music":/m munt-smf2wav -m /m/mt32rom -s /m/out.wav /m/song.mid
```

DQ3 全 24 軌(MBG 18 + EBG 6 戰鬥音樂,docs/61)的自動轉換(MIDI → MT-32 wav → OGG)整合在
[`../tools/export_music_mt32.sh`](../tools/export_music_mt32.sh)(EBG 段會自動抽取+轉 MIDI 18-23)。

## 遊戲內音源切換(remake 設定選單)

remake 設定選單加了「**音源:SB FM / MT-32**」一列(僅當 MT-32 音檔就緒才出現):
- **SB FM** = 現有 OPL2 即時合成(`dq3_opl2`),無額外資產。
- **MT-32** = 播放預先 render 的音檔(從你的 MT-32 ROM 出來的),由 SDL audio callback 串流 + 重取樣。

**音檔格式:優先 OGG,退回 WAV。**
- build 時 CMake 偵測 `libvorbisfile`(裝 `libvorbis-dev`)→ 定義 `DQ3_HAVE_VORBIS`,remake 直接解碼播放 `.ogg`(~16MB / 24 軌,小)。
- 沒有 libvorbis 的平台 → 用 `.wav`(`tools/export_mt32_wav.sh` 產 22050Hz mono,~37MB,零解碼依賴)。
- 放置:把音檔放到 `<assets>/mt32/track_NN.ogg`(或 `.wav`),或設 env `DQ3_MT32_DIR`。設定值存 `dq3.cfg` 的 `audio=sb|mt32`。

> 匯出 24 軌 OGG(MBG 18 + EBG 6 戰鬥):`tools/export_music_mt32.sh`(已含,輸出 `work/music/export/mt32/`,複製到 `<assets>/mt32/` 即用)。

## 版權立場

MT-32 ROM 與遊戲音樂資料皆來自使用者合法持有的硬體與遊戲檔,僅作個人保存與技術研究,不隨原始碼散布。
MT-32 版屬「再想像」(精訊當年未做此音源),非歷史原貌,於 [`58`](58-taiwan-jingxun-music-preservation.md) 已標明。
