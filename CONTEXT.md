# CONTEXT — 術語表 + 知識庫索引

精訊版 DQ3 反組譯專案的單一入口:**canonical 術語**(命名 / 文件 / 程式一致用詞)與
**知識庫索引**(`docs/` 全文件按主題分組)。新概念先進這裡再用;模糊詞列在末尾待釐清。

## Canonical 術語

格式:`詞 — 定義。_Avoid_: 禁用同義詞`。

### 位址 / 反組譯
- **logical (seg0-relative)** — `exe_funcs.json` / `sub_XXXX` 用的位址基準。
- **file offset** — `tools/re_disasm.py` 輸出的位址基準。換算 `file = logical + 0x1370`。_Avoid_: 兩者混稱「位址」(見 [`docs/00`](docs/00-re-methodology.md) §3 陷阱)。
- **DGROUP** — 資料段;變數 `[DS_off]` 的檔內位置 `file = 0x16140 + DS_off`。
- **scripted-event** — runner `0xabb2` 的 event-id 派發基址是 `DGROUP 0x3baa`；直接 subtype2 NPC dispatcher logical `0x4fe5..0x5001` 則以 `DGROUP 0x3bb4 + byte4×2` 取 handler。兩者是同一張表相差 5 entries，不得互換 index。
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
- **NPC 槽 / 移動** — NPC 載到 `[0xb66]` 8-byte 槽(byte0=X/1=Y/3=移動控制/6=底 tile)。slot[3] **bit2=移動開關**(set=隨機走動、clear=靜止)、bits0-1=朝向、bit7=凍結。mover=`L02025`(file 0x3395),每幀 RNG(4)==1 才評估、1/20 轉向、不貼玩家 3 格。_Avoid_: mover 位址寫成 `0x6240`/`0x62e9`(舊誤標,是 init/劇情碼)。
- **鑰匙門 (locked door)** — 門 tile 的 attr **低 byte bits6-7** = 所需鑰匙等級(1/2/3);隊伍持有 ≥ 該級鑰匙(道具 id 0x55 盜賊 / 0x56 魔法 / 0x57 最終,級=id−0x54)即就地改 tile 開鎖。內嵌轉場 handler `0x488f`(`0x48c3` 掃級 / `0x4906` 開門),非獨立機制。_Avoid_: 把低 byte 0xc0 當「角落/出口」(那是高 byte 0xc000)。
- **特殊地形 type** — 移動函式 `sub_11A6A` 對 `attr&0xe000` 兩次 `rol` 後 `and 3`；
  例如塔的 `0x2000`=type0 跌落/離場，`0x4000`=type1 樓梯。_Avoid_: 把所有高位 attr
  都籠統叫樓梯，或把 logical `0xf6d6` 滑鼠 hit-test 當玩家座標 region。

### 事件 / 文字
- **section 事件表** — `section+8` 指向;entry 4-byte `{type, param(u16), p2}`。type 0=調べる/寶箱、1/3=給道具、2=傳送/門/階梯、else=給道具+訊息。
- **D3TXT** — 對話 / 劇情腳本檔 `D3TXT00~09.TXT`:指標表 + 2-byte LE 記錄(`0xffff` 終止);值 `<1476`=字模索引、`>=0xffed`=控制碼。
- **FON** — 點陣字型容器(`.FON`);16×16 row-major MSB,字身 16×14、row14/15 留白。
- **glyph index** — 字模索引(0..1475),OCR → Unicode 才能 dump 純文字。
- **control-code** — 訊息字串中的負值 word(`0xffff` 結束、`0xfffc` 捲動等);全為文字格式 / 變數插入,**無** run-event 碼。
- **rec400 指令窗** — D3TXT00 記錄 400 = 精訊版野外指令窗本尊(命令 + 2×3 格 對話/咒文/狀況/道具/裝備/調查);UI 選單以「框線 glyph + 標籤」存成單一 record(同類:rec407 狀態畫面、rec421 道具 使用/給予/丟掉、rec441-444 戰鬥指令)。_Avoid_: 人工猜指令中文名(docs/12 舊標)。
- **共用對話視窗 rect** — DGROUP `0x3e6e`（file `0x19fae`）的原版結構；`+2=0x13`
  是 X VGA byte、`+4=0xee` 是 Y pixel、`+6=0x2c` 是內容寬 VGA byte、`+8=0x60`
  是內容高 pixel。`sub_1fb36` 加左右 1 byte／上下 `0x10` 後，外框為
  `(152,238,352,96)`，文字 inset `(16,16)`、20 欄、每頁 4 行；`360×112` 只是
  `sub_1fb36` 關閉時背景備份範圍，不能當可見框。_Avoid_: 沿用 remake
  的 `(24,244,592,96)` 或把 `+6/+8` 寬高對調（見 [`docs/94`](docs/94-dialogue-window-and-monster-mask-re.md)）。
- **道具名 rec** — D3TXT00 記錄號 = 道具碼 + 1(docs/33);咒文名 rec = `spell_id + 0x79`。

### 角色 / 數值
- **7 屬性序** — 成長表 14 byte = 7 個 (base,slope) 對,`kind*2`=列內 offset:0 HP、2 MP、4 速度、6 力量、8 聰明、A 耐力、C 運氣。語意三方交叉確認(成長表樣式 + 升級訊息 rec191-197 + BBS 存檔佈局 docs/history)。_Avoid_: 把屬性欄標成 B4/B6/B8/BA 不明名(舊 enum)。
- **怪物 AI 欄位** — `D3MNS.DAT` 記錄內驅動敵方行動的 4 欄:`+0x0d` 施咒機率(/256)、`+0x0e..+0x13` 已知咒文 bitmask(6 byte/48 bit,放咒時均勻隨機挑)、`+0x17` 逃跑觸發閾值(對我方強度 `[0x5094]`)、`+0x18` 逃跑機率。決策樹 file 0xbcf0;見 [docs/37](docs/37-monster-ai.md)。_Avoid_: 把怪物施咒當「挑最佳咒」(實為隨機)。
- **怪物掉落欄位** — `D3MNS.DAT` 記錄 `+0x25` 是掉落判定閾值、`+0x26` 是掉落道具；
  勝利結算在 logical `0xc425`（file `0xd795`）以 `DGROUP 0x2321` 的 formation 第一組
  怪物 raw ID 選記錄，依 `roll(256) <= +0x25` 判定，再讀
  `+0x26`。戰鬥狀態 `DGROUP 0x2518 bit1` 會抑制掉落。`+0x27` 尚未閉合，維持
  `unknown_27`。_Avoid_: 沿用舊 parser 把 `+0x27` 稱為掉落率。
- **SHP AND-mask** — `DQ3MNS.SHP` 每隻怪的四個色彩 plane 後另有逐列 RLE 遮罩；mask bit
  1 保留背景（透明），bit 0 清背景後寫色彩（不透明）。調色盤索引 0 是黑色且可為實體像素。
  _Avoid_: `palette index 0 = transparent`（見 [`docs/94`](docs/94-dialogue-window-and-monster-mask-re.md)）。
- **戰鬥睡眠／混亂恢復** — 玩家角色 `+0x38 bit0x20` 在行動前
  `roll<=0x64` 清除；怪物 bit0x80 是 `roll<0x64` 清除，醒來當回合都跳過。見
  [`docs/37`](docs/37-monster-ai.md)、[`docs/91`](docs/91-garuna-satori-book-production-trace.md)。
  _Avoid_: 舊 C remake 的「持續整場、不會自行解除」結論。
- **日邦格八頭大蛇兩階段事件** — 第一階段由 CTY19 sec1 handler45、固定 NPC
  `(35,12)` 進入；怪物75的原始掉落給草薙大劍，勝利後向右移動兩步並轉至 CTY21
  顯示 D3TXT04 rec70。第二階段由 CTY21 sec0 handler36、NPC `(18,7)` 進入，rec66
  Yes 分支只顯示 rec67 且可重談，No 分支以 rec68 開第二戰；第二戰設掉落抑制，勝利後
  rec69 與旗標交易開放同格寶箱取得紫寶珠。詳見 `docs/95`。_Avoid_: 第一戰直接給紫寶珠、
  省略自動轉場，或把兩戰合併為單一 `bossTrigger`。
- **愛丁貝亞追蹤守衛** — CTY39 sec0 handler29（raw `0x1d`）在玩家踏入
  `(13,37)/(14,37)` 時，未隱形會令 slot0 守衛沿 X 追到玩家同欄；`DGROUP 0x4f46`
  bit `0x8000` set 時直接略過追蹤。隱形草 `0x5d` 會消耗並設定 25 步暫態；玩家仍不能
  穿越一般 NPC，只能走守衛未占據的右欄。詳見
  [`docs/96`](docs/96-invisibility-grass-eginbear-production-trace.md)。_Avoid_: 把隱形解讀成
  全域 NPC passthrough，或把朗錫爾 CTY47/75 當成 `0x5d` 貨架證據。
- **愛丁貝亞推石** — CTY76 sec0 的 NPC slot0–2 初始位於 `(3,10)/(5,10)/(7,10)`。
  玩家碰撞 consumer 以 runtime NPC `ctrl bit0x40` 判定可推，並要求前方地形與 NPC stamp
  可用；`DGROUP 0x4f46 & 0xc000` 非零時禁止推動。handler30（raw `0x1e`）固定檢查
  slot0–2 的 Y 是否全為 5，成立才 clear flag `0x3b` 並開 passage；event0／flag `0x3a`
  才是乾渴壺 `0x5e`。詳見 [`docs/97`](docs/97-eginbear-push-puzzle-production-trace.md)。
  _Avoid_: 以 sprite `B2==40` 判 pushability、加入 remake 自造 `sokoban_solved`，或把
  event1／flag `0x3b` 當第二個乾渴壺。
- **勇氣試煉暫時單人模式** — CTY47 sec0 handler37 `(25,16)` 接受後保存 active party
  count、強制 count=1、寫 world `(82,165)` 並設 `DGROUP 0x4f46` bit `0x80`；CTY75
  handler62 `(26,7)` 還原 count 並清 bit。flag `0x13` 是完成分支 reader，writer 仍
  `unknown`；原版 world reader 為 clear→CTY75、set→CTY47。見 [`docs/100`](docs/100-lancel-courage-blue-orb-production-trace.md)。
  _Avoid_: 接受或取得藍寶珠時自行 set flag `0x13`、反向 `owPortal`、宣稱 solo 機制 moot。
- **海盜村紅寶珠密道** — CTY27 sec0 `(26,9)` 物件的 runtime control `0xc1` 含
  `bit0x40`，推開後露出 transition subid1，進 sec1 `(5,9)`；event0 是紅寶珠 `0x68`，
  flag `0x3f` 同時控制寶箱與入口物件 visibility。`B2=0x28` 的精確 sprite 語意尚未閉合，
  canonical 名稱只用「可推物件」。見 [`docs/101`](docs/101-pirates-red-orb-production-trace.md)。
  _Avoid_: 只走普通入口、以攻略名稱取代 sprite 證據，或不經正式魯拉船重定位就注入座標。
- **遭遇區表** — overworld region map `0x4966`(16×16 grid,cell=(X>>4)+(Y&0xf0)→region)+ 遭遇表 `0x4a56`(region×0x20 = 4 子表×8 byte;byte2=背景頁、byte4-7=候選怪 0xff 終止)。CTY 用 `[0xd77]` 存的 region(0=安全)。見 [docs/39](docs/39-encounter-zones.md)。
- **咒文習得表** — `sub_db5f`(file 0xeecf)讀 DGROUP `0x36f9` 起 per-系 stride 8 指標表 `{spellA,levelA,spellB,levelB}`;系基底 0/8/0x10=勇者/僧侶/魔法系;`level[i]`/`spell[i]` 平行,清單長度由下一指標界定(越界=#5 亂學咒 bug)。職業→系:勇者→勇者系、僧侶→僧侶系、魔法使→魔法系、賢者→僧侶+魔法、其餘無咒。

> ⚠ 命名硬規則:標位址一律顯式標 `(file)` 或 `(logical)`,不可只寫裸位址。

## 知識庫索引(`docs/`)

### 方法論(先讀)
- [`00`](docs/00-re-methodology.md) 逆向方法論:羅塞塔範本、跳表派發、位址基準陷阱、資料反證標註(可重用 RE *技術*)
- ⭐ [經驗教訓 Lessons Learned](docs/lessons-learned.md) 多輪協作教訓:別鬼打牆(已完成別當未完成重查)、ground truth=code+使用者、靜態 RE 優先不依賴 DOSBox、柵欄原則用對、斷言前驗證、NULL=現行零回歸、記憶只存教訓(*怎麼工作不繞遠路*)

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
- [地圖連通圖](docs/maps/map_graph.md) — 全 CTY 轉場連接(metadata 自動抽出;89 CTY/50 跨 CTY 連結)
- [`36`](docs/36-original-tavern-playthrough.md) 原版 DOSBox 實機玩法:開機→姓名輸入→露依達酒場創角(知識庫;姓名輸入按鍵語意實機校正)
- [`37`](docs/37-monster-ai.md) 怪物 AI(敵方行動邏輯)反組譯:狀態→逃跑→物理/咒文決策樹;D3MNS AI 欄位(+0x0d 施咒率/+0x0e..13 咒文 bitmask/+0x17 逃閾/+0x18 逃率),資料驅動驗證
- [`38`](docs/38-monster-stats.md) 完整怪物屬性表(130 隻):HP/攻/防/速/經驗/金錢 + AI(施咒%/逃跑)+ 會用的法術(bitmask→真咒名)
- [`39`](docs/39-encounter-zones.md) 遭遇區系統:overworld region map(0x4966 16×16 grid)+ 遭遇表(0x4a56,region→4 子表→候選怪);位置→region→怪

### 歷史 / 一手史料
- [BBS 1994–1995 討論串](docs/history/dq3-bbs-1994.md) — 當年 BBS 攻略 / bug 修改 / 存檔佈局 / 道具咒文表;精訊官方公告(1993 開發、ENIX 拒、離職員工外流);**RE 一手交叉佐證**(魔王打不死 / 彩虹水滴 / 五頭龍當機修改法對上 bug #1/#2/#3、存檔屬性欄位佈局對上成長表)
- [攻略來源致敬 + 地名對照](docs/history/dq3-walkthrough-credits.md) — 杜勝利 DQ3 攻略 + Mike's RPG Center 世界地圖致敬;FC 英文↔精訊中文↔CTY 地名對照 + 進度順序(供 CTY→地名 收尾)

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
- [`bugs.md`](docs/bugs.md) 青衫 bug 清單原文(7 個 bug 的一手描述,docs/18 的來源)

### 設施 / 道具 / 船 / 傳送 / 接線(40+)
- [`40`](docs/40-facility-shops.md) 設施/商店系統 · [`41`](docs/41-treasures.md) 寶箱 ·
  [`49`](docs/49-item-use.md) 道具使用效果 · [`22`](docs/22-item-fix.md) 裝備/ITEM.DAT
- [`43`](docs/43-reachability.md) 可達性/傳送 · [`44`](docs/44-scripted-warp-re.md) scripted warp ·
  [`45`](docs/45-binary-structures.md) 二進位結構彙整 · [`42`](docs/42-npc-dialogue.md) NPC 對話
- [`48`](docs/48-ship-navigation.md) 船航行 · [`50`](docs/50-ship-acquisition.md) 取船 ·
  [`51`](docs/51-palette-and-overworld-sprites.md) palette/地表 sprite · [`46`](docs/46-debug-hook.md) DEBUG 口 ·
  [`69`](docs/69-why-remake-looked-done-but-wasnt.md) 失敗經驗與防再犯方法、
  [`74`](docs/74-ebiten-remake-completion-plan.md) 現行完成計畫 ·
  [`84`](docs/84-game-pack-json-contract.md) 共用 engine／versioned game pack JSON 契約 ·
  [`108`](docs/108-battle-text-pack.md) 戰鬥文字 record／glyph 對映 ·
  [`109`](docs/109-battle-hud-rects-re.md) IDA 戰鬥 HUD rect／raw flags／共用 frame JSON 證據 ·
  [`110`](docs/110-field-command-labels-re.md) 地表命令窗雙字模資料化 ·
  [`111`](docs/111-battle-scene-layout-re.md) 戰鬥場景帶／游標 glyph 資料化與 D2 限制 ·
  [`112`](docs/112-newgame-labels-re.md) 開場／創角 glyph、45 格功能列資料化與 D2 限制 ·
  [`113`](docs/113-newgame-geometry-re.md) IDA 9.4 創角 raw window、格距、功能焦點與像素 rect ·
  [`114`](docs/114-release-artifacts.md) AppImage、Windows ZIP、macOS ZIP 與推廣片產物／驗證界線 ·
  [`115`](docs/115-encounter-cty-gate-re.md) IDA 9.4 CTY `+0x11` 遭遇 gate polarity、步數計數器與 Go 接線 ·
  [`116`](docs/116-field-status-panel-re.md) IDA 9.4 `DGROUP 0x3DA8`／record 407 詳細狀況窗、JSON anchors 與 E2／V2 runtime 對拍 ·
  [`88`](docs/88-norud-guided-passage-production-trace.md) 諾魯德密道 handler50/57、
  NPC 移動腳本與正式玩家輸入追蹤 ·
  [`91`](docs/91-garuna-satori-book-production-trace.md) 加爾那之塔領悟之書、睡眠恢復
  IDA 證據與正式玩家輸入追蹤 ·
  [`98`](docs/98-thirsty-pitcher-final-key-production-trace.md) 乾渴壺唯一成功格、world-state
  `0x08`、`5×4` 地圖 patch、CTY40 最終鑰匙與正式玩家輸入追蹤 ·
  [`99`](docs/99-teidon-final-key-green-orb-production-trace.md) 最終鑰匙正式使用、魯拉船重定位、
  提頓 handler35 綠色寶珠與 boot production trace ·
  [`100`](docs/100-lancel-courage-blue-orb-production-trace.md) 蘭西爾 handler37／62、暫時單人、
  CTY23 藍寶珠、途中／復隊存讀檔與 boot production trace ·
  [`101`](docs/101-pirates-red-orb-production-trace.md) CTY27 可推入口、密道 transition、
  紅寶珠、魯拉船重定位與 boot production trace ·
  [`102`](docs/102-merchant-settlement-world-entrance-re.md) 商人聚落 ordered gate、建城商人
  transaction、CTY83 黃寶珠 ·
  [`103`](docs/103-samanoasa-mirror-production-trace.md) CTY24 拉之鏡 JSON、商人城 CTY60
  自然銜接、沙曼歐莎假王與怪力魔 boot production trace

### 資料 / 研究(`docs/data/`)
- [道具取得鏈](docs/data/quest-items.md) · [咒文效果研究](docs/data/spell-effects-research.md) ·
  [晝夜系統](docs/data/day-night-system.md) · [special 事件盤點](docs/data/special-events-audit.md) ·
  [boss 觸發點](docs/data/boss-trigger-points.md) · [攻略接線盤點](docs/data/walkthrough-flow-audit.md)
- [sub2 結構](docs/data/sub2-struct.md) · [sub2 handlers](docs/data/sub2-handlers.md) ·
  [sub2 live NPC](docs/data/sub2-npcs-live.md) · [DGROUP 表](docs/data/dgroup-tables.md) ·
  [oracle 驗證](docs/data/oracle-validation.md) · [原版已知 bug](docs/data/original-known-bugs.md) ·
  [結局未完成](docs/data/endtxt-incomplete.md)

### RE 攻關日誌 / 教訓
- [`[0x722]` runner 狀態機](docs/re-log-722-state-machine.md) ·
  [咒文效果 dispatcher](docs/re-log-spell-effect-dispatch.md) ·
  [remake 接線教訓](docs/remake-wiring-lessons.md)

### 2026-08-11 目前 IDA story-flag E2 接線
- `dq3_cht` pack schema `0.1.29` 的 `events.json.story_flag_runtime_events` 保存四條已閉合
  handler transaction：74／CTY80 `0x0d`、69／CTY70 `0x11`＋map byte `0x80`、70／CTY65
  `0x1a/0x1b`、33／CTY44 `DI=0x0bff→record71` 的 `0x24/0x21`。
- 共用 engine 只執行固定 primitive；`game/story_flags_test.go` 以正式 command／battle path、
  Phoenix save/load map replay、Baramos aftermath 與 ending writer 驗收。完整 V3 route、
  palette register、SHP／PCM timing、boss repeat-N 與未閉合 story 語意仍保持 `unknown`／fail-closed。
- CTY07→CTY08 route 勘誤（2026-08-11）：原始 transition chain 已接通；塔區遭遇由正式聖水／
  戰鬥輸入處理後可抵達 CTY08 sec3。完整 campaign 後段仍有其他路徑 fixture，不能把這條
  route 的修正誤稱為 E3 全程通關；呈現與時序近似依 README 的 `V3-approx` 標記。
- 詳細目前狀態：[`docs/119`](docs/119-daynight-story-flag-audit.md)、
  [`docs/123`](docs/123-static-battle-daynight-re.md)、[`docs/74`](docs/74-ebiten-remake-completion-plan.md)、
  [`docs/84`](docs/84-game-pack-json-contract.md)。

### 雜項工具筆記
- [js-dos bundle 下載技巧](docs/js-dos-bundle.md)(dos.zczc.cz 線上模擬器抽 bundle zip)

## Flagged ambiguities(待釐清)

- 「event」一詞橫跨 section 事件表(4-byte entry)與 scripted-event(跳表 id);文件中需標明哪一種。
- 「物件 / NPC / 實體」三詞在 `docs/31` 尚未收斂(實體表 `[0x4f15]` vs CTY 內 NPC 記錄)。
