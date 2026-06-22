# C 重編(matching decompilation):splice 框架與 byte-match 覆蓋率

[`docs/17`](17-build-toolchain.md) 已證明:我們的反組譯能用獨立組譯器(nasm)100%
還原原版 `DQ3.EXE`(sha256 相同)。那是「ASM 重組」層級——seg0 主碼以 `db` 固定 bytes
表示。本篇處理下一步「正路 (b)」:**把 seg0 的函式逐一反編譯成 C,用原版編譯器
(Microsoft C 5.1)編回去,且編出來的機器碼與原版該函式逐 byte 相同**(matching
decompilation,如 N64 / SM64 decomp)。這份文件記錄為此建立的**管線基礎設施**——
框架本身,不是「已 match 多少函式」(那是另一條長期工作)。

## 為什麼是「splice」框架

不需要等所有函式都反編譯完才能重組出可執行檔。採「split」策略:

- 能反編譯成 C 且**已 byte-match** 的函式 → 由 MSC 5.1 編成 `.obj`,抽出機器碼,
  **splice(覆蓋)進 docs/17 的重組產物**,取代該函式對應的 `db` bytes。
- 尚未/不可能成 C 的部分(runtime 段、資料、MZ header / reloc / padding)→ 沿用
  docs/17 的 ASM 重組(原版 `db` bytes)。
- 把兩者組/連成 EXE,`sha256` 與原版比。

### 核心不變量(為什麼框架一定正確)

| inv | 內容 |
|---|---|
| **inv.1** | matching-decomp 的判準 = 編出來的函式 bytes 與原版該函式 `[lo,hi)` **逐 byte 相同**。因此一個「已 match」的 C 函式 splice 進去,**不會改變** seg0 內容 → 整檔 sha256 不變。 |
| **inv.2** | 0 個 C 函式時,seg0 = 原版 seg0 → 整檔重組 sha256 **== 原版**(框架正確性的基準)。 |
| **inv.3** | 函式邊界 `[lo,hi)` 取自 [`docs/data/exe_funcs.json`](data/exe_funcs.json)(seg0 邏輯 offset)。splice 寫入位置 = seg0 buffer 的 `[lo:hi)`;長度必須吻合(C obj 函式 byte 數 == `hi-lo`)。 |

inv.1 + inv.2 一起保證:**只要每個被 splice 的函式真的 exact-match,整檔 sha256
就維持與原版相同**——隨 match 函式增加,愈來愈多 `db` bytes 被「具語意的 C source」
取代,但二進位輸出不變,逐步逼近 100% matching decompilation。

> 反過來說:若某個被宣告為 match 的函式其實沒 match,splice 後 sha256 會立刻變,
> 重組失敗——這正是 byte-identical 重組當作 pass/fail 訊號的價值。為避免「半 match」
> 污染重組,工具**只 splice 通過 exact-match 的函式**,未 match 者維持原版 db。

## 架構與資料流

```
re/match/*.c  ──(MSC 5.1, docker dq3-msc)──▶  *.obj
     │  @match entry=0xNNNN                        │ OMF LEDATA
     │                                             ▼
     │                              抽函式機器碼(splice_lib.extract_obj_code)
     │                                             │
     │                          與原版 [lo,hi) 逐 byte 比對(check_func_match)
     │                                             │
     │                              exact-match? ──┴── 否 → 不 splice(維持原版 db)
     │                                   │ 是
     ▼                                   ▼
原版 seg0 bytes ──── build_spliced_seg0(覆蓋 [lo:hi)) ──▶ spliced seg0
                                                              │
            原版 MZ header / reloc / padding / runtime+data ──┤ emit_asm
                                                              ▼
                                                        dq3.asm(nasm)
                                                              │ nasm -f bin
                                                              ▼
                                                  work/RE-DQ3.EXE ── sha256 比對原版
```

## 工具(`tools/build/`)

| 檔 | 角色 |
|---|---|
| `splice_lib.py` | 共用核心:解析 MZ / seg0、讀 `exe_funcs.json`、抽 OMF 函式 bytes、`check_func_match`、`build_spliced_seg0`、`emit_asm`、`coverage_report`。純函式,不跑 docker。OMF 解析重用 `tools/re_match.py`。 |
| `match_check.py` | byte-match 覆蓋率工具(見下)。 |
| `splice_rebuild.py` | docker 內被 `rebuild.sh` 呼叫:抽 exact-match 函式 bytes、splice、產 `.asm`。 |
| `rebuild.sh` | full-rebuild:兩階段(host 編 .obj → docker splice+nasm+sha256)。 |
| `msc_compile.sh` | (既有)docker 內單檔 `.c` → `.obj`。 |

中間產物在 `work/csplice/`(gitignore)。

## 用法

### byte-match 覆蓋率:`match_check.py`

每個 `re/match/*.c` 宣告它對應哪個原版函式,二擇一(優先序由上而下):

```c
// @match entry=0xe6b9            (明確 marker;entry = exe_funcs.json 的 seg0 邏輯 offset)
// @match entry=0xe6b9 model=/AS  (可選 memory model,預設 /AS small)
```

或沿用 `sub_XXXX` 命名慣例——**檔名或函式名**含 `sub_<hex>`(如 `sub_e6b9.c` /
`unsigned sub_e6b9(void)`),工具會把 `XXXX` 當 seg0 offset 推出 entry,既有 match
檔免加 marker 即可被檢查;檔頭若有 `model=/A?` 也會被採用。

```bash
tools/build/match_check.py                      # 掃 re/match/*.c
tools/build/match_check.py re/match/foo.c       # 指定檔
tools/build/match_check.py --json work/match_report.json
tools/build/match_check.py --no-compile         # 用已存在 .obj(除錯)
```

每個函式輸出 `MATCH` / `DIFF`、`match/total`、`match%`、size 是否吻合;不符時印
first diff + orig/cand hexdump。最後印整體覆蓋率(exact-match 函式數 / 280、match
bytes / 函式總 bytes / seg0)。編譯失敗或 entry 不存在 → 回非 0(供 CI)。

### full-rebuild:`rebuild.sh`

```bash
tools/build/rebuild.sh
```

階段 1(host docker `dq3-msc`)編每個 `re/match/*.c` 成 `.obj`;階段 2(docker
`uv`+nasm)splice exact-match 函式、nasm 組成 `work/RE-DQ3.EXE`、`sha256` 比對原版,
印覆蓋率。docker 同步前景,不污染 host。

## 目前覆蓋率與驗證狀態

| 項目 | 狀態 |
|---|---|
| 框架基準(inv.2):**0 函式時 `RE-DQ3.EXE` sha256 == 原版** | ✅ PASS(`5178fdc8…530c`) |
| splice 機制(inv.1):用原版 bytes splice 多個函式 → 仍 byte-identical | ✅ PASS(self-test 5 函式) |
| 覆蓋率工具(compile→extract→compare)端到端可跑 | ✅ 可跑(對非 match 函式正確報 DIFF,不誤判) |
| **已 exact-match 的 C 函式數** | **0 / 280**(本任務不寫 match C;由另一工作推進) |

亦即:**框架就緒,覆蓋率可量測,0 函式時整檔 byte-identical**。

## 到 100% 的 gap

1. **match C source(主要工作量)**:280 個 seg0 函式逐一反編譯成 MSC 5.1 能編出
   byte-identical 機器碼的 C。難點在 codegen 細節——docs/17 §3.3 指出原版偏好
   accumulator 短碼(`3d`/`05`/`a1`/`a3`)、near ret、特定 NOP padding、CONST 段、
   `rol` 旋轉展開,需逐函式調 C 寫法 + 編譯 flag 逼出相同指令序列。
2. **多函式單 obj / 符號定位**:目前 `extract_obj_code` 取 .obj 最大 code 段
   (適合單函式 .c)。當一個 .c 含多函式或編譯器混入 CONST/DATA 段時,需改以
   OMF 符號表(PUBDEF / LNAMES)精準定位目標函式 bytes,而非靠 size 比對排除。
3. **跨函式重定位(reloc / far call)**:函式內含絕對位址或 far call 時,`.obj` 的
   bytes 帶 fixup 佔位(常為 0),與原版已 link 的絕對位址不同。要嘛比對「去 fixup
   後的骨架」,要嘛把 splice 推進到「link 階段」用 MSC LINK 串 .obj + 原版 reloc。
   目前框架在 raw bytes 層 splice,適用無外部 reloc 的純運算函式;有 reloc 的函式
   先標記、暫不 splice(維持原版 db,不影響 sha256)。
4. **runtime 段(0x11370 起)**:VGA planar / Sound Blaster / 鍵盤等手寫組語,非 C
   產物,維持 ASM 重組(db),不在 C 重編範圍。

> 設計上,以上 gap 都不破壞 inv.2:任何尚未解決的函式維持原版 db,整檔 sha256
> 持續 == 原版。覆蓋率只會隨真正 exact-match 的函式單調上升。
