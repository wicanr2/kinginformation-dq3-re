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

## 兩版 remake 的 bug(code 核實)

| | 原版 | C remake | ebitan |
|---|---|---|---|
| b2<4 NPC | **顯示**(不丟) | `dq3_scene.c:172` `if(b2<4)continue` **丟棄** | `worldmap.go:171` `if n.B2<4 continue` **丟棄** |
| cache 容量 | 13 slot | `n_npc_spr<8`(**上限 8,少 5**) | 用 Go map 無上限(不受此限) |

**影響量化**(全 CTY*.DAT NPC 表統計):**b2<4 的 NPC 共 1661 個被兩版丟棄**。熱點:
CTY01 sec5 有 24 個 b2<4、CTY03 sec4 有 21 個…。阿里阿罕(CTY00)sec0 剛好 0 個 b2<4(24 NPC 全 b2≥4),
所以起始城感覺還好;但多數城鎮少掉大批平民 NPC → 使用者「NPC 沒那麼多」的精確根因。

## 修法(C-3 批次)

- ebitan `worldmap.go`:**移除 `b2<4 continue`**;b2<4 的 sprite 改載 DQ3LIN.BLS(entry 對應待定錨:
  先驗 b2 0/1/2/3 → DQ3LIN 哪隻,用 lindump PNG × 原版城鎮截圖比對)。
- C remake 一併修:`b2<4 continue` 移除 + cache 上限 8→13(另一刀,等 ebitan 定錨後同步)。
- 驗收:阿里阿罕以外某城(如雷貝 CTY01)ebitan NPC 數對上 DOSBox 原版截圖。

## 附:sec0 出城轉場表(靜態解,供 U2 出城機制)

CTY00 sec0 transition 表(section+0xc)7 條:`{0,1,(14,6)} {0,2,(8,2)} {0,3,(7,1)} {0,4,(10,10)}
{25,0,(15,30)} {25,4,(2,4)} {0,4,(5,5)}`。**無 destSec=0xff(出地表)項** → 阿里阿罕 sec0 的出城
不走 transition 表,推測靠走到地圖邊界(待 DOSBox 實測②補;實測① agent 卡在 sec4 房間、實測② stalled)。
