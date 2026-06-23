# 逆向方法論:本專案重複用到的可重用技巧

整檔反組譯 `DQ3.EXE`(16-bit real-mode、Microsoft C 5.x)與解碼各素材的過程中,有幾個技巧是
**反覆奏效、跨子系統重用**的。這裡把它們抽離出來,附上本專案的實證,方便日後逆向同類老遊戲
(DOS/16-bit、固定格式素材)時直接套用。其餘逐子系統的細節在 `docs/01`–`docs/35`。

## 1. 用已知正確的資料當「羅塞塔範本」鎖定格式

破解未知二進位格式時,先找**含已知內容、最容易判讀的那一份**當基準,肉眼確認排列慣例
(尺寸、row/col-major、MSB/LSB),再把這份已知真值套到其他同類檔。

- **字模實證**:`D3TXT00.FON` 含拉丁字母與數字 `0-9 A-Z`,在 16×16 row-major MSB 一次解出,
  排列慣例就此定案;再把它每個 32-byte 字模當範本,在別的字型檔逐 byte 搜尋完全相符片段,
  得到「地面真值」對齊位置,勝過任何啟發式。
- **相位鎖定**:點陣字常在底部留 1–2 列空白當行距(DQ3 字身 16×14,row14/15 實測 97.9%/99.2%
  全空)。抽取時強制 `row15==0 且 row14 留白`,半格錯位的視窗必在這兩列見筆畫而被否決,
  相位自動鎖定。變寬/變位字型(`CHINA.FON`)改用「貪婪非重疊掃描」:合格字模就輸出並 +字模長,
  否則 +1 重新對齊,自動跳過索引表與插入位元組。

## 2. 找不到呼叫者時:判定 fall-through 或跳表間接派發

某 handler 全檔掃不到呼叫者,**不代表是死碼**。標準排查順序(DQ3 `0x77ce` 彩虹水滴合成實證):

1. 全檔掃直接呼叫:near `E8`(target = `i+3+rel`)、far `9A off seg`(phys = `seg*16+off`)。
2. 全檔掃指標表:`word == handler` 的 image-phys、或 `dword (off,seg)` far 指標。
3. 三者皆 0 → 多半是 **(a) fall-through 尾段**(往前找前一個 `ret` 定出真正入口),或
   **(b) 跳表間接派發**。
4. 找跳表:把 handler 入口 image-phys 當 word 全檔搜 → 命中處是 handler-offset 表;
   看周邊是否單調遞增的 offset 群,定出表 base。
5. 找派發器:掃 `FF /2`(call)/`/4`(jmp)且 disp16 在表 base 附近的 indexed call,
   典型樣式 `mov bx,[id]; dec bx; shl bx,1; call word[bx+TABLE]`。
6. 反推 event id:`index = (handler_off − table_base)/2`,再依 `dec bx` 等還原原始 id。
   id 常**資料驅動**(來自地圖/劇本資料,經 `[A]→[B]` 鏈載入,全檔掃不到 `mov [id],imm`)。

實例:`0x77ce` → handler `file 0x776c` → 跳表 `DS 0x3baa` idx82 → runner `file 0xabb2`
(`bx=[0x722]`)→ scripted-event **id 83**。

**何時收手轉動態**:runner/handler 群 + 跳表**全無可靜態解析的進入點**(near/far call、jmp、
指標 data 全 0,且掃描器已用已知 far call 反查驗證無誤)→ 代表經執行期計算指標 / 事件腳本 VM
進入。別再硬追,改 DOSBox debugger 在 runner phys 下中斷點,觸發後讀地圖 id / 座標 / 呼叫堆疊。

## 3. 位址基準混用陷阱:工具輸出 ≠ 符號表位址

最容易吃悶虧的一類錯:**反組譯工具印的位址與符號表/xref 用的位址是不同基準**,混用不報錯,
只會把你帶進「另一個剛好也存在、語意看似合理」的函式,結論整條歪掉。

- **本專案換算**:`tools/re_disasm.py` 印 **file offset**;`exe_funcs.json` / `sub_XXXX` 用
  **logical(seg0-relative)**,`file = logical + 0x1370`;DGROUP `file = 0x16140 + DS_off`。
- **踩過的雷**:examine_chain 內 `call (file)0x9cd6`(= logical 0x8966)被誤當 `sub_9cd6`
  (logical 0x9cd6 = file 0xb046),一路追進**戰鬥碼**,把 `[0x37c4]` 誤判成野外 NPC 對話表
  (實為怪物/戰鬥事件表)。
- **防呆**:開工就把換算常數釘死、每次跨「工具輸出 ↔ 符號名」顯式標 `(file)`/`(logical)`;
  用一個滿地都是的 far call(DQ3:`lcall 0x10b6,0x166` 共 56 個)全檔反查,數量/落點對得上
  才信掃描器與換算。結論可疑時,先懷疑追錯基準,再懷疑邏輯。同類問題也見於 PE 的
  file offset vs RVA、ELF 的 file offset vs `p_vaddr`。

## 4. 用資料內容反證程式碼標註

替函式下標註時不要只靠 call-graph 與直覺,**把那段碼實際讀寫的資料/字串解出來,用內容反證**。
內容對不上 = 標註錯,不管 call-graph 看起來多合理。這是技巧 1(羅塞塔範本)的推廣:
已知真值的資料是最強的地面真值。

- **實證**:一條被標成「野外 examine/話す」的路徑(confirm_worker / `sub_ba71`),把它顯示的
  D3TXT 記錄解出來,全是 `{V}受到{V}點的傷害。`/`{V}被打倒了!`/`怪物們被打倒了!` ——
  全是戰鬥訊息。⇒ 該路徑是戰鬥 handler、`[0x37c4]` 是怪物/戰鬥事件表,**不是**野外對話表
  (舊註誤標,照它做 NPC 對話會錯)。
- **反向證「不存在」**:訊息字串印表器掃完整個 control-code 分派表,只有文字格式 / 變數插入碼、
  沒有「run scripted-event」碼 → 確認祠堂合成事件**不經訊息系統**觸發,該往別處追。掃整張表
  反證「此路徑不負責 X」,比沿一條路徑追到底更省。
- 標註與內容衝突時:改標註,並回頭檢查是不是位址基準(技巧 3)追錯。
