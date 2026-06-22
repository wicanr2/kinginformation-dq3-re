# 精訊 DQ3 數值類 bug 修正(#4 勇者 MP / #5 升級錯亂 / #6 數值 255 溢位)

這三個 bug 青衫攻略沒有官方修正碼(見 docs/21),需自行從反組譯定位後修。
本文記錄根因核對、安全修法與 DOSBox 驗證,並修正 docs/18 早期分析的兩處誤判。

工具:`tools/re_stat_patch.py`(docker capstone 內跑);機讀 patch:`docs/data/stat_patches.json`。

## 摘要

| # | bug | 根因 | 修法 | 狀態 |
|---|---|---|---|---|
| 4 | 勇者 MaxMP 成長偏低 | EXE 靜態成長表勇者 MP base/slope 偏低 | EXE in-place 資料 patch(2 byte) | 已修,DOSBox 開機驗證通過 |
| 5 | 高等級(Lv44+)升級錯亂 | 門檻表 MAX_LEVEL=43,lv44 越界讀 0 → 連升 | clamp level≤43;需 code cave | 未修(cave 不足,留 C 層) |
| 6 | 數值 255 溢位 | 非成長公式(已澄清);疑顯示寬度 / 與 #5 互動 | uint16 + 明確 clamp | 未修(無可信靜態截斷點,留 C 層) |

## 修正 docs/18 的兩處誤判

docs/18 對 #4/#6 的定位有兩處與實際不符,經本次反組譯核對更正:

1. **成長表不在 dragon0.dat,而在 EXE 靜態映像。** dragon0.dat 實為 2172 byte 的存檔
   (內容多為 0xff 空槽),不是成長表。成長表是初始化資料,位於 DGROUP(進入點
   `mov ax,0x14dd; mov ds,ax`)的 `DS:0x4366`,換算 file offset = `0x1370 + 0x14dd×16 +
   0x4366 = 0x1a4a6`,仍落在 115282 byte 檔案內。因此 #4 可直接做 EXE in-place 資料 patch,
   不需 dump dragon0.dat。

2. **成長公式不是 #6 的來源。** 公式 `target = base + (slope×level)>>1` 全程 16-bit:
   `mul dl`(slope×level → AX)、`shr ax,1`、`add al,base; adc ah,0`(8-bit 加但有正確
   進位傳遞到 AH),產出正確 16-bit target;增量再以 word 加進屬性欄位
   (`add word ptr [si+..]`)。屬性欄 `[si+0x24]`/`[si+0x26]` 全檔只有 word(8b/89)讀寫、
   查無 byte(8a/88)截斷。所以 255-wrap 不在成長公式內。

## Bug 4:勇者 MaxMP 成長偏低(已修)

成長公式 `sub_d9cc`(file 0xed3c)對 8 個職業共用一段碼,職業 = `[si+1]`(0 勇者…7 遊玩者),
查 `bx = 職業×14` 的成長表列。每列 14 byte,MP 對應 `+2 base`、`+3 slope`:

```
f00edb0: mov al,[bx+0x4369]   ; MP slope
         mul dl               ; × level
         shr ax,1             ; ÷2
         add al,[bx+0x4368]   ; + MP base   (target = base + slope×level/2)
         adc ah,0
         sub ax,[si+0x1e]     ; delta = target − 現值MP
         ... delta>0 → call sub_fa57(delta) ...
```

關鍵:`sub_fa57(delta)` 回傳的不是 delta 本身,而是 **`rng % delta`(下限 1)** —— 實得 MP
增量是「1 到 delta 之間的亂數」。勇者那列 MP base=3、slope=5(target 每級僅多 ~2.5),delta 小,
亂數又常落在 1,於是表現成「每級只 +1」。模擬到 Lv43 勇者 MaxMP 僅約 107,放不出費 MP 的
全體補血咒「比荷瑪拉(ベホマラー)」。

**修法(EXE in-place 資料 patch)**:把勇者列 MP base/slope 調高,delta 拉大、亂數下限提升。

| file offset | 欄位 | 原 | 新 |
|---|---|---|---|
| 0x1a4a8 | 勇者 MP base | 03 | 08 |
| 0x1a4a9 | 勇者 MP slope | 05 | 0a |

模擬(忠實重現 8-bit 加 + min-1 亂數):Lv43 勇者 MaxMP 由 ~107 提升到 ~217,且 8-bit 路徑
與真實 16-bit 結果在 lv1..43 完全一致(無 255 wrap)。屬性以 word 累加,不溢位。

**DOSBox 驗證**:`work/dq3_stat_game/`(原素材 + patched DQ3.EXE)在 dq3-dosbox 容器啟動 →
開機 → 標題「DRAGON FIGHTER III 傳說的終章 ©1993 精訊資訊有限公司」→ 進新遊戲命名畫面
(輸入姓名 / 輸入注音),全程正常,截圖與未改 baseline 逐張一致,2 byte 資料改動未造成回歸。

## Bug 5:高等級升級錯亂(未修,留 C 層)

升級門檻表:職業指標表 `DS:0x43d6` → 8 個職業各一張門檻表,**每張 44 entry(lv0..43)、
每 entry 4 byte(32-bit 累積經驗)**。8 張表在記憶體連續排列。

外迴圈 `sub_d94c`(file 0xecbc):對角色反覆呼叫 `sub_ecdb`,回傳 1(可升級)就再呼叫一次
(連升)。`sub_ecdb`(file 0xece9)以等級當索引查表:

```
mov bl,[si+0x15]   ; bl = 等級(byte)
xor bh,bh
shl bx,1
shl bx,1           ; bx = 等級×4
mov dx,[bx+di+2]   ; 門檻 高 word
mov ax,[bx+di]     ; 門檻 低 word
```

**根因 = MAX_LEVEL=43,越界讀表**。等級到 43 之後,`entry[44]` 落到**下一職業門檻表的
entry[0]**,而每職業 entry[0]=0(lv0 門檻)。於是「升級門檻 = 0」,角色經驗必然 ≥0 → 連升,
正是青衫描述「Lv4x 之後連升 30~40 級」。學咒文 `sub_db5f` 同理用爆走的等級索引滑進相鄰職業
咒文表,於是學到別職業咒文。

**修法**:查門檻 / 咒文表前 `if (level > 43) level = 43;`。

**為何未做 binary patch**:in-place 不可——原碼「載入等級→×4」僅 9 byte,插 clamp 需多 ~7 byte。
code cave 也不可——`tools/re_codecave_scan.py` 掃全檔(常駐碼段 HDR..0x12000)最大可用 cave
只有 8 byte(file 0x1169b、0x11ecf 各 8 byte 零填充;file 0xa40f 6 byte NOP;檔尾無 slack),
放不下 x86-16 的「載入+cmp+jbe+mov+回跳」clamp + trampoline(約 11~19 byte),拆兩段非相鄰
cave 又會在升級 / 戰鬥路徑放大崩潰風險。依**安全優先**原則不冒險,留待 SDL2/C 層重寫時補
clamp(該層為單行 if)。

## Bug 6:數值 255 溢位(未修,留 C 層)

青衫:除防禦力外,屬性超過 255 就從 0 算;防禦力上限約 1023。

如上「修正 docs/18」所述,成長公式 target 為 16-bit 正確、屬性欄位 word 讀寫無截斷,所以
**255-wrap 不在成長公式**。較可能來源:屬性顯示欄位寬度(以 byte / 2~3 位數渲染)、或與
Bug 5 等級爆走後讀到的垃圾值互動。靜態反組譯找不到單一可信的 byte 截斷點,需 DOSBox 把
屬性堆到 >255 實機觀察才能坐實位置。防禦力「1023」疑為 10-bit 欄位 / 顯示遮罩(全檔唯一
`and ax,0x3ff` 在 file 0x9321,上下文是道具 / 咒文清單繪製取 10-bit ID,非防禦 clamp)。

**修法(C 層)**:屬性中間值 / 顯示一律 uint16,寫回前明確 `min(v, 9999)`,不靠自然 wrap;
防禦若確為 10-bit 改 u16 + 明確上限。源於型別 / 顯示寬度,組語層無單點可同長度 patch。

## 產生與驗證

```
# 反組譯核對 + 產 patched EXE(docker capstone)
docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone \
           && python tools/re_stat_patch.py table \
           && python tools/re_stat_patch.py thresholds \
           && python tools/re_stat_patch.py build"   # -> work/DQ3_stat_fixed.EXE

# DOSBox 開機驗證
cp -r assets_raw work/dq3_stat_game && cp work/DQ3_stat_fixed.EXE work/dq3_stat_game/DQ3.EXE
docker run --rm -v "$PWD":/work -v "$PWD/work/dq3_stat_game":/game dq3-dosbox \
  bash /work/tools/dosbox_run.sh statfix    # -> dosbox/statfix_*.png
```

## 殘留 / 待釐清

- #4 的數值選擇(base 8 / slope 10)為求穩健的保守值;真要對齊 FC DQ3 勇者 MP 曲線,
  可日後實機比對微調(仍 in-place 2 byte)。
- #5 / #6 的 clamp 在 SDL2/C 重寫時各為單行,屆時一併補入(level clamp + uint16 屬性)。
- #6 的實際 255-wrap 觸發點需 DOSBox 把屬性堆過 255 觀察後再定論。
