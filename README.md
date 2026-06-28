# 精訊版《勇者鬥惡龍 III：傳說的終章》反組譯

中文標題「傳說的終章」對應日版 DQ3《そして伝説へ…》（精訊版改名為 *Dragon Fighter 3* 以避商標）。本 repo 記錄精訊資訊（King Information）在 1990 年代製作、**最終未發售**的中文版 DQ3 的反組譯與素材還原工作。

---

1993 年，精訊資訊把《勇者鬥惡龍 III》整套中文化做了出來——字型、劇本、地圖、選單都翻好了。然後 ENIX 不同意在台發售，開發喊停。一份半成品從某台 BBS 流出，在 14 吋螢幕前陪一整代台灣玩家卡關、罵 bug、交換用 hex editor 摸出來的修改法，然後被時間蓋過去。它從來沒有一個「完成」的版本。

這個專案做的，是把它**完成**。

我們從那支 `DQ3.EXE`（115,282 bytes、16-bit real-mode、Microsoft C 5.x 編譯）一個函式一個函式反組譯回 C——反組譯重組後 `sha256` 與原版**逐 byte 100% 相同**，證明讀懂的是真的它；再以這份乾淨的 C 為基礎，用現代 C99 + SDL2 **跨平台重寫**成一個完整可玩、**主線可破關**的版本。素材全解碼（字型 1,476 字、地圖、怪物 sprite、文字劇本），每個數值與邏輯都從原始檔抽出、以 DOSBox 原版逐張比對驗證。

接著我們拿著當年杜勝利、青衫詩客的攻略當 ground truth，把整條流程一格一格接回來：露依達酒場創角、誘惑洞窟的魔法球、金字塔的按鈕、提頓村的夜、勇氣神殿的單獨試煉、耶進貝亞地下室那三顆要推進藍白地面的大石、帶商人去草原蓋出一座新城鎮……11 項攻略缺口逐一 RE 透或自製還原（過程揪出並修掉 3 個道具碼錯位 + 1 個賢者轉職認錯道具的 bug，`game_tester` 93/93 全綠）。

我們還補回了當年沒做完的兩格怪物圖——**id 128 歐里狄加、id 129 五頭龍大王**，正是索瑪城「父親 vs 五頭龍」決鬥的雙方；缺圖讓那場戲三十年來一進去就當機，現在依精訊風格重繪、注入原格式，在純黑的最終決戰背景下，那場仗第一次真的打得起來。打完索瑪，勇者被冊封為「洛特」，手中的王者之劍化為洛特之劍——**そして伝説へ…**，圓上那個副標。

> 三十年前他們說沒有中文版，要大家別再傳了。三十年後，我們把它一行一行讀回來、一格一格補起來，讓它跑到了結局。這不是為了散布——原版版權屬 Enix／精訊，遊戲執行需搭配你**合法持有**的原版 `DQ3.EXE`——而是以工程的方式，替一個未竟的時代，補上那聲「傳說，於此展開」。

---

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
| `dq3_remake/` | **SDL2 現代重製**（C99，可玩 + 可破關）；[`WORKLIST.md`](dq3_remake/WORKLIST.md) 進度、[`DIST_README.md`](dq3_remake/DIST_README.md) 打包說明 |
| `re/` | 反組譯重建的 C 原始碼 + SDL2 移植（`re/sdl/`）|
| `docs/` | **全部技術文件**（格式分析、EXE 反組譯、事件/地圖、bug、驗證、資料/研究、RE 日誌）|
| `tools/` | 解析 / 抽取 / 渲染素材的腳本（Python，於 docker uv venv 執行） |
| `references/` | 外部參考資料；[`walkthroughs/`](references/walkthroughs/) **收錄台灣本土攻略原文**（杜勝利 / 青衫 / 孔方兄，註明出處,歷史保存）|
| `assets_raw/` `assets_out/` | （git 排除）原版素材 / 抽出產物 |

## 📚 文件索引（可點擊）

> 完整術語表 + 知識庫入口見 [`CONTEXT.md`](CONTEXT.md)。以下依主題索引重要文件。

**核心 / 進度 / 方法論**
- [`dq3_remake/WORKLIST.md`](dq3_remake/WORKLIST.md) — remake 現況 + 真正剩餘工作（唯一真相）
- 🔊 **音訊還原** — [計畫 `docs/56`](docs/56-audio-fm-plan.md) · [CMF/SB FM 反組譯經驗 `docs/57`](docs/57-cmf-audio-re.md) · [WORKLIST 音訊段](dq3_remake/WORKLIST.md)：精訊 SB FM(OPL2 CMF)音樂從 `MBG.MCX` RE 抽出（18 軌）+ 重建 OPL2 合成器在遊戲內播放 + VOC 音效（46 個）+ MT-32 真音源 OGG
- 🎼 **音樂保存 / MT-32** — [台灣電玩史保存紀錄 `docs/58`](docs/58-taiwan-jingxun-music-preservation.md) · [munt MT-32 build 經驗 `docs/59`](docs/59-munt-mt32-build.md)：24 軌離線匯出兩版本（SB OPL2 FM + 真 MT-32 ROM 經 munt render），供個人保存
- 🎵 **配樂對應 + 靜態 RE** — [`docs/61`](docs/61-music-scene-mapping.md)：哪個場景/地圖播哪首曲(含逐 CTY 城堡/迷宮/城鎮分類)；**戰鬥音樂在獨立檔 `EBG.MCX`**(非 MBG);track→曲目的靜態 RE 過程
- 📦 **打包 / 發行** — 已併入 [`dq3_remake/WORKLIST.md`](dq3_remake/WORKLIST.md) 的「發行打包」段：Linux/Windows/AppImage 基礎包已交付,完整 MT-32 包 / macOS / Android / dev-setup 為剩餘 delta(移植分析見 [`docs/55`](docs/55-android-macos-port-plan.md))
- ⭐ [**經驗教訓 Lessons Learned**](docs/lessons-learned.md) — 多輪協作踩雷 → 根因 → 紀律（別鬼打牆 / 靜態 RE 優先 / 斷言前驗證 / 零回歸改法…）
- [`docs/00`](docs/00-re-methodology.md) 逆向方法論（可重用 RE 技術）· [`docs/47`](docs/47-remaining-work.md) 剩餘工作盤點 · [`docs/46`](docs/46-debug-hook.md) DEBUG 口

**素材格式與還原** — [`01`](docs/01-asset-inventory.md) inventory · [`02`](docs/02-font-format.md) 字型 · [`03`](docs/03-text-format.md) 文字腳本 · [`04`](docs/04-map-format.md) 地圖 tile · [`27`](docs/27-bls-character-sprites.md) 角色 sprite · [`33`](docs/33-name-tables.md) 名稱表 · [`51`](docs/51-palette-and-overworld-sprites.md) palette/地表 sprite

**DQ3.EXE 反組譯（RE → C）** — [`05`](docs/05-exe-recon.md) 偵察 · [`06`](docs/06-exe-callmap.md) callmap · [`08`](docs/08-exe-functions.md) 函式圖 · [`09`](docs/09-exe-loaders.md) loaders · [`10`](docs/10-exe-states.md) 狀態機 · [`11`](docs/11-exe-gameplay.md) 場景/移動 · [`12`](docs/12-exe-commands.md) 野外指令/對話 · [`13`](docs/13-exe-battle.md) 戰鬥 · [`15`](docs/15-exe-nameinput.md) 注音輸入 · [`16`](docs/16-monsters-ui.md) 怪物/UI

**事件 / 地圖 / 腳本 / 傳送** — [`31`](docs/31-event-system.md) 事件系統 · [`32`](docs/32-world-locations.md) 世界地點 · [`34`](docs/34-cty-load-format.md) CTY 載入 · [`35`](docs/35-script-format.md) 腳本/轉場格式 · [`42`](docs/42-npc-dialogue.md) NPC 對話 · [`43`](docs/43-reachability.md) 可達性/傳送 · [`44`](docs/44-scripted-warp-re.md) scripted warp · [`45`](docs/45-binary-structures.md) 二進位結構 · [`60`](docs/60-daynight-npc-mechanism.md) **日夜 NPC 機制**(section 兩份清單,RE + 還原) · [地圖連通圖](docs/maps/map_graph.md) · [CTY↔地名](docs/maps/cty-names.md)

**戰鬥 / 怪物 / 數值 / 咒文** — [`13`](docs/13-exe-battle.md) 戰鬥 · [`37`](docs/37-monster-ai.md) 怪物 AI · [`38`](docs/38-monster-stats.md) 怪物屬性表 · [`39`](docs/39-encounter-zones.md) 遭遇區 · [`23`](docs/23-stat-fixes.md) 數值修正 · [咒文效果研究](docs/data/spell-effects-research.md) · [晝夜系統](docs/data/day-night-system.md)

**道具 / 設施 / 船 / 取得鏈** — [`22`](docs/22-item-fix.md) 道具特效修正 · [`40`](docs/40-facility-shops.md) 設施/商店 · [`41`](docs/41-treasures.md) 寶箱 · [`48`](docs/48-ship-navigation.md) 船航行 · [`49`](docs/49-item-use.md) 道具使用 · [`50`](docs/50-ship-acquisition.md) 取船 · [道具取得鏈](docs/data/quest-items.md)

**Bug 分析 / 修正 / 正確性驗證** — [`18`](docs/18-bug-analysis.md) bug 反組譯定位 · [`20`](docs/20-fixed-version.md) 修正版 · [`21`](docs/21-official-patches.md) 青衫 patch 表 · [`28`](docs/28-battle-palette-bug.md) 調色盤 bug · [`14`](docs/14-dosbox-validation.md) 素材黃金對照 · [`17`](docs/17-build-toolchain.md) byte-identical 重組 · [`19`](docs/19-re-correctness.md) MSC 5.x 確認 · [`29`](docs/29-dosbox-oracle.md)/[`30`](docs/30-phase5-comparison.md) oracle 比對 · [oracle 驗證](docs/data/oracle-validation.md) · [原版已知 bug](docs/data/original-known-bugs.md) · [bugs.md](docs/bugs.md)（青衫 bug 清單原文）

**資料 / 研究（`docs/data/`）** — [道具取得鏈](docs/data/quest-items.md) · [咒文效果研究](docs/data/spell-effects-research.md) · [晝夜系統](docs/data/day-night-system.md) · [special 事件盤點](docs/data/special-events-audit.md) · [boss 觸發點](docs/data/boss-trigger-points.md) · [sub2 結構](docs/data/sub2-struct.md)/[handlers](docs/data/sub2-handlers.md)/[live NPC](docs/data/sub2-npcs-live.md) · [攻略接線盤點](docs/data/walkthrough-flow-audit.md) · [DGROUP 表](docs/data/dgroup-tables.md)

**RE 攻關日誌 / 教訓** — [`[0x722]` runner 狀態機](docs/re-log-722-state-machine.md) · [咒文效果 dispatcher](docs/re-log-spell-effect-dispatch.md) · [remake 接線教訓](docs/remake-wiring-lessons.md)

**歷史 / 一手史料** — [BBS 1994–95 討論串](docs/history/dq3-bbs-1994.md) · [攻略致敬 + 地名對照](docs/history/dq3-walkthrough-credits.md)

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
| **SDL2 現代跨平台重寫**(`dq3_remake/`,C99+SDL2) | ✅ **完整可玩 + 主線可破關 + 93/93 驗證 + 打包**(見下「remake 可玩狀態」)| [`dq3_remake/`](dq3_remake/)、[`dq3_remake/WORKLIST.md`](dq3_remake/WORKLIST.md) |

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

**杜勝利 + 青衫攻略逐章接線**（B 道具鏈 + A boss + 過場/解謎，攻略當 ground truth）：夢幻紅寶石 / 變身杖 / 蓋亞之劍 /
鑰匙鏈 / **彩虹水滴鏈 100% 可玩** / 領悟之書賢者轉職 / 王者之劍（瑪依拉換）/ 甘達特金皇冠正源 /
**甘達特盜賊巢穴**（CTY14 走過去按空白 → 甘達特手下怪27 → 甘達特怪26，與古布達黑胡椒鏈閉環）/
**巴拉摩斯本體 + 六大魔人（怪106 ×6）+ 索瑪前序列**（怨靈122 → 殭屍123 → 索瑪124）/ 不死鳥拉米亞坐騎 /
**勇氣神殿神父挑戰**（flag 0x13 驅動 owportal）/ **諾魯特密道**（波魯多加授國王的信 → 開東方通道）/
**蓋亞那跳坑下降閘**（巴拉摩斯敗後）/ **帶商人建城**（商人 class6 寄存 → 建城）。

> **攻略全流程缺口盤點**（[`docs/data/walkthrough-flow-audit.md`](docs/data/walkthrough-flow-audit.md)）：原列 11 缺口逐一 RE 透 / code 核實 →
> **10/11 已解**（4 個其實早已 data-driven 可取、1 個真 bug 已修、5 個已接線），剩 B-7 倉庫番 puzzle、E-11 post-game、B-6 視覺增生（皆邊際、不阻塞破關）。
> 過程揪出並修正 **3 個道具碼標籤錯位**（0x40 勇者之盾、0x4a 領悟之書、0x5b 國王的信）+ **1 個 correctness bug**（賢者轉職 gate 原認錯道具 0x40，已修為領悟之書 0x4a）。boss/事件觸發點見 [`boss-trigger-points.md`](docs/data/boss-trigger-points.md)。

**步數晝夜系統**：地表走路自動循環白天 → 黃昏 → 黑夜 → 黎明（步數驅動,各相位 palette 調暗）；
ラナルータ 咒切換、黑暗之燈道具變夜。**設定選單**（title「設定」→ RNG 模式 DOS/真實,自建字形補原版缺字）。
**per-member 裝備管理**（選隊員 → 換裝/卸下）。**忠實 rng 成長**（創角/升級 `rng(0..(target−當前))`,RE sub_d9cc）。
**戰鬥狀態咒**（拜基魯多/史卡拉/魯卡尼/瑪努莎/瑪荷頓 等 base==0 套修正）+ **野外咒文**（魯拉/烈米特/特黑洛斯）。

**離開 / 存讀檔**：F10 = 離開遊戲（「離開遊戲嗎」Yes/No 確認 + 自動存檔），ESC = 取消當前選單（非離開）；
存檔 v8（多 slot）存名冊 / 隊伍 / 道具 / 位置 / 船 / **劇情進度 storyflags** / 晝夜相位 / 魯拉去過城鎮,DQ3_LOAD 讀檔續玩,
game_tester 含存讀檔 roundtrip 斷言。
**交付驗證**:`tools/game_tester.sh` **93/93 全綠**（15 單元 + 主線一條龍 + 全 89 城零崩潰 +
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
確認 ✅、SDL2 跨平台重寫（完整可玩 + 可破關 + **93/93 驗證** + **Linux/Windows/AppImage 三包**）✅、
`[0x722]` runner/region scripted-event 系統靜態攻克 ✅、**杜勝利+青衫攻略全流程缺口盤點 10/11 已接**
（含揪出並修正 3 個道具碼標籤錯位 + 1 個賢者轉職 gate correctness bug，見
[`walkthrough-flow-audit.md`](docs/data/walkthrough-flow-audit.md)）✅。**攻略 11 缺口全接**(含 B-7 耶進貝亞倉庫番
sokoban 推三石、B-6 帶商人建城後城鎮繁榮、E-11「そして伝説へ」洛特裝備時代結尾)。剩餘:

- glyph index → Unicode 對照表（OCR 1,476 字模）→ 把劇情 dump 成純 UTF-8。
- `CHINA.FON` 殘留 ~7% 對齊、段內索引表精確語意。
- macOS / Android 移植(規劃見 [`docs/55`](docs/55-android-macos-port-plan.md))。
  細項見 [`dq3_remake/WORKLIST.md`](dq3_remake/WORKLIST.md)、[`docs/47-remaining-work.md`](docs/47-remaining-work.md)。
