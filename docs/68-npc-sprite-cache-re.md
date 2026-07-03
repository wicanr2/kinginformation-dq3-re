# 68 — NPC sprite cache RE:兩版 remake「NPC 變少」的根因(2026-07-03)

> 使用者實測 U3:remake 城鎮 NPC 比原版少。靜態反組譯 `DQ3.EXE` sprite cache 建立迴圈,
> 定位為**兩版 remake 共有的移植錯誤**(非資料缺失)。方法:capstone 反組譯 + 全 CTY NPC 表統計。

## 原版機制(反組譯 file 0x4634–0x46ab,L032c4–0x3338)

NPC sprite 去重快取(DGROUP `0x265d`=key 表 / `0x265e`=狀態表),**13 個 slot**(bx 8..0x22 step 2):

```
L032a3 (0x4613): 清 13 slot key = 0xff        ; mov cx,0xd; mov bx,8; [bx+0x265d]=0xff; bx+=2; loop
L032c7 (0x4637): 掃 NPC 清單(di from 0xb66, stride 8)
  al = [di+2]                                 ; ★ b2(sprite key)直接取,無任何 <4 判斷
  在 8..0x22(13 slot)找 cache 命中
  命中 → [bx+0x265e]=1、[di+2]=bx>>1(改存 slot index)
  未命中 → [bp*2+0x260d]=di(記入待載清單)、bp++
```

**關鍵事實(反組譯確鑿)**:
1. **原版不丟任何 NPC** —— cache 迴圈對每個 b2 一視同仁,**沒有 `b2<4` 的跳過分支**。
2. cache 容量 **13 slot**(不是 8)。
3. sprite entry 計算(0xffc3):`bp==1` 走一路、否則 `ax-=4` 再 `×0xf00+6`(DQ3MAN.BLS)。
   b2<4 時 `(b2-4)` 為負 → 指向另一 sprite 源(**假說:DQ3LIN.BLS**,46086=12×0xf00+6,已 dump
   出 12 隻小人 sprite;確切 b2→源對應待實作 C-3 時再定錨,見 docs/27 DQ3LIN 待辨識項)。

## ★ C-3 定錨結果:DQ3LIN.BLS 假說已證偽,b2<4 = 原版 undefined behavior

反組譯 `sub_0xffc3`(file 0xffc3=L0ec53)+ 呼叫端(`file 0xffa4`=L0ec34,NPC 載入迴圈)確認:

```
NPC 載入迴圈(0xffa4)一律 bp=2 呼叫 sub_0xffc3(cmp bp,1 恆不成立)
  → 無條件 sub ax,4(ax=b2,u16)
  → mul 0xf00 → +6 → seek(INT21 AH=0x42)到 handle=[0x2609]=DQ3MAN.BLS(確認:file open
     `dq3man.bls` string @ 0x16210 存到 [0x2609],`dq3mst.bls` @ 0x16205 存到 [0x260b])
  → read 0xf00 bytes(INT21 AH=0x3f)到 seg_blspage 頁緩衝
```

b2∈{0,1,2,3} 時 `ax-4` 對 u16 underflow(0xFFFC..0xFFFF),`×0xf00+6` 產生 seek offset
≈0x0EFFxxxx(≈240MB),遠超 DQ3MAN.BLS 實際大小(222726B)。DOS `INT21/AH=42h` 允許 seek
超出 EOF 不報錯;其後 `INT21/AH=3Fh` 在該位置讀 0xf00 bytes 會**回傳 0 byte**(AX=0,無錯誤),
頁緩衝維持**讀取前的殘留內容**(哪個 NPC 之前用過同一 cache slot 就殘留誰的 sprite page)。

**DQ3LIN.BLS 確認不是 b2<4 的來源** —— 它由完全獨立的另一個 loader 載入:

```
file 0x13387(L12017):open "dq3lin.bls"(string @ 0x1621b)→ handle [0x2604](第三個獨立 handle,
  與 DQ3MAN=[0x2609]/DQ3MST=[0x260b] 都不同)
  迴圈 18 次,每次讀 0x1e0(480B,= 一個 sub-frame,非整格 8-frame 角色頁 0xf00)
  seek offset 來自查表 `word[byte[si]*2 + 0x2b7a] + 6`(si 指向固定 index 陣列 @ 0x9f1)
  呼叫端:file 0x133e3(L12073)`lea si,[0x9f1]; call 0x13387` —— 固定、非 b2 key 驅動
```

即 DQ3LIN.BLS 是給某個**寫死 18-subframe 清單的特定場景**用(呼叫路徑未完整追溯,推測是特定
演出/過場,非城鎮 NPC 一般渲染路徑),與 NPC `b2` key 完全無關聯。

**結論**:b2<4 在原版是**未定義行為**(讀 0 byte、沿用 stale 頁緩衝),不是「指向 DQ3LIN.BLS 某隻」。
這可能是精訊此 DOS 移植版本本身未修好的潛在 bug(該版從未發售,未經完整 QA,與專案主旨「挖掘
未發售版本技術」相符)。remake 據此改採「不丟棄、不硬猜 DQ3LIN、以 DQ3MAN.BLS entry0 當誠實
fallback」(見下)。

## 兩版 remake 的 bug(code 核實)

> **狀態(2026-07-03 更新)**:下表「C remake / ebitan(修復前)」是**發現當時**的狀態;**兩版現皆已修**
> ——ebitan commit(C-3/8b76ac1 一帶)、**C remake W5 commit f27b51d**(見底部)。表保留當時證據供追溯。

| | 原版 | C remake(修復前) | ebitan(修復前) |
|---|---|---|---|
| b2<4 NPC | **顯示**,sprite = undefined(頁緩衝殘留) | `dq3_scene.c:172` `if(b2<4)continue` **丟棄** | `worldmap.go:171` `if n.B2<4 continue` **丟棄** |
| cache 容量 | 13 slot | `n_npc_spr<8`(**上限 8,少 5**) | 用 Go map 無上限(不受此限) |

**影響量化**(全 CTY*.DAT NPC 表統計):**b2<4 的 NPC 共 1661 個被兩版丟棄**。阿里阿罕(CTY00)
sec0 剛好 0 個 b2<4(24 NPC 全 b2≥4),所以起始城感覺還好;但多數城鎮少掉大批平民 NPC →
使用者「NPC 沒那麼多」的精確根因。

> 熱點清單校正:先前這裡列的「CTY01 sec5 有 24 個、CTY03 sec4 有 21 個」來自另一支未存檔的
> 離線統計腳本,section 索引口徑與 `internal/dq3data.OpenTown`(word-table 直接索引、無 count
> 前綴)不一致——實際用 OpenTown 讀 CTY01 sec5/CTY03 sec4 會回 `layout oob`(該 word 落在
> section 表未初始化區、非有效 section)。第一輪重掃只檢查 `OpenTown` 不報錯就採信,結果誤收了
> **CTY01 sec9**(w=257、h=3079,tile 陣列需 158 萬 byte,但 `CTY01.DAT` 全檔僅 2735B——`townU16`
> 越界安全回 0,不會報錯,但整張圖幾乎全讀 0,dump 出來只有主角站在空地,是偽陽性熱點)。加上
> 「tile 陣列 `tbase+2*w*h` 必須落在檔案大小內」的嚴格檢查後,重新掃描全 CTY*.DAT 得到的有效
> 熱點改為 **CTY03 sec15(18×18、33 NPC、25 個 b2<4)、CTY90 sec13(11×11、55 NPC、26 個 b2<4)**,
> C-3 視覺核驗改用這兩筆。

## 修法(C-3 批次,已實作 ebitan)

- ebitan `worldmap.go`:**移除 `b2<4 continue`**;b2<4 一律 fallback 成 `b2=4`(DQ3MAN.BLS entry 0,
  第一個角色)—— 原版該值 undefined,不硬猜 DQ3LIN(已證偽),取「顯示但誠實非原始」優先於「丟棄」。
- **C remake ✅ 已同步(W5,commit f27b51d)**:`dq3_scene.c` b2<4→entry0 fallback(不丟棄)+ cache 上限 8→13(`npc_spr[13]`)+ 逐筆 story-flag 過濾(見 docs/71/72)。
- 驗收:CTY03 sec15 / CTY90 sec13(NPC 密度熱點,OpenTown + tile 陣列邊界雙重校正後)ebitan
  dump PNG,NPC **數量**(非 sprite 圖像本身,因原版該值本就 undefined)應對上 DOSBox 原版截圖。

## 附:sec0 出城轉場表(靜態解,供 U2 出城機制)

CTY00 sec0 transition 表(section+0xc)7 條:`{0,1,(14,6)} {0,2,(8,2)} {0,3,(7,1)} {0,4,(10,10)}
{25,0,(15,30)} {25,4,(2,4)} {0,4,(5,5)}`。**無 destSec=0xff(出地表)項** → 阿里阿罕 sec0 的出城
不走 transition 表,推測靠走到地圖邊界(待 DOSBox 實測②補;實測① agent 卡在 sec4 房間、實測② stalled)。

## ⚠ Open item(誠實揭露,未閉環)

C-3 修復解決的是「NPC **數量**」(移除丟棄 → npcdump 確認 b2<4 NPC 重新顯示,CTY03 sec15/CTY90 sec13)。
**但「原版這些 b2<4 NPC 實際長怎樣、是否可見」未經 DOSBox 城內實測確認** —— agent 靠反組譯推論
「載入讀 0 byte = undefined」,非實測看到。兩種可能未排除:(a) 原版顯示 stale 垃圾圖;(b) b2<4 其實
是某種特殊 sprite(如固定劇情角色),原版有我們尚未找到的專門處理。remake 現用 DQ3MAN entry-0
fallback(一律畫成第一隻角色)→ 數量對、外觀非原始。**最終驗收需 DOSBox 進城內數 NPC + 比對外觀**
(實測①②都卡在起始房間/stalled 未達成)。在那之前,此修復標「數量已對齊、外觀待實測」。
