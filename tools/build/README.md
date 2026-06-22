# RE C-code 編譯工具鏈(docker)

把反組譯還原的 C(`re/*.c`)編回 16-bit DOS 的 build 環境。原版經指紋分析確認由
**Microsoft C 5.x(near-code model)** 編譯(證據見 [`docs/17`](../../docs/17-build-toolchain.md)、
[`docs/19`](../../docs/19-re-correctness.md)),故主工具鏈用 **MSC 5.1**,以「逐函式 byte-match」
為正確性判準。

> **版權**:古董編譯器二進位(MSC 5.1 / Turbo C floppy image、抽出的 BIN/LIB/INCLUDE、.obj)
> **不入 git**(見 `.gitignore`);本目錄只記錄 **Dockerfile + build 腳本 + 取得步驟**,可重現。

## 檔案

| 檔 | 用途 |
|---|---|
| `Dockerfile.msc` | Microsoft C 5.1(1988)在 DOSBox 內;提供 `CL/C1/C2/C3/LINK`。**主工具鏈** |
| `msc_compile.sh` | 在 docker 內把單一 `.c` 編成 `.obj`(`/c /AS /Ox`,model 可選 `/AS/AM/AC/AL/AH`)|
| `Dockerfile.turboc` | Borland Turbo C 2.01;比對用(已排除為原版編譯器)|
| `tcc_compile.sh` | TC 版編譯腳本 |
| `Dockerfile.sdl` / `sdl_build.sh` | 現代 SDL2 移植版的 build(見 `re/sdl/`,終極目標)|
| `poc*.c` | byte-match PoC 原始碼(亂數函式 `sub_e6b9` 等)|

## 取得 MSC 5.1 並建 image(可重現步驟)

```bash
# 1) 從 archive.org 取 MSC 5.1 floppy .img,解到 tools/build/msc51_img/(不入 git)
# 2) 用 mtools 把元件抽到 tools/build/msc/(BIN / LIB / INCLUDE)
tools/build/msc_extract.sh          # (抽取腳本)
# 3) build image
docker build -f tools/build/Dockerfile.msc -t dq3-msc tools/build
```

## 編譯單一函式(byte-match)

```bash
tools/build/msc_compile.sh re/somefile.c /c /AS /Ox   # → re/somefile.as.msc.obj
# 再用 tools/re_match.py 反組譯 .obj 並與原版 DQ3.EXE 該函式逐 byte 比對
```

## 現況與目標

- **byte-match PoC**:亂數函式以 MSC 5.1 `/AS /Ox` 編出與原版逐 byte 相符(指令序列、NOP padding、
  CONST 段、near ret 全對),證明「逐函式 byte-match 反編譯」可行(詳見 docs/17)。
- **層級**(見 docs/19):①binary patch ②ASM 重組(已 byte-identical 100%)③**C 重編**(本工具鏈,進行中)。
- **TODO(正路 b)**:完成 `re/*.c` 全函式反編譯 → 用本工具鏈逐函式 byte-match → link 成 `RE-DQ3.EXE`;
  bug 修進 C source(非 binary patch),再以此為基礎做 `re/sdl/` 的 SDL2 現代移植。
- 目前的「修正版」(`tools/build_fixed_version.py`)是**層級 ①binary patch**,當 C 重編版的**對照組**。
