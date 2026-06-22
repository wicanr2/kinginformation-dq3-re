# DQ3.EXE 反組譯偵察(recon)

第一輪靜態偵察結論。**修正先前判斷**:DQ3.EXE 不是 Borland Pascal 保護模式程式。

## 是主程式

EXE 內含全部素材檔名引用(明碼):`CHINA.FON`、`d3txt00.fon/.txt`、`cty00.dat`、`item.dat`、
`player.dat`、`dragon0.dat`、`dq3con.map`、`dq3und.map`、`dq3.blk`/`dq31.blk`、`blkbm.dat`、
`dq3mst/man/lin.bls`、`dq3.pal`、`mnsbk.pal`、`d3mns.dat`、`dq3mns.shp`、`packbg.scr`、
`mbg.mcx`/`ebg.mcx`、`nvoc.vcx`/`fvoc.vcx`、`first.scr`、`TITA.P` 等。檔名以 8.3 FCB 格式存放。
→ DQ3.EXE 即遊戲主程式,負責載入並驅動所有素材。

## 架構:16-bit real-mode、large model(非保護模式)

| 項目 | 值 |
|---|---|
| 檔型 | DOS MZ EXE,115,282 bytes |
| 「MZP」 | **誤判**。byte2=`0x50` 只是 `e_cblp`=80(最後頁位元組數),非 Borland 保護模式標記 |
| header | 311 paragraphs = 0x1370 bytes;之後為載入影像 |
| relocations | 1,232 筆(typical 大型分段程式)|
| 入口 | `CS:IP = 0000:9299`(檔案 off 0x1370+0x9299=0xA609)|
| 記憶體模型 | **large model**:程式碼滿是 `lcall seg:off` 遠呼叫 |
| 模式 | **real-mode**:直接 `int 21h`(`mov ah,4Ch` 結束)、直接 I/O port;**無 DPMI、無保護模式、無 RTM** |

啟動碼(0xA609 起)是一連串對低段 runtime(seg 0xfe1/0x100a/0x1053/0x109c/0x10b6/0x1104/
0x11d3/0x1209…)的遠呼叫,以 `AH:AL` 傳參數的暫存器式 API,最後 `mov ah,4Ch; int 21h` 結束。

## 自寫低階組語 runtime library

runtime dispatcher(如 seg 0x1104)經反組譯為**直接硬體驅動**:
`in al,0x64` / `out 0x60,al`(8042 鍵盤控制器)、`cli`/`sti`、bit 運算。
→ 程式鏈結了一套**手寫組語的低階 runtime**(鍵盤,推測也含 VGA/音效直驅),典型 1990 年代
台灣遊戲開發手法。低段(~0xfe1 paragraph 起)為 runtime library,前 ~64KB(seg 0)為主程式碼。

## 編譯器/語言:未定,但幾乎可排除 Borland Pascal

- **無** Borland Pascal/Turbo Pascal 的 `Runtime error` 簽章與 banner。
- **無** 任何編譯器 banner(Microsoft/Watcom/Borland C/…)。
- 有 DOS 嚴重錯誤訊息表(`disk is write-protected$drive not ready$…`,C/Pascal runtime 皆常見)。
- 暫存器式遠呼叫 + 手寫組語驅動 → 偏向 **C(large model)+ 組語驅動**,或大量手寫組語。
- **結論:先前「Borland Pascal DPMI」判斷有誤;還原目標語言需重新討論。**

## 工具

- `tools/re_disasm.py` — capstone 16-bit 反組譯(docker 內跑;`entry` 或給檔案 offset)。

## 下一步(待定方向)

- 釐清還原目標語言(見專案根 CLAUDE.md / 與使用者確認):RE → C?可讀虛擬碼?或仍嘗試重寫為 Pascal(屬 re-implement 而非還原)?
- 切分段:列出所有遠呼叫目標 seg,分辨「主程式碼 / runtime library / 資料段」。
- 找主迴圈與素材載入函式(從引用檔名的資料位址回溯 xref)。
- runtime API 編目(鍵盤/VGA/音效/檔案 I/O),建立可命名的函式表。
- CTY 城鎮佈局、BLKATT 屬性語意、海面 palette cycling 等素材謎團,可由主程式繪製邏輯一併解開。
