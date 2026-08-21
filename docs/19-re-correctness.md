# 我們怎麼知道反組譯是「對的」?——RE 正確性的確認

> **2026-08-22 勘誤：**逐函式 byte-match 是機器碼保真與反編譯品質的強證據，
> 但即使全 EXE byte-match，也不能數學上證明每個資料欄位與玩家可見語意均已理解。
> 本篇的全覆蓋目標從未完成；現況以 [`docs/135`](135-re-assertion-audit-20260822.md)
> 為準。

反組譯最容易自我感覺良好:讀完組語、寫出一份「看起來等價」的 C,就以為懂了。但**「看起來對」不等於「證明對」**。這份文件記錄本專案如何把 RE 正確性從「判讀」推進到「證實」,以及一路上對編譯器的三次修正。

## 兩個層級的確認

| 層級 | 對象 | 方法 | 狀態 |
|---|---|---|---|
| L1 素材 | 字型 / 圖 / 地圖 / 文字 / 調色盤 | 用解出的格式重新 render,與 **DOSBox 跑原版**的實機畫面逐張對照 | ✅ 已確認 |
| L2 程式碼 | 反組譯出的邏輯 | **byte-matching 反編譯**:把我們的 C 用原版的編譯器編回去,逐函式對上原版機器碼 | 🔄 方法已證、擴展中 |

**L1 已確認**:標題(`TITG.P`)、開場(`TITA.P`)、主選單、城鎮 tile+palette、注音輸入流程,都與 DOSBox 實機像素/行為吻合(見 [`docs/14`](14-dosbox-validation.md))。

**L2 是關鍵**:程式邏輯(`re/*.c`)是我們對組語的判讀,要證明它正確,最嚴謹的方法是 **byte-matching**——若我們的 C 經正確編譯器編出來，bytes 與原版那段**完全相同**，就證明判讀無誤。這正是現代 decomp 專案(SM64 / Zelda 等)的黃金標準。要做這件事,得先知道**原版用哪顆編譯器**。

## 編譯器偵探:三次修正

| 階段 | 判斷 | 為何錯 / 證據 |
|---|---|---|
| ① 最初 | Borland **Pascal** 7.0 DPMI 保護模式 | 只看檔頭 `MZP`。**錯**:`P` 只是 `e_cblp`=80,非保護模式標記(見 [`docs/07`](07-dpmi-note.md)) |
| ② recon | 16-bit real-mode、**large model** C | 滿是 `lcall` 遠呼叫 → 以為 large model |
| ③ 定案 | **Microsoft C 5.x，near-code model** | 逐函式指紋(見下),修正 ②;Borland 排除 |

### 鎖定 MSC 5.x 的指紋證據

- **near `ret`(0xC3)佔 268/280 函式**、far ret 僅 1 → 程式碼是 near(small/medium),不是 large。
- **`LOOP` 指令 184 次** + `rol`/`ror` 旋轉 + `cdq` + **459 個 NOP padding** + AX 為中心的定址 → 典型 **Microsoft C**;Borland Turbo C 幾乎不發 `LOOP`、不用旋轉、預設帶 BP frame。
- OMF 物件層:MSC 有 **CONST 段**與 `MS C` 浮水印;Turbo C 無 CONST、浮水印為 `TC86 Borland Turbo C 2.01`。
- runtime 內含 int 24h(嚴重錯誤)handler 字串「Press X to stop trying…」+ `disk is write-protected$…` → MSC `_hardresume` 系措辭。

> 教訓:**別只看檔頭/單一特徵就定編譯器**。要逐函式指紋(ret 類型、特殊指令、padding、OMF 段、runtime 字串)交叉比對。

## byte-match PoC:方法論已通

- 標的:亂數產生器 `sub_e6b9`(LCG)。
- 結果:**MSC 5.1 `/AS /Ox` 編出來與原版逐 byte 對上**——`mov ax,[seed]; add ax,0x9018`(原版 `a1 5a0b 05 1890`)同序、NOP padding、CONST 段、near ret 全對。殘差僅旋轉展開(原版 3×`rol` vs 編譯器 `shl`)與省略的 BP frame,屬點版本/flag 微調。
- 工具:`tools/re_match.py`(OMF LEDATA 抽碼 + capstone + diff);docker 內 MSC 5.1 編譯可重現(見 [`docs/17`](17-build-toolchain.md))。

**意義**:「docker 編 C → 反組譯 → 逐 byte 比對原版」這個迴圈跑通了一個函式,代表整條路可行。

## byte-match 能證明什麼

- **不是**「跑得起來、看起來像」。
- **較強的程式碼證據**:把反編譯的 C 用 MSC 5.x 編回去，讓目標函式 byte-identical。
  這證明該編譯設定下的機器碼重現，仍須另以 caller、writer-consumer、原始資料與
  玩家可見 oracle 證明所宣稱的語意；不能把 byte-match 當成全域語意證明。
- 在此之前,程式碼層的 RE 是「高信心判讀」,**尚未證實**——這份文件誠實標示這個邊界。

## 現況與下一步

- ✅ L1 素材:DOSBox 實機確認。
- ✅ L2 方法:編譯器鎖定(MSC 5.x)+ 1 函式 byte-match。
- ⏸ 歷史 L2 規劃:曾規劃把 byte-match 推到全部 280 函式；現行 Go／Ebitengine
  remake 不以此作交付 gate，只對玩家 blocker 閉合必要資料流。

> 現行解讀：byte-matching 用來證明指定機器碼的重現；原版語意與玩家結果仍要各自
> 閉合，不能由整檔 bytes 相同代替。
