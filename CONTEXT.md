# CONTEXT — 術語表 + 知識庫索引

精訊版 DQ3 反組譯專案的單一入口:**canonical 術語**(命名 / 文件 / 程式一致用詞)與
**知識庫索引**(`docs/` 全文件按主題分組)。新概念先進這裡再用;模糊詞列在末尾待釐清。

## Canonical 術語

格式:`詞 — 定義。_Avoid_: 禁用同義詞`。

### 位址 / 反組譯
- **logical (seg0-relative)** — `exe_funcs.json` / `sub_XXXX` 用的位址基準。
- **file offset** — `tools/re_disasm.py` 輸出的位址基準。換算 `file = logical + 0x1370`。_Avoid_: 兩者混稱「位址」(見 [`docs/00`](docs/00-re-methodology.md) §3 陷阱)。
- **DGROUP** — 資料段;變數 `[DS_off]` 的檔內位置 `file = 0x16140 + DS_off`。
- **scripted-event** — 由 runner `0xabb2` 經跳表 `DS 0x3baa`(id→handler)派發的事件;id 資料驅動(如彩虹水滴合成 = id 83)。
- **oracle** — DOSBox 跑原版 `DQ3.EXE` 的實機畫面 / 記憶體,當還原結果的黃金對照。_Avoid_: 「參考版」。

### 地圖 / 場景
- **CTY** — 城鎮或洞窟地圖檔 `CTYNN.DAT`;城鎮與洞窟同格式,差別在大小與 section 數。
- **section** — 一個 CTY 內的房間 / 子區。CTY 檔**無 count 前綴**,`word[i]` = section i 的 base offset(`0xffff`=空)。section 0 = 從地表進入看到的那層。
- **section header** — section_base 起的欄位:`+0/+2` NPC 清單、`+8` 事件表指標、`+0xe` layout、`+0x13/+0x14` spawn、`+0x10/+0x11` 遭遇安全旗標。
- **layout** — section 的 tile 矩陣:`{w(2), h(2), tiles…}`,tiles 在 `+4`。
- **spawn** — 玩家進入 section 的初始座標(`section_base + 0x13/0x14`)。_Avoid_: 把 spawn 當 layout 內欄位。
- **subid** — tile 高 byte 低 5 bit = 事件碼;查 section 事件表(指標在 `section+8`)。
- **BLK** — tile 圖庫 `DQ3?.BLK`,由 `map_blknum` 依 CTY 選用。
- **warp 表 (0x4ea0)** — 門 / 階梯目的地,`[param*3 + 0x4ea0]` = `{dest, X, Y}`(3 byte);type-2 事件用。
- **鑰匙門 (locked door)** — 門 tile 的 attr **低 byte bits6-7** = 所需鑰匙等級(1/2/3);隊伍持有 ≥ 該級鑰匙(道具 id 0x55 盜賊 / 0x56 魔法 / 0x57 最終,級=id−0x54)即就地改 tile 開鎖。內嵌轉場 handler `0x488f`(`0x48c3` 掃級 / `0x4906` 開門),非獨立機制。_Avoid_: 把低 byte 0xc0 當「角落/出口」(那是高 byte 0xc000)。

### 事件 / 文字
- **section 事件表** — `section+8` 指向;entry 4-byte `{type, param(u16), p2}`。type 0=調べる/寶箱、1/3=給道具、2=傳送/門/階梯、else=給道具+訊息。
- **D3TXT** — 對話 / 劇情腳本檔 `D3TXT00~09.TXT`:指標表 + 2-byte LE 記錄(`0xffff` 終止);值 `<1476`=字模索引、`>=0xffed`=控制碼。
- **FON** — 點陣字型容器(`.FON`);16×16 row-major MSB,字身 16×14、row14/15 留白。
- **glyph index** — 字模索引(0..1475),OCR → Unicode 才能 dump 純文字。
- **control-code** — 訊息字串中的負值 word(`0xffff` 結束、`0xfffc` 捲動等);全為文字格式 / 變數插入,**無** run-event 碼。

> ⚠ 命名硬規則:標位址一律顯式標 `(file)` 或 `(logical)`,不可只寫裸位址。

## 知識庫索引(`docs/`)

### 方法論(先讀)
- [`00`](docs/00-re-methodology.md) 逆向方法論:羅塞塔範本、跳表派發、位址基準陷阱、資料反證標註(可重用)

### 素材格式與還原
- [`01`](docs/01-asset-inventory.md) 資產 inventory 與格式偵察(194 檔)
- [`02`](docs/02-font-format.md) 字型格式(D3TXT00 / SETTXT / CHINA.FON)
- [`03`](docs/03-text-format.md) 文字腳本格式 D3TXT*.TXT
- [`04`](docs/04-map-format.md) 地圖 / tile 素材格式與還原
- [`27`](docs/27-bls-character-sprites.md) DQ3MAN.BLS 角色 sprite 格式
- [`33`](docs/33-name-tables.md) D3TXT00 全域名稱表(武器/道具/咒文/城鎮)

### DQ3.EXE 反組譯(RE → C)
- [`05`](docs/05-exe-recon.md) 反組譯偵察 ·
  [`06`](docs/06-exe-callmap.md) Call Map ·
  [`07`](docs/07-dpmi-note.md) DPMI 筆記
- [`08`](docs/08-exe-functions.md) 函式圖 / 啟動鏈 / 主迴圈 ·
  [`09`](docs/09-exe-loaders.md) 素材載入函式
- [`10`](docs/10-exe-states.md) 主迴圈狀態機(鍵碼跳表) ·
  [`11`](docs/11-exe-gameplay.md) 場景繪製 / 移動碰撞 / HUD
- [`12`](docs/12-exe-commands.md) 野外指令 + 對話流程 ·
  [`13`](docs/13-exe-battle.md) 戰鬥系統
- [`15`](docs/15-exe-nameinput.md) 注音姓名輸入(中文化亮點) ·
  [`16`](docs/16-monsters-ui.md) 怪物資料線 + UI 字串

### 事件 / 地圖 / 腳本系統
- [`31`](docs/31-event-system.md) 事件 / 物件 / 互動子系統(校正中)
- [`32`](docs/32-world-locations.md) 世界地點圖(cty_loc → 標記)
- [`34`](docs/34-cty-load-format.md) load_cty 載入邏輯與 CTY 格式
- [`35`](docs/35-script-format.md) 腳本 / 事件 / 轉場格式(靜態 RE 全解)

### 正確性驗證
- [`14`](docs/14-dosbox-validation.md) DOSBox 黃金對照(素材) ·
  [`17`](docs/17-build-toolchain.md) byte-identical 重組
- [`19`](docs/19-re-correctness.md) RE 正確性的確認(MSC 5.x) ·
  [`24`](docs/24-c-rebuild.md) C 重編 splice 框架 ·
  [`25`](docs/25-match-progress.md) byte-match 進度
- [`29`](docs/29-dosbox-oracle.md) DOSBox oracle 自動推進 ·
  [`30`](docs/30-phase5-comparison.md) 原版 vs SDL remake 逐畫面比對

### Bug 分析與修正
- [`18`](docs/18-bug-analysis.md) bug 反組譯定位與修正 ·
  [`20`](docs/20-fixed-version.md) 修正版(RE 價值證明)
- [`21`](docs/21-official-patches.md) 青衫官方 EXE patch 全表 ·
  [`22`](docs/22-item-fix.md) 道具特效修正 ·
  [`23`](docs/23-stat-fixes.md) 數值類 bug 修正 ·
  [`28`](docs/28-battle-palette-bug.md) 戰鬥後調色盤未還原

## Flagged ambiguities(待釐清)

- 「event」一詞橫跨 section 事件表(4-byte entry)與 scripted-event(跳表 id);文件中需標明哪一種。
- 「物件 / NPC / 實體」三詞在 `docs/31` 尚未收斂(實體表 `[0x4f15]` vs CTY 內 NPC 記錄)。
