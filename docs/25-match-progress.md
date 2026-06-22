# 逐函式 byte-match 進度 (matching decompilation)

正路 (b):把 seg0 函式逐一寫成「用 **MSC 5.1** 編出 byte-identical 原版機器碼」的 C。
本篇記錄第一批 leaf 函式的 byte-match 嘗試、可重複的 workflow、各函式 match% 與殘差成因,
以及下一批的建議。

> 前置:編譯器已鎖定 **Microsoft C 5.x small/near model**(指紋見 `docs/17`、`docs/19`)。
> 整檔已用 nasm 以 db 形式 byte-identical 重組(`docs/17` §2,sha256 相同)。本篇是更高一層:
> 不用 db 固定原版 bytes,而是**從 C 重新編出**相同機器碼。

## 結論先講

- 本批挑了 5 個最簡單的 leaf 函式(純全域 / 旋轉 / port out / memcpy / 填表迴圈)。
- **沒有一個達到 100% byte-identical**;但對「無 local、無迴圈」的直線函式,**指令選擇與結構已逼近原版**
  (`sub_e91e` frameless、長度精確相同、指令集完全一致,僅首對 `mov dx`/`mov ax` 載入順序相反)。
- 卡點全部歸類清楚(見「殘差分析」),都是 **MSC 5.1 codegen 的固定選擇**(旋轉編碼、intrinsic 引數求值順序、
  memcpy 策略、**有 local 必開 BP frame**),不是 C 寫法錯誤。
- 最關鍵的單一卡點:**這顆 MSC 5.1 build 對任何「含 local 變數」的函式都會發 `push bp; mov bp,sp` 並把
  變數配到 stack/SI/DI**,與原版 280 函式全為 frameless + `LOOP` + BX 定址的指紋衝突。解掉這個,迴圈類函式可大幅提升。

## 可重複 workflow

```bash
# 0) 一次性:建一顆帶 capstone 的 python image (host 不污染)
#    Dockerfile: FROM python:3.12-slim; RUN pip install capstone   → tag dq3-recap

# 1) 讀原版該函式反組譯
docker run --rm -v "$PWD":/w -w /w dq3-recap python3 tools/re_match.py orig <seg0_off_hex> <size>

# 2) 寫等價 C → re/match/sub_XXXX.c

# 3) MSC 5.1 編 (DOSBox in docker)
tools/build/msc_compile.sh re/match/sub_XXXX.c /c /AS /Ox   # → re/match/sub_XXXX.as.msc.obj

# 4) 抽 OBJ code bytes 與原版逐 byte 比對
docker run --rm -v "$PWD":/w -w /w dq3-recap python3 \
    tools/re_matchbatch.py <seg0_off_hex> <size> re/match/sub_XXXX.as.msc.obj
#  或整批:re_matchbatch.py manifest tools/re_match_manifest.json
```

`tools/re_matchbatch.py` 同時報兩個數字:
- **raw match%**:逐 byte 完全相同的比率。
- **masked match%**:把「16-bit 絕對位址運算元」(全域 / 陣列的 linker 指派位址,屬 data reloc 差異,
  原版位址 ≠ 我們的 linker 位址)遮罩後的比率 — 用來看「指令序列 / opcode / 結構」是否一致,
  排除掉「位址不同」這種與 codegen 無關的差異。

> 注意:`msc_compile.sh` 用 DOSBox,單次約 30–60s 且序列化(同時間只跑一個 DOSBox)。
> 產出檔名只依 model tag(`/AS`→`.as.`),**不含優化 flag**,故換 flag 重編會覆蓋同名 obj。

## 本批結果

| 函式 | seg0 | size | raw% | masked% | 用的 flag | 狀態 |
|---|---|---|---|---|---|---|
| `sub_e6b9` | 0xe6b9 | 16 | 31% | 62% | `/AS /Ox` + `#pragma intrinsic(_rotl)` | 頭尾同碼,旋轉編碼殘差 |
| `sub_e91e` | 0xe91e | 20 | 70% | 70% | `/AS /Ox` + `#pragma intrinsic(outpw)` | frameless、長度相同、指令集全同,僅首對載入順序 |
| `sub_edab` | 0xedab | 18 | 0% | 23% | `/AS /Ox` + `#pragma intrinsic(memcpy)` | memcpy 策略不同(見下) |
| `sub_e96d` | 0xe96d | 21 | 2% | 7% | `/AS /Ox` | 迴圈,BP frame + SI/DI 配置殘差 |
| `sub_32a3` | 0x32a3 | 17 | 0% | 9% | `/AS /Ox`(`/Gs` 無差) | 迴圈,BP frame + SI/DI 配置殘差 |

對齊最好的兩個:

```
sub_e91e (VGA GC 暫存器初始化, 4 x outpw):
  orig: bace03 b80300 ef b80508 ef b80700 ef b808ff ef c3   (mov dx,3ce; mov ax,3; out; ...)
  ours: b80300 bace03 ef b80508 ef b80700 ef b808ff ef c3   (mov ax,3; mov dx,3ce; out; ...)
        ^^^^^^^^^^^^^^  只有首對 mov 順序相反;其後 14 bytes 完全相同

sub_e6b9 (LCG 亂數前進):
  orig: a1 5a0b | 05 1890 | d1c0 d1c0 d1c0 | a3 5a0b | c3   (mov ax,[g]; add; rol×3; mov [g]; ret)
  ours: a1 0000 | 05 1890 | b103 d3c0       | a3 0000 | c3   (mov ax,[g]; add; mov cl,3+rol cl; mov [g]; ret)
        頭 (mov ax,[g]; add ax,0x9018) 與尾 (mov [g],ax; ret) 同碼;
        位址 5a0b vs 0000 屬 data reloc;只有中間旋轉編碼不同。
```

## 殘差分析(逐一歸類,均為 MSC 5.1 codegen 固定選擇)

1. **旋轉編碼(`sub_e6b9`)**
   原版三個 `rol ax,1`(`d1c0 d1c0 d1c0`,6 bytes,each = D1 /0 立即 1 形式)。
   MSC 5.1 對 C 旋轉慣用語 `(x<<3)|(x>>13)` **不認得**,會展成 `shl/shr cl + or`(更長)。
   用 `_rotl` intrinsic 才接近,但 intrinsic 一律 `mov cl,n; rol ax,cl`(`b103 d3c0`,4 bytes),
   即使 `_rotl(x,1)` 也不發 `d1c0`。
   → **MSC 5.1 的 `_rotl` 不會發 1-bit 立即旋轉**;原版那三個 `rol ax,1` 極可能來自 **inline asm**
   或更晚版本 MSC codegen。這顆 5.1 純 C 無法重現該 6-byte 形式。

2. **intrinsic 引數求值順序(`sub_e91e`)**
   `outpw(port, val)` intrinsic 在 `/Ox` 下先把 `val` 載入 ax、再把 `port` 載入 dx(右到左 cdecl 求值),
   原版相反(先 `mov dx,port` 再 `mov ax,val`)。其後三次 out 因 dx 不變、只重載 ax,**完全對齊**。
   → 只差首對兩條 `mov` 的順序;intrinsic 求值順序非 C 端可控(試 `/Os` 反而退化成真的 call)。

3. **memcpy 策略(`sub_edab`)**
   原版 = `push es; push ds; pop es; mov si,[src]; lea di,[dst]; mov cx,0x30; rep movsb; pop es`
   (純 byte 複製,不存 si/di,顯式 ES=DS)。
   MSC `memcpy` intrinsic = `push si/di` + word/byte 拆分(`shr cx,1; rep movsw; adc cx,cx; rep movsb`)。
   → 兩種不同 memcpy 實作;原版那種 frameless + 不存 si/di + 純 movsb,像手寫或 `movmem` 風,
   非 MSC 5.1 memcpy intrinsic。

4. **BP frame + 暫存器配置(`sub_e96d`, `sub_32a3`)— 最關鍵卡點**
   原版迴圈函式:**無 BP frame**、計數放 **CX 用 `LOOP`**、index/指標放 **BX/DI**、累加值放 **AX**。
   這顆 MSC 5.1 build:**只要函式有 local 變數就一定發 `push bp; mov bp,sp; sub sp,N`**,
   且把 register 變數配到 **SI/DI** 並以 `dec+jne` 取代 `loop`,結尾還把變數 spill 回 frame。
   試過 `register` 關鍵字、`/Gs`(無 stack probe)均無法去掉 frame。
   → 這與原版 280 函式 **全 frameless + `LOOP` 184 次 + BX 定址** 的指紋直接衝突。
   推測原版的 frame omission 來自某個尚未找到的 flag(或不同 build 的 MSC 5.x)。
   **解掉這點,所有迴圈類 leaf 函式可望大幅提升甚至 100%。**

5. **資料絕對位址(全函式共通,非 codegen 差異)**
   全域 / 陣列的 16-bit 絕對位址(如 `0xb5a`、`0x265d`、`0x289a`)由 linker 指派,與原版不同。
   單獨編譯(`/c`)的 obj 此處為 `0000` 並帶 reloc record。這屬 **data relocation 差異**,
   待整體 link、佈局對齊原版 DGROUP 後可一併消除;`re_matchbatch.py` 的 masked% 已把這類位元組排除。

## 下一批建議

1. **優先攻破 frame omission**:這是迴圈類函式 byte-match 的總開關。可試方向:
   - 比對原版某個「確定有 local」的函式,逆推它用的 MSC flag 組合;
   - 試 `CL` 其他優化 flag(`/Ol` loop opt、`/Oa`、`/Oc`)或不同 MSC 5.x sub-version;
   - 確認 docs/17 §7 統計的 frameless 是否其實對應「source 無 local」的函式子集。
2. **直線無 local 函式**先做(最好控):VGA / port I/O 初始化序列(`out` 系列)、純全域指派、
   常數表寫入。`sub_e91e` 已證實 frameless + intrinsic 路可行,只差求值順序。
3. **避開 intrinsic 求值順序問題**:找只有「單一 out / 單一 memcpy」或載入順序無歧義的函式。
4. **旋轉 / 特殊指令**:若原版大量用 `rol r,1`,評估是否原始碼即含 inline asm;此類保留 db 形式
   (見 docs/17 §6 後續 1),不強求純 C。
5. **data reloc 收斂**:待累積數個函式後,做一次「整批 link + 對齊 DGROUP 佈局」實驗,
   把絕對位址也對上,屆時 raw% 才能反映真實 byte-identical。

## 受阻誠實揭露

- 本批 **0 個達 100%**。最接近 `sub_e91e`(frameless、長度相同、指令集全同,僅首對載入順序)、
  `sub_e6b9`(頭尾同碼)。
- 主因是 **MSC 5.1 codegen 固定選擇**(旋轉編碼、intrinsic 求值順序、memcpy 策略、**有 local 必開 frame**),
  非 C 判讀錯誤。
- workflow(讀原版 → 寫 C → MSC 5.1 編 → 抽 obj 比對 + masked%)**已可重複**,
  工具 `tools/re_matchbatch.py` + `tools/re_match_manifest.json` 一鍵整批。
- 推進 100% 的單一最高槓桿:**找回 MSC 5.x 的 frame-pointer omission 行為**(見「下一批建議」1)。
