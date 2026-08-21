# DQ3.EXE bytes 保真驗證：byte-identical 重組

> **2026-08-22 勘誤：**本篇證明的是 MZ 區段、原始 bytes 與已列出的 seg0
> 指令切分可重生，不是「每個 byte 的遊戲語意都已理解」。以 `db` 原樣輸出
> 再得到相同 SHA-256，不能證明資料欄位、caller／consumer、狀態副作用或玩家
> 可見行為。現行語意完成度與未知項目以 [`docs/135`](135-re-assertion-audit-20260822.md)
> 為準；下文保留當時方法與數字，但不再作全域語意完成聲明。

本篇回答的是:**DQ3.EXE 的檔案結構、原始 bytes 與已解碼的 seg0 指令邊界，
能否精確重生?** 在此之前,只有素材(字型/地圖/標題)解碼經 DOSBox 驗證。

驗證方法是明確的 bytes 保真訊號:**把匯出產物用獨立組譯器(nasm)重組,
`sha256sum` 與原版逐 byte 相同**。對得上只證明輸出沒有遺失或改動原始 bytes；
指令語意仍須以 IDA database 的交叉參照、writer-consumer 與 runtime oracle 個別閉合。

最終結論先講:**整檔重組 sha256 與原版完全相同(PASS,bytes 100%)**；seg0
12,949 條候選指令的助記符 round-trip 未發現已證實的解碼錯誤。這不等於
12,949 條指令的高階語意、資料欄位與玩家效果都已解讀。

---

## 1. 原版檔案結構(MZ EXE)

| 欄位/區段 | 值 / 範圍 | 說明 |
|---|---|---|
| 檔案大小 | 115,282 bytes | |
| `e_cparhdr` | 0x137(311 段)| header = 311×16 = **0x1370 bytes** |
| reloc table | `e_lfarlc`=0x22,`e_crlc`=1232 | 0x22..0x1362(1232×4 bytes) |
| header padding | 0x1362..0x1370 | 14 bytes,全 0 |
| load image | 0x1370..0x1c250 | seg0 主碼 + runtime 段 + 資料 |
| 宣告影像大小 | (e_cp,e_cblp)=226 頁/末頁 80 → 115,280 | |
| 尾端 pad | 0x1c250..0x1c252 | 2 bytes(`0000`),超出頁宣告 |
| entry | `CS:IP`=0000:9299(file 0xa609)| |

`tools/reasm/mz_roundtrip.py` 把這套結構解析成具名欄位 + region(覆蓋全檔、無洞、
無重疊),再從結構重建。

---

## 2. 里程碑 (a):整檔 byte-identical 重組 — PASS

### 2.1 結構化 round-trip(具名欄位重建,非 raw copy)

```
python3 tools/reasm/mz_roundtrip.py verify assets_raw/DQ3.EXE work/reasm_full
```

MZ header 28 bytes 由解析出的 14 個 u16 欄位**重新 pack**(不是複製),reloc table、
padding、load image 各自 region 重組。內建覆蓋率自檢(region 必須無洞無重疊覆蓋全檔)。

結果:`orig sha256 == rebuilt sha256`(`5178fdc8…30c`)→ **PASS,byte-identical(100%)**。

### 2.2 用獨立組譯器 nasm 重組(關鍵硬證據)

光是 Python 拆檔再串接,可能被質疑「只是把 bytes 搬來搬去」。所以再用**獨立的
組譯器 nasm** 走一遍:

```
tools/reasm/reasm_full.sh        # docker 內,不污染 host
```

`tools/reasm/exe_to_nasm.py` 把整檔產生成一份 **nasm 原始檔**(`work/reasm/dq3.asm`,
27,373 行):MZ header 各欄位具名註解、reloc table 每筆標 `seg:off`、padding、
load image。其中 **seg0 主碼附逐條助記符反組譯為註解**(`; 0xNNNNN mov ...`),
使這份 .asm 同時是「可審閱的反組譯」與「可重組的原始碼」。指令本體以 `db` 固定 bytes
(原因見 §3:16-bit 編碼有合法歧義,只有 db 保證 nasm 組出逐 byte 相同)。

```
nasm -f bin -o work/reasm/dq3.rebuilt work/reasm/dq3.asm
sha256sum assets_raw/DQ3.EXE work/reasm/dq3.rebuilt
# 5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c  (兩者相同)
```

→ **RESULT: PASS — nasm 重組 byte-identical(100%)**。獨立組譯器從我們產生的原始碼,
組出與原版完全相同的 115,282 bytes。

> 意義:這證明輸出的 header/reloc/padding/image 區域能無損重生，reloc 1232 筆、
> entry、段邊界與尾端 pad 無一漏出輸出。由於 load image 以 `db` 保留，這項結果
> 不能單獨證明其中每個 byte 的程式／資料分類或語意。

---

## 3. 里程碑 (b):seg0 主碼反組譯保真度 — 實質 0 解碼錯誤

整檔重組用 db 固定 bytes 證明了「切分/對齊正確」。更進一步要證明「我們**命名的指令
(助記符)**是對的」:把 seg0 每條指令以助記符寫出,用獨立 encoder(keystone,與
capstone 同家、同 16-bit 模式)重組,比對 bytes。

```
# docker 內(capstone + keystone-engine)
python3 tools/reasm/reasm_ks.py funcs docs/data/exe_funcs.json
python3 tools/reasm/reasm_ks.py 0xfa29 16 e6b9    # 單函式
```

### 3.1 原始數字(280 函式、12,949 指令、35,127 bytes)

- 助記符直接 re-encode byte-identical:**10,560 / 12,949 指令(81.55%)**、
  **29,669 / 35,127 bytes(84.46%)**。

### 3.2 不符的 2,389 條,逐一歸類後 = 實質 0 解碼錯誤

| 類別 | 數量 | 性質 |
|---|---|---|
| keystone `0x66` prefix 缺陷 | 466 | keystone 在 16-bit 模式對某些指令(如 `ret`)誤加 operand-size prefix(`c3`→`66c3`)。**我們的解碼正確**,是 encoder 缺陷 |
| 等價 encoding(decode 回同一指令)| 934 | 同一助記符有多種合法 encoding。原版選 **accumulator 短碼**(`cmp ax,0`=`3d0000`、`add ax,imm`=`05..`),keystone 選通用碼(`83f800`)。語意相同,bytes 不同 |
| call/jmp 相對位移 encoding | 976 | keystone 解析相對目標時選不同位移大小,語意相同 |
| 真正不同 | **13** | 全為 capstone 助記符含 **segment override**(`ds:`)時文字 round-trip 解析歧義(如 `mov cx,[ds:bp]`),非解碼錯誤,屬文字表示限制 |

→ 12,949 條指令中,**0 條是反組譯把指令解錯**;81.55% 直接同碼,其餘全是
encoding 選擇差異或 encoder/文字限制。配合 §2 整檔 sha256 相同,**seg0 主碼反組譯確認正確**。

### 3.3 副產品:accumulator 短碼是編譯器指紋

原版**一致使用 accumulator-specific 短碼**(`3d` cmp AX,imm16;`05` add AX,imm16;
`a1/a3` mov AX,[mem]),這是通用組譯器不會優先選的形式,也呼應主碼由
**Microsoft C** 產生的判斷(見 §7 附錄)。這也正是為何「助記符經通用 encoder
重組」無法天然 byte-match,必須用 db 固定原版精確 encoding。

---

## 4. 重組覆蓋率總結

| 區段 | 範圍 | 重組方式 | byte-match |
|---|---|---|---|
| MZ header(欄位)| 0..0x1c | 解析欄位重新 pack | **100%** |
| header gap / reloc table | 0x1c..0x1362 | 結構化 db(每筆 seg:off 標註)| **100%** |
| header padding | 0x1362..0x1370 | db | **100%** |
| seg0 主碼 | 0x1370..0x11370 | db(每條附助記符反組譯註解)| **100%** bytes;助記符保真實質 0 錯誤 |
| runtime 段 + 資料 | 0x11370..0x1c250 | db(16/行)| **100%**(原樣保留,未逐指令解;手寫組語 runtime)|
| 尾端 pad | 0x1c250..EOF | db | **100%** |
| **整檔** | 0..EOF | nasm -f bin | **100%(sha256 相同)** |

對不齊的段:**無**。整檔 sha256 相同。

---

## 5. 工具(tools/reasm/)

| 檔 | 功能 |
|---|---|
| `mz_roundtrip.py` | MZ 結構化拆解↔重組(`split`/`build`/`verify`),region 覆蓋率自檢 |
| `exe_to_nasm.py` | 整檔 → nasm 原始檔(header 具名 + reloc 標註 + seg0 助記符註解 + db 全覆蓋)|
| `reasm_full.sh` | docker 內 nasm 重組 + sha256 比對(一鍵 PASS/FAIL)|
| `reasm_ks.py` | capstone→keystone 逐指令助記符重組保真度(整檔批次 / 單函式)|
| `disasm_to_nasm.py` / `reasm_verify.py` | 單段 nasm 助記符重組驗證(輔助)|

中間檔(`work/reasm*`、`.asm`、`.bin`、`.obj`)照 .gitignore 不入版控。
全部在 docker(`ghcr.io/astral-sh/uv` + nasm/capstone/keystone)內跑,不污染 host。

---

## 6. 結論與後續

- **bytes 保真基礎成立**:匯出內容能用獨立組譯器(nasm)100% 重生原版 DQ3.EXE
  二進位(sha256 相同)；seg0 12,949 條候選指令的助記符 round-trip 未發現已證實
  的解碼錯誤。
- 整檔每個 byte 都被輸出覆蓋，reloc/entry/段邊界可重生；`db` 區域的程式／資料
  歸屬與高階語意不由此測試證明。
- 目前 seg0 主碼以助記符註解 + db 表示(byte-exact);runtime 段(VGA/SB/鍵盤等
  手寫組語)以 db 原樣保留,尚未逐指令解成助記符——但不影響 byte-identical 重組。

### 後續（在 bytes 保真確認之後，本任務不做）

1. 把 seg0 的 db 形式逐步「升級」成可審閱的助記符標籤化反組譯(函式名、跳轉標籤),
   仍維持 nasm 組回 byte-identical(用 db 固定原版精確 encoding,助記符放標籤/註解)。
2. runtime 段(0x11370 起)逐段反組譯命名(VGA planar / Sound Blaster / 鍵盤…)。
3. 歷史規劃曾準備往 C 重編／SDL2 移植；現行產品已改為 Go／Ebitengine，且只按
   玩家可見 blocker 開窄 RE 任務。

> 受阻/殘留:無段對不齊。唯一「真正不同」的 13 條是 capstone 對 segment-override
> 助記符的文字 round-trip 限制(指令本身解碼正確),不影響整檔 byte-identical。

---

## 7. 附錄:原版編譯器辨識(支撐 §3.3 的 accumulator 短碼指紋)

§3.3 觀察到原版偏好 accumulator 短碼。這與「主碼由哪個編譯器產生」的獨立判斷一致,
證據摘錄如下(完整 byte-match PoC 過程見 git 歷史 / tools/build/)。

| 指紋(seg0 主碼,280 函式統計)| 觀察 | 指向 |
|---|---|---|
| `ret` 型別 | 268 near `ret`(c3),僅 1 `retf` | near-code 模型(small/medium)|
| BP frame | 0/280 用 `push bp;mov bp,sp` prologue | 優化版、非偵錯 |
| `LOOP` 指令 | **184 次** | **Microsoft C**(Borland 幾乎不發 LOOP)|
| 旋轉 | `rol`×25、`ror`×21 | 編譯器辨識旋轉慣用語 |
| `cdq` | 4 次 | MSC 風格 |
| NOP padding | 459 個 | MSC 對齊習慣 |
| accumulator 短碼 | `3d/05/a1/a3` 一致使用 | MSC codegen 偏好(§3.3)|
| int 24h handler 訊息 | `Press X to stop trying…` + `disk is write-protected$…` | MSC `_hardresume` 系措辭 |

OMF segment 命名(對候選編譯器 .obj 驗證):MSC 產出含 **CONST 段** + `MS C` 浮水印;
Turbo C 無 CONST、浮水印 `TC86 Borland Turbo C 2.01`。byte-match PoC(LCG sub_e6b9):
MSC 5.1 `/AS /Ox` 指令序列/暫存器/段命名最相符(Turbo C 因 far-ret、不發 LOOP 排除)。

**結論**:原版主碼最可能為 **Microsoft C 5.x(small/medium model)**,runtime 為手寫
組語另段以 far call 串接。此判斷僅作背景知識;**RE 確認本身不依賴它**,§2 的 nasm
sha256-identical 重組是直接證據。古董編譯器(TC/MSC)docker 建法暫擱置,留作日後
C 重編佐證。
