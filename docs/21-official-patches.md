# 青衫攻略官方 EXE patch 全表(8 段)— 反組譯交叉驗證

青衫整理的攻略(`references/勇者鬥惡龍3:傳說的終章.html`,本機,git 排除)藏有「修改:DQ3.EXE」
一節,列出當年玩家社群實際用來通關的 **8 段 EXE patch**,外加存檔 HEX 修改法與 128 道具編號表。
本文件把這 8 段 patch 全部抽出來,逐 byte 在 `assets_raw/DQ3.EXE` 反組譯命中、驗證原 bytes,
並判斷哪些是「修 bug」、哪些是金手指/相容性調整,以及哪些補進修正版。

抽取與驗證工具:`tools/re_official_patches.py`(docker 內跑)。機讀結果:`docs/data/official_patches.json`。

## 結論先講

- **8 段 patch 的原始 bytes 全部在 EXE 內精準命中、逐 byte 相符**(唯一命中,致命一擊第二塊有 2 處
  命中但屬金手指、不收)。1990 年代社群的 patch 等於替本專案的反組譯定位做了獨立交叉驗證。
- 只有 **2 段是「修官方已知卡關 bug」**:第 2 段(對映 bug#1 巴拉摩斯)、第 4 段(對映 bug#3 五頭龍當機)。
  這兩段與既有 `docs/data/bug_patches.json` 的定位 **完全一致**,本次不重複收錄(避免與既有修正版衝突),
  僅在 `official_patches.json` 標 `source="青衫官方"` 留底備查。
- **關鍵負面發現:青衫官方 patch 與存檔修改法都沒有任何一段在修 bug#4(勇者MP+1)、bug#5(高等級升級錯亂)、
  bug#6(255 溢位)。** 換言之,當年社群也沒給這三個「樂趣型」bug 任何 byte patch —— 這與 docs/18 的判定
  (#4 在 EXE DGROUP 成長表、#5/#6 需 C 層 clamp/改型別)互相印證:當年社群沒給這三個「樂趣型」bug 官方 byte patch。
  (更正:#4 後來已做成 **2-byte EXE-data patch**(`file 0x1a4a8/9`,docs/23/stat_patches.json),並非「無 EXE 單點」;
  只是當年官方未提供。)
- 其餘 6 段:2 段是真實的**當機防護但屬硬體/環境相容**(聲霸卡、特定機器),4 段是**純金手指作弊**
  (敵人消失、穿牆、致命一擊、生命不減)。都不對映青衫列的 7 個遊戲性 bug,**不補進修正版**。

## 8 段官方 patch 全表

格式說明:每段在 HTML 內是「原 bytes 行」+「mask 行」(`--` = 保留、兩碼 = 新值)。
下表「命中 file offset」為原 pattern 在 EXE 的起點(已驗證原 bytes 100% 相符)。

| # | 名稱 | 命中 file offset | 改動(file:原→新) | 類別 | 對映 bug | 補進修正版 |
|---|---|---|---|---|---|---|
| 1 | 聲霸卡防當法 | 0x11b90 | 0x11b91:e8→f0、0x11b92:80→77 | 硬體相容(Sound Blaster 偵測) | — | 否 |
| 2 | 魔王打不死修正法 | 0xbe02 | 0xbe04:36→34、0xbe0a:36→34 | **修 bug** | **#1 巴拉摩斯** | 已在(bug_patches) |
| 3 | 敵人消失法 | 0xd118 | 0xd118-1b:→90×4、0xd11c:75→eb | 金手指(作弊) | — | 否 |
| 4 | 五頭龍大王當機修正 | 0x1b02c / 0x1b038 | 0x1b02c:80→7e、0x1b02e:81→7d、0x1b038:81→7c | **修 bug** | **#3 五頭龍當機** | 已在(bug_patches) |
| 5 | 修正某些機器跑會當機 | 0x11c60 / 0x11c6a | 0x11c63:74→90、0x11c64:1d→90、0x11c6d:74→eb | 機器相容(crash 防護) | — | 否 |
| 6 | 穿牆 | 0x3b64 / 0x2e3d / 0xa95f | 0x3b69:71→00、0x2e42:6d→00、0xa962:74→eb | 金手指(作弊) | — | 否 |
| 7 | 致命一擊 | 0xc9d7 / 0xca02 | 0xc9da:39→00、0xca02..06:32 ff 8b 6d 07→8b 5d 07 8b eb | 金手指(作弊) | — | 否 |
| 8 | 生命不減 | 0xc10b | 0xc10b-0d:89 6d 16→90 90 90 | 金手指(作弊) | — | 否 |

> 致命一擊第二塊 pattern `32 FF 8B 6D 07 2B EB` 在 EXE 有 2 處命中(0xca02、0xca2c);因屬金手指、
> 不收進任何修正版,本次不再消歧義。其餘 7 段全唯一命中。

## 修 bug 兩段的反組譯對映(與既有定位一致)

### 第 2 段「魔王打不死修正法」→ bug#1 巴拉摩斯打不死

命中 `0xbe02`,正是 `sub_aa69` 的敵方累積值 clamp(docs/18 bug#1)。HTML 原碼
`39 87 36 23 73 04 8B 87 36 23`,把兩個 disp16 高位 `36`→`34`:

```
saa92: cmp [bx+0x2336], ax   ; 39 87 36 23  →  cmp [bx+0x2334], ax
saa96: jae  ok               ; 73 04
saa98: mov ax, [bx+0x2336]   ; 8b 87 36 23  →  mov ax, [bx+0x2334]
```

clamp 基準從錯誤欄 +0x2336 改回 +0x2334(累積值本身)。file 0xbe04 / 0xbe0a 原 byte 經工具驗證
皆為 `36`,與 docs/data/bug_patches.json 既收的 bug#1 兩處 patch **完全相同**。

### 第 4 段「五頭龍大王當機修正」→ bug#3 boss 缺 sprite 當機

命中 boss 遭遇表 overlay 區 file 0x1b02c。HTML 給的換怪法:第一場 `80→7E`(歐里狄加→冰河魔人)、
`81→7D`(五頭龍大王→魔文);第二場 `81→7C`(五頭龍大王→索瑪)。三處原 byte 經工具驗證為
`80 / 81 / 81`,與 bug_patches.json 既收的 bug#3 三處 patch **完全相同**。

> 注:本專案的修正版(`tools/build_fixed_version.py`)對 bug#3 採「**補回真實 sprite**」
> (`tools/make_sprites.py` 重繪 id128/id129),而非官方的「換怪 reroute」。兩種修法殊途同歸:
> 官方碼讓既有 sprite 接上 boss(快速避當,但顯示替代怪),本專案復原父子決鬥雙方原貌。
> 官方碼在此作為「定位正確」的交叉驗證,不取代本專案 SHP 修法。

## 其餘 6 段的性質判讀(為何不補進修正版)

- **第 1 段 聲霸卡防當法**:命中 file 0x11b90(音效/硬體初始化區),把 `mov si, 0x80e8` 的立即數
  改成 `0x77f0`(`BE E8 80`→`BE F0 77`)。屬 Sound Blaster 偵測常數調整以避免特定音效卡開機當機。
  非青衫列的 7 個遊戲性 bug;DOSBox 下音效卡為模擬,影響有限。
- **第 5 段 某些機器跑會當機**:命中 file 0x11c60(同初始化區),把一個 `je` 兩 byte NOP 掉、另一個
  `je`→`jmp short`(繞過/改寫兩段 `cmp ax,0` + `lcall` 的分支)。屬特定真實機器的相容性 crash 防護。
- **第 3 / 6 / 7 / 8 段 = 金手指作弊**:敵人消失(把遞減+條件跳轉 NOP)、穿牆(碰撞旗標檢查改永遠通過)、
  致命一擊(改攻擊判定)、生命不減(HP 寫回 NOP)。這些是改變遊戲規則的作弊,不是修 bug。

第 1、5 段雖是真實的 crash 防護,但對映的是硬體/環境而非遊戲邏輯 bug,且 DOSBox 環境下未必觸發;
保留在本文件全表備查,**不收進 official_patches.json 的 bug-fix 清單、不補進修正版**。

## bug#4 / #5 / #6 沒有官方修法 —— 任務目標的核實結果

本任務原始目標是「找官方 patch 修掉 bug#4/#5/#6」。逐段核對 8 個 EXE patch + 存檔 HEX 修改法
(彩虹水滴製作、五頭龍避戰存檔改 monster_id)+ 128 道具表後,**確認當年社群並未對這三個 bug 提供
任何 byte patch 或存檔修改**:

- **bug#4 勇者 MaxMP 成長+1**:無官方 patch。根因 = 成長表勇者列 MP base/slope 偏低(公式碼本身正確);
  成長表**在 EXE DGROUP**(DS:0x4366→`file 0x1a4a6`,非 dragon0.dat),已做成 **2-byte EXE-data patch**(file 0x1a4a8/9,docs/23)。
  當年官方未提供此修改,純因屬「樂趣型」非卡關 bug。
- **bug#5 高等級升級錯亂**:無官方 patch。根因是等級 byte 當索引越界查門檻/咒文表,需 clamp(code cave),
  real-mode 原碼無空檔塞同長度修正 —— 與「官方也沒給」一致。
- **bug#6 數值 255 溢位**:無官方 patch。根因是成長中間值 8-bit `mul`+`adc` wrap(C 型別),
  組語層無單點可同長度翻轉。

這是負面證據,但與 docs/18/20 的既有判定完全吻合:#4/#5/#6 屬 C 層/資料層,留待 SDL2 重寫處理。
青衫攻略本身也只能靠「存檔 HEX 改數值」事後補救,而非從 EXE 修掉根因。

## 工具與重現

```bash
# 解析 + 驗證 8 段官方 patch(印全表),不寫檔:
docker run --rm -v "$PWD":/work -w /work python:3.12-slim \
  python tools/re_official_patches.py

# 寫出 docs/data/official_patches.json(只收 2/4 兩段修 bug 的 in-place patch):
docker run --rm -v "$PWD":/work -w /work python:3.12-slim \
  python tools/re_official_patches.py write
```

- `tools/re_official_patches.py` — 8 段 patch 硬編 + 在 EXE 搜 pattern + 逐 byte 驗證原值 + dump JSON。
- `docs/data/official_patches.json` — 機讀版,只含對映 bug 的同長度 in-place patch,欄位與
  `bug_patches.json` 相容並加 `official_id` / `source="青衫官方"`。
- 完整分析(各 bug 反組譯)見 [`docs/18-bug-analysis.md`](18-bug-analysis.md);修正版組裝見
  [`docs/20-fixed-version.md`](20-fixed-version.md)。
