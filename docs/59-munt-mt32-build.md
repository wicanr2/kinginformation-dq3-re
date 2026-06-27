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

## 使用

```bash
# 把一個 MIDI 經 MT-32 ROM render 成 wav(-m 指向放 ROM 的目錄,-s 輸出檔)
docker run --rm -v "$PWD/work/music":/m munt-smf2wav -m /m/mt32rom -s /m/out.wav /m/song.mid
```

DQ3 全 18 軌的自動轉換(MIDI → MT-32 wav → OGG)整合在
[`../tools/export_music_mt32.sh`](../tools/export_music_mt32.sh)。

## 版權立場

MT-32 ROM 與遊戲音樂資料皆來自使用者合法持有的硬體與遊戲檔,僅作個人保存與技術研究,不隨原始碼散布。
MT-32 版屬「再想像」(精訊當年未做此音源),非歷史原貌,於 [`58`](58-taiwan-jingxun-music-preservation.md) 已標明。
