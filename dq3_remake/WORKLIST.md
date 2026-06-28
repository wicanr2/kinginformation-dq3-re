# dq3_remake WORKLIST

> 目標:把精訊版 DQ3(Dragon Fighter III, 1993)用現代 C + SDL2 **跨平台重製**,放在 `dq3_remake/`,
> 行為與原版**一模一樣**(DOSBox 原版當 oracle),且 **7 個 bug 全在 C 層修對**。
> 紀律:**不用 subagent,全部自己在主執行緒做**;一切編譯/執行在 **docker**(不污染 host);每階段有產出就 commit+push。
> 細部 RE 機制/位址在 `docs/`;本檔只留**現況 + 真正剩餘工作**,完成項收進「已完成」摘要,不做 session 日記。

## 現況快照(2026-06-27)

`dq3_remake/`(C99 + SDL2)= **完整、可玩、資料驅動**的精訊版 DQ3 核心,**主線可破關**。
每個數值/邏輯都從 `DQ3.EXE` 或遊戲資料抽出。`tools/game_tester.sh` **93/93 全綠**(交付 gate)。
已打包 Linux tar + **Linux AppImage** + **Windows x64 zip**(交叉編譯,wine 驗證);macOS/Android 待移植(規劃見 docs/55)。

```
阿里阿罕起步 → 露依達酒場創角(職業→注音/英數命名→性別)→ 名冊/隊伍
  → 地表(真實 DQ3CON.MAP + 原版起點 + 真實遭遇區 + 步數晝夜)→ 城鎮(CTY/NPC/繁中對話/門/轉場)
  → 戰鬥(指令窗 + 手動/自動選咒[EXE 威力] + 怪物 AI + 狀態咒 + 裝備加成 + rng 升級 + 回寫名冊 + 金錢)
  → 商店(真實 ITEM.DAT)+ 裝備管理 → つよさ/じゅもん/道具 → 設定 → 存讀檔(多 slot,含劇情進度)
```

## 已完成(大模組,細節見 docs/)

- **素材全解碼**:字型(02)、文字+UTF-8 劇本(03)、地圖 tile/世界圖/CTY(04)、標題/PCX、怪物 sprite+數值(16,37-39)、道具(22)、戰鬥背景 packbg(PACKBG.SCR)。
- **EXE 反組譯**:recon(05)、callmap/runtime(06)、函式圖(08)、loaders(09)、戰鬥(13)、注音(15);整檔反組譯→nasm 重組 **sha256 == 原版**(17),編譯器 MSC 5.x(19)。
- **remake 引擎**:scene(攝影機/tile/碰撞/走動)、field/town、主角+NPC sprite(27)、text/dialogue、battle(指令/AI/施法/升級)、roster/tavern、注音 IME(英數↔注音 Tab 切換)、status/spell/item 畫面、鑰匙門(35§八)、NPC mover(35§九)、retro-cjk hires 畫布、**步數晝夜系統**(4 相位+palette 調暗+ラナルータ/黑暗之燈)、**自建字形 pipeline**(`gen_custom_glyphs.py`,補原版缺字)。
- **7 bug C 層全修**:#1 結算 / #2 合成(產 0x75)/ #3 blit guard+復原 sprite / #4-6 rng 升級系統 / #7a 隼劍雙擊 / #7b 魔甲抗魔 / #8 palette 還原。#7c 祈禱之戒(回 MP+~25% 損壞)依 RE 忠實實作。
- **主線可破關**:8 里程碑真實 NPC 觸發;`[0x722]` runner/region scripted-event 靜態攻克(re-log-722,57 setter)。
- **杜勝利 + 青衫攻略接線**:B 道具鏈(夢幻紅寶石/變身杖/蓋亞之劍/鑰匙鏈/彩虹水滴/領悟之書/王者之劍)、A boss(巴拉摩斯+怨靈122/殭屍123/索瑪124、甘達特金皇冠、甘達特巢穴 CTY14、八頭大蛇 CTY19、**六大魔人怪106**)、不死鳥拉米亞坐騎(六珠)、**勇氣神殿神父 gate(flag0x13)**、**蓋亞那跳坑下降閘(flag0x213)**。全流程缺口盤點見 `docs/data/walkthrough-flow-audit.md`(11 缺口→8 已解,剩 B-5/B-6/B-7/E-11 不阻塞)。**賢者 gate 道具碼 bug 已修(領悟之書 0x4a)**。道具取得鏈全列(`quest-items.md`,無 stale)。
- **戰鬥系統忠實化**:傷害公式逐一對反組譯確認精確(物理/弱攻/會心×2/咒文)、怪物 AI/數值/遭遇全 RE、**狀態咒**(拜基魯多/史卡拉/魯卡尼/瑪努莎/瑪荷頓 等 base==0 套修正,research+impl 對照 `spell-effects-research.md`)、**rng 升級成長**(RE sub_d9cc:`+= rng(0..(target−當前))`,創角+升級,NULL=確定性零回歸)、手動選咒選單、教會復活費 RE level 表。
- **野外/系統**:野外咒文施放(魯拉/烈米特/特黑洛斯)、per-member 裝備管理 UI、設定選單(RNG 模式 DOS/REAL 存 dq3.cfg)、寶箱開過回饋。
- **存讀檔 v8**:多 slot(6)+ F5/F9 + 標題選單;存名冊/隊伍/道具/位置/船/**劇情進度 storyflags**/晝夜相位/魯拉去過城鎮(magic DQ3SAVE8)。
- **special 事件**:全分類,0 review 待辦。**overworld 旗標 portal**:RE 0x396e 全分支確認 3 條靜態 portal 完備。
- **場景轉場/傳送全接**:門/階梯/出城/下降(section +0xc 轉場表)、type-2 examine warp(0x4ea0,旅の扉/洞穴出口→城)、overworld 旗標 portal。洞窟/迷宮↔城連接在遊戲內可走;`dq3_locwarp`(0xd1f9 codepath 的同連接冗餘資料表)未單獨接,連接由上述機制覆蓋。
- **金字塔施咒當機**:RE 確認 remake 結構免疫(咒文與地圖資源解耦),不復刻原版 crash(`original-known-bugs.md`)。
- **知識庫**:`CONTEXT.md`、`docs/00-re-methodology.md`、`docs/data/*`(oracle-validation / spell-effects-research / day-night-system 等)、`~/.claude` 跨 session 記憶。

## 真正剩餘工作(已對 ground truth 核實)

### 收尾打磨(非阻塞)
- [x] **全專案 markdown 過期審查 ✅**(2026-06-28,第一性原理對 code 核實):掃 102 篇 .md,修正 ~24 項過期/錯誤斷言
  —— 含 ① 順帶揪出**功能回歸**:`playthrough_check.sh` 蓋美拉翅膀回地表測試輸入未跟上新增的傳送選單(改 `cdredde`→`cdreddee`,
  game_tester 回 93/93);② 存檔版本(README/WORKLIST v6→v8)、測試數(86/86→93/93、79/79→動態)、
  日夜 NPC 雙清單(docs/34/35 的 +0/+2 由「Y<0x12c 座標」更正為晝夜選擇)、docs/18 bug 索引(#4/#7a 早已 binary 修好、
  成長表在 EXE DGROUP、255-wrap 非 8-bit 公式)、多處「已實作卻標未實作」(戰鬥施法/狀態系統解毒/商店/音訊引擎/寶箱翻面/
  NPC 擋路/對話派發)、docs/43 撤回的 byte4 假陽性。**殘留 minor**(不誤導,自校或需另跑):docs/19/24 byte-match 確切通過數待 `match_check.py`、
  docs/28/36/27/29 開頭框架語(同文後段已自校)。
- [x] **魯拉/蓋美拉 傳送目的地存讀檔持久 ✅**(2026-06-28):visited 城鎮清單(cty+最後位置+section)寫進
  `dq3_save_pos`(save v8,magic DQ3SAVE8;pos size 變自動擋舊檔);存/讀檔雙向接好;test_save 加 round-trip 斷言。
- [x] **城鎮日夜 NPC 行為還原**(使用者觀察,2026-06-28;**完成**):原版城鎮白天(含黃昏)NPC 走動、
  黑夜部分睡覺/隱藏、有的白天隱藏夜間才現。
  - ✅ **RE 出真實機制**(原版 file 0x452d):**每個 section 開頭兩個 word 是兩份 NPC 表指標**(相對 base)——
    `word[0]`=白天表(人多)、`word[1]`=黑夜表(少數)。原版 `[0x526c]`(日=0/夜=1,由時刻 `[0x251d]` 衍生,
    僅 overworld 前進)選表載入。docs/35 的 0x4f70 旗標只是 NPC **個別劇情可見性**,日夜是**另一層雙清單切換**。
  - ✅ **實作**(`dq3_scene.c` `dq3_scene_load_npcs`):依當前相位(`get_daynight()==2`=黑夜→word1;否則 word0)選表,
    主表無效(0xffff/越界)退另一份(部分 section 日夜共用)。時間僅 overworld 前進 → 進城選一次即還原,與原版一致。
  - ✅ **城內即時切換**(`main.c` `reload_town_daynight()`):拉那魯達/黑暗之燈在城內切日夜 → 重載當前 section NPC(保留座標)。
  - ✅ **驗證**:CTY00 24→6、CTY02 31→10、CTY01 11→4、CTY03 12→4(白天→黑夜),runtime 與資料表一致。
  - 「睡覺」NPC = 室內 section 夜間清單裡帶睡姿 sprite id 的記錄,由同一雙清單切換**資料驅動**載入(slot[2]=sprite id),無需額外碼。
- [~] **晝夜精校**(夜 gated 事件 ✅ 2026-06-27;palette 為必要近似):
  - ✅ **提頓村=テドン 夜 gated 還原**:綠寶珠改夜限定(`g_dn_phase==黑夜` 才開牢門給珠;白天牢門深鎖只見留書),
    忠實 day-night doc §9「夜晚進村開牢門」。game_tester 加白天/夜晚兩斷言(80/80)。
  - 步數已設使用者指定值(白天→黑夜 120 步,每相位 60);原版確切步數計數器多輪 RE 未定位(多層 handler 鏈)。
  - 各相位 palette 為近似:**此環境無 DOSBox oracle 可逐格比色**(記憶 `dq3-no-dosbox-debugger`),
    屬「靜態不可驗」殘留,非 pending dev task;現值(夜 42/44/70%、黃昏 82/62/58%、黎明 72/74/92%)為合理近似。
- [x] **per-member 裝備 4 槽 ✅**(2026-06-27,★RE 更正 5→4 槽):第一性原理從 ITEM.DAT b4 高位反推
  ——精訊版實為 **4 裝備槽**(武器 0x2_/鎧 0x4_/盾 0x6_/兜 0x8_,def 遞增佐證),**非 5 槽**;飾品(戒指/
  手環)無乾淨 b4 部位編碼(0x00 或同道具 0x18),不設槽。`dq3_item_equip_slot`=`(b4>>5)−1`;
  `dq3_recruit` 加 shield/head;戰鬥 def=耐力/2+(鎧+盾+兜 b1 總和);`equip_modal` 4 槽 2×2 管理
  (修舊 `cat&0x40` 把盾 0x60 誤判成鎧的 bug);save v7。dump 驗證 4 槽畫面、game_tester 80/80。

### 打包
- [x] **Windows x64 跨平台打包 ✅**(2026-06-27):mingw-w64 + SDL2 2.30.9 mingw dev,docker 內交叉編譯
  (`scripts/Dockerfile.mingw` + `cmake/mingw-w64-x86_64.cmake` + `tools/package_win.sh`)→ `work/dq3_remake_win64.zip`
  (dq3_remake.exe 647KB + SDL2.dll + run.bat)。entry 用 `SDL_MAIN_HANDLED` 自管 main(免 SDL2main/WinMain);
  `-static-libgcc`(僅依賴 KERNEL32/SDL2/msvcrt);setenv→`_putenv_s` 可攜化。**wine ABI 實機驗證**:載真實
  素材、繪地表一幀正確;Linux build 80/80 無回歸。
- [x] **Linux AppImage ✅**(2026-06-27):`tools/package_appimage.sh` —— 容器內手工 AppDir + ldd 打包 SDL2 及
  46 個相依 .so + AppRun(LD_LIBRARY_PATH + 自動定位 assets_raw)+ 自製圖示(不嵌版權標題畫面);appimagetool
  `--appimage-extract-and-run`(免 FUSE)+ type2-runtime。產 `dist/linux/dq3_remake-x86_64.AppImage`(7.3MB,slim)。
  **實測執行**:bundled SDL2 載真實素材繪地表一幀正確。注:較新 glibc 建置,極舊發行版請改自編。

### 實機展示
- [x] **demo MP4 錄製 ✅**(2026-06-27):引擎加 `DQ3_RECDIR` 逐幀錄製 hook(`dq3_present` 內,headless 也可錄);
  `tools/record_demo.sh` 兩階段(docker 錄 title/地表晝夜/指令窗/裝備 4 槽/戰鬥 → host ffmpeg 組片)。風格對齊作者
  IJ 專案:深藍金字成果說明卡 + 實機片段 + 淡入淡出,1280×720/25fps,~49s。產 `work/video/dq3_remake_demo.mp4`。
  **mp4/截圖含版權渲染像素,依 .gitignore 政策不入版控**(腳本可重生,可掛 GitHub Release)。

### 跨平台延伸(規劃)
- [ ] **macOS / Android 移植**:規劃見 `docs/55-android-macos-port-plan.md`。macOS 工作量小(SDL2 原生 + .app 打包 +
  CI),Android 中–大(SDLActivity/NDK + 素材解壓 + 觸控 UI)。可先做零回歸前置:進入點 `#ifndef __ANDROID__` 守衛、
  素材/存檔路徑 base-dir 抽象。

### 不在此環境做 / 不依賴 DOSBox(記憶 `dq3-no-dosbox-debugger`)
- **DOSBox 逐畫面 oracle**:此環境**無 DOSBox runtime**,不掛在它上面當待辦。已能靜態驗的都驗了
  (整檔反組譯 byte-identical docs/17、標題/注音 IME/boss sprite 對使用者提供截圖一致)。其餘逐畫面
  比對屬「此環境不驗」,**不是** pending dev task,也不作為任何卡關項的下一步。靜態 RE 能解的一律靜態解。

## 關鍵參考 / 工具 / 紀律

- **格式/位址**:docs/02–16(素材)、docs/05–13/15(EXE 邏輯)、`docs/data/exe_funcs.json`(函式圖)、`docs/35`(轉場/鑰匙門/NPC mover)、`docs/37-39`(怪物 AI/數值/遭遇)。
- **build / 測試(容器內,傳 /repo /build)**:
  ```bash
  docker run --rm -v "$PWD":/repo -v dq3build:/build dq3-remake bash -lc 'cd /build && cmake --build .'
  docker run --rm -v "$PWD":/repo -v dq3build:/build dq3-remake bash -lc 'bash /repo/tools/game_tester.sh /repo/assets_raw /build/dq3_remake'
  ```
  > 坑:`game_tester.sh` 硬編 `/repo`/`/build`,host 直跑→全假失敗(環境非回歸)。
- **其他工具**:`scripts/build.sh`、`tools/package.sh`(Linux 打包)、`tools/package_win.sh`(Windows x64 交叉編譯打包,
  用 `scripts/Dockerfile.mingw` + `cmake/mingw-w64-x86_64.cmake`)、`tools/dockrun.sh`(docker uv python)、
  `tools/gen_custom_glyphs.py`(自建字形)、`tools/dosbox_run.sh`+`dq3-dosbox`(原版 oracle)。
- **原始素材**:`assets_raw/`(版權,gitignore;remake 執行期指向它)。
- **不公開**:可玩遊戲包(`work/`)、原始檔、第三方攻略/截圖、render 的版權畫面像素 —— 見各 `.gitignore`。
- **紀律**:不用 subagent;docker first;每階段 commit+push;斷言完成前 git diff/檔案中介數字驗證(memory `re-verify-via-git-not-tool-echo`);commit message 繁中 + `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`。
