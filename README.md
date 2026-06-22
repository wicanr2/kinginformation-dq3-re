# 精訊版《勇者鬥惡龍 III：傳說的終章》反組譯

中文標題「傳說的終章」對應日版 DQ3《そして伝説へ…》（精訊版改名為 *Dragon Fighter 3* 以避商標）。本 repo 記錄精訊資訊（King Information）在 1990 年代製作、**最終未發售**的中文版 DQ3 的反組譯與素材還原工作。

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

**怪物圖鑑**（`DQ3MNS.SHP` sprite + `mnsbk.pal`，128 隻全彩；id 5 = 史萊姆、id 2 = 金屬史萊姆，數值見 [`docs/data/d3mns_stats.json`](docs/data/d3mns_stats.json)）：

![怪物圖鑑](docs/monsters/monster_sheet.png)

> 城鎮地圖（`CTY*.DAT`）見 [`docs/maps/cty00_town.png`](docs/maps/cty00_town.png)；劇情純文字劇本見 [`docs/script/`](docs/script/)；更多標題 / 過場圖（片尾、年代過場）見 [`docs/title/`](docs/title/)。

## 目標

1. **反組譯主程式** — 還原 `DQ3.EXE` 的程式邏輯。經偵察，它是 **16-bit real-mode、large model 的分段 DOS 程式**（非保護模式、非 Borland Pascal；先前依檔頭 `MZP` 的判斷已更正，詳見 [`docs/05-exe-recon.md`](docs/05-exe-recon.md)），鏈結了手寫組語的低階硬體驅動。還原目標語言（C／可讀虛擬碼／其他）待定。
2. **拆解遊戲素材** — 字型、地圖、圖檔、文字腳本、音樂音效等，還原成可檢視 / 可再利用的格式。
3. **挖掘技術** — 記錄這套未發售中文版採用的中文字型、文字編碼、地圖與封包等技術細節。

## 驗證方式

還原產出須能重建並在 **DOSBox**（容器內執行，不污染 host）跑出與原版一致的行為，以原版執行畫面為黃金對照基準。

## 版權與素材

原始遊戲檔（`DQ3.EXE` 與全部素材）版權屬 Enix / 精訊資訊，**不納入本公開 repo**。要重現本專案的工作，需自行提供原版 `dq3.zip` 並解壓到 `assets_raw/`（已被 `.gitignore` 排除）。本 repo 只記錄：

- 反組譯產出的 Pascal 原始碼（`re/`）
- 格式分析與技術文件（`docs/`）
- 操作原始檔的工具腳本（`tools/`）

## 目錄結構

| 目錄 | 內容 |
|---|---|
| `re/` | 反組譯重建的 Pascal 原始碼 |
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
| `DQ3.EXE` 反組譯 → C | 🔄 主結構成 C(啟動/主迴圈/狀態機/野外指令/對話/戰鬥/場景/素材載入,~2900 行) | [`docs/05`](docs/05-exe-recon.md)–[`13`](docs/13-exe-battle.md)、[`re/`](re/) |

### 已解出的重點

- **字模格式**：三個 `.FON` 共用 16×16 row-major MSB 點陣（字身 16×14，row14/15 為行距留白）。
- **`D3TXT00.FON`**：完整解出 1,476 字 = 遊戲內全部字元（拉丁／注音／職業／武器／防具／道具／咒文／符號）。
- **`CHINA.FON`**：變位母字庫（~5,400 字），run-based 相位對齊正確率 ≈93%。早期誤判的「字模混淆」已更正為 **1-byte 對齊漂移**（`deobf@b ≡ read@(b−1)`，非防拷混淆）。
- **`D3TXT00~09.TXT`**：遊戲對話／劇情腳本。指標表 + 2-byte LE 記錄（`0xffff` 終止）；值 `<1476` 為 `D3TXT00.FON` 字模索引，`>=0xffed` 為控制碼（換行／換頁／動態插值）。逐碼 render 出通順繁體中文劇情（道具說明、各城鎮 NPC、結局、下層 DQ1 世界）。

### 待續

- glyph index → Unicode 對照表（OCR 1,476 字模）→ 把劇情 dump 成純 UTF-8。
- `CHINA.FON` 殘留 ~7% 對齊、段內索引表精確語意。
- 地圖（`DQ3*.BLK` / `CTY*.DAT`）格式與還原。
- `DQ3.EXE`（Borland Pascal DPMI）反組譯為可重編譯 Pascal。
