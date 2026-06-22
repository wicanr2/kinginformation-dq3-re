# DQ3.EXE 編譯器辨識、build 工具鏈與 SDL2 移植骨架

本篇處理「怎麼把反組譯產物變回可驗證的可執行檔」這條線,並接上最終目標——
把精訊 DQ3(《DRAGON FIGHTER III 傳說的終章》,© 1993 精訊資訊)反組譯成乾淨 C,
再用 SDL2 改寫成跨平台現代版。

涵蓋三塊:
1. **原版編譯器辨識**(指紋證據 → Microsoft C 5.x)。
2. **古董編譯器 docker 建法 + byte-match 方法論 PoC**(可重現)。
3. **SDL2 移植骨架 + 最小可跑 PoC**(已驗證:解碼標題畫面並顯示)。

---

## 1. 原版編譯器辨識:Microsoft C 5.x(near-code 模型)

先前 docs/05 推測「large model」。逐函式反組譯後修正:**主程式碼段(seg0)是
near-call / near-ret**,只有對 runtime 段才用 far call。證據如下。

### 1.1 codegen 指紋(seg0 主碼,280 函式統計)

| 指紋 | 觀察 | 指向 |
|---|---|---|
| `ret` 型別 | 268 函式 near `ret`(c3),僅 1 個 `retf` | **near code**(small/medium,非 large/huge) |
| BP frame | **0/280** 用 `push bp;mov bp,sp` prologue | 非偵錯版;優化打開 |
| `LOOP` 指令 | **184 次**(`loop`,如 sub_32a3 `mov cx,0xd;…;loop`)| **Microsoft C**(Borland 幾乎不發 LOOP) |
| 旋轉 | `rol`×25、`ror`×21(如 LCG `d1c0 d1c0 d1c0`)| 編譯器辨識旋轉慣用語 |
| `cdq` | 4 次(32-bit 符號延伸)| MSC 風格 |
| NOP padding | **459 個 `nop`**(byte 寫入後 `c606…;90`)| MSC 對齊習慣 |
| 暫存器配置 | 全域以 `mov ax,[abs]`(AX-centric)讀寫 | 直接定址 + DGROUP |

### 1.2 OMF / segment 命名指紋

對候選編譯器產出的 `.obj` 解析 OMF LNAMES + COMENT:

- **Turbo C 2.01**:COMENT 含 `TC86 Borland Turbo C 2.01`;LNAMES =
  `_TEXT/CODE/_DATA/DATA/_BSS/BSS/DGROUP`。**無 CONST 段**。
- **Microsoft C 5.1**:COMENT 含 `MS C` + `SLIBCE`(small lib + 浮點 emulator)+ `0sO`(/Os);
  LNAMES = `DGROUP/_TEXT/CODE/_DATA/DATA/`**`CONST`**`/_BSS/BSS`。**有 CONST 段**。

原版 DQ3.EXE 的 runtime DOS 嚴重錯誤處理(int 24h)字串表明碼存在:
```
Error with drive ?: - $Press X to stop trying, or correct and press another key to try again...$
disk is write-protected$unknown unit$drive not ready$ … $general failure$
```
這套 `$`-結尾訊息 + 「Press X to stop trying」措辭是 C runtime 的 int 24h handler 內嵌資料
(MSC `_hardresume` 系)。EXE 本身無編譯器 banner(link 後 OMF COMENT 被剝除,屬正常)。

### 1.3 結論

綜合「LOOP/rol/ror/cdq/NOP-padding/AX-centric/near-ret + MSC int24h 訊息措辭」,
**原版最可能為 Microsoft C 5.x(small 或 medium model)**,runtime 為手寫組語另段
(VGA/SB/鍵盤/滑鼠/檔案),以 far call 串接。Borland Turbo C 已用 codegen 證據排除
(TC 不發 LOOP、不辨識旋轉、預設帶 BP frame、無 CONST 段)。

> 殘留不確定:exact MSC point-release(5.0 / 5.1)與 flag 組合(`/Ox` vs `/Os /Ot`)
> 影響「旋轉是否展開成 3×`rol`」「是否省 BP frame」。要做到逐函式 100% byte-match
> 需把這兩個 knob 掃出來(見 §2.4 gap)。

---

## 2. 古董編譯器 docker 建法 + byte-match 方法論

全部在 docker 內跑,不污染 host。原版商業編譯器映像來自 archive.org(版權屬各家,
不入 git;見 .gitignore)。

### 2.1 取得編譯器(本機,不入版控)

```bash
# Turbo C 2.01(archive.org TC201)
curl -L -o tools/build/tc201.zip \
  https://archive.org/download/TC201/TC201.ZIP

# Microsoft C 5.1(archive.org,floppy .img 集合)
curl -L -o tools/build/msc51.zip \
  "https://archive.org/download/microsoft-c-5.1-optimizing-compiler-5.25.-7z/Microsoft%20C%205.1%20Optimizing%20Compiler.zip"
unzip -o tools/build/msc51.zip -d tools/build/msc51_img
# 用 mtools 把 CL/C1/C2/C3/LINK + SLIBCR/MLIBCR/LLIBCR + headers 抽出到 tools/build/msc/
#   (Compiler 1/2 disk、Libraries (Small/Medium) disk、Include disk)
```

### 2.2 建 image

```bash
docker build -f tools/build/Dockerfile.turboc -t dq3-turboc tools/build   # Turbo C 2.01 + DOSBox
docker build -f tools/build/Dockerfile.msc    -t dq3-msc    tools/build   # MSC 5.1 + DOSBox
```

兩者都用 DOSBox(headless,`SDL_VIDEODRIVER=dummy`)跑 16-bit 編譯器,把 `.c` 編成 `.obj`。

### 2.3 編譯 + 反組譯比對

```bash
tools/build/tcc_compile.sh tools/build/poc.c  -ms -c       # Turbo C, small model
tools/build/msc_compile.sh tools/build/poc3.c /c /AS /Ox   # MSC, small model, full opt
# 反組譯候選 .obj(OMF LEDATA 抽 code bytes -> capstone 16-bit)
tools/dockrun_cap.sh tools/re_match.py obj tools/build/poc3.as.msc.obj
# 與原版某函式逐 byte 比對
tools/dockrun_cap.sh tools/re_match.py cmp e6b9 16 <candidate_hex>
```

`tools/re_match.py`:極簡 OMF 解析(抽 LEDATA code)+ capstone 反組譯 + byte diff / match%。

### 2.4 byte-match PoC 結果(以 LCG 亂數 sub_e6b9 為標的)

原版 `sub_e6b9`(亂數狀態前進,16 bytes,near ret):
```
a15a0b 051890 d1c0 d1c0 d1c0 a35a0b c3
  mov ax,[g]; add ax,0x9018; rol ax,1; rol ax,1; rol ax,1; mov [g],ax; ret
```

| 候選 | 關鍵差異 | 結論 |
|---|---|---|
| Turbo C 2.01 `-ml` | `retf`(cb)、帶 BP frame、旋轉編成 `shl/shr/or`、無 LOOP | **排除**(far ret、不發 LOOP) |
| Turbo C 2.01 `-ms` | near `ret`(c3)✓,但旋轉仍 `shl/shr/or`、迴圈用 `inc/cmp/jl` 非 LOOP | model 對、codegen 不對 |
| **MSC 5.1 `/AS /Ox`** | near `ret`(c3)✓、`mov ax,[g];add ax,0x9018` ✓(與原版 `a15a0b 051890` 同序)、NOP padding ✓、CONST 段 ✓、`MS C` 浮水印 ✓ | **最相符**;旋轉以 `shl cl=3`(原版 3×`rol`)、含 BP frame(原版無)為殘差 |

**方法論已驗證可行**:docker 內跑古董編譯器 → 產 16-bit `.obj` → 反組譯 → 與原版逐 byte 比對的迴圈
全程可重現。MSC 5.1 已能逐指令逼近(指令序列、暫存器選擇、段命名相符)。

**到逐函式 100% byte-match 的 gap**:
1. MSC point-release(5.0 vs 5.1 vs 6.0)與 flag(`/Ox` vs `/Os /Ot` vs `/G2`)會改變
   旋轉展開(`rol×3` vs `shl cl`)與 BP frame 省略——需掃 knob。
2. 全域變數要 link 到原版同一 DGROUP offset(`[0xb5a]` 等),需自訂 link script / 對齊。
3. runtime 段是手寫組語,不走 C 編譯(見 §3 混合策略)。

> 註:本案最終目標是 **SDL2 跨平台改寫**(§3),不是把整檔編回 byte-identical 的 1993 EXE。
> 上述編譯器辨識的價值在於「**正確理解原版 runtime 行為**」,直接餵養 SDL2 shim 的語意還原。

---

## 3. SDL2 移植骨架(最終目標的地基)

把 real-mode DOS runtime 改寫成 SDL2 / stdio shim(adapters at the edges),
遊戲邏輯(re/*.c 反編譯出的 C)日後改成呼叫這層,而非 far-call wrapper。

### 3.1 結構(re/sdl/)

| 檔 | 角色 |
|---|---|
| `re/sdl/dq3_runtime.h/.c` | DOS runtime → SDL2 shim:framebuffer(4bpp packed + palette)、輸入(SDL keysym → BIOS scancode,對應 docs/08 主迴圈)、計時(SDL_Delay/Ticks)、檔案 I/O(stdio 取代 DOS AH=3Dh/3Fh/3Eh) |
| `re/sdl/dq3_pcx.h/.c` | TIT*.P 標題畫面解碼(ZSoft PCX,640×350,4 plane,RLE;對應已驗證的 tools/title_pcx.py) |
| `re/sdl/sdl_main.c` | PoC main:載入標題 → 解碼 → SDL 顯示 + 最小事件迴圈 |
| `re/sdl/CMakeLists.txt` | 跨平台 build(Linux/Windows/macOS;find SDL2) |

設計套用 deep-modules(窄介面藏複雜度)與 retro-cjk-hires-canvas
(內部畫布 nearest-neighbor 放大、`DQ3_SCALE` 倍率、CJK 預留 24×24 點陣空間)。

### 3.2 runtime shim 對應表(DOS far-call → SDL2)

| 原版 runtime(docs/06) | SDL2 shim |
|---|---|
| seg 0x111b VGA planar 繪圖 | framebuffer 4bpp packed → ARGB8888 texture(palette 查表)+ `SDL_RenderCopy` nearest 放大 |
| seg 0x1104 鍵盤 scancode | `SDL_PollEvent` → `dq3_poll_scancode()` 回 BIOS scancode(0x50/4b/48/4d…) |
| seg 0x10b6 滑鼠 | (待補)SDL mouse |
| seg 0x1053 檔案 I/O | `dq3_load_file()` = fopen/fread |
| BIOS tick / sub_ee23 frame_wait | `dq3_delay_ms` / `dq3_ticks_ms` |
| seg 0x129c Sound Blaster | (待補)SDL_mixer / SDL audio |

### 3.3 build + headless 驗證(docker,不污染 host)

```bash
docker build -f tools/build/Dockerfile.sdl -t dq3-sdl tools/build
tools/build/sdl_build.sh        # cmake build + headless 解碼 TITG.P → PPM
```

`sdl_build.sh` 用 `SDL_VIDEODRIVER=dummy` 在容器內 build 並跑 headless 解碼驗證
(`DQ3_DUMP=out.ppm` 模式只跑 loader+decoder+dump,不開視窗,適合 CI)。

### 3.4 PoC 結果(關鍵里程碑:解碼資產 → SDL 顯示管線通)

- `re/sdl` 在 SDL2 dev + cmake 容器內**編譯通過**(dq3_sdl,23488 bytes)。
- headless 載入 `assets_raw/TITG.P`(39610 bytes)→ PCX 解碼 → dump 640×350 PPM。
- PPM 轉 PNG 檢視:**正確還原標題畫面**「DRAGON FIGHTER III / 傳說的終章 / © 1993 精訊資訊有限公司」,
  11 色 EGA palette,非空白非雜訊。

→ 證明「**已解的 loader/格式 + SDL2 shim + 顯示管線**」在現代跨平台環境可運作。
這是 remake 的地基:資產解碼正確、palette 正確、planar→packed→RGB 正確、視窗顯示正確。

---

## 4. 完美 RE / remake 的混合策略與 gap

最終驗證以 **DOSBox 跑原版當黃金對照**(oracle,docs/14),SDL 版日後逐畫面比對。
function 採混合策略:

- **能反編譯的** → 乾淨 C(可讀、可維護、可移植到 SDL2)。
- **攻不下的**(手寫組語 runtime、極致最佳化、奇怪控制流)→ 不強解 C;
  在 SDL2 port 直接以「等價語意的 C/SDL shim」重寫(不需 byte-identical,只需行為一致)。
- 若要走「byte-identical 重組」驗證理解度(§2 路線),屬另一條可選佐證線,**非 remake 必要**。

### 到完整可玩 remake 還缺(gap)

1. **主迴圈接線**:把 re/mainloop.c 的主迴圈改成呼叫 dq3_runtime shim(讀 scancode → 移動 →
   狀態機跳表 → 繪場景),取代 far-call wrapper。
2. **地圖/tile 繪製**:把 BLK/SHP tile(docs/04)+ CTY 城鎮佈局接到 framebuffer;
   draw_scene(sub_255b)/ update_player(sub_19b8)的捲動與碰撞改 SDL。
3. **palette cycling**:海面動畫(docs/04 謎團)在 shim 層做 palette rotate。
4. **文字/對話**:D3TXT*.TXT 控制碼(docs/03)+ CHINA.FON / d3txt00.fon 點陣 → framebuffer;
   套 retro-cjk-hires-canvas(內部畫布拉高、CJK 24×24)。
5. **音效**:Sound Blaster(seg 0x129c)→ SDL audio / SDL_mixer;BLS 音樂格式待解。
6. **存檔 / 選單 / 戰鬥**:re/commands.c、re/battle.c、re/nameinput.c 邏輯接 shim。
7. **跨平台打包**:Windows(SDL2 mingw)、AppImage(Linux)、universal `.app`(macOS)。

### 殘留 / 受阻

- byte-identical 重組路線未做(本案改走 SDL2 remake,優先級下調);若要做,
  最確定的是 §4 提的「反組譯→NASM/Python 重組整段 code,sha256 對齊」——尚未實作。
- MSC exact point-release / flag 未掃到 100% 逐函式 match(§2.4 gap),不影響 remake。
- runtime 段(VGA/SB)未逐指令反編譯為 C;remake 走「重寫等價 shim」而非還原該段。
