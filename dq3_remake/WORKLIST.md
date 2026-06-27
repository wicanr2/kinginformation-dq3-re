# dq3_remake WORKLIST

> 目標:把精訊版 DQ3(Dragon Fighter III, 1993)用現代 C + SDL2 **跨平台重製**,放在 `dq3_remake/`,
> 行為與原版**一模一樣**(DOSBox 原版當 oracle),且 **7 個 bug 全在 C 層修對**。
> 紀律:**不用 subagent,全部自己在主執行緒做**;一切編譯/執行在 **docker**(不污染 host);每階段有產出就 commit+push。
> 細部 RE 機制/位址/逐步紀錄在 `docs/`(本檔只留現況 + 真正剩餘工作,不做 session 日記)。

## 現況快照(2026-06-27)

`dq3_remake/`(C99 + SDL2)= **完整、可玩、資料驅動**的精訊版 DQ3 核心,**主線可破關**。
每個數值/邏輯都從 `DQ3.EXE` 或遊戲資料抽出。`tools/game_tester.sh` **79/79 全綠**(交付 gate)。
已打包 Linux distributable;Windows/AppImage 未做。

```
阿里阿罕起步 → 露依達酒場創角(職業→注音/英數命名→性別)→ 名冊/隊伍
  → 地表(真實 DQ3CON.MAP + 原版起點 + 真實遭遇區)→ 城鎮(CTY/NPC/繁中對話/門/轉場)
  → 戰鬥(指令窗 + 手動/自動選咒[EXE 威力] + 怪物 AI + 裝備加成 + 升級 + 回寫名冊 + 金錢)
  → 商店(真實 ITEM.DAT)→ つよさ/じゅもん/道具 → 存讀檔(多 slot)
```

## 已完成(大模組,細節見 docs/)

- **素材全解碼**:字型(02)、文字+UTF-8 劇本(03)、地圖 tile/世界圖/CTY 城鎮(04)、標題/PCX、怪物 sprite+數值(16,37-39)、道具(22)。
- **EXE 反組譯**:recon(05)、callmap+runtime(06)、函式圖(08)、loaders(09)、戰鬥(13)、注音輸入(15);整檔反組譯→nasm 重組 **sha256 == 原版(byte-identical)**(17),編譯器=MSC 5.x(19)。
- **remake 引擎**:scene(攝影機/tile/碰撞/走動)、field/town、真主角+NPC sprite(DQ3MAN.BLS,27)、text/dialogue(控制碼)、battle(指令/AI/施法/升級回寫)、roster/tavern、注音組字 IME、status/spell/item 畫面、多 slot 存讀檔(DQ3SAVE5 + F5/F9 + 標題選單)、鑰匙門(35§八)、NPC mover(35§九)、retro-cjk hires 畫布(nearest 整數放大)。
- **7 bug C 層全修**(階段④,誠實狀態見 docs/28 及各 dq3_* 單測):#1 結算/#2 合成(產 0x75)/#3 blit guard+復原 sprite/#4-6 升級系統(成長表+clamp)/#7a 隼劍雙擊/#7b 魔甲抗魔/#8 palette 還原。#7c 祈禱之戒原版本就壞,不修。
- **主線可破關**:8 里程碑真實 NPC 觸發一條龍;`[0x722]` runner/region scripted-event **靜態攻克**(re-log-722,57 setter)。
- **杜勝利攻略接線**:B 道具鏈(夢幻紅寶石/變身杖/蓋亞之劍/鑰匙鏈/彩虹水滴/領悟之書/王者之劍)、A boss(巴拉摩斯本體+怨靈122/殭屍123/索瑪124 序列、甘達特金皇冠正源、甘達特巢穴 CTY14、八頭大蛇 CTY19)、不死鳥拉米亞坐騎(六珠)。
- **道具取得鏈全 ✅**:`docs/data/quest-items.md` 全列(寶箱/sub2 NPC/位置 item-use/boss/合成),無 stale「待」。
- **special 事件**:`special-events-classified.json` 全分類,**0 review 待辦**(CTY6 折扣/CTY65 換蛋已第一性原理確認非缺口,audit §44-71)。
- **知識庫**:`CONTEXT.md`(術語表+docs 索引)、`docs/00-re-methodology.md`、`docs/data/oracle-validation.md`(DOSBox 比對)、`~/.claude` 跨 session 記憶。

## 真正剩餘工作(已對 ground truth 核實,非 stale)

### 內容 / 接線
- [ ] **scripted warp 全接**(8 個 0xd1f9):`dq3_locwarp` 有落點,缺「源觸發 tile」位置 → 需各 call site disasm 抽 `[0x4f33/35]==XY` 比較值。(忠實完整性,非可玩性阻塞)
- [ ] **overworld 旗標 portal 全表**:`dq3_owportal` 目前數條;完整需抽 0x396e 全分支。(忠實完整性)
- [x] **怪物施加狀態 ✅**(已確認 + 修 phantom-damage):第一性原理 RE(`docs/re-log-spell-effect-dispatch.md`)定論——精訊引擎效果由 base 值分類、**無「咒→狀態」表**;base==0 中只有 144 睡/152 混亂=玩家不能行動 → 映射麻痺(已接),其餘 base==0(buff/debuff/封咒/幻/傳走)未模型化。本次修掉「base==0 咒 fall through 成 base=24 假傷害」bug → 詠唱不致傷。中毒(overworld/戰鬥毒傷+解)亦閉環。
- (移除)~~CTY→地名對照收尾~~:使用者確認不影響遊戲,不列待辦。

### 晝夜系統 ✅(2026-06-27 實作;`docs/data/day-night-system.md`)
- [x] **步數驅動晝夜 ✅**(★更正:原版**走路就會自動循環**白天/黃昏/黑夜/黎明,**步數計數**非時間;
  先前誤判「不自動」已撤回,使用者 ground truth 指正)。實作:`dq3_scene` 4 相位 + palette 調暗
  (夜暗偏藍/黃昏暗偏暖/黎明微暗;sprite 較淺保持可見);`main.c` 地表每 28 步(近似)推進相位;
  ラナルータ(rec177 咒)toggle + 黑暗之燈(0x5f 道具)變夜;debug `dn:N`。驗證:phase0 vs phase2
  dump diff 494437、夜景調暗、game_tester 79/79。
- [ ] 晝夜精校/補完:確切原版步數閾值 + 各相位 palette(目前近似)、存檔持久化(載入回白天)、
  夜 gated 事件接 g_dn_phase(提頓村牢房)。

### 系統 / UI
- [ ] **遊戲內設定選單 UI**:title 加「設定」→ 切 RNG 模式(DOS/REAL)存 dq3.cfg。後端 `dq3_config` 已備,只差 UI(未來音量/解析度/語言同此)。
- [x] **手動選咒選單 ✅**(已實作,WORKLIST 此項 stale):`dq3_battlescene` 戰鬥指令「咒文」開
  互動選咒子選單(列領頭施法成員已學可施放咒 + MP,選定設 `g_manual_cast_*`);無選則自動放最強攻擊咒。
- [x] **注音↔英數 切換鍵 ✅**(已實作,此項 stale):`dq3_tavern.c` 創角姓名輸入 **Tab 鍵**
  (sc 0x0f)`name_mode ^= 1` 切換英數盤(`dq3_nameinput`)↔注音盤(`dq3_zhuyin`),各自渲染/輸入。
- [x] **設定選單 UI ✅**(2026-06-27):title 加「設定」第三項 → `config_modal` 切 RNG 模式
  (原版 DOS / 真實 REAL)存 dq3.cfg。★解 glyph blocker:**自建字形 pipeline**
  (`tools/gen_custom_glyphs.py` 用 NotoSansCJK rasterize CJK → 16×16 font 格式;`dq3_customglyph.c/h`;
  `dq3_text_draw_glyph` 對 idx≥0xf000 fallback)補原版字庫缺的「設/版」。dump 驗證畫面清楚可讀。
- [x] **per-member 裝備管理 UI ✅**(2026-06-27):cmd「裝備」開互動管理 `equip_modal`:選隊員→選槽
  (武器/防具)→ 從背包換裝(可裝備性 `dq3_item_can_equip` + category 過濾)/ 卸下(舊品回背包)。
  戰鬥加成沿用既有(武器 b0→atk、防具 b1→def、#7a 雙擊/#7b 抗魔)。debug:DQ3_EQUIP_DUMP+equip token。
  待:擴 5 槽(盾/頭/飾)— 目前 2 槽(武器/防具)。
- [ ] **忠實初始擲值 / RNG 成長**(需 RE + 動升級系統):★RE 已定位(2026-06-27)——成長 handler
  sub_d9cc(file 0xed3c)算 `base+slope×level` 後 **`call 0xfa57`(= EXE RNG `[0xb5a]+0x9018;rol×3`)
  套 rng 變異**;docs「確定性上限」= target 是上限、實際值 = rng 到上限。remake 升級(#4/#5/#6)現用
  **確定性上限**(已測試簡化)。要忠實 = 把創角 + 每次升級改成 rng(0..target),需重做並重驗升級單測
  (回歸風險)。非可玩性阻塞。精訊版無性格系統(已確認)。

### 校準(已靜態 RE 攻克;原列「需 DOSBox」已不成立)
- [x] **旅社/教會收費公式 ✅**(靜態 RE,2026-06-27):**旅社**已精確(`fac->inn_cost × 人數` = RE
  handler 0x86f5 `mov ax,es:[di+1]; mul cl`,設施 block +1 per-town);**教會復活**改成 RE level 表
  (handler 0x85ff,表 @ DGROUP 0x3c6c / file 0x19dac,index=min(lv,40)−1;`DQ3_REVIVE_COST[40]`),
  取代「level×10」近似(Lv1-3 應同 10、Lv40 應 1610)。解毒(0x849b 5 固定)未驗證確為解毒,保守未改。
- [x] **傷害公式 ✅ 驗證 RE 精確**(2026-06-27,無需校準):物理 `(atk-def)/2+rng(0..(atk-def)/4)`、
  弱攻 `rng<0x80→1 else miss`、會心 ×2(`[0x24b7]&0x10`)、咒文 `base/2+rng(base/2)` —— 逐一對
  反組譯(file 0xc03e/0xc07a/0xc0cd/0xc22e)確認 remake **完全一致**。會心率 1/32 為 DQ3 標準值
  (×2 機制已驗;精確 roll 指令在攻擊指令 handler,未定位但屬單一標準常數)。
- [ ] **DOSBox 逐畫面 oracle 擴展**:標題逐像素/注音 IME 字序/boss sprite 已驗一致;其餘畫面待補。
- [ ] **packbg 戰鬥背景解碼**:戰鬥背景目前用通用天空/地面,packbg 壓縮背景未解碼。

### 咒文效果(2026-06-27,研究驅動實作;研究+實作對照 `docs/data/spell-effects-research.md`)
- [x] **戰鬥狀態/輔助咒 ✅**:base==0 咒(拜基魯多→敵攻×2、史卡拉/史克魯多→敵守↑、魯卡尼/魯加南→
  我守↓、瑪努莎→幻惑物攻失手、瑪荷頓→封咒)在 `dq3_battlescene` 套 per-battle 修正狀態。
  ★修真 bug:前版這些咒 fall through 成 base=24 假傷害 → 改套修正/詠唱不致傷。其餘 base==0(皮歐里姆/
  夫巴哈/阿斯特龍/反彈等)誠實標「未模型化、不致傷」。出處:BBS 一手史料 + Wiki/RoD + RE re-log。
- [x] **野外咒文 ✅**:cmd「咒文」改可施放 `field_cast_modal`:魯拉/烈米特→回地表、特黑洛斯→驅弱敵
  (扣會該咒隊員 MP)。清單檢視移到「狀況」Space。待:因帕斯(米米克判定)/多拉瑪那(防傷地板)/
  拉那魯達(晝夜)/帕爾奔得(隨機)需更多世界系統。

### 打包
- [ ] **Windows / AppImage 跨平台打包**:目前 `tools/package.sh` 只產 Linux tar(`work/dq3_remake_dist.tar.gz`)。

### Chesterton's Fence 三項 → 已執行(使用者要求,2026-06-27)
- [x] **#7c 祈禱之戒 ✅**:RE file 0x5ad0/0x5af4 → 回 MP + 每次使用 ~25.4%(al≤0x40)損壞消失,
  接進 `dq3_item_use`(野外/debug 兩路徑)+ 單測。青衫「永不壞」前提經 RE 推翻(損壞邏輯本就存在)。
- [x] **寶箱開過回饋 ✅**(remake 增強,非還原):原版取後不翻 tile(docs/31);使用者選「疊變暗/開蓋標記」
  → `dq3_scene_mark_opened_tile` 程式疊繪(不動 BLK),`overlay_opened_chests` 對已取格套用。dump 驗證。
- [x] **金字塔施咒當機 ✅**(RE + 驗證免疫):原版逐圖資源型 crash;remake 咒文與地圖資源解耦
  (戰鬥 map-independent、野外咒文不跑 overlay)→ 結構免疫,headless 驗 CTY13 進場+施咒 exit=0
  不 crash(`docs/data/original-known-bugs.md` 詳記)。
- 仍保留不復刻:原版金字塔 crash 本身、#7c 的「永不壞」誤傳。

## 關鍵參考 / 工具 / 紀律

- **格式/位址**:docs/02–16(素材)、docs/05–13/15(EXE 邏輯)、`docs/data/exe_funcs.json`(函式圖)、`docs/35`(轉場/鑰匙門/NPC mover)、`docs/37-39`(怪物 AI/數值/遭遇)。
- **build / 測試(容器內,傳 /repo /build)**:
  ```bash
  docker run --rm -v "$PWD":/repo -v dq3build:/build dq3-remake bash -lc 'cd /build && cmake --build .'
  docker run --rm -v "$PWD":/repo -v dq3build:/build dq3-remake bash -lc 'bash /repo/tools/game_tester.sh /repo/assets_raw /build/dq3_remake'
  ```
  > 坑:`game_tester.sh` 硬編 `/repo`/`/build`,host 直跑→全假失敗(環境非回歸)。
- **其他工具**:`scripts/build.sh`(docker SDL2 + headless 標題驗證)、`tools/package.sh`(打包)、`tools/dockrun.sh`(docker uv python)、`tools/dosbox_run.sh`+`dq3-dosbox`(原版 oracle)。
- **原始素材**:`assets_raw/`(版權,gitignore;remake 執行期指向它)。
- **不公開**:可玩遊戲包(`work/`)、原始檔、第三方攻略/截圖、render 的版權畫面像素 —— 見各 `.gitignore`。
- **紀律**:不用 subagent;docker first;每階段 commit+push;斷言完成前 git diff/檔案中介數字驗證(memory `re-verify-via-git-not-tool-echo`);commit message 繁中 + `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`。
