# 筆記:DPMI 在精訊 DQ3 中的用途

**結論:DQ3 完全不使用 DPMI。** 整個遊戲跑在 **16-bit real mode**(真實模式),沒有任何保護模式、
DPMI、EMS/XMS。下面記錄這個判斷的由來與證據,以免日後再被 `MZP` 標頭誤導。

## 為何一度以為用了 DPMI

`DQ3.EXE` 開頭 4 個位元組是 `4D 5A 50 00`,讀起來像 **"MZP"**。坊間有個常見啟發式:
把 `MZ` 後面接 `P`(0x50)當成 **Borland Pascal 7.0 保護模式(DPMI)執行檔**的標記。
本專案早期即依此把目標訂為「Borland Pascal + DPMI 還原」。

**這是誤判。** MZ EXE 標頭的第 3 個位元組(offset 0x02)是 `e_cblp`(最後一頁使用的位元組數),
這裡的值剛好是 `0x50` = 80。`DQ3.EXE` 大小 115,282 = 225×512 + 80,`e_cblp=80` 完全合理。
那個 `P` 只是 80 的低位元組,**不是**保護模式標記。詳見 [`docs/05-exe-recon.md`](05-exe-recon.md)。

## 證據:沒有 DPMI

對整個 `DQ3.EXE` 反組譯後統計中斷呼叫:

| 中斷 | 次數 | 用途 |
|---|---|---|
| `int 21h` | 101 | DOS(檔案 I/O、結束)|
| `int 10h` | 14 | BIOS 視訊(模式設定)|
| `int 33h` | 13 | 滑鼠 |
| `int 31h` | **0** | DPMI 服務 API ✗ |
| `int 2Fh` | **0** | DPMI host 偵測(AX=1687h)✗ |
| `int 67h` | **0** | EMS 擴充記憶體 ✗ |

- **無 `int 31h`** → 從不呼叫任何 DPMI 服務。
- **無 `int 2Fh` / AX=1687h** → 從不偵測或進入 DPMI 環境。
- **無 `int 67h`** → 不使用 EMS。
- 程式碼為 large-model real-mode:`lcall seg:off` 遠呼叫、直接 `int 21h`、直接硬體埠 I/O
  (VGA 0x3c4、Sound Blaster DMA、8042 鍵盤),這些在保護模式下會觸發例外,只能在 real mode 直接做。

## 那麼沒有 DPMI / EMS,DQ3 怎麼管大素材?

最大素材 `PACKBG.SCR` 有 3.7MB,遠超 real-mode 的 640KB 慣常記憶體。DQ3 的作法是
**從磁碟串流、用到才載入**,不是一次全部塞進記憶體:

- 素材以 DOS `AH=3Dh 開檔 → AH=3Fh 讀 → AH=3Eh 關檔` 逐檔載入到固定的 real-mode 段緩衝
  (見 [`docs/06-exe-callmap.md`](06-exe-callmap.md) 的 BLK 載入函式:讀 6-byte 表頭後把 tile
  資料讀進 `DS=[0x2532]` 指的遠段)。
- 帶序號的素材(`cty00`、`dq3?.blk`、`blkbm`)以檔名模板就地改號,按需要載入對應場景片段。
- 即每個畫面/場景只把當下需要的 tile/背景/文字載入有限的慣常記憶體,用完換下一批。

## 對 RE → C 的意涵

- 還原目標是 **real-mode large-model C**(可用當年 DOS C 編譯器重建),**不需要**任何 DPMI/保護模式
  支援,也不需要 DOS extender。
- 驗證在 DOSBox 的純 real-mode 即可,毋須 DPMI host。

> 一句話:DQ3 是純 real-mode DOS 遊戲,DPMI 在其中的用途是「無」。先前的 DPMI/Pascal 假設源自
> 對 `MZP` 標頭的誤讀,已更正。
