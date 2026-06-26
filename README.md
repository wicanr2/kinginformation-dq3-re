# 精訊版《勇者鬥惡龍 III：傳說的終章》反組譯

中文標題「傳說的終章」對應日版 DQ3《そして伝説へ…》（精訊版改名為 *Dragon Fighter 3* 以避商標）。本 repo 記錄精訊資訊（King Information）在 1990 年代製作、**最終未發售**的中文版 DQ3 的反組譯與素材還原工作。

> 📖 知識庫入口:[`CONTEXT.md`](CONTEXT.md) — canonical 術語表 + `docs/` 全文件主題索引;逆向方法論見 [`docs/00`](docs/00-re-methodology.md)。

## 成果展示

從原版資料檔反組譯、解碼、還原出的成果（皆由 `tools/` 內的工具從原始素材重新產生）。標題、開場、主選單、城鎮 tile / palette 已在 **DOSBox 實機跑原版逐張比對驗證**（見 [`docs/14`](docs/14-dosbox-validation.md)），與還原結果一致：

**標題畫面**（`TITG.P` → 解出為標準 ZSoft PCX：640×350、4bpp、RLE、palette 內含於檔）：

![標題畫面](docs/title/title_logo.png)

**開場畫面**（`FIRST.SCR` → 640×350 4bpp row-interleaved planar 解碼）：

![開場原畫](docs/title/first_opening.png)

**世界地圖**（`DQ3CON.MAP` 244×205 + `DQ3.BLK` tile 圖庫 + `DQ3.PAL`）：

![世界地圖](docs/maps/world_con.png)

**遊戲內中文字庫**（`D3TXT00.FON` 完整 1,476 字：拉丁／注音／職業／武器／防具／道具／咒文）：

![字庫](docs/fonts/D3TXT00.FON.atlas.png)

**怪物圖鑑**（`DQ3MNS.SHP` sprite + `mnsbk.pal`，130 隻全彩；id 5 = 史萊姆、id 2 = 金屬史萊姆，數值見 [`docs/data/d3mns_stats.json`](docs/data/d3mns_stats.json)。最後兩格 128/129 為下方復原的未完成 boss）：

![怪物圖鑑](docs/monsters/monster_sheet.png)

**復原未完成的 sprite**（未發售版 130 隻中 **2 隻沒做完**:id 128 歐里狄加、id 129 五頭龍大王——正是索瑪城「父親 vs 五頭龍」決鬥那場戲的雙方,缺圖導致該戰當機。依參考圖以精訊風格重繪、注入原格式,經遊戲 SHP 解碼器驗證;五頭龍配色對齊實機索瑪最終戰的綠色 King Hydra。remake 戰鬥已接此復原 sprite,bug #3 從「不當機」進到「真顯示」):

![復原父子決鬥](docs/monsters/restored_128_129.png)

> 城鎮地圖（`CTY*.DAT`）見 [`docs/maps/cty00_town.png`](docs/maps/cty00_town.png)；劇情純文字劇本見 [`docs/script/`](docs/script/)；更多標題 / 過場圖（片尾、年代過場）見 [`docs/title/`](docs/title/)。全 86 個場域地圖（各 section 標號）見 [`docs/maps/cty/`](docs/maps/cty/)。

## 🥚 開發者彩蛋：阿里阿罕左上角的「LIH」

把起始城鎮**阿里阿罕**（`CTY00` section 0）的地圖渲染出來，左上角那片磚牆不是隨便砌的——它排成三個字母 **L · I · H**。三十年前某位精訊的地圖設計師，在玩家永遠不會特別留意的城牆角落，悄悄留了自己的簽名。直到把 `CTY00.DAT` 逐 tile 解碼、整鎮攤平，這個躲在牆裡的彩蛋才重見天日。

![LIH 彩蛋](docs/maps/lih_easter_egg.png)

> 全鎮地圖見 [`docs/maps/cty/CTY00.png`](docs/maps/cty/CTY00.png)，彩蛋在 section 0 左上角（tile 約 `(0–10, 0–4)`）。

## 📜 當年的 BBS：一封三十年前的存證信函

精訊版 DQ3 從未正式發售。1993 年精訊資訊做好了中文版，ENIX 不同意在台發售，開發中止；
一位離職員工把未完成版放上 BBS，從此在 FidoNet / 90Net 各站之間流傳。一整代玩家在
14 吋螢幕前玩它、卡關、罵它一堆 bug，然後彼此交換用 PCTOOLS 摸出來的修改法。

我們找到一份當年（1994–1995）的 BBS 討論串存檔，把它整理成了
[**歷史紀錄**](docs/history/dq3-bbs-1994.md)。裡面有完整的故事攻略、咒文道具表、存檔修改大全——
還有精訊資訊董事長**李培民**親自貼出的那份公告，要求各站「切勿再任意拷貝、陳列以及散佈」。

最有意思的是：當年玩家用 hex editor 一個 byte 一個 byte 試出來的 bug 修改法，
和我們今天從 `DQ3.EXE` 反組譯出的**是同一批 bug**——魔王打不死、彩虹水滴合成錯、五頭龍當機，
三十年前的 `Find/Edit` patch 正好對得上我們的 docs/18–22。那份討論串，成了這個專案最意外的
一份外部交叉驗證。

> 他們說沒有中文版，要大家別再傳了。三十年後，我們把它一行一行讀回來。

## RE 正確性驗證 + 修正版

**反組譯正確性已證實(兩個層級):**
- **素材**:標題 / 開場 / 主選單 / 城鎮 tile / palette 在 [DOSBox 跑原版逐張比對](docs/14-dosbox-validation.md)一致。
- **程式碼**:整檔反組譯 → nasm 重組,`sha256` 與原版**逐 byte 100% 相同**([docs/17](docs/17-build-toolchain.md))。
- 原版編譯器經逐函式指紋鎖定為 **Microsoft C 5.x**([docs/19](docs/19-re-correctness.md));byte-match PoC 已通。

**精訊 DQ3 修正版(binary patch 對照組,7 個 bug 處理 5 個)** —— 證明 RE 能精準修原版([docs/18](docs/18-bug-analysis.md)、[docs/20](docs/20-fixed-version.md)):

| Bug(青衫先生記錄) | 修法 |
|---|---|
| 巴拉摩斯打不死 / 彩虹水滴卡關 / 勇者 MP 固定+1 / 隼劍只打一次 | ✅ EXE binary patch(共 25 byte,同長度不位移)|
| 九頭龍/歐里狄加戰當機 | ✅ **重繪未完成的 sprite 補圖**(父子決鬥雙方)|
| 祈禱之戒 | ✅ 反組譯確認本就生效,不需修 |
| 高等級升級錯亂 / 數值 255 溢位 / 魔甲無抗魔 | ⬜ 留「C 重編」修(real-mode 無 cave 空間)|

> 兩階段策略:**(a) binary patch 對照組(已完成)→ (b) 從 re 出的 C source 重編** —— bug 修進 C(無 real-mode 限制)、用 MSC 5.1 重編出 `RE-DQ3.EXE`,並以此乾淨 C 為基礎做 SDL2 現代移植。

## 目標

1. **反組譯主程式 → 用 SDL2 重寫成現代版** — 還原 `DQ3.EXE` 的程式邏輯。經偵察與逐函式指紋，它是 **16-bit real-mode、near-code model 的 DOS C 程式，由 Microsoft C 5.x 編譯**（非保護模式、非 Pascal、非 large model；先前依檔頭 `MZP` 與遠呼叫的判斷已逐步更正，詳見 [`docs/05`](docs/05-exe-recon.md)、[`docs/07`](docs/07-dpmi-note.md)、[`docs/19`](docs/19-re-correctness.md)），鏈結手寫組語的硬體驅動。最終目標:把它**用現代 C + SDL2 跨平台重寫**,以 DOSBox 原版為 oracle 驗證一致。
2. **拆解遊戲素材** — 字型、地圖、圖檔、文字腳本、音樂音效等，還原成可檢視 / 可再利用的格式。
3. **挖掘技術** — 記錄這套未發售中文版採用的中文字型、文字編碼、地圖與封包等技術細節。

## 驗證方式

還原產出須能重建並在 **DOSBox**（容器內執行，不污染 host）跑出與原版一致的行為，以原版執行畫面為黃金對照基準。

## 版權與素材

原始遊戲檔（`DQ3.EXE` 與全部素材）版權屬 Enix / 精訊資訊，**不納入本公開 repo**。要重現本專案的工作，需自行提供原版 `dq3.zip` 並解壓到 `assets_raw/`（已被 `.gitignore` 排除）。本 repo 只記錄：

- 反組譯產出的 C 原始碼 + SDL2 重寫（`re/`）
- 格式分析與技術文件（`docs/`）
- 操作原始檔的工具腳本（`tools/`）

## 目錄結構

| 目錄 | 內容 |
|---|---|
| `re/` | 反組譯重建的 C 原始碼 + SDL2 移植（`re/sdl/`）|
| `docs/` | 格式分析、結構地圖、技術筆記 |
| `tools/` | 解析 / 抽取 / 渲染素材的腳本（Python，於 docker uv venv 執行） |
| `references/` | 外部參考資料（青衫先生攻略等） |
| `assets_raw/` | （git 排除）原版素材，使用者自備 |
| `assets_out/` | （git 排除）抽出的素材產物 |

## 進度

逐迭代推進，每輪有產出即更新本 repo。

| 項目 | 狀態 | 文件 |
|---|---|---|
| 資產 inventory 與格式偵察（194 檔） | ✅ | [`docs/01-asset-inventory.md`](docs/01-asset-inventory.md) |
| 字型解碼 | ✅ | [`docs/02-font-format.md`](docs/02-font-format.md) |
| 文字腳本解碼 + 純文字（UTF-8）dump | ✅ | [`docs/03-text-format.md`](docs/03-text-format.md)、[`docs/script/`](docs/script/) |
| 地圖 tile + 世界地圖還原（城鎮佈局待 EXE） | ✅ | [`docs/04-map-format.md`](docs/04-map-format.md)、[`docs/maps/`](docs/maps/) |
| `DQ3.EXE` 反組譯 → C | 🔄 主結構成 C(啟動/主迴圈/狀態機/野外指令/對話/戰鬥/場景/素材載入/注音輸入) | [`docs/05`](docs/05-exe-recon.md)–[`19`](docs/19-re-correctness.md)、[`re/`](re/) |
| RE 正確性確認(原版編譯器 = **MSC 5.x**,byte-match 驗證) | 🔄 方法已證、擴展中 | [`docs/19`](docs/19-re-correctness.md)、[`docs/17`](docs/17-build-toolchain.md) |
| **SDL2 現代跨平台重寫**(`dq3_remake/`,C99+SDL2) | ✅ **完整可玩 + 主線可破關 + 79/79 驗證 + 打包**(見下「remake 可玩狀態」)| [`dq3_remake/`](dq3_remake/)、[`dq3_remake/WORKLIST.md`](dq3_remake/WORKLIST.md) |

### remake 可玩狀態（`dq3_remake/` — 核心 RPG 迴圈完整）

`dq3_remake/`（C99 + SDL2，docker 編譯、不污染 host）已是一個**完整、可玩、資料驅動**的精訊版
DQ3 核心。整條經典迴圈都在，且每個數值 / 邏輯都從 `DQ3.EXE` 或遊戲資料抽出，不是湊的：

```
阿里阿罕城鎮起步 → 露依達酒場創角（職業 → 注音/英數命名 → 性別 → 登錄）→ 名冊/隊伍
  → 出城到地表（真實 DQ3CON.MAP + 原版起點 153,174 + 真實遭遇區）
  → 城鎮（CTY / NPC / 繁中對話 / 門 / 轉場）
  → 戰鬥：野外指令窗 → 手動/自動選咒（EXE 精確威力）+ 怪物 AI（施咒率/逃跑/真實法術）
          + 裝備加成 + 升級（255 clamp）+ 回寫名冊 + 戰利金錢
  → 商店（真實 ITEM.DAT 價格 / 職業裝備限制）→ 裝備變強
  → つよさ / じゅもん / 道具 畫面 → 存檔
```

**7 個 bug 全在 C 層修對**。怪物全 RE（[AI](docs/37-monster-ai.md) / [數值+法術](docs/38-monster-stats.md) /
[遭遇區](docs/39-encounter-zones.md)）、咒文全 RE（descriptor 威力 + 習得表）、道具全 RE
（[ITEM.DAT](docs/22-item-fix.md) 攻/防/價/職業限制）。中文化亮點**注音組字輸入**已重建。

**主線可破關**:8 里程碑真實 NPC 觸發（盜賊鑰匙 / 魔法玉 / 羅馬利亞皇冠 / 達瑪轉職 / 取船 /
彩虹合成 / 下降 / 索瑪終戰）一條龍推進到 9/9；索瑪二階段（光之珠）+ 結局捲動。先前判定
「需 DOSBox debugger」的 `[0x722]` runner/region scripted-event 觸發系統**已靜態攻克**
（[`docs/re-log-722-state-machine.md`](docs/re-log-722-state-machine.md)，57 個 setter 推翻「無 setter」結論）。

**杜勝利攻略逐章接線**（B 道具鏈 + A boss + 過場，攻略當 ground truth）：夢幻紅寶石 / 變身杖 / 蓋亞之劍 /
鑰匙鏈 / **彩虹水滴鏈 100% 可玩** / 領悟之書賢者轉職 / 王者之劍（瑪依拉換）/ 甘達特金皇冠正源 /
**甘達特盜賊巢穴**（CTY14 走過去按空白 → 甘達特手下怪27 → 甘達特怪26，與古布達黑胡椒鏈閉環）/
**巴拉摩斯本體 + 索瑪前序列**（怨靈122 → 殭屍123 → 索瑪124）/ 不死鳥拉米亞坐騎。boss/事件「正式觸發點」
用 `kind=special` 事件格定位（[`docs/data/boss-trigger-points.md`](docs/data/boss-trigger-points.md)、
`special-events-audit.md`，`kind=special` 事件全分類、**0 個 review 待辦**，可玩性無缺口）。

**離開 / 存讀檔**：F10 = 離開遊戲（「離開遊戲嗎」Yes/No 確認 + 自動存檔），ESC = 取消當前選單（非離開）；
自動存檔 `dq3_save.dat`（名冊 / 隊伍 / 道具 / 位置），DQ3_LOAD 讀檔續玩，game_tester 含存讀檔 roundtrip 斷言。
**交付驗證**:`tools/game_tester.sh` **79/79 全綠**（15 單元 + 主線一條龍 + 全 89 城零崩潰 +
boss 事件 + B 道具鏈 + 巴拉摩斯序列 + 甘達特巢穴 + 草薙之劍 + 存讀檔多 slot roundtrip）+ DOSBox oracle（標題逐像素一致、注音 IME 字序一致，
[`docs/data/oracle-validation.md`](docs/data/oracle-validation.md)）;`tools/package.sh` 打包跨平台
distributable（[`dq3_remake/DIST_README.md`](dq3_remake/DIST_README.md),不含版權素材）。

### 已解出的重點

- **字模格式**：三個 `.FON` 共用 16×16 row-major MSB 點陣（字身 16×14，row14/15 為行距留白）。
- **`D3TXT00.FON`**：完整解出 1,476 字 = 遊戲內全部字元（拉丁／注音／職業／武器／防具／道具／咒文／符號）。
- **`CHINA.FON`**：變位母字庫（~5,400 字），run-based 相位對齊正確率 ≈93%。早期誤判的「字模混淆」已更正為 **1-byte 對齊漂移**（`deobf@b ≡ read@(b−1)`，非防拷混淆）。
- **`D3TXT00~09.TXT`**：遊戲對話／劇情腳本。指標表 + 2-byte LE 記錄（`0xffff` 終止）；值 `<1476` 為 `D3TXT00.FON` 字模索引，`>=0xffed` 為控制碼（換行／換頁／動態插值）。逐碼 render 出通順繁體中文劇情（道具說明、各城鎮 NPC、結局、下層 DQ1 世界）。

### 待續

已完成:地圖（`DQ3*.BLK` / `CTY*.DAT`）格式與還原 ✅、`DQ3.EXE`（MSC 5.x）反組譯 + byte-match
確認 ✅、SDL2 跨平台重寫（完整可玩 + 可破關 + 79/79 驗證 + 打包）✅、`[0x722]` runner/region
scripted-event 系統靜態攻克 ✅。剩餘:

- glyph index → Unicode 對照表（OCR 1,476 字模）→ 把劇情 dump 成純 UTF-8。
- `CHINA.FON` 殘留 ~7% 對齊、段內索引表精確語意。
- remake 內容完整度延伸（更多已解出的 sub2 handler 接線、未發售版未接好的劇情補洞);
  細項見 [`dq3_remake/WORKLIST.md`](dq3_remake/WORKLIST.md)、[`docs/47-remaining-work.md`](docs/47-remaining-work.md)。
