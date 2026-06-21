# 精訊版《勇者鬥惡龍 III：傳說的終章》反組譯

中文標題「傳說的終章」對應日版 DQ3《そして伝説へ…》。本 repo 記錄精訊資訊（King Information）在 1990 年代製作、**最終未發售**的中文版 DQ3 的反組譯與素材還原工作。

## 目標

1. **還原 Pascal 原始碼** — 主程式 `DQ3.EXE` 為 Borland Pascal 7.0 DPMI 保護模式執行檔（檔頭 `MZP`），原始語言即 Pascal。目標是把它反組譯回可重新編譯的 Pascal 原始碼。
2. **拆解遊戲素材** — 字型、地圖、圖檔、文字腳本、音樂音效等，還原成可檢視 / 可再利用的格式。
3. **挖掘技術** — 記錄這套未發售中文版採用的中文字型、文字編碼、地圖與封包等技術細節。

## 驗證方式

反組譯產出的 Pascal 須能重新編譯，並在 **DOSBox**（容器內執行，不污染 host）跑出與原版一致的行為，以原版執行畫面為黃金對照基準。

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
| 文字腳本解碼 | ✅ | [`docs/03-text-format.md`](docs/03-text-format.md) |
| 地圖（BLK / CTY）還原 | ⬜ | — |
| `DQ3.EXE` → Pascal 反組譯 | ⬜ | — |

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
